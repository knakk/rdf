// Package rdf provides functionality for working with RDF resources, including
// parsing and serialization of the various RDF formats.
//
// Data model
//
// The package adhers to the RDF data model as described in http://www.w3.org/TR/rdf11-concepts/.
//
// Data structures
//
// TODO.
//
// Encoding and decoding
//
// The package aims to support all the RDF serialization formats standardized by W3C. Currently the following are implemented:
//  Format     | Decode | Encode
//  -----------|--------|--------
//  RDF/XML    | x      | -
//  N-Triples  | x      | x
//  N-Quads    | x      | x
//  Turtle     | x      | x
//  TriG       | -      | -
//  JSON-LD    | -      | -
//
// The parsers are implemented as streaming decoders, consuming an io.Reader
// and emitting triples/quads as soon as they are available. Simply call
// Decode() until the reader is exhausted and emits io.EOF:
//
//    f, err := os.Open("mytriples.ttl")
//    if err != nil {
//        // handle error
//    }
//    dec := rdf.NewTripleDecoder(f, rdf.Turtle)
//    for triple, err := dec.Decode(); err != io.EOF; triple, err = dec.Decode() {
//        // do something with triple ..
//    }
//
// The encoders work similarily.
// For a complete working example, see the rdf2rdf application, which converts between different serialization formats using the decoders and encoders of the rdf package: https://github.com/knakk/rdf2rdf.
package rdf

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// DateFormat defines the string representation of xsd:DateTime values. You can override
// it if you need another layout.
var DateFormat = time.RFC3339

// The XML schema built-in datatypes (xsd):
// https://dvcs.w3.org/hg/rdf/raw-file/default/rdf-concepts/index.html#xsd-datatypes
var (
	// Core types:                                                    // Corresponding Go datatype:

	xsdString  = IRI{str: "http://www.w3.org/2001/XMLSchema#string"}  // string
	xsdBoolean = IRI{str: "http://www.w3.org/2001/XMLSchema#boolean"} // bool
	xsdDecimal = IRI{str: "http://www.w3.org/2001/XMLSchema#decimal"} // float64
	xsdInteger = IRI{str: "http://www.w3.org/2001/XMLSchema#integer"} // int

	// IEEE floating-point numbers:

	xsdDouble = IRI{str: "http://www.w3.org/2001/XMLSchema#double"} // float64
	xsdFloat  = IRI{str: "http://www.w3.org/2001/XMLSchema#float"}  // float64

	// Time and date:

	//xsdDate = IRI{str: "http://www.w3.org/2001/XMLSchema#date"}
	//xsdTime          = IRI{str: "http://www.w3.org/2001/XMLSchema#time"}
	xsdDateTime = IRI{str: "http://www.w3.org/2001/XMLSchema#dateTime"} // time.Time
	//xsdDateTimeStamp = IRI{str: "http://www.w3.org/2001/XMLSchema#dateTimeStamp"}

	// Recurring and partial dates:

	//xsdYear              = IRI{str: "http://www.w3.org/2001/XMLSchema#gYear"}
	//xsdMonth             = IRI{str: "http://www.w3.org/2001/XMLSchema#gMonth"}
	//xsdDay               = IRI{str: "http://www.w3.org/2001/XMLSchema#gDay"}
	//xsdYearMonth         = IRI{str: "http://www.w3.org/2001/XMLSchema#gYearMonth"}
	//xsdDuration          = IRI{str: "http://www.w3.org/2001/XMLSchema#Duration"}
	//xsdYearMonthDuration = IRI{str: "http://www.w3.org/2001/XMLSchema#yearMonthDuration"}
	//xsdDayTimeDuration   = IRI{str: "http://www.w3.org/2001/XMLSchema#dayTimeDuration"}

	// Limited-range integer numbers

	xsdByte = IRI{str: "http://www.w3.org/2001/XMLSchema#byte"} // []byte
	//xsdShort = IRI{str: "http://www.w3.org/2001/XMLSchema#short"} // int16
	xsdInt = IRI{str: "http://www.w3.org/2001/XMLSchema#int"} // int32
	//xsdLong  = IRI{str: "http://www.w3.org/2001/XMLSchema#long"}  // int64

	// Various

	rdfLangString = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"} // string
	xmlLiteral    = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral"} // string
)

