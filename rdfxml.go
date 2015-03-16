package rdf

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"strings"
)

const (
	rdfNS = `http://www.w3.org/1999/02/22-rdf-syntax-ns#`
	xmlNS = `http://www.w3.org/XML/1998/namespace`
)

// evalCtx represents the evaluation context for an xml node.
type evalCtx struct {
	Base IRI
	Subj Subject
	Lang string
	LiN  int
	Ctx  context
}

type rdfXMLDecoder struct {
	dec *xml.Decoder

	// xml parser state
	state     parseXMLFn // current state function
	nextState parseXMLFn // which state function enter on the next call to Decode()
	bnodeN    int        // anonymous blank node counter
	tok       xml.Token  // current XML token
	topElem   string     // top level element (namespace+localname)
	reifyID   string     // if not "", id to be resolved against the current in-scope Base IRI
	dt        *IRI       // datatype of the Literal to be parsed
	current   Triple     // the current triple beeing parsed
	ctx       evalCtx    // current node evaluation context
	ctxStack  []evalCtx  // stack of parent evaluation contexts
	triples   []Triple   // complete, valid triples to be emitted
}

func newRDFXMLDecoder(r io.Reader) *rdfXMLDecoder {
	return &rdfXMLDecoder{dec: xml.NewDecoder(r), nextState: parseXMLTopElem}
}

// SetBase sets the base IRI of the decoder, to be used resolving relative IRIs.
func (d *rdfXMLDecoder) SetBase(i IRI) {
	d.ctx.Base = i
}

var rgxRDFN = regexp.MustCompile(`_[1-9]\d*$`)

// xmlLit is a struct used to decode the object node of a predicate
// (when parseType="Literal") as a XML literal.
// TODO if possible find a way to parse this by accessing the []byte directly,
// without going via a struct..
type xmlLit struct {
	XML string `xml:",innerxml"`
}

// Decode parses a RDF/XML document, and returns the next available triple,
// or an error.
func (d *rdfXMLDecoder) Decode() (t Triple, err error) {
	defer d.recover(&err)

	if len(d.triples) == 0 {
		// Run the parser state machine.
		d.nextXMLToken()
		for d.state = d.nextState; d.state != nil; {
			d.state = d.state(d)
		}

		if len(d.triples) == 0 {
			// No triples left in document
			return t, io.EOF
		}
	}

	t = d.triples[0]
	d.triples = d.triples[1:]
	return t, err
}

// DecodeAll parses a compete RDF/XML document and returns the valid triples,
// or an error.
func (d *rdfXMLDecoder) DecodeAll() ([]Triple, error) {
	var ts []Triple
	for t, err := d.Decode(); err != io.EOF; t, err = d.Decode() {
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}

// parseXMLFn represents the state of the parser as a function that returns the
// next state. A new xml.Token is assumed to be generated and stored in d.tok
// before entering a new state function.
type parseXMLFn func(*rdfXMLDecoder) parseXMLFn

// parseXMLTopElem parses the top-level XML document element.
// This is usually rdf:RDF, but can be any node element when
// there is only one top-level element.
func parseXMLTopElem(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.StartElement:
		// Store the top-level element, so we know we are done
		// parsing when we reach the corresponding closing tag=TODO!
		d.topElem = resolve(elem.Name.Space, elem.Name.Local)
		d.nextState = parseXMLNodeElem
		d.nextXMLToken()
		return parseXMLNodeElem
	default: // xml.Comment, xml.CharData, xml.Directive, xml.ProcInst, xml.EndElement
		// We only care about the top-level start element.
		d.nextXMLToken()
		return parseXMLTopElem
	}
}

