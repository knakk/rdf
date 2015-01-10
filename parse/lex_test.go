package parse

import (
	"fmt"
	"strings"
	"testing"
)

// Make the token types prettyprint.
var tokenName = map[tokenType]string{
	tokenError:             "Error",
	tokenEOL:               "EOL",
	tokenEOF:               "EOF",
	tokenIRIAbs:            "IRI (absolute)",
	tokenIRIRel:            "IRI (relative)",
	tokenLiteral:           "Literal",
	tokenLiteralInteger:    "Literal (integer shorthand syntax)",
	tokenLiteralDouble:     "Literal (double shorthand syntax)",
	tokenLiteralDecimal:    "Literal (decimal shorthand syntax)",
	tokenBNode:             "Blank node",
	tokenLangMarker:        "Language tag marker",
	tokenLang:              "Language tag",
	tokenDataTypeMarker:    "Literal datatype marker",
	tokenDot:               "Dot",
	tokenSemicolon:         "Semicolon",
	tokenComma:             "Comma",
	tokenRDFType:           "rdf:type",
	tokenPrefix:            "@prefix",
	tokenPrefixLabel:       "Prefix label",
	tokenIRISuffix:         "IRI suffix",
	tokenBase:              "@base",
	tokenSparqlBase:        "BASE",
	tokenSparqlPrefix:      "PREFIX",
	tokenAnonBNode:         "Anonymous blank node",
	tokenPropertyListStart: "Property list start",
	tokenPropertyListEnd:   "Property list end",
	tokenCollectionStart:   "Collection start",
	tokenCollectionEnd:     "Collection end",
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
		if tk.typ == tokenEOF || tk.typ == tokenError {
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
			{tokenEOF, ""}},
		},
		{" \t ", []testToken{
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{" \n ", []testToken{
			{tokenEOL, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{"<a>", []testToken{
			{tokenIRIRel, "a"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{"<http://www.w3.org/1999/02/22-rdf-syntax-ns#type>", []testToken{
			{tokenIRIAbs, "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`<s:1> a <o:1>`, []testToken{
			{tokenIRIAbs, "s:1"},
			{tokenRDFType, "a"},
			{tokenIRIAbs, "o:1"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`<a>a<b>.`, []testToken{
			{tokenIRIRel, "a"},
			{tokenRDFType, "a"},
			{tokenIRIRel, "b"},
			{tokenDot, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`  <x><y> <z>  <\u0053> `, []testToken{
			{tokenIRIRel, "x"},
			{tokenIRIRel, "y"},
			{tokenIRIRel, "z"},
			{tokenIRIRel, "S"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`<s><p><o>.#comment`, []testToken{
			{tokenIRIRel, "s"},
			{tokenIRIRel, "p"},
			{tokenIRIRel, "o"},
			{tokenDot, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`"a"`, []testToken{
			{tokenLiteral, "a"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`"""a"""`, []testToken{
			{tokenLiteral, ""},
			{tokenLiteral, "a"},
			{tokenLiteral, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`"æøå üçgen こんにちは" # comments text`, []testToken{
			{tokenLiteral, "æøå üçgen こんにちは"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`"KI\u0053\U00000053ki⚡⚡"`, []testToken{
			{tokenLiteral, "KISSki⚡⚡"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`"she said: \"hi\""`, []testToken{
			{tokenLiteral, `she said: "hi"`},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`"escapes:\t\\\"\n\t"`, []testToken{
			{tokenLiteral, `escapes:	\"
	`},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`"hei"@nb-no "hi"@en #language tags`, []testToken{
			{tokenLiteral, "hei"},
			{tokenLangMarker, "@"},
			{tokenLang, "nb-no"},
			{tokenLiteral, "hi"},
			{tokenLangMarker, "@"},
			{tokenLang, "en"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
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
			{tokenEOL, ""},
			{tokenEOF, ""}},
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
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`_:a-b.c.`, []testToken{
			{tokenBNode, "a-b.c"},
			{tokenDot, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
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
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`@prefix a: <http:/a.org/>.`, []testToken{
			{tokenPrefix, "prefix"},
			{tokenPrefixLabel, "a"},
			{tokenIRIAbs, "http:/a.org/"},
			{tokenDot, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`p:s <http://a.example/p> <http://a.example/o>`, []testToken{
			{tokenPrefixLabel, "p"},
			{tokenIRISuffix, "s"},
			{tokenIRIAbs, "http://a.example/p"},
			{tokenIRIAbs, "http://a.example/o"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{"@base <http:/a.org/>.", []testToken{
			{tokenBase, "base"},
			{tokenIRIAbs, "http:/a.org/"},
			{tokenDot, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{"BASE <http:/a.org/>", []testToken{
			{tokenSparqlBase, "BASE"},
			{tokenIRIAbs, "http:/a.org/"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{"[] <a> <b> .", []testToken{
			{tokenAnonBNode, ""},
			{tokenIRIRel, "a"},
			{tokenIRIRel, "b"},
			{tokenDot, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{"[\t ]", []testToken{
			{tokenAnonBNode, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`[] foaf:knows [ foaf:name "Bob" ] .`, []testToken{
			{tokenAnonBNode, ""},
			{tokenPrefixLabel, "foaf"},
			{tokenIRISuffix, "knows"},
			{tokenPropertyListStart, ""},
			{tokenPrefixLabel, "foaf"},
			{tokenIRISuffix, "name"},
			{tokenLiteral, "Bob"},
			{tokenPropertyListEnd, ""},
			{tokenDot, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{"( .99 1 -2 3.14 4.2e9 )", []testToken{
			{tokenCollectionStart, ""},
			{tokenLiteralDecimal, ".99"},
			{tokenLiteralInteger, "1"},
			{tokenLiteralInteger, "-2"},
			{tokenLiteralDecimal, "3.14"},
			{tokenLiteralDouble, "4.2e9"},
			{tokenCollectionEnd, ""},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
		{`1+2`, []testToken{
			{tokenError, "bad literal: illegal number syntax (number followed by '+')"}},
		},
		{`0.99a`, []testToken{
			{tokenError, "bad literal: illegal number syntax (number followed by 'a')"}},
		},
		{"<s> <p> 1, 2, 3", []testToken{
			{tokenIRIRel, "s"},
			{tokenIRIRel, "p"},
			{tokenLiteralInteger, "1"},
			{tokenComma, ","},
			{tokenLiteralInteger, "2"},
			{tokenComma, ","},
			{tokenLiteralInteger, "3"},
			{tokenEOL, ""},
			{tokenEOF, ""}},
		},
	}

	for _, tt := range lexTests {
		lex := newLexer(strings.NewReader(tt.in))
		res := []testToken{}
		for _, t := range collect(lex) {
			res = append(res, t)
		}

		if !equalTokens(tt.want, res) {
			t.Errorf("lexing %q, got:\n\t%v\nexpected:\n\t%v", tt.in, res, tt.want)
		}
	}
}
