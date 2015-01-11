package parse

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/knakk/rdf"
)

func parseAllNT(s string) (r []rdf.Triple, err error) {
	dec := NewNTDecoder(bytes.NewBufferString(s))
	for tr, err := dec.DecodeTriple(); err != io.EOF; tr, err = dec.DecodeTriple() {
		if err != nil {
			return r, err
		}
		r = append(r, tr)
	}
	return r, err
}

func TestNT(t *testing.T) {
	for _, test := range ntTestSuite {
		triples, err := parseAllNT(test.input)
		if test.errWant != "" && err == nil {
			t.Errorf("parseNT(%s) => <no error>, want %q", test.input, test.errWant)
			continue
		}

		if test.errWant != "" && err != nil {
			if !strings.HasSuffix(err.Error(), test.errWant) {
				t.Errorf("parseNT(%s) => %v, want %q", test.input, err.Error(), test.errWant)
			}
			continue
		}

		if test.errWant == "" && err != nil {
			t.Errorf("parseNT(%s) => %v, want %q", test.input, err.Error(), test.want)
			continue
		}

		if !reflect.DeepEqual(triples, test.want) {
			t.Errorf("parseNT(%s) => %v, want %v", test.input, triples, test.want)
		}
	}
}

var empty = []rdf.Triple{rdf.Triple{}}

