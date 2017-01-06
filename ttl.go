package rdf

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"time"
)

type ttlDecoder struct {
	l *lexer

	state     parseFn           // state of parser
	base      IRI               // base (default IRI)
	bnodeN    int               // anonymous blank node counter
	ns        map[string]string // map[prefix]namespace
	tokens    [3]token          // 3 token lookahead
	peekCount int               // number of tokens peeked at (position in tokens lookahead array)
	current   ctxTriple         // the current triple beeing parsed

	// ctxStack keeps track of current and parent triple contexts,
	// needed for parsing recursive structures (list/collections).
	ctxStack []ctxTriple

	// triples contains complete triples ready to be emitted. Usually it will have just one triple,
	// but can have more when parsing nested list/collections. Decode() will always return the first item.
	triples []Triple
}

func newTTLDecoder(r io.Reader) *ttlDecoder {
	return &ttlDecoder{
		l:        newLexer(r),
		ns:       make(map[string]string),
		ctxStack: make([]ctxTriple, 0, 8),
		triples:  make([]Triple, 0, 4),
	}
}

// SetOption sets a ParseOption to the give value
func (d *ttlDecoder) SetOption(o ParseOption, v interface{}) error {
	switch o {
	case Base:
		iri, ok := v.(IRI)
		if !ok {
			return fmt.Errorf("ParseOption \"Base\" must be an IRI.")
		}
		d.base = iri
	default:
		return fmt.Errorf("RDF/XML decoder doesn't support option: %v", o)
	}
	return nil
}

// Decode parses a Turtle document, and returns the next valid triple, or an error.
func (d *ttlDecoder) Decode() (t Triple, err error) {
	defer d.recover(&err)

	// Check if there is allready a triple in the pipeline:
	if len(d.triples) >= 1 {
		goto done
	}

	// Return io.EOF when there is no more tokens to parse.
	if d.next().typ == tokenEOF {
		return t, io.EOF
	}
	d.backup()

	// Run the parser state machine.
	for d.state = parseStart; d.state != nil; {
		d.state = d.state(d)
	}

	if len(d.triples) == 0 {
		// No triples to emit, i.e only comments and possibly directives was parsed.
		return t, io.EOF
	}

done:
	t = d.triples[0]
	d.triples = d.triples[1:]
	return t, err
}

