package rdf

import (
	"fmt"
	"io"
	"runtime"
)

// ntDecoder is a N-Triples parser.
type ntDecoder struct {
	l         *lexer   // Turtle lexer (N-Triples is a subset of Turtle)
	tokens    [2]token // 2 token lookahead
	peekCount int      // Number of tokens peeked at (position in tokens lookahead array)
}

// newNTDecoder returns a new N-Triples parser on the given io.Reader.
func newNTDecoder(r io.Reader) *ntDecoder {
	return &ntDecoder{l: newLineLexer(r)}
}

// Decode parses a N-Triples document and returns the next valid Triple or an error.
func (d *ntDecoder) Decode() (t Triple, err error) {
	defer d.recover(&err)

again:
	for d.peek().typ == tokenEOL {
		d.next()
		goto again
	}
	if d.peek().typ == tokenEOF {
		return t, io.EOF
	}

	// parse triple subject
	tok := d.expectAs("subject", tokenIRIAbs, tokenBNode)
	if tok.typ == tokenIRIAbs {
		t.Subj = IRI{str: tok.text}
	} else {
		t.Subj = Blank{id: tok.text}
	}

	// parse triple predicate
	tok = d.expect1As("predicate", tokenIRIAbs)
	t.Pred = IRI{str: tok.text}

	// parse triple object
	tok = d.expectAs("object", tokenIRIAbs, tokenBNode, tokenLiteral)

	switch tok.typ {
	case tokenBNode:
		t.Obj = Blank{id: tok.text}
	case tokenLiteral:
		val := tok.text
		l := Literal{
			str:      val,
			DataType: xsdString,
		}
		p := d.peek()
		switch p.typ {
		case tokenLangMarker:
			d.next() // consume peeked token
			tok = d.expect1As("literal language", tokenLang)
			l.lang = tok.text
			l.DataType = rdfLangString
		case tokenDataTypeMarker:
			d.next() // consume peeked token
			tok = d.expect1As("literal datatype", tokenIRIAbs)
			l.DataType = IRI{str: tok.text}
		}
		t.Obj = l
	case tokenIRIAbs:
		t.Obj = IRI{str: tok.text}
	}

	// parse final dot
	d.expect1As("dot (.)", tokenDot)

	// check for extra tokens, assert we reached end of line
	d.expect1As("end of line", tokenEOL)

	if d.peek().typ == tokenEOF {
		// drain lexer of final EOF token
		d.next()
	}

	return t, err
}

// DecodeAll parses a compete N-Triples document and returns the valid triples,
// or an error.
func (d *ntDecoder) DecodeAll() ([]Triple, error) {
	var ts []Triple
	for t, err := d.Decode(); err != io.EOF; t, err = d.Decode() {
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}

// SetOption sets a ParseOption to the give value
func (d *ntDecoder) SetOption(o ParseOption, v interface{}) error {
	switch o {
	default:
		return fmt.Errorf("N-Triples decoder doesn't support option: %v", o)
	}
}

// Parsing functions:

// next returns the next token.
func (d *ntDecoder) next() token {
	if d.peekCount > 0 {
		d.peekCount--
	} else {
		d.tokens[0] = d.l.nextToken()
	}

	return d.tokens[d.peekCount]
}

// peek returns but does not consume the next token.
func (d *ntDecoder) peek() token {
	if d.peekCount > 0 {
		return d.tokens[d.peekCount-1]
	}
	d.peekCount = 1
	d.tokens[0] = d.l.nextToken()
	return d.tokens[0]
}

// recover catches non-runtime panics and binds the panic error
// to the given error pointer.
func (d *ntDecoder) recover(errp *error) {
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

// errorf formats the error and terminates parsing.
func (d *ntDecoder) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("%s", format)
	panic(fmt.Errorf(format, args...))
}

// unexpected complains about the given token and terminates parsing.
func (d *ntDecoder) unexpected(t token, context string) {
	d.errorf("%d:%d unexpected %v as %s", t.line, t.col, t.typ, context)
}

// expect1As consumes the next token and guarantees that it has the expected type.
func (d *ntDecoder) expect1As(context string, expected tokenType) token {
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
func (d *ntDecoder) expectAs(context string, expected ...tokenType) token {
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
