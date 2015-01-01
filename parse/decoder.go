package parse

import (
	"bufio"
	"bytes"
	"errors"
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
	return d.parseNT(line)
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
		d.next()
	case tokenLiteral:
		lit, err := d.parseLiteral(false)
		if err != nil {
			return t, err
		}
		t.Obj = lit
		if d.cur.typ != tokenDot {
			// TODO document why, or find a better solution
			d.next()
		}
	case tokenIRIAbs:
		t.Obj = rdf.URI{URI: d.cur.text}
		d.next()
	}

	// parse final dot (d.next() called in ojbect parse to check for langtag, datatype)
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
	q := rdf.Quad{}

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
		d.next()
	case tokenLiteral:
		lit, err := d.parseLiteral(false)
		if err != nil {
			return q, err
		}
		q.Obj = lit
		if d.cur.typ != tokenDot {
			// TODO document why, or find a better solution
			d.next()
		}
	case tokenIRIAbs:
		q.Obj = rdf.URI{URI: d.cur.text}
		d.next()
	}

	// parse graph (optional) or final dot (d.next() called in ojbect parse to check for langtag, datatype).
	if err := d.expect(tokenDot, tokenIRIAbs, tokenBNode); err != nil {
		return q, err
	}
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
func (d *Decoder) parseLiteral(relIRI bool) (rdf.Literal, error) {
	if d.cur.typ != tokenLiteral {
		panic("interal parse error: parseLiteral() expects current token to be a tokenLiteral")
	}
	l := rdf.Literal{}
	l.Val = d.cur.text
	l.DataType = rdf.XSDString
	d.next() // need to see if literal has language tag or datatype URI
	switch d.cur.typ {
	case tokenLang:
		l.Lang = d.cur.text
		return l, nil
	case tokenDataTypeRel:
		if !relIRI {
			return l, errors.New("Literal data type IRI must be absolute")
		}
		fallthrough
	case tokenDataTypeAbs:
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
			// TODO: other xsd dataypes that maps to Go data types
		}
		return l, nil
	case tokenEOL:
		return l, fmt.Errorf("%d:%d: unexpected end of line", d.cur.line, d.cur.col)
	case tokenError:
		return l, fmt.Errorf("%d:%d: syntax error: %s", d.cur.line, d.cur.col, d.cur.text)
	default:
		// literal not follwed by language tag or datatype
		return l, nil
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
		return fmt.Errorf("%d:%d: expected %s, got %s", d.cur.line, d.cur.col, tt[0], d.cur.typ)
	}
	var types = make([]string, 0, len(tt))
	for _, t := range tt {
		types = append(types, fmt.Sprintf("%s", t))
	}
	return fmt.Errorf("%d:%d: expected %s, got %s", d.cur.line, d.cur.col, strings.Join(types, " / "), tt[0])
}