// parseXMLNodeElem parses node elements, establishing the subject of the triple.
func parseXMLNodeElem(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.StartElement:
		if elem.Name.Space == rdfNS {
			switch elem.Name.Local {
			case "Description":
				for _, a := range elem.Attr {
					if a.Name.Space == rdfNS && a.Name.Local == "about" {
						d.current.Subj = IRI{str: a.Value}
						d.nextXMLToken()
						return parseXMLPropertyElem
					}
				}
			case "Bag":
				d.current.Subj = Blank{id: "_:bag"}
				d.ctx.Ctx = ctxBag
				d.triples = append(d.triples,
					Triple{
						Subj: d.current.Subj.(Blank),
						Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"}, // TODO global var
						Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag"},  // TODO global var
					})
				d.nextXMLToken()
				return parseXMLPropertyElem
			case "li":
				panic(fmt.Errorf("disallowed as top node element: rdf:%s", elem.Name.Local))
			default:
				panic(fmt.Errorf("parseXMLNodeElem: TODO: %s", elem))
			}
		}
		// not rdf namepsace:
		d.current.Subj = IRI{str: resolve(elem.Name.Space, elem.Name.Local)}
		if len(elem.Attr) == 0 {
			// only ask for new token if there are no attributes TODO explain why
			d.nextXMLToken()
		}
		return parseXMLPropertyElem
	case xml.EndElement:
		if resolve(elem.Name.Space, elem.Name.Local) == d.topElem {
			// Reached closing tag of top-level element
			d.nextState = nil
			return nil // or panic(io.EOF)?
		}
		panic(fmt.Errorf("parseXMLNodeElem: unexpected closing tag: %v", elem))
	default: // xml.Comment, xml.CharData, xml.Directive, xml.ProcInst:
		d.nextXMLToken()
		return parseXMLNodeElem
	}
}

// parseXMLPropertyElem parses property elements, establishing the element predicative.
// Subject must be set when entering this state.
func parseXMLPropertyElem(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.StartElement:
		if elem.Name.Space != rdfNS {
			d.current.Pred = IRI{str: resolve(elem.Name.Space, elem.Name.Local)}
		} else {
			// handle rdf:li rdf:_n (& TODO rdf:Description) if present
			switch elem.Name.Local {
			case "li":
				// We're in a rdf:Bag
				if d.ctx.Ctx != ctxBag {
					d.ctx.Ctx = ctxBag
					d.ctx.LiN = 0
					d.triples = append(d.triples,
						Triple{
							Subj: Blank{id: "_:bag"},
							Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
							Obj:  d.current.Subj.(Object),
						})
					// Parent node is not subject
					d.current.Subj = Blank{id: "_:bag"}
				}
				d.ctx.LiN++
				d.current.Pred = IRI{str: fmt.Sprintf("http://www.w3.org/1999/02/22-rdf-syntax-ns#_%d", d.ctx.LiN)}
			default:
				// check for rdf:_n
				if ln := rgxRDFN.FindString(elem.Name.Local); ln != "" {
					// We're in a rdf:Bag
					if d.ctx.Ctx != ctxBag {
						d.ctx.Ctx = ctxBag
						d.ctx.LiN = 0
						d.triples = append(d.triples,
							Triple{
								Subj: Blank{id: "_:bag"},
								Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
								Obj:  d.current.Subj.(Object),
							})
						// Parent node is not subject
						d.current.Subj = Blank{id: "_:bag"}
					}
					d.current.Pred = IRI{str: fmt.Sprintf("http://www.w3.org/1999/02/22-rdf-syntax-ns#%s", ln)}
				} else {
					// other rdf:Local TODO valid?
					d.current.Pred = IRI{str: resolve(elem.Name.Space, elem.Name.Local)}
				}
			}

		}
		if len(elem.Attr) == 0 {
			// Object is the literal character data enclosed by property element tags:
			d.nextXMLToken()
			return parseXMLObjectData
		}

		// TODO it's a bit messy to have to check for all these different attributes,
		// but given that they can come in any order, it's hard to solve by just looping
		// over []elem.Attr since some attributes takes precedens over others...

		if as := attrRDF(elem, "datatype"); as != nil {
			d.dt = &IRI{as[0].Value}
		}

		if as := attrXMLLang(elem); as != nil {
			d.ctx.Lang = as[0].Value
		}

		if as := attrRDF(elem, "ID"); as != nil {
			d.reifyID = as[0].Value
		}

		if as := attrRDF(elem, "parseType"); as != nil {
			switch as[0].Value {
			case "Literal":
				o := xmlLit{}
				err := d.dec.DecodeElement(&o, &elem)
				if err != nil {
					panic(err)
				}
				d.current.Obj = Literal{
					str:      strings.TrimLeft(o.XML, "\n\t "),
					DataType: xmlLiteral,
				}
				d.triples = append(d.triples, d.current)
				// we're done TODO assert d.nextState=parseXMLPropertyElemOrNodeEnd?
				return nil
			case "Resource":
				d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
				d.bnodeN++
				d.pushContext()
				d.triples = append(d.triples, d.current)
				d.current.Subj = d.current.Obj.(Subject)
				return nil
			default:
				panic(fmt.Errorf("TODO parseType=%q", as[0].Value))
			}
		}

		if as := attrRDF(elem, "resource"); as != nil {
			// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-empty-property-elements
			d.current.Obj = IRI{str: as[0].Value}
			d.triples = append(d.triples, d.current)
			d.nextXMLToken()
			return parseXMLPropertyElemEnd
		}

		if as := attrRest(elem); as != nil {
			// The attribute is predicate and object on the containing property element which is made an empty element.
			// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-property-attributes-on-property-element
			d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
			d.bnodeN++
			d.triples = append(d.triples, d.current)
			dt := xsdString
			if d.ctx.Lang != "" {
				dt = rdfLangString
			}
			for _, a := range as {
				if TermsEqual(dt, rdfLangString) {
					d.triples = append(d.triples,
						Triple{
							Subj: d.current.Obj.(Blank),
							Pred: IRI{str: resolve(a.Name.Space, a.Name.Local)},
							Obj:  Literal{str: a.Value, DataType: dt, lang: d.ctx.Lang},
						})
				} else {
					d.triples = append(d.triples,
						Triple{
							Subj: d.current.Obj.(Blank),
							Pred: IRI{str: resolve(a.Name.Space, a.Name.Local)},
							Obj:  Literal{str: a.Value, DataType: dt},
						})
				}
			}
			d.nextXMLToken()
			return parseXMLPropertyElemEnd
		}

		d.nextXMLToken()
		return parseXMLObjectData
	case xml.EndElement:
		panic(fmt.Errorf("parseXMLPropertyElem: unexpected %v", elem))
	default: // xml.Comment, xml.CharData, xml.Directive, xml.ProcInst:
		d.nextXMLToken()
		return parseXMLPropertyElem
	}
}

