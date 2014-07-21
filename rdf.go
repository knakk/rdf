// Package rdf introduces data structures and functions for creating and
// working with RDF resources.
//
// The main use case is representing data coming from or going to a
// triple/quad-store via the SPARQL protocol.
// The package will not include graph traversing or querying functions, as
// this is much more efficently handled by a SPARQL query engine.
package rdf

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Exported errors.
var (
	ErrBlankNodeMissingID   = errors.New("blank node cannot have an empty ID")
	ErrURIEmptyInput        = errors.New("URI cannot be an empty string")
	ErrURIInvalidCharacters = errors.New(`URI cannot contain space or any of the charaters: <>{}|\^'"`)
)

// DateFormat defines the string representation of xsd:DateTime values. You can override
// it if you need another layout.
var DateFormat = time.RFC3339

// The XML schema built-in datatypes (xsd). See here for documentation:
// https://dvcs.w3.org/hg/rdf/raw-file/default/rdf-concepts/index.html#xsd-datatypes
var (
	// Core types:
	XSDString  = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#string")
	XSDBoolean = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#boolean")
	XSDDecimal = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#decimal")
	XSDInteger = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#integer")

	// IEEE floating-point numbers:
	XSDDouble = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#double")
	XSDFloat  = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#float")

	// Time and date:
	XSDDate          = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#date")
	XSDTime          = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#time")
	XSDDateTime      = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#dateTime")
	XSDDateTimeStamp = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#dateTimeStamp")

	// Recurring and partial dates:
	XSDYear              = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#gYear")
	XSDMonth             = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#gMonth")
	XSDDay               = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#gDay")
	XSDYearMonth         = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#gYearMonth")
	XSDDuration          = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#Duration")
	XSDYearMonthDuration = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#yearMonthDuration")
	XSDDayTimeDuration   = NewURIUnsafe("http://www.w3.org/2001/XMLSchema#dayTimeDuration")

	// Limited-range integer numbers
	// TODO

	// Encoded binary data
	// TODO

	// Miscellaneous XSD types
	// TODO
)

// Term is the interface for the RDF term types: blank node, literal and URI.
type Term interface {
	// String should return the string representation of a RDF term, in a
	// form suitable for insertion into a SPARQL query.
	String() string

	// Eq tests for equality with another RDF term.
	Eq(other Term) bool

	// Type returns the RDF term type.
	Type() termType
}

type termType int

// Exported RDF term types.
const (
	TermBlank termType = iota
	TermURI
	TermLiteral
)

// Blank represents a RDF blank node; an unqualified URI with an ID.
type Blank struct {
	ID string
}

// String returns the string representation of a blank node.
func (b *Blank) String() string {
	return "_:" + b.ID
}

// Eq tests a blank node's equality with other RDF terms.
func (b *Blank) Eq(other Term) bool {
	if other.Type() != b.Type() {
		return false
	}
	return b.String() == other.String()
}

// Type returns the termType of a blank node.
func (b *Blank) Type() termType {
	return TermBlank
}

// NewBlank returns a new blank node with a given ID. It returns
// an error only if the supplied ID is blank.
func NewBlank(id string) (*Blank, error) {
	if len(strings.TrimSpace(id)) == 0 {
		return nil, ErrBlankNodeMissingID
	}
	return &Blank{ID: id}, nil
}

// NewBlankUnsafe is like NewBlank, except it doesn't fail on invalid input.
func NewBlankUnsafe(id string) *Blank {
	return &Blank{ID: id}
}

// URI represents a RDF URI resource.
//
// The URI term type is actially an IRI, meaning it can consist of non-latin
// characters as well.
type URI struct {
	URI string
}

// String returns the string representation of an URI.
func (u *URI) String() string {
	return "<" + u.URI + ">"
}

// Eq tests a URI's equality with other RDF terms.
func (u *URI) Eq(other Term) bool {
	if other.Type() != u.Type() {
		return false
	}
	return u.String() == other.String()
}

// Type returns the termType of a URI.
func (u *URI) Type() termType {
	return TermURI
}

