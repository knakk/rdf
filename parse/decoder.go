package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/knakk/rdf"
)

type format int

const (
	formatUnknown format = iota
	formatRDFXML
	formatTTL
	formatNT
	formatNQ
	formatTriG
)

// Decoder implements a Turtle/Trig parser
type Decoder struct {
	r *bufio.Reader
	l *lexer
	f format
	g rdf.Term

	cur, prev token
}

// NewTTLDecoder creates a Turtle decoder
func NewTTLDecoder(r io.Reader) *Decoder {
	return &Decoder{
		l: newLexer(),
		r: bufio.NewReader(r),
		f: formatTTL,
	}
}

// NewNTDecoder creates a N-Triples decoder
func NewNTDecoder(r io.Reader) *Decoder {
	return &Decoder{
		l: newLexer(),
		r: bufio.NewReader(r),
		f: formatNT,
	}
}

// NewNQDecoder creates a N-Quads decoder.
// defaultGraph must be ether a rdf.URI or rdf.Blank.
func NewNQDecoder(r io.Reader, defaultGraph rdf.Term) *Decoder {
	if _, ok := defaultGraph.(rdf.Literal); ok {
		panic("defaultGraph must be either an URI or Blank node")
	}
	return &Decoder{
		l: newLexer(),
		r: bufio.NewReader(r),
		f: formatNQ,
		g: defaultGraph,
	}
}

// DecodeTriple returns the next valid triple, or an error
func (d *Decoder) DecodeTriple() (rdf.Triple, error) {
	line, err := d.r.ReadBytes('\n')
	if err != nil && len(line) == 0 {
		d.l.stop() // reader drained, stop lexer
		return rdf.Triple{}, err
	}
	line = bytes.TrimSpace(line)
	if len(line) == 0 || bytes.HasPrefix(line, []byte("#")) {
		// skip empty lines or comment lines
		d.l.line++
		return d.DecodeTriple()
	}
	if d.f == formatNT {
		return d.parseNT(line)
	}
	return d.parseTTL(line)
}

// DecodeQuad returns the next valid quad, or an error
func (d *Decoder) DecodeQuad() (rdf.Quad, error) {
	line, err := d.r.ReadBytes('\n')
	if err != nil && len(line) == 0 {
		d.l.stop() // reader drained, stop lexer
		return rdf.Quad{}, err
	}
	line = bytes.TrimSpace(line)
	if len(line) == 0 || bytes.HasPrefix(line, []byte("#")) {
		// skip empty lines or comment lines
		d.l.line++
		return d.DecodeQuad()
	}
	return d.parseNQ(line)
}

