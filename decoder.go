package rdf

import (
	"encoding/xml"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"time"
)

// TODO make TripleDecoder & QuadDecoder interfaces ?
// so we don't have to mix xml- & turtle-decoding state in one struct.

const rdfNS = `http://www.w3.org/1999/02/22-rdf-syntax-ns#`

type format int

// ctxTriple contains a Triple, plus the context in which the Triple appears.
type ctxTriple struct {
	Triple
	Ctx context
}

type context int

const (
	ctxTop context = iota
	ctxCollection
	ctxList
	ctxBag
)

// TODO remove when done
func (ctx context) String() string {
	switch ctx {
	case ctxTop:
		return "top context"
	case ctxList:
		return "list"
	case ctxCollection:
		return "collection"

	default:
		return "unknown context"
	}
}

// TripleDecoder parses RDF triples in one of the following formats:
// N-Triples, Turtle, RDF/XML.
//
// For streaming parsing, use the Decode() method to decode a single Triple
// at a time. Or, if you want to read the whole source in one go, DecodeAll().
type TripleDecoder struct {
	l      *lexer
	format Format

	state     parseFn           // state of parser
	Base      IRI               // base (default IRI)
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

// NewTripleDecoder returns a new TripleDecoder capable of parsing triples
// from the given io.Reader in the given serialization format.
func NewTripleDecoder(r io.Reader, f Format) *TripleDecoder {
	var l *lexer
	var x *xml.Decoder
	switch f {
	case FormatNT:
		l = newLineLexer(r)
	case FormatRDFXML:
		x = xml.NewDecoder(r)
	default:
		l = newLexer(r)
	}
	d := TripleDecoder{
		l:        l,
		xmlDec:   x,
		format:   f,
		ns:       make(map[string]string),
		ctxStack: make([]ctxTriple, 0, 8),
		triples:  make([]Triple, 0, 4),
		Base:     IRI{str: ""},
	}
	return &d
}

// Decode returns the next valid Triple, or an error.
func (d *TripleDecoder) Decode() (Triple, error) {
	switch d.format {
	case FormatNT:
		return d.parseNT()
	case FormatTTL:
		return d.parseTTL()
	case FormatRDFXML:
		return d.parseRDFXML()
	}

	return Triple{}, fmt.Errorf("can't decode triples in format %v", d.format)
}

// DecodeAll decodes and returns all Triples from source, or an error
func (d *TripleDecoder) DecodeAll() ([]Triple, error) {
	var ts []Triple
	for t, err := d.Decode(); err != io.EOF; t, err = d.Decode() {
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}

// pushContext pushes the current triple and context to the context stack.
func (d *TripleDecoder) pushContext() {
	d.ctxStack = append(d.ctxStack, d.current)
}

// popContext restores the next context on the stack as the current context.
// If allready at the topmost context, it clears the current triple.
func (d *TripleDecoder) popContext() {
	switch len(d.ctxStack) {
	case 0:
		d.current.Ctx = ctxTop
		d.current.Subj = nil
		d.current.Pred = nil
		d.current.Obj = nil
	case 1:
		d.current = d.ctxStack[0]
		d.ctxStack = d.ctxStack[:0]
	default:
		d.current = d.ctxStack[len(d.ctxStack)-1]
		d.ctxStack = d.ctxStack[:len(d.ctxStack)-1]
	}
}

// emit adds the current triple to the slice of completed triples.
func (d *TripleDecoder) emit() {
	d.triples = append(d.triples, d.current.Triple)
}

// next returns the next token.
func (d *TripleDecoder) next() token {
	if d.peekCount > 0 {
		d.peekCount--
	} else {
		d.tokens[0] = d.l.nextToken()
	}

	return d.tokens[d.peekCount]
}

// peek returns but does not consume the next token.
func (d *TripleDecoder) peek() token {
	if d.peekCount > 0 {
		return d.tokens[d.peekCount-1]
	}
	d.peekCount = 1
	d.tokens[0] = d.l.nextToken()
	return d.tokens[0]
}

// backup backs the input stream up one token.
func (d *TripleDecoder) backup() {
	d.peekCount++
}

// backup2 backs the input stream up two tokens.
func (d *TripleDecoder) backup2(t1 token) {
	d.tokens[1] = t1
	d.peekCount = 2
}

// backup3 backs the input stream up three tokens.
func (d *TripleDecoder) backup3(t2, t1 token) {
	d.tokens[1] = t1
	d.tokens[2] = t2
	d.peekCount = 3
}

// Parsing:

// parseFn represents the state of the parser as a function that returns the next state.
type parseFn func(*TripleDecoder) parseFn

// errorf formats the error and terminates parsing.
func (d *TripleDecoder) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("%s", format)
	panic(fmt.Errorf(format, args...))
}

// unexpected complains about the given token and terminates parsing.
func (d *TripleDecoder) unexpected(t token, context string) {
	d.errorf("%d:%d unexpected %v as %s", t.line, t.col, t.typ, context)
}

// recover catches non-runtime panics and binds the panic error
// to the given error pointer.
func (d *TripleDecoder) recover(errp *error) {
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
func (d *TripleDecoder) expect1As(context string, expected tokenType) token {
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
func (d *TripleDecoder) expectAs(context string, expected ...tokenType) token {
	t := d.next()
	for _, e := range expected {
		if t.typ == e {
			return t
		}
	}
	if t.typ == tokenError {
		d.errorf("syntax error: %s", t.text)
	} else {
		d.unexpected(t, context)
	}
	return t
}

// parseLiteral
func parseLiteral(val, datatype string) (interface{}, error) {
	switch datatype {
	case xsdString.str:
		return val, nil
	case xsdInteger.str:
		i, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		return i, nil
	case xsdFloat.str, xsdDouble.str, xsdDecimal.str:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	case xsdBoolean.str:
		bo, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		return bo, nil
	case xsdDateTime.str:
		t, err := time.Parse(DateFormat, val)
		if err != nil {
			// Unfortunately, xsd:dateTime allows dates without timezone information
			// Try parse again unspecified timzeone (defaulting to UTC)
			t, err = time.Parse("2006-01-02T15:04:05", val)
			if err != nil {
				return nil, err
			}
			return t, nil
		}
		return t, nil
	case xsdByte.str:
		return []byte(val), nil
		// TODO: other xsd dataypes that maps to Go data types
	default:
		return val, nil
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
