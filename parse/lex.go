package parse

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode/utf8"
)

type tokenType int

const (
	// special tokens
	tokenEOL   tokenType = iota // end of line
	tokenEOF                    // end of input to be scanned TODO remove?
	tokenError                  // an illegal token

	// turtle tokens
	tokenIRI      // RDF IRI reference
	tokenBNode    // RDF blank node
	tokenLiteral  // RDF literal
	tokenLang     // literal language tag
	tokenDataType // literal data type
	tokenDot      // .
)

const eof = -1

var hex = []byte("0123456789ABCDEFabcdef")

var badIRIRunes = [...]rune{' ', '<', '"', '{', '}', '|', '^', '`'}

// isAlpha tests if rune is in the set [a-zA-Z]
func isAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isAlphaOrDigit tests if rune is in the set [a-zA-Z0-9]
func isAlphaOrDigit(r rune) bool {
	return isAlpha(r) || (r >= '0' && r <= '9')
}

// token represents a token emitted by the lexer.
type token struct {
	typ  tokenType // type of token
	line int       // line number
	col  int       // column number (measured in runes, not bytes)
	text string    // the value of the token
}

// stateFn represents the state of the lexer as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer for trig/turtle (and their line-based subsets n-triples & n-quads).
//
// The lexer is assumed to be working on one line at a time. When end of line
// is reached, tokenEOL is emitted, and the caller may supply more lines to
// the incoming channel. If there are no more input to be scanned, the user
// must call stop(), which will terminate the lexer and emit a final tokenEOF. TODO NO?
//
// Tokens for whitespace and comments are not emitted.
//
// The design of the lexer and indeed much of the implementation is lifted from
// the template lexer in Go's standard library, and is governed by a BSD licence
// and Copyright 2011 The Go Authors.
type lexer struct {
	incoming chan []byte

	input  []byte     // the input being scanned (should not inlcude newlines)
	state  stateFn    // the next lexing function to enter
	line   int        // the current line number
	pos    int        // the current position in input
	width  int        // width of the last rune read from input
	start  int        // start of current token
	unEsc  bool       // true when current token needs to be unescaped
	tokens chan token // channel of scanned tokens
}

func newLexer() *lexer {
	l := lexer{
		incoming: make(chan []byte),
		tokens:   make(chan token),
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
	r, w := utf8.DecodeRune(l.input[l.pos:])
	// TODO:
	/*if r == utf8.RuneError {
		l.errorf("invalid utf-8 encoding")
	}*/
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

func (l *lexer) unescape(s string) string {
	if !l.unEsc {
		return s
	}
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
	l.unEsc = false
	return buf.String()
}

// emit publishes a token back to the comsumer.
func (l *lexer) emit(typ tokenType) {
	l.tokens <- token{
		typ:  typ,
		line: l.line,
		col:  l.start,
		text: l.unescape(string(l.input[l.start:l.pos])),
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

// nextToken returns the next token from the input.
func (l *lexer) nextToken() token {
	token := <-l.tokens
	return token
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
again:
	line := <-l.incoming
	l.input = line
	l.pos = 0
	l.start = 0
	l.line++
	if line == nil {
		// closed incoming channel
		return
	}
	for l.state = lexAny; l.state != nil; {
		l.state = l.state(l)
	}
	goto again
}

func (l *lexer) stop() {
	close(l.incoming)
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
	case '"':
		l.ignore()
		return lexLiteral
	case ' ', '\t':
		// whitespace tokens are not emitted, we continue
		return lexAny
	case '.':
		l.ignore()
		l.emit(tokenDot)
		return lexAny
	case '\n':
		// Lexer should not get input with newlines.
		// TODO panic or handle somehow?
		return nil
	case '#', eof:
		l.ignore()
		l.emit(tokenEOL)
		return nil
	default:
		return l.errorf("illegal character %q", r)
	}
}

func _lexIRI(l *lexer) stateFn {
	for {
		r := l.next()
		if r == eof {
			return l.errorf("bad IRI: no closing '>'")
		}
		for _, bad := range badIRIRunes {
			if r == bad {
				return l.errorf("bad IRI: disallowed character %q", r)
			}
		}

		if r == '\\' {
			// handle numeric escape sequences for unicode points:
			esc := l.peek()
			switch esc {
			case 'u':
				l.next() // cosume 'u'
				if !l.acceptRunMin(hex, 4) {
					return l.errorf("bad IRI: insufficent hex digits in unicode escape")
				}
				l.unEsc = true
			case 'U':
				l.next() // cosume 'U'
				if !l.acceptRunMin(hex, 8) {
					return l.errorf("bad IRI: insufficent hex digits in unicode escape")
				}
				l.unEsc = true
			case eof:
				return l.errorf("bad IRI: no closing '>'")
			default:
				return l.errorf("bad IRI: disallowed escape character %q", esc)
			}
		}
		if r == '>' {
			// reached end of IRI
			break
		}
	}
	l.backup()
	return nil
}

func lexIRI(l *lexer) stateFn {
	res := _lexIRI(l)
	if res != nil {
		return res
	}
	l.emit(tokenIRI)

	// ignore '>'
	l.pos++
	l.ignore()

	return lexAny
}

func lexIRIDT(l *lexer) stateFn {
	res := _lexIRI(l)
	if res != nil {
		return res
	}
	l.emit(tokenDataType)

	// ignore '>'
	l.pos++
	l.ignore()

	return lexAny
}

func lexLiteral(l *lexer) stateFn {
	for {
		r := l.next()
		if r == eof {
			return l.errorf("bad Literal: no closing '\"'")
		}
		if r == '\\' {
			// handle numeric escape sequences for unicode points:
			esc := l.peek()
			switch esc {
			case 't', 'b', 'n', 'r', 'f', '"', '\'', '\\':
				l.next() // consume '\'
				l.unEsc = true
			case 'u':
				l.next() // cosume 'u'
				if !l.acceptRunMin(hex, 4) {
					return l.errorf("bad literal: insufficent hex digits in unicode escape")
				}
				l.unEsc = true
			case 'U':
				l.next() // cosume 'U'
				if !l.acceptRunMin(hex, 8) {
					return l.errorf("bad literal: insufficent hex digits in unicode escape")
				}
				l.unEsc = true
			case eof:
				return l.errorf("bad literal: no closing '\"'")
			default:
				return l.errorf("bad literal: disallowed escape character %q", esc)
			}
		}
		if r == '"' {
			// reached end of Literal
			break
		}
	}
	l.backup()

	l.emit(tokenLiteral)

	// ignore '"'
	l.pos++
	l.ignore()
	return lexLang
}

func lexBNode(l *lexer) stateFn {
	// TODO make this according to spec
	r := l.next()
	if r == eof {
		return l.errorf("bad blank node: unexpected end of line")
	}
	for {
		if r == '<' || r == '"' || r == ' ' || r == eof {
			l.backup()
			break
		}
		r = l.next()
	}
	l.emit(tokenBNode)
	return lexAny
}

func lexLang(l *lexer) stateFn {
	if l.peek() != '@' {
		return lexDataType
	}
	// consume and ignore '@'
	l.next()
	l.ignore()

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

func lexDataType(l *lexer) stateFn {
	if l.peek() != '^' {
		return lexAny
	}
	l.next() // consume '^'
	if l.peek() != '^' {
		return l.errorf("bad literal: incomplete datatype marker")
	}
	l.next() // consume '^'
	if l.peek() != '<' {
		return l.errorf("bad literal: expected IRI after datatype marker, found %q", l.peek())
	}
	l.next() // consume and ignore '^^<'
	l.ignore()

	return lexIRIDT
}
