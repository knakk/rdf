package rdf

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
	tokenLiteral3:          "Literal (triple-quoted string)",
	tokenLiteralInteger:    "Literal (integer shorthand syntax)",
	tokenLiteralDouble:     "Literal (double shorthand syntax)",
	tokenLiteralDecimal:    "Literal (decimal shorthand syntax)",
	tokenLiteralBoolean:    "Literal (boolean shorthand syntax)",
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
			{tokenEOF, ""}},
		},
		{" \n ", []testToken{
			{tokenEOF, ""}},
		},
		{"<a>", []testToken{
			{tokenIRIRel, "a"},
			{tokenEOF, ""}},
		},
		{"<http://www.w3.org/1999/02/22-rdf-syntax-ns#type>", []testToken{
			{tokenIRIAbs, "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			{tokenEOF, ""}},
		},
		{`<s:1> a <o:1>`, []testToken{
			{tokenIRIAbs, "s:1"},
			{tokenRDFType, "a"},
			{tokenIRIAbs, "o:1"},
			{tokenEOF, ""}},
		},
		{`<a>a<b>.`, []testToken{
			{tokenIRIRel, "a"},
			{tokenRDFType, "a"},
			{tokenIRIRel, "b"},
			{tokenDot, ""},
			{tokenEOF, ""}},
		},
		{"<a>\r<b>\r\n<c>\n.", []testToken{
			{tokenIRIRel, "a"},
			{tokenIRIRel, "b"},
			{tokenIRIRel, "c"},
			{tokenDot, ""},
			{tokenEOF, ""}},
		},
		{`  <x><y> <z>  <\u0053> `, []testToken{
			{tokenIRIRel, "x"},
			{tokenIRIRel, "y"},
			{tokenIRIRel, "z"},
			{tokenIRIRel, "S"},
			{tokenEOF, ""}},
		},
		{`<s><p><o>.#comment`, []testToken{
			{tokenIRIRel, "s"},
			{tokenIRIRel, "p"},
			{tokenIRIRel, "o"},
			{tokenDot, ""},
			{tokenEOF, ""}},
		},
		{`""`, []testToken{
			{tokenLiteral, ""},
			{tokenEOF, ""}},
		},
		{`"a"`, []testToken{
			{tokenLiteral, "a"},
			{tokenEOF, ""}},
		},
		{"'a'", []testToken{
			{tokenLiteral, "a"},
			{tokenEOF, ""}},
		},
		{`"""""" """a"""`, []testToken{
			{tokenLiteral3, ""},
			{tokenLiteral3, "a"},
			{tokenEOF, ""}},
		},
		{`'''xyz'''`, []testToken{
			{tokenLiteral3, "xyz"},
			{tokenEOF, ""}},
		},
		{`'''a''b'''`, []testToken{
			{tokenLiteral3, "a''b"},
			{tokenEOF, ""}},
		},
		{`"""multi
line
string"""`, []testToken{
			{tokenLiteral3, "multi\nline\nstring"},
			{tokenEOF, ""}},
		},
		{`'''one
two
3'''`, []testToken{
			{tokenLiteral3, "one\ntwo\n3"},
			{tokenEOF, ""}},
		},
		{`"æøå üçgen こんにちは" # comments text`, []testToken{
			{tokenLiteral, "æøå üçgen こんにちは"},
			{tokenEOF, ""}},
		},
		{`"KI\u0053\U00000053ki⚡⚡"`, []testToken{
			{tokenLiteral, "KISSki⚡⚡"},
			{tokenEOF, ""}},
		},
		{`"she said: \"hi\""`, []testToken{
			{tokenLiteral, `she said: "hi"`},
			{tokenEOF, ""}},
		},
		{`"escapes:\t\\\"\n\t"`, []testToken{
			{tokenLiteral, `escapes:	\"
	`},
			{tokenEOF, ""}},
		},
		{`"hei"@nb-no "hi"@en #language tags`, []testToken{
			{tokenLiteral, "hei"},
			{tokenLangMarker, "@"},
			{tokenLang, "nb-no"},
			{tokenLiteral, "hi"},
			{tokenLangMarker, "@"},
			{tokenLang, "en"},
			{tokenEOF, ""}},
		},
		{`"hei"@`, []testToken{
			{tokenLiteral, "hei"},
			{tokenLangMarker, "@"},
			{tokenError, "bad literal: invalid language tag"}},
		},
		{`"Neishe"@zh-latn-pinyin-x-notone`, []testToken{
			{tokenLiteral, "Neishe"},
			{tokenLangMarker, "@"},
			{tokenLang, "zh-latn-pinyin-x-notone"},
			{tokenEOF, ""}},
		},
		{`"a"^^<s://mydatatype>`, []testToken{
			{tokenLiteral, "a"},
			{tokenDataTypeMarker, "^^"},
			{tokenIRIAbs, "s://mydatatype"},
			{tokenEOF, ""}},
		},
		{`"1"^^xsd:integer`, []testToken{
			{tokenLiteral, "1"},
			{tokenDataTypeMarker, "^^"},
			{tokenPrefixLabel, "xsd"},
			{tokenIRISuffix, "integer"},
			{tokenEOF, ""}},
		},
		{`"a"^`, []testToken{
			{tokenLiteral, "a"},
			{tokenError, "bad literal: invalid datatype IRI"}},
		},
		{`"a"^^`, []testToken{
			{tokenLiteral, "a"},
			{tokenDataTypeMarker, "^^"},
			{tokenEOF, ""}},
		},
		{`"a"^^xyz`, []testToken{
			{tokenLiteral, "a"},
			{tokenDataTypeMarker, "^^"},
			{tokenError, `illegal token: "xyz"`}},
		},
		{`_:a_BlankLabel123.`, []testToken{
			{tokenBNode, "_:a_BlankLabel123"},
			{tokenDot, ""},
			{tokenEOF, ""}},
		},
		{`<s> <p> true; <p2> false.`, []testToken{
			{tokenIRIRel, "s"},
			{tokenIRIRel, "p"},
			{tokenLiteralBoolean, "true"},
			{tokenSemicolon, ";"},
			{tokenIRIRel, "p2"},
			{tokenLiteralBoolean, "false"},
			{tokenDot, ""},
			{tokenEOF, ""}},
		},
		{`_:a-b.c.`, []testToken{
			{tokenBNode, "_:a-b.c"},
			{tokenDot, ""},
			{tokenEOF, ""}},
		},
		{`#comment
		  <s><p><o>.#comment
		  # comment

		  <s><p2> "yo"
		  ####
		  `, []testToken{
			{tokenIRIRel, "s"},
			{tokenIRIRel, "p"},
			{tokenIRIRel, "o"},
			{tokenDot, ""},
			{tokenIRIRel, "s"},
			{tokenIRIRel, "p2"},
			{tokenLiteral, "yo"},
			{tokenEOF, ""}},
		},
		{`@prefix a: <http:/a.org/>.`, []testToken{
			{tokenPrefix, "prefix"},
			{tokenPrefixLabel, "a"},
			{tokenIRIAbs, "http:/a.org/"},
			{tokenDot, ""},
			{tokenEOF, ""}},
		},
		{`p:D\.C\.`, []testToken{
			{tokenPrefixLabel, "p"},
			{tokenIRISuffix, "D.C."},
			{tokenEOF, ""}},
		},
		{`p:s <http://a.example/p> <http://a.example/o>`, []testToken{
			{tokenPrefixLabel, "p"},
			{tokenIRISuffix, "s"},
			{tokenIRIAbs, "http://a.example/p"},
			{tokenIRIAbs, "http://a.example/o"},
			{tokenEOF, ""}},
		},
		{`p:	<http://a.example/p> <http://a.example/o>`, []testToken{
			{tokenPrefixLabel, "p"},
			{tokenIRISuffix, ""},
			{tokenIRIAbs, "http://a.example/p"},
			{tokenIRIAbs, "http://a.example/o"},
			{tokenEOF, ""}},
		},
		{"@base <http:/a.org/>.", []testToken{
			{tokenBase, "base"},
			{tokenIRIAbs, "http:/a.org/"},
			{tokenDot, ""},
			{tokenEOF, ""}},
		},
		{"BASE <http:/a.org/>", []testToken{
			{tokenSparqlBase, "BASE"},
			{tokenIRIAbs, "http:/a.org/"},
			{tokenEOF, ""}},
		},
		{"basE <a>\nPReFiX p:<b>", []testToken{
			{tokenSparqlBase, "basE"},
			{tokenIRIRel, "a"},
			{tokenSparqlPrefix, "PReFiX"},
			{tokenPrefixLabel, "p"},
			{tokenIRIRel, "b"},
			{tokenEOF, ""}},
		},
		{"[] <a> <b> .", []testToken{
			{tokenAnonBNode, ""},
			{tokenIRIRel, "a"},
			{tokenIRIRel, "b"},
			{tokenDot, ""},
			{tokenEOF, ""}},
		},
		{"[\t ]", []testToken{
			{tokenAnonBNode, ""},
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
			{tokenEOF, ""}},
		},
		{"( .99. 1, -2 3.14 4.2e9. )", []testToken{
			{tokenCollectionStart, ""},
			{tokenLiteralDecimal, ".99"},
			{tokenDot, ""},
			{tokenLiteralInteger, "1"},
			{tokenComma, ","},
			{tokenLiteralInteger, "-2"},
			{tokenLiteralDecimal, "3.14"},
			{tokenLiteralDouble, "4.2e9"},
			{tokenDot, ""},
			{tokenCollectionEnd, ""},
			{tokenEOF, ""}},
		},
		{`123e`, []testToken{
			{tokenError, "bad literal: illegal number syntax: missing exponent"}},
		},
		{`1+2`, []testToken{
			{tokenError, "bad literal: illegal number syntax (number followed by '+')"}},
		},
		{`0.99a`, []testToken{
			{tokenError, "bad literal: illegal number syntax (number followed by 'a')"}},
		},
		{"<s> <p> 1, 2, 3.", []testToken{
			{tokenIRIRel, "s"},
			{tokenIRIRel, "p"},
			{tokenLiteralInteger, "1"},
			{tokenComma, ","},
			{tokenLiteralInteger, "2"},
			{tokenComma, ","},
			{tokenLiteralInteger, "3"},
			{tokenDot, ""},
			{tokenEOF, ""}},
		},
	}

	for _, tt := range lexTests {
		lex := newLexer(strings.NewReader(tt.in))
		res := []testToken{}
		res = append(res, collect(lex)...)

		if !equalTokens(tt.want, res) {
			t.Fatalf("lexing %q, got:\n\t%v\nexpected:\n\t%v", tt.in, res, tt.want)
		}
	}
}