// parseXMLObjectData parses the inner character data of a property node,
// establishing the object of the triple.
// The subject and predicate of the triple must be set when entering this state.
func parseXMLObjectData(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.CharData:
		if d.dt != nil {
			// The datatype have been specified with the rdf:datatype attribute
			if TermsEqual(*d.dt, rdfLangString) {
				d.current.Obj = Literal{str: string(elem), DataType: rdfLangString, lang: d.ctx.Lang}
			} else {
				d.current.Obj = Literal{str: string(elem), DataType: *d.dt}
			}
			d.dt = nil
		} else {
			if d.ctx.Lang != "" {
				d.current.Obj = Literal{str: string(elem), DataType: rdfLangString, lang: d.ctx.Lang}
			} else {
				d.current.Obj = Literal{str: string(elem), DataType: xsdString}
			}
		}
		// We have a full triple:
		d.triples = append(d.triples, d.current)

		d.nextXMLToken()
		return parseXMLPropertyElemEnd
	case xml.Comment:
		d.nextXMLToken()
		return parseXMLPropertyElem
	default:
		panic(fmt.Errorf("parseXMLPropertyElem: unexpected %v", elem))
	}
}

// parseXMLPropertyElemEnd parses the closing tag of a property element.
// The current triple must be complete when entering this state.
func parseXMLPropertyElemEnd(d *rdfXMLDecoder) parseXMLFn {
	d.reifyCheck()
	switch elem := d.tok.(type) {
	case xml.EndElement:
		// When parsing again, we might get another property element, or
		// the closing of the node element.
		d.nextState = parseXMLPropertyElemOrNodeEnd
		// We're done, return the triple
		return nil
	case xml.Comment:
		d.nextXMLToken()
		return parseXMLPropertyElemEnd
	default:
		panic(fmt.Errorf("parseXMLPropertyElemEnd: unexpected %v", elem))
	}
}