// NewURI returns a new URI, or an error if it's not valid.
func NewURI(uri string) (*URI, error) {
	if len(strings.TrimSpace(uri)) == 0 {
		return nil, ErrURIEmptyInput
	}
	for _, r := range uri {
		switch r {
		case '<', '>', '"', '{', '}', '|', '^', '`', '\\':
			return nil, ErrURIInvalidCharacters
		}
	}
	return &URI{URI: uri}, nil
}

// NewURIUnsafe returns a new URI, with no validation performed on input.
func NewURIUnsafe(uri string) *URI {
	return &URI{uri}
}

// Literal represents a RDF literal; a value with a datatype and
// (optionally) an associated language tag for strings.
//
// So called untyped literals are given the datatype xsd:string, so in practice
// they are not untyped anymore. This is according to the RDF1.1 spec:
// http://www.w3.org/TR/2014/REC-rdf11-concepts-20140225/#section-Graph-Literal
type Literal struct {
	// Value represents the typed value of a RDF Literal, boxed in an empty interface.
	// A type assertion is needed to get the value in the corresponding Go type.
	Value interface{}

	// Lang, if not empty, represents the language tag of a string.
	Lang string

	// The datatype of the Literal.
	DataType *URI
}

// String returns the string representation of a Literal.
func (l *Literal) String() string {
	if l.Lang != "" {
		return fmt.Sprintf("\"%v\"@%s", l.Value, l.Lang)
	}
	if l.DataType != nil {
		switch t := l.Value.(type) {
		case bool, int, float64:
			return fmt.Sprintf("%v", t)
		case string:
			return fmt.Sprintf("\"%v\"", t)
		case time.Time:
			return fmt.Sprintf("\"%v\"^^%v", t.Format(DateFormat), l.DataType)
		default:
			return fmt.Sprintf("%v^^%v", t, l.DataType)
		}
	}
	return fmt.Sprintf("\"%v\"", l.Value)
}

// Eq tests a Literal's equality with other RDF terms.
func (l *Literal) Eq(other Term) bool {
	if other.Type() != l.Type() {
		return false
	}
	return l.String() == other.String()
}

// Type returns the termType of a Literal.
func (l *Literal) Type() termType {
	return TermLiteral
}

// NewLiteral returns a new Literal, or an error on invalid input. It tries
// to map the given Go values to a corresponding xsd datatype.
// If you need a custom datatype, you must create the literal with the normal
// struct syntax:
//    l := Literal{Val: "my-val", DataType: NewURIUnsafe("my uri")}
func NewLiteral(v interface{}) (*Literal, error) {
	switch t := v.(type) {
	case bool:
		return &Literal{Value: t, DataType: XSDBoolean}, nil
	case int:
		return &Literal{Value: t, DataType: XSDInteger}, nil
	case string:
		return &Literal{Value: t, DataType: XSDString}, nil
	case float64:
		return &Literal{Value: t, DataType: XSDFloat}, nil
	case time.Time:
		return &Literal{Value: t, DataType: XSDDateTime}, nil
	default:
		return &Literal{}, fmt.Errorf("cannot infer xsd:datatype from %v", t)
	}
}

// NewLiteralUnsafe returns a new literal without performing any validation
// on input. Any input on which type cannot be inferred, will be forced to xsd:string.
func NewLiteralUnsafe(v interface{}) *Literal {
	l, err := NewLiteral(v)
	if err != nil {
		l, _ = NewLiteral(fmt.Sprintf("%v", v))
	}
	return l
}

// NewLangLiteral creates a RDF literal with a givne language tag.
// No validation is performed to check if the language tag conforms
// to the BCP 47 spec: http://tools.ietf.org/html/bcp47
func NewLangLiteral(v, lang string) *Literal {
	return &Literal{Value: v, Lang: lang, DataType: XSDString}
}

// Triple represents a RDF triple.
type Triple struct {
	Subj, Pred, Obj Term
}

// Quad represents a RDF quad; that is, a triple with a named graph.
type Quad struct {
	Graph     URI
	Statement Triple
}