// Format represents a RDF serialization format.
type Format int

// Supported parser/serialization formats for Triples and Quads.
const (
	// Triple serialization:

	NTriples Format = iota
	Turtle
	RDFXML
	// TODO: JSON-LD

	// Quad serialization:

	NQuads // N-Quads
	// TODO: Format TriG

	// Internal formats
	formatInternal
)

// Term represents an RDF term. There are 3 term types: Blank node, Literal and IRI.
type Term interface {
	// Serialize returns a string representation of the Term in the specified serialization format.
	Serialize(Format) string

	// String returns the term as it is stored, without any modifications.
	String() string

	// Type returns the Term type.
	Type() TermType
}

// TermType describes the type of RDF term: Blank node, IRI or Literal
type TermType int

// Exported RDF term types.
const (
	TermBlank TermType = iota
	TermIRI
	TermLiteral
)

// Blank represents a RDF blank node; an unqualified IRI with identified by a label.
type Blank struct {
	id string
}

// validAsSubject denotes that a Blank node is valid as a Triple's Subject.
func (b Blank) validAsSubject() {}

// validAsObject denotes that a Blank node is valid as a Triple's Object.
func (b Blank) validAsObject() {}

// Serialize returns a string representation of a Blank node.
func (b Blank) Serialize(f Format) string {
	return b.id
}

// Type returns the TermType of a blank node.
func (b Blank) Type() TermType {
	return TermBlank
}

// String returns the Blank node label
func (b Blank) String() string {
	return b.id[2:]
}

// NewBlank returns a new blank node with a given label. It returns
// an error only if the supplied label is blank.
func NewBlank(id string) (Blank, error) {
	if len(strings.TrimSpace(id)) == 0 {
		return Blank{}, errors.New("blank id")
	}
	return Blank{id: "_:" + id}, nil
}

// IRI represents a RDF IRI resource.
type IRI struct {
	str string
}

// validAsSubject denotes that an IRI is valid as a Triple's Subject.
func (u IRI) validAsSubject() {}

// validAsPredicate denotes that an IRI is valid as a Triple's Predicate.
func (u IRI) validAsPredicate() {}

// validAsObject denotes that an IRI is valid as a Triple's Object.
func (u IRI) validAsObject() {}

// Type returns the TermType of a IRI.
func (u IRI) Type() TermType {
	return TermIRI
}

// String returns the IRI string.
func (u IRI) String() string {
	return u.str
}

// Serialize returns a string representation of an IRI.
func (u IRI) Serialize(f Format) string {
	return fmt.Sprintf("<%s>", u.str)
}

// Split returns the prefix and suffix of the IRI string, splitted at the first
// '/' or '#' character, in reverse order of the string.
func (u IRI) Split() (prefix, suffix string) {
	i := len(u.str)
	for i > 0 {
		r, w := utf8.DecodeLastRuneInString(u.str[0:i])
		if r == '/' || r == '#' {
			prefix, suffix = u.str[0:i], u.str[i:len(u.str)]
			break
		}
		i -= w
	}
	return prefix, suffix
}

// NewIRI returns a new IRI, or an error if it's not valid.
//
// A valid IRI cannot be empty, or contain any of the disallowed characters: [\x00-\x20<>"{}|^`\].
func NewIRI(iri string) (IRI, error) {
	// http://www.ietf.org/rfc/rfc3987.txt
	if len(iri) == 0 {
		return IRI{}, errors.New("empty IRI")
	}
	for _, r := range iri {
		if r >= '\x00' && r <= '\x20' {
			return IRI{}, fmt.Errorf("disallowed character: %q", r)
		}
		switch r {
		case '<', '>', '"', '{', '}', '|', '^', '`', '\\':
			return IRI{}, fmt.Errorf("disallowed character: %q", r)
		}
	}
	return IRI{str: iri}, nil
}

