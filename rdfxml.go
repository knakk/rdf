package rdf

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"
)

const (
	rdfNS = `http://www.w3.org/1999/02/22-rdf-syntax-ns#`
	xmlNS = `http://www.w3.org/XML/1998/namespace`

	// XML elements
	elAbout           = "about"
	elAboutEach       = "aboutEach"
	elAboutEachPrefix = "aboutEachPrefix"
	elAlt             = "Alt"
	elBag             = "Bag"
	elBagID           = "bagID"
	elBase            = "base"
	elCollection      = "Collection"
	elDataType        = "datatype"
	elDescription     = "Description"
	elID              = "ID"
	elLang            = "lang"
	elLi              = "li"
	elNodeID          = "nodeID"
	elParseType       = "parseType"
	elRDF             = "RDF"
	elResource        = "resource"
	elSeq             = "Seq"
	elType            = "type"
	elXMLNS           = "xmlns"
)

var (
	rdfType      = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"}
	rdfFirst     = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"}
	rdfRest      = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"}
	rdfNil       = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"}
	rdfSubj      = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#subject"}
	rdfPred      = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate"}
	rdfObj       = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#object"}
	rdfStatement = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement"}
)

var rgxpNCName = regexp.MustCompile(`^[\pL_][\d\pL\pM_.-]*$`)

// evalCtx represents the evaluation context for an xml node.
type evalCtx struct {
	Base string
	Subj Subject
	Lang string
	LiN  int
	NS   []string
}

