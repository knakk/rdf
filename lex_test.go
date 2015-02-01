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

func benchmarkTokenizerEx(i int, b *testing.B) {
	input := ttlBenchInputs[i]
	for n := 0; n < b.N; n++ {
		lex := newLexer(strings.NewReader(input))
		for _ = range collect(lex) {
		}
	}
}

func BenchmarkTokenizerEx1(b *testing.B)  { benchmarkTokenizerEx(0, b) }
func BenchmarkTokenizerEx2(b *testing.B)  { benchmarkTokenizerEx(1, b) }
func BenchmarkTokenizerEx3(b *testing.B)  { benchmarkTokenizerEx(2, b) }
func BenchmarkTokenizerEx4(b *testing.B)  { benchmarkTokenizerEx(3, b) }
func BenchmarkTokenizerEx5(b *testing.B)  { benchmarkTokenizerEx(4, b) }
func BenchmarkTokenizerEx6(b *testing.B)  { benchmarkTokenizerEx(5, b) }
func BenchmarkTokenizerEx7(b *testing.B)  { benchmarkTokenizerEx(6, b) }
func BenchmarkTokenizerEx8(b *testing.B)  { benchmarkTokenizerEx(7, b) }
func BenchmarkTokenizerEx9(b *testing.B)  { benchmarkTokenizerEx(8, b) }
func BenchmarkTokenizerEx10(b *testing.B) { benchmarkTokenizerEx(9, b) }
func BenchmarkTokenizerEx11(b *testing.B) { benchmarkTokenizerEx(10, b) }
func BenchmarkTokenizerEx12(b *testing.B) { benchmarkTokenizerEx(11, b) }
func BenchmarkTokenizerEx13(b *testing.B) { benchmarkTokenizerEx(12, b) }
func BenchmarkTokenizerEx14(b *testing.B) { benchmarkTokenizerEx(13, b) }
func BenchmarkTokenizerEx15(b *testing.B) { benchmarkTokenizerEx(14, b) }
func BenchmarkTokenizerEx16(b *testing.B) { benchmarkTokenizerEx(15, b) }
func BenchmarkTokenizerEx17(b *testing.B) { benchmarkTokenizerEx(16, b) }
func BenchmarkTokenizerEx18(b *testing.B) { benchmarkTokenizerEx(17, b) }
func BenchmarkTokenizerEx19(b *testing.B) { benchmarkTokenizerEx(18, b) }
func BenchmarkTokenizerEx20(b *testing.B) { benchmarkTokenizerEx(19, b) }
func BenchmarkTokenizerEx21(b *testing.B) { benchmarkTokenizerEx(20, b) }
func BenchmarkTokenizerEx22(b *testing.B) { benchmarkTokenizerEx(21, b) }
func BenchmarkTokenizerEx23(b *testing.B) { benchmarkTokenizerEx(22, b) }
func BenchmarkTokenizerEx24(b *testing.B) { benchmarkTokenizerEx(23, b) }
func BenchmarkTokenizerEx25(b *testing.B) { benchmarkTokenizerEx(24, b) }
func BenchmarkTokenizerEx26(b *testing.B) { benchmarkTokenizerEx(25, b) }
func BenchmarkTokenizerEx27(b *testing.B) { benchmarkTokenizerEx(26, b) }
func BenchmarkTokenizerEx28(b *testing.B) { benchmarkTokenizerEx(27, b) }
func BenchmarkTokenizerEx29(b *testing.B) { benchmarkTokenizerEx(28, b) }

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
			{tokenBNode, "a_BlankLabel123"},
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
			{tokenBNode, "a-b.c"},
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
		{`p:s <http://a.example/p> <http://a.example/o>`, []testToken{
			{tokenPrefixLabel, "p"},
			{tokenIRISuffix, "s"},
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
		for _, t := range collect(lex) {
			res = append(res, t)
		}

		if !equalTokens(tt.want, res) {
			t.Fatalf("lexing %q, got:\n\t%v\nexpected:\n\t%v", tt.in, res, tt.want)
		}
	}
}
