package parse

import (
	"fmt"
	"strings"
	"testing"
)

// Make the token types prettyprint.
var tokenName = map[tokenType]string{
	tokenNone:           "None",
	tokenError:          "Error",
	tokenEOL:            "EOL",
	tokenIRIAbs:         "IRI (absolute)",
	tokenIRIRel:         "IRI (relative)",
	tokenLiteral:        "Literal",
	tokenBNode:          "Blank node",
	tokenLangMarker:     "Language tag marker",
	tokenLang:           "Language tag",
	tokenDataTypeMarker: "Literal datatype marker",
	tokenDot:            "Dot",
}

func (t tokenType) String() string {
	s := tokenName[t]
	if s == "" {
		return fmt.Sprintf("token%d", int(t))
	}
	return s
}

type testToken struct {
	Typ  tokenType
	Text string
}

func collect(l *lexer) []testToken {
	tokens := []testToken{}
	for {
		tk := l.nextToken() // <- blocking
		tokens = append(tokens, testToken{tk.typ, tk.text})
		if tk.typ == tokenEOL || tk.typ == tokenError {
			break
		}

	}
	return tokens
}

// equalTokens tests if two slice of testToken are equal
func equalTokens(a, b []testToken) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if a[k].Typ != b[k].Typ {
			return false
		}
		if a[k].Text != b[k].Text {
			return false
		}
	}
	return true
}

func TestTokens(t *testing.T) {
	lexTests := []struct {
		in   string
		want []testToken
	}{
		{"", []testToken{
			{tokenEOL, ""}},
		},
		{" \t ", []testToken{
			{tokenEOL, ""}},
		},
		{" \n", []testToken{
			{tokenEOL, ""},
			{tokenEOL, ""}},
		},
		{"<a>", []testToken{
			{tokenIRIRel, "a"},
			{tokenEOL, ""}},
		},
		{"<http://www.w3.org/1999/02/22-rdf-syntax-ns#type>", []testToken{
			{tokenIRIAbs, "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			{tokenEOL, ""}},
		},
		{`  <x><y> <z>  <\u0053> `, []testToken{
			{tokenIRIRel, "x"},
			{tokenIRIRel, "y"},
			{tokenIRIRel, "z"},
			{tokenIRIRel, "S"},
			{tokenEOL, ""}},
		},
		{`<s><p><o>.#comment`, []testToken{
			{tokenIRIRel, "s"},
			{tokenIRIRel, "p"},
			{tokenIRIRel, "o"},
			{tokenDot, ""},
			{tokenEOL, ""}},
		},
		{`"a"`, []testToken{
			{tokenLiteral, "a"},
			{tokenEOL, ""}},
		},
		{`"""a"""`, []testToken{
			{tokenLiteral, ""},
			{tokenLiteral, "a"},
			{tokenLiteral, ""},
			{tokenEOL, ""}},
		},
		{`"æøå üçgen こんにちは" # comments should be ignored`, []testToken{
			{tokenLiteral, "æøå üçgen こんにちは"},
			{tokenEOL, ""}},
		},
		{`"KI\u0053\U00000053ki⚡⚡"`, []testToken{
			{tokenLiteral, "KISSki⚡⚡"},
			{tokenEOL, ""}},
		},
		{`"she said: \"hi\""`, []testToken{
			{tokenLiteral, `she said: "hi"`},
			{tokenEOL, ""}},
		},
		{`"escapes:\t\\\"\n\t"`, []testToken{
			{tokenLiteral, `escapes:	\"
	`},
			{tokenEOL, ""},
		}},
		{`"hei"@nb-no "hi"@en #language tags`, []testToken{
			{tokenLiteral, "hei"},
			{tokenLangMarker, "@"},
			{tokenLang, "nb-no"},
			{tokenLiteral, "hi"},
			{tokenLangMarker, "@"},
			{tokenLang, "en"},
			{tokenEOL, ""}},
		},
		{`"hei"@`, []testToken{
			{tokenLiteral, "hei"},
			{tokenLangMarker, "@"},
			{tokenError, "bad literal: invalid language tag"}},
		},
		{`"a"^^<s://mydatatype>`, []testToken{
			{tokenLiteral, "a"},
			{tokenDataTypeMarker, "^^"},
			{tokenIRIAbs, "s://mydatatype"},
			{tokenEOL, ""}},
		},
		{`"a"^`, []testToken{
			{tokenLiteral, "a"},
			{tokenError, "bad literal: invalid datatype IRI"}},
		},
		{`"a"^^`, []testToken{
			{tokenLiteral, "a"},
			{tokenDataTypeMarker, "^^"},
			{tokenError, "bad literal: invalid datatype IRI"}},
		},
		{`"a"^^xyz`, []testToken{
			{tokenLiteral, "a"},
			{tokenDataTypeMarker, "^^"},
			{tokenError, "bad literal: invalid datatype IRI"}},
		},
		{`_:a_BlankLabel123.`, []testToken{
			{tokenBNode, "a_BlankLabel123"},
			{tokenDot, ""},
			{tokenEOL, ""}},
		},
		{`_:a-b.c.`, []testToken{
			{tokenBNode, "a-b.c"},
			{tokenDot, ""},
			{tokenEOL, ""}},
		},
		{`#comment
		  <s><p><o>.#comment
		  # comment	

		  <s><p2> "yo"
		  ####
		  `, []testToken{
			{tokenEOL, ""},
			{tokenIRIRel, "s"},
			{tokenIRIRel, "p"},
			{tokenIRIRel, "o"},
			{tokenDot, ""},
			{tokenEOL, ""},
			{tokenEOL, ""},
			{tokenEOL, ""},
			{tokenIRIRel, "s"},
			{tokenIRIRel, "p2"},
			{tokenLiteral, "yo"},
			{tokenEOL, ""},
			{tokenEOL, ""},
			{tokenEOL, ""}},
		},
	}

	lex := newLexer()
	for _, tt := range lexTests {
		res := []testToken{}
		for _, l := range strings.Split(tt.in, "\n") {
			lex.incoming <- []byte(l)
			for _, t := range collect(lex) {
				res = append(res, t)
			}

		}
		if !equalTokens(tt.want, res) {
			t.Errorf("lexing %v, got:\n\t%v\nexpected:\n\t%v", tt.in, res, tt.want)
		}
	}
	lex.stop()
}