// rdfXMLDecoder decodes Triples from an XML stream.
//
// Deviations from the RDF/XML specification at http://www.w3.org/TR/rdf-syntax-grammar/ :
// - A valid RDF/XML document cannot have to elements with the same ID, but this
//   decoder only emits valid triples as soon as they are available in a stream, and then
//   it's up to the consumer to decide what to do with duplicates.
type rdfXMLDecoder struct {
	dec *xml.Decoder

	// xml parser state
	state     parseXMLFn // current state function
	nextState parseXMLFn // which state function enter on the next call to Decode()
	ns        []string   // prefix and namespaces (only from the top-level element, usually rdf:RDF)
	base      string     // top level xml:base
	bnodeN    int        // anonymous blank node counter
	tok       xml.Token  // current XML token
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

// SetOption sets a ParseOption to the give value
func (d *rdfXMLDecoder) SetOption(o ParseOption, v interface{}) error {
	switch o {
	case Base:
		iri, ok := v.(IRI)
		if !ok {
			return fmt.Errorf("ParseOption \"Base\" must be an IRI.")
		}
		d.ctx.Base = iri.str
	default:
		return fmt.Errorf("RDF/XML decoder doesn't support option: %v", o)
	}
	return nil
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
		// parsing when we reach the corresponding closing tag.
		d.topElem = elem.Name.Space + elem.Name.Local

		// Store prefix and namespaces. Go's XML encoder provides the
		// name space in eachr xml.StartElement, but we need the space
		// to prefix mapping in case of any parseType="Literal".
		d.storePrefixNS(elem)

		// Store top-level base
		if as := attrXML(elem, elBase); as != nil {
			d.base = as[0].Value
		}

		// Store top-level prefix and namespaces
		if as := attrXMLNS(elem); as != nil {
			for _, a := range as {
				d.ns = append(d.ns, a.Value, a.Name.Local)
			}
		}

		if elem.Name.Space != rdfNS || elem.Name.Local != elRDF {
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
			case elDescription:
				d.storePrefixNS(elem)

				if as := attrRDF(elem, elAbout); as != nil {
					d.current.Subj = IRI{str: d.resolve(d.ctx.Base, as[0].Value)}
				}

				if as := attrRDF(elem, elID); as != nil {
					if a := attrRDF(elem, elNodeID); a != nil {
						panic(errors.New("A node element cannot have both rdf:ID and rdf:nodeID"))
					}

					// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-ID-xml-base
					d.current.Subj = IRI{str: d.resolve(d.ctx.Base, "#"+as[0].Value)}
				}

				if as := attrRDF(elem, elNodeID); as != nil {
					if a := attrRDF(elem, elAbout); a != nil {
						panic(errors.New("A node element cannot have both rdf:about and rdf:nodeID"))
					}
					d.current.Subj = Blank{id: fmt.Sprintf("_:%s", as[0].Value)}
				}

				if as := attrRDF(elem, elType); as != nil {
					d.current.Pred = rdfType
					d.current.Obj = IRI{str: d.resolve(d.ctx.Base, as[0].Value)}
					d.triples = append(d.triples, d.current)

					d.nextState = parseXMLPropElemOrNodeEnd
					return nil

					// TODO what if as := attrRest(elem); as != nil ?
				}

				if l := attrXML(elem, elLang); l != nil {
					d.ctx.Lang = l[0].Value
				}

				if len(elem.Attr) == 0 || d.current.Subj == nil {
					// A rdf:Description with no ID or about attribute describes an
					// un-named resource, aka a bNode.
					d.current.Subj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
					d.bnodeN++
				}

				if as := attrRest(elem); as != nil {
					// When a property element's content is string literal, it may be possible
					// to use it as an XML attribute on the containing node element. This can be
					// done for multiple properties on the same node element only if the property
					// element name is not repeated (required by XML — attribute names are unique
					// on an XML element) and any in-scope xml:lang on the property element's
					// string literal (if any) are the same.
					for _, a := range as {
						d.current.Pred = IRI{str: a.Name.Space + a.Name.Local}
						d.parseObjLiteral(a.Value)
						d.triples = append(d.triples, d.current)
					}

					// We now have one or more complete triples and can return.
					// On the next call to Decode(), continue looking for property elements,
					// or the end of the containing node element.
					d.nextState = parseXMLPropElemOrNodeEnd
					return nil
				}

				// Fallthrough scenario; predicate not established, continue parsing
				// property element
				d.nextXMLToken()
				return parseXMLPropElem
			case elBag, elSeq, elAlt:
				d.storePrefixNS(elem)
				d.pushContext() // TODO explain why?

				// Handled as typed node element below
			case elLi, elRDF, elID, elBagID, elAbout, elParseType, elResource, elNodeID, elAboutEach, elAboutEachPrefix:
				panic(fmt.Errorf("disallowed as node element name: rdf:%s", elem.Name.Local))
			default:
				// all other local names are valid as node element name
				// continue as typed node element below
			}
		}
		// By here, all cases of rdf:XXX have returned to another state, except
		// when the element is a typed node element.
		// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-typed-nodes

		if as := attrRDF(elem, elAbout); as != nil {
			d.current.Subj = IRI{str: d.resolve(d.ctx.Base, as[0].Value)}
		}

		if as := attrRDF(elem, elID); as != nil {
			// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-ID-xml-base
			d.current.Subj = IRI{str: d.resolve(d.ctx.Base, "#"+as[0].Value)}
		}

		if d.current.Subj == nil {
			// A typed element without with no attributes
			d.current.Subj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
			d.bnodeN++
		}

		d.current.Pred = rdfType
		d.current.Obj = IRI{elem.Name.Space + elem.Name.Local}
		d.triples = append(d.triples, d.current)

		if as := attrRestWithLn(elem); as != nil {
			for _, a := range as {
				d.current.Pred = IRI{str: a.Name.Space + a.Name.Local}
				d.parseObjLiteral(a.Value)
				d.triples = append(d.triples, d.current)
			}
		}

		d.nextState = parseXMLPropElemOrNodeEnd
		return nil
	case xml.EndElement:
		if elem.Name.Space+elem.Name.Local == d.topElem {
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

		if elem.Name.Space == rdfNS && (elem.Name.Local == elLi || isLn(elem.Name.Local)) {
			// We're in a rdf:Bag, TODO and/or rdf:XXX?
			return parseXMLPropElem
		}

		if len(elem.Attr) == 0 {
			// Since there are no attributes we don't get any hint of
			// the content of the property element. It can either be a
			// string literal, or a new node element. In either case, store
			// the relation from current subject as predicate before continuing.
			d.current.Pred = IRI{str: elem.Name.Space + elem.Name.Local}
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
// This function will establish the object of the triple.
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
			case elDescription:
				// A new element
				if len(elem.Attr) == 0 {
					// Element is a blank node
					d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
					d.bnodeN++
					d.triples = append(d.triples, d.current)

					d.current.Subj = d.current.Obj.(Subject)
					d.nextState = parseXMLPropElemOrNodeEnd
					return nil
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

		d.reifyCheck()

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
			case elDescription:
				// A new element

				d.storePrefixNS(elem)

				if as := attrRest(elem); as != nil {
					// Element is an anonymous blank node
					d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
					d.bnodeN++
					d.triples = append(d.triples, d.current)
					d.reifyCheck()

					d.current.Subj = d.current.Obj.(Subject)

					// Construct triples from attribute elements
					for _, a := range as {
						d.current.Pred = IRI{str: a.Name.Space + a.Name.Local}
						d.parseObjLiteral(a.Value)
						d.triples = append(d.triples, d.current)
					}

					d.nextState = parseXMLPropElemOrNodeEnd
					return nil
				}

				if as := attrRDF(elem, elNodeID); as != nil {
					d.current.Obj = Blank{id: fmt.Sprintf("_:%s", as[0].Value)}
					d.triples = append(d.triples, d.current)
					d.reifyCheck()

					d.current.Subj = d.current.Obj.(Subject)
					d.nextState = parseXMLPropElemOrNodeEnd
					return nil
				}

				// Default case, Element is a blank node
				d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
				d.bnodeN++
				d.triples = append(d.triples, d.current)
				d.reifyCheck()

				d.current.Subj = d.current.Obj.(Subject)
				d.nextState = parseXMLPropElemOrNodeEnd
				return nil

			default:
				panic(fmt.Errorf("parseXMLCharDataOrElemNode second: TODO rdf:!Description"))
			}
		} else {
			if as := attrRDF(elem, elAbout); as != nil {
				d.current.Obj = IRI{str: as[0].Value}
				d.triples = append(d.triples, d.current)

				d.current.Subj = d.current.Obj.(Subject)
				d.nextState = parseXMLPropElemOrNodeEnd
				return nil
			}
			panic(fmt.Errorf("parseXMLCharDataOrElemNode second: TODO not rdf name space, not rdf:about"))
		}
	case xml.EndElement:
		// The closing of the property element; it meanst hat charData
		// represents the string literal as the object.
		d.parseObjLiteral(charData)

		// Emit the complete triple and return
		d.triples = append(d.triples, d.current)
		d.nextState = parseXMLPropElemOrNodeEnd
		return parseXMLPropElemEnd // will return nil after clearing lang and reifying
	default: // xml.Comment, xml.Directive, xml.ProcInst:
		d.nextXMLToken()
		goto second
	}
}

