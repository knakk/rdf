package rdf

import (
	"bytes"
	"fmt"
	"testing"
)

func TestIRI(t *testing.T) {
	errTests := []struct {
		input string
		want  string
	}{
		{"", "empty IRI"},
		{"http://dott\ncom", "disallowed character: '\\n'"},
		{"<a>", "disallowed character: '<'"},
		{"here are spaces", "disallowed character: ' '"},
		{"myscheme://abc/xyz/伝言/æøå#hei?f=88", "<nil>"},
	}

	for _, tt := range errTests {
		_, err := NewIRI(tt.input)
		if fmt.Sprintf("%v", err) != tt.want {
			t.Errorf("NewURI(%q) => %v; want %v", tt.input, err, tt.want)
		}
	}

}

func TestLiteral(t *testing.T) {
	inferTypeTests := []struct {
		input     interface{}
		dt        IRI
		errString string
	}{
		{1, xsdInteger, ""},
		{int64(1), xsdInteger, ""},
		{int32(1), xsdInteger, ""},
		{3.14, xsdDouble, ""},
		{float32(3.14), xsdDouble, ""},
		{float64(3.14), xsdDouble, ""},
		{true, xsdBoolean, ""},
		{false, xsdBoolean, ""},
		{"a", xsdString, ""},
		{[]byte("123"), xsdByte, ""},
		{struct{ a, b string }{"1", "2"}, IRI{}, `cannot infer XSD datatype from struct { a string; b string }{a:"1", b:"2"}`},
	}

	for _, tt := range inferTypeTests {
		l, err := NewLiteral(tt.input)
		if err != nil {
			if tt.errString == "" {
				t.Errorf("NewLiteral(%#v) failed with %v; want no error", tt.input, err)
				continue
			}
			if tt.errString != err.Error() {
				t.Errorf("NewLiteral(%#v) failed with %v; want %v", tt.input, err, tt.errString)
				continue
			}
		}
		if err == nil && tt.errString != "" {
			t.Errorf("NewLiteral(%#v) => <no error>; want error %v", tt.input, tt.errString)
			continue
		}
		if l.DataType != tt.dt {
			t.Errorf("NewLiteral(%#v).DataType => %v; want %v", tt.input, l.DataType, tt.dt)
		}
	}

	langTagTests := []struct {
		tag     string
		errWant string
	}{
		{"en", ""},
		{"en-GB", ""},
		{"nb-no2", ""},
		{"no-no-a", "invalid language tag: only one '-' allowed"},
		{"1", "invalid language tag: unexpected character: '1'"},
		{"fr-ø", "invalid language tag: unexpected character: 'ø'"},
		{"en-", "invalid language tag: trailing '-' disallowed"},
		{"-en", "invalid language tag: must start with a letter"},
	}
	for _, tt := range langTagTests {
		_, err := NewLangLiteral("string", tt.tag)
		if err != nil {
			if tt.errWant == "" {
				t.Errorf("NewLangLiteral(\"string\", %#v) failed with %v; want no error", tt.tag, err)
				continue
			}
			if tt.errWant != err.Error() {
				t.Errorf("NewLangLiteral(\"string\", %#v) failed with %v; want %v", tt.tag, err, tt.errWant)
				continue
			}
		}
		if err == nil && tt.errWant != "" {
			t.Errorf("NewLangLiteral(\"string\", %#v) => <no error>; want error %v", tt.tag, tt.errWant)
			continue
		}

	}
}