// DecodeAll parses a compete Trutle document and returns the valid triples,
// or an error.
func (d *ttlDecoder) DecodeAll() ([]Triple, error) {
	var ts []Triple
	for t, err := d.Decode(); err != io.EOF; t, err = d.Decode() {
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}

// parseStart parses top context
func parseStart(d *ttlDecoder) parseFn {
	switch d.next().typ {
	case tokenPrefix:
		label := d.expect1As("prefix label", tokenPrefixLabel)
		if label.text == "" {
			println("empty label")
		}
		tok := d.expectAs("prefix IRI", tokenIRIAbs, tokenIRIRel)
		if tok.typ == tokenIRIRel {
			// Resolve against document base IRI
			d.ns[label.text] = d.base.str + tok.text
		} else {
			d.ns[label.text] = tok.text
		}
		d.expect1As("directive trailing dot", tokenDot)
	case tokenSparqlPrefix:
		label := d.expect1As("prefix label", tokenPrefixLabel)
		uri := d.expect1As("prefix IRI", tokenIRIAbs)
		d.ns[label.text] = uri.text
	case tokenBase:
		tok := d.expectAs("base IRI", tokenIRIAbs, tokenIRIRel)
		if tok.typ == tokenIRIRel {
			// Resolve against document base IRI
			d.base.str = d.base.str + tok.text
		} else {
			d.base.str = tok.text
		}
		d.expect1As("directive trailing dot", tokenDot)
	case tokenSparqlBase:
		uri := d.expect1As("base IRI", tokenIRIAbs)
		d.base.str = uri.text
	case tokenEOF:
		return nil
	default:
		d.backup()
		return parseTriple
	}
	return parseStart
}

// parseEnd parses punctuation [.,;\])] before emitting the current triple.
func parseEnd(d *ttlDecoder) parseFn {
	tok := d.next()
	switch tok.typ {
	case tokenSemicolon:
		switch d.peek().typ {
		case tokenSemicolon:
			// parse multiple semicolons in a row
			return parseEnd
		case tokenDot:
			// parse trailing semicolon
			return parseEnd
		case tokenEOF:
			// trailing semicolon without final dot not allowed
			// TODO only allowed in property lists?
			d.errorf("%d:%d: expected triple termination, got %v", tok.line, tok.col, tok.typ)
			return nil
		}
		d.current.Pred = nil
		d.current.Obj = nil
		d.pushContext()
		return nil
	case tokenComma:
		d.current.Obj = nil
		d.pushContext()
		return nil
	case tokenPropertyListEnd:
		d.popContext()
		if d.peek().typ == tokenDot {
			// Reached end of statement
			d.next()
			return nil
		}
		if d.current.Pred == nil {
			// Property list was subject, push context with subject to stack.
			d.pushContext()
			return nil
		}
		// Property list was object, need to check for more closing property lists.
		return parseEnd
	case tokenCollectionEnd:
		// Emit collection closing triple { bnode rdf:rest rdf:nil }
		d.current.Pred = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"}
		d.current.Obj = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"}
		d.emit()

		// Restore parent triple
		d.popContext()
		if d.current.Pred == nil {
			// Collection was subject, push context with subject to stack.
			d.pushContext()
			return nil
		}
		// Collection was object, need to check for more closing collection.
		return parseEnd
	case tokenDot:
		if d.current.Ctx == ctxColl {
			return parseEnd
		}
		return nil
	case tokenError:
		d.errorf("%d:%d: syntax error: %v", tok.line, tok.col, tok.text)
		return nil
	default:
		if d.current.Ctx == ctxColl {
			d.backup() // unread collection item, to be parsed on next iteration

			d.bnodeN++
			d.current.Pred = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"}
			d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
			d.emit()

			d.current.Subj = d.current.Obj.(Subject)
			d.current.Obj = nil
			d.current.Pred = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"}
			d.pushContext()
			return nil
		}
		d.errorf("%d:%d: expected triple termination, got %v", tok.line, tok.col, tok.typ)
		return nil
	}

}

func parseTriple(d *ttlDecoder) parseFn {
	return parseSubject
}

func parseSubject(d *ttlDecoder) parseFn {
	// restore triple context, or clear current
	d.popContext()

	if d.current.Subj != nil {
		return parsePredicate
	}
	tok := d.next()
	switch tok.typ {
	case tokenIRIAbs:
		d.current.Subj = IRI{str: tok.text}
	case tokenIRIRel:
		d.current.Subj = IRI{str: d.base.str + tok.text}
	case tokenBNode:
		d.current.Subj = Blank{id: tok.text}
	case tokenAnonBNode:
		d.bnodeN++
		d.current.Subj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
	case tokenPrefixLabel:
		ns, ok := d.ns[tok.text]
		if !ok {
			d.errorf("missing namespace for prefix: '%s'", tok.text)
		}
		suf := d.expect1As("IRI suffix", tokenIRISuffix)
		d.current.Subj = IRI{str: ns + suf.text}
	case tokenPropertyListStart:
		// Blank node is subject of a new triple
		d.bnodeN++
		d.current.Subj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
		d.pushContext() // Subj = bnode, top context
		d.current.Ctx = ctxList
	case tokenCollectionStart:
		if d.peek().typ == tokenCollectionEnd {
			// An empty collection
			d.current.Subj = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"}
			break
		}
		d.bnodeN++
		d.current.Subj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
		d.pushContext()
		d.current.Pred = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"}
		d.current.Ctx = ctxColl
		return parseObject
	case tokenError:
		d.errorf("%d:%d: syntax error: %v", tok.line, tok.col, tok.text)
	default:
		d.errorf("unexpected %v as subject", tok.typ)
	}

	return parsePredicate
}

func parsePredicate(d *ttlDecoder) parseFn {
	if d.current.Pred != nil {
		return parseObject
	}
	tok := d.next()
	switch tok.typ {
	case tokenIRIAbs:
		d.current.Pred = IRI{str: tok.text}
	case tokenIRIRel:
		d.current.Pred = IRI{str: d.base.str + tok.text}
	case tokenRDFType:
		d.current.Pred = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"}
	case tokenPrefixLabel:
		ns, ok := d.ns[tok.text]
		if !ok {
			d.errorf("missing namespace for prefix: '%s'", tok.text)
		}
		suf := d.expect1As("IRI suffix", tokenIRISuffix)
		d.current.Pred = IRI{str: ns + suf.text}
	case tokenError:
		d.errorf("%d:%d: syntax error: %v", tok.line, tok.col, tok.text)
	default:
		d.errorf("%d:%d: unexpected %v as predicate", tok.line, tok.col, tok.typ)
	}

	return parseObject
}

func parseObject(d *ttlDecoder) parseFn {
	tok := d.next()
	switch tok.typ {
	case tokenIRIAbs:
		d.current.Obj = IRI{str: tok.text}
	case tokenIRIRel:
		d.current.Obj = IRI{str: d.base.str + tok.text}
	case tokenBNode:
		d.current.Obj = Blank{id: tok.text}
	case tokenAnonBNode:
		d.bnodeN++
		d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
	case tokenLiteral, tokenLiteral3:
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
			tok = d.expectAs("literal datatype", tokenIRIAbs, tokenPrefixLabel)
			switch tok.typ {
			case tokenIRIAbs:
				l.DataType = IRI{str: tok.text}
			case tokenPrefixLabel:
				ns, ok := d.ns[tok.text]
				if !ok {
					d.errorf("missing namespace for prefix: '%s'", tok.text)
				}
				tok2 := d.expect1As("IRI suffix", tokenIRISuffix)
				l.DataType = IRI{str: ns + tok2.text}
			}
		}
		d.current.Obj = l
	case tokenLiteralDouble:
		d.current.Obj = Literal{
			str:      tok.text,
			DataType: xsdDouble,
		}
	case tokenLiteralDecimal:
		d.current.Obj = Literal{
			str:      tok.text,
			DataType: xsdDecimal,
		}
	case tokenLiteralInteger:
		d.current.Obj = Literal{
			str:      tok.text,
			DataType: xsdInteger,
		}
	case tokenLiteralBoolean:
		d.current.Obj = Literal{
			str:      tok.text,
			DataType: xsdBoolean,
		}
	case tokenPrefixLabel:
		ns, ok := d.ns[tok.text]
		if !ok {
			d.errorf("missing namespace for prefix: '%s'", tok.text)
		}
		suf := d.expect1As("IRI suffix", tokenIRISuffix)
		d.current.Obj = IRI{str: ns + suf.text}
	case tokenPropertyListStart:
		// Blank node is object of current triple
		// Save current context, to be restored after the list ends
		d.pushContext()

		d.bnodeN++
		d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
		d.emit()

		// Set blank node as subject of the next triple. Push to stack and return.
		d.current.Subj = d.current.Obj.(Subject)
		d.current.Pred = nil
		d.current.Obj = nil
		d.current.Ctx = ctxList
		d.pushContext()
		return nil
	case tokenCollectionStart:
		if d.peek().typ == tokenCollectionEnd {
			// an empty collection
			d.next() // consume ')'
			d.current.Obj = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"}
			break
		}
		// Blank node is object of current triple
		// Save current context, to be restored after the collection ends
		d.pushContext()

		d.bnodeN++
		d.current.Obj = Blank{id: fmt.Sprintf("_:b%d", d.bnodeN)}
		d.emit()
		d.current.Subj = d.current.Obj.(Subject)
		d.current.Pred = IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"}
		d.current.Obj = nil
		d.current.Ctx = ctxColl
		d.pushContext()
		return nil
	case tokenError:
		d.errorf("%d:%d: syntax error: %v", tok.line, tok.col, tok.text)
	default:
		d.errorf("%d:%d: unexpected %v as object", tok.line, tok.col, tok.typ)
	}

	// We now have a full tripe, emit it.
	d.emit()

	return parseEnd
}

