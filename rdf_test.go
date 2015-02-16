package rdf

import (
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
