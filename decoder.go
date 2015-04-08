package rdf

import (
	"fmt"
	"io"
	"runtime"
)

// A ParseOption allows to customize the behaviour of a decoder.
type ParseOption int

// Options which can configure a decoder.
const (
	// Base IRI to resolve relative IRIs against (for formats that support
	// relative IRIs: Turtle, RDF/XML, TriG, JSON-LD)
	Base ParseOption = iota

	// Strict mode determines how the decoder responds to errors.
	// When true (the default), it will fail on any malformed input. When
	// false, it will try to continue parsing, discarding only the malformed
	// parts.
	// Strict

	// ErrOut
)

// TripleDecoder parses RDF documents (serializations of an RDF graph).
//
// For streaming parsing, use the Decode() method to decode a single Triple
// at a time. Or, if you want to read the whole document in one go, use DecodeAll().
//
// The decoder can be instructed with numerous options. Note that not all options
// are supported by all formats. Consult the following table:
//
//  Option      Description        Value      (default)       Format support
//  ------------------------------------------------------------------------------
//  Base        Base IRI           IRI        (empty IRI)     Turtle, RDF/XML
//  Strict      Strict mode        true/false (true)          TODO
//  ErrOut      Error output       io.Writer  (nil)           TODO
type TripleDecoder interface {
	// Decode parses a RDF document and return the next valid triple.
	// It returns io.EOF when the whole document is parsed.
	Decode() (Triple, error)

	// DecodeAll parses the entire RDF document and return all valid
	// triples, or an error.
	DecodeAll() ([]Triple, error)

	// SetOption sets a parsing option to the given value. Not all options
	// are supported by all serialization formats.
	SetOption(ParseOption, interface{}) error
}

// NewTripleDecoder returns a new TripleDecoder capable of parsing triples
// from the given io.Reader in the given serialization format.
func NewTripleDecoder(r io.Reader, f Format) TripleDecoder {
	switch f {
	case NTriples:
		return newNTDecoder(r)
	case RDFXML:
		return newRDFXMLDecoder(r)
	case Turtle:
		return newTTLDecoder(r)
	default:
		panic(fmt.Errorf("Decoder for serialization format %v not implemented", f))
	}
}

// QuadDecoder parses RDF quads in one of the following formats:
// N-Quads.
//
// For streaming parsing, use the Decode() method to decode a single Quad
// at a time. Or, if you want to read the whole source in one go, DecodeAll().
type QuadDecoder struct {
	l      *lexer
	format Format

	DefaultGraph Context  // default graph
	tokens       [3]token // 3 token lookahead
	peekCount    int      // number of tokens peeked at (position in tokens lookahead array)
}

// NewQuadDecoder returns a new QuadDecoder capable of parsing quads
// from the given io.Reader in the given serialization format.
func NewQuadDecoder(r io.Reader, f Format) *QuadDecoder {
	return &QuadDecoder{
		l:            newLineLexer(r),
		format:       f,
		DefaultGraph: Blank{id: "_:defaultGraph"},
	}
}

// Decode returns the next valid Quad, or an error
func (d *QuadDecoder) Decode() (Quad, error) {
	return d.parseNQ()
}

// DecodeAll decodes and returns all Quads from source, or an error
func (d *QuadDecoder) DecodeAll() ([]Quad, error) {
	var qs []Quad
	for q, err := d.Decode(); err != io.EOF; q, err = d.Decode() {
		if err != nil {
			return nil, err
		}
		qs = append(qs, q)
	}
	return qs, nil
}

// next returns the next token.
func (d *QuadDecoder) next() token {
	if d.peekCount > 0 {
		d.peekCount--
	} else {
		d.tokens[0] = d.l.nextToken()
	}

	return d.tokens[d.peekCount]
}

// peek returns but does not consume the next token.
func (d *QuadDecoder) peek() token {
	if d.peekCount > 0 {
		return d.tokens[d.peekCount-1]
	}
	d.peekCount = 1
	d.tokens[0] = d.l.nextToken()
	return d.tokens[0]
}

// recover catches non-runtime panics and binds the panic error
// to the given error pointer.
func (d *QuadDecoder) recover(errp *error) {
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

// expect1As consumes the next token and guarantees that it has the expected type.
func (d *QuadDecoder) expect1As(context string, expected tokenType) token {
	t := d.next()
	if t.typ != expected {
		if t.typ == tokenError {
			d.errorf("%d:%d: syntax error: %s", t.line, t.col, t.text)
		} else {
			d.unexpected(t, context)
		}
	}
	return t
}

// expectAs consumes the next token and guarantees that it has the one of the expected types.
func (d *QuadDecoder) expectAs(context string, expected ...tokenType) token {
	t := d.next()
	for _, e := range expected {
		if t.typ == e {
			return t
		}
	}
	if t.typ == tokenError {
		d.errorf("%d:%d: syntax error: %v", t.line, t.col, t.text)
	} else {
		d.unexpected(t, context)
	}
	return t
}

// errorf formats the error and terminates parsing.
func (d *QuadDecoder) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("%s", format)
	panic(fmt.Errorf(format, args...))
}

// unexpected complains about the given token and terminates parsing.
func (d *QuadDecoder) unexpected(t token, context string) {
	d.errorf("%d:%d unexpected %v as %s", t.line, t.col, t.typ, context)
}
