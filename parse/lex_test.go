package parse

import (
	"fmt"
	"testing"
)

// Make the token types prettyprint.
var tokenName = map[tokenType]string{
	tokenError:    "Error",
	tokenEOL:      "EOL",
	tokenEOF:      "EOF",
	tokenIRI:      "IRI",
	tokenLiteral:  "Literal",
	tokenLang:     "Language tag",
	tokenDataType: "Literal data type",
	tokenDot:      ".",
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
			{tokenIRI, "a"},
			{tokenEOL, ""}},
		},
		{"<http://www.w3.org/1999/02/22-rdf-syntax-ns#type>", []testToken{
			{tokenIRI, "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			{tokenEOL, ""}},
		},
		{`  <x><y> <z>  <\u0053> `, []testToken{
			{tokenIRI, "x"},
			{tokenIRI, "y"},
			{tokenIRI, "z"},
			{tokenIRI, "S"},
			{tokenEOL, ""}},
		},
		{`<s><p><o>.#comment`, []testToken{
			{tokenIRI, "s"},
			{tokenIRI, "p"},
			{tokenIRI, "o"},
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
		{`"a"^^<mydatatype>`, []testToken{
			{tokenLiteral, "a"},
			{tokenDataType, "mydatatype"},
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
