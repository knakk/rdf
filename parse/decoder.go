package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"runtime"
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

func (f format) String() string {
	switch f {
	case formatRDFXML:
		return "RDF/XML"
	case formatTTL:
		return "Turtle"
	case formatNT:
		return "N-Triples"
	case formatNQ:
		return "N-Quads"
	case formatTriG:
		return "Tri-G"
	default:
		return "Unknown format"
	}
}

// Decoder implements a Turtle/Trig parser
type Decoder struct {
	r *bufio.Reader
	l *lexer
	f format

	base        string            // base (default IRI)
	bnodeN      int               // anonymous blank node counter
	g           rdf.Term          // default graph
	ns          map[string]string // map[prefix]namespace
	cur, prev   token             // current and previous lexed tokens
	tokens      [3]token          // 3 token lookahead
	peekCount   int               // number of tokens peeked at (position in tokens)
	curSubj     rdf.Term          // current subject
	curPred     rdf.Term          // current predicate
	tripleStack []rdf.Triple      // stack of parent triples, needed for nested property lists
}

// NewTTLDecoder creates a Turtle decoder
func NewTTLDecoder(r io.Reader) *Decoder {
	return &Decoder{
		l:  newLexer(),
		r:  bufio.NewReader(r),
		f:  formatTTL,
		ns: make(map[string]string),
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

// Public decoder methods:

// DecodeTriple returns the next valid triple, or an error
func (d *Decoder) DecodeTriple() (rdf.Triple, error) {
	line, err := d.r.ReadBytes('\n')
	if err != nil && len(line) == 0 {
		d.l.stop() // reader drained, stop lexer
		return rdf.Triple{}, err
	}
	line = bytes.TrimSpace(line)
	if len(line) == 0 || line[0] == '#' {
		// skip empty lines or comment lines
		d.l.line++
		return d.DecodeTriple()
	}
	if d.f == formatNT {
		return d.parseNT(line)
	}
	if line[0] == '@' || line[0] == 'P' || line[0] == 'B' {
		// parse @prefix / @base / PREFIX / BASE
		d.cur.typ = tokenNone
		d.l.incoming <- line
		d.l.line++
		err := d.parseDirectives()
		if err != nil {
			return rdf.Triple{}, err
		}
		return d.DecodeTriple()
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
	if len(line) == 0 || line[0] == '#' {
		// skip empty lines or comment lines
		d.l.line++
		return d.DecodeQuad()
	}
	return d.parseNQ(line)
}

// Private parsing helpers:

// next returns the next token.
func (d *Decoder) next() token {
	if d.peekCount > 0 {
		d.peekCount--
	} else {
		d.tokens[0] = d.l.nextToken()
	}
	return d.tokens[d.peekCount]
}

// peek returns but does not consume the next token.
func (d *Decoder) peek() token {
	if d.peekCount > 0 {
		return d.tokens[d.peekCount-1]
	}
	d.peekCount = 1
	d.tokens[0] = d.l.nextToken()
	return d.tokens[0]
}

// backup backs the input stream up one token.
func (d *Decoder) backup() {
	d.peekCount++
}

// backup2 backs the input stream up two tokens.
func (d *Decoder) backup2(t1 token) {
	d.tokens[1] = t1
	d.peekCount = 2
}

// backup3 backs the input stream up three tokens.
func (d *Decoder) backup3(t2, t1 token) {
	d.tokens[1] = t1
	d.tokens[2] = t2
	d.peekCount = 3
}

// Parsing:

// errorf formats the error and terminates parsing.
func (d *Decoder) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("%s", format)
	panic(fmt.Errorf(format, args...))
}

// unexpected complains about the given token and terminates parsing.
func (d *Decoder) unexpected(t token, context string) {
	d.errorf("%d:%d unexpected %v as %s", t.line, t.col, t.typ, context)
}

// recover catches non-runtime panics and binds the panic error
// to the given error pointer.
func (d *Decoder) recover(errp *error) {
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
func (d *Decoder) expect1As(context string, expected tokenType) token {
	t := d.next()
	if t.typ != expected {
		if t.typ == tokenError {
			d.errorf("syntax error: %s", t.text)
		} else {
			d.unexpected(t, context)
		}
	}
	return t
}

// expectAs consumes the next token and guarantees that it has the one of the expected types.
func (d *Decoder) expectAs(context string, expected ...tokenType) token {
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

// parseNT parses a line of N-Triples and returns a valid triple or an error.
func (d *Decoder) parseNT(line []byte) (t rdf.Triple, err error) {
	defer d.recover(&err)
	d.l.incoming <- line

	// parse triple subject
	tok := d.expectAs("subject", tokenIRIAbs, tokenBNode)
	if tok.typ == tokenIRIAbs {
		t.Subj = rdf.URI{URI: tok.text}
	} else {
		t.Subj = rdf.Blank{ID: tok.text}
	}

	// parse triple predicate
	tok = d.expectAs("predicate", tokenIRIAbs, tokenBNode)
	if tok.typ == tokenIRIAbs {
		t.Pred = rdf.URI{URI: tok.text}
	} else {
		t.Pred = rdf.Blank{ID: tok.text}
	}

	// parse triple object
	tok = d.expectAs("object", tokenIRIAbs, tokenBNode, tokenLiteral)

	switch tok.typ {
	case tokenBNode:
		t.Obj = rdf.Blank{ID: tok.text}
	case tokenLiteral:
		val := tok.text
		l := rdf.Literal{
			Val:      val,
			DataType: rdf.XSDString,
		}
		p := d.peek()
		switch p.typ {
		case tokenLangMarker:
			d.next() // consume peeked token
			tok = d.expect1As("literal language", tokenLang)
			l.Lang = tok.text
		case tokenDataTypeMarker:
			d.next() // consume peeked token
			tok = d.expect1As("literal datatype", tokenIRIAbs)
			v, err := parseLiteral(val, tok.text)
			if err == nil {
				l.Val = v
			}
			l.DataType = rdf.URI{URI: tok.text}
		}
		t.Obj = l
	case tokenIRIAbs:
		t.Obj = rdf.URI{URI: tok.text}
	}

	// parse final dot
	d.expect1As("dot (.)", tokenDot)

	// check for extra tokens, assert we reached end of line
	d.expect1As("end of line", tokenEOL)

	return t, err
}

// parseNQ parses a line of N-Quads and returns a valid quad or an error.
func (d *Decoder) parseNQ(line []byte) (q rdf.Quad, err error) {
	defer d.recover(&err)
	d.l.incoming <- line

	// Set Quad graph to default graph
	q.Graph = d.g

	// parse quad subject
	tok := d.expectAs("subject", tokenIRIAbs, tokenBNode)
	if tok.typ == tokenIRIAbs {
		q.Subj = rdf.URI{URI: tok.text}
	} else {
		q.Subj = rdf.Blank{ID: tok.text}
	}

	// parse quad predicate
	tok = d.expectAs("predicate", tokenIRIAbs, tokenBNode)
	if tok.typ == tokenIRIAbs {
		q.Pred = rdf.URI{URI: tok.text}
	} else {
		q.Pred = rdf.Blank{ID: tok.text}
	}

	// parse quad object
	tok = d.expectAs("object", tokenIRIAbs, tokenBNode, tokenLiteral)

	switch tok.typ {
	case tokenBNode:
		q.Obj = rdf.Blank{ID: tok.text}
	case tokenLiteral:
		val := tok.text
		l := rdf.Literal{
			Val:      val,
			DataType: rdf.XSDString,
		}
		p := d.peek()
		switch p.typ {
		case tokenLangMarker:
			d.next() // consume peeked token
			tok = d.expect1As("literal language", tokenLang)
			l.Lang = tok.text
		case tokenDataTypeMarker:
			d.next() // consume peeked token
			tok = d.expect1As("literal datatype", tokenIRIAbs)
			v, err := parseLiteral(val, tok.text)
			if err == nil {
				l.Val = v
			}
			l.DataType = rdf.URI{URI: tok.text}
		}
		q.Obj = l
	case tokenIRIAbs:
		q.Obj = rdf.URI{URI: tok.text}
	}

	// parse optional graph
	p := d.peek()
	switch p.typ {
	case tokenIRIAbs:
		tok = d.next() // consume peeked token
		q.Graph = rdf.URI{URI: tok.text}
	case tokenBNode:
		tok = d.next() // consume peeked token
		q.Graph = rdf.Blank{ID: tok.text}
	case tokenDot:
		break
	default:
		d.expectAs("graph", tokenIRIAbs, tokenBNode)
	}

	// parse final dot
	d.expect1As("dot (.)", tokenDot)

	// check for extra tokens, assert we reached end of line
	d.expect1As("end of line", tokenEOL)

	return q, err
}

// *******************************
func (d *Decoder) parseDirectives() error {
	d.next()
	if err := d.expect(tokenPrefix, tokenSparqlPrefix, tokenBase, tokenSparqlBase); err != nil {
		return err
	}
	t := d.cur.typ
	switch t {
	case tokenPrefix, tokenSparqlPrefix:
		d.next()
		if err := d.expect(tokenPrefixLabel); err != nil {
			return err
		}
		d.next()
		if err := d.expect(tokenIRIAbs); err != nil {
			return err
		}

		// store namespace prefix
		d.ns[d.prev.text] = d.cur.text

		if t == tokenPrefix {
			// @prefix directives end in '.', but not SPARQL PREFIX directive
			d.next()
			if err := d.expect(tokenDot); err != nil {
				return err
			}
		}

		d.next()
		if d.cur.typ != tokenEOL {
			return fmt.Errorf("%d:%d: illegal token after end of directive: %s", d.cur.line, d.cur.col, d.cur.text)
		}
	case tokenBase, tokenSparqlBase:
		d.next()
		if err := d.expect(tokenIRIAbs); err != nil {
			return err
		}
		d.base = d.cur.text

		if t == tokenBase {
			// @base directives end in '.', but not SPARQL BAE directive
			d.next()
			if err := d.expect(tokenDot); err != nil {
				return err
			}
		}

		d.next()
		if d.cur.typ != tokenEOL {
			return fmt.Errorf("%d:%d: illegal token after end of directive: %s", d.cur.line, d.cur.col, d.cur.text)
		}
	}
	return nil
}

func (d *Decoder) _ttlParseSubj() (rdf.Term, error) {
	d.next()
	if err := d.expect(tokenIRIAbs, tokenIRIRel, tokenBNode, tokenAnonBNode, tokenPropertyListStart, tokenPrefixLabel); err != nil {
		return nil, err
	}
	switch d.cur.typ {
	case tokenIRIAbs:
		return rdf.URI{URI: d.cur.text}, nil
	case tokenIRIRel:
		return rdf.URI{URI: d.base + d.cur.text}, nil
		// TODO err if no base
	case tokenBNode:
		return rdf.Blank{ID: d.cur.text}, nil
	case tokenAnonBNode:
		d.bnodeN++
		return rdf.Blank{ID: fmt.Sprintf("b%d", d.bnodeN)}, nil
	case tokenPropertyListStart:
		d.bnodeN++
		d.curSubj = rdf.Blank{ID: fmt.Sprintf("b%d", d.bnodeN)}
		return d.curSubj, nil
	case tokenPrefixLabel:
		ns, ok := d.ns[d.cur.text]
		if !ok {
			return nil, fmt.Errorf("missing namespace for prefix: %s", d.cur.text)
		}
		d.next()
		if err := d.expect(tokenIRISuffix); err != nil {
			return nil, err
		}
		return rdf.URI{URI: ns + d.cur.text}, nil
	}
	panic("unreachable")
}

func (d *Decoder) _ttlParsePred() (rdf.Term, error) {
	d.next()
	if err := d.expect(tokenIRIAbs, tokenIRIRel, tokenBNode, tokenRDFType, tokenPrefixLabel); err != nil {
		return nil, err
	}
	switch d.cur.typ {
	case tokenIRIAbs:
		return rdf.URI{URI: d.cur.text}, nil
	case tokenIRIRel:
		return rdf.URI{URI: d.base + d.cur.text}, nil
		// TODO err if no base
	case tokenBNode:
		return rdf.Blank{ID: d.cur.text}, nil
	case tokenRDFType:
		return rdf.URI{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"}, nil
	case tokenPrefixLabel:
		ns, ok := d.ns[d.cur.text]
		if !ok {
			return nil, fmt.Errorf("missing namespace for prefix: %s", d.cur.text)
		}
		d.next()
		if err := d.expect(tokenIRISuffix); err != nil {
			return nil, err
		}
		return rdf.URI{URI: ns + d.cur.text}, nil
	}
	panic("unreachable")
}

func (d *Decoder) _ttlParseObj() (rdf.Term, error) {
	d.next()
	if err := d.expect(tokenIRIAbs, tokenIRIRel, tokenBNode, tokenAnonBNode, tokenPropertyListStart, tokenLiteral, tokenPrefixLabel); err != nil {
		return nil, err
	}

	switch d.cur.typ {
	case tokenBNode:
		d.next() // becasue case tokenLiteral consumes (at least) two tokens
		return rdf.Blank{ID: d.prev.text}, nil
	case tokenAnonBNode:
		d.bnodeN++
		d.next() // becasue case tokenLiteral consumes (at least) two tokens
		return rdf.Blank{ID: fmt.Sprintf("b%d", d.bnodeN)}, nil
	case tokenPropertyListStart:
		d.bnodeN++
		d.curSubj = rdf.Blank{ID: fmt.Sprintf("b%d", d.bnodeN)}
		return d.curSubj, nil
	case tokenLiteral:
		val := d.cur.text
		l := rdf.Literal{
			Val:      val,
			DataType: rdf.XSDString,
		}
		d.next()
		if err := d.expect(tokenLangMarker, tokenDataTypeMarker, tokenDot); err != nil {
			return nil, err
		}
		switch d.cur.typ {
		case tokenDot:
			return l, nil
		case tokenLangMarker:
			d.next()
			if err := d.expect(tokenLang); err != nil {
				return nil, err
			}
			l.Lang = d.cur.text
		case tokenDataTypeMarker:
			d.next()
			if err := d.expect(tokenIRIAbs); err != nil {
				return nil, err
			}
			v, err := parseLiteral(val, d.cur.text)
			if err == nil {
				l.Val = v
			}
			l.DataType = rdf.URI{URI: d.cur.text}
		}
		return l, nil
	case tokenIRIAbs:
		d.next() // becasue case tokenLiteral consumes (at least) two tokens
		return rdf.URI{URI: d.prev.text}, nil
	case tokenIRIRel:
		// TODO err if no base
		d.next() // becasue case tokenLiteral consumes (at least) two tokens
		return rdf.URI{URI: d.base + d.prev.text}, nil
	case tokenPrefixLabel:
		ns, ok := d.ns[d.cur.text]
		if !ok {
			return nil, fmt.Errorf("missing namespace for prefix: %s", d.cur.text)
		}
		d.next()
		if err := d.expect(tokenIRISuffix); err != nil {
			return nil, err
		}
		d.next() // becasue case tokenLiteral consumes (at least) two tokens
		return rdf.URI{URI: ns + d.prev.text}, nil
	}
	panic("unreachable")
}

func (d *Decoder) parseTTL(line []byte) (rdf.Triple, error) {
	d.l.incoming <- line
	d.cur.typ = tokenNone
	t := rdf.Triple{}
	var term rdf.Term
	var err error

	if d.curSubj != nil {
		// we are in a blankNodePropertyList
		t.Subj = d.curSubj
	} else {
		// parse triple subject
		term, err := d._ttlParseSubj()
		if err != nil {
			return t, err
		}
		t.Subj = term
	}

	if d.curPred != nil {
		// we are in a predicateObjectList
		t.Subj = d.curSubj
	} else {
		// parse triple predicate
		term, err = d._ttlParsePred()
		if err != nil {
			return t, err
		}
		t.Pred = term
	}

	// parse triple object
	term, err = d._ttlParseObj()
	if err != nil {
		return t, err
	}
	t.Obj = term

	// d.next() called in _ttlParseObj()
	// parse final dot, or end of proerty list
	if err := d.expect(tokenDot, tokenPropertyListEnd, tokenPropertyListStart); err != nil {
		return t, err
	}
	switch d.cur.typ {
	case tokenPropertyListEnd:
		d.next()
		if d.cur.typ == tokenDot {
			break
		}
		return t, nil
	case tokenDot:
		d.curSubj = nil
		d.curPred = nil
	case tokenPropertyListStart:
		d.next()
		fmt.Printf("curSubj: %v\n", d.curSubj)
		fmt.Printf("curPred: %v\n", d.curPred)
		fmt.Printf("%v %v", d.cur.typ, d.cur.text)
		return t, nil
	}

	// check for extra tokens, assert we reached end of line
	d.next()
	if d.cur.typ != tokenEOL {
		return t, fmt.Errorf("found extra token after end of statement: %q(%v)", d.cur.text, d.cur.typ)
	}

	return t, nil
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
		return fmt.Errorf("%d:%d: expected %v, got %v", d.cur.line, d.cur.col, tt[0], d.cur.typ)
	}
	var types = make([]string, 0, len(tt))
	for _, t := range tt {
		types = append(types, fmt.Sprintf("%v", t))
	}
	return fmt.Errorf("%d:%d: expected %v, got %v", d.cur.line, d.cur.col, strings.Join(types, " / "), d.cur.typ)
}
