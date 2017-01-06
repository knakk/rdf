package rdf

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

var defaultGraph = Blank{id: "_:defaultGraph"}

func BenchmarkDecodeNQ(b *testing.B) {
	input := "#comment\n<http://example/s> <http://example/p> \"123\"^^<http://www.w3.org/2001/XMLSchema#integer> <http://example/g>"
	for n := 0; n < b.N; n++ {
		dec := NewQuadDecoder(bytes.NewBufferString(input), NQuads)
		for _, err := dec.Decode(); err != io.EOF; _, err = dec.Decode() {
		}
	}
	b.SetBytes(int64(len(input)))
}

func TestNQ(t *testing.T) {
	for _, test := range nqTestSuite {
		dec := NewQuadDecoder(bytes.NewBufferString(test.input), NQuads)
		quads, err := dec.DecodeAll()
		if test.errWant != "" && err == nil {
			t.Errorf("parseNQ(%s) => <no error>, want %q", test.input, test.errWant)
			continue
		}

		if test.errWant != "" && err != nil {
			if !strings.HasSuffix(err.Error(), test.errWant) {
				t.Errorf("parseNQ(%s) => %v, want %q", test.input, err.Error(), test.errWant)
			}
			continue
		}

		if test.errWant == "" && err != nil {
			t.Errorf("parseNQ(%s) => %v, want %q", test.input, err.Error(), test.want)
			continue
		}

		if !reflect.DeepEqual(quads, test.want) {
			t.Errorf("parseNQ(%s) => %v, want %v", test.input, quads, test.want)
		}
	}
}