// pushContext pushes the current triple and context to the context stack.
func (d *ttlDecoder) pushContext() {
	d.ctxStack = append(d.ctxStack, d.current)
}

// popContext restores the next context on the stack as the current context.
// If allready at the topmost context, it clears the current triple.
func (d *ttlDecoder) popContext() {
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
func (d *ttlDecoder) emit() {
	d.triples = append(d.triples, d.current.Triple)
}

// next returns the next token.
func (d *ttlDecoder) next() token {
	if d.peekCount > 0 {
		d.peekCount--
	} else {
		d.tokens[0] = d.l.nextToken()
	}

	return d.tokens[d.peekCount]
}

// peek returns but does not consume the next token.
func (d *ttlDecoder) peek() token {
	if d.peekCount > 0 {
		return d.tokens[d.peekCount-1]
	}
	d.peekCount = 1
	d.tokens[0] = d.l.nextToken()
	return d.tokens[0]
}

// backup backs the input stream up one token.
func (d *ttlDecoder) backup() {
	d.peekCount++
}

// backup2 backs the input stream up two tokens.
func (d *ttlDecoder) backup2(t1 token) {
	d.tokens[1] = t1
	d.peekCount = 2
}

// backup3 backs the input stream up three tokens.
func (d *ttlDecoder) backup3(t2, t1 token) {
	d.tokens[1] = t1
	d.tokens[2] = t2
	d.peekCount = 3
}

// Parsing:

// parseFn represents the state of the parser as a function that returns the next state.
type parseFn func(*ttlDecoder) parseFn

// errorf formats the error and terminates parsing.
func (d *ttlDecoder) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("%s", format)
	panic(fmt.Errorf(format, args...))
}

// unexpected complains about the given token and terminates parsing.
func (d *ttlDecoder) unexpected(t token, context string) {
	d.errorf("%d:%d unexpected %v as %s", t.line, t.col, t.typ, context)
}

// recover catches non-runtime panics and binds the panic error
// to the given error pointer.
func (d *ttlDecoder) recover(errp *error) {
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
func (d *ttlDecoder) expect1As(context string, expected tokenType) token {
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
func (d *ttlDecoder) expectAs(context string, expected ...tokenType) token {
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

// ctxTriple contains a Triple, plus the context in which the Triple appears.
type ctxTriple struct {
	Triple
	Ctx context
}

type context int

const (
	ctxTop context = iota
	ctxColl
	ctxList
)

// TODO remove when done
func (ctx context) String() string {
	switch ctx {
	case ctxTop:
		return "top context"
	case ctxList:
		return "list"
	case ctxColl:
		return "collection"

	default:
		return "unknown context"
	}
}
