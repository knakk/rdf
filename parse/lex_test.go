package parse

import (
	"fmt"
	"testing"
)

// Make the token types prettyprint.
var tokenName = map[tokenType]string{
	tokenError:       "Error",
	tokenEOL:         "EOL",
	tokenEOF:         "EOF",
	tokenIRIAbs:      "IRI (absolute)",
	tokenIRIRel:      "IRI (relative)",
	tokenLiteral:     "Literal",
	tokenBNode:       "Blank node",
	tokenLang:        "Language tag",
	tokenDataTypeAbs: "Literal data type IRI (absolute)",
	tokenDataTypeRel: "Literal data typ IRI (relative)",
	tokenDot:         ".",
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
		tk := l.nextToken()
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
		{`"hei"@nb-no "hi"@en #language tags`, []testToken{
			{tokenLiteral, "hei"},
			{tokenLang, "nb-no"},
			{tokenLiteral, "hi"},
			{tokenLang, "en"},
			{tokenEOL, ""}},
		},
		{`"a"^^<s://mydatatype>`, []testToken{
			{tokenLiteral, "a"},
			{tokenDataTypeAbs, "s://mydatatype"},
			{tokenEOL, ""}},
		},
		{`_:a`, []testToken{
			{tokenBNode, "a"},
			{tokenEOL, ""}},
		},
	}

	lex := newLexer()
	for _, tt := range lexTests {
		lex.incoming <- []byte(tt.in)
		res := collect(lex)
		if !equalTokens(tt.want, res) {
			t.Errorf("lexing %v, got:\n\t%v\nexpected:\n\t%v", tt.in, res, tt.want)
		}
	}
	lex.stop()
}

//func TestTokenErrors(t *testing.T) { }
