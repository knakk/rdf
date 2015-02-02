package rdf

import (
	"fmt"
	"testing"
	"time"
)

func TestIRI(t *testing.T) {
	var errTests = []struct {
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

func TestTermTypeLiteral(t *testing.T) {
	_, err := NewLiteral([]int{1, 2, 3})
	if err == nil {
		t.Errorf("Expected an error creating Literal, got nil")
	}

	l1, _ := NewLiteral(42)
	l2, _ := NewLiteral(42.00001)
	l3, _ := NewLiteral(true)
	l4, _ := NewLiteral(false)
	l5 := NewLangLiteral("fisk", "nno")
	l6 := NewLangLiteral("fisk", "no")
	l7, _ := NewLiteral("fisk")
	l8, _ := NewLiteral(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))

	var formatTests = []struct {
		l    Literal
		want string
	}{
		{l1, "42"},
		{l2, "42.00001"},
		{l3, "true"},
		{l4, "false"},
		{l5, `"fisk"@nno`},
		{l6, `"fisk"@no`},
		{l7, `"fisk"`},
		{l8, `"2009-11-10T23:00:00Z"^^<http://www.w3.org/2001/XMLSchema#dateTime>`},
	}

	for _, tt := range formatTests {
		if tt.l.String() != tt.want {
			t.Errorf("Literal string formatting \"%v\", want \"%v\"", tt.want, tt.l.String())
		}
	}
}
