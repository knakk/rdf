package rdf

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"strings"
)

const rdfNS = `http://www.w3.org/1999/02/22-rdf-syntax-ns#`

// evalCtx represents the evaluation context for an xml node.
type evalCtx struct {
	Base IRI
	Subj Subject
	Lang IRI
	LiN  int
	Ctx  context
}

type rdfXMLDecoder struct {
	dec *xml.Decoder

	// xml parser state
	state    parseXMLFn // statefunction - TODO store it between calls to Decode()?
	bnodeN   int        // anonymous blank node counter
	tok      xml.Token  // current XML token
	topElem  string     // top level element (namespace+localname)
	reifyID  string     // if not "", id to be resolved against the current in-scope Base IRI
	current  Triple     // the current triple beeing parsed
	ctx      evalCtx    // current node evaluation context
	ctxStack []evalCtx  // stack of parent evaluation contexts
	triples  []Triple   // complete, valid triples to be emitted
}

func newRDFXMLDecoder(r io.Reader) *rdfXMLDecoder {
	return &rdfXMLDecoder{dec: xml.NewDecoder(r)}
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
		for d.state = parseXMLTopElem; d.state != nil; {
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

// Parse functions:

// parseXMLTopElem parses the top-level XML document element.
// This is usually rdf:RDF, but can be any node element when
// there is only one top-level element.
func parseXMLTopElem(d *rdfXMLDecoder) parseXMLFn {
	if d.topElem != "" {
		return parseXMLSubjectNode
	}

	d.nextXMLToken()

	switch elem := d.tok.(type) {
	case xml.StartElement:
		// Store the top-level element, so we know we are done
		// parsing when we reach the closing tag. TODO is it necessary?
		d.topElem = resolve(elem.Name.Space, elem.Name.Local)

		return parseXMLSubjectNode
	default:
		// case xml.Comment, xml.CharData, xml.Directive, xml.ProcInst, xml.EndElement:
		// We only care about the top-level element at this point.
		return parseXMLTopElem
	}
}

func parseXMLSubjectNode(d *rdfXMLDecoder) parseXMLFn {
	if d.current.Subj != nil {
		return parseXMLPredicate
	}
	d.nextXMLToken()
	switch elem := d.tok.(type) {
	case xml.Comment, xml.CharData:
		return parseXMLSubjectNode
	case xml.StartElement:
		if elem.Name.Space == rdfNS {
			switch elem.Name.Local {
			case "Description":
				for _, a := range elem.Attr {
					if a.Name.Space == rdfNS && a.Name.Local == "about" {
						d.current.Subj = IRI{str: a.Value}
						return parseXMLPredicate
					}
				}
			case "Bag":
				d.ctx.LiN = 1
				d.current.Subj = Blank{id: "_:bag"}
				d.ctx.Ctx = ctxBag
				d.triples = append(d.triples,
					Triple{
						Subj: d.current.Subj.(Blank),
						Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
						Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag"},
					})
			case "li":
				panic(fmt.Errorf("disallowed as top node element: rdf:%s", elem.Name.Local))
			}
		} else {
			d.current.Subj = IRI{str: resolve(elem.Name.Space, elem.Name.Local)}
			if len(elem.Attr) > 0 {
				for _, a := range elem.Attr {
					if a.Name.Space == rdfNS {
						switch a.Name.Local {
						default:
							panic(fmt.Errorf("disallowed as attribute: rdf:%s", a.Name.Local))
						}
					} // TODO if not rdf:xx ?
				}

			}
		}
	case xml.EndElement:
		d.current.Subj = nil
		return nil
	default:
		panic(errors.New("parseXMLSubjectNode not xml.StartElement"))
	}

	return parseXMLPredicate
}

func parseXMLPredicate(d *rdfXMLDecoder) parseXMLFn {
	d.nextXMLToken()
	switch elem := d.tok.(type) {
	case xml.Comment, xml.CharData:
		return parseXMLPredicate
	case xml.StartElement:
		if elem.Name.Space == rdfNS {
			switch elem.Name.Local {
			case "li":
				// We're in a rdf:Bag
				if d.ctx.Ctx != ctxBag {
					d.ctx.Ctx = ctxBag
					d.ctx.LiN = 1
					d.triples = append(d.triples,
						Triple{
							Subj: Blank{id: "_:bag"},
							Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
							Obj:  d.current.Subj.(Object),
						})
					// Parent node is not subject
					d.current.Subj = Blank{id: "_:bag"}
				}
				d.current.Pred = IRI{str: fmt.Sprintf("http://www.w3.org/1999/02/22-rdf-syntax-ns#_%d", d.ctx.LiN)}
				d.ctx.LiN++
			default:
				if ln := rgxRDFN.FindString(elem.Name.Local); ln != "" {
					// We're in a rdf:Bag
					if d.ctx.Ctx != ctxBag {
						d.ctx.Ctx = ctxBag
						d.ctx.LiN = 1
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
					d.current.Pred = IRI{str: resolve(elem.Name.Space, elem.Name.Local)}
				}
			}
		} else {
			d.current.Pred = IRI{str: resolve(elem.Name.Space, elem.Name.Local)}
		}
		dt := xsdString
		for _, a := range elem.Attr {
			if a.Name.Space == rdfNS {
				switch a.Name.Local {
				case "datatype":
					dt = IRI{str: a.Value}
					goto parseCharData
				case "ID":
					d.reifyID = a.Value
				case "parseType":
					switch a.Value {
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
						return nil // TODO end parseXMLFn
					case "Resource":
						// TODO work here
						// d.pushContext()
					default:
						panic(fmt.Errorf("parseType = %q", a.Value))
					}
				default:
					panic(fmt.Errorf("disallowed as attribute: rdf:%s", a.Name.Local))
				}
			} else {
				// object is a blank node
				d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
				d.bnodeN++
				d.triples = append(d.triples, d.current)
				d.triples = append(d.triples,
					Triple{
						Subj: d.current.Obj.(Blank),
						Pred: IRI{str: resolve(a.Name.Space, a.Name.Local)},
						Obj:  Literal{str: a.Value, DataType: dt},
					})
				return parseXMLCloseStatement
			}
		}
	parseCharData:
		d.current.Obj = Literal{DataType: dt}
		return parseXMLObject
	case xml.EndElement:
		d.current.Subj = nil
		return nil
	default:
		panic(errors.New("parseXMLPredicate not xml.StartElement"))
	}
}

func parseXMLObject(d *rdfXMLDecoder) parseXMLFn {
	d.nextXMLToken()
	switch elem := d.tok.(type) {
	case xml.Comment:
		return parseXMLObject
	case xml.CharData:
		l := d.current.Obj.(Literal)
		l.str = string(elem)
		d.current.Obj = l
		d.triples = append(d.triples, d.current)
		//fmt.Printf(d.current.Triple.Serialize(FormatNT))
	default:
		panic(errors.New("parseXMLObject not xml.CharData"))
	}
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
	return parseXMLCloseStatement
}

func parseXMLCloseStatement(d *rdfXMLDecoder) parseXMLFn {
	d.nextXMLToken()
	switch elem := d.tok.(type) {
	case xml.Comment, xml.CharData:
		return parseXMLCloseStatement
	case xml.EndElement:
		return nil
	default:
		panic(fmt.Errorf("parseXMLCloseStatement: TODO: %v", elem))
	}
}

// parseXMLFn represents the state of the parser as a function that returns the next state.
type parseXMLFn func(*rdfXMLDecoder) parseXMLFn

// Helper functions:

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
