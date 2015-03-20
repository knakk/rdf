package rdf

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"runtime"
	"strings"
)

const (
	rdfNS = `http://www.w3.org/1999/02/22-rdf-syntax-ns#`
	xmlNS = `http://www.w3.org/XML/1998/namespace`
)

var (
	rdfType = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"}
)

// evalCtx represents the evaluation context for an xml node.
type evalCtx struct {
	Base string
	Subj Subject
	Lang string
	LiN  int
	NS   []string
	Ctx  context
}

type rdfXMLDecoder struct {
	dec *xml.Decoder

	// xml parser state
	state     parseXMLFn // current state function
	nextState parseXMLFn // which state function enter on the next call to Decode()
	ns        []string   // prefix and namespaces (only from the top-level element, usually rdf:RDF)
	bnodeN    int        // anonymous blank node counter
	tok       xml.Token  // current XML token
	peekCount int        // number of tokens peeked at (position in tokens lookahead array)
	topElem   string     // top level element (namespace+localname)
	reifyID   string     // if not "", id to be resolved against the current in-scope Base IRI
	dt        *IRI       // datatype of the Literal to be parsed
	lang      string     // xml element in-scope xml:lang
	current   Triple     // the current triple beeing parsed
	ctx       evalCtx    // current node evaluation context
	ctxStack  []evalCtx  // stack of parent evaluation contexts

	triples []Triple // complete, valid triples to be emitted
}

func newRDFXMLDecoder(r io.Reader) *rdfXMLDecoder {
	return &rdfXMLDecoder{dec: xml.NewDecoder(r), nextState: parseXMLTopElem}
}

// SetBase sets the base IRI of the decoder, to be used resolving relative IRIs.
func (d *rdfXMLDecoder) SetBase(i IRI) {
	d.ctx.Base = i.str
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

		// Store prefix and namespaces. Go's XML encoder provides the
		// name space in eachr xml.StartElement, but we need the space
		// to prefix mapping in case of any parseType="Literal".
		d.storePrefixNS(elem)

		if elem.Name.Space != rdfNS || elem.Name.Local != "RDF" {
			// When there is only one top-level node element,
			// rdf:RDF can be omitted.
			return parseXMLNodeElem
		}

		d.nextXMLToken()
		return parseXMLNodeElem
	default: // xml.Comment, xml.CharData, xml.Directive, xml.ProcInst, xml.EndElement
		// We only care about the top-level start element.
		d.nextXMLToken()
		return parseXMLTopElem
	}
}

