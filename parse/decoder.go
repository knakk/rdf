package parse

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"strconv"
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

	state      parseFn           // state of parser
	startState parseFn           // start state (parseOutsideTriple, or, parseListItem if in a recursive structure)
	base       string            // base (default IRI)
	bnodeN     int               // anonymous blank node counter
	g          rdf.Term          // default graph
	ns         map[string]string // map[prefix]namespace
	tokens     [3]token          // 3 token lookahead
	peekCount  int               // number of tokens peeked at (position in tokens)
	lineMode   bool              // true if parsing line-based formats (N-Triples and N-Quads)

	// stack to keep track of nested patterns (collections, blank node property lists)
	// the top of the stack is always the next triple to be emitted
	tripleStack []rdf.Triple
}

// NewTTLDecoder creates a Turtle decoder
func NewTTLDecoder(r io.Reader) *Decoder {
	d := Decoder{
		l:           newLexer(r),
		f:           formatTTL,
		ns:          make(map[string]string),
		tripleStack: make([]rdf.Triple, 1, 4),
		startState:  parseOutsideTriple,
	}
	return &d
}

// NewNTDecoder creates a N-Triples decoder
func NewNTDecoder(r io.Reader) *Decoder {
	d := Decoder{
		l:        newLexer(r),
		f:        formatNT,
		lineMode: true,
	}
	return &d
}

// NewNQDecoder creates a N-Quads decoder.
// defaultGraph must be ether a rdf.URI or rdf.Blank.
func NewNQDecoder(r io.Reader, defaultGraph rdf.Term) *Decoder {
	if _, ok := defaultGraph.(rdf.Literal); ok {
		panic("defaultGraph must be either an URI or Blank node")
	}
	return &Decoder{
		l:        newLexer(r),
		f:        formatNQ,
		g:        defaultGraph,
		lineMode: true,
	}
}

// Public decoder methods:

// DecodeTriple returns the next valid triple, or an error
func (d *Decoder) DecodeTriple() (rdf.Triple, error) {
	switch d.f {
	case formatNT:
		return d.parseNT()
	case formatTTL:
		return d.parseTTL()
	}

	return rdf.Triple{}, fmt.Errorf("can't decode triples in format %v", d.f)
}

// DecodeQuad returns the next valid quad, or an error
func (d *Decoder) DecodeQuad() (rdf.Quad, error) {
	return d.parseNQ()
}

// Private parsing helpers:

// curTriple returns a pointer to the next triple to be emitted (the top of tripleStack).
func (d *Decoder) curTriple() *rdf.Triple {
	return &d.tripleStack[len(d.tripleStack)-1]
}

// insertBeforeTop a triple to the position before top of the stack
func (d *Decoder) insertBeforeTop(t rdf.Triple) {
	d.tripleStack = append(d.tripleStack, rdf.Triple{})
	d.tripleStack[len(d.tripleStack)-1] = d.tripleStack[len(d.tripleStack)-2]
	d.tripleStack[len(d.tripleStack)-2] = t
}

// remove the top-1 triple from the stack
func (d *Decoder) removeBeforeTop() {
	if len(d.tripleStack) >= 2 {
		i := len(d.tripleStack) - 2
		d.tripleStack = d.tripleStack[:i+copy(d.tripleStack[i:], d.tripleStack[i+1:])]
	} else {
		println("removeBeforeTop called with len(d.tripleStack) < 2") // TODO panic?
	}
}

// next returns the next token, or the next non-EOL token if the
// parser is not in linemode.
func (d *Decoder) next() token {
	for {
		if d.peekCount > 0 {
			d.peekCount--
		} else {
			d.tokens[0] = d.l.nextToken()
		}
		if !d.lineMode && d.tokens[d.peekCount].typ == tokenEOL {
			continue
		}
		break
	}
	return d.tokens[d.peekCount]
}

