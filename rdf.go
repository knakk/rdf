// Package rdf introduces data structures for representing RDF resources,
// and includes functions for parsing and serialization of RDF data.
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
	ErrIRIEmptyInput        = errors.New("IRI cannot be an empty string")
	ErrIRIInvalidCharacters = errors.New(`IRI cannot contain space or any of the charaters: <>{}|\^'"`)
)

// DateFormat defines the string representation of xsd:DateTime values. You can override
// it if you need another layout.
var DateFormat = time.RFC3339

// The XML schema built-in datatypes (xsd):
// https://dvcs.w3.org/hg/rdf/raw-file/default/rdf-concepts/index.html#xsd-datatypes
var (
	// Core types:                                                    // Corresponding Go datatype:

	xsdString  = IRI{IRI: "http://www.w3.org/2001/XMLSchema#string"}  // string
	xsdBoolean = IRI{IRI: "http://www.w3.org/2001/XMLSchema#boolean"} // bool
	xsdDecimal = IRI{IRI: "http://www.w3.org/2001/XMLSchema#decimal"} // float64
	xsdInteger = IRI{IRI: "http://www.w3.org/2001/XMLSchema#integer"} // int

	// IEEE floating-point numbers:

	xsdDouble = IRI{IRI: "http://www.w3.org/2001/XMLSchema#double"} // float64
	xsdFloat  = IRI{IRI: "http://www.w3.org/2001/XMLSchema#float"}  // float64

	// Time and date:

	xsdDate          = IRI{IRI: "http://www.w3.org/2001/XMLSchema#date"}
	xsdTime          = IRI{IRI: "http://www.w3.org/2001/XMLSchema#time"}
	xsdDateTime      = IRI{IRI: "http://www.w3.org/2001/XMLSchema#dateTime"}
	xsdDateTimeStamp = IRI{IRI: "http://www.w3.org/2001/XMLSchema#dateTimeStamp"}

	// Recurring and partial dates:

	xsdYear              = IRI{IRI: "http://www.w3.org/2001/XMLSchema#gYear"}
	xsdMonth             = IRI{IRI: "http://www.w3.org/2001/XMLSchema#gMonth"}
	xsdDay               = IRI{IRI: "http://www.w3.org/2001/XMLSchema#gDay"}
	xsdYearMonth         = IRI{IRI: "http://www.w3.org/2001/XMLSchema#gYearMonth"}
	xsdDuration          = IRI{IRI: "http://www.w3.org/2001/XMLSchema#Duration"}
	xsdYearMonthDuration = IRI{IRI: "http://www.w3.org/2001/XMLSchema#yearMonthDuration"}
	xsdDayTimeDuration   = IRI{IRI: "http://www.w3.org/2001/XMLSchema#dayTimeDuration"}

	// Limited-range integer numbers

	xsdByte = IRI{IRI: "http://www.w3.org/2001/XMLSchema#byte"} // []byte
)