func TestNTSerialization(t *testing.T) {
	tests := []struct {
		t   Triple
		out string
	}{
		{
			Triple{Subj: IRI{IRI: "http://example/s"}, Pred: IRI{IRI: "http://example/p"}, Obj: IRI{IRI: "http://example/o"}},
			`<http://example/s> <http://example/p> <http://example/o> .
`,
		},
		{
			Triple{
				Subj: IRI{IRI: "http://example/æøå"},
				Pred: IRI{IRI: "http://example/禅"},
				Obj:  Literal{Val: "\"\\\r\n Здра́вствуйте\t☺", DataType: xsdString},
			},
			`<http://example/æøå> <http://example/禅> "\"\\\r\n Здра́вствуйте	☺" .
`,
		},
		{
			Triple{Subj: Blank{ID: "he"}, Pred: IRI{IRI: "http://xmlns.com/foaf/0.1/knows"}, Obj: Blank{ID: "she"}},
			`_:he <http://xmlns.com/foaf/0.1/knows> _:she .
`,
		},
		{
			Triple{
				Subj: IRI{IRI: "http://example/s"},
				Pred: IRI{IRI: "http://example/p"},
				Obj:  Literal{Val: 1, DataType: xsdInteger},
			},
			`<http://example/s> <http://example/p> "1"^^<http://www.w3.org/2001/XMLSchema#integer> .
`,
		},
		{
			Triple{
				Subj: IRI{IRI: "http://example/s"},
				Pred: IRI{IRI: "http://example/p"},
				Obj:  Literal{Val: "bonjour", DataType: xsdString, Lang: "fr"},
			},
			`<http://example/s> <http://example/p> "bonjour"@fr .
`,
		},
	}

	for _, tt := range tests {
		s := tt.t.Serialize(FormatNT)
		if s != tt.out {
			t.Errorf("Serializing %v, \ngot:\n\t%s\nwant:\n\t%s", tt.t, s, tt.out)
		}
	}

	triples := Triples{
		Triple{
			Subj: IRI{IRI: "http://example.org/resource1"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  IRI{IRI: "http://example.org/resource2"},
		},
		Triple{
			Subj: Blank{ID: "anon"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  IRI{IRI: "http://example.org/resource2"},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource2"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Blank{ID: "anon"},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource3"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  IRI{IRI: "http://example.org/resource2"},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource4"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  IRI{IRI: "http://example.org/resource2"},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource5"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  IRI{IRI: "http://example.org/resource2"},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource6"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  IRI{IRI: "http://example.org/resource2"},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource7"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "simple literal", DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource8"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: `backslash:\`, DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource9"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: `dquote:"`, DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource10"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "newline:\n", DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource11"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "return\r", DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource12"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "tab:\t", DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource13"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  IRI{IRI: "http://example.org/resource2"},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource14"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "x", DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource15"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Blank{ID: "anon"},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource16"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "é", DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource17"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "€", DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource21"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "", DataType: IRI{IRI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource22"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: " ", DataType: IRI{IRI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource23"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "x", DataType: IRI{IRI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource23"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: `"`, DataType: IRI{IRI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource24"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "<a></a>", DataType: IRI{IRI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource25"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "a <b></b>", DataType: IRI{IRI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource26"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "a <b></b> c", DataType: IRI{IRI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource26"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "a\n<b></b>\nc", DataType: IRI{IRI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource27"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "chat", DataType: IRI{IRI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource30"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "chat", Lang: "fr", DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource31"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "chat", Lang: "en", DataType: xsdString},
		},
		Triple{
			Subj: IRI{IRI: "http://example.org/resource32"},
			Pred: IRI{IRI: "http://example.org/property"},
			Obj:  Literal{Val: "abc", DataType: IRI{IRI: "http://example.org/datatype1"}},
		},
	}
	var buf bytes.Buffer
	want := `<http://example.org/resource1> <http://example.org/property> <http://example.org/resource2> .
_:anon <http://example.org/property> <http://example.org/resource2> .
<http://example.org/resource2> <http://example.org/property> _:anon .
<http://example.org/resource3> <http://example.org/property> <http://example.org/resource2> .
<http://example.org/resource4> <http://example.org/property> <http://example.org/resource2> .
<http://example.org/resource5> <http://example.org/property> <http://example.org/resource2> .
<http://example.org/resource6> <http://example.org/property> <http://example.org/resource2> .
<http://example.org/resource7> <http://example.org/property> "simple literal" .
<http://example.org/resource8> <http://example.org/property> "backslash:\\" .
<http://example.org/resource9> <http://example.org/property> "dquote:\"" .
<http://example.org/resource10> <http://example.org/property> "newline:\n" .
<http://example.org/resource11> <http://example.org/property> "return\r" .
<http://example.org/resource12> <http://example.org/property> "tab:	" .
<http://example.org/resource13> <http://example.org/property> <http://example.org/resource2> .
<http://example.org/resource14> <http://example.org/property> "x" .
<http://example.org/resource15> <http://example.org/property> _:anon .
<http://example.org/resource16> <http://example.org/property> "é" .
<http://example.org/resource17> <http://example.org/property> "€" .
<http://example.org/resource21> <http://example.org/property> ""^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource22> <http://example.org/property> " "^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource23> <http://example.org/property> "x"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource23> <http://example.org/property> "\""^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource24> <http://example.org/property> "<a></a>"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource25> <http://example.org/property> "a <b></b>"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource26> <http://example.org/property> "a <b></b> c"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource26> <http://example.org/property> "a\n<b></b>\nc"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource27> <http://example.org/property> "chat"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource30> <http://example.org/property> "chat"@fr .
<http://example.org/resource31> <http://example.org/property> "chat"@en .
<http://example.org/resource32> <http://example.org/property> "abc"^^<http://example.org/datatype1> .
`
	err := triples.Serialize(&buf, FormatNT)
	if err != nil {
		t.Fatalf("Serializing N-Triples to io.Writer failed: %v", err)
	}
	if buf.String() != want {
		t.Errorf("Serializing N-Triples:\n%v\ngot:\n%v\nwant:%v", triples, buf.String(), want)
	}
}
