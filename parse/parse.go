package parse

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/knakk/rdf"
)

// Decoder implements a Turtle/Trig parser
type Decoder struct {
	r      *bufio.Reader
	l      *lexer
	format string

	cur, prev token
	errors    []error
}

// NewNTDecoder creates a N-Triples decoder
func NewNTDecoder(r io.Reader) *Decoder {
	return &Decoder{
		l:      newLexer(),
		r:      bufio.NewReader(r),
		format: "N-Triples",
	}
}

// Parser helper functions

// isAbsoluteIRI checks if an IRI is absolute (i.e has a scheme).
// RFC 2396: scheme = alpha *( alpha | digit | "+" | "-" | "." )
func isAbsoluteIRI(s string) bool {
	iri := []rune(s)
	i := 1
	if isAlpha(iri[0]) {
		for _, r := range iri[1:] {
			if isAlphaOrDigit(r) || r == '+' || r == '-' || r == '.' {
				i++
				continue
			}
			break
		}
	}
	if i >= 1 && len(iri) > i && iri[i] == ':' {
		return true
	}
	return false
}

// Parser functions:

// Decode returns the next valid triple, or an error
func (d *Decoder) Decode() (rdf.Triple, error) {
	line, err := d.r.ReadBytes('\n')
	if err != nil && len(line) == 0 {
		d.l.stop() // reader drained, stop lexer
		return rdf.Triple{}, err
	}
	line = bytes.TrimSpace(line)
	if len(line) == 0 || bytes.HasPrefix(line, []byte("#")) {
		return d.Decode()
	}
	return d.parseNT(line)
}

func (d *Decoder) parseNT(line []byte) (rdf.Triple, error) {
	d.l.incoming <- line
	t := rdf.Triple{}

	// parse triple subject
	if !d.next() {
		// cannot be tokenEOL, given we check for len(line) and comment in line in Decode(),
		// so there must me something, at least tokenError
		return t, fmt.Errorf("%d:%d: %s", d.cur.line, d.cur.col, d.cur.text)
	}
	if !d.oneOf(tokenIRI, tokenBNode) {
		return t, fmt.Errorf("subject must be IRI or Blank node, got %v", d.cur.typ)
	}
	if d.cur.typ == tokenIRI {
		if d.format == "N-Triples" && !isAbsoluteIRI(d.cur.text) {
			return t, fmt.Errorf("%d:%d: relative IRIs not allowed", d.cur.line, d.cur.col)
		}
		t.Subj = rdf.URI{URI: d.cur.text}
	} else {
		t.Subj = rdf.Blank{ID: d.cur.text}
	}

	// parse triple predicate
	if !d.next() {
		if d.cur.typ == tokenEOL {
			return t, fmt.Errorf("%d:%d: unexpected end of line", d.cur.line, d.cur.col)
		}
		return t, fmt.Errorf("%d:%d: %s", d.cur.line, d.cur.col, d.cur.text)
	}
	if !d.oneOf(tokenIRI, tokenBNode) {
		return t, fmt.Errorf("predicate must be IRI or Blank node, got %v", d.cur.typ)
	}
	if d.cur.typ == tokenIRI {
		if d.format == "N-Triples" && !isAbsoluteIRI(d.cur.text) {
			return t, fmt.Errorf("%d:%d: relative IRIs not allowed", d.cur.line, d.cur.col)
		}
		t.Pred = rdf.URI{URI: d.cur.text}
	} else {
		t.Pred = rdf.Blank{ID: d.cur.text}
	}

	// parse triple object
	if !d.next() {
		if d.cur.typ == tokenEOL {
			return t, fmt.Errorf("%d:%d: unexpected end of line", d.cur.line, d.cur.col)
		}
		return t, fmt.Errorf("%d:%d: %s", d.cur.line, d.cur.col, d.cur.text)
	}
	if !d.oneOf(tokenIRI, tokenLiteral, tokenBNode) {
		return t, fmt.Errorf("expected IRI/Literal as object, got %v", d.cur.typ)
	}
	switch d.cur.typ {
	case tokenBNode:
		t.Obj = rdf.Blank{ID: d.cur.text}
		d.next()
	case tokenLiteral:
		lit, err := d.parseLiteral()
		if err != nil {
			return t, err
		}
		t.Obj = lit
		if d.cur.typ == tokenDot {
			return t, nil
		}
		d.next()
	case tokenIRI:
		if d.format == "N-Triples" && !isAbsoluteIRI(d.cur.text) {
			return t, fmt.Errorf("%d:%d: relative IRIs not allowed", d.cur.line, d.cur.col)
		}
		t.Obj = rdf.URI{URI: d.cur.text}
		d.next()
	}

	// parse final dot
	if d.cur.typ != tokenDot {
		return t, errors.New("missing '.' at end of triple statement")
	}

	// check for extra tokens, assert we reached end of line
	d.next()
	if d.cur.typ != tokenEOL {
		return t, fmt.Errorf("found extra token after end of statement: %q", d.cur.text)
	}

	return t, nil
}

// parseLiteral
func (d *Decoder) parseLiteral() (rdf.Literal, error) {
	if d.cur.typ != tokenLiteral {
		panic("interal parse error: parseLiteral() expects current token to be a tokenLiteral")
	}
	l := rdf.Literal{}
	l.Val = d.cur.text
	l.DataType = rdf.XSDString
	if d.next() {
		switch d.cur.typ {
		case tokenLang:
			l.Lang = d.cur.text
			return l, nil
		case tokenDataType:
			if d.format == "N-Triples" && !isAbsoluteIRI(d.cur.text) {
				return l, fmt.Errorf("%d:%d: relative IRIs not allowed", d.cur.line, d.cur.col)
			}
			l.DataType = rdf.URI{URI: d.cur.text}
			switch d.cur.text {
			case rdf.XSDInteger.URI:
				i, err := strconv.Atoi(d.prev.text)
				if err != nil {
					//TODO set datatype to xsd:string?
					return l, nil
				}
				l.Val = i
			case rdf.XSDFloat.URI: // TODO also XSDDouble ?
				f, err := strconv.ParseFloat(d.prev.text, 64)
				if err != nil {
					return l, nil
				}
				l.Val = f
			case rdf.XSDBoolean.URI:
				bo, err := strconv.ParseBool(d.prev.text)
				if err != nil {
					return l, nil
				}
				l.Val = bo
			case rdf.XSDDateTime.URI:
				t, err := time.Parse(rdf.DateFormat, d.prev.text)
				if err != nil {
					return l, nil
				}
				l.Val = t
				// TODO: other xsd dataypes
			}
			return l, nil
		default:
			// literal not follwed by language tag or datatype
			if d.cur.typ != tokenDot {
				d.backup()
			}
		}
	}
	return l, nil
}

func (d *Decoder) next() bool {
	if d.cur.typ != tokenEOL {
		d.prev = d.cur
	}
	d.cur = d.l.nextToken()
	if d.cur.typ == tokenEOL || d.cur.typ == tokenError {
		return false
	}
	return true
}

func (d *Decoder) backup() {
	d.cur = d.prev
}

func (d *Decoder) oneOf(tt ...tokenType) bool {
	for i := range tt {
		if d.cur.typ == tt[i] {
			return true
		}
	}
	return false
}