// Term is the interface for the RDF term types: blank node, literal and IRI.
type Term interface {
	// String returns a string representation of the Term.
	String() string

	// Value returns the typed value of a RDF term, boxed in an empty interface.
	// For IRIs and Blank nodes this would return the uri and blank label as strings.
	Value() interface{}

	// Type returns the RDF term type.
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

// Blank represents a RDF blank node; an unqualified IRI with an ID.
type Blank struct {
	ID string
}

// Value returns the string label of the blank node, without the '_:' prefix.
func (b Blank) Value() interface{} {
	return b.ID
}

// String returns the string representation of a blank node.
func (b Blank) String() string {
	return "_:" + b.ID
}

// Type returns the TermType of a blank node.
func (b Blank) Type() TermType {
	return TermBlank
}

// NewBlank returns a new blank node with a given ID. It returns
// an error only if the supplied ID is blank.
func NewBlank(id string) (Blank, error) {
	if len(strings.TrimSpace(id)) == 0 {
		return Blank{}, ErrBlankNodeMissingID
	}
	return Blank{ID: id}, nil
}

// IRI represents a RDF IRI resource.
type IRI struct {
	IRI string
}

// String returns the string representation of an IRI.
func (u IRI) String() string {
	return "<" + u.IRI + ">"
}

// Value returns the IRI as a string, without the enclosing <>.
func (u IRI) Value() interface{} {
	return u.IRI
}

// Type returns the TermType of a IRI.
func (u IRI) Type() TermType {
	return TermIRI
}

// NewIRI returns a new IRI, or an error if it's not valid.
func NewIRI(uri string) (IRI, error) {
	if len(strings.TrimSpace(uri)) == 0 {
		return IRI{}, ErrIRIEmptyInput
	}
	for _, r := range uri {
		switch r {
		case '<', '>', '"', '{', '}', '|', '^', '`', '\\':
			return IRI{}, ErrIRIInvalidCharacters
		}
	}
	return IRI{IRI: uri}, nil
}

// Literal represents a RDF literal; a value with a datatype and
// (optionally) an associated language tag for strings.
//
// So called untyped literals are given the datatype xsd:string, so in practice
// they are not untyped anymore. This is according to the RDF1.1 spec:
// http://www.w3.org/TR/2014/REC-rdf11-concepts-20140225/#section-Graph-Literal
type Literal struct {
	// Val represents the typed value of a RDF Literal, boxed in an empty interface.
	// A type assertion is needed to get the value in the corresponding Go type.
	Val interface{}

	// Lang, if not empty, represents the language tag of a string.
	Lang string

	// The datatype of the Literal.
	// TODO should be a pointer, to easily check for nil??
	DataType IRI
}

// String returns the string representation of a Literal.
func (l Literal) String() string {
	if l.Lang != "" {
		return fmt.Sprintf("\"%v\"@%s", l.Val, l.Lang)
	}
	if l.DataType.String() != "" {
		switch t := l.Val.(type) {
		case bool, int, float64:
			return fmt.Sprintf("%v", t)
		case string:
			if l.DataType == xsdString || l.DataType.String() == "" {
				return fmt.Sprintf("\"%v\"", t)
			}
			return fmt.Sprintf("\"%v\"^^%v", t, l.DataType)
		case time.Time:
			return fmt.Sprintf("\"%v\"^^%v", t.Format(DateFormat), l.DataType)
		default:
			return fmt.Sprintf("%v^^%v", t, l.DataType)
		}
	}
	return fmt.Sprintf("\"%v\"", l.Val)
}

// Value returns the string representation of an IRI.
func (l Literal) Value() interface{} {
	return l.Val
}

// Type returns the TermType of a Literal.
func (l Literal) Type() TermType {
	return TermLiteral
}

// NewLiteral returns a new Literal, or an error on invalid input. It tries
// to map the given Go values to a corresponding xsd datatype.
func NewLiteral(v interface{}) (Literal, error) {
	switch t := v.(type) {
	case bool:
		return Literal{Val: t, DataType: xsdBoolean}, nil
	case int:
		return Literal{Val: t, DataType: xsdInteger}, nil
	case string:
		return Literal{Val: t, DataType: xsdString}, nil
	case float64:
		return Literal{Val: t, DataType: xsdFloat}, nil
	case time.Time:
		return Literal{Val: t, DataType: xsdDateTime}, nil
	default:
		return Literal{}, fmt.Errorf("cannot infer xsd:datatype from %v", t)
	}
}

// NewLangLiteral creates a RDF literal with a givne language tag.
// No validation is performed to check if the language tag conforms
// to the BCP 47 spec: http://tools.ietf.org/html/bcp47
func NewLangLiteral(v, lang string) Literal {
	return Literal{Val: v, Lang: lang, DataType: xsdString}
}

// Triple represents a RDF triple.
type Triple struct {
	Subj, Pred, Obj Term
}

// NT returns a string representation of the triple in N-Triples format.
func (t Triple) NT() string {
	// TODO only xsd:string doesn't need datatype, all others do
	return fmt.Sprintf("%v %v %v .", t.Subj, t.Pred, t.Obj)
}

// Quad represents a RDF quad; that is, a triple with a named graph.
type Quad struct {
	Triple
	Graph Term // IRI or BNode (Literal not valid as graph)
}
