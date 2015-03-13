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

type rdfXMLDecoder struct {
	format Format

	state     parseXMLFn        // state of parser
	base      IRI               // base (default IRI)
	bnodeN    int               // anonymous blank node counter
	ns        map[string]string // map[prefix]namespace
	tokens    [3]token          // 3 token lookahead
	peekCount int               // number of tokens peeked at (position in tokens lookahead array)
	current   ctxTriple         // the current triple beeing parsed

	// xml decoder state
	xmlDec     *xml.Decoder
	xmlTok     xml.Token
	xmlTopElem string
	xmlListN   int
	xmlReifyID string

	// ctxStack keeps track of current and parent triple contexts,
	// needed for parsing recursive structures (list/collections).
	ctxStack []ctxTriple

	// triples contains complete triples ready to be emitted. Usually it will have just one triple,
	// but can have more when parsing nested list/collections. Decode() will always return the first item.
	triples []Triple
}

func newRDFXMLDecoder(r io.Reader) *rdfXMLDecoder {
	return &rdfXMLDecoder{xmlDec: xml.NewDecoder(r)}
}

// SetBase sets the base IRI of the decoder, to be used resolving relative IRIs.
func (d *rdfXMLDecoder) SetBase(i IRI) {
	d.base = i
}

var rgxRDFN = regexp.MustCompile(`_[1-9]\d*$`)

// xmlLit is a struct used to decode the object node of a predicate
// (when parseType="Literal") as a XML literal.
type xmlLit struct {
	XML string `xml:",innerxml"`
}

// Decode parses a RDF/XML document, and returns the next available triple,
// or an error.
func (d *rdfXMLDecoder) Decode() (t Triple, err error) {
	defer d.recover(&err)

	if len(d.triples) == 0 {
		// Run the parser state machine.
		for d.state = parseXMLRootElem; d.state != nil; {
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

func parseXMLRootElem(d *rdfXMLDecoder) parseXMLFn {
	if d.xmlTopElem != "" {
		//
		return parseXMLSubjectNode
	}
	d.nextXMLToken()
	switch elem := d.xmlTok.(type) {
	case xml.StartElement:
		d.xmlTopElem = resolve(elem.Name.Space, elem.Name.Local)
	case xml.Comment, xml.CharData:
		return parseXMLRootElem
	default:
		panic(errors.New("parseXMLRootElem not xml.StartElement"))
	}
	return parseXMLSubjectNode
}

func parseXMLSubjectNode(d *rdfXMLDecoder) parseXMLFn {
	if d.current.Subj != nil {
		return parseXMLPredicate
	}
	d.nextXMLToken()
	switch elem := d.xmlTok.(type) {
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
				d.xmlListN = 1
				d.current.Subj = Blank{id: "_:bag"}
				d.current.Ctx = ctxBag
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
	switch elem := d.xmlTok.(type) {
	case xml.Comment, xml.CharData:
		return parseXMLPredicate
	case xml.StartElement:
		if elem.Name.Space == rdfNS {
			switch elem.Name.Local {
			case "li":
				// We're in a rdf:Bag
				if d.current.Ctx != ctxBag {
					d.current.Ctx = ctxBag
					d.xmlListN = 1
					d.triples = append(d.triples,
						Triple{
							Subj: Blank{id: "_:bag"},
							Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
							Obj:  d.current.Subj.(Object),
						})
					// Parent node is not subject
					d.current.Subj = Blank{id: "_:bag"}
				}
				d.current.Pred = IRI{str: fmt.Sprintf("http://www.w3.org/1999/02/22-rdf-syntax-ns#_%d", d.xmlListN)}
				d.xmlListN++
			default:
				if ln := rgxRDFN.FindString(elem.Name.Local); ln != "" {
					// We're in a rdf:Bag
					if d.current.Ctx != ctxBag {
						d.current.Ctx = ctxBag
						d.xmlListN = 1
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
					d.xmlReifyID = a.Value
				case "parseType":
					switch a.Value {
					case "Literal":
						o := xmlLit{}
						err := d.xmlDec.DecodeElement(&o, &elem)
						if err != nil {
							panic(err)
						}
						d.current.Obj = Literal{
							str:      strings.TrimLeft(o.XML, "\n\t "),
							DataType: xmlLiteral,
						}
						d.triples = append(d.triples, d.current.Triple)
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
				d.triples = append(d.triples, d.current.Triple)
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
	switch elem := d.xmlTok.(type) {
	case xml.Comment:
		return parseXMLObject
	case xml.CharData:
		l := d.current.Obj.(Literal)
		l.str = string(elem)
		d.current.Obj = l
		d.triples = append(d.triples, d.current.Triple)
		//fmt.Printf(d.current.Triple.Serialize(FormatNT))
	default:
		panic(errors.New("parseXMLObject not xml.CharData"))
	}
	if d.xmlReifyID != "" {
		// Now that we have a full triple, we can reify if needed
		d.triples = append(d.triples,
			Triple{
				Subj: IRI{str: resolve(d.base.str, d.xmlReifyID)},
				Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
				Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement"},
			},
			Triple{
				Subj: IRI{str: resolve(d.base.str, d.xmlReifyID)},
				Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#subject"},
				Obj:  d.current.Subj.(Object),
			},
			Triple{
				Subj: IRI{str: resolve(d.base.str, d.xmlReifyID)},
				Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate"},
				Obj:  d.current.Pred.(Object),
			},
			Triple{
				Subj: IRI{str: resolve(d.base.str, d.xmlReifyID)},
				Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#object"},
				Obj:  d.current.Obj.(Object),
			})
		d.xmlReifyID = ""
	}
	return parseXMLCloseStatement
}

func parseXMLCloseStatement(d *rdfXMLDecoder) parseXMLFn {
	d.nextXMLToken()
	switch elem := d.xmlTok.(type) {
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
	d.xmlTok, err = d.xmlDec.Token()
	if err != nil {
		panic(err)
	}
}

func resolve(iri string, s string) string {
	if len(iri) == 0 {
		// A valid IRI cannot be empty, but given some corrupt input,
		// the decoder might generate one.
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