// parseXMLNodeElem parses node elements. It will establish the subject of the triple,
// unless its a empty rdf:Description node.
func parseXMLNodeElem(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.StartElement:
		if elem.Name.Space == rdfNS {
			switch elem.Name.Local {
			case "Description":
				d.storePrefixNS(elem)

				// Check for rdf:about, rdf:ID, rdf:nodeID, xml:lang or no attributes, in the order of
				// most common to least common (I belive, TODO gather statistics?)

				if as := attrRDF(elem, "about"); as != nil {
					d.current.Subj = IRI{str: resolve(d.ctx.Base, as[0].Value)}
				}

				if as := attrRDF(elem, "ID"); as != nil {
					// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-ID-xml-base
					d.current.Subj = IRI{str: resolve(d.ctx.Base, "#"+as[0].Value)}
				}

				if as := attrRDF(elem, "nodeID"); as != nil {
					d.current.Subj = Blank{id: fmt.Sprintf("_:%s", as[0].Value)}
				}

				if d.current.Subj != nil {
					// Continue with parsing property elemets if rdf:about/rdfnodeID is the only attribute.
					if len(elem.Attr) == 1 {
						d.nextXMLToken()
						return parseXMLPropElem
					}

					// Assign xml:lang to context if present
					if l := attrXML(elem, "lang"); l != nil {
						d.ctx.Lang = l[0].Value
					}

					// When a property element's content is string literal, it may be possible
					// to use it as an XML attribute on the containing node element. This can be
					// done for multiple properties on the same node element only if the property
					// element name is not repeated (required by XML — attribute names are unique
					// on an XML element) and any in-scope xml:lang on the property element's
					// string literal (if any) are the same.
					if as := attrRest(elem); as != nil {
						for _, a := range as {
							d.current.Pred = IRI{str: resolve(a.Name.Space, a.Name.Local)}
							d.parseObjLiteral(a.Value)
							d.triples = append(d.triples, d.current)
						}

						// We now have one or more complete triples and can return.
						// On the next call to Decode(), continue looking for property elements,
						// or the end of the containing node element.
						d.nextState = parseXMLPropElemOrNodeEnd
						return nil
					}
				}

				if len(elem.Attr) == 0 {
					// A rdf:Description with no ID or about attribute describes an
					// un-named resource, aka a bNode. However, the subject can be
					// defined by the followowing property element (rdf:li, rdf:ID)
					panic(fmt.Errorf("parseXMLNodeElem TODO: empty rdf:Desciption"))
				}

				// Fallthrough scenario; predicate not established, continue parsing
				// property element
				d.nextXMLToken()
				return parseXMLPropElem
			default:
				panic(fmt.Errorf("parseXMLNodeElem TODO: rdf:%s", elem.Name.Local))
			}
		} // By here, all cases of rdf:XXX have returned another parseXMLFn

		if as := attrRDF(elem, "about"); as != nil {
			// Typed node element
			// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-typed-nodes

			d.current.Subj = IRI{str: as[0].Value}
			d.current.Pred = rdfType
			d.current.Obj = IRI{resolve(elem.Name.Space, elem.Name.Local)}
			d.triples = append(d.triples, d.current)

			d.nextState = parseXMLPropElemOrNodeEnd
			return nil
		}
		panic(fmt.Errorf("parseXMLNodeElem TODO: unhandeled node element:%v", elem))
	case xml.EndElement:
		if resolve(elem.Name.Space, elem.Name.Local) == d.topElem {
			// Reached closing tag of top-level element

			// A valid RDF/XML document cannot contain more nodes, so
			// we don't want to allow any more parsing.
			d.nextState = nil

			// TODO: consider checking for more tokens to give feedback on malformed RDF/XML.
			return nil
		}
		panic(fmt.Errorf("parseXMLNodeElem: unexpected closing tag: %v", elem))
	default: // xml.Comment, xml.CharData, xml.Directive, xml.ProcInst:
		d.nextXMLToken()
		return parseXMLNodeElem
	}
}

// parseXMLPropElemOrNodeEnd parses property elements of a containing
// element node, or the end of that element node.
func parseXMLPropElemOrNodeEnd(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.StartElement:
		// The element node has more property elements.

		if len(elem.Attr) == 0 {
			// Since there are no attributes we don't get any hint of
			// the content of the property element. It can either be a
			// string literal, or a new node element. In either case, store
			// the relation from current subject as predicate before continuing.
			d.current.Pred = IRI{str: resolve(elem.Name.Space, elem.Name.Local)}
			d.nextXMLToken()

			return parseXMLCharDataOrElemNode
		}

		// Handle default case, not covered in above contitionals:
		return parseXMLPropElem
	case xml.EndElement:
		// Reached the end of an element node.

		// Restore parent context, if any:
		d.popContext()

		if d.current.Subj != nil {
			// Parent context restored, with subject set.
			// Continue looking for property elements, or the closing
			// of that element node:
			d.nextXMLToken()
			return parseXMLPropElemOrNodeEnd
		}

		// Continue look for more node elements:
		d.nextXMLToken()
		return parseXMLNodeElem
	default: // xml.Comment, xml.CharData, xml.Directive, xml.ProcInst:
		d.nextXMLToken()
		return parseXMLPropElemOrNodeEnd
	}
}