// parseXMLPropertyElemOrNodeEnd parses a property element (begining a new triple),
// or the end of a node element.
// Subject must be set when entering this state.
func parseXMLPropertyElemOrNodeEnd(d *rdfXMLDecoder) parseXMLFn {
	d.reifyCheck()
	switch elem := d.tok.(type) {
	case xml.Comment, xml.CharData:
		d.nextXMLToken()
		return parseXMLPropertyElemOrNodeEnd
	case xml.StartElement:
		// Entering a new triple with the same subject as previously emitted triple
		return parseXMLPropertyElem
	case xml.EndElement:
		// Restore parent context if any
		d.popContext()
		d.nextState = parseXMLNodeElem

		// Look for more node elements, or property elements if subject is established
		d.nextXMLToken()
		if d.current.Subj != nil {
			return parseXMLPropertyElem
		}
		return parseXMLNodeElem
	default:
		panic(fmt.Errorf("parseXMLPropertyElemOrNodeEnd: unexpected %v", elem))
	}
}

func (d *rdfXMLDecoder) reifyCheck() {
	if d.reifyID != "" {
		// Now that we have a full triple, we can reify if needed
		d.triples = append(d.triples,
			Triple{
				Subj: IRI{str: resolve(d.ctx.Base.str, d.reifyID)},
				Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
				Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement"},
			},
			Triple{
				Subj: IRI{str: resolve(d.ctx.Base.str, d.reifyID)},
				Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#subject"},
				Obj:  d.current.Subj.(Object),
			},
			Triple{
				Subj: IRI{str: resolve(d.ctx.Base.str, d.reifyID)},
				Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate"},
				Obj:  d.current.Pred.(Object),
			},
			Triple{
				Subj: IRI{str: resolve(d.ctx.Base.str, d.reifyID)},
				Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#object"},
				Obj:  d.current.Obj.(Object),
			})
		d.reifyID = ""
	}
}

// pushContext pushes the current context on to the context stack.
func (d *rdfXMLDecoder) pushContext() {
	d.ctx.Subj = d.current.Subj
	d.ctxStack = append(d.ctxStack, d.ctx)
}

// popContext restores the next context on the stack as the current context.
// If allready at the topmost context, it clears the context and triple subject.
func (d *rdfXMLDecoder) popContext() {
	switch len(d.ctxStack) {
	case 0:
		d.ctx = evalCtx{}
		d.current.Subj = nil
	case 1:
		d.ctx = d.ctxStack[0]
		d.current.Subj = d.ctxStack[0].Subj
		d.ctxStack = d.ctxStack[:0]
	default:
		d.ctx = d.ctxStack[len(d.ctxStack)-1]
		d.current.Subj = d.ctx.Subj
		d.ctxStack = d.ctxStack[:len(d.ctxStack)-1]
	}
}

// recover catches non-runtime panics and binds the panic error
// to the given error pointer.
func (d *rdfXMLDecoder) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			// Don't recover from runtime errors.
			panic(e)
		}
		//d.stop() something to clean up?
		*errp = e.(error)
	}
	return
}

func (d *rdfXMLDecoder) nextXMLToken() {
	var err error
	d.tok, err = d.dec.Token()
	if err != nil {
		panic(err)
	}
}

func resolve(iri string, s string) string {
	// TODO implement as described in:
	// http://www.w3.org/TR/xmlbase/#resolution
	if len(iri) == 0 {
		return s
	}
	switch iri[len(iri)-1] {
	case '#':
		if len(s) > 1 && s[0] == '#' {
			return iri + s[1:]
		}
		return iri + s
	case '/':
		return iri + s
	default:
		if len(s) > 1 && s[0] == '#' {
			return iri + s
		}
		return iri + "#" + s
	}
}

func attrRDF(e xml.StartElement, lname string) []xml.Attr {
	var as []xml.Attr
	for _, a := range e.Attr {
		if a.Name.Space == rdfNS {
			switch a.Name.Local {
			case lname:
				as = append(as, a)
				break
			case "ID", "datatype", "parseType", "resource":
				// valid rdf:attributes
			default:
				panic(fmt.Errorf("unexpected as attribute: rdf:%s", a.Name.Local))
			}
		}
	}
	return as
}

func attrXMLLang(e xml.StartElement) []xml.Attr {
	var as []xml.Attr
	for _, a := range e.Attr {
		if a.Name.Space == xmlNS && a.Name.Local == "lang" {
			as = append(as, a)
			break
		}
	}
	return as

}

func attrRest(e xml.StartElement) []xml.Attr {
	// attrs not namepsace rdf or xml:lang
	var as []xml.Attr
	for _, a := range e.Attr {
		if a.Name.Space != rdfNS && a.Name.Space != xmlNS {
			as = append(as, a)
		}
	}
	return as
}