// nqTestSuite is a representation of the official W3C test suite for N-Quads
// which is found at: http://www.w3.org/2013/N-QuadsTests/
var nqTestSuite = []struct {
	input   string
	errWant string
	want    []Quad
}{
	//<#nq-syntax-uri-01> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-uri-01" ;
	//   rdfs:comment "IRI graph with IRI triple" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-uri-01.nq> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o> <http://example/g> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			IRI{str: "http://example/g"},
		},
	}},

	//<#nq-syntax-uri-02> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-uri-02" ;
	//   rdfs:comment "IRI graph with BNode subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-uri-02.nq> ;
	//   .

	{`_:s <http://example/p> <http://example/o> <http://example/g> .`, "", []Quad{
		Quad{
			Triple{
				Subj: Blank{id: "_:s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			IRI{str: "http://example/g"},
		},
	}},

	//<#nq-syntax-uri-03> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-uri-03" ;
	//   rdfs:comment "IRI graph with BNode object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-uri-03.nq> ;
	//   .

	{`<http://example/s> <http://example/p> _:o <http://example/g> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Blank{id: "_:o"},
			},
			IRI{str: "http://example/g"},
		},
	}},

	//<#nq-syntax-uri-04> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-uri-04" ;
	//   rdfs:comment "IRI graph with simple literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-uri-04.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "o" <http://example/g> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "o", DataType: xsdString},
			},
			IRI{str: "http://example/g"},
		},
	}},

	//<#nq-syntax-uri-05> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-uri-05" ;
	//   rdfs:comment "IRI graph with language tagged literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-uri-05.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "o"@en <http://example/g> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "o", lang: "en", DataType: rdfLangString},
			},
			IRI{str: "http://example/g"},
		},
	}},

	//<#nq-syntax-uri-06> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-uri-06" ;
	//   rdfs:comment "IRI graph with datatyped literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-uri-06.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "o"^^<http://www.w3.org/2001/XMLSchema#string> <http://example/g> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "o", DataType: xsdString},
			},
			IRI{str: "http://example/g"},
		},
	}},

	//<#nq-syntax-bnode-01> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-bnode-01" ;
	//   rdfs:comment "BNode graph with IRI triple" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bnode-01.nq> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o> _:g .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			Blank{id: "_:g"},
		},
	}},

	//<#nq-syntax-bnode-02> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-bnode-02" ;
	//   rdfs:comment "BNode graph with BNode subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bnode-02.nq> ;
	//   .

	{`_:s <http://example/p> <http://example/o> _:g .`, "", []Quad{
		Quad{
			Triple{
				Subj: Blank{id: "_:s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			Blank{id: "_:g"},
		},
	}},

	//<#nq-syntax-bnode-03> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-bnode-03" ;
	//   rdfs:comment "BNode graph with BNode object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bnode-03.nq> ;
	//   .

	{`<http://example/s> <http://example/p> _:o _:g .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Blank{id: "_:o"},
			},
			Blank{id: "_:g"},
		},
	}},

	//<#nq-syntax-bnode-04> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-bnode-04" ;
	//   rdfs:comment "BNode graph with simple literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bnode-04.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "o" _:g .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "o", DataType: xsdString},
			},
			Blank{id: "_:g"},
		},
	}},

	//<#nq-syntax-bnode-05> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-bnode-05" ;
	//   rdfs:comment "BNode graph with language tagged literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bnode-05.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "o"@en _:g .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "o", lang: "en", DataType: rdfLangString},
			},
			Blank{id: "_:g"},
		},
	}},

	//<#nq-syntax-bnode-06> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nq-syntax-bnode-06" ;
	//   rdfs:comment "BNode graph with datatyped literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bnode-06.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "o"^^<http://www.w3.org/2001/XMLSchema#string> _:g .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "o", DataType: xsdString},
			},
			Blank{id: "_:g"},
		},
	}},

	//<#nq-syntax-bad-literal-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nq-syntax-bad-literal-01" ;
	//   rdfs:comment "Graph name may not be a simple literal (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bad-literal-01.nq> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o> "o" .`,
		"unexpected Literal as graph", []Quad{}},

	//<#nq-syntax-bad-literal-02> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nq-syntax-bad-literal-02" ;
	//   rdfs:comment "Graph name may not be a language tagged literal (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bad-literal-02.nq> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o> "o"@en .`,
		"unexpected Literal as graph", []Quad{}},

	//<#nq-syntax-bad-literal-03> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nq-syntax-bad-literal-03" ;
	//   rdfs:comment "Graph name may not be a datatyped literal (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bad-literal-03.nq> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o> "o"^^<http://www.w3.org/2001/XMLSchema#string> .`,
		"unexpected Literal as graph", []Quad{}},

	//<#nq-syntax-bad-uri-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nq-syntax-bad-uri-01" ;
	//   rdfs:comment "Graph name IRI must be absolute (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bad-uri-01.nq> ;
	//   .

	{`# No relative IRIs in N-Quads
<http://example/s> <http://example/p> <http://example/o> <g>.`,
		"unexpected IRI (relative) as graph", []Quad{}},

	//<#nq-syntax-bad-quint-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nq-syntax-bad-quint-01" ;
	//   rdfs:comment "N-Quads does not have a fifth element (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nq-syntax-bad-quint-01.nq> ;
	//   .

	{`# N-Quads rejects a quint
<http://example/s> <http://example/p> <http://example/o> <http://example/g> <http://example/n> .`,
		"unexpected IRI (absolute) as dot (.)", []Quad{}},

	//<#nt-syntax-file-01> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-file-01" ;
	//   rdfs:comment "Empty file" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-file-01.nq> ;
	//   .

	{``, "", nil},

	//<#nt-syntax-file-02> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-file-02" ;
	//   rdfs:comment "Only comment" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-file-02.nq> ;
	//   .

	{`#Empty file.`, "", nil},

	//<#nt-syntax-file-03> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-file-03" ;
	//   rdfs:comment "One comment, one empty line" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-file-03.nq> ;
	//   .

	{`#One comment, one empty line.
`, "", nil},

	//<#nt-syntax-uri-01> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-uri-01" ;
	//   rdfs:comment "Only IRIs" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-uri-01.nq> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-uri-02> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-uri-02" ;
	//   rdfs:comment "IRIs with Unicode escape" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-uri-02.nq> ;
	//   .

	{`# x53 is capital S
<http://example/\u0053> <http://example/p> <http://example/o> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/S"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-uri-03> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-uri-03" ;
	//   rdfs:comment "IRIs with long Unicode escape" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-uri-03.nq> ;
	//   .

	{`# x53 is capital S
<http://example/\U00000053> <http://example/p> <http://example/o> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/S"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-uri-04> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-uri-04" ;
	//   rdfs:comment "Legal IRIs" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-uri-04.nq> ;
	//   .

	{`# IRI with all chars in it.
<http://example/s> <http://example/p> <scheme:!$%25&'()*+,-./0123456789:/@ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~?#> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "scheme:!$%25&'()*+,-./0123456789:/@ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~?#"},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-string-01> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-string-01" ;
	//   rdfs:comment "string literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-string-01.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "string" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "string", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-string-02> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-string-02" ;
	//   rdfs:comment "langString literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-string-02.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "string"@en .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "string", DataType: rdfLangString, lang: "en"},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-string-03> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-string-03" ;
	//   rdfs:comment "langString literal with region" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-string-03.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "string"@en-uk .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "string", DataType: rdfLangString, lang: "en-uk"},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-str-esc-01> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-str-esc-01" ;
	//   rdfs:comment "string literal with escaped newline" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-str-esc-01.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "a\n" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "a\n", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-str-esc-02> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-str-esc-02" ;
	//   rdfs:comment "string literal with Unicode escape" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-str-esc-02.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "a\u0020b" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "a b", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-str-esc-03> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-str-esc-03" ;
	//   rdfs:comment "string literal with long Unicode escape" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-str-esc-03.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "a\U00000020b" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "a b", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-bnode-01> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-bnode-01" ;
	//   rdfs:comment "bnode subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bnode-01.nq> ;
	//   .

	{`_:a  <http://example/p> <http://example/o> .`, "", []Quad{
		Quad{
			Triple{
				Subj: Blank{id: "_:a"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-bnode-02> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-bnode-02" ;
	//   rdfs:comment "bnode object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bnode-02.nq> ;
	//   .

	{`<http://example/s> <http://example/p> _:a .
_:a  <http://example/p> <http://example/o> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Blank{id: "_:a"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: Blank{id: "_:a"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-bnode-03> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-bnode-03" ;
	//   rdfs:comment "Blank node labels may start with a digit" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bnode-03.nq> ;
	//   .

	{`<http://example/s> <http://example/p> _:1a .
_:1a  <http://example/p> <http://example/o> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Blank{id: "_:1a"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: Blank{id: "_:1a"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-datatypes-01> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-datatypes-01" ;
	//   rdfs:comment "xsd:byte literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-datatypes-01.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "123"^^<http://www.w3.org/2001/XMLSchema#byte> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "123", DataType: IRI{str: "http://www.w3.org/2001/XMLSchema#byte"}},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-datatypes-02> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-datatypes-02" ;
	//   rdfs:comment "integer as xsd:string" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-datatypes-02.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "123"^^<http://www.w3.org/2001/XMLSchema#string> .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "123", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#nt-syntax-bad-uri-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-01" ;
	//   rdfs:comment "Bad IRI : space (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-uri-01.nq> ;
	//   .

	{`# Bad IRI : space.
<http://example/ space> <http://example/p> <http://example/o> .`, "bad IRI: disallowed character ' '", []Quad{}},

	//<#nt-syntax-bad-uri-02> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-02" ;
	//   rdfs:comment "Bad IRI : bad escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-uri-02.nq> ;
	//   .

	{`# Bad IRI : bad escape
<http://example/\u00ZZ11> <http://example/p> <http://example/o> .`, "bad IRI: insufficent hex digits in unicode escape", []Quad{}},

	//<#nt-syntax-bad-uri-03> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-03" ;
	//   rdfs:comment "Bad IRI : bad long escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-uri-03.nq> ;
	//   .

	{`# Bad IRI : bad escape
<http://example/\U00ZZ1111> <http://example/p> <http://example/o> .`, "bad IRI: insufficent hex digits in unicode escape", []Quad{}},

	//<#nt-syntax-bad-uri-04> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-04" ;
	//   rdfs:comment "Bad IRI : character escapes not allowed (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-uri-04.nq> ;
	//   .

	{`# Bad IRI : character escapes not allowed.
<http://example/\n> <http://example/p> <http://example/o> .`, "bad IRI: disallowed escape character 'n'", []Quad{}},

	//<#nt-syntax-bad-uri-05> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-05" ;
	//   rdfs:comment "Bad IRI : character escapes not allowed (2) (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-uri-05.nq> ;
	//   .

	{`# Bad IRI : character escapes not allowed.
<http://example/\/> <http://example/p> <http://example/o> .`, "bad IRI: disallowed escape character '/'", []Quad{}},

	//<#nt-syntax-bad-uri-06> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-06" ;
	//   rdfs:comment "Bad IRI : relative IRI not allowed in subject (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-uri-06.nq> ;
	//   .

	{`# No relative IRIs in N-Triples
<s> <http://example/p> <http://example/o> .`, "unexpected IRI (relative) as subject", []Quad{}},

	//<#nt-syntax-bad-uri-07> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-07" ;
	//   rdfs:comment "Bad IRI : relative IRI not allowed in predicate (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-uri-07.nq> ;
	//   .

	{`# No relative IRIs in N-Triples
<http://example/s> <p> <http://example/o> .`, "unexpected IRI (relative) as predicate", []Quad{}},

	//<#nt-syntax-bad-uri-08> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-08" ;
	//   rdfs:comment "Bad IRI : relative IRI not allowed in object (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-uri-08.nq> ;
	//   .

	{`# No relative IRIs in N-Triples
<http://example/s> <http://example/p> <o> .`, "unexpected IRI (relative) as object", []Quad{}},

	//<#nt-syntax-bad-uri-09> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-09" ;
	//   rdfs:comment "Bad IRI : relative IRI not allowed in datatype (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-uri-09.nq> ;
	//   .

	{`# No relative IRIs in N-Triples
<http://example/s> <http://example/p> "foo"^^<dt> .`, "unexpected IRI (relative) as literal datatype", []Quad{}},

	//<#nt-syntax-bad-prefix-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-prefix-01" ;
	//   rdfs:comment "@prefix not allowed in n-triples (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-prefix-01.nq> ;
	//   .

	{`@prefix : <http://example/> .`, "unexpected @prefix as subject", []Quad{}},

	//<#nt-syntax-bad-base-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-base-01" ;
	//   rdfs:comment "@base not allowed in N-Triples (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-base-01.nq> ;
	//   .

	{`@base <http://example/> .`, " unexpected @base as subject", []Quad{}},

	//<#nt-syntax-bad-struct-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-struct-01" ;
	//   rdfs:comment "N-Triples does not have objectList (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-struct-01.nq> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o>, <http://example/o2> .`,
		"unexpected Comma as graph", []Quad{}},

	//<#nt-syntax-bad-struct-02> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-struct-02" ;
	//   rdfs:comment "N-Triples does not have predicateObjectList (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-struct-02.nq> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o>; <http://example/p2>, <http://example/o2> .`,
		"unexpected Semicolon as graph", []Quad{}},

	//<#nt-syntax-bad-lang-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-lang-01" ;
	//   rdfs:comment "langString with bad lang (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-lang-01.nq> ;
	//   .

	{`# Bad lang tag
<http://example/s> <http://example/p> "string"@1 .`, "syntax error: bad literal: invalid language tag", []Quad{}},

	//<#nt-syntax-bad-esc-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-esc-01" ;
	//   rdfs:comment "Bad string escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-esc-01.nq> ;
	//   .

	{`# Bad string escape
<http://example/s> <http://example/p> "a\zb" .`, "syntax error: bad literal: disallowed escape character 'z'", []Quad{}},

	//<#nt-syntax-bad-esc-02> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-esc-02" ;
	//   rdfs:comment "Bad string escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-esc-02.nq> ;
	//   .

	{`# Bad string escape
<http://example/s> <http://example/p> "\uWXYZ" .`, "syntax error: bad literal: insufficent hex digits in unicode escape", []Quad{}},

	//<#nt-syntax-bad-esc-03> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-esc-03" ;
	//   rdfs:comment "Bad string escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-esc-03.nq> ;
	//   .

	{`# Bad string escape
<http://example/s> <http://example/p> "\U0000WXYZ" .`, "syntax error: bad literal: insufficent hex digits in unicode escape", []Quad{}},

	//<#nt-syntax-bad-string-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-01" ;
	//   rdfs:comment "mismatching string literal open/close (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-string-01.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "abc' .`, "syntax error: bad literal: no closing quote: '\"'", []Quad{}},

	//<#nt-syntax-bad-string-02> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-02" ;
	//   rdfs:comment "mismatching string literal open/close (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-string-02.nq> ;
	//   .

	{`<http://example/s> <http://example/p> 1.0 .`, "unexpected Literal (decimal shorthand syntax) as object", []Quad{}},

	//<#nt-syntax-bad-string-03> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-03" ;
	//   rdfs:comment "single quotes (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-string-03.nq> ;
	//   .

	{`<http://example/s> <http://example/p> 1.0e1 .`, "unexpected Literal (double shorthand syntax) as object", []Quad{}},

	//<#nt-syntax-bad-string-04> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-04" ;
	//   rdfs:comment "long single string literal (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-string-04.nq> ;
	//   .

	{`<http://example/s> <http://example/p> '''abc''' .`, " unexpected Literal (triple-quoted string) as object", []Quad{}},

	//<#nt-syntax-bad-string-05> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-05" ;
	//   rdfs:comment "long double string literal (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-string-05.nq> ;
	//   .

	{`<http://example/s> <http://example/p> """abc""" .`, "unexpected Literal (triple-quoted string) as object", []Quad{}},

	//<#nt-syntax-bad-string-06> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-06" ;
	//   rdfs:comment "string literal with no end (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-string-06.nq> ;
	//   .

	{`<http://example/s> <http://example/p> "abc .`, "syntax error: bad literal: no closing quote: '\"'", []Quad{}},

	//<#nt-syntax-bad-string-07> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-07" ;
	//   rdfs:comment "string literal with no start (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-string-07.nq> ;
	//   .

	{`<http://example/s> <http://example/p> abc" .`, `syntax error: illegal token: "abc\""`, []Quad{}},

	//<#nt-syntax-bad-num-01> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-num-01" ;
	//   rdfs:comment "no numbers in N-Triples (integer) (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-num-01.nq> ;
	//   .

	{`<http://example/s> <http://example/p> 1 .`, "unexpected Literal (integer shorthand syntax) as object", []Quad{}},

	//<#nt-syntax-bad-num-02> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-num-02" ;
	//   rdfs:comment "no numbers in N-Triples (decimal) (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-num-02.nq> ;
	//   .

	{`<http://example/s> <http://example/p> 1.0 .`, "unexpected Literal (decimal shorthand syntax) as object", []Quad{}},

	//<#nt-syntax-bad-num-03> a rdft:TestNQuadsNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-num-03" ;
	//   rdfs:comment "no numbers in N-Triples (float) (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-bad-num-03.nq> ;
	//   .

	{`<http://example/s> <http://example/p> 1.0e0 .`, "unexpected Literal (double shorthand syntax) as object", []Quad{}},

	//<#nt-syntax-subm-01> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "nt-syntax-subm-01" ;
	//   rdfs:comment "Submission test from Original RDF Test Cases" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nt-syntax-subm-01.nq> ;
	//   .

	{`#
# Copyright World Wide Web Consortium, (Massachusetts Institute of
# Technology, Institut National de Recherche en Informatique et en
# Automatique, Keio University).
#
# All Rights Reserved.
#
# Please see the full Copyright clause at
# <http://www.w3.org/Consortium/Legal/copyright-software.html>
#
# Test file with a variety of legal N-Triples
#
# Dave Beckett - http://purl.org/net/dajobe/
# 
# $Id: test.nt,v 1.7 2003/10/06 15:52:19 dbeckett2 Exp $
# 
#####################################################################

# comment lines
  	  	   # comment line after whitespace
# empty blank line, then one with spaces and tabs

         	
<http://example.org/resource1> <http://example.org/property> <http://example.org/resource2> .
_:anon <http://example.org/property> <http://example.org/resource2> .
<http://example.org/resource2> <http://example.org/property> _:anon .
# spaces and tabs throughout:
 	 <http://example.org/resource3> 	 <http://example.org/property>	 <http://example.org/resource2> 	.	 

# line ending with CR NL (ASCII 13, ASCII 10)
<http://example.org/resource4> <http://example.org/property> <http://example.org/resource2> .

# 2 statement lines separated by single CR (ASCII 10)
<http://example.org/resource5> <http://example.org/property> <http://example.org/resource2> .
<http://example.org/resource6> <http://example.org/property> <http://example.org/resource2> .


# All literal escapes
<http://example.org/resource7> <http://example.org/property> "simple literal" .
<http://example.org/resource8> <http://example.org/property> "backslash:\\" .
<http://example.org/resource9> <http://example.org/property> "dquote:\"" .
<http://example.org/resource10> <http://example.org/property> "newline:\n" .
<http://example.org/resource11> <http://example.org/property> "return\r" .
<http://example.org/resource12> <http://example.org/property> "tab:\t" .

# Space is optional before final .
<http://example.org/resource13> <http://example.org/property> <http://example.org/resource2>.
<http://example.org/resource14> <http://example.org/property> "x".
<http://example.org/resource15> <http://example.org/property> _:anon.

# \u and \U escapes
# latin small letter e with acute symbol \u00E9 - 3 UTF-8 bytes #xC3 #A9
<http://example.org/resource16> <http://example.org/property> "\u00E9" .
# Euro symbol \u20ac  - 3 UTF-8 bytes #xE2 #x82 #xAC
<http://example.org/resource17> <http://example.org/property> "\u20AC" .
# resource18 test removed
# resource19 test removed
# resource20 test removed

# XML Literals as Datatyped Literals
<http://example.org/resource21> <http://example.org/property> ""^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource22> <http://example.org/property> " "^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource23> <http://example.org/property> "x"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource23> <http://example.org/property> "\""^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource24> <http://example.org/property> "<a></a>"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource25> <http://example.org/property> "a <b></b>"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource26> <http://example.org/property> "a <b></b> c"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource26> <http://example.org/property> "a\n<b></b>\nc"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
<http://example.org/resource27> <http://example.org/property> "chat"^^<http://www.w3.org/2000/01/rdf-schema#XMLLiteral> .
# resource28 test removed 2003-08-03
# resource29 test removed 2003-08-03

# Plain literals with languages
<http://example.org/resource30> <http://example.org/property> "chat"@fr .
<http://example.org/resource31> <http://example.org/property> "chat"@en .

# Typed Literals
<http://example.org/resource32> <http://example.org/property> "abc"^^<http://example.org/datatype1> .
# resource33 test removed 2003-08-03`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource1"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  IRI{str: "http://example.org/resource2"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: Blank{id: "_:anon"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  IRI{str: "http://example.org/resource2"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource2"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Blank{id: "_:anon"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource3"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  IRI{str: "http://example.org/resource2"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource4"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  IRI{str: "http://example.org/resource2"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource5"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  IRI{str: "http://example.org/resource2"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource6"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  IRI{str: "http://example.org/resource2"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource7"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "simple literal", DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource8"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: `backslash:\`, DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource9"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: `dquote:"`, DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource10"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "newline:\n", DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource11"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "return\r", DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource12"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "tab:\t", DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource13"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  IRI{str: "http://example.org/resource2"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource14"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "x", DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource15"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Blank{id: "_:anon"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource16"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "é", DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource17"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "€", DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource21"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "", DataType: IRI{str: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource22"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: " ", DataType: IRI{str: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource23"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "x", DataType: IRI{str: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource23"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: `"`, DataType: IRI{str: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource24"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "<a></a>", DataType: IRI{str: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource25"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "a <b></b>", DataType: IRI{str: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource26"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "a <b></b> c", DataType: IRI{str: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource26"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "a\n<b></b>\nc", DataType: IRI{str: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource27"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "chat", DataType: IRI{str: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource30"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "chat", lang: "fr", DataType: rdfLangString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource31"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "chat", lang: "en", DataType: rdfLangString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/resource32"},
				Pred: IRI{str: "http://example.org/property"},
				Obj:  Literal{str: "abc", DataType: IRI{str: "http://example.org/datatype1"}},
			},
			defaultGraph,
		},
	}},

	//<#comment_following_triple> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "comment_following_triple" ;
	//   rdfs:comment "Tests comments after a triple" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <comment_following_triple.nq> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o> . # comment
<http://example/s> <http://example/p> _:o . # comment
<http://example/s> <http://example/p> "o" . # comment
<http://example/s> <http://example/p> "o"^^<http://example/dt> . # comment
<http://example/s> <http://example/p> "o"@en . # comment`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Blank{id: "_:o"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "o", DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "o", DataType: IRI{str: "http://example/dt"}},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "o", lang: "en", DataType: rdfLangString},
			},
			defaultGraph,
		},
	}},

	//<#literal_ascii_boundaries> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_ascii_boundaries" ;
	//   rdfs:comment "literal_ascii_boundaries '\\x00\\x26\\x28...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_ascii_boundaries.nq> ;
	//   .

	{"<http://a.example/s> <http://a.example/p> \"\x00	&([]\" .", "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj: Literal{str: "\x00	&([]", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_UTF8_boundaries> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_UTF8_boundaries" ;
	//   rdfs:comment "literal_with_UTF8_boundaries '\\x80\\x7ff\\x800\\xfff...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_UTF8_boundaries.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_all_controls> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_all_controls" ;
	//   rdfs:comment "literal_all_controls '\\x00\\x01\\x02\\x03\\x04...'" ;
	//   rdft:approval rdft:Approved ;
	//   rdft:approval rdft:Approved ;
	//   mf:action   <literal_all_controls.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\u0000\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u0008\t\u000B\u000C\u000E\u000F\u0010\u0011\u0012\u0013\u0014\u0015\u0016\u0017\u0018\u0019\u001A\u001B\u001C\u001D\u001E\u001F" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "\x00\x01\x02\x03\x04\x05\x06\a\b\t\v\f\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_all_punctuation> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_all_punctuation" ;
	//   rdfs:comment "literal_all_punctuation '!\"#$%&()...'" ;
	//   rdft:approval rdft:Approved ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_all_punctuation.nq> ;
	//   .

	{"<http://a.example/s> <http://a.example/p> \" !\\\"#$%&():;<=>?@[]^_`{|}~\".", "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: " !\"#$%&():;<=>?@[]^_`{|}~", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_squote> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_squote" ;
	//   rdfs:comment "literal with squote \"x'y\"" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_squote.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x'y" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "x'y", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_2_squotes> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_2_squotes" ;
	//   rdfs:comment "literal with 2 squotes \"x''y\"" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_2_squotes.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x''y" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "x''y", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal" ;
	//   rdfs:comment "literal \"\"\"x\"\"\"" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "x", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_dquote> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_dquote" ;
	//   rdfs:comment 'literal with dquote "x\"y"' ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_dquote.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x\"y" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: `x"y`, DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_2_dquotes> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_2_dquotes" ;
	//   rdfs:comment "literal with 2 squotes \"\"\"a\"\"b\"\"\"" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_2_dquotes.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x\"\"y" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: `x""y`, DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_REVERSE_SOLIDUS2> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name    "literal_with_REVERSE_SOLIDUS2" ;
	//   rdfs:comment "REVERSE SOLIDUS at end of literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_REVERSE_SOLIDUS2.nq> ;
	//   .

	{`<http://example.org/ns#s> <http://example.org/ns#p1> "test-\\" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/ns#s"},
				Pred: IRI{str: "http://example.org/ns#p1"},
				Obj:  Literal{str: `test-\`, DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_CHARACTER_TABULATION> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_CHARACTER_TABULATION" ;
	//   rdfs:comment "literal with CHARACTER TABULATION" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_CHARACTER_TABULATION.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\t" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "\t", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_BACKSPACE> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_BACKSPACE" ;
	//   rdfs:comment "literal with BACKSPACE" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_BACKSPACE.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\b" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "\b", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_LINE_FEED> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_LINE_FEED" ;
	//   rdfs:comment "literal with LINE FEED" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_LINE_FEED.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\n" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "\n", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_CARRIAGE_RETURN> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_CARRIAGE_RETURN" ;
	//   rdfs:comment "literal with CARRIAGE RETURN" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_CARRIAGE_RETURN.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\r" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "\r", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_FORM_FEED> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_FORM_FEED" ;
	//   rdfs:comment "literal with FORM FEED" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_FORM_FEED.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\f" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "\f", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_REVERSE_SOLIDUS> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_REVERSE_SOLIDUS" ;
	//   rdfs:comment "literal with REVERSE SOLIDUS" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_REVERSE_SOLIDUS.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\\" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "\\", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_numeric_escape4> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_numeric_escape4" ;
	//   rdfs:comment "literal with numeric escape4 \\u" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_numeric_escape4.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\u006F" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "o", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#literal_with_numeric_escape8> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "literal_with_numeric_escape8" ;
	//   rdfs:comment "literal with numeric escape8 \\U" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_numeric_escape8.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\U0000006F" .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "o", DataType: xsdString},
			},
			defaultGraph,
		},
	}},

	//<#langtagged_string> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "langtagged_string" ;
	//   rdfs:comment "langtagged string \"x\"@en" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <langtagged_string.nq> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "chat"@en .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://a.example/s"},
				Pred: IRI{str: "http://a.example/p"},
				Obj:  Literal{str: "chat", lang: "en", DataType: rdfLangString},
			},
			defaultGraph,
		},
	}},

	//<#lantag_with_subtag> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "lantag_with_subtag" ;
	//   rdfs:comment "lantag with subtag \"x\"@en-us" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <lantag_with_subtag.nq> ;
	//   .

	{`<http://example.org/ex#a> <http://example.org/ex#b> "Cheers"@en-UK .`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example.org/ex#a"},
				Pred: IRI{str: "http://example.org/ex#b"},
				Obj:  Literal{str: "Cheers", lang: "en-UK", DataType: rdfLangString},
			},
			defaultGraph,
		},
	}},

	//<#minimal_whitespace> a rdft:TestNQuadsPositiveSyntax ;
	//   mf:name      "minimal_whitespace" ;
	//   rdfs:comment "tests absense of whitespace between subject, predicate, object and end-of-statement" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <minimal_whitespace.nq> ;
	//   .

	{`<http://example/s><http://example/p><http://example/o>.
<http://example/s><http://example/p>"Alice".
<http://example/s><http://example/p>_:o.
_:s<http://example/p><http://example/o>.
_:s<http://example/p>"Alice".
_:s<http://example/p>_:bnode1.`, "", []Quad{
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "Alice", DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: IRI{str: "http://example/s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Blank{id: "_:o"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: Blank{id: "_:s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  IRI{str: "http://example/o"},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: Blank{id: "_:s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Literal{str: "Alice", DataType: xsdString},
			},
			defaultGraph,
		},
		Quad{
			Triple{
				Subj: Blank{id: "_:s"},
				Pred: IRI{str: "http://example/p"},
				Obj:  Blank{id: "_:bnode1"},
			},
			defaultGraph,
		},
	}},
}

func TestEncodeNQuads(t *testing.T) {
	input := `<http://one.example/subject1> <http://one.example/predicate1> <http://one.example/object1> <http://example.org/graph3> . # comments here
# or on a line by themselves
_:subject1 <http://an.example/predicate1> "object1" <http://example.org/graph1> .
_:subject2 <http://an.example/predicate2> "object2" <http://example.org/graph5> .
<http://one.example/subject2> <http://one.example/predicate3> "xyz" . # use default graph
`
	dec := NewQuadDecoder(bytes.NewBufferString(input), NQuads)
	quads, err := dec.DecodeAll()
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	enc := NewQuadEncoder(&out, NQuads)
	if err := enc.EncodeAll(quads); err != nil {
		t.Fatal(err)
	}
	enc.Close()

	dec = NewQuadDecoder(&out, NQuads)
	quads2, err := dec.DecodeAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(quads) != len(quads2) {
		t.Fatal("N-Quads decode-encode-decode roundtrip failed")
	}
	for i, q := range quads {
		if !QuadsEqual(q, quads2[i]) {
			t.Fatal("N-Quads decode-encode-decode roundtrip failed")
		}
	}
}