func (d *Decoder) parseTTL(line []byte) (rdf.Triple, error) {
	d.cur.typ = tokenNone
	d.l.incoming <- line
	t := rdf.Triple{}

	// parse triple subject
	d.next()
	if err := d.expect(tokenIRIAbs, tokenBNode); err != nil {
		return t, err
	}
	if d.cur.typ == tokenIRIAbs {
		t.Subj = rdf.URI{URI: d.cur.text}
	} else {
		t.Subj = rdf.Blank{ID: d.cur.text}
	}

	// parse triple predicate
	d.next()
	if err := d.expect(tokenIRIAbs, tokenIRIRel, tokenBNode, tokenRDFType); err != nil {
		return t, err
	}
	switch d.cur.typ {
	case tokenIRIAbs, tokenIRIRel:
		t.Pred = rdf.URI{URI: d.cur.text}
	case tokenBNode:
		t.Pred = rdf.Blank{ID: d.cur.text}
	case tokenRDFType:
		t.Pred = rdf.URI{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"}
	}

	// parse triple object
	d.next()
	if err := d.expect(tokenIRIAbs, tokenBNode, tokenLiteral); err != nil {
		return t, err
	}

	switch d.cur.typ {
	case tokenBNode:
		t.Obj = rdf.Blank{ID: d.cur.text}
	case tokenLiteral:
		val := d.cur.text
		l := rdf.Literal{
			Val:      val,
			DataType: rdf.XSDString,
		}
		d.next()
		if err := d.expect(tokenLangMarker, tokenDataTypeMarker, tokenDot); err != nil {
			return t, err
		}
		switch d.cur.typ {
		case tokenDot:
			t.Obj = l
			goto consumedDot
		case tokenLangMarker:
			d.next()
			if err := d.expect(tokenLang); err != nil {
				return t, err
			}
			l.Lang = d.cur.text
		case tokenDataTypeMarker:
			d.next()
			if err := d.expect(tokenIRIAbs); err != nil {
				return t, err
			}
			v, err := parseLiteral(val, d.cur.text)
			if err == nil {
				l.Val = v
			}
			l.DataType = rdf.URI{URI: d.cur.text}
		}
		t.Obj = l
	case tokenIRIAbs:
		t.Obj = rdf.URI{URI: d.cur.text}
	}

	// parse final dot
	d.next()
consumedDot:
	if err := d.expect(tokenDot); err != nil {
		return t, err
	}

	// check for extra tokens, assert we reached end of line
	d.next()
	if d.cur.typ != tokenEOL {
		return t, fmt.Errorf("found extra token after end of statement: %q(%v)", d.cur.text, d.cur.typ)
	}

	return t, nil
}

func (d *Decoder) parseNT(line []byte) (rdf.Triple, error) {
	d.cur.typ = tokenNone
	d.l.incoming <- line
	t := rdf.Triple{}

	// parse triple subject
	d.next()
	if err := d.expect(tokenIRIAbs, tokenBNode); err != nil {
		return t, err
	}
	if d.cur.typ == tokenIRIAbs {
		t.Subj = rdf.URI{URI: d.cur.text}
	} else {
		t.Subj = rdf.Blank{ID: d.cur.text}
	}

	// parse triple predicate
	d.next()
	if err := d.expect(tokenIRIAbs, tokenBNode); err != nil {
		return t, err
	}
	if d.cur.typ == tokenIRIAbs {
		t.Pred = rdf.URI{URI: d.cur.text}
	} else {
		t.Pred = rdf.Blank{ID: d.cur.text}
	}

	// parse triple object
	d.next()
	if err := d.expect(tokenIRIAbs, tokenBNode, tokenLiteral); err != nil {
		return t, err
	}

	switch d.cur.typ {
	case tokenBNode:
		t.Obj = rdf.Blank{ID: d.cur.text}
	case tokenLiteral:
		val := d.cur.text
		l := rdf.Literal{
			Val:      val,
			DataType: rdf.XSDString,
		}
		d.next()
		if err := d.expect(tokenLangMarker, tokenDataTypeMarker, tokenDot); err != nil {
			return t, err
		}
		switch d.cur.typ {
		case tokenDot:
			t.Obj = l
			goto consumedDot
		case tokenLangMarker:
			d.next()
			if err := d.expect(tokenLang); err != nil {
				return t, err
			}
			l.Lang = d.cur.text
		case tokenDataTypeMarker:
			d.next()
			if err := d.expect(tokenIRIAbs); err != nil {
				return t, err
			}
			v, err := parseLiteral(val, d.cur.text)
			if err == nil {
				l.Val = v
			}
			l.DataType = rdf.URI{URI: d.cur.text}
		}
		t.Obj = l
	case tokenIRIAbs:
		t.Obj = rdf.URI{URI: d.cur.text}
	}

	// parse final dot
	d.next()
consumedDot:
	if err := d.expect(tokenDot); err != nil {
		return t, err
	}

	// check for extra tokens, assert we reached end of line
	d.next()
	if d.cur.typ != tokenEOL {
		return t, fmt.Errorf("found extra token after end of statement: %q(%v)", d.cur.text, d.cur.typ)
	}

	return t, nil
}

func (d *Decoder) parseNQ(line []byte) (rdf.Quad, error) {
	d.cur.typ = tokenNone
	d.l.incoming <- line
	q := rdf.Quad{Graph: d.g}

	// parse quad subject
	d.next()
	if err := d.expect(tokenIRIAbs, tokenBNode); err != nil {
		return q, err
	}
	if d.cur.typ == tokenIRIAbs {
		q.Subj = rdf.URI{URI: d.cur.text}
	} else {
		q.Subj = rdf.Blank{ID: d.cur.text}
	}

	// parse quad predicate
	d.next()
	if err := d.expect(tokenIRIAbs, tokenBNode); err != nil {
		return q, err
	}
	if d.cur.typ == tokenIRIAbs {
		q.Pred = rdf.URI{URI: d.cur.text}
	} else {
		q.Pred = rdf.Blank{ID: d.cur.text}
	}

	// parse quad object
	d.next()
	if err := d.expect(tokenIRIAbs, tokenBNode, tokenLiteral); err != nil {
		return q, err
	}

	switch d.cur.typ {
	case tokenBNode:
		q.Obj = rdf.Blank{ID: d.cur.text}
	case tokenLiteral:
		val := d.cur.text
		l := rdf.Literal{
			Val:      val,
			DataType: rdf.XSDString,
		}
		d.next()
		if err := d.expect(tokenLangMarker, tokenDataTypeMarker, tokenIRIAbs, tokenBNode, tokenDot); err != nil {
			return q, err
		}
		switch d.cur.typ {
		case tokenDot, tokenIRIAbs, tokenBNode:
			q.Obj = l
			goto consumedDot
		case tokenLangMarker:
			d.next()
			if err := d.expect(tokenLang); err != nil {
				return q, err
			}
			l.Lang = d.cur.text
		case tokenDataTypeMarker:
			d.next()
			if err := d.expect(tokenIRIAbs); err != nil {
				return q, err
			}
			v, err := parseLiteral(val, d.cur.text)
			if err == nil {
				l.Val = v
			}
			l.DataType = rdf.URI{URI: d.cur.text}
		}
		q.Obj = l
	case tokenIRIAbs:
		q.Obj = rdf.URI{URI: d.cur.text}
	}

	// parse graph (optional) or final dot (d.next() called in ojbect parse to check for langtag, datatype).
	d.next()
	if err := d.expect(tokenDot, tokenIRIAbs, tokenBNode); err != nil {
		return q, err
	}
consumedDot:
	switch d.cur.typ {
	case tokenDot: // do nothing
	case tokenBNode:
		q.Graph = rdf.Blank{ID: d.cur.text}
		d.next()
		if err := d.expect(tokenDot); err != nil {
			return q, err
		}
	case tokenIRIAbs:
		q.Graph = rdf.URI{URI: d.cur.text}
		d.next()
		if err := d.expect(tokenDot); err != nil {
			return q, err
		}
	}

	// check for extra tokens, assert we reached end of line
	d.next()
	if d.cur.typ != tokenEOL {
		return q, fmt.Errorf("found extra token after end of statement: %q(%v)", d.cur.text, d.cur.typ)
	}

	return q, nil
}

// parseLiteral
func parseLiteral(val, datatype string) (interface{}, error) {
	switch val {
	case rdf.XSDInteger.URI:
		i, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		return i, nil
	case rdf.XSDFloat.URI: // TODO also XSDDouble ?
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	case rdf.XSDBoolean.URI:
		bo, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		return bo, nil
	case rdf.XSDDateTime.URI:
		t, err := time.Parse(rdf.DateFormat, val)
		if err != nil {
			return nil, err
		}
		return t, nil
		// TODO: other xsd dataypes that maps to Go data types
	default:
		return nil, fmt.Errorf("don't know how to represent %q with datatype %q as a Go type", val, datatype)
	}
}

func (d *Decoder) next() bool {
	if d.cur.typ == tokenEOL || d.cur.typ == tokenError {
		return false
	}
	d.prev = d.cur
	d.cur = d.l.nextToken()
	if d.cur.typ == tokenEOL || d.cur.typ == tokenError {
		return false
	}
	return true
}

func (d *Decoder) expect(tt ...tokenType) error {
	if d.cur.typ == tokenEOL {
		return fmt.Errorf("%d:%d: unexpected end of line", d.cur.line, d.cur.col)
	}
	if d.cur.typ == tokenError {
		return fmt.Errorf("%d:%d: syntax error: %s", d.cur.line, d.cur.col, d.cur.text)
	}
	for i := range tt {
		if d.cur.typ == tt[i] {
			return nil
		}
	}
	if len(tt) == 1 {
		return fmt.Errorf("%d:%d: expected %v, got %s", d.cur.line, d.cur.col, tt[0], d.cur.typ)
	}
	var types = make([]string, 0, len(tt))
	for _, t := range tt {
		types = append(types, fmt.Sprintf("%s", t))
	}
	return fmt.Errorf("%d:%d: expected %v, got %s", d.cur.line, d.cur.col, strings.Join(types, " / "), d.cur.typ)
}