// peek returns but does not consume the next token.
func (d *Decoder) peek() token {
	for {
		if d.peekCount > 0 {
			return d.tokens[d.peekCount-1]
		}
		d.peekCount = 1
		d.tokens[0] = d.l.nextToken()
		if !d.lineMode && d.tokens[0].typ == tokenEOL {
			continue
		}
		break
	}
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

// parseFn represents the state of the parser as a function that returns the next state.
type parseFn func(*Decoder) parseFn

// parseOutsideTriple parses entering a new triple, or after a triple (before punctuation)
func parseOutsideTriple(d *Decoder) parseFn {
	switch d.next().typ {
	case tokenSemicolon:
		// We have a full triple, entering a predicate list.
		// Push to stack a triple with current subject set.
		d.insertBeforeTop(rdf.Triple{
			Subj: d.curTriple().Subj,
		})
		return nil // emit current triple
	case tokenComma:
		// We have a full triple, entering a object list.
		// Push to stack a triple with current subject and predicate set.
		d.insertBeforeTop(rdf.Triple{
			Subj: d.curTriple().Subj,
			Pred: d.curTriple().Pred,
		})
		return nil // emit current triple
	case tokenPrefix:
		label := d.expect1As("prefix label", tokenPrefixLabel)
		uri := d.expect1As("prefix URI", tokenIRIAbs)
		d.ns[label.text] = uri.text
		d.expect1As("directive trailing dot", tokenDot)
	case tokenSparqlPrefix:
		label := d.expect1As("prefix label", tokenPrefixLabel)
		uri := d.expect1As("prefix URI", tokenIRIAbs)
		d.ns[label.text] = uri.text
	case tokenBase:
		uri := d.expect1As("base URI", tokenIRIAbs)
		d.base = uri.text
		d.expect1As("directive trailing dot", tokenDot)
	case tokenSparqlBase:
		uri := d.expect1As("base URI", tokenIRIAbs)
		d.base = uri.text
	case tokenPropertyListEnd:
		// If not entering a predicate list or object list, we must
		// discard the previous entry in the stack.
		// TODO but what if last object was end of list? then top-1 will be { bnode rdf:rest rdf:nil }
		switch d.peek().typ {
		case tokenIRIAbs, tokenIRIRel, tokenPrefixLabel:
			// property list is subject, continue with predicate and object after this triple is emitted
		case tokenSemicolon, tokenComma:
			d.removeBeforeTop()
			d.next() // consume ';'
			fmt.Printf("%v\n", d.tripleStack)
			// property list is object,
		case tokenPropertyListEnd:
			d.removeBeforeTop()
			return parseOutsideTriple
		case tokenDot:
			d.next() // consume '.'
		default:
			d.removeBeforeTop()
		}

		return nil
	case tokenCollectionEnd:
		// insert { bnode rdf:last rdf:nil } in the stack, so that it will be emitted after current triple.
		d.removeBeforeTop()
		d.insertBeforeTop(
			rdf.Triple{
				Subj: d.curTriple().Subj,
				Pred: rdf.URI{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
				Obj:  rdf.URI{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
			})
		return nil
	case tokenDot:
		return nil // emit triple
	default:
		d.backup()
		return parseTriple
	}
	return parseOutsideTriple
}

func parseTriple(d *Decoder) parseFn {
	return parseSubject
}

func parseSubject(d *Decoder) parseFn {
	if d.curTriple().Subj != nil {
		return parsePredicate
	}
	tok := d.next()
	switch tok.typ {
	case tokenIRIAbs:
		d.curTriple().Subj = rdf.URI{URI: tok.text}
	case tokenIRIRel:
		d.curTriple().Subj = rdf.URI{URI: d.base + tok.text}
		// TODO err if d.base == ""
	case tokenBNode:
		d.curTriple().Subj = rdf.Blank{ID: tok.text}
	case tokenAnonBNode:
		d.bnodeN++
		d.curTriple().Subj = rdf.Blank{ID: fmt.Sprintf("b%d", d.bnodeN)}
	case tokenPrefixLabel:
		ns, ok := d.ns[tok.text]
		if !ok {
			d.errorf("missing namespace for prefix: %s", tok.text)
		}
		suf := d.expect1As("IRI suffix", tokenIRISuffix)
		d.curTriple().Subj = rdf.URI{URI: ns + suf.text}
	case tokenPropertyListStart:
		d.bnodeN++
		d.curTriple().Subj = rdf.Blank{ID: fmt.Sprintf("b%d", d.bnodeN)}
		// Blank node is subject.
		// insert triple with subj of property list, to be restored when property list ends
		d.insertBeforeTop(rdf.Triple{Subj: d.curTriple().Subj})
	case tokenCollectionStart:
		if d.peek().typ == tokenCollectionEnd {
			d.curTriple().Subj = rdf.URI{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"}
			break
		}
		d.bnodeN++
		d.curTriple().Subj = rdf.Blank{ID: fmt.Sprintf("b%d", d.bnodeN)}
		d.curTriple().Pred = rdf.URI{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"}

		d.insertBeforeTop(
			rdf.Triple{
				Subj: d.curTriple().Subj,
			})
		return parseObject
	case tokenError:
		d.errorf("%d:%d: syntax error: %v", tok.line, tok.col, tok.text)
	default:
		d.errorf("unexpected %v as subject", tok.typ)
	}

	return parsePredicate
}

func parsePredicate(d *Decoder) parseFn {
	if d.curTriple().Pred != nil {
		return parseObject
	}
	tok := d.next()
	switch tok.typ {
	case tokenIRIAbs:
		d.curTriple().Pred = rdf.URI{URI: tok.text}
	case tokenIRIRel:
		d.curTriple().Pred = rdf.URI{URI: d.base + tok.text}
		// TODO err if d.base == ""
	case tokenRDFType:
		d.curTriple().Pred = rdf.URI{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"}
	case tokenBNode:
		d.curTriple().Pred = rdf.Blank{ID: tok.text}
	case tokenPrefixLabel:
		ns, ok := d.ns[tok.text]
		if !ok {
			d.errorf("missing namespace for prefix: %s", tok.text)
		}
		suf := d.expect1As("IRI suffix", tokenIRISuffix)
		d.curTriple().Pred = rdf.URI{URI: ns + suf.text}
	case tokenError:
		d.errorf("syntax error: %v", tok.text)
	default:
		d.errorf("%d:%d: unexpected %v as predicate", tok.line, tok.col, tok.typ)
	}

	//	d.tripleStack[len(d.tripleStack)-1].Pred = d.curTriple().Pred

	return parseObject
}

func parseObject(d *Decoder) parseFn {
	if d.curTriple().Obj != nil {
		return nil
	}
	tok := d.next()
	switch tok.typ {
	case tokenIRIAbs:
		d.curTriple().Obj = rdf.URI{URI: tok.text}
	case tokenIRIRel:
		d.curTriple().Obj = rdf.URI{URI: d.base + tok.text}
		// TODO err if d.base == ""
	case tokenBNode:
		d.curTriple().Obj = rdf.Blank{ID: tok.text}
	case tokenAnonBNode:
		d.bnodeN++
		d.curTriple().Obj = rdf.Blank{ID: fmt.Sprintf("b%d", d.bnodeN)}
	case tokenLiteral, tokenLiteral3:
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
		d.curTriple().Obj = l
	case tokenLiteralDecimal:
		// we can ignore the error, because we know it's an correctly lexed decimal value:
		f, _ := strconv.ParseFloat(tok.text, 64) // TODO math/bigINt?
		d.curTriple().Obj = rdf.Literal{
			Val:      f,
			DataType: rdf.XSDDecimal,
		}
	case tokenLiteralInteger:
		// we can ignore the error, because we know it's an correctly lexed integer value:
		i, _ := strconv.Atoi(tok.text)
		d.curTriple().Obj = rdf.Literal{
			Val:      i,
			DataType: rdf.XSDInteger,
		}
	case tokenPrefixLabel:
		ns, ok := d.ns[tok.text]
		if !ok {
			d.errorf("missing namespace for prefix: %s", tok.text)
		}
		suf := d.expect1As("IRI suffix", tokenIRISuffix)
		d.curTriple().Obj = rdf.URI{URI: ns + suf.text}
	case tokenPropertyListStart:
		d.bnodeN++
		d.curTriple().Obj = rdf.Blank{ID: fmt.Sprintf("b%d", d.bnodeN)}
		d.insertBeforeTop(rdf.Triple{Subj: d.curTriple().Subj})
		d.insertBeforeTop(rdf.Triple{Subj: d.curTriple().Obj})
		return nil
	case tokenCollectionStart:
		if d.peek().typ == tokenCollectionEnd {
			d.next() // consume ')'
			d.curTriple().Obj = rdf.URI{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"}
			break
		}
		d.bnodeN++
		d.curTriple().Obj = rdf.Blank{ID: fmt.Sprintf("b%d", d.bnodeN)}

		d.insertBeforeTop(
			rdf.Triple{
				Subj: d.curTriple().Obj,
				Pred: rdf.URI{URI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			})
		return nil
	case tokenError:
		d.errorf("%d:%d: syntax error: %v", tok.line, tok.col, tok.text)
	default:
		d.errorf("%d:%d: unexpected %v as object", tok.line, tok.col, tok.typ)
	}

	return d.startState //parseOutsideTriple | parseListItem
}

func parseCollectionItem(d *Decoder) parseFn {
	switch d.next().typ {
	case tokenCollectionEnd:

	}
	return nil
}

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
func (d *Decoder) parseNT() (t rdf.Triple, err error) {
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

	if d.peek().typ == tokenEOF {
		// drain lexer
		d.next()
	}

	return t, err
}

// parseNQ parses a line of N-Quads and returns a valid quad or an error.
func (d *Decoder) parseNQ() (q rdf.Quad, err error) {
	defer d.recover(&err)

	for d.peek().typ == tokenEOL {
		d.next()
	}
	if d.peek().typ == tokenEOF {
		return q, io.EOF
	}

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

	if d.peek().typ == tokenEOF {
		// drain lexer
		d.next()
	}
	return q, err
}

// parseTTL parses a Turtle document, and returns the first available triple.
func (d *Decoder) parseTTL() (t rdf.Triple, err error) {
	defer d.recover(&err)
	if d.next().typ == tokenEOF {
		//fmt.Printf("final stack: %v\n", d.tripleStack)
		return t, io.EOF
	}
	d.backup()

	if len(d.tripleStack) > 1 {
		// Remove the last emitted triple from stack.
		d.tripleStack = d.tripleStack[:len(d.tripleStack)-1]
	} else {
		// Clear the only triple on stack.
		d.tripleStack[0] = rdf.Triple{}
	}

	//fmt.Printf("start stack: %v\n", d.tripleStack)
	for d.state = d.startState; d.state != nil; {
		d.state = d.state(d)
	}
	t = *d.curTriple()
	//fmt.Printf("emit: %v\n\n", t)
	return t, err
}

// Helper functions:

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
