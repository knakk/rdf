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
		{`<s> <p> <o> .`, "",
			rdf.Triple{
				Subj: rdf.URI{URI: "s"},
				Pred: rdf.URI{URI: "p"},
				Obj:  rdf.URI{URI: "o"}}},
		{`<s> <p> "o"^^<xyz> .`, "",
			rdf.Triple{
				Subj: rdf.URI{URI: "s"},
				Pred: rdf.URI{URI: "p"},
				Obj:  rdf.Literal{Val: "o", DataType: rdf.URI{URI: "xyz"}}}},
		{`<s> <p> "3.14"^^<http://www.w3.org/2001/XMLSchema#float> .`, "",
			rdf.Triple{
				Subj: rdf.URI{URI: "s"},
				Pred: rdf.URI{URI: "p"},
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
