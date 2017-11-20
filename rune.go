package rdf

import (
	"bytes"
	"math"
	"unicode"
)

// Rune helper values and functions:
var (
	hex            = []byte("0123456789ABCDEFabcdef")
	pnLocalEsc     = [...]rune{'_', '~', '.', '-', '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', '/', '?', '#', '@', '%'}
	badIRIRunes    = [...]rune{' ', '<', '"', '{', '}', '|', '^', '`'}
	badIRIRunesEsc = [...]rune{' ', '<', '"', '{', '}', '|', '^', '`', '>'}
	okAfterRDFType = [...]rune{' ', '\t', '<', '"', '\''}
	whitespace     = [...]rune{' ', '\t', '\r', '\n'}
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
		':', ':',
		'0', '9',
		'%', '%',
		'\\', '\\', // last of PN_LOCAL first character
		'-', '-',
		0x00B7, 0x00B7,
		0x0300, 0x036F,
		0x203F, 0x2040,
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
	return check(r, plTab[:2*19])
}

func isPnLocalMid(r rune) bool {
	return check(r, plTab)
}

func isWhitespace(r rune) bool {
	for _, w := range whitespace {
		if r == w {
			return true
		}
	}
	return false
}

const (
	runeError = math.MaxInt32

	t1 = 0x00 // 0000 0000
	tx = 0x80 // 1000 0000
	t2 = 0xC0 // 1100 0000
	t3 = 0xE0 // 1110 0000
	t4 = 0xF0 // 1111 0000
	t5 = 0xF8 // 1111 1000

	maskx = 0x3F // 0011 1111
	mask2 = 0x1F // 0001 1111
	mask3 = 0x0F // 0000 1111
	mask4 = 0x07 // 0000 0111

	rune1Max = 1<<7 - 1
	rune2Max = 1<<11 - 1
	rune3Max = 1<<16 - 1
)

// decodeRune is utf8.DecodeRune from the standard library, except it uses a different
// value for illegal runes. The value used by utf8.RuneError i \uFFFD, which is accepted
// in parts of the turtle grammar.
//
// Go's utf8 package is copyright 2009 The Go Authurs, and goverened by a BSD-license.
// The idea and implementation of the custom decodeRune function is lifted from
// https://github.com/cznic/scanner/blob/master/nquads/etc.go
// and is Copyright 2014 The scanner Authors, also governed by a BSD-license.
func decodeRune(s []byte) (r rune, size int) {
	n := len(s)
	if n < 1 {
		return 0, 0
	}
	c0 := s[0]

	// 1-byte, 7-bit sequence?
	if c0 < tx {
		return rune(c0), 1
	}

	// unexpected continuation byte?
	if c0 < t2 {
		return runeError, 1
	}

	// need first continuation byte
	if n < 2 {
		return runeError, 1
	}
	c1 := s[1]
	if c1 < tx || t2 <= c1 {
		return runeError, 1
	}

	// 2-byte, 11-bit sequence?
	if c0 < t3 {
		r = rune(c0&mask2)<<6 | rune(c1&maskx)
		if r <= rune1Max {
			return runeError, 1
		}
		return r, 2
	}

	// need second continuation byte
	if n < 3 {
		return runeError, 1
	}
	c2 := s[2]
	if c2 < tx || t2 <= c2 {
		return runeError, 1
	}

	// 3-byte, 16-bit sequence?
	if c0 < t4 {
		r = rune(c0&mask3)<<12 | rune(c1&maskx)<<6 | rune(c2&maskx)
		if r <= rune2Max {
			return runeError, 1
		}
		return r, 3
	}

	// need third continuation byte
	if n < 4 {
		return runeError, 1
	}
	c3 := s[3]
	if c3 < tx || t2 <= c3 {
		return runeError, 1
	}

	// 4-byte, 21-bit sequence?
	if c0 < t5 {
		r = rune(c0&mask4)<<18 | rune(c1&maskx)<<12 | rune(c2&maskx)<<6 | rune(c3&maskx)
		if r <= rune3Max || unicode.MaxRune < r {
			return runeError, 1
		}
		return r, 4
	}

	// error
	return runeError, 1
}

// escapeLiteral escapes a Literal string for serialization to N-Triples (canonical form).
func escapeLiteral(l string) string {
	var buf bytes.Buffer
	for _, r := range l {
		switch r {
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