// parseXMLCharDataOrElemNode parses the tokens after a property element
// with no attributes, finding either a string literal or a new element node.
func parseXMLCharDataOrElemNode(d *rdfXMLDecoder) parseXMLFn {
	var charData string

first:
	switch elem := d.tok.(type) {
	case xml.CharData:
		// Could be string literal or the white space between two tokens,
		// store it until we know.
		charData = string(elem)
	case xml.StartElement:
		// Entering a new element. We need to push current context to stack:
		d.pushContext()
		d.pushContext() // TODO doc why twice

		if elem.Name.Space == rdfNS {
			switch elem.Name.Local {
			case "Description":
				// A new element

				if len(elem.Attr) == 0 {
					// Element is a blank node, but we have to parse the next
					// tokens to establish the node identifier (next could be rdf:li etc..)
					goto third

				}
			default:
				panic(fmt.Errorf("parseXMLCharDataOrElemNode first: TODO rdf:!Decsription"))
			}
		} else {
			panic(fmt.Errorf("parseXMLCharDataOrElemNode first: TODO not rdf name space"))
		}
	case xml.EndElement:
		// It's an empty string literal
		d.parseObjLiteral("")

		// Emit the complete triple and return
		d.triples = append(d.triples, d.current)

		d.nextState = parseXMLPropElemOrNodeEnd
		return nil
	default: // xml.Comment, xml.Directive, xml.ProcInst:
		d.nextXMLToken()
		goto first
	}

	d.nextXMLToken()

second:
	switch elem := d.tok.(type) {
	case xml.StartElement:
		// A new node element.
		// (it means that charData was only whitespace between tokens)

		// We need to push stack twice, since popContext is called on the
		// closing of both property element tag, and node element tag.
		d.pushContext()
		d.pushContext()

		if elem.Name.Space == rdfNS {
			switch elem.Name.Local {
			case "Description":
				// A new element

				d.storePrefixNS(elem)

				if as := attrRest(elem); as != nil {
					// Element is an anonymous blank node
					d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
					d.bnodeN++
					d.triples = append(d.triples, d.current)

					d.current.Subj = d.current.Obj.(Subject)

					// Construct triples from attribute elements
					for _, a := range as {
						d.current.Pred = IRI{str: resolve(a.Name.Space, a.Name.Local)}
						d.parseObjLiteral(a.Value)
						d.triples = append(d.triples, d.current)
					}

					d.nextState = parseXMLPropElemOrNodeEnd
					return nil
				}

				if len(elem.Attr) == 0 {
					// Element is a blank node, but we have to parse the next
					// tokens to establish the node identifier (next could be rdf:li etc..)
					goto third

				}
			default:
				panic(fmt.Errorf("parseXMLCharDataOrElemNode second: TODO rdf:!Decsription"))
				// hm return parseXMLNodeElem?
			}
		} else {
			panic(fmt.Errorf("parseXMLCharDataOrElemNode second: TODO not rdf name space"))
		}
	case xml.EndElement:
		// The closing of the property element; it meanst hat charData
		// represents the string literal as the object.
		d.parseObjLiteral(charData)

		// Emit the complete triple and return
		d.triples = append(d.triples, d.current)

		d.nextState = parseXMLPropElemOrNodeEnd
		return nil
	default: // xml.Comment, xml.Directive, xml.ProcInst:
		d.nextXMLToken()
		goto second
	}

third:
	panic("parseXMLCharDataOrElemNode third TODO")
	return nil
}

// parseXMLPropElemEnd parses the closing tag of a property element. It should
// only be called when a full triple is ready to be emitted.
func parseXMLPropElemEnd(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.EndElement:
		d.lang = "" // clear the in-scope xml:lang

		return nil
	default:
		panic(fmt.Errorf("unexpected XML token: %v", elem))
	}
}

// parseXMLPropElem parses a property elements.
func parseXMLPropElem(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.StartElement:
		d.storePrefixNS(elem)

		d.current.Pred = IRI{str: resolve(elem.Name.Space, elem.Name.Local)}

		if as := attrRDF(elem, "resource"); as != nil {
			d.current.Obj = IRI{str: resolve(d.ctx.Base, as[0].Value)}

			// We have a full triple
			d.triples = append(d.triples, d.current)

			// Before returning, consume the closing tag of the property element.
			d.nextXMLToken()
			d.nextState = parseXMLPropElemOrNodeEnd
			return parseXMLPropElemEnd // this will return nil, without changing d.nextState
		}

		if as := attrRDF(elem, "parseType"); as != nil {
			switch as[0].Value {
			case "Literal":
				// The inner tokens and character data are stored as an XML literal
				d.parseXMLLiteral(elem)
				d.triples = append(d.triples, d.current)

				d.nextState = parseXMLPropElemOrNodeEnd
				return nil
			case "Resource":
				// Omitting rdf:Decsription for blank node
				// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-parsetype-resource
				d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
				d.bnodeN++

				d.triples = append(d.triples, d.current)

				d.pushContext()
				d.current.Subj = d.current.Obj.(Subject)
				d.nextState = parseXMLPropElemOrNodeEnd
				return nil
			default:
				panic(fmt.Errorf("parseXMLPropElem TODO parseType=%s", as[0].Value))
			}
		}

		if as := attrRDF(elem, "nodeID"); as != nil {
			// predicate is pointing to a blank node,
			// create it and return

			d.current.Obj = Blank{id: fmt.Sprintf("_:%s", as[0].Value)}
			d.triples = append(d.triples, d.current)

			d.pushContext()
			d.nextState = parseXMLPropElemOrNodeEnd
			return nil
		}

		if a := attrRDF(elem, "datatype"); a != nil {
			d.dt = &IRI{str: resolve(d.ctx.Base, a[0].Value)}
		} else {
			// Only check for xml:lang if datatype not found
			// TODO or error if both?
			if l := attrXML(elem, "lang"); l != nil {
				// store as in-scope lang
				d.lang = l[0].Value
			}
		}

		if as := attrRest(elem); as != nil {
			// If all of the property elements on a blank node element have
			// string literal values with the same in-scope xml:lang value
			// and each of these property elements appears at most once and
			// there is at most one rdf:type property element with a IRI object node,
			// these can be abbreviated by moving them to be property attributes
			// on the containing property element which is made an empty element.
			// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-property-attributes-on-property-element
			d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
			d.bnodeN++
			d.triples = append(d.triples, d.current)
			d.pushContext()

			d.current.Subj = d.current.Obj.(Subject)
			for _, a := range as {
				d.current.Pred = IRI{str: resolve(a.Name.Space, a.Name.Local)}
				d.parseObjLiteral(a.Value)
				d.triples = append(d.triples, d.current)
			}

			d.nextState = parseXMLPropElemOrNodeEnd
			return nil
		}

		// Continue parsing the literal data
		d.nextXMLToken()
		return parseXMLCharData
	case xml.EndElement:
		panic(fmt.Errorf("unexpected end of property element: %v", elem))
	default: // xml.Comment, xml.CharData, xml.Directive, xml.ProcInst:
		d.nextXMLToken()
		return parseXMLPropElem
	}
}

