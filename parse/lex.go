package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type tokenType int

const (
	// special tokens
	tokenEOF   tokenType = iota // end of input
	tokenEOL                    // end of line
	tokenError                  // an illegal token

	// turtle tokens
	tokenIRIAbs            // RDF IRI reference (absolute)
	tokenIRIRel            // RDF IRI reference (relative)
	tokenBNode             // RDF blank node
	tokenLiteral           // RDF literal
	tokenLiteral3          // RDF literal (triple-quoted string)
	tokenLiteralInteger    // RDF literal (integer)
	tokenLiteralDouble     // RDF literal (double-precision floating point)
	tokenLiteralDecimal    // RDF literal (arbritary-precision decimal)
	tokenLangMarker        // '@''
	tokenLang              // literal language tag
	tokenDataTypeMarker    // '^^'
	tokenDot               // '.'
	tokenSemicolon         // ';'
	tokenComma             // ','
	tokenRDFType           // 'a' => <http://www.w3.org/1999/02/22-rdf-syntax-ns#type>
	tokenPrefix            // @prefix
	tokenPrefixLabel       // @prefix tokenPrefixLabel: IRI
	tokenIRISuffix         // prefixLabel:IRISuffix
	tokenBase              // Base marker
	tokenSparqlPrefix      // PREFIX
	tokenSparqlBase        // BASE
	tokenAnonBNode         // [ws*]
	tokenPropertyListStart // '['
	tokenPropertyListEnd   // ']'
	tokenCollectionStart   // '('
	tokenCollectionEnd     // ')'
)

const eof = -1

// Rune helper values and functions:
var (
	hex            = []byte("0123456789ABCDEFabcdef")
	pnLocalEsc     = [...]rune{'_', '~', '.', '-', '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', '/', '?', '#', '@', '%'}
	badIRIRunes    = [...]rune{' ', '<', '"', '{', '}', '|', '^', '`'}
	okAfterRDFType = [...]rune{' ', '\t', '<', '"', '\''}
	pnTab          = []rune{
		'A', 'Z',
		'a', 'z',
		0x00C0, 0x00D6,
		0x00D8, 0x00F6,
		0x00F8, 0x02FF,
		0x0370, 0x037D,
		0x037F, 0x1FFF,
		0x200C, 0x200D,
		0x2070, 0x218F,
		0x2C00, 0x2FEF,
		0x3001, 0xD7FF,
		0xF900, 0xFDCF,
		0xFDF0, 0xFFFD,
		0x10000, 0xEFFFF, // last of PN_CHARS_BASE
		'_', '_',
		':', ':', // last of PN_CHARS_U
		'-', '-',
		'0', '9',
		0x00B7, 0x00B7,
		0x0300, 0x036F,
		0x203F, 0x2040, // last of PN_CHARS
	}
	plTab = []rune{
		'A', 'Z',
		'a', 'z',
		0x00C0, 0x00D6,
		0x00D8, 0x00F6,
		0x00F8, 0x02FF,
		0x0370, 0x037D,
		0x037F, 0x1FFF,
		0x200C, 0x200D,
		0x2070, 0x218F,
		0x2C00, 0x2FEF,
		0x3001, 0xD7FF,
		0xF900, 0xFDCF,
		0xFDF0, 0xFFFD,
		0x10000, 0xEFFFF, // last of PN_CHARS_BASE
		'_', '_', // last of PN_CHARS_U
		'-', '-',
		'0', '9',
		0x00B7, 0x00B7,
		0x0300, 0x036F,
		0x203F, 0x2040, // last of PN_CHARS
		'%', '%',
		'\\', '\\', // last of PN_LOCAL first character
		':', ':',
		'.', '.', // last of PN_LOCAL (except last character)
	}
)

func isAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigit(r rune) bool {
	return (r >= '0' && r <= '9')
}

func isAlphaOrDigit(r rune) bool {
	return isAlpha(r) || isDigit(r)
}

func check(r rune, tab []rune) bool {
	for i := 0; i < len(tab); i += 2 {
		if r >= tab[i] && r <= tab[i+1] {
			return true
		}
	}
	return false
}

func isPnCharsBase(r rune) bool {
	return check(r, pnTab[:2*14])
}