// Literal represents a RDF literal; a value with a datatype and
// (optionally) an associated language tag for strings.
type Literal struct {
	// The literal is always stored as a string, regardless of datatype.
	str string

	// Val represents the typed value of a RDF Literal, boxed in an empty interface.
	// A type assertion is needed to get the value in the corresponding Go type.
	val interface{}

	// lang, if not empty, represents the language tag of a string.
	// A language tagged string has the datatype: rdf:langString.
	lang string

	// The datatype of the Literal.
	DataType IRI
}

// Serialize returns a string representation of a Literal.
func (l Literal) Serialize(f Format) string {
	if TermsEqual(l.DataType, rdfLangString) {
		return fmt.Sprintf("\"%s\"@%s", escapeLiteral(l.str), l.Lang())
	}
	if l.DataType != xsdString {
		switch f {
		case formatInternal:
			return l.str
		case NTriples, NQuads:
			return fmt.Sprintf("\"%s\"^^%s", escapeLiteral(l.str), l.DataType.Serialize(f))
		case Turtle:
			switch l.DataType {
			case xsdInteger, xsdDecimal, xsdBoolean, xsdDouble:
				return l.str
			case xsdDateTime:
				return fmt.Sprintf("\"%s\"^^%s", l.str, l.DataType.Serialize(f))
			default:
				return fmt.Sprintf("\"%s\"^^%s", escapeLiteral(l.str), l.DataType.Serialize(f))
			}
		default:
			panic("TODO")
		}
	}
	return fmt.Sprintf("\"%s\"", escapeLiteral(l.str))
}

// Type returns the TermType of a Literal.
func (l Literal) Type() TermType {
	return TermLiteral
}

// Lang returns the language of a language-tagged string.
func (l Literal) Lang() string {
	return l.lang
}

// String returns the literal string.
func (l Literal) String() string {
	return l.str
}

// Typed tries to parse the Literal's value into a Go type, acordig to the
// the DataType.
func (l Literal) Typed() (interface{}, error) {
	if l.val == nil {
		switch l.DataType.str {
		case xsdInteger.str, xsdInt.str:
			i, err := strconv.Atoi(l.str)
			if err != nil {
				return nil, err
			}
			l.val = i
			return i, nil
		case xsdDouble.str, xsdDecimal.str:
			f, err := strconv.ParseFloat(l.str, 64)
			if err != nil {
				return nil, err
			}
			l.val = f
			return f, nil
		case xsdBoolean.str:
			b, err := strconv.ParseBool(l.str)
			if err != nil {
				return nil, err
			}
			l.val = b
			return b, nil
		case xsdByte.str:
			return []byte(l.str), nil
			// TODO xsdDateTime etc
		default:
			return l.str, nil
		}
	}
	return l.val, nil
}

// validAsObject denotes that a Literal is valid as a Triple's Object.
func (l Literal) validAsObject() {}

// NewLiteral returns a new Literal, or an error on invalid input. It tries
// to map the given Go values to a corresponding xsd datatype.
func NewLiteral(v interface{}) (Literal, error) {
	switch t := v.(type) {
	case bool:
		return Literal{val: t, str: fmt.Sprintf("%v", t), DataType: xsdBoolean}, nil
	case int, int32, int64:
		return Literal{val: t, str: fmt.Sprintf("%v", t), DataType: xsdInteger}, nil
	case string:
		return Literal{str: t, DataType: xsdString}, nil
	case float32, float64:
		return Literal{val: t, str: fmt.Sprintf("%v", t), DataType: xsdDouble}, nil
	case time.Time:
		return Literal{val: t, str: t.Format(DateFormat), DataType: xsdDateTime}, nil
	case []byte:
		return Literal{val: t, str: string(t), DataType: xsdByte}, nil
	default:
		return Literal{}, fmt.Errorf("cannot infer XSD datatype from %#v", t)
	}
}

