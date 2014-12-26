package parse

import (
	"reflect"
	"strings"
	"testing"

	"github.com/knakk/rdf"
)

func TestNTriples(t *testing.T) {
	tests := []struct {
		in   string
		err  string
		want rdf.Triple
	}{
		{`<s:/s> <s:/p> <s:/o> .`, "",
			rdf.Triple{
				Subj: rdf.URI{URI: "s:/s"},
				Pred: rdf.URI{URI: "s:/p"},
				Obj:  rdf.URI{URI: "s:/o"}}},
		{`<s:/s> <s:/p> "s:/o"^^<s:/xyz> .`, "",
			rdf.Triple{
				Subj: rdf.URI{URI: "s:/s"},
				Pred: rdf.URI{URI: "s:/p"},
				Obj:  rdf.Literal{Val: "s:/o", DataType: rdf.URI{URI: "s:/xyz"}}}},
		{`<s:/s> <s:/p> "3.14"^^<http://www.w3.org/2001/XMLSchema#float> .`, "",
			rdf.Triple{
				Subj: rdf.URI{URI: "s:/s"},
				Pred: rdf.URI{URI: "s:/p"},
				Obj:  rdf.Literal{Val: 3.14, DataType: rdf.XSDFloat}}},
	}

	for _, tt := range tests {
		d := NewNTDecoder(strings.NewReader(tt.in))
		tr, err := d.Decode()
		if tt.err != "" && err.Error() != tt.err {
			t.Errorf("parsing %q, got error %v, wanted %v", tt.in, err, tt.err)
			continue
		}
		if err != nil {
			t.Errorf("parsing %q, got error %v, wanted no error", tt.in, err)
			continue
		}
		if !reflect.DeepEqual(tr, tt.want) {
			t.Errorf("parsing %q got:\n\t%v\nwanted:\n\t%v", tt.in, tr, tt.want)
		}
	}
}