// ntTestSuite is a representation of the official W3C test suite for N-Triples
// which is found at: http://www.w3.org/2013/N-TriplesTests/
var ntTestSuite = []struct {
	input   string
	errWant string
	want    []rdf.Triple
}{
	//<#nt-syntax-file-01> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-file-01" ;
	//   rdfs:comment "Empty file" ;
	//   mf:action    <nt-syntax-file-01.nt> ;
	//   .

	{``, "", nil},

	//<#nt-syntax-file-02> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-file-02" ;
	//   rdfs:comment "Only comment" ;
	//   mf:action    <nt-syntax-file-02.nt> ;
	//   .

	{`#Empty file.`, "", nil},

	//<#nt-syntax-file-03> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-file-03" ;
	//   rdfs:comment "One comment, one empty line" ;
	//   mf:action    <nt-syntax-file-03.nt> ;
	//   .

	{`#One comment, one empty line.
	`, "", nil},

	//<#nt-syntax-uri-01> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-uri-01" ;
	//   rdfs:comment "Only IRIs" ;
	//   mf:action    <nt-syntax-uri-01.nt> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o> .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.URI{URI: "http://example/o"},
		},
	},
	},

	//<#nt-syntax-uri-02> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-uri-02" ;
	//   rdfs:comment "IRIs with Unicode escape" ;
	//   mf:action    <nt-syntax-uri-02.nt> ;
	//   .

	{`# x53 is capital S
	<http://example/\u0053> <http://example/p> <http://example/o> .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/S"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.URI{URI: "http://example/o"},
		},
	},
	},

	//<#nt-syntax-uri-03> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-uri-03" ;
	//   rdfs:comment "IRIs with long Unicode escape" ;
	//   mf:action    <nt-syntax-uri-03.nt> ;
	//   .

	{`# x53 is capital S
	<http://example/\U00000053> <http://example/p> <http://example/o> .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/S"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.URI{URI: "http://example/o"},
		},
	}},

	//<#nt-syntax-uri-04> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-uri-04" ;
	//   rdfs:comment "Legal IRIs" ;
	//   mf:action    <nt-syntax-uri-04.nt> ;
	//   .

	{`# IRI with all chars in it.
	<http://example/s> <http://example/p> <scheme:!$%25&'()*+,-./0123456789:/@ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~?#> .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.URI{URI: "scheme:!$%25&'()*+,-./0123456789:/@ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~?#"},
		},
	}},

	//<#nt-syntax-string-01> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-string-01" ;
	//   rdfs:comment "string literal" ;
	//   mf:action    <nt-syntax-string-01.nt> ;
	//   .

	{`<http://example/s> <http://example/p> "string" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "string", DataType: rdf.XSDString},
		},
	}},

	//<#nt-syntax-string-02> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-string-02" ;
	//   rdfs:comment "langString literal" ;
	//   mf:action    <nt-syntax-string-02.nt> ;
	//   .

	{`<http://example/s> <http://example/p> "string"@en .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "string", DataType: rdf.XSDString, Lang: "en"},
		},
	}},

	//<#nt-syntax-string-03> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-string-03" ;
	//   rdfs:comment "langString literal with region" ;
	//   mf:action    <nt-syntax-string-03.nt> ;
	//   .

	{`<http://example/s> <http://example/p> "string"@en-uk .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "string", DataType: rdf.XSDString, Lang: "en-uk"},
		},
	}},

	//<#nt-syntax-str-esc-01> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-str-esc-01" ;
	//   rdfs:comment "string literal with escaped newline" ;
	//   mf:action    <nt-syntax-str-esc-01.nt> ;
	//   .

	{`<http://example/s> <http://example/p> "a\n" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "a\n", DataType: rdf.XSDString},
		},
	}},

	//<#nt-syntax-str-esc-02> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-str-esc-02" ;
	//   rdfs:comment "string literal with Unicode escape" ;
	//   mf:action    <nt-syntax-str-esc-02.nt> ;
	//   .

	{`<http://example/s> <http://example/p> "a\u0020b" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "a b", DataType: rdf.XSDString},
		},
	}},

	//<#nt-syntax-str-esc-03> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-str-esc-03" ;
	//   rdfs:comment "string literal with long Unicode escape" ;
	//   mf:action    <nt-syntax-str-esc-03.nt> ;
	//   .

	{`<http://example/s> <http://example/p> "a\U00000020b" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "a b", DataType: rdf.XSDString},
		},
	}},

	//<#nt-syntax-bnode-01> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-bnode-01" ;
	//   rdfs:comment "bnode subject" ;
	//   mf:action    <nt-syntax-bnode-01.nt> ;
	//   .

	{`_:a  <http://example/p> <http://example/o> .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.Blank{ID: "a"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.URI{URI: "http://example/o"},
		},
	}},

	//<#nt-syntax-bnode-02> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-bnode-02" ;
	//   rdfs:comment "bnode object" ;
	//   mf:action    <nt-syntax-bnode-02.nt> ;
	//   .

	{`<http://example/s> <http://example/p> _:a .
	_:a  <http://example/p> <http://example/o> .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Blank{ID: "a"},
		},
		rdf.Triple{
			Subj: rdf.Blank{ID: "a"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.URI{URI: "http://example/o"},
		},
	}},

	//<#nt-syntax-bnode-03> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-bnode-03" ;
	//   rdfs:comment "Blank node labels may start with a digit" ;
	//   mf:action    <nt-syntax-bnode-03.nt> ;
	//   .

	{`<http://example/s> <http://example/p> _:1a .
	_:1a  <http://example/p> <http://example/o> .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Blank{ID: "1a"},
		},
		rdf.Triple{
			Subj: rdf.Blank{ID: "1a"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.URI{URI: "http://example/o"},
		},
	}},

	//<#nt-syntax-datatypes-01> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-datatypes-01" ;
	//   rdfs:comment "xsd:byte literal" ;
	//   mf:action    <nt-syntax-datatypes-01.nt> ;
	//   .

	{`<http://example/s> <http://example/p> "123"^^<http://www.w3.org/2001/XMLSchema#byte> .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "123", DataType: rdf.URI{URI: "http://www.w3.org/2001/XMLSchema#byte"}},
		},
	}},

	//<#nt-syntax-datatypes-02> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-datatypes-02" ;
	//   rdfs:comment "integer as xsd:string" ;
	//   mf:action    <nt-syntax-datatypes-02.nt> ;
	//   .

	{`<http://example/s> <http://example/p> "123"^^<http://www.w3.org/2001/XMLSchema#string> .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "123", DataType: rdf.XSDString},
		},
	}},

	//<#nt-syntax-bad-uri-01> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-01" ;
	//   rdfs:comment "Bad IRI : space (negative test) ;
	//   mf:action    <nt-syntax-bad-uri-01.nt> ;
	//   .

	{`# Bad IRI : space.
	<http://example/ space> <http://example/p> <http://example/o> .`, "bad IRI: disallowed character ' '", nil},

	//<#nt-syntax-bad-uri-02> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-02" ;
	//   rdfs:comment "Bad IRI : bad escape (negative test) ;
	//   mf:action    <nt-syntax-bad-uri-02.nt> ;
	//   .

	{`# Bad IRI : bad escape
	<http://example/\u00ZZ11> <http://example/p> <http://example/o> .`, "bad IRI: insufficent hex digits in unicode escape", nil},

	//<#nt-syntax-bad-uri-03> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-03" ;
	//   rdfs:comment "Bad IRI : bad long escape (negative test) ;
	//   mf:action    <nt-syntax-bad-uri-03.nt> ;
	//   .

	{`# Bad IRI : bad escape
	<http://example/\U00ZZ1111> <http://example/p> <http://example/o> .`, "bad IRI: insufficent hex digits in unicode escape", nil},

	//<#nt-syntax-bad-uri-04> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-04" ;
	//   rdfs:comment "Bad IRI : character escapes not allowed (negative test) ;
	//   mf:action    <nt-syntax-bad-uri-04.nt> ;
	//   .

	{`# Bad IRI : character escapes not allowed.
	<http://example/\n> <http://example/p> <http://example/o> .`, "bad IRI: disallowed escape character 'n'", nil},

	//<#nt-syntax-bad-uri-05> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-05" ;
	//   rdfs:comment "Bad IRI : character escapes not allowed (2) (negative test) ;
	//   mf:action    <nt-syntax-bad-uri-05.nt> ;
	//   .

	{`# Bad IRI : character escapes not allowed.
	<http://example/\/> <http://example/p> <http://example/o> .`, "bad IRI: disallowed escape character '/'", nil},

	//<#nt-syntax-bad-uri-06> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-06" ;
	//   rdfs:comment "Bad IRI : relative IRI not allowed in subject (negative test) ;
	//   mf:action    <nt-syntax-bad-uri-06.nt> ;
	//   .

	{`# No relative IRIs in N-Triples
	<s> <http://example/p> <http://example/o> .`, "unexpected IRI (relative) as subject", nil},

	//<#nt-syntax-bad-uri-07> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-07" ;
	//   rdfs:comment "Bad IRI : relative IRI not allowed in predicate (negative test) ;
	//   mf:action    <nt-syntax-bad-uri-07.nt> ;
	//   .

	{`# No relative IRIs in N-Triples
	<http://example/s> <p> <http://example/o> .`, "unexpected IRI (relative) as predicate", nil},

	//<#nt-syntax-bad-uri-08> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-08" ;
	//   rdfs:comment "Bad IRI : relative IRI not allowed in object (negative test) ;
	//   mf:action    <nt-syntax-bad-uri-08.nt> ;
	//   .

	{`# No relative IRIs in N-Triples
	<http://example/s> <http://example/p> <o> .`, "unexpected IRI (relative) as object", nil},

	//<#nt-syntax-bad-uri-09> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-uri-09" ;
	//   rdfs:comment "Bad IRI : relative IRI not allowed in datatype (negative test) ;
	//   mf:action    <nt-syntax-bad-uri-09.nt> ;
	//   .

	{`# No relative IRIs in N-Triples
	<http://example/s> <http://example/p> "foo"^^<dt> .`, "unexpected IRI (relative) as literal datatype", nil},

	//<#nt-syntax-bad-prefix-01> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-prefix-01" ;
	//   rdfs:comment "@prefix not allowed in n-triples (negative test) ;
	//   mf:action    <nt-syntax-bad-prefix-01.nt> ;
	//   .

	{`@prefix : <http://example/> .`, "unexpected @prefix as subject", nil},

	//<#nt-syntax-bad-base-01> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-base-01" ;
	//   rdfs:comment "@base not allowed in N-Triples (negative test) ;
	//   mf:action    <nt-syntax-bad-base-01.nt> ;
	//   .

	{`@base <http://example/> .`, "unexpected @base as subject", nil},

	//<#nt-syntax-bad-struct-01> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-struct-01" ;
	//   rdfs:comment "N-Triples does not have objectList (negative test) ;
	//   mf:action    <nt-syntax-bad-struct-01.nt> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o>, <http://example/o2> .`, "unexpected Comma as dot (.)", nil},

	//<#nt-syntax-bad-struct-02> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-struct-02" ;
	//   rdfs:comment "N-Triples does not have predicateObjectList (negative test) ;
	//   mf:action    <nt-syntax-bad-struct-02.nt> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o>; <http://example/p2>, <http://example/o2> .`, "unexpected Semicolon as dot (.)", nil},

	//<#nt-syntax-bad-lang-01> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-lang-01" ;
	//   rdfs:comment "langString with bad lang (negative test) ;
	//   mf:action    <nt-syntax-bad-lang-01.nt> ;
	//   .

	{`# Bad lang tag
	<http://example/s> <http://example/p> "string"@1 .`, "syntax error: bad literal: invalid language tag", nil},

	//<#nt-syntax-bad-esc-01> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-esc-01" ;
	//   rdfs:comment "Bad string escape (negative test) ;
	//   mf:action    <nt-syntax-bad-esc-01.nt> ;
	//   .

	{`# Bad string escape
	<http://example/s> <http://example/p> "a\zb" .`, "syntax error: bad literal: disallowed escape character 'z'", nil},

	//<#nt-syntax-bad-esc-02> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-esc-02" ;
	//   rdfs:comment "Bad string escape (negative test) ;
	//   mf:action    <nt-syntax-bad-esc-02.nt> ;
	//   .

	{`# Bad string escape
	<http://example/s> <http://example/p> "\uWXYZ" .`, "syntax error: bad literal: insufficent hex digits in unicode escape", nil},

	//<#nt-syntax-bad-esc-03> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-esc-03" ;
	//   rdfs:comment "Bad string escape (negative test) ;
	//   mf:action    <nt-syntax-bad-esc-03.nt> ;
	//   .

	{`# Bad string escape
	<http://example/s> <http://example/p> "\U0000WXYZ" .`, "syntax error: bad literal: insufficent hex digits in unicode escape", nil},

	//<#nt-syntax-bad-string-01> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-01" ;
	//   rdfs:comment "mismatching string literal open/close (negative test) ;
	//   mf:action    <nt-syntax-bad-string-01.nt> ;
	//   .

	{`<http://example/s> <http://example/p> "abc' .`, "syntax error: bad literal: no closing quote: '\"'", nil},

	//<#nt-syntax-bad-string-02> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-02" ;
	//   rdfs:comment "mismatching string literal open/close (negative test) ;
	//   mf:action    <nt-syntax-bad-string-02.nt> ;
	//   .

	{`<http://example/s> <http://example/p> 1.0 .`, "unexpected Literal (decimal shorthand syntax) as object", nil},

	//<#nt-syntax-bad-string-03> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-03" ;
	//   rdfs:comment "single quotes (negative test) ;
	//   mf:action    <nt-syntax-bad-string-03.nt> ;
	//   .

	{`<http://example/s> <http://example/p> 1.0e1 .`, "unexpected Literal (double shorthand syntax) as object", nil},

	//<#nt-syntax-bad-string-04> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-04" ;
	//   rdfs:comment "long single string literal (negative test) ;
	//   mf:action    <nt-syntax-bad-string-04.nt> ;
	//   .

	{`<http://example/s> <http://example/p> '''abc''' .`, "unexpected Literal (triple-quoted string) as object", nil},

	//<#nt-syntax-bad-string-05> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-05" ;
	//   rdfs:comment "long double string literal (negative test) ;
	//   mf:action    <nt-syntax-bad-string-05.nt> ;
	//   .

	{`<http://example/s> <http://example/p> """abc""" .`, "unexpected Literal (triple-quoted string) as object", nil},

	//<#nt-syntax-bad-string-06> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-06" ;
	//   rdfs:comment "string literal with no end (negative test) ;
	//   mf:action    <nt-syntax-bad-string-06.nt> ;
	//   .

	{`<http://example/s> <http://example/p> "abc .`, "syntax error: bad literal: no closing quote: '\"'", nil},

	//<#nt-syntax-bad-string-07> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-string-07" ;
	//   rdfs:comment "string literal with no start (negative test) ;
	//   mf:action    <nt-syntax-bad-string-07.nt> ;
	//   .

	{`<http://example/s> <http://example/p> abc" .`, "syntax error: illegal token: abc\"", nil},

	//<#nt-syntax-bad-num-01> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-num-01" ;
	//   rdfs:comment "no numbers in N-Triples (integer) (negative test) ;
	//   mf:action    <nt-syntax-bad-num-01.nt> ;
	//   .

	{`<http://example/s> <http://example/p> 1 .`, "unexpected Literal (integer shorthand syntax) as object", nil},

	//<#nt-syntax-bad-num-02> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-num-02" ;
	//   rdfs:comment "no numbers in N-Triples (decimal) (negative test) ;
	//   mf:action    <nt-syntax-bad-num-02.nt> ;
	//   .

	{`<http://example/s> <http://example/p> 1.0 .`, "unexpected Literal (decimal shorthand syntax) as object", nil},

	//<#nt-syntax-bad-num-03> rdf:type rdft:TestNTriplesNegativeSyntax ;
	//   mf:name    "nt-syntax-bad-num-03" ;
	//   rdfs:comment "no numbers in N-Triples (float) (negative test) ;
	//   mf:action    <nt-syntax-bad-num-03.nt> ;
	//   .

	{`<http://example/s> <http://example/p> 1.0e0 .`, "unexpected Literal (double shorthand syntax) as object", nil},

	//<#nt-syntax-subm-01> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "nt-syntax-subm-01" ;
	//   rdfs:comment "Submission test from Original RDF Test Cases" ;
	//   mf:action    <nt-syntax-subm-01.nt> ;
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
	# resource33 test removed 2003-08-03`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource1"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.URI{URI: "http://example.org/resource2"},
		},
		rdf.Triple{
			Subj: rdf.Blank{ID: "anon"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.URI{URI: "http://example.org/resource2"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource2"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Blank{ID: "anon"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource3"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.URI{URI: "http://example.org/resource2"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource4"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.URI{URI: "http://example.org/resource2"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource5"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.URI{URI: "http://example.org/resource2"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource6"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.URI{URI: "http://example.org/resource2"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource7"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "simple literal", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource8"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: `backslash:\`, DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource9"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: `dquote:"`, DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource10"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "newline:\n", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource11"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "return\r", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource12"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "tab:\t", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource13"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.URI{URI: "http://example.org/resource2"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource14"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "x", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource15"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Blank{ID: "anon"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource16"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "é", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource17"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "€", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource21"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "", DataType: rdf.URI{URI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource22"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: " ", DataType: rdf.URI{URI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource23"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "x", DataType: rdf.URI{URI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource23"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: `"`, DataType: rdf.URI{URI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource24"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "<a></a>", DataType: rdf.URI{URI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource25"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "a <b></b>", DataType: rdf.URI{URI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource26"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "a <b></b> c", DataType: rdf.URI{URI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource26"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "a\n<b></b>\nc", DataType: rdf.URI{URI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource27"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "chat", DataType: rdf.URI{URI: "http://www.w3.org/2000/01/rdf-schema#XMLLiteral"}},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource30"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "chat", Lang: "fr", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource31"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "chat", Lang: "en", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/resource32"},
			Pred: rdf.URI{URI: "http://example.org/property"},
			Obj:  rdf.Literal{Val: "abc", DataType: rdf.URI{URI: "http://example.org/datatype1"}},
		},
	}},

	//<#comment_following_triple> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "comment_following_triple" ;
	//   rdfs:comment "Tests comments after a triple" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <comment_following_triple.nt> ;
	//   .

	{`<http://example/s> <http://example/p> <http://example/o> . # comment
	<http://example/s> <http://example/p> _:o . # comment
	<http://example/s> <http://example/p> "o" . # comment
	<http://example/s> <http://example/p> "o"^^<http://example/dt> . # comment
	<http://example/s> <http://example/p> "o"@en . # comment`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.URI{URI: "http://example/o"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Blank{ID: "o"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "o", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "o", DataType: rdf.URI{URI: "http://example/dt"}},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "o", Lang: "en", DataType: rdf.XSDString},
		},
	}},

	//<#literal_ascii_boundaries> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_ascii_boundaries" ;
	//   rdfs:comment "literal_ascii_boundaries '\\x00\\x26\\x28...'" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_ascii_boundaries.nt> ;
	//   .

	{"<http://a.example/s> <http://a.example/p> \"\x00	&([]\" .", "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj: rdf.Literal{Val: "\x00	&([]", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_UTF8_boundaries> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_UTF8_boundaries" ;
	//   rdfs:comment "literal_with_UTF8_boundaries '\\x80\\x7ff\\x800\\xfff...'" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_UTF8_boundaries.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽", DataType: rdf.XSDString},
		},
	}},

	//<#literal_all_controls> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_all_controls" ;
	//   rdfs:comment "literal_all_controls '\\x00\\x01\\x02\\x03\\x04...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action   <literal_all_controls.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\u0000\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u0008\t\u000B\u000C\u000E\u000F\u0010\u0011\u0012\u0013\u0014\u0015\u0016\u0017\u0018\u0019\u001A\u001B\u001C\u001D\u001E\u001F" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "\x00\x01\x02\x03\x04\x05\x06\a\b\t\v\f\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f", DataType: rdf.XSDString},
		},
	}},

	//<#literal_all_punctuation> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_all_punctuation" ;
	//   rdfs:comment "literal_all_punctuation '!\"#$%&()...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_all_punctuation.nt> ;
	//   .

	{"<http://a.example/s> <http://a.example/p> \" !\\\"#$%&():;<=>?@[]^_`{|}~\".", "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: " !\"#$%&():;<=>?@[]^_`{|}~", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_squote> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_squote" ;
	//   rdfs:comment "literal with squote \"x'y\"" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_squote.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x'y" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "x'y", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_2_squotes> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_2_squotes" ;
	//   rdfs:comment "literal with 2 squotes \"x''y\"" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_2_squotes.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x''y" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "x''y", DataType: rdf.XSDString},
		},
	}},

	//<#literal> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal" ;
	//   rdfs:comment "literal \"\"\"x\"\"\"" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "x", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_dquote> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_dquote" ;
	//   rdfs:comment 'literal with dquote "x\"y"' ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_dquote.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x\"y" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: `x"y`, DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_2_dquotes> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_2_dquotes" ;
	//   rdfs:comment "literal with 2 squotes \"\"\"a\"\"b\"\"\"" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_2_dquotes.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x\"\"y" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: `x""y`, DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_REVERSE_SOLIDUS2> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name    "literal_with_REVERSE_SOLIDUS2" ;
	//   rdfs:comment "REVERSE SOLIDUS at end of literal" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_REVERSE_SOLIDUS2.nt> ;
	//   .

	{`<http://example.org/ns#s> <http://example.org/ns#p1> "test-\\" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/ns#s"},
			Pred: rdf.URI{URI: "http://example.org/ns#p1"},
			Obj:  rdf.Literal{Val: `test-\`, DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_CHARACTER_TABULATION> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_CHARACTER_TABULATION" ;
	//   rdfs:comment "literal with CHARACTER TABULATION" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_CHARACTER_TABULATION.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\t" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "\t", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_BACKSPACE> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_BACKSPACE" ;
	//   rdfs:comment "literal with BACKSPACE" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_BACKSPACE.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\b" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "\b", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_LINE_FEED> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_LINE_FEED" ;
	//   rdfs:comment "literal with LINE FEED" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_LINE_FEED.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\n" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "\n", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_CARRIAGE_RETURN> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_CARRIAGE_RETURN" ;
	//   rdfs:comment "literal with CARRIAGE RETURN" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_CARRIAGE_RETURN.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\r" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "\r", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_FORM_FEED> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_FORM_FEED" ;
	//   rdfs:comment "literal with FORM FEED" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_FORM_FEED.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\f" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "\f", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_REVERSE_SOLIDUS> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_REVERSE_SOLIDUS" ;
	//   rdfs:comment "literal with REVERSE SOLIDUS" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_REVERSE_SOLIDUS.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\\" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "\\", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_numeric_escape4> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_numeric_escape4" ;
	//   rdfs:comment "literal with numeric escape4 \\u" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_numeric_escape4.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\u006F" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "o", DataType: rdf.XSDString},
		},
	}},

	//<#literal_with_numeric_escape8> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "literal_with_numeric_escape8" ;
	//   rdfs:comment "literal with numeric escape8 \\U" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <literal_with_numeric_escape8.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\U0000006F" .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "o", DataType: rdf.XSDString},
		},
	}},

	//<#langtagged_string> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "langtagged_string" ;
	//   rdfs:comment "langtagged string \"x\"@en" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <langtagged_string.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "chat"@en .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://a.example/s"},
			Pred: rdf.URI{URI: "http://a.example/p"},
			Obj:  rdf.Literal{Val: "chat", Lang: "en", DataType: rdf.XSDString},
		},
	}},

	//<#lantag_with_subtag> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "lantag_with_subtag" ;
	//   rdfs:comment "lantag with subtag \"x\"@en-us" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <lantag_with_subtag.nt> ;
	//   .

	{`<http://example.org/ex#a> <http://example.org/ex#b> "Cheers"@en-UK .`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example.org/ex#a"},
			Pred: rdf.URI{URI: "http://example.org/ex#b"},
			Obj:  rdf.Literal{Val: "Cheers", Lang: "en-UK", DataType: rdf.XSDString},
		},
	}},

	//<#minimal_whitespace> rdf:type rdft:TestNTriplesPositiveSyntax ;
	//   mf:name      "minimal_whitespace" ;
	//   rdfs:comment "tests absense of whitespace between subject, predicate, object and end-of-statement" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <minimal_whitespace.nt> ;
	//   .

	{`<http://example/s><http://example/p><http://example/o>.
	<http://example/s><http://example/p>"Alice".
	<http://example/s><http://example/p>_:o.
	_:s<http://example/p><http://example/o>.
	_:s<http://example/p>"Alice".
	_:s<http://example/p>_:bnode1.`, "", []rdf.Triple{
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.URI{URI: "http://example/o"},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "Alice", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.URI{URI: "http://example/s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Blank{ID: "o"},
		},
		rdf.Triple{
			Subj: rdf.Blank{ID: "s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.URI{URI: "http://example/o"},
		},
		rdf.Triple{
			Subj: rdf.Blank{ID: "s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Literal{Val: "Alice", DataType: rdf.XSDString},
		},
		rdf.Triple{
			Subj: rdf.Blank{ID: "s"},
			Pred: rdf.URI{URI: "http://example/p"},
			Obj:  rdf.Blank{ID: "bnode1"},
		},
	}},
}