// NewLangLiteral creates a RDF literal with a given language tag, or fails
// if the language tag is not well-formed.
//
// The literal will have the datatype IRI xsd:String.
func NewLangLiteral(v, lang string) (Literal, error) {
	afterDash := false
	if len(lang) >= 1 && lang[0] == '-' {
		return Literal{}, errors.New("invalid language tag: must start with a letter")
	}
	for _, r := range lang {
		switch {
		case (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'):
			continue
		case r == '-':
			if afterDash {
				return Literal{}, errors.New("invalid language tag: only one '-' allowed")
			}
			afterDash = true
		case r >= '0' && r <= '9':
			if afterDash {
				continue
			}
			fallthrough
		default:
			return Literal{}, fmt.Errorf("invalid language tag: unexpected character: %q", r)
		}
	}
	if lang[len(lang)-1] == '-' {
		return Literal{}, errors.New("invalid language tag: trailing '-' disallowed")
	}
	return Literal{str: v, lang: lang, DataType: rdfLangString}, nil
}

// NewTypedLiteral returns a literal with the given datatype.
func NewTypedLiteral(v string, dt IRI) Literal {
	return Literal{str: v, DataType: dt}
}

// Subject interface distiguishes which Terms are valid as a Subject of a Triple.
type Subject interface {
	Term
	validAsSubject()
}

// Predicate interface distiguishes which Terms are valid as a Predicate of a Triple.
type Predicate interface {
	Term
	validAsPredicate()
}

// Object interface distiguishes which Terms are valid as a Object of a Triple.
type Object interface {
	Term
	validAsObject()
}

// Context interface distiguishes which Terms are valid as a Quad's Context.
// Incidently, this is the same as Terms valid as a Subject of a Triple.
type Context interface {
	Term
	validAsSubject()
}

// Triple represents a RDF triple.
type Triple struct {
	Subj Subject
	Pred Predicate
	Obj  Object
}

// Serialize returns a string representation of a Triple in the specified format.
//
// However, it will only serialize the triple itself, and not include the prefix directives.
// For a full serialization including directives, use the TripleEncoder.
func (t Triple) Serialize(f Format) string {
	var s, o string
	switch term := t.Subj.(type) {
	case IRI:
		s = term.Serialize(f)
	case Blank:
		s = term.Serialize(f)
	}
	switch term := t.Obj.(type) {
	case IRI:
		o = term.Serialize(f)
	case Literal:
		o = term.Serialize(f)
	case Blank:
		o = term.Serialize(f)
	}
	return fmt.Sprintf(
		"%s %s %s .\n",
		s,
		t.Pred.(IRI).Serialize(f),
		o,
	)
}

// Quad represents a RDF Quad; a Triple plus the context in which it occurs.
type Quad struct {
	Triple
	Ctx Context
}

// Serialize serializes the Quad in the given format (assumed to be NQuads atm).
func (q Quad) Serialize(f Format) string {
	var s, o, g string
	switch term := q.Subj.(type) {
	case IRI:
		s = term.Serialize(f)
	case Blank:
		s = term.Serialize(f)
	}
	switch term := q.Obj.(type) {
	case IRI:
		o = term.Serialize(f)
	case Literal:
		o = term.Serialize(f)
	case Blank:
		o = term.Serialize(f)
	}
	switch term := q.Ctx.(type) {
	case IRI:
		g = term.Serialize(f)
	case Blank:
		g = term.Serialize(f)
	}
	return fmt.Sprintf(
		"%s %s %s %s .\n",
		s,
		q.Pred.(IRI).Serialize(f),
		o,
		g,
	)
}

// TermsEqual returns true if two Terms are equal, or false if they are not.
func TermsEqual(a, b Term) bool {
	if a.Type() != b.Type() {
		return false
	}
	return a.Serialize(formatInternal) == b.Serialize(formatInternal)
}

// TriplesEqual tests if two Triples are identical.
func TriplesEqual(a, b Triple) bool {
	return TermsEqual(a.Subj, b.Subj) && TermsEqual(a.Pred, b.Pred) && TermsEqual(a.Obj, b.Obj)
}

// QuadsEqual tests if two Quads are identical.
func QuadsEqual(a, b Quad) bool {
	return TermsEqual(a.Ctx, b.Ctx) && TriplesEqual(a.Triple, b.Triple)
}