// parseXMLPropElemEnd parses the closing tag of a property element. It should
// only be called when a full triple is ready to be emitted.
func parseXMLPropElemEnd(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.EndElement:
		d.reifyCheck()
		d.lang = "" // clear the in-scope xml:lang

		return nil
	case xml.CharData, xml.Comment, xml.ProcInst:
		d.nextXMLToken()
		return parseXMLPropElemEnd
	default:
		panic(fmt.Errorf("unexpected XML token: %v", elem))
	}
}

// parseXMLPropElem parses a property elements.
func parseXMLPropElem(d *rdfXMLDecoder) parseXMLFn {
	switch elem := d.tok.(type) {
	case xml.StartElement:
		d.storePrefixNS(elem)

		if elem.Name.Space == rdfNS {
			switch elem.Name.Local {
			case elLi:
				d.ctx.LiN++
				d.current.Pred = IRI{str: fmt.Sprintf("http://www.w3.org/1999/02/22-rdf-syntax-ns#_%d", d.ctx.LiN)}
			case elDescription, elRDF, elID, elAbout, elBagID, elParseType, elResource, elNodeID, elAboutEach, elAboutEachPrefix:
				panic(fmt.Errorf("disallowed as property element name: rdf:%s", elem.Name.Local))
			default:
				if isLn(elem.Name.Local) {
					d.current.Pred = IRI{str: fmt.Sprintf("http://www.w3.org/1999/02/22-rdf-syntax-ns#_%s", elem.Name.Local[1:])}
				}
				// Default case, rdf name space
				d.current.Pred = IRI{str: elem.Name.Space + elem.Name.Local}
			}
		} else {
			// Default case, not rdf name space
			d.current.Pred = IRI{str: elem.Name.Space + elem.Name.Local}
		}

		if a := attrRDF(elem, elID); a != nil {
			// Store ID to be used to create the IRI for reified statements (in parseXMLPropElemEnd)
			d.reifyID = "#" + a[0].Value
		}

		if as := attrRDF(elem, elParseType); as != nil {
			switch as[0].Value {
			case "Resource":
				// Omitting rdf:Decsription for blank node
				// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-parsetype-resource
				d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
				d.bnodeN++

				d.triples = append(d.triples, d.current)
				d.reifyCheck()

				d.pushContext()
				d.current.Subj = d.current.Obj.(Subject)
				d.nextXMLToken()
				return parseXMLPropElemOrNodeEnd
				//return nil
			case elCollection:
				// http://www.w3.org/TR/rdf-syntax-grammar/#section-Syntax-parsetype-Collection
				// If we emit triple by triple, we have to keep track of more state to know
				// we're in a collection, to know previous and next nodes and so on, so we
				// parse all items in one go in the followeing state function:
				return parseXMLColl
			default: // case "Literal"
				// All rdf:parseType attribute values other than the strings "Resource",
				// "Literal" or "Collection" are treated as if the value was "Literal".
				if as := attrRDF(elem, elResource); as != nil {
					// TODO find a more generic way to check for attributes which
					// doesn't go together
					panic(errors.New("cannot have both rdf:parseType=\"Literal\" and rdf:resource"))
				}
				// The inner tokens and character data are stored as an XML literal
				d.parseXMLLiteral(elem)
				d.triples = append(d.triples, d.current)

				d.nextState = parseXMLPropElemOrNodeEnd
				return nil
			}
		}

		if as := attrRDF(elem, elResource); as != nil {
			if a := attrRDF(elem, elNodeID); a != nil {
				panic(errors.New("A property element cannot have both rdf:resource and rdf:nodeID"))
			}
			d.current.Obj = IRI{str: d.resolve(d.ctx.Base, as[0].Value)}

			// We have a full triple
			d.triples = append(d.triples, d.current)
			d.reifyCheck()

			if ar := attrRest(elem); ar != nil {
				d.pushContext()
				d.current.Subj = d.current.Obj.(Subject)
				for _, a := range ar {
					d.current.Pred = IRI{str: a.Name.Space + a.Name.Local}
					d.parseObjLiteral(a.Value)
					d.triples = append(d.triples, d.current)
				}
				d.popContext()
			}

			// Before returning, consume the closing tag of the property element.
			d.nextXMLToken()
			d.nextState = parseXMLPropElemOrNodeEnd
			return parseXMLPropElemEnd // this will return nil, without changing d.nextState
		}

		if as := attrRDF(elem, elNodeID); as != nil {
			// predicate is pointing to a blank node,
			// create it and return

			d.current.Obj = Blank{id: fmt.Sprintf("_:%s", as[0].Value)}
			d.triples = append(d.triples, d.current)
			d.reifyCheck()

			d.pushContext()
			d.nextState = parseXMLPropElemOrNodeEnd
			return nil
		}

		if a := attrRDF(elem, elDataType); a != nil {
			d.dt = &IRI{str: d.resolve(d.ctx.Base, a[0].Value)}
		} else {
			// Only check for xml:lang if datatype not found
			// TODO or error if both?
			if l := attrXML(elem, elLang); l != nil {
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

			// We need to reify before we change predicate & object
			d.reifyCheck()

			d.current.Subj = d.current.Obj.(Subject)
			for _, a := range as {
				d.current.Pred = IRI{str: a.Name.Space + a.Name.Local}
				d.parseObjLiteral(a.Value)
				d.triples = append(d.triples, d.current)
			}

			d.nextState = parseXMLPropElemOrNodeEnd
			return nil
		}

		// We don't know what's next; literal data, or new node:
		d.nextXMLToken()
		return parseXMLCharDataOrElemNode
	case xml.EndElement:
		// The containing node element closes. It's valid RDF/XML, but it means
		// there are no more triples to generate on this node.
		return parseXMLPropElemOrNodeEnd
	default: // xml.Comment, xml.CharData, xml.Directive, xml.ProcInst:
		d.nextXMLToken()
		return parseXMLPropElem
	}
}

// parseXMLColl parses collections; all the node elements in a property
// element with attribute parseType="Collection".
// Subject an Predicate is set.
func parseXMLColl(d *rdfXMLDecoder) parseXMLFn {
	d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
	d.bnodeN++

	d.triples = append(d.triples, d.current)

	d.current.Subj = d.current.Obj.(Subject)
	tag := d.tok.(xml.StartElement).Name.Space + d.tok.(xml.StartElement).Name.Local
	first := true
outer:
	for {
		d.nextXMLToken()

		switch elem := d.tok.(type) {
		case xml.StartElement:
			if elem.Name.Space == rdfNS && elem.Name.Local == elDescription {
				if a := attrRDF(elem, elAbout); a != nil {
					if first {
						d.current.Pred = rdfFirst
						d.current.Obj = IRI{str: a[0].Value}
						d.triples = append(d.triples, d.current)
						first = false
					} else {
						d.current.Pred = rdfRest
						d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
						d.bnodeN++
						d.triples = append(d.triples, d.current)

						d.current.Subj = d.current.Obj.(Subject)
						d.current.Pred = rdfFirst
						d.current.Obj = IRI{str: a[0].Value}
						d.triples = append(d.triples, d.current)
					}
				} else {
					panic(fmt.Errorf("parseXMLColl: TODO <rdf:Description> element without rdf:about %v", elem))
				}
			} else {
				panic(fmt.Errorf("parseXMLColl: TODO element node not <rdf:Description> %v", elem))
			}
		case xml.EndElement:
			if elem.Name.Space+elem.Name.Local == tag {
				// we're done
				break outer
			}
		default: // xml.CharData, xml.Comment etc
			continue outer
		}
	}

	// add final statement marking the end of the collection
	d.current.Pred = rdfRest
	d.current.Obj = rdfNil
	d.triples = append(d.triples, d.current)

	return nil
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
	curTok := elem.Name.Space + elem.Name.Local
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
			if elem.Name.Space+elem.Name.Local == curTok {
				// We're done
				break parseLiteral
			}
			b.Write([]byte("</"))
			if elem.Name.Space != "" {
				b.WriteString(d.getPrefix(elem.Name.Space))
				b.Write([]byte(":"))
			}
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

func (d *rdfXMLDecoder) reifyCheck() {
	if d.reifyID != "" {
		iri := IRI{str: d.resolve(d.ctx.Base, d.reifyID)}
		d.triples = append(d.triples,
			Triple{
				Subj: iri,
				Pred: rdfType,
				Obj:  rdfStatement,
			},
			Triple{
				Subj: iri,
				Pred: rdfSubj,
				Obj:  d.current.Subj.(Object),
			},
			Triple{
				Subj: iri,
				Pred: rdfPred,
				Obj:  d.current.Pred.(Object),
			},
			Triple{
				Subj: iri,
				Pred: rdfObj,
				Obj:  d.current.Obj,
			})
		d.reifyID = ""
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

	panic(fmt.Errorf("no prefix found for name space: %q", ns))
}

// getNS returns the in-scope name space for the prefix.
func (d *rdfXMLDecoder) getNS(prefix string) string {
	// First check for context local declarations
	for i := 0; i < len(d.ctx.NS); i += 2 {
		if d.ctx.NS[i+1] == prefix {
			return d.ctx.NS[i]
		}
	}

	// Check in top-level declarations
	for i := 0; i < len(d.ns); i += 2 {
		if d.ns[i+1] == prefix {
			return d.ns[i]
		}
	}

	panic(fmt.Errorf("no name space found for prefix: %q", prefix))
}

// storePrefixNS stores any name space prefixes declared to the element context.
// It also stores the base URI, if xml:base is present.
// TODO also store xml:lang?
func (d *rdfXMLDecoder) storePrefixNS(elem xml.StartElement) {
	if as := attrXMLNS(elem); as != nil {
		for _, a := range as {
			d.ctx.NS = append(d.ctx.NS, a.Value, a.Name.Local)
		}
	}
	if as := attrXML(elem, elBase); as != nil {
		d.ctx.Base = as[0].Value
	}
}

// pushContext pushes the current context on to the context stack, and reset
// the state of current context. It should be called when entering a new element node.
func (d *rdfXMLDecoder) pushContext() {
	d.ctx.Subj = d.current.Subj
	d.ctxStack = append(d.ctxStack, d.ctx)

	// Reset li-counter and subject of current context
	// Base and Lang are inherited and doesn't have to be cleared TODO hm?
	d.ctx.LiN = 0
}

// popContext restores the next context on the stack as the current context.
// If allready at the topmost context, it clears the context and triple subject.
// It should be called when closing an element node.
func (d *rdfXMLDecoder) popContext() {
	switch len(d.ctxStack) {
	case 0:
		d.ctx = evalCtx{}
		d.current.Subj = nil
		d.ctx.Base = d.base
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

func (d *rdfXMLDecoder) resolve(base string, path string) string {
	for i := 0; i < len(path); {
		r, w := utf8.DecodeRuneInString(path[i:])
		i += w
		if r == ':' {
			if strings.HasPrefix(path[i:], "//") {
				// scheme found; it means the path is not relative
				return path
			}
			if i < len(path) {
				// URI is composed of prefix:suffix
				return d.getNS(path[:i-1]) + path[i:]
			}
			break
		}
	}
	if len(base) == 0 {
		return path
	}
	if len(path) == 0 {
		return base[:iriFragmentIdx(base)]
	}
	switch path[0] {
	case '#':
		return base[:iriFragmentIdx(base)] + path
	case '/':
		if len(path) > 1 && path[1] == '/' {
			return base[:iriSchemeEnd(base)] + path
		}
		return base[:iriHostEnd(base)] + path
	case '.':
		numLevels := len(strings.Split(path, "../"))
		return base[:iriSlashIdx(base, numLevels)] + strings.TrimLeft(path, "../")
	default:
		i := iriLastSlashIdx(base)
		if base[i-1] != '/' {
			return base + "/" + path
		}
		return base[:i] + path
	}
}

func iriFragmentIdx(s string) int {
	i := len(s)
	for i > 0 {
		r, w := utf8.DecodeLastRuneInString(s[0:i])
		i -= w
		if r == '#' {
			return i
		}
		if r == '/' {
			break
		}
	}
	return len(s)
}

func iriHostEnd(s string) int {
	i := 0
	l := len(s)
	for i < l {
		r, w := utf8.DecodeRuneInString(s[i:])
		if w == 0 {
			return i
		}
		i += w
		if r == '.' {
			for r, w = utf8.DecodeRuneInString(s[i:]); isAlpha(r); r, w = utf8.DecodeRuneInString(s[i:]) {
				i += w
			}
			if w == 0 {
				return i
			}
			if r == '/' {
				return i
			}
		}
	}
	return i
}

func iriSchemeEnd(s string) int {
	if strings.HasPrefix(s, "http://") {
		return 5
	}
	i := 0
	l := len(s)
	for i < l {
		r, w := utf8.DecodeRuneInString(s[i:])
		if w == 0 {
			return i
		}
		i += w
		if r == ':' {
			if i+2 < l && s[i] == '/' && s[i+1] == '/' {
				return i
			}
		}
	}
	return i
}

func iriLastSlashIdx(s string) int {
	i := len(s)
	for i > 0 {
		r, w := utf8.DecodeLastRuneInString(s[0:i])
		if r == '/' {
			if i > 1 && s[i-w-1] == '/' {
				return len(s)
			}
			return i
		}
		i -= w
	}
	return 0
}

func iriSlashIdx(s string, n int) int {
	c := 0
	i := len(s)
	for {
		r, w := utf8.DecodeLastRuneInString(s[:i])
		if w == 0 {
			break
		}
		i -= w

		if r == '/' {
			if i > 0 && s[i-1] == '/' {
				// Return if we have reached the URI scheme.
				return i + 1
			}
			c++
			if c == n {
				return i + 1
			}
		}
	}
	return i
}

// isLn checks if string matches ^_[1-9]\d*$
func isLn(s string) bool {
	if len(s) < 2 {
		return false
	}
	if s[0] != '_' {
		return false
	}
	if s[1] < '1' || s[1] > '9' {
		return false
	}
	if len(s) > 2 {
		for _, r := range s[2:] {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
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
				if lname == elNodeID || lname == elID {
					if !rgxpNCName.MatchString(a.Value) {
						panic(fmt.Errorf("rdf:%s is not a valid XML NCName: %q", a.Name.Local, a.Value))
					}
					as = append(as, a)
				}
				as = append(as, a)
			case elLi:
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
		if a.Name.Space == elXMLNS {
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
			case elAbout, elParseType, elResource, elDataType, elLi, elType:
				continue
			case elID, elNodeID:
				// validate as NCName:
				if !rgxpNCName.MatchString(a.Value) {
					panic(fmt.Errorf("rdf:%s is not a valid XML NCName: %q", a.Name.Local, a.Value))
				}
				continue
			case elAboutEach, elAboutEachPrefix, elBagID:
				panic(fmt.Errorf("deprecated: rdf:%s", a.Name.Local))
			default:
				if isLn(a.Name.Local) {
					continue
				}
				as = append(as, a)
				continue
			}
		}
		if a.Name.Space == xmlNS || a.Name.Local == elXMLNS || a.Name.Space == "" {
			continue
		}
		as = append(as, a)
	}
	return as
}

// attrRestWithLn is like AttrRest, but includes rdf:_n attributes.
func attrRestWithLn(e xml.StartElement) []xml.Attr {
	var as []xml.Attr
	for _, a := range e.Attr {
		if a.Name.Space == rdfNS {
			switch a.Name.Local {
			case elAbout, elParseType, elResource, elDataType, elLi, elType:
				continue
			case elID, elNodeID:
				// validate as NCName:
				if !rgxpNCName.MatchString(a.Value) {
					panic(fmt.Errorf("rdf:%s is not a valid XML NCName: %q", a.Name.Local, a.Value))
				}
				continue
			case elAboutEach, elAboutEachPrefix, elBagID:
				panic(fmt.Errorf("deprecated: rdf:%s", a.Name.Local))
			default:
				as = append(as, a)
				continue
			}
		}
		if a.Name.Space == xmlNS || a.Name.Local == elXMLNS {
			continue
		}
		as = append(as, a)
	}
	return as
}