func isPnCharsU(r rune) bool {
	return check(r, pnTab[:2*16])
}

func isPnChars(r rune) bool {
	return check(r, pnTab)
}

func isPnLocalFirst(r rune) bool {
	return check(r, plTab[:2*22])
}

func isPnLocalMid(r rune) bool {
	return check(r, plTab)
}

// token represents a token emitted by the lexer.
type token struct {
	typ  tokenType // type of token
	line int       // line number
	col  int       // column number (NB measured in bytes, not runes)
	text string    // the value of the token
}

// stateFn represents the state of the lexer as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer for trig/turtle (and their line-based subsets n-triples & n-quads).
//
// Tokens for whitespace are comments are not not emitted.
//
// The design of the lexer and indeed much of the implementation is lifted from
// the template lexer in Go's standard library, and is governed by a BSD licence
// and Copyright 2011 The Go Authors.
type lexer struct {
	rdr *bufio.Reader

	input  []byte     // the input being scanned (should not inlcude newlines)
	state  stateFn    // the next lexing function to enter
	line   int        // the current line number
	pos    int        // the current position in input
	width  int        // width of the last rune read from input
	start  int        // start of current token
	unEsc  bool       // true when current token needs to be unescaped
	tokens chan token // channel of scanned tokens
}

func newLexer(r io.Reader) *lexer {
	l := lexer{
		rdr:    bufio.NewReader(r),
		tokens: make(chan token),
	}
	go l.run()
	return &l
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := decodeRune(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) unescape(s string, t tokenType) string {
	if !l.unEsc {
		return s
	}
	l.unEsc = false
	if t == tokenIRISuffix {
		return unescapeReservedChars(s)
	}
	return unescapeNumericString(s)
}

func unescapeNumericString(s string) string {
	r := []rune(s)
	buf := bytes.NewBuffer(make([]byte, 0, len(r)))

	for i := 0; i < len(r); {
		switch r[i] {
		case '\\':
			i++
			var c byte
			switch r[i] {
			case 't':
				c = '\t'
			case 'b':
				c = '\b'
			case 'n':
				c = '\n'
			case 'r':
				c = '\r'
			case 'f':
				c = '\f'
			case '"':
				c = '"'
			case '\'':
				c = '\''
			case '\\':
				c = '\\'
			case 'u':
				rc, _ := strconv.ParseInt(string(r[i+1:i+5]), 16, 32)
				// we can safely assume no error, because we allready veryfied
				// the escape sequence in the lex state funcitons
				buf.WriteRune(rune(rc))
				i += 5
				continue
			case 'U':
				rc, _ := strconv.ParseInt(string(r[i+1:i+9]), 16, 32)
				// we can safely assume no error, because we allready veryfied
				// the escape sequence in the lex state funcitons
				buf.WriteRune(rune(rc))
				i += 9
				continue
			}
			buf.WriteByte(c)
		default:
			buf.WriteRune(r[i])
		}
		i++
	}
	return buf.String()
}

func unescapeReservedChars(s string) string {
	r := []rune(s)
	buf := bytes.NewBuffer(make([]byte, 0, len(r)))

	for i := 0; i < len(r); {
		switch r[i] {
		case '\\':
			i++
			var c rune
			switch r[i] {
			case '_', '~', '.', '-', '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', '/', '?', '#', '@', '%':
				c = r[i]
			}
			buf.WriteRune(c)
		default:
			buf.WriteRune(r[i])
		}
		i++
	}
	return buf.String()
}

// emit publishes a token back to the comsumer.
func (l *lexer) emit(typ tokenType) {
	l.tokens <- token{
		typ:  typ,
		line: l.line,
		col:  l.start,
		text: l.unescape(string(l.input[l.start:l.pos]), typ),
	}

	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// acceptRunMin consumes a run of runes from the valid set, returning
// true if a minimum number of runes where consumed.
func (l *lexer) acceptRunMin(valid []byte, num int) bool {
	c := 0
	for bytes.IndexRune(valid, l.next()) >= 0 {
		c++
	}
	l.backup()
	return c >= num
}

// acceptExact consumes the given string in l.input and returns true,
// or otherwise false if the string is not matched in l.input.
// The string must not contain multi-byte runes.
func (l *lexer) acceptExact(s string) bool {
	if len(l.input[l.start:]) < len(s) {
		return false
	}
	if string(l.input[l.start:l.pos+len(s)-1]) == s {
		l.pos = l.pos + len(s) - 1
		return true
	}
	return false
}

// nextToken returns the next token from the input.
func (l *lexer) nextToken() token {
	token := <-l.tokens
	return token
}

func (l *lexer) feed(overwrite bool) bool {
again:
	line, err := l.rdr.ReadBytes('\n')
	if err != nil && len(line) == 0 {
		return false
	}

	l.line++
	if len(line) == 0 || line[0] == '#' {
		// skip empty lines and lines starting with comment
		l.emit(tokenEOL)
		goto again
	}

	if overwrite {
		// multi-line literal, concat to current input
		l.input = append(l.input, line...)
	} else {
		l.input = line
		l.pos = 0
		l.start = 0
	}

	return true
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for {
		if !l.feed(false) {
			break
		}

		for l.state = lexAny; l.state != nil; {
			l.state = l.state(l)
		}
	}

	// No more input to lex, emit final EOF token and terminate.
	// The value of the closed tokens channel is tokenEOF.
	close(l.tokens)
}

// state functions:

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextToken.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{
		tokenError,
		l.line,
		l.pos,
		fmt.Sprintf(format, args...),
	}
	return nil
}

func lexAny(l *lexer) stateFn {
	r := l.next()
	switch r {
	case '@':
		n := l.next()
		switch n {
		case 'p':
			l.start++ // consume '@''
			return lexPrefix
		case 'b':
			l.start++ // consume '@''
			return lexBase
		default:
			l.backup()
			return l.errorf("illegal character %q", r)
		}
	case '_':
		if l.peek() != ':' {
			return l.errorf("illegal character %q in blank node identifier", l.peek())
		}
		// consume & ignore '_:'
		l.next()
		l.ignore()
		return lexBNode
	case '<':
		l.ignore()
		return lexIRI
	case 'a':
		p := l.peek()
		for _, a := range okAfterRDFType {
			if p == a {
				l.emit(tokenRDFType)
				return lexAny
			}
		}
		// If not 'a' as rdf:type, it can be a prefixed local name starting with 'a'
		l.pos-- // undread 'a'
		return lexPrefixLabel
	case ':':
		// default namespace, no prefix
		l.backup()
		return lexPrefixLabel
	case '\'':
		l.backup()
		return lexLiteral
	case '"':
		l.backup()
		return lexLiteral
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '+', '-':
		l.backup()
		return lexNumber
	case ' ', '\t':
		// whitespace tokens are not emitted, so we ignore and continue
		l.ignore()
		return lexAny
	case '[':
		// Can be either an anonymous blank node: '[' WS* ']',
		// or start of blank node property list: '[' verb objectList (';' (verb objectList)?)* ']'
		for r = l.next(); r == ' ' || r == '\t'; r = l.next() {
		}
		if r == ']' {
			l.ignore()
			l.emit(tokenAnonBNode)
			return lexAny
		}
		l.backup()
		l.ignore()
		l.emit(tokenPropertyListStart)
		return lexAny
	case ']':
		// This must be closing of a blank node property list,
		// since the closing of anonymous blank node is lexed above.
		l.ignore()
		l.emit(tokenPropertyListEnd)
		return lexAny
	case '(':
		l.ignore()
		l.emit(tokenCollectionStart)
		return lexAny
	case ')':
		l.ignore()
		l.emit(tokenCollectionEnd)
		return lexAny
	case '.':
		if isDigit(l.peek()) {
			l.pos -= 2 // can only backup once with l.backup()
			return lexNumber
		}
		l.ignore()
		l.emit(tokenDot)
		return lexAny
	case '\r':
		if l.peek() == '\n' {
			l.next()
			return lexAny
		}
		l.ignore()
		l.emit(tokenEOL)
		return lexAny
	case '\n':
		l.ignore()
		l.emit(tokenEOL)
		return nil
	case ';':
		l.emit(tokenSemicolon)
		return lexAny
	case ',':
		l.emit(tokenComma)
		return lexAny
	case '#', eof:
		// comment tokens are not emitted, so treated as eof
		l.ignore()
		l.emit(tokenEOL)
		return nil // This parks the lexer until it gets more input
	case 'P':
		if l.acceptExact("PREFIX") {
			l.emit(tokenSparqlPrefix)
			// consume and ignore any whitespace before localname
			for r := l.next(); r == ' ' || r == '\t'; r = l.next() {
			}
			l.backup()
			l.ignore()
			return lexPrefixLabelInDirective
		}
		fallthrough // continue to default
	case 'B':
		if l.acceptExact("BASE") {
			l.emit(tokenSparqlBase)
			return lexAny
		}
		fallthrough // continue to default
	default:
		if isPnCharsBase(r) {
			l.backup()
			return lexPrefixLabel
		}
		return l.errorf("illegal character %q", r)
	}
}

func hasValidScheme(l *lexer) bool {
	// RFC 2396: scheme = alpha *( alpha | digit | "+" | "-" | "." )

	// decode first rune, must be in set [a-zA-Z]
	r, w := decodeRune(l.input[l.start:])
	if !isAlpha(r) {
		return false
	}

	// remaining runes must be alphanumeric or '+', '-', '.''
	for p := l.start + w; p < l.pos-w; p = p + w {
		r, w = decodeRune(l.input[p:])
		if isAlphaOrDigit(r) || r == '+' || r == '-' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func _lexIRI(l *lexer) (stateFn, bool) {
	hasScheme := false    // does it have a scheme? defines if IRI is absolute or relative
	maybeAbsolute := true // false if we reach a non-valid scheme rune before ':'
	for {
		r := l.next()
		if r == eof {
			return l.errorf("bad IRI: no closing '>'"), false
		}
		for _, bad := range badIRIRunes {
			if r == bad {
				return l.errorf("bad IRI: disallowed character %q", r), false
			}
		}

		if r == '\\' {
			// handle numeric escape sequences for unicode points:
			esc := l.peek()
			switch esc {
			case 'u':
				l.next() // cosume 'u'
				if !l.acceptRunMin(hex, 4) {
					return l.errorf("bad IRI: insufficent hex digits in unicode escape"), false
				}
				l.unEsc = true
			case 'U':
				l.next() // cosume 'U'
				if !l.acceptRunMin(hex, 8) {
					return l.errorf("bad IRI: insufficent hex digits in unicode escape"), false
				}
				l.unEsc = true
			case eof:
				return l.errorf("bad IRI: no closing '>'"), false
			default:
				return l.errorf("bad IRI: disallowed escape character %q", esc), false
			}
		}
		if maybeAbsolute && r == ':' {
			// Check if we have an absolute IRI
			if l.pos != l.start && hasValidScheme(l) && l.peek() != eof {
				hasScheme = true
				maybeAbsolute = false // stop checking for absolute
			}
		}
		if r == '>' {
			// reached end of IRI
			break
		}
	}
	l.backup()
	return nil, hasScheme
}

func lexIRI(l *lexer) stateFn {
	res, absolute := _lexIRI(l)
	if res != nil {
		return res
	}
	if absolute {
		l.emit(tokenIRIAbs)
	} else {
		l.emit(tokenIRIRel)
	}

	// ignore '>'
	l.pos++
	l.ignore()

	return lexAny
}

func lexLiteral(l *lexer) stateFn {
	quote := l.next()
	quoteCount := 1
	var r rune

	l.ignore()
	for quoteCount <= 3 {
		r = l.next()
		if r != quote {
			break
		}
		l.ignore()
		quoteCount++
	}
	if quoteCount == 2 {
		// Empty string
		quoteCount = 0
		l.pos = l.start
		goto done
	}
outer:
	for {
		switch r {
		case '\n':
			if quoteCount != 3 {
				return l.errorf("bad literal: newline not allowed in single-quoted string")
			}
			// triple-quoted strings can contain newlines
			if !l.feed(true) {
				return l.errorf("bad literal: no closing quote: %q", quote)
			}
		case '\r':
			if quoteCount != 3 {
				return l.errorf("bad literal: carriage return not allowed in single-quoted string")
			}
		case eof:
			return l.errorf("bad literal: no closing quote: %q", quote)
		case '\\':
			// handle numeric escape sequences for unicode points:
			esc := l.next()
			switch esc {
			case 't', 'b', 'n', 'r', 'f', '"', '\'', '\\':
				l.unEsc = true
			case 'u':
				if !l.acceptRunMin(hex, 4) {
					return l.errorf("bad literal: insufficent hex digits in unicode escape")
				}
				l.unEsc = true
			case 'U':
				if !l.acceptRunMin(hex, 8) {
					return l.errorf("bad literal: insufficent hex digits in unicode escape")
				}
				l.unEsc = true
			case eof:
				return l.errorf("bad literal: no closing quote %q", quote)
			default:
				return l.errorf("bad literal: disallowed escape character %q", esc)
			}
		case quote:
			// reached end of Literal
			if quoteCount == 3 {
				if l.next() != quote {
					return l.errorf("bad literal: missing triple quote closing %q", quote)
				}
				if l.next() != quote {
					return l.errorf("bad literal: missing triple quote closing %q", quote)
				}
			}
			l.pos -= quoteCount
			break outer
		}
		r = l.next()
	}
done:
	if quoteCount == 3 {
		l.emit(tokenLiteral3)
	} else {
		l.emit(tokenLiteral)
	}

	// ignore quote(s)
	l.pos += quoteCount
	l.ignore()

	// check if literal has language tag or datatype URI:
	r = l.next()
	switch r {
	case '@':
		l.emit(tokenLangMarker)
		return lexLang
	case '^':
		if l.next() != '^' {
			return l.errorf("bad literal: invalid datatype IRI")
		}
		l.emit(tokenDataTypeMarker)
		if l.next() != '<' {
			return l.errorf("bad literal: invalid datatype IRI")
		}
		l.ignore() // ignore '<'
		return lexIRI
	case ' ', '\t':
		return lexAny
	default:
		l.backup()
		return lexAny
	}
}

func lexNumber(l *lexer) stateFn {
	// integer: [+-]?[0-9]+
	// decimal: [+-]?[0-9]*\.[0-9]+
	// dobule: (([+-]?[0-9]+\.[0-9]+)|
	//          ([+-]?\.[0-9]+)|
	//          ([+-]?[0-9]))
	//         (e|E)[+-]?[0-9]+
	gotDot := false
	gotE := false
	r := l.next()
	switch r {
	case '+', '-':
	case '.':
		// cannot be an integer
		gotDot = true
	default:
		// a digit
	outer:
		for {
			r = l.next()
			switch {
			case isDigit(r):
				continue
			case r == '.':
				if gotDot {
					return l.errorf("bad literal: illegal number syntax")
				}
				gotDot = true
			case r == 'e', r == 'E':
				if gotE {
					return l.errorf("bad literal: illegal number syntax")
				}
				gotE = true
				p := l.peek()
				if p == '+' || p == '-' {
					l.next()
				}
			default:
				// TODO can number be followed by something else than:
				// - whitespace
				// - comma (if in a object list)
				// - semicolon (if in a predicate list)
				// - eof (end of line/input)
				// - ')' end of list
				// - ']' end of list
				// - What about dot?? as end of triples statement, not part of number
				if r == ' ' || r == ',' || r == ';' || r == eof || r == ')' || r == ']' {
					l.backup()
					break outer
				}
				l.errorf("bad literal: illegal number syntax (number followed by %q)", r)
			}
		}

		switch {
		case gotE:
			l.emit(tokenLiteralDouble)
		case gotDot:
			l.emit(tokenLiteralDecimal)
		default:
			l.emit(tokenLiteralInteger)
		}

	}

	return lexAny
}

func lexBNode(l *lexer) stateFn {
	r := l.next()
	if r == eof {
		return l.errorf("bad blank node: unexpected end of line")
	}
	if !(isPnCharsU(r) || isDigit(r)) {
		return l.errorf("bad blank node: invalid character %q", r)
	}

	for {
		r = l.next()

		if r == '.' {
			// Blank node labels can include '.', except as the final character
			if isPnChars(l.peek()) {
				continue
			} else {
				l.pos-- // backup, '.' has width 1
				break
			}
		}

		if !(isPnChars(r)) {
			l.backup()
			break
		}
	}
	l.emit(tokenBNode)
	return lexAny
}

func lexLang(l *lexer) stateFn {
	// consume [A-Za-z]+
	c := 0
	for r := l.next(); isAlpha(r); r = l.next() {
		c++
	}
	l.backup()
	if c == 0 {
		return l.errorf("bad literal: invalid language tag")
	}

	// consume ('-' [A-Za-z0-9])* if present
	if l.peek() == '-' {
		l.next() // consume '-'

		c = 0
		for r := l.next(); isAlphaOrDigit(r); r = l.next() {
			c++
		}
		l.backup()
		if c == 0 {
			return l.errorf("bad literal: invalid language tag")
		}
	}

	l.emit(tokenLang)
	return lexAny
}

func lexPrefix(l *lexer) stateFn {
	if l.acceptExact("prefix") {
		l.emit(tokenPrefix)
		// consume and ignore any whitespace before localname
		for r := l.next(); r == ' ' || r == '\t'; r = l.next() {
		}
		l.backup()
		l.ignore()
		return lexPrefixLabelInDirective
	}
	return l.errorf("invalid character 'p'")
}

func lexPrefixLabelInDirective(l *lexer) stateFn {
	r := l.next()
	if r == ':' {
		//PN_PREFIX can be empty
		l.emit(tokenPrefixLabel)
		return lexAny
	}
	if !isPnCharsBase(r) {
		return l.errorf("illegal token: %s", string(l.input[l.start:l.pos]))
	}
	for {
		r = l.next()
		if r == ':' {
			l.backup()
			break
		}
		if !(isPnChars(r) || (r == '.' && l.peek() != ':')) {
			return l.errorf("illegal token: %s", string(l.input[l.start:l.pos]))
		}
	}

	l.emit(tokenPrefixLabel)

	// consume and ignore ':'
	l.next()
	l.ignore()
	// consume and ignore any whitespace
	for r = l.next(); r == ' ' || r == '\t'; r = l.next() {
	}
	if r == '<' {
		l.ignore()
		return lexIRI
	}

	l.backup()
	return lexIRISuffix
}

func lexPrefixLabel(l *lexer) stateFn {
	r := l.next()
	if r == ':' {
		//PN_PREFIX can be empty
		l.emit(tokenPrefixLabel)
		return lexIRISuffix
	}
	if !isPnCharsBase(r) {
		return l.errorf("illegal token: %s", string(l.input[l.start:l.pos]))
	}
	for {
		r = l.next()
		if r == ':' {
			l.backup()
			break
		}
		if !(isPnChars(r) || (r == '.' && l.peek() != ':')) {
			return l.errorf("illegal token: %s", string(l.input[l.start:l.pos]))
		}
	}

	l.emit(tokenPrefixLabel)

	// consume and ignore ':'
	l.next()
	l.ignore()

	return lexIRISuffix

}

func lexIRISuffix(l *lexer) stateFn {
	//(PN_CHARS_U | ':' | [0-9] | PLX) ((PN_CHARS | '.' | ':' | PLX)* (PN_CHARS | ':' | PLX))?
	r := l.next()
	if r == ' ' {
		// prefix ony IRI
		l.ignore()
		l.emit(tokenIRISuffix)
		return lexAny
	}
	if !isPnLocalFirst(r) {
		return l.errorf("invalid character %q", r)
	}
	for r = l.next(); isPnLocalMid(r); r = l.next() {
		if r == '\\' {
			p := l.next()
			for _, esc := range pnLocalEsc {
				if esc == p {
					l.unEsc = true
					continue
				}
			}
		}
		// TODO check validity of:
		// - ('%' hex hex)
	}
	l.backup()
	if l.input[l.pos] == '.' {
		// last rune cannot be dot, otherwise isPnLocalMid(r) is valid for last position as well
		return l.errorf("illegal token: %s", string(l.input[l.start:l.pos]))
	}
	l.emit(tokenIRISuffix)
	return lexAny
}

func lexBase(l *lexer) stateFn {
	if l.acceptExact("base") {
		l.emit(tokenBase)
		// consume and ignore any whitespace before base IRI
		for r := l.next(); r == ' ' || r == '\t'; r = l.next() {
		}
		l.backup()
		l.ignore()
		return lexAny
	}
	return l.errorf("invalid character 'b'")
}