func parseXMLCharData(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.CharData:
		d.parseObjLiteral(string(elem))

		// We have a full triple:
		d.triples = append(d.triples, d.current)

		// Parse the closing of the containg property element before returning
		d.nextState = parseXMLPropElemOrNodeEnd
		d.nextXMLToken()
		return parseXMLPropElemEnd
	case xml.Comment:
		d.nextXMLToken()
		return parseXMLCharData
	default:
		panic(fmt.Errorf("parseXMLCharData: unexpected %v", elem))
	}
}

// parseObjLiteral parses the object from the given character data,
// making sure it get's the in-scope xml:lang and correct datatype.
func (d *rdfXMLDecoder) parseObjLiteral(data string) {
	if d.dt != nil {
		d.current.Obj = Literal{str: data, DataType: *d.dt, lang: d.lang}
		d.dt = nil
	} else if d.lang != "" {
		d.current.Obj = Literal{str: data, DataType: rdfLangString, lang: d.lang}
	} else if d.ctx.Lang != "" {
		d.current.Obj = Literal{str: data, DataType: rdfLangString, lang: d.ctx.Lang}
	} else {
		d.current.Obj = Literal{str: data, DataType: xsdString}
	}
}

// parseXMLLiteral parses XML literals, making sure to declare any
// name spaces used (so that the result is a self-contained XML document).
func (d *rdfXMLDecoder) parseXMLLiteral(elem xml.StartElement) {
	var b bytes.Buffer
	curTok := resolve(elem.Name.Space, elem.Name.Local)
	prefixes := make(map[string]struct{})
parseLiteral:
	for {
		d.nextXMLToken()
		switch elem := d.tok.(type) {
		case xml.StartElement:
			b.Write([]byte("<"))
			if elem.Name.Space != "" {
				b.WriteString(d.getPrefix(elem.Name.Space))
				b.Write([]byte(":"))
				b.WriteString(elem.Name.Local)
				if _, ok := prefixes[elem.Name.Space]; !ok {
					b.Write([]byte(" xmlns:"))
					b.WriteString(d.getPrefix(elem.Name.Space))
					b.Write([]byte("=\""))
					b.WriteString(elem.Name.Space)
					b.Write([]byte("\""))
					prefixes[elem.Name.Space] = struct{}{}
				}
			} else {
				b.WriteString(elem.Name.Local)
			}
			for _, a := range elem.Attr {
				b.Write([]byte(" "))
				if a.Name.Space != "" {
					b.WriteString(d.getPrefix(a.Name.Space))
					b.Write([]byte(":"))
					b.WriteString(a.Name.Local)
					if _, ok := prefixes[a.Name.Space]; !ok {
						b.Write([]byte(" xmlns:"))
						b.WriteString(d.getPrefix(a.Name.Space))
						b.Write([]byte("=\""))
						b.WriteString(a.Name.Space)
						b.Write([]byte("\""))
						prefixes[a.Name.Space] = struct{}{}
					}
				} else {
					b.WriteString(a.Name.Local)
				}
				b.Write([]byte("=\""))
				b.WriteString(a.Value)
				b.Write([]byte("\""))
			}
			b.Write([]byte(">"))
		case xml.EndElement:
			if resolve(elem.Name.Space, elem.Name.Local) == curTok {
				// We're done
				break parseLiteral
			}
			b.Write([]byte("</"))
			b.WriteString(d.getPrefix(elem.Name.Space))
			b.Write([]byte(":"))
			b.WriteString(elem.Name.Local)
			b.Write([]byte(">"))
		case xml.CharData:
			b.Write(elem)
		default:
			panic(fmt.Errorf("parseXMLPropElem: TODO parseType=Literal & token: %v", elem))
		}
	}

	d.current.Obj = Literal{
		str:      b.String(),
		DataType: xmlLiteral,
	}
}

// getPrefix returns the in-scope prefix for the given name space.
func (d *rdfXMLDecoder) getPrefix(ns string) string {
	// First check for context local declarations
	for i := 0; i < len(d.ctx.NS); i += 2 {
		if d.ctx.NS[i] == ns {
			return d.ctx.NS[i+1]
		}
	}

	// Check in top-level declarations
	for i := 0; i < len(d.ns); i += 2 {
		if d.ns[i] == ns {
			return d.ns[i+1]
		}
	}

	panic(fmt.Errorf("no prefix found for name space: %s", ns))
}

// storePrefixNS stores any name space prefixes declared to the element context.
// It also stores the base URI, if xml:base is present.
func (d *rdfXMLDecoder) storePrefixNS(elem xml.StartElement) {
	if as := attrXMLNS(elem); as != nil {
		for _, a := range as {
			d.ctx.NS = append(d.ns, a.Value, a.Name.Local)
		}
	}
	if as := attrXML(elem, "base"); as != nil {
		d.ctx.Base = as[0].Value
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
	if !isRelative(s) {
		return s
	}
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

func isRelative(iri string) bool {
	// TODO implement properly detecting any shceme, see TTL-parser
	return !strings.HasPrefix(iri, "http://")
}

// isLn checks if string matches ^_[1-9]\d*$, and returns the
// digits part of tje match, otherwise empty string.
func isLn(s string) string {
	if s[0] != '_' {
		return ""
	}
	if s[1] < '1' || s[1] > '9' {
		return ""
	}
	for _, r := range s[2:] {
		if r < '0' || r > '9' {
			return ""
		}
	}
	return s[1:]
}

// attrRDF looks for a attribute in the rdf namespace with the given
// local name. It returns the first one found. It returns a slice to
// easily check for no results. It Panics on disallowed attributes.
func attrRDF(e xml.StartElement, lname string) []xml.Attr {
	var as []xml.Attr
	for _, a := range e.Attr {
		if a.Name.Space == rdfNS {
			switch a.Name.Local {
			case lname:
				as = append(as, a)
			case "li":
				panic(fmt.Errorf("unexpected as attribute: rdf:%s", a.Name.Local))
			default:
				// continue
			}
		}
	}
	return as
}

func attrXMLNS(e xml.StartElement) []xml.Attr {
	var as []xml.Attr
	for _, a := range e.Attr {
		if a.Name.Space == "xmlns" {
			as = append(as, a)
		}
	}
	return as
}

func attrXML(e xml.StartElement, lname string) []xml.Attr {
	var as []xml.Attr
	for _, a := range e.Attr {
		if a.Name.Space == xmlNS && a.Name.Local == lname {
			as = append(as, a)
			break
		}
	}
	return as
}

// attrRest filters out all xml and rdf syntax attributes, leaving those
// assumed to be string literal values of the containing node element.
func attrRest(e xml.StartElement) []xml.Attr {
	var as []xml.Attr
	for _, a := range e.Attr {
		if a.Name.Space == rdfNS {
			switch a.Name.Local {
			case "ID", "about", "parseType", "resource", "nodeID", "datatype", "li":
				continue
			default:
				if ln := isLn(a.Name.Local); ln != "" {
					continue
				}
				as = append(as, a)
				continue
			}
		}
		if a.Name.Space == xmlNS {
			continue
		}
		as = append(as, a)
	}
	return as
}
