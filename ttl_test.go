package rdf

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

var ttlBenchInputs = []string{
	`# example 1
@base <http://example.org/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix rel: <http://www.perceive.net/schemas/relationship/> .

<#green-goblin>
    rel:enemyOf <#spiderman> ;
    a foaf:Person ;    # in the context of the Marvel universe
    foaf:name "Green Goblin" .

<#spiderman>
    rel:enemyOf <#green-goblin> ;
    a foaf:Person ;
    foaf:name "Spiderman", "Человек-паук"@ru .`,

	`# example 2
<http://example.org/#spiderman> <http://www.perceive.net/schemas/relationship/enemyOf> <http://example.org/#green-goblin> .`,

	`# example 3
<http://example.org/#spiderman> <http://www.perceive.net/schemas/relationship/enemyOf> <http://example.org/#green-goblin> ;
				<http://xmlns.com/foaf/0.1/name> "Spiderman" .`,

	`# example 4
<http://example.org/#spiderman> <http://www.perceive.net/schemas/relationship/enemyOf> <http://example.org/#green-goblin> .
<http://example.org/#spiderman> <http://xmlns.com/foaf/0.1/name> "Spiderman" .`,

	`# example 5
<http://example.org/#spiderman> <http://xmlns.com/foaf/0.1/name> "Spiderman", "Человек-паук"@ru .`,

	`# example 6
<http://example.org/#spiderman> <http://xmlns.com/foaf/0.1/name> "Spiderman" .
<http://example.org/#spiderman> <http://xmlns.com/foaf/0.1/name> "Человек-паук"@ru .`,

	`# example 7
@prefix somePrefix: <http://www.perceive.net/schemas/relationship/> .

<http://example.org/#green-goblin> somePrefix:enemyOf <http://example.org/#spiderman> .`,

	`# example 8
PREFIX somePrefix: <http://www.perceive.net/schemas/relationship/>

<http://example.org/#green-goblin> somePrefix:enemyOf <http://example.org/#spiderman> .`,

	`# example 9
# A triple with all absolute IRIs
<http://one.example/subject1> <http://one.example/predicate1> <http://one.example/object1> .

@base <http://one.example/> .
<subject2> <predicate2> <object2> .     # relative IRIs, e.g. http://one.example/subject2

BASE <http://one.example/>
<subject2> <predicate2> <object2> .     # relative IRIs, e.g. http://one.example/subject2

@prefix p: <http://two.example/> .
p:subject3 p:predicate3 p:object3 .     # prefixed name, e.g. http://two.example/subject3

PREFIX p: <http://two.example/>
p:subject3 p:predicate3 p:object3 .     # prefixed name, e.g. http://two.example/subject3

@prefix p: <path/> .                    # prefix p: now stands for http://one.example/path/
p:subject4 p:predicate4 p:object4 .     # prefixed name, e.g. http://one.example/path/subject4

@prefix : <http://another.example/> .    # empty prefix
:subject5 :predicate5 :object5 .        # prefixed name, e.g. http://another.example/subject5

:subject6 a :subject7 .                 # same as :subject6 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> :subject7 .

<http://伝言.example/?user=أكرم&amp;channel=R%26D> a :subject8 . # a multi-script subject IRI .`,

	`# example 10
@prefix foaf: <http://xmlns.com/foaf/0.1/> .

<http://example.org/#green-goblin> foaf:name "Green Goblin" .

<http://example.org/#spiderman> foaf:name "Spiderman" .`,

	`# example 11
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix show: <http://example.org/vocab/show/> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .

show:218 rdfs:label "That Seventies Show"^^xsd:string .            # literal with XML Schema string datatype
show:218 rdfs:label "That Seventies Show"^^<http://www.w3.org/2001/XMLSchema#string> . # same as above
show:218 rdfs:label "That Seventies Show" .                                            # same again
show:218 show:localName "That Seventies Show"@en .                 # literal with a language tag
show:218 show:localName 'Cette Série des Années Soixante-dix'@fr . # literal delimited by single quote
show:218 show:localName "Cette Série des Années Septante"@fr-be .  # literal with a region subtag
																   # literal with embedded new lines and quotes
show:218 show:blurb '''This is a multi-line
literal with many quotes (""""")
and up to two sequential apostrophes ('').''' .`,

	`# example 12
@prefix : <http://example.org/elements> .
<http://en.wikipedia.org/wiki/Helium>
    :atomicNumber 2 ;               # xsd:integer
    :atomicMass 4.002602 ;          # xsd:decimal
    :specificGravity 1.663E-4 .     # xsd:double`,

	`# example 13
@prefix : <http://example.org/stats> .
<http://somecountry.example/census2007>
    :isLandlocked false .           # xsd:boolean`,

	`# example 14
@prefix foaf: <http://xmlns.com/foaf/0.1/> .

_:alice foaf:knows _:bob .
_:bob foaf:knows _:alice .`,

	`# example 15
@prefix foaf: <http://xmlns.com/foaf/0.1/> .

# Someone knows someone else, who has the name "Bob".
[] foaf:knows [ foaf:name "Bob" ] .`,

	`# example 16
@prefix foaf: <http://xmlns.com/foaf/0.1/> .

[ foaf:name "Alice" ] foaf:knows [
    foaf:name "Bob" ;
    foaf:knows [
        foaf:name "Eve" ] ;
    foaf:mbox <bob@example.com> ] .`,

	`# example 17
_:a <http://xmlns.com/foaf/0.1/name> "Alice" .
_:a <http://xmlns.com/foaf/0.1/knows> _:b .
_:b <http://xmlns.com/foaf/0.1/name> "Bob" .
_:b <http://xmlns.com/foaf/0.1/knows> _:c .
_:c <http://xmlns.com/foaf/0.1/name> "Eve" .
_:b <http://xmlns.com/foaf/0.1/mbox> <bob@example.com> .`,

	`# example 18
@prefix : <http://example.org/foo> .
# the object of this triple is the RDF collection blank node
:subject :predicate ( :a :b :c ) .

# an empty collection value - rdf:nil
:subject :predicate2 () .`,

	`# example 19
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix dc: <http://purl.org/dc/elements/1.1/> .
@prefix ex: <http://example.org/stuff/1.0/> .

<http://www.w3.org/TR/rdf-syntax-grammar>
  dc:title "RDF/XML Syntax Specification (Revised)" ;
  ex:editor [
    ex:fullname "Dave Beckett";
    ex:homePage <http://purl.org/net/dajobe/>
  ] .`,

	`# example 20
PREFIX : <http://example.org/stuff/1.0/>
:a :b ( "apple" "banana" ) .
         `,
	`# example 21
@prefix : <http://example.org/stuff/1.0/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
:a :b
  [ rdf:first "apple";
    rdf:rest [ rdf:first "banana";
               rdf:rest rdf:nil ]
  ] .`,

	`# example 22
@prefix : <http://example.org/stuff/1.0/> .

:a :b "The first line\nThe second line\n  more" .

:a :b """The first line
The second line
  more""" .`,

	`# example 23
@prefix : <http://example.org/stuff/1.0/> .
(1 2.0 3E1) :p "w" .`,

	`# example 24
@prefix : <http://example.org/stuff/1.0/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
    _:b0  rdf:first  1 ;
          rdf:rest   _:b1 .
    _:b1  rdf:first  2.0 ;
          rdf:rest   _:b2 .
    _:b2  rdf:first  3E1 ;
          rdf:rest   rdf:nil .
    _:b0  :p         "w" . `,

	`# example 25
PREFIX : <http://example.org/stuff/1.0/>
(1 [:p :q] ( 2 ) ) :p2 :q2 .`,

	`# example 26
@prefix : <http://example.org/stuff/1.0/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
    _:b0  rdf:first  1 ;
          rdf:rest   _:b1 .
    _:b1  rdf:first  _:b2 .
    _:b2  :p         :q .
    _:b1  rdf:rest   _:b3 .
    _:b3  rdf:first  _:b4 .
    _:b4  rdf:first  2 ;
          rdf:rest   rdf:nil .
    _:b3  rdf:rest   rdf:nil .`,

	`# example 27
@prefix ericFoaf: <http://www.w3.org/People/Eric/ericP-foaf.rdf#> .
@prefix : <http://xmlns.com/foaf/0.1/> .
ericFoaf:ericP :givenName "Eric" ;
              :knows <http://norman.walsh.name/knows/who/dan-brickley> ,
                      [ :mbox <mailto:timbl@w3.org> ] ,
                      <http://getopenid.com/amyvdh> .
         `,
	`# example 28
@prefix dc: <http://purl.org/dc/terms/> .
@prefix frbr: <http://purl.org/vocab/frbr/core#> .

<http://books.example.com/works/45U8QJGZSQKDH8N> a frbr:Work ;
     dc:creator "Wil Wheaton"@en ;
     dc:title "Just a Geek"@en ;
     frbr:realization <http://books.example.com/products/9780596007683.BOOK>,
         <http://books.example.com/products/9780596802189.EBOOK> .

<http://books.example.com/products/9780596007683.BOOK> a frbr:Expression ;
     dc:type <http://books.example.com/product-types/BOOK> .

<http://books.example.com/products/9780596802189.EBOOK> a frbr:Expression ;
     dc:type <http://books.example.com/product-types/EBOOK> .`,

	`# example 29
@prefix frbr: <http://purl.org/vocab/frbr/core#> .

<http://books.example.com/works/45U8QJGZSQKDH8N> a frbr:Work .`,
}

var ttlBenchOutputs = []string{
	`@prefix ns0:	<http://example.org/#> .
@prefix ns1:	<http://www.perceive.net/schemas/relationship/> .
ns0:green-goblin	ns1:enemyOf	ns0:spiderman .
@prefix ns2:	<http://xmlns.com/foaf/0.1/> .
ns0:green-goblin	a	ns2:Person ;
	ns2:name	"Green Goblin" .
ns0:spiderman	ns1:enemyOf	ns0:green-goblin ;
	a	ns2:Person ;
	ns2:name	"Spiderman" ,
			"Человек-паук"@ru .`,

	`@prefix ns0:	<http://example.org/#> .
@prefix ns1:	<http://www.perceive.net/schemas/relationship/> .
ns0:spiderman	ns1:enemyOf	ns0:green-goblin .`,

	`@prefix ns0:	<http://example.org/#> .
@prefix ns1:	<http://www.perceive.net/schemas/relationship/> .
ns0:spiderman	ns1:enemyOf	ns0:green-goblin .
@prefix ns2:	<http://xmlns.com/foaf/0.1/> .
ns0:spiderman	ns2:name	"Spiderman" .`,

	`@prefix ns0:	<http://example.org/#> .
@prefix ns1:	<http://www.perceive.net/schemas/relationship/> .
ns0:spiderman	ns1:enemyOf	ns0:green-goblin .
@prefix ns2:	<http://xmlns.com/foaf/0.1/> .
ns0:spiderman	ns2:name	"Spiderman" .`,

	`@prefix ns0:	<http://xmlns.com/foaf/0.1/> .
@prefix ns1:	<http://example.org/#> .
ns1:spiderman	ns0:name	"Spiderman" ,
			"Человек-паук"@ru .`,

	`@prefix ns0:	<http://xmlns.com/foaf/0.1/> .
@prefix ns1:	<http://example.org/#> .
ns1:spiderman	ns0:name	"Spiderman" ,
			"Человек-паук"@ru .`,

	`@prefix ns0:	<http://example.org/#> .
@prefix ns1:	<http://www.perceive.net/schemas/relationship/> .
ns0:green-goblin	ns1:enemyOf	ns0:spiderman .`,

	`@prefix ns0:	<http://example.org/#> .
@prefix ns1:	<http://www.perceive.net/schemas/relationship/> .
ns0:green-goblin	ns1:enemyOf	ns0:spiderman .`,

	`@prefix ns0:	<http://another.example/> .
ns0:subject5	ns0:predicate5	ns0:object5 .
ns0:subject6	a	ns0:subject7 .
@prefix ns1:	<http://one.example/path/> .
ns1:subject4	ns1:predicate4	ns1:object4 .
@prefix ns2:	<http://one.example/> .
ns2:subject1	ns2:predicate1	ns2:object1 .
ns2:subject2	ns2:predicate2	ns2:object2 .
@prefix ns3:	<http://two.example/> .
ns3:subject3	ns3:predicate3	ns3:object3 .
@prefix ns4:	<http://伝言.example/> .
ns4:?user=أكرم&amp;channel=R%26D	a	ns0:subject8 .`,

	`@prefix ns0:	<http://xmlns.com/foaf/0.1/> .
@prefix ns1:	<http://example.org/#> .
ns1:green-goblin	ns0:name	"Green Goblin" .
ns1:spiderman	ns0:name	"Spiderman" .`,

	`@prefix ns0:	<http://example.org/vocab/show/> .
ns0:218	ns0:blurb	"This is a multi-line\nliteral with many quotes (\"\"\"\"\")\nand up to two sequential apostrophes ('')." ;
	ns0:localName	"That Seventies Show"@en ,
			"Cette Série des Années Soixante-dix"@fr ,
			"Cette Série des Années Septante"@fr-be .
@prefix ns1:	<http://www.w3.org/2000/01/rdf-schema#> .
ns0:218	ns1:label	"That Seventies Show" .`,

	`@prefix ns0:	<http://example.org/> .
@prefix ns1:	<http://en.wikipedia.org/wiki/> .
ns1:Helium	ns0:elementsatomicMass	4.002602 ;
	ns0:elementsatomicNumber	2 ;
	ns0:elementsspecificGravity	1.663E-4 .`,

	`@prefix ns0:	<http://example.org/> .
@prefix ns1:	<http://somecountry.example/> .
ns1:census2007	ns0:statsisLandlocked	false .`,

	`@prefix ns0:	<http://xmlns.com/foaf/0.1/> .
_:alice	ns0:knows	_:bob .
_:bob	ns0:knows	_:alice .`,

	`@prefix ns0:	<http://xmlns.com/foaf/0.1/> .
_:b1	ns0:knows	_:b2 .
_:b2	ns0:name	"Bob" .`,

	`@prefix ns0:	<http://xmlns.com/foaf/0.1/> .
_:b1	ns0:knows	_:b2 ;
	ns0:name	"Alice" .
_:b2	ns0:knows	_:b3 ;
	ns0:mbox	<bob@example.com> ;
	ns0:name	"Bob" .
_:b3	ns0:name	"Eve" .`,

	`@prefix ns0:	<http://xmlns.com/foaf/0.1/> .
_:a	ns0:knows	_:b ;
	ns0:name	"Alice" .
_:b	ns0:knows	_:c ;
	ns0:mbox	<bob@example.com> ;
	ns0:name	"Bob" .
_:c	ns0:name	"Eve" .`,

	`@prefix rdf:	<http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix ns0:	<http://example.org/> .
ns0:foosubject	ns0:foopredicate2	rdf:nil ;
	ns0:foopredicate	_:b1 .
_:b1	rdf:first	ns0:fooa ;
	rdf:rest	_:b2 .
_:b2	rdf:first	ns0:foob ;
	rdf:rest	_:b3 .
_:b3	rdf:first	ns0:fooc ;
	rdf:rest	rdf:nil .`,

	`@prefix ns0:	<http://example.org/stuff/1.0/> .
@prefix ns1:	<http://www.w3.org/TR/> .
ns1:rdf-syntax-grammar	ns0:editor	_:b1 .
@prefix ns2:	<http://purl.org/dc/elements/1.1/> .
ns1:rdf-syntax-grammar	ns2:title	"RDF/XML Syntax Specification (Revised)" .
_:b1	ns0:fullname	"Dave Beckett" .
@prefix ns3:	<http://purl.org/net/dajobe/> .
_:b1	ns0:homePage	ns3: .`,

	`@prefix ns0:	<http://example.org/stuff/1.0/> .
ns0:a	ns0:b	_:b1 .
@prefix rdf:	<http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
_:b1	rdf:first	"apple" ;
	rdf:rest	_:b2 .
_:b2	rdf:first	"banana" ;
	rdf:rest	rdf:nil .`,

	`@prefix ns0:	<http://example.org/stuff/1.0/> .
ns0:a	ns0:b	_:b1 .
@prefix rdf:	<http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
_:b1	rdf:first	"apple" ;
	rdf:rest	_:b2 .
_:b2	rdf:first	"banana" ;
	rdf:rest	rdf:nil .`,

	`@prefix ns0:	<http://example.org/stuff/1.0/> .
ns0:a	ns0:b	"The first line\nThe second line\n  more" .`,

	`@prefix ns0:	<http://example.org/stuff/1.0/> .
_:b1	ns0:p	"w" .
@prefix rdf:	<http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
_:b1	rdf:first	1 ;
	rdf:rest	_:b2 .
_:b2	rdf:first	2.0 ;
	rdf:rest	_:b3 .
_:b3	rdf:first	3E1 ;
	rdf:rest	rdf:nil .`,

	`@prefix ns0:	<http://example.org/stuff/1.0/> .
_:b0	ns0:p	"w" .
@prefix rdf:	<http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
_:b0	rdf:first	1 ;
	rdf:rest	_:b1 .
_:b1	rdf:first	2.0 ;
	rdf:rest	_:b2 .
_:b2	rdf:first	3E1 ;
	rdf:rest	rdf:nil .`,

	`@prefix ns0:	<http://example.org/stuff/1.0/> .
_:b1	ns0:p2	ns0:q2 .
@prefix rdf:	<http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
_:b1	rdf:first	1 ;
	rdf:rest	_:b2 .
_:b2	rdf:first	_:b3 ;
	rdf:rest	_:b4 .
_:b3	ns0:p	ns0:q .
_:b4	rdf:first	_:b5 ;
	rdf:rest	rdf:nil .
_:b5	rdf:first	2 ;
	rdf:rest	rdf:nil .`,

	`@prefix rdf:	<http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
_:b0	rdf:first	1 ;
	rdf:rest	_:b1 .
_:b1	rdf:first	_:b2 ;
	rdf:rest	_:b3 .
@prefix ns0:	<http://example.org/stuff/1.0/> .
_:b2	ns0:p	ns0:q .
_:b3	rdf:first	_:b4 ;
	rdf:rest	rdf:nil .
_:b4	rdf:first	2 ;
	rdf:rest	rdf:nil .`,

	`@prefix ns0:	<http://xmlns.com/foaf/0.1/> .
@prefix ns1:	<http://www.w3.org/People/Eric/ericP-foaf.rdf#> .
ns1:ericP	ns0:givenName	"Eric" .
@prefix ns2:	<http://norman.walsh.name/knows/who/> .
ns1:ericP	ns0:knows	ns2:dan-brickley ,
			_:b1 .
@prefix ns3:	<http://getopenid.com/> .
ns1:ericP	ns0:knows	ns3:amyvdh .
_:b1	ns0:mbox	<mailto:timbl@w3.org> .`,

	`@prefix ns0:	<http://books.example.com/product-types/> .
@prefix ns1:	<http://purl.org/dc/terms/> .
@prefix ns2:	<http://books.example.com/products/> .
ns2:9780596007683.BOOK	ns1:type	ns0:BOOK .
@prefix ns3:	<http://purl.org/vocab/frbr/core#> .
ns2:9780596007683.BOOK	a	ns3:Expression .
ns2:9780596802189.EBOOK	ns1:type	ns0:EBOOK ;
	a	ns3:Expression .
@prefix ns4:	<http://books.example.com/works/> .
ns4:45U8QJGZSQKDH8N	ns1:creator	"Wil Wheaton"@en ;
	ns1:title	"Just a Geek"@en ;
	ns3:realization	ns2:9780596007683.BOOK ,
			ns2:9780596802189.EBOOK ;
	a	ns3:Work .`,

	`@prefix ns0:	<http://purl.org/vocab/frbr/core#> .
@prefix ns1:	<http://books.example.com/works/> .
ns1:45U8QJGZSQKDH8N	a	ns0:Work .`,
}

func BenchmarkDecodeTTL(b *testing.B) {
	var bf bytes.Buffer
	for _, i := range ttlBenchInputs {
		bf.WriteString(i)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dec := NewTripleDecoder(bytes.NewReader(bf.Bytes()), Turtle)
		for _, err := dec.Decode(); err != io.EOF; _, err = dec.Decode() {
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	b.SetBytes(int64(len(bf.Bytes())))
}

func TestEncodingTTL(t *testing.T) {
	for i, ex := range ttlBenchInputs {
		dec := NewTripleDecoder(bytes.NewBufferString(ex), Turtle)
		triples, err := dec.DecodeAll()
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		enc := NewTripleEncoder(&buf, Turtle)

		//test custom namespace as well
		enc.Namespaces["http://www.w3.org/1999/02/22-rdf-syntax-ns#"] = "rdf"

		err = enc.EncodeAll(triples)

		if err != nil {
			t.Fatal(err)
		}
		err = enc.Close()
		if err != nil {
			t.Fatal(err)
		}
		if buf.String() != ttlBenchOutputs[i] {
			t.Fatalf("Decode/Encode roundtrip failed, re-encoding:\n%v\ngot:\n%v\nwant:\n%v", ex, buf.String(), ttlBenchOutputs[i])
		}
	}
}
func TestTTL(t *testing.T) {
	for _, test := range ttlTestSuite {
		dec := NewTripleDecoder(bytes.NewBufferString(test.input), Turtle)
		triples, err := dec.DecodeAll()
		if err != nil {
			if test.errWant == "" {
				t.Fatalf("ParseTTL(%s) => %v, want %v\ntriples:%v", test.input, err, test.want, triples)
				continue
			}
			if strings.HasSuffix(err.Error(), test.errWant) {
				continue
			}
			t.Fatalf("ParseTTL(%s) => %q, want %q", test.input, err, test.errWant)
			continue
		}

		if !reflect.DeepEqual(triples, test.want) {
			t.Fatalf("ParseTTL(%s) => %v,\nwant: %v", test.input, triples, test.want)
		}
	}
}

// ttlTestSuite is a representation of the official W3C test suite for Turtle
// which is found at: http://www.w3.org/2013/TurtleTests/
var ttlTestSuite = []struct {
	input   string
	errWant string
	want    []Triple
}{
	//# atomic tests
	//
	//<#IRI_subject> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "IRI_subject" ;
	//   rdfs:comment "IRI subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <IRI_subject.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#IRI_with_four_digit_numeric_escape> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "IRI_with_four_digit_numeric_escape" ;
	//   rdfs:comment "IRI with four digit numeric escape (\\u)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <IRI_with_four_digit_numeric_escape.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`<http://a.example/\u0073> <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#IRI_with_eight_digit_numeric_escape> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "IRI_with_eight_digit_numeric_escape" ;
	//   rdfs:comment "IRI with eight digit numeric escape (\\U)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <IRI_with_eight_digit_numeric_escape.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`<http://a.example/\U00000073> <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#IRI_with_all_punctuation> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "IRI_with_all_punctuation" ;
	//   rdfs:comment "IRI with all punctuation" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <IRI_with_all_punctuation.ttl> ;
	//   mf:result    <IRI_with_all_punctuation.nt> ;
	//   .

	{`<scheme:!$%25&amp;'()*+,-./0123456789:/@ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~?#> <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "scheme:!$%25&amp;'()*+,-./0123456789:/@ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~?#"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#bareword_a_predicate> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "bareword_a_predicate" ;
	//   rdfs:comment "bareword a predicate" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <bareword_a_predicate.ttl> ;
	//   mf:result    <bareword_a_predicate.nt> ;
	//   .

	{`<http://a.example/s> a <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#old_style_prefix> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "old_style_prefix" ;
	//   rdfs:comment "old-style prefix" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <old_style_prefix.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
p:s <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#SPARQL_style_prefix> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "SPARQL_style_prefix" ;
	//   rdfs:comment "SPARQL-style prefix" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <SPARQL_style_prefix.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`PREFIX p: <http://a.example/>
p:s <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#prefixed_IRI_predicate> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "prefixed_IRI_predicate" ;
	//   rdfs:comment "prefixed IRI predicate" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <prefixed_IRI_predicate.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
<http://a.example/s> p:p <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#prefixed_IRI_object> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "prefixed_IRI_object" ;
	//   rdfs:comment "prefixed IRI object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <prefixed_IRI_object.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
<http://a.example/s> <http://a.example/p> p:o .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#prefix_only_IRI> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "prefix_only_IRI" ;
	//   rdfs:comment "prefix-only IRI (p:)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <prefix_only_IRI.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`@prefix p: <http://a.example/s>.
p: <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#prefix_with_PN_CHARS_BASE_character_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "prefix_with_PN_CHARS_BASE_character_boundaries" ;
	//   rdfs:comment "prefix with PN CHARS BASE character boundaries (prefix: AZazÀÖØöø...:)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <prefix_with_PN_CHARS_BASE_character_boundaries.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`@prefix AZazÀÖØöø˿ͰͽͿ῿‌‍⁰↏Ⰰ⿯、퟿豈﷏ﷰ�𐀀󯿽: <http://a.example/> .
<http://a.example/s> <http://a.example/p> AZazÀÖØöø˿ͰͽͿ῿‌‍⁰↏Ⰰ⿯、퟿豈﷏ﷰ�𐀀󯿽:o .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#prefix_with_non_leading_extras> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "prefix_with_non_leading_extras" ;
	//   rdfs:comment "prefix with_non_leading_extras (_:a·̀ͯ‿.⁀)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <prefix_with_non_leading_extras.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`@prefix a·̀ͯ‿.⁀: <http://a.example/>.
a·̀ͯ‿.⁀:s <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#localName_with_assigned_nfc_bmp_PN_CHARS_BASE_character_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "localName_with_assigned_nfc_bmp_PN_CHARS_BASE_character_boundaries" ;
	//   rdfs:comment "localName with assigned, NFC-normalized, basic-multilingual-plane PN CHARS BASE character boundaries (p:AZazÀÖØöø...)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <localName_with_assigned_nfc_bmp_PN_CHARS_BASE_character_boundaries.ttl> ;
	//   mf:result    <localName_with_assigned_nfc_bmp_PN_CHARS_BASE_character_boundaries.nt> ;
	//   .

	{`@prefix p: <http://a.example/> .
<http://a.example/s> <http://a.example/p> p:AZazÀÖØöø˿Ͱͽ΄῾‌‍⁰↉Ⰰ⿕、ퟻ﨎ﷇﷰ￯ .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/AZazÀÖØöø˿Ͱͽ΄῾‌‍⁰↉Ⰰ⿕、ퟻ﨎ﷇﷰ￯"},
		},
	}},

	//<#localName_with_assigned_nfc_PN_CHARS_BASE_character_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "localName_with_assigned_nfc_PN_CHARS_BASE_character_boundaries" ;
	//   rdfs:comment "localName with assigned, NFC-normalized PN CHARS BASE character boundaries (p:AZazÀÖØöø...)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <localName_with_assigned_nfc_PN_CHARS_BASE_character_boundaries.ttl> ;
	//   mf:result    <localName_with_assigned_nfc_PN_CHARS_BASE_character_boundaries.nt> ;
	//   .

	{`@prefix p: <http://a.example/> .
<http://a.example/s> <http://a.example/p> p:AZazÀÖØöø˿Ͱͽ΄῾‌‍⁰↉Ⰰ⿕、ퟻ﨎ﷇﷰ￯𐀀󠇯 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/AZazÀÖØöø˿Ͱͽ΄῾‌‍⁰↉Ⰰ⿕、ퟻ﨎ﷇﷰ￯𐀀󠇯"},
		},
	}},

	//<#localName_with_nfc_PN_CHARS_BASE_character_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "localName_with_nfc_PN_CHARS_BASE_character_boundaries" ;
	//   rdfs:comment "localName with nfc-normalize PN CHARS BASE character boundaries (p:AZazÀÖØöø...)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <localName_with_nfc_PN_CHARS_BASE_character_boundaries.ttl> ;
	//   mf:result    <localName_with_nfc_PN_CHARS_BASE_character_boundaries.nt> ;
	//   .

	{`@prefix p: <http://a.example/> .
<http://a.example/s> <http://a.example/p> p:AZazÀÖØöø˿ͰͽͿ῿‌‍⁰↏Ⰰ⿯、퟿﨎﷏ﷰ￯𐀀󯿽 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/AZazÀÖØöø˿ͰͽͿ῿‌‍⁰↏Ⰰ⿯、퟿﨎﷏ﷰ￯𐀀󯿽"},
		},
	}},

	//<#default_namespace_IRI> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "default_namespace_IRI" ;
	//   rdfs:comment "default namespace IRI (:ln)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <default_namespace_IRI.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`@prefix : <http://a.example/>.
:s <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#prefix_reassigned_and_used> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "prefix_reassigned_and_used" ;
	//   rdfs:comment "prefix reassigned and used" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <prefix_reassigned_and_used.ttl> ;
	//   mf:result    <prefix_reassigned_and_used.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
@prefix p: <http://b.example/>.
p:s <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://b.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#reserved_escaped_localName> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "reserved_escaped_localName" ;
	//   rdfs:comment "reserved-escaped local name" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <reserved_escaped_localName.ttl> ;
	//   mf:result    <reserved_escaped_localName.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
p:\_\~\.-\!\$\&\'\(\)\*\+\,\;\=\/\?\#\@\%00 <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/_~.-!$&'()*+,;=/?#@%00`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#percent_escaped_localName> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "percent_escaped_localName" ;
	//   rdfs:comment "percent-escaped local name" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <percent_escaped_localName.ttl> ;
	//   mf:result    <percent_escaped_localName.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
p:%25 <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/%25`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#HYPHEN_MINUS_in_localName> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "HYPHEN_MINUS_in_localName" ;
	//   rdfs:comment "HYPHEN-MINUS in local name" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <HYPHEN_MINUS_in_localName.ttl> ;
	//   mf:result    <HYPHEN_MINUS_in_localName.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
p:s- <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/s-`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#underscore_in_localName> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "underscore_in_localName" ;
	//   rdfs:comment "underscore in local name" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <underscore_in_localName.ttl> ;
	//   mf:result    <underscore_in_localName.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
p:s_ <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/s_`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#localname_with_COLON> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "localname_with_COLON" ;
	//   rdfs:comment "localname with COLON" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <localname_with_COLON.ttl> ;
	//   mf:result    <localname_with_COLON.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
p:s: <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/s:`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#localName_with_leading_underscore> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "localName_with_leading_underscore" ;
	//   rdfs:comment "localName with leading underscore (p:_)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <localName_with_leading_underscore.ttl> ;
	//   mf:result    <localName_with_leading_underscore.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
p:_ <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/_`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#localName_with_leading_digit> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "localName_with_leading_digit" ;
	//   rdfs:comment "localName with leading digit (p:_)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <localName_with_leading_digit.ttl> ;
	//   mf:result    <localName_with_leading_digit.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
p:0 <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/0`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#localName_with_non_leading_extras> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "localName_with_non_leading_extras" ;
	//   rdfs:comment "localName with_non_leading_extras (_:a·̀ͯ‿.⁀)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <localName_with_non_leading_extras.ttl> ;
	//   mf:result    <localName_with_non_leading_extras.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
p:a·̀ͯ‿.⁀ <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/a·̀ͯ‿.⁀`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#old_style_base> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "old_style_base" ;
	//   rdfs:comment "old-style base" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <old_style_base.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`@base <http://a.example/>.
<s> <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/s`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#SPARQL_style_base> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "SPARQL_style_base" ;
	//   rdfs:comment "SPARQL-style base" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <SPARQL_style_base.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`BASE <http://a.example/>
<s> <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/s`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#labeled_blank_node_subject> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "labeled_blank_node_subject" ;
	//   rdfs:comment "labeled blank node subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <labeled_blank_node_subject.ttl> ;
	//   mf:result    <labeled_blank_node_subject.nt> ;
	//   .

	{`_:s <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#labeled_blank_node_object> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "labeled_blank_node_object" ;
	//   rdfs:comment "labeled blank node object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <labeled_blank_node_object.ttl> ;
	//   mf:result    <labeled_blank_node_object.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> _:o .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/s`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:o"},
		},
	}},

	//<#labeled_blank_node_with_PN_CHARS_BASE_character_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "labeled_blank_node_with_PN_CHARS_BASE_character_boundaries" ;
	//   rdfs:comment "labeled blank node with PN_CHARS_BASE character boundaries (_:AZazÀÖØöø...)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <labeled_blank_node_with_PN_CHARS_BASE_character_boundaries.ttl> ;
	//   mf:result    <labeled_blank_node_object.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> _:AZazÀÖØöø˿ͰͽͿ῿‌‍⁰↏Ⰰ⿯、퟿豈﷏ﷰ�𐀀󯿽 .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/s`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:AZazÀÖØöø˿ͰͽͿ῿‌‍⁰↏Ⰰ⿯、퟿豈﷏ﷰ�𐀀󯿽"},
		},
	}},

	//<#labeled_blank_node_with_leading_underscore> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "labeled_blank_node_with_leading_underscore" ;
	//   rdfs:comment "labeled blank node with_leading_underscore (_:_)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <labeled_blank_node_with_leading_underscore.ttl> ;
	//   mf:result    <labeled_blank_node_object.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> _:_ .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/s`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:_"},
		},
	}},

	//<#labeled_blank_node_with_leading_digit> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "labeled_blank_node_with_leading_digit" ;
	//   rdfs:comment "labeled blank node with_leading_digit (_:0)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <labeled_blank_node_with_leading_digit.ttl> ;
	//   mf:result    <labeled_blank_node_object.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> _:0 .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/s`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:0"},
		},
	}},

	//<#labeled_blank_node_with_non_leading_extras> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "labeled_blank_node_with_non_leading_extras" ;
	//   rdfs:comment "labeled blank node with_non_leading_extras (_:a·̀ͯ‿.⁀)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <labeled_blank_node_with_non_leading_extras.ttl> ;
	//   mf:result    <labeled_blank_node_object.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> _:a·̀ͯ‿.⁀ .`, "", []Triple{
		Triple{
			Subj: IRI{str: `http://a.example/s`},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:a·̀ͯ‿.⁀"},
		},
	}},

	//<#anonymous_blank_node_subject> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "anonymous_blank_node_subject" ;
	//   rdfs:comment "anonymous blank node subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <anonymous_blank_node_subject.ttl> ;
	//   mf:result    <labeled_blank_node_subject.nt> ;
	//   .

	{`[] <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#anonymous_blank_node_object> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "anonymous_blank_node_object" ;
	//   rdfs:comment "anonymous blank node object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <anonymous_blank_node_object.ttl> ;
	//   mf:result    <labeled_blank_node_object.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> [] .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:b1"},
		},
	}},

	//<#sole_blankNodePropertyList> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "sole_blankNodePropertyList" ;
	//   rdfs:comment "sole blankNodePropertyList [ <p> <o> ] ." ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <sole_blankNodePropertyList.ttl> ;
	//   mf:result    <labeled_blank_node_subject.nt> ;
	//   .

	{`[ <http://a.example/p> <http://a.example/o> ] .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#blankNodePropertyList_as_subject> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "blankNodePropertyList_as_subject" ;
	//   rdfs:comment "blankNodePropertyList as subject [ … ] <p> <o> ." ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <blankNodePropertyList_as_subject.ttl> ;
	//   mf:result    <blankNodePropertyList_as_subject.nt> ;
	//   .

	{`[ <http://a.example/p> <http://a.example/o> ] <http://a.example/p2> <http://a.example/o2> .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p2"},
			Obj:  IRI{str: "http://a.example/o2"},
		},
	}},

	//<#blankNodePropertyList_as_object> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "blankNodePropertyList_as_object" ;
	//   rdfs:comment "blankNodePropertyList as object <s> <p> [ … ] ." ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <blankNodePropertyList_as_object.ttl> ;
	//   mf:result    <blankNodePropertyList_as_object.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> [ <http://a.example/p2> <http://a.example/o2> ] .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:b1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p2"},
			Obj:  IRI{str: "http://a.example/o2"},
		},
	}},

	//<#blankNodePropertyList_with_multiple_triples> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "blankNodePropertyList_with_multiple_triples" ;
	//   rdfs:comment "blankNodePropertyList with multiple triples [ <s> <p> ; <s2> <p2> ]" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <blankNodePropertyList_with_multiple_triples.ttl> ;
	//   mf:result    <blankNodePropertyList_with_multiple_triples.nt> ;
	//   .

	{`[ <http://a.example/p1> <http://a.example/o1> ; <http://a.example/p2> <http://a.example/o2> ] <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p1"},
			Obj:  IRI{str: "http://a.example/o1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p2"},
			Obj:  IRI{str: "http://a.example/o2"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#nested_blankNodePropertyLists> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "nested_blankNodePropertyLists" ;
	//   rdfs:comment "nested blankNodePropertyLists [ <p1> [ <p2> <o2> ] ; <p3> <o3> ]" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nested_blankNodePropertyLists.ttl> ;
	//   mf:result    <nested_blankNodePropertyLists.nt> ;
	//   .

	{`[ <http://a.example/p1> [ <http://a.example/p2> <http://a.example/o2> ] ; <http://a.example/p> <http://a.example/o> ].`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p1"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://a.example/p2"},
			Obj:  IRI{str: "http://a.example/o2"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#blankNodePropertyList_containing_collection> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "blankNodePropertyList_containing_collection" ;
	//   rdfs:comment "blankNodePropertyList containing collection [ <p1> ( … ) ]" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <blankNodePropertyList_containing_collection.ttl> ;
	//   mf:result    <blankNodePropertyList_containing_collection.nt> ;
	//   .

	{`[ <http://a.example/p1> (1) ] .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p1"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#collection_subject> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "collection_subject" ;
	//   rdfs:comment "collection subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <collection_subject.ttl> ;
	//   mf:result    <collection_subject.nt> ;
	//   .

	{`(1) <http://a.example/p> <http://a.example/o> .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#collection_object> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "collection_object" ;
	//   rdfs:comment "collection object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <collection_object.ttl> ;
	//   mf:result    <collection_object.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> (1) .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:b1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#empty_collection> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "empty_collection" ;
	//   rdfs:comment "empty collection ()" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <empty_collection.ttl> ;
	//   mf:result    <empty_collection.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> () .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#nested_collection> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "nested_collection" ;
	//   rdfs:comment "nested collection (())" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <nested_collection.ttl> ;
	//   mf:result    <nested_collection.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> ((1)) .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:b1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#first> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "first" ;
	//   rdfs:comment "first, not last, non-empty nested collection" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <first.ttl> ;
	//   mf:result    <first.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> ((1) 2) .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:b1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  Blank{id: "_:b3"},
		},
		Triple{
			Subj: Blank{id: "_:b3"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "2", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b3"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#last> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "last" ;
	//   rdfs:comment "last, not first, non-empty nested collection" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <last.ttl> ;
	//   mf:result    <last.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> (1 (2)) .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Blank{id: "_:b1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Blank{id: "_:b3"},
		},
		Triple{
			Subj: Blank{id: "_:b3"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "2", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b3"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#LITERAL1> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL1" ;
	//   rdfs:comment "LITERAL1 'x'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL1.ttl> ;
	//   mf:result    <LITERAL1.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> 'x' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "x", DataType: xsdString},
		},
	}},

	//<#LITERAL1_ascii_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL1_ascii_boundaries" ;
	//   rdfs:comment "LITERAL1_ascii_boundaries '\\x00\\x09\\x0b\\x0c\\x0e\\x26\\x28...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL1_ascii_boundaries.ttl> ;
	//   mf:result    <LITERAL1_ascii_boundaries.nt> ;
	//   .

	{"<http://a.example/s> <http://a.example/p> '\x00	&([]' .", "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\u0000\t\u000B\u000C\u000E&([]\u007F", DataType: xsdString},
		},
	}},

	//<#LITERAL1_with_UTF8_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL1_with_UTF8_boundaries" ;
	//   rdfs:comment "LITERAL1_with_UTF8_boundaries '\\x80\\x7ff\\x800\\xfff...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL1_with_UTF8_boundaries.ttl> ;
	//   mf:result    <LITERAL_with_UTF8_boundaries.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽", DataType: xsdString},
		},
	}},

	//<#LITERAL1_all_controls> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL1_all_controls" ;
	//   rdfs:comment "LITERAL1_all_controls '\\x00\\x01\\x02\\x03\\x04...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL1_all_controls.ttl> ;
	//   mf:result    <LITERAL1_all_controls.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "\u0000\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u0008\t\u000B\u000C\u000E\u000F\u0010\u0011\u0012\u0013\u0014\u0015\u0016\u0017\u0018\u0019\u001A\u001B\u001C\u001D\u001E\u001F" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\u0000\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u0008\t\u000B\u000C\u000E\u000F\u0010\u0011\u0012\u0013\u0014\u0015\u0016\u0017\u0018\u0019\u001A\u001B\u001C\u001D\u001E\u001F", DataType: xsdString},
		},
	}},

	//<#LITERAL1_all_punctuation> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL1_all_punctuation" ;
	//   rdfs:comment "LITERAL1_all_punctuation '!\"#$%&()...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL1_all_punctuation.ttl> ;
	//   mf:result    <LITERAL1_all_punctuation.nt> ;
	//   .

	{"<http://a.example/s> <http://a.example/p> ' !\"#$%&():;<=>?@[]^_`{|}~' .", "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: " !\"#$%&():;<=>?@[]^_`{|}~", DataType: xsdString},
		},
	}},

	//<#LITERAL_LONG1> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL_LONG1" ;
	//   rdfs:comment "LITERAL_LONG1 '''x'''" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG1.ttl> ;
	//   mf:result    <LITERAL1.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '''x''' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "x", DataType: xsdString},
		},
	}},

	//<#LITERAL_LONG1_ascii_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL_LONG1_ascii_boundaries" ;
	//   rdfs:comment "LITERAL_LONG1_ascii_boundaries '\\x00\\x26\\x28...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG1_ascii_boundaries.ttl> ;
	//   mf:result    <LITERAL_LONG1_ascii_boundaries.nt> ;
	//   .

	{"<http://a.example/s> <http://a.example/p> '\x00&([]' .", "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\x00&([]", DataType: xsdString},
		},
	}},

	//<#LITERAL_LONG1_with_UTF8_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL_LONG1_with_UTF8_boundaries" ;
	//   rdfs:comment "LITERAL_LONG1_with_UTF8_boundaries '\\x80\\x7ff\\x800\\xfff...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG1_with_UTF8_boundaries.ttl> ;
	//   mf:result    <LITERAL_with_UTF8_boundaries.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '''߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽''' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽", DataType: xsdString},
		},
	}},

	//<#LITERAL_LONG1_with_1_squote> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL_LONG1_with_1_squote" ;
	//   rdfs:comment "LITERAL_LONG1 with 1 squote '''a'b'''" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG1_with_1_squote.ttl> ;
	//   mf:result    <LITERAL_LONG1_with_1_squote.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '''x'y''' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "x'y", DataType: xsdString},
		},
	}},

	//<#LITERAL_LONG1_with_2_squotes> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL_LONG1_with_2_squotes" ;
	//   rdfs:comment "LITERAL_LONG1 with 2 squotes '''a''b'''" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG1_with_2_squotes.ttl> ;
	//   mf:result    <LITERAL_LONG1_with_2_squotes.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '''x''y''' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "x''y", DataType: xsdString},
		},
	}},

	//<#LITERAL2> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL2" ;
	//   rdfs:comment "LITERAL2 \"x\"" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL2.ttl> ;
	//   mf:result    <LITERAL1.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "x" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "x", DataType: xsdString},
		},
	}},

	//<#LITERAL2_ascii_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL2_ascii_boundaries" ;
	//   rdfs:comment "LITERAL2_ascii_boundaries '\\x00\\x09\\x0b\\x0c\\x0e\\x21\\x23...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL2_ascii_boundaries.ttl> ;
	//   mf:result    <LITERAL2_ascii_boundaries.nt> ;
	//   .

	{"<http://a.example/s> <http://a.example/p> \"\x00	!#[]\" .", "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj: Literal{str: "\x00	!#[]", DataType: xsdString},
		},
	}},

	//<#LITERAL2_with_UTF8_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL2_with_UTF8_boundaries" ;
	//   rdfs:comment "LITERAL2_with_UTF8_boundaries '\\x80\\x7ff\\x800\\xfff...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL2_with_UTF8_boundaries.ttl> ;
	//   mf:result    <LITERAL_with_UTF8_boundaries.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽", DataType: xsdString},
		},
	}},

	//<#LITERAL_LONG2> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL_LONG2" ;
	//   rdfs:comment "LITERAL_LONG2 \"\"\"x\"\"\"" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG2.ttl> ;
	//   mf:result    <LITERAL1.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> """x""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "x", DataType: xsdString},
		},
	}},

	//<#LITERAL_LONG2_ascii_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL_LONG2_ascii_boundaries" ;
	//   rdfs:comment "LITERAL_LONG2_ascii_boundaries '\\x00\\x21\\x23...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG2_ascii_boundaries.ttl> ;
	//   mf:result    <LITERAL_LONG2_ascii_boundaries.nt> ;
	//   .

	{"<http://a.example/s> <http://a.example/p> \"\x00!#[]\" .", "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\x00!#[]", DataType: xsdString},
		},
	}},

	//<#LITERAL_LONG2_with_UTF8_boundaries> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL_LONG2_with_UTF8_boundaries" ;
	//   rdfs:comment "LITERAL_LONG2_with_UTF8_boundaries '\\x80\\x7ff\\x800\\xfff...'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG2_with_UTF8_boundaries.ttl> ;
	//   mf:result    <LITERAL_with_UTF8_boundaries.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> """߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "߿ࠀ࿿က쿿퀀퟿�𐀀𿿽񀀀󿿽􀀀􏿽", DataType: xsdString},
		},
	}},

	//<#LITERAL_LONG2_with_1_squote> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL_LONG2_with_1_squote" ;
	//   rdfs:comment "LITERAL_LONG2 with 1 squote \"\"\"a\"b\"\"\"" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG2_with_1_squote.ttl> ;
	//   mf:result    <LITERAL_LONG2_with_1_squote.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> """x"y""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: `x"y`, DataType: xsdString},
		},
	}},

	//<#LITERAL_LONG2_with_2_squotes> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "LITERAL_LONG2_with_2_squotes" ;
	//   rdfs:comment "LITERAL_LONG2 with 2 squotes \"\"\"a\"\"b\"\"\"" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG2_with_2_squotes.ttl> ;
	//   mf:result    <LITERAL_LONG2_with_2_squotes.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> """x""y""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: `x""y`, DataType: xsdString},
		},
	}},

	//<#literal_with_CHARACTER_TABULATION> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_CHARACTER_TABULATION" ;
	//   rdfs:comment "literal with CHARACTER TABULATION" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_CHARACTER_TABULATION.ttl> ;
	//   mf:result    <literal_with_CHARACTER_TABULATION.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '	' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj: Literal{str: `	`, DataType: xsdString},
		},
	}},

	//<#literal_with_BACKSPACE> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_BACKSPACE" ;
	//   rdfs:comment "literal with BACKSPACE" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_BACKSPACE.ttl> ;
	//   mf:result    <literal_with_BACKSPACE.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj: Literal{str: "", DataType: xsdString},
		},
	}},

	//<#literal_with_LINE_FEED> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_LINE_FEED" ;
	//   rdfs:comment "literal with LINE FEED" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_LINE_FEED.ttl> ;
	//   mf:result    <literal_with_LINE_FEED.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '''
''' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\n", DataType: xsdString},
		},
	}},

	//<#literal_with_CARRIAGE_RETURN> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_CARRIAGE_RETURN" ;
	//   rdfs:comment "literal with CARRIAGE RETURN" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_CARRIAGE_RETURN.ttl> ;
	//   mf:result    <literal_with_CARRIAGE_RETURN.nt> ;
	//   .

	{"<http://a.example/s> <http://a.example/p> '''\r''' .", "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\r", DataType: xsdString},
		},
	}},

	//<#literal_with_FORM_FEED> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_FORM_FEED" ;
	//   rdfs:comment "literal with FORM FEED" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_FORM_FEED.ttl> ;
	//   mf:result    <literal_with_FORM_FEED.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj: Literal{str: "", DataType: xsdString},
		},
	}},

	//<#literal_with_REVERSE_SOLIDUS> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_REVERSE_SOLIDUS" ;
	//   rdfs:comment "literal with REVERSE SOLIDUS" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_REVERSE_SOLIDUS.ttl> ;
	//   mf:result    <literal_with_REVERSE_SOLIDUS.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '\\' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: `\`, DataType: xsdString},
		},
	}},

	//<#literal_with_escaped_CHARACTER_TABULATION> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_escaped_CHARACTER_TABULATION" ;
	//   rdfs:comment "literal with escaped CHARACTER TABULATION" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_escaped_CHARACTER_TABULATION.ttl> ;
	//   mf:result    <literal_with_CHARACTER_TABULATION.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '\t' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\t", DataType: xsdString},
		},
	}},

	//<#literal_with_escaped_BACKSPACE> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_escaped_BACKSPACE" ;
	//   rdfs:comment "literal with escaped BACKSPACE" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_escaped_BACKSPACE.ttl> ;
	//   mf:result    <literal_with_BACKSPACE.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '\b' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\b", DataType: xsdString},
		},
	}},

	//<#literal_with_escaped_LINE_FEED> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_escaped_LINE_FEED" ;
	//   rdfs:comment "literal with escaped LINE FEED" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_escaped_LINE_FEED.ttl> ;
	//   mf:result    <literal_with_LINE_FEED.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '\n' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\n", DataType: xsdString},
		},
	}},

	//<#literal_with_escaped_CARRIAGE_RETURN> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_escaped_CARRIAGE_RETURN" ;
	//   rdfs:comment "literal with escaped CARRIAGE RETURN" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_escaped_CARRIAGE_RETURN.ttl> ;
	//   mf:result    <literal_with_CARRIAGE_RETURN.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '\r' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\r", DataType: xsdString},
		},
	}},

	//<#literal_with_escaped_FORM_FEED> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_escaped_FORM_FEED" ;
	//   rdfs:comment "literal with escaped FORM FEED" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_escaped_FORM_FEED.ttl> ;
	//   mf:result    <literal_with_FORM_FEED.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '\f' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "\f", DataType: xsdString},
		},
	}},

	//<#literal_with_numeric_escape4> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_numeric_escape4" ;
	//   rdfs:comment "literal with numeric escape4 \\u" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_numeric_escape4.ttl> ;
	//   mf:result    <literal_with_numeric_escape4.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '\u006F' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "o", DataType: xsdString},
		},
	}},

	//<#literal_with_numeric_escape8> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_with_numeric_escape8" ;
	//   rdfs:comment "literal with numeric escape8 \\U" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_with_numeric_escape8.ttl> ;
	//   mf:result    <literal_with_numeric_escape4.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> '\U0000006F' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "o", DataType: xsdString},
		},
	}},

	//<#IRIREF_datatype> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "IRIREF_datatype" ;
	//   rdfs:comment "IRIREF datatype \"\"^^<t>" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <IRIREF_datatype.ttl> ;
	//   mf:result    <IRIREF_datatype.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "1"^^<http://www.w3.org/2001/XMLSchema#integer> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
	}},

	//<#prefixed_name_datatype> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "prefixed_name_datatype" ;
	//   rdfs:comment "prefixed name datatype \"\"^^p:t" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <prefixed_name_datatype.ttl> ;
	//   mf:result    <IRIREF_datatype.nt> ;
	//   .

	{`@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .
<http://a.example/s> <http://a.example/p> "1"^^xsd:integer .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
	}},

	//<#bareword_integer> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "bareword_integer" ;
	//   rdfs:comment "bareword integer" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <bareword_integer.ttl> ;
	//   mf:result    <IRIREF_datatype.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> 1 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
	}},

	//<#bareword_decimal> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "bareword_decimal" ;
	//   rdfs:comment "bareword decimal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <bareword_decimal.ttl> ;
	//   mf:result    <bareword_decimal.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> 1.0 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "1.0", DataType: xsdDecimal},
		},
	}},

	//<#bareword_double> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "bareword_double" ;
	//   rdfs:comment "bareword double" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <bareword_double.ttl> ;
	//   mf:result    <bareword_double.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> 1E0 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "1E0", DataType: xsdDouble},
		},
	}},

	//<#double_lower_case_e> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "double_lower_case_e" ;
	//   rdfs:comment "double lower case e" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <double_lower_case_e.ttl> ;
	//   mf:result    <double_lower_case_e.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> 1e0 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "1e0", DataType: xsdDouble},
		},
	}},

	//<#negative_numeric> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "negative_numeric" ;
	//   rdfs:comment "negative numeric" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <negative_numeric.ttl> ;
	//   mf:result    <negative_numeric.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> -1 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "-1", DataType: xsdInteger},
		},
	}},

	//<#positive_numeric> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "positive_numeric" ;
	//   rdfs:comment "positive numeric" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <positive_numeric.ttl> ;
	//   mf:result    <positive_numeric.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> +1 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "+1", DataType: xsdInteger},
		},
	}},

	//<#numeric_with_leading_0> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "numeric_with_leading_0" ;
	//   rdfs:comment "numeric with leading 0" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <numeric_with_leading_0.ttl> ;
	//   mf:result    <numeric_with_leading_0.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> 01 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "01", DataType: xsdInteger},
		},
	}},

	//<#literal_true> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_true" ;
	//   rdfs:comment "literal true" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_true.ttl> ;
	//   mf:result    <literal_true.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> true .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "true", DataType: xsdBoolean},
		},
	}},

	//<#literal_false> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "literal_false" ;
	//   rdfs:comment "literal false" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <literal_false.ttl> ;
	//   mf:result    <literal_false.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> false .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "false", DataType: xsdBoolean},
		},
	}},

	//<#langtagged_non_LONG> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "langtagged_non_LONG" ;
	//   rdfs:comment "langtagged non-LONG \"x\"@en" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <langtagged_non_LONG.ttl> ;
	//   mf:result    <langtagged_non_LONG.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "chat"@en .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "chat", lang: "en", DataType: rdfLangString},
		},
	}},

	//<#langtagged_LONG> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "langtagged_LONG" ;
	//   rdfs:comment "langtagged LONG \"\"\"x\"\"\"@en" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <langtagged_LONG.ttl> ;
	//   mf:result    <langtagged_non_LONG.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> """chat"""@en .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "chat", lang: "en", DataType: rdfLangString},
		},
	}},

	//<#lantag_with_subtag> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "lantag_with_subtag" ;
	//   rdfs:comment "lantag with subtag \"x\"@en-us" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <lantag_with_subtag.ttl> ;
	//   mf:result    <lantag_with_subtag.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> "chat"@en-us .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  Literal{str: "chat", lang: "en-us", DataType: rdfLangString},
		},
	}},

	//<#objectList_with_two_objects> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "objectList_with_two_objects" ;
	//   rdfs:comment "objectList with two objects … <o1>,<o2>" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <objectList_with_two_objects.ttl> ;
	//   mf:result    <objectList_with_two_objects.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p> <http://a.example/o1>, <http://a.example/o2> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o1"},
		},
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o2"},
		},
	}},

	//<#predicateObjectList_with_two_objectLists> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "predicateObjectList_with_two_objectLists" ;
	//   rdfs:comment "predicateObjectList with two objectLists … <o1>,<o2>" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <predicateObjectList_with_two_objectLists.ttl> ;
	//   mf:result    <predicateObjectList_with_two_objectLists.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p1> <http://a.example/o1>; <http://a.example/p2> <http://a.example/o2> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p1"},
			Obj:  IRI{str: "http://a.example/o1"},
		},
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p2"},
			Obj:  IRI{str: "http://a.example/o2"},
		},
	}},

	//<#repeated_semis_at_end> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "repeated_semis_at_end" ;
	//   rdfs:comment "repeated semis at end <s> <p> <o> ;; <p2> <o2> ." ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <repeated_semis_at_end.ttl> ;
	//   mf:result    <predicateObjectList_with_two_objectLists.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p1> <http://a.example/o1>;; <http://a.example/p2> <http://a.example/o2> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p1"},
			Obj:  IRI{str: "http://a.example/o1"},
		},
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p2"},
			Obj:  IRI{str: "http://a.example/o2"},
		},
	}},

	//<#repeated_semis_not_at_end> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "repeated_semis_not_at_end" ;
	//   rdfs:comment "repeated semis not at end <s> <p> <o> ;;." ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <repeated_semis_not_at_end.ttl> ;
	//   mf:result    <repeated_semis_not_at_end.nt> ;
	//   .

	{`<http://a.example/s> <http://a.example/p1> <http://a.example/o1>;; .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p1"},
			Obj:  IRI{str: "http://a.example/o1"},
		},
	}},

	//# original tests-ttl
	//<#turtle-syntax-file-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-file-01" ;
	//   rdfs:comment "Empty file" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-file-01.ttl> ;
	//   .

	{``, "", nil},

	//<#turtle-syntax-file-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-file-02" ;
	//   rdfs:comment "Only comment" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-file-02.ttl> ;
	//   .

	{`#Empty file.`, "", nil},

	//<#turtle-syntax-file-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-file-03" ;
	//   rdfs:comment "One comment, one empty line" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-file-03.ttl> ;
	//   .

	{`#One comment, one empty line.
`, "", nil},

	//<#turtle-syntax-uri-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-uri-01" ;
	//   rdfs:comment "Only IRIs" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-uri-01.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-uri-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-uri-02" ;
	//   rdfs:comment "IRIs with Unicode escape" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-uri-02.ttl> ;
	//   .

	{`# x53 is capital S
<http://www.w3.org/2013/TurtleTests/\u0053> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/S"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-uri-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-uri-03" ;
	//   rdfs:comment "IRIs with long Unicode escape" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-uri-03.ttl> ;
	//   .

	{`# x53 is capital S
<http://www.w3.org/2013/TurtleTests/\U00000053> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/S"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-uri-04> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-uri-04" ;
	//   rdfs:comment "Legal IRIs" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-uri-04.ttl> ;
	//   .

	{`# IRI with all chars in it.
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p>
<scheme:!$%25&'()*+,-./0123456789:/@ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~?#> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "scheme:!$%25&'()*+,-./0123456789:/@ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~?#"},
		},
	}},

	//<#turtle-syntax-base-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-base-01" ;
	//   rdfs:comment "@base" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-base-01.ttl> ;
	//   .

	{`@base <http://www.w3.org/2013/TurtleTests/> .`, "", nil},

	//<#turtle-syntax-base-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-base-02" ;
	//   rdfs:comment "BASE" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-base-02.ttl> ;
	//   .

	{`BASE <http://www.w3.org/2013/TurtleTests/>`, "", nil},

	//<#turtle-syntax-base-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-base-03" ;
	//   rdfs:comment "@base with relative IRIs" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-base-03.ttl> ;
	//   .

	{`@base <http://www.w3.org/2013/TurtleTests/> .
<s> <p> <o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-base-04> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-base-04" ;
	//   rdfs:comment "base with relative IRIs" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-base-04.ttl> ;
	//   .

	{`base <http://www.w3.org/2013/TurtleTests/>
<s> <p> <o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-prefix-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-prefix-01" ;
	//   rdfs:comment "@prefix" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-prefix-01.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .`, "", nil},

	//<#turtle-syntax-prefix-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-prefix-02" ;
	//   rdfs:comment "PreFIX" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-prefix-02.ttl> ;
	//   .

	{`PreFIX : <http://www.w3.org/2013/TurtleTests/>`, "", nil},

	//<#turtle-syntax-prefix-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-prefix-03" ;
	//   rdfs:comment "Empty PREFIX" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-prefix-03.ttl> ;
	//   .

	{`PREFIX : <http://www.w3.org/2013/TurtleTests/>
:s :p :123 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/123"},
		},
	}},

	//<#turtle-syntax-prefix-04> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-prefix-04" ;
	//   rdfs:comment "Empty @prefix with % escape" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-prefix-04.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p :%20 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/%20"},
		},
	}},

	//<#turtle-syntax-prefix-05> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-prefix-05" ;
	//   rdfs:comment "@prefix with no suffix" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-prefix-05.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
: : : .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/"},
		},
	}},

	//<#turtle-syntax-prefix-06> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-prefix-06" ;
	//   rdfs:comment "colon is a legal pname character" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-prefix-06.ttl> ;
	//   .

	{`# colon is a legal pname character
@prefix : <http://www.w3.org/2013/TurtleTests/> .
@prefix x: <http://www.w3.org/2013/TurtleTests/> .
:a:b:c  x:d:e:f :::: .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/a:b:c"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/d:e:f"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/:::"},
		},
	}},

	//<#turtle-syntax-prefix-07> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-prefix-07" ;
	//   rdfs:comment "dash is a legal pname character" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-prefix-07.ttl> ;
	//   .

	{`# dash is a legal pname character
@prefix x: <http://www.w3.org/2013/TurtleTests/> .
x:a-b-c  x:p x:o .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/a-b-c"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-prefix-08> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-prefix-08" ;
	//   rdfs:comment "underscore is a legal pname character" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-prefix-08.ttl> ;
	//   .

	{`# underscore is a legal pname character
@prefix x: <http://www.w3.org/2013/TurtleTests/> .
x:_  x:p_1 x:o .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/_"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p_1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-prefix-09> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-prefix-09" ;
	//   rdfs:comment "percents in pnames" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-prefix-09.ttl> ;
	//   .

	{`# percents
@prefix : <http://www.w3.org/2013/TurtleTests/> .
@prefix x: <http://www.w3.org/2013/TurtleTests/> .
:a%3E  x:%25 :a%3Eb .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/a%3E"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/%25"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/a%3Eb"},
		},
	}},

	//<#turtle-syntax-string-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-01" ;
	//   rdfs:comment "string literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-01.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "string" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "string", DataType: xsdString},
		},
	}},

	//<#turtle-syntax-string-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-02" ;
	//   rdfs:comment "langString literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-02.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "string"@en .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "string", DataType: rdfLangString, lang: "en"},
		},
	}},

	//<#turtle-syntax-string-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-03" ;
	//   rdfs:comment "langString literal with region" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-03.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "string"@en-uk .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "string", DataType: rdfLangString, lang: "en-uk"},
		},
	}},

	//<#turtle-syntax-string-04> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-04" ;
	//   rdfs:comment "squote string literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-04.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> 'string' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "string", DataType: xsdString},
		},
	}},

	//<#turtle-syntax-string-05> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-05" ;
	//   rdfs:comment "squote langString literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-05.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> 'string'@en .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "string", DataType: rdfLangString, lang: "en"},
		},
	}},

	//<#turtle-syntax-string-06> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-06" ;
	//   rdfs:comment "squote langString literal with region" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-06.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> 'string'@en-uk .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "string", DataType: rdfLangString, lang: "en-uk"},
		},
	}},

	//<#turtle-syntax-string-07> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-07" ;
	//   rdfs:comment "long string literal with embedded single- and double-quotes" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-07.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> """abc""def''ghi""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: `abc""def''ghi`, DataType: xsdString},
		},
	}},

	//<#turtle-syntax-string-08> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-08" ;
	//   rdfs:comment "long string literal with embedded newline" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-08.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> """abc
def""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "abc\ndef", DataType: xsdString},
		},
	}},

	//<#turtle-syntax-string-09> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-09" ;
	//   rdfs:comment "squote long string literal with embedded single- and double-quotes" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-09.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> '''abc
def''' .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "abc\ndef", DataType: xsdString},
		},
	}},

	//<#turtle-syntax-string-10> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-10" ;
	//   rdfs:comment "long langString literal with embedded newline" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-10.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> """abc
def"""@en .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "abc\ndef", DataType: rdfLangString, lang: "en"},
		},
	}},

	//<#turtle-syntax-string-11> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-string-11" ;
	//   rdfs:comment "squote long langString literal with embedded newline" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-string-11.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> '''abc
def'''@en .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "abc\ndef", DataType: rdfLangString, lang: "en"},
		},
	}},

	//<#turtle-syntax-str-esc-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-str-esc-01" ;
	//   rdfs:comment "string literal with escaped newline" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-str-esc-01.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "a\n" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "a\n", DataType: xsdString},
		},
	}},

	//<#turtle-syntax-str-esc-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-str-esc-02" ;
	//   rdfs:comment "string literal with Unicode escape" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-str-esc-02.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "a\u0020b" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "a b", DataType: xsdString},
		},
	}},

	//<#turtle-syntax-str-esc-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-str-esc-03" ;
	//   rdfs:comment "string literal with long Unicode escape" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-str-esc-03.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "a\U00000020b" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Literal{str: "a b", DataType: xsdString},
		},
	}},

	//<#turtle-syntax-pname-esc-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-pname-esc-01" ;
	//   rdfs:comment "pname with back-slash escapes" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-pname-esc-01.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p :\~\.-\!\$\&\'\(\)\*\+\,\;\=\/\?\#\@\_\%AA .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/~.-!$&'()*+,;=/?#@_%AA"},
		},
	}},

	//<#turtle-syntax-pname-esc-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-pname-esc-02" ;
	//   rdfs:comment "pname with back-slash escapes (2)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-pname-esc-02.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p :0123\~\.-\!\$\&\'\(\)\*\+\,\;\=\/\?\#\@\_\%AA123 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/0123~.-!$&'()*+,;=/?#@_%AA123"},
		},
	}},

	//<#turtle-syntax-pname-esc-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-pname-esc-03" ;
	//   rdfs:comment "pname with back-slash escapes (3)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-pname-esc-03.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:xyz\~ :abc\.:  : .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/xyz~"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/abc.:"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/"},
		},
	}},

	//<#turtle-syntax-bnode-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-bnode-01" ;
	//   rdfs:comment "bnode subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bnode-01.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
[] :p :o .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-bnode-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-bnode-02" ;
	//   rdfs:comment "bnode object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bnode-02.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p [] .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Blank{id: "_:b1"},
		},
	}},

	//<#turtle-syntax-bnode-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-bnode-03" ;
	//   rdfs:comment "bnode property list object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bnode-03.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p [ :q :o ] .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Blank{id: "_:b1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/q"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-bnode-04> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-bnode-04" ;
	//   rdfs:comment "bnode property list object (2)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bnode-04.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p [ :q1 :o1 ; :q2 :o2 ] .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Blank{id: "_:b1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/q1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/q2"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o2"},
		},
	}},

	//<#turtle-syntax-bnode-05> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-bnode-05" ;
	//   rdfs:comment "bnode property list subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bnode-05.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
[ :q1 :o1 ; :q2 :o2 ] :p :o .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/q1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/q2"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o2"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-bnode-06> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-bnode-06" ;
	//   rdfs:comment "labeled bnode subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bnode-06.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
_:a  :p :o .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:a"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-bnode-07> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-bnode-07" ;
	//   rdfs:comment "labeled bnode subject and object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bnode-07.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s  :p _:a .
_:a  :p :o .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Blank{id: "_:a"},
		},
		Triple{
			Subj: Blank{id: "_:a"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-bnode-08> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-bnode-08" ;
	//   rdfs:comment "bare bnode property list" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bnode-08.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
[ :p  :o ] .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-bnode-09> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-bnode-09" ;
	//   rdfs:comment "bnode property list" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bnode-09.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
[ :p  :o1,:2 ] .
:s :p :o  .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/2"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-bnode-10> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-bnode-10" ;
	//   rdfs:comment "mixed bnode property list and triple" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bnode-10.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .

:s1 :p :o .
[ :p1  :o1 ; :p2 :o2 ] .
:s2 :p :o .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p2"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o2"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s2"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-number-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-01" ;
	//   rdfs:comment "integer literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-01.ttl> ;
	//   .

	{`<s> <p> 123 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "123", DataType: xsdInteger},
		},
	}},

	//<#turtle-syntax-number-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-02" ;
	//   rdfs:comment "negative integer literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-02.ttl> ;
	//   .

	{`<s> <p> -123 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "-123", DataType: xsdInteger},
		},
	}},

	//<#turtle-syntax-number-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-03" ;
	//   rdfs:comment "positive integer literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-03.ttl> ;
	//   .

	{`<s> <p> +123 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "+123", DataType: xsdInteger},
		},
	}},

	//<#turtle-syntax-number-04> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-04" ;
	//   rdfs:comment "decimal literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-04.ttl> ;
	//   .

	{`# This is a decimal.
<s> <p> 123.0 . `, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "123.0", DataType: xsdDecimal},
		},
	}},

	//<#turtle-syntax-number-05> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-05" ;
	//   rdfs:comment "decimal literal (no leading digits)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-05.ttl> ;
	//   .

	{`# This is a decimal.
<s> <p> .1 . `, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: ".1", DataType: xsdDecimal},
		},
	}},

	//<#turtle-syntax-number-06> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-06" ;
	//   rdfs:comment "negative decimal literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-06.ttl> ;
	//   .

	{`# This is a decimal.
<s> <p> -123.0 . `, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "-123.0", DataType: xsdDecimal},
		},
	}},

	//<#turtle-syntax-number-07> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-07" ;
	//   rdfs:comment "positive decimal literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-07.ttl> ;
	//   .

	{`# This is a decimal.
<s> <p> +123.0 . `, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "+123.0", DataType: xsdDecimal},
		},
	}},

	//<#turtle-syntax-number-08> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-08" ;
	//   rdfs:comment "integer literal with decimal lexical confusion" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-08.ttl> ;
	//   .

	{`# This is an integer
<s> <p> 123.`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "123", DataType: xsdInteger},
		},
	}},

	//<#turtle-syntax-number-09> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-09" ;
	//   rdfs:comment "double literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-09.ttl> ;
	//   .

	{`<s> <p> 123.0e1 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "123.0e1", DataType: xsdDouble},
		},
	}},

	//<#turtle-syntax-number-10> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-10" ;
	//   rdfs:comment "negative double literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-10.ttl> ;
	//   .

	{`<s> <p> -123e-1 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "-123e-1", DataType: xsdDouble},
		},
	}},

	//<#turtle-syntax-number-11> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-number-11" ;
	//   rdfs:comment "double literal no fraction" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-number-11.ttl> ;
	//   .

	{`<s> <p> 123.E+1 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "123.E+1", DataType: xsdDouble},
		},
	}},

	//<#turtle-syntax-datatypes-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-datatypes-01" ;
	//   rdfs:comment "xsd:byte literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-datatypes-01.ttl> ;
	//   .

	{`@prefix xsd:     <http://www.w3.org/2001/XMLSchema#> .
<s> <p> "123"^^xsd:byte .`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "123", DataType: xsdByte},
		},
	}},

	//<#turtle-syntax-datatypes-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-datatypes-02" ;
	//   rdfs:comment "integer as xsd:string" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-datatypes-02.ttl> ;
	//   .

	{`@prefix rdf:     <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix xsd:     <http://www.w3.org/2001/XMLSchema#> .
<s> <p> "123"^^xsd:string .`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "123", DataType: xsdString},
		},
	}},

	//<#turtle-syntax-kw-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-kw-01" ;
	//   rdfs:comment "boolean literal (true)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-kw-01.ttl> ;
	//   .

	{`<s> <p> true .`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "true", DataType: xsdBoolean},
		},
	}},

	//<#turtle-syntax-kw-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-kw-02" ;
	//   rdfs:comment "boolean literal (false)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-kw-02.ttl> ;
	//   .

	{`<s> <p> false .`, "", []Triple{
		Triple{
			Subj: IRI{str: "s"},
			Pred: IRI{str: "p"},
			Obj:  Literal{str: "false", DataType: xsdBoolean},
		},
	}},

	//<#turtle-syntax-kw-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-kw-03" ;
	//   rdfs:comment "'a' as keyword" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-kw-03.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s a :C .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/C"},
		},
	}},

	//<#turtle-syntax-struct-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-struct-01" ;
	//   rdfs:comment "object list" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-struct-01.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p :o1 , :o2 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o1"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o2"},
		},
	}},

	//<#turtle-syntax-struct-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-struct-02" ;
	//   rdfs:comment "predicate list with object list" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-struct-02.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p1 :o1 ;
   :p2 :o2 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o1"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p2"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o2"},
		},
	}},

	//<#turtle-syntax-struct-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-struct-03" ;
	//   rdfs:comment "predicate list with object list and dangling ';'" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-struct-03.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p1 :o1 ;
   :p2 :o2 ;
   .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o1"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p2"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o2"},
		},
	}},

	//<#turtle-syntax-struct-04> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-struct-04" ;
	//   rdfs:comment "predicate list with multiple ;;" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-struct-04.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p1 :o1 ;;
   :p2 :o2
   .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o1"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p2"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o2"},
		},
	}},

	//<#turtle-syntax-struct-05> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-struct-05" ;
	//   rdfs:comment "predicate list with multiple ;;" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-struct-05.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p1 :o1 ;
   :p2 :o2 ;;
   .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o1"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p2"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o2"},
		},
	}},

	//<#turtle-syntax-lists-01> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-lists-01" ;
	//   rdfs:comment "empty list" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-lists-01.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p () .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#turtle-syntax-lists-02> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-lists-02" ;
	//   rdfs:comment "mixed list" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-lists-02.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p (1 "2" :o) .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Blank{id: "_:b1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "2", DataType: xsdString},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  Blank{id: "_:b3"},
		},
		Triple{
			Subj: Blank{id: "_:b3"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
		Triple{
			Subj: Blank{id: "_:b3"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#turtle-syntax-lists-03> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-lists-03" ;
	//   rdfs:comment "isomorphic list as subject and object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-lists-03.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
(1) :p (1) .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#turtle-syntax-lists-04> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-lists-04" ;
	//   rdfs:comment "lists of lists" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-lists-04.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
(()) :p (()) .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#turtle-syntax-lists-05> rdf:type rdft:TestTurtlePositiveSyntax ;
	//   mf:name    "turtle-syntax-lists-05" ;
	//   rdfs:comment "mixed lists with embedded lists" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-lists-05.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
(1 2 (1 2)) :p (( "a") "b" :o) .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "2", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  Blank{id: "_:b3"},
		},
		Triple{
			Subj: Blank{id: "_:b3"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Blank{id: "_:b4"},
		},
		Triple{
			Subj: Blank{id: "_:b4"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b4"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  Blank{id: "_:b5"},
		},
		Triple{
			Subj: Blank{id: "_:b5"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "2", DataType: xsdInteger},
		},
		Triple{
			Subj: Blank{id: "_:b5"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b3"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  Blank{id: "_:b6"},
		},
		Triple{
			Subj: Blank{id: "_:b6"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Blank{id: "_:b7"},
		},
		Triple{
			Subj: Blank{id: "_:b7"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "a", DataType: xsdString},
		},
		Triple{
			Subj: Blank{id: "_:b7"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
		Triple{
			Subj: Blank{id: "_:b6"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  Blank{id: "_:b8"},
		},
		Triple{
			Subj: Blank{id: "_:b8"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "b", DataType: xsdString},
		},
		Triple{
			Subj: Blank{id: "_:b8"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  Blank{id: "_:b9"},
		},
		Triple{
			Subj: Blank{id: "_:b9"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
		Triple{
			Subj: Blank{id: "_:b9"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#turtle-syntax-bad-uri-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-uri-01" ;
	//   rdfs:comment "Bad IRI : space (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-uri-01.ttl> ;
	//   .

	{`# Bad IRI : space.
<http://www.w3.org/2013/TurtleTests/ space> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`,
		"bad IRI: disallowed character ' '", []Triple{}},

	//<#turtle-syntax-bad-uri-02> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-uri-02" ;
	//   rdfs:comment "Bad IRI : bad escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-uri-02.ttl> ;
	//   .

	{`# Bad IRI : bad escape
<http://www.w3.org/2013/TurtleTests/\u00ZZ11> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`,
		"bad IRI: insufficent hex digits in unicode escape", []Triple{}},

	//<#turtle-syntax-bad-uri-03> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-uri-03" ;
	//   rdfs:comment "Bad IRI : bad long escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-uri-03.ttl> ;
	//   .

	{`# Bad IRI : bad escape
<http://www.w3.org/2013/TurtleTests/\U00ZZ1111> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`,
		"bad IRI: insufficent hex digits in unicode escape", []Triple{}},

	//<#turtle-syntax-bad-uri-04> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-uri-04" ;
	//   rdfs:comment "Bad IRI : character escapes not allowed (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-uri-04.ttl> ;
	//   .

	{`# Bad IRI : character escapes not allowed.
<http://www.w3.org/2013/TurtleTests/\n> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`,
		"bad IRI: disallowed escape character 'n'", []Triple{}},

	//<#turtle-syntax-bad-uri-05> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-uri-05" ;
	//   rdfs:comment "Bad IRI : character escapes not allowed (2) (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-uri-05.ttl> ;
	//   .

	{`# Bad IRI : character escapes not allowed.
<http://www.w3.org/2013/TurtleTests/\/> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`,
		`bad IRI: disallowed escape character '/'`, []Triple{}},

	//<#turtle-syntax-bad-prefix-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-prefix-01" ;
	//   rdfs:comment "No prefix (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-prefix-01.ttl> ;
	//   .

	{`# No prefix
:s <http://www.w3.org/2013/TurtleTests/p> "x" .`, "missing namespace for prefix: ':'", []Triple{}},

	//<#turtle-syntax-bad-prefix-02> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-prefix-02" ;
	//   rdfs:comment "No prefix (2) (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-prefix-02.ttl> ;
	//   .

	{`# No prefix
@prefix rdf:     <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
<http://www.w3.org/2013/TurtleTests/s> rdf:type :C .`, "missing namespace for prefix: ':'", []Triple{}},

	//<#turtle-syntax-bad-prefix-03> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-prefix-03" ;
	//   rdfs:comment "@prefix without IRI (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-prefix-03.ttl> ;
	//   .

	{`# @prefix without IRI.
@prefix ex: .`, "unexpected Dot as prefix IRI", []Triple{}},

	//<#turtle-syntax-bad-prefix-04> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-prefix-04" ;
	//   rdfs:comment "@prefix without prefix name (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-prefix-04.ttl> ;
	//   .

	{`# @prefix without prefix name .
@prefix <http://www.w3.org/2013/TurtleTests/> .`, "unexpected character: '<'", []Triple{}},

	//<#turtle-syntax-bad-prefix-05> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-prefix-05" ;
	//   rdfs:comment "@prefix without ':' (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-prefix-05.ttl> ;
	//   .

	{`# @prefix without :
@prefix x <http://www.w3.org/2013/TurtleTests/> .`, `illegal token: "x "`, []Triple{}},

	//<#turtle-syntax-bad-base-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-base-01" ;
	//   rdfs:comment "@base without IRI (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-base-01.ttl> ;
	//   .

	{`# @base without IRI.
@base .`, "unexpected Dot as base IRI", []Triple{}},

	//<#turtle-syntax-bad-base-02> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-base-02" ;
	//   rdfs:comment "@base in wrong case (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-base-02.ttl> ;
	//   .

	{`# @base in wrong case.
@BASE <http://www.w3.org/2013/TurtleTests/> .`, "unrecognized directive", []Triple{}},

	//<#turtle-syntax-bad-base-03> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-base-03" ;
	//   rdfs:comment "BASE without IRI (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-base-03.ttl> ;
	//   .

	{`# FULL STOP used after SPARQL BASE
BASE <http://www.w3.org/2013/TurtleTests/> .
<s> <p> <o> .`, "unexpected Dot as subject", []Triple{}},

	//<#turtle-syntax-bad-struct-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-01" ;
	//   rdfs:comment "Turtle is not TriG (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-01.ttl> ;
	//   .

	{`# Turtle is not TriG
{ <http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> }`,
		"unexpected character: '{'", []Triple{}},

	//<#turtle-syntax-bad-struct-02> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-02" ;
	//   rdfs:comment "Turtle is not N3 (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-02.ttl> ;
	//   .

	{`# Turtle is not N3
<http://www.w3.org/2013/TurtleTests/s> = <http://www.w3.org/2013/TurtleTests/o> .`,
		"unexpected character: '='", []Triple{}},

	//<#turtle-syntax-bad-struct-03> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-03" ;
	//   rdfs:comment "Turtle is not NQuads (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-03.ttl> ;
	//   .

	{`# Turtle is not NQuads
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> <http://www.w3.org/2013/TurtleTests/g> .`,
		"expected triple termination, got IRI (absolute)", []Triple{}},

	//<#turtle-syntax-bad-struct-04> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-04" ;
	//   rdfs:comment "Turtle does not allow literals-as-subjects (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-04.ttl> ;
	//   .

	{`# Turtle does not allow literals-as-subjects
"hello" <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`,
		"unexpected Literal as subject", []Triple{}},

	//<#turtle-syntax-bad-struct-05> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-05" ;
	//   rdfs:comment "Turtle does not allow literals-as-predicates (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-05.ttl> ;
	//   .

	{`# Turtle does not allow literals-as-predicates
<http://www.w3.org/2013/TurtleTests/s> "hello" <http://www.w3.org/2013/TurtleTests/o> .`,
		"unexpected Literal as predicate", []Triple{}},

	//<#turtle-syntax-bad-struct-06> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-06" ;
	//   rdfs:comment "Turtle does not allow bnodes-as-predicates (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-06.ttl> ;
	//   .

	{`# Turtle does not allow bnodes-as-predicates
<http://www.w3.org/2013/TurtleTests/s> [] <http://www.w3.org/2013/TurtleTests/o> .`,
		"unexpected Anonymous blank node as predicate", []Triple{}},

	//<#turtle-syntax-bad-struct-07> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-07" ;
	//   rdfs:comment "Turtle does not allow labeled bnodes-as-predicates (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-07.ttl> ;
	//   .

	{`# Turtle does not allow bnodes-as-predicates
<http://www.w3.org/2013/TurtleTests/s> _:p <http://www.w3.org/2013/TurtleTests/o> .`,
		"unexpected Blank node as predicate", []Triple{}},

	//<#turtle-syntax-bad-kw-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-kw-01" ;
	//   rdfs:comment "'A' is not a keyword (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-kw-01.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s A :C .`, `illegal token: "A "`, []Triple{}},

	//<#turtle-syntax-bad-kw-02> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-kw-02" ;
	//   rdfs:comment "'a' cannot be used as subject (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-kw-02.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
a :p :o .`, "unexpected rdf:type as subject", []Triple{}},

	//<#turtle-syntax-bad-kw-03> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-kw-03" ;
	//   rdfs:comment "'a' cannot be used as object (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-kw-03.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p a .`, "unexpected rdf:type as object", []Triple{}},

	//<#turtle-syntax-bad-kw-04> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-kw-04" ;
	//   rdfs:comment "'true' cannot be used as subject (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-kw-04.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
true :p :o .`, "unexpected Literal (boolean shorthand syntax) as subject", []Triple{}},

	//<#turtle-syntax-bad-kw-05> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-kw-05" ;
	//   rdfs:comment "'true' cannot be used as object (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-kw-05.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s true :o .`, "unexpected Literal (boolean shorthand syntax) as predicate", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-01" ;
	//   rdfs:comment "{} fomulae not in Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-01.ttl> ;
	//   .

	{`# {} fomulae not in Turtle
@prefix : <http://www.w3.org/2013/TurtleTests/> .

{ :a :q :c . } :p :z .
`, "unexpected character: '{'", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-02> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-02" ;
	//   rdfs:comment "= is not Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-02.ttl> ;
	//   .

	{`# = is not Turtle
@prefix : <http://www.w3.org/2013/TurtleTests/> .

:a = :b .`, "unexpected character: '='", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-03> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-03" ;
	//   rdfs:comment "N3 paths not in Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-03.ttl> ;
	//   .

	{`# N3 paths
@prefix : <http://www.w3.org/2013/TurtleTests/> .
@prefix ns: <http://www.w3.org/2013/TurtleTests/p#> .

:x.
  ns:p.
    ns:q :p :z .`, "unexpected Dot as predicate", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-04> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-04" ;
	//   rdfs:comment "N3 paths not in Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-04.ttl> ;
	//   .

	{`# N3 paths
@prefix : <http://www.w3.org/2013/TurtleTests/> .
@prefix ns: <http://www.w3.org/2013/TurtleTests/p#> .

:x^ns:p :p :z .`, "syntax error: unexpected character: '^'", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-05> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-05" ;
	//   rdfs:comment "N3 is...of not in Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-05.ttl> ;
	//   .

	{`# N3 is...of
@prefix : <http://www.w3.org/2013/TurtleTests/> .

:z is :p of :x .`, `illegal token: "is "`, []Triple{}},

	//<#turtle-syntax-bad-n3-extras-06> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-06" ;
	//   rdfs:comment "N3 paths not in Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-06.ttl> ;
	//   .

	{`# = is not Turtle
@prefix : <http://www.w3.org/2013/TurtleTests/> .

:a.:b.:c .`, "unexpected Dot as predicate", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-07> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-07" ;
	//   rdfs:comment "@keywords is not Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-07.ttl> ;
	//   .

	{`# @keywords is not Turtle
@keywords a .
x a Item .`, "unrecognized directive", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-08> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-08" ;
	//   rdfs:comment "@keywords is not Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-08.ttl> ;
	//   .

	{`# @keywords is not Turtle
@keywords a .
x a Item .`, "unrecognized directive", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-09> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-09" ;
	//   rdfs:comment "=> is not Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-09.ttl> ;
	//   .

	{`# => is not Turtle
@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s => :o .`, "unexpected character: '='", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-10> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-10" ;
	//   rdfs:comment "<= is not Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-10.ttl> ;
	//   .

	{`# <= is not Turtle
@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s <= :o .`, "bad IRI: disallowed character ' '", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-11> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-11" ;
	//   rdfs:comment "@forSome is not Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-11.ttl> ;
	//   .

	{`# @forSome is not Turtle
@prefix : <http://www.w3.org/2013/TurtleTests/> .
@forSome :x .`, "unrecognized directive", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-12> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-12" ;
	//   rdfs:comment "@forAll is not Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-12.ttl> ;
	//   .

	{`# @forAll is not Turtle
@prefix : <http://www.w3.org/2013/TurtleTests/> .
@forAll :x .`, "unrecognized directive", []Triple{}},

	//<#turtle-syntax-bad-n3-extras-13> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-n3-extras-13" ;
	//   rdfs:comment "@keywords is not Turtle (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-n3-extras-13.ttl> ;
	//   .

	{`# @keywords is not Turtle
@keywords .
x @a Item .`, "unrecognized directive", []Triple{}},

	//<#turtle-syntax-bad-struct-08> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-08" ;
	//   rdfs:comment "missing '.' (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-08.ttl> ;
	//   .

	{`# No DOT
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o>`,
		"expected triple termination, got EOF", []Triple{}},

	//<#turtle-syntax-bad-struct-09> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-09" ;
	//   rdfs:comment "extra '.' (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-09.ttl> ;
	//   .

	{`# Too many DOT
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> . .`,
		"unexpected Dot as subject", []Triple{}},

	//<#turtle-syntax-bad-struct-10> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-10" ;
	//   rdfs:comment "extra '.' (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-10.ttl> ;
	//   .

	{`# Too many DOT
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> . .
<http://www.w3.org/2013/TurtleTests/s1> <http://www.w3.org/2013/TurtleTests/p1> <http://www.w3.org/2013/TurtleTests/o1> .`,
		"unexpected Dot as subject", []Triple{}},

	//<#turtle-syntax-bad-struct-11> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-11" ;
	//   rdfs:comment "trailing ';' no '.' (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-11.ttl> ;
	//   .

	{`# Trailing ;
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> ;`,
		"expected triple termination, got Semicolon", []Triple{}},

	//<#turtle-syntax-bad-struct-12> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-12" ;
	//   rdfs:comment "subject, predicate, no object (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-12.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> `,
		"unexpected EOF as object", []Triple{}},

	//<#turtle-syntax-bad-struct-13> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-13" ;
	//   rdfs:comment "subject, predicate, no object (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-13.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> `,
		"unexpected EOF as object", []Triple{}},

	//<#turtle-syntax-bad-struct-14> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-14" ;
	//   rdfs:comment "literal as subject (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-14.ttl> ;
	//   .

	{`# Literal as subject
"abc" <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/p>  .`,
		"unexpected Literal as subject", []Triple{}},

	//<#turtle-syntax-bad-struct-15> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-15" ;
	//   rdfs:comment "literal as predicate (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-15.ttl> ;
	//   .

	{`# Literal as predicate
<http://www.w3.org/2013/TurtleTests/s> "abc" <http://www.w3.org/2013/TurtleTests/p>  .`,
		"unexpected Literal as predicate", []Triple{}},

	//<#turtle-syntax-bad-struct-16> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-16" ;
	//   rdfs:comment "bnode as predicate (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-16.ttl> ;
	//   .

	{`# BNode as predicate
<http://www.w3.org/2013/TurtleTests/s> [] <http://www.w3.org/2013/TurtleTests/p>  .`,
		"unexpected Anonymous blank node as predicate", []Triple{}},

	//<#turtle-syntax-bad-struct-17> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-struct-17" ;
	//   rdfs:comment "labeled bnode as predicate (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-struct-17.ttl> ;
	//   .

	{`# BNode as predicate
<http://www.w3.org/2013/TurtleTests/s> _:a <http://www.w3.org/2013/TurtleTests/p>  .`,
		"unexpected Blank node as predicate", []Triple{}},

	//<#turtle-syntax-bad-lang-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-lang-01" ;
	//   rdfs:comment "langString with bad lang (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-lang-01.ttl> ;
	//   .

	{`# Bad lang tag
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "string"@1 .`,
		"bad literal: invalid language tag", []Triple{}},

	//<#turtle-syntax-bad-esc-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-esc-01" ;
	//   rdfs:comment "Bad string escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-esc-01.ttl> ;
	//   .

	{`# Bad string escape
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "a\zb" .`,
		"bad literal: disallowed escape character 'z'", []Triple{}},

	//<#turtle-syntax-bad-esc-02> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-esc-02" ;
	//   rdfs:comment "Bad string escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-esc-02.ttl> ;
	//   .

	{`# Bad string escape
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "\uWXYZ" .`,
		"bad literal: insufficent hex digits in unicode escape", []Triple{}},

	//<#turtle-syntax-bad-esc-03> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-esc-03" ;
	//   rdfs:comment "Bad string escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-esc-03.ttl> ;
	//   .

	{`# Bad string escape
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "\U0000WXYZ" .`,
		"bad literal: insufficent hex digits in unicode escape", []Triple{}},

	//<#turtle-syntax-bad-esc-04> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-esc-04" ;
	//   rdfs:comment "Bad string escape (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-esc-04.ttl> ;
	//   .

	{`# Bad string escape
<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> "\U0000WXYZ" .`,
		"bad literal: insufficent hex digits in unicode escape", []Triple{}},

	//<#turtle-syntax-bad-pname-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-pname-01" ;
	//   rdfs:comment "'~' must be escaped in pname (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-pname-01.ttl> ;
	//   .

	{`# ~ must be escaped.
@prefix : <http://www.w3.org/2013/TurtleTests/> .
:a~b :p :o .`, "unexpected character: '~'", []Triple{}},

	//<#turtle-syntax-bad-pname-02> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-pname-02" ;
	//   rdfs:comment "Bad %-sequence in pname (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-pname-02.ttl> ;
	//   .

	{`# Bad %-sequence
@prefix : <http://www.w3.org/2013/TurtleTests/> .
:a%2 :p :o .`, "invalid hex escape sequence", []Triple{}},

	//<#turtle-syntax-bad-pname-03> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-pname-03" ;
	//   rdfs:comment "Bad unicode escape in pname (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-pname-03.ttl> ;
	//   .

	{`# No \u (x39 is "9")
@prefix : <http://www.w3.org/2013/TurtleTests/> .
:a\u0039 :p :o .`, "invalid escape charater 'u'", []Triple{}},

	//<#turtle-syntax-bad-string-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-string-01" ;
	//   rdfs:comment "mismatching string literal open/close (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-string-01.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p "abc' .`, "bad literal: no closing quote: '\"'", []Triple{}},

	//<#turtle-syntax-bad-string-02> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-string-02" ;
	//   rdfs:comment "mismatching string literal open/close (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-string-02.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p 'abc" .`, "bad literal: no closing quote: '\\''", []Triple{}},

	//<#turtle-syntax-bad-string-03> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-string-03" ;
	//   rdfs:comment "mismatching string literal long/short (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-string-03.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p '''abc' .`, "bad literal: no closing quote: '\\''", []Triple{}},

	//<#turtle-syntax-bad-string-04> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-string-04" ;
	//   rdfs:comment "mismatching long string literal open/close (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-string-04.ttl> ;
	//   .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p """abc''' .`, "bad literal: no closing quote: '\"'", []Triple{}},

	//<#turtle-syntax-bad-string-05> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-string-05" ;
	//   rdfs:comment "Long literal with missing end (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-string-05.ttl> ;
	//   .

	{`# Long literal with missing end
@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p """abc
def`, "bad literal: no closing quote: '\"'", []Triple{}},

	//<#turtle-syntax-bad-string-06> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-string-06" ;
	//   rdfs:comment "Long literal with extra quote (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-string-06.ttl> ;
	//   .

	{`# Long literal with 4"
@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p """abc""""@en .`, "bad literal: no closing quote: '\"'", []Triple{}},

	//<#turtle-syntax-bad-string-07> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-string-07" ;
	//   rdfs:comment "Long literal with extra squote (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-string-07.ttl> ;
	//   .

	{`# Long literal with 4'
@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p '''abc''''@en .`, "bad literal: no closing quote: '\\''", []Triple{}},

	//<#turtle-syntax-bad-num-01> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-num-01" ;
	//   rdfs:comment "Bad number format (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-num-01.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> 123.abc .`,
		"illegal token: \"abc \"", []Triple{}},

	//<#turtle-syntax-bad-num-02> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-num-02" ;
	//   rdfs:comment "Bad number format (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-num-02.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> 123e .`,
		"bad literal: illegal number syntax: missing exponent", []Triple{}},

	//<#turtle-syntax-bad-num-03> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-num-03" ;
	//   rdfs:comment "Bad number format (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-num-03.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> 123abc .`,
		" bad literal: illegal number syntax (number followed by 'a')", []Triple{}},

	//<#turtle-syntax-bad-num-04> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-num-04" ;
	//   rdfs:comment "Bad number format (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-num-04.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> 0x123 .`,
		" bad literal: illegal number syntax (number followed by 'x')", []Triple{}},

	//<#turtle-syntax-bad-num-05> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-num-05" ;
	//   rdfs:comment "Bad number format (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-num-05.ttl> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> +-1 .`,
		"bad literal: illegal number syntax: ('+' not followed by number)", []Triple{}},

	//<#turtle-eval-struct-01> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-eval-struct-01" ;
	//   rdfs:comment "triple with IRIs" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-eval-struct-01.ttl> ;
	//   mf:result    <turtle-eval-struct-01.nt> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-eval-struct-02> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-eval-struct-02" ;
	//   rdfs:comment "triple with IRIs and embedded whitespace" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-eval-struct-02.ttl> ;
	//   mf:result    <turtle-eval-struct-02.nt> ;
	//   .

	{`<http://www.w3.org/2013/TurtleTests/s>
      <http://www.w3.org/2013/TurtleTests/p1> <http://www.w3.org/2013/TurtleTests/o1> ;
      <http://www.w3.org/2013/TurtleTests/p2> <http://www.w3.org/2013/TurtleTests/o2> ;
      .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o1"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p2"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o2"},
		},
	}},

	//<#turtle-subm-01> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-01" ;
	//   rdfs:comment "Blank subject" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-01.ttl> ;
	//   mf:result    <turtle-subm-01.nt> ;
	//   .

	{`@prefix : <#> .
[] :x :y .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "#x"},
			Obj:  IRI{str: "#y"},
		},
	}},

	//<#turtle-subm-02> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-02" ;
	//   rdfs:comment "@prefix and qnames" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-02.ttl> ;
	//   mf:result    <turtle-subm-02.nt> ;
	//   .

	{`# Test @prefix and qnames
@prefix :  <http://example.org/base1#> .
@prefix a: <http://example.org/base2#> .
@prefix b: <http://example.org/base3#> .
:a :b :c .
a:a a:b a:c .
:a a:a b:a .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/base1#a"},
			Pred: IRI{str: "http://example.org/base1#b"},
			Obj:  IRI{str: "http://example.org/base1#c"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/base2#a"},
			Pred: IRI{str: "http://example.org/base2#b"},
			Obj:  IRI{str: "http://example.org/base2#c"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/base1#a"},
			Pred: IRI{str: "http://example.org/base2#a"},
			Obj:  IRI{str: "http://example.org/base3#a"},
		},
	}},

	//<#turtle-subm-03> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-03" ;
	//   rdfs:comment ", operator" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-03.ttl> ;
	//   mf:result    <turtle-subm-03.nt> ;
	//   .

	{`# Test , operator
@prefix : <http://example.org/base#> .
:a :b :c,
      :d,
      :e .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/base#a"},
			Pred: IRI{str: "http://example.org/base#b"},
			Obj:  IRI{str: "http://example.org/base#c"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/base#a"},
			Pred: IRI{str: "http://example.org/base#b"},
			Obj:  IRI{str: "http://example.org/base#d"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/base#a"},
			Pred: IRI{str: "http://example.org/base#b"},
			Obj:  IRI{str: "http://example.org/base#e"},
		},
	}},

	//<#turtle-subm-04> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-04" ;
	//   rdfs:comment "; operator" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-04.ttl> ;
	//   mf:result    <turtle-subm-04.nt> ;
	//   .

	{`# Test ; operator
@prefix : <http://example.org/base#> .
:a :b :c ;
   :d :e ;
   :f :g .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/base#a"},
			Pred: IRI{str: "http://example.org/base#b"},
			Obj:  IRI{str: "http://example.org/base#c"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/base#a"},
			Pred: IRI{str: "http://example.org/base#d"},
			Obj:  IRI{str: "http://example.org/base#e"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/base#a"},
			Pred: IRI{str: "http://example.org/base#f"},
			Obj:  IRI{str: "http://example.org/base#g"},
		},
	}},

	//<#turtle-subm-05> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-05" ;
	//   rdfs:comment "empty [] as subject and object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-05.ttl> ;
	//   mf:result    <turtle-subm-05.nt> ;
	//   .

	{`# Test empty [] operator; not allowed as predicate
@prefix : <http://example.org/base#> .
[] :a :b .
:c :d [] .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://example.org/base#a"},
			Obj:  IRI{str: "http://example.org/base#b"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/base#c"},
			Pred: IRI{str: "http://example.org/base#d"},
			Obj:  Blank{id: "_:b2"},
		},
	}},

	//<#turtle-subm-06> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-06" ;
	//   rdfs:comment "non-empty [] as subject and object" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-06.ttl> ;
	//   mf:result    <turtle-subm-06.nt> ;
	//   .

	{`# Test non empty [] operator; not allowed as predicate
@prefix : <http://example.org/base#> .
[ :a :b ] :c :d .
:e :f [ :g :h ] .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://example.org/base#a"},
			Obj:  IRI{str: "http://example.org/base#b"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://example.org/base#c"},
			Obj:  IRI{str: "http://example.org/base#d"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/base#e"},
			Pred: IRI{str: "http://example.org/base#f"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://example.org/base#g"},
			Obj:  IRI{str: "http://example.org/base#h"},
		},
	}},

	//<#turtle-subm-07> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-07" ;
	//   rdfs:comment "'a' as predicate" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-07.ttl> ;
	//   mf:result    <turtle-subm-07.nt> ;
	//   .

	{`# 'a' only allowed as a predicate
@prefix : <http://example.org/base#> .
:a a :b .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/base#a"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			Obj:  IRI{str: "http://example.org/base#b"},
		},
	}},

	//<#turtle-subm-08> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-08" ;
	//   rdfs:comment "simple collection" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-08.ttl> ;
	//   mf:result    <turtle-subm-08.nt> ;
	//   .

	{`@prefix : <http://example.org/stuff/1.0/> .
:a :b ( "apple" "banana" ) .
`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/stuff/1.0/a"},
			Pred: IRI{str: "http://example.org/stuff/1.0/b"},
			Obj:  Blank{id: "_:b1"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "apple", DataType: xsdString},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#first"},
			Obj:  Literal{str: "banana", DataType: xsdString},
		},
		Triple{
			Subj: Blank{id: "_:b2"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#turtle-subm-09> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-09" ;
	//   rdfs:comment "empty collection" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-09.ttl> ;
	//   mf:result    <turtle-subm-09.nt> ;
	//   .

	{`@prefix : <http://example.org/stuff/1.0/> .
:a :b ( ) .
`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/stuff/1.0/a"},
			Pred: IRI{str: "http://example.org/stuff/1.0/b"},
			Obj:  IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"},
		},
	}},

	//<#turtle-subm-10> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-10" ;
	//   rdfs:comment "integer datatyped literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-10.ttl> ;
	//   mf:result    <turtle-subm-10.nt> ;
	//   .

	{`# Test integer datatyped literals using an OWL cardinality constraint
@prefix owl: <http://www.w3.org/2002/07/owl#> .

# based on examples in the OWL Reference

_:hasParent a owl:ObjectProperty .

[] a owl:Restriction ;
  owl:onProperty _:hasParent ;
  owl:maxCardinality 2 .`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:hasParent"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			Obj:  IRI{str: "http://www.w3.org/2002/07/owl#ObjectProperty"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			Obj:  IRI{str: "http://www.w3.org/2002/07/owl#Restriction"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2002/07/owl#onProperty"},
			Obj:  Blank{id: "_:hasParent"},
		},
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://www.w3.org/2002/07/owl#maxCardinality"},
			Obj:  Literal{str: "2", DataType: xsdInteger},
		},
	}},

	//<#turtle-subm-11> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-11" ;
	//   rdfs:comment "decimal integer canonicalization" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-11.ttl> ;
	//   mf:result    <turtle-subm-11.nt> ;
	//   .

	{`<http://example.org/res1> <http://example.org/prop1> 000000 .
<http://example.org/res2> <http://example.org/prop2> 0 .
<http://example.org/res3> <http://example.org/prop3> 000001 .
<http://example.org/res4> <http://example.org/prop4> 2 .
<http://example.org/res5> <http://example.org/prop5> 4 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/res1"},
			Pred: IRI{str: "http://example.org/prop1"},
			Obj:  Literal{str: "000000", DataType: xsdInteger},
		},
		Triple{
			Subj: IRI{str: "http://example.org/res2"},
			Pred: IRI{str: "http://example.org/prop2"},
			Obj:  Literal{str: "0", DataType: xsdInteger},
		},
		Triple{
			Subj: IRI{str: "http://example.org/res3"},
			Pred: IRI{str: "http://example.org/prop3"},
			Obj:  Literal{str: "000001", DataType: xsdInteger},
		},
		Triple{
			Subj: IRI{str: "http://example.org/res4"},
			Pred: IRI{str: "http://example.org/prop4"},
			Obj:  Literal{str: "2", DataType: xsdInteger},
		},
		Triple{
			Subj: IRI{str: "http://example.org/res5"},
			Pred: IRI{str: "http://example.org/prop5"},
			Obj:  Literal{str: "4", DataType: xsdInteger},
		},
	}},

	//<#turtle-subm-12> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-12" ;
	//   rdfs:comment "- and _ in names and qnames" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-12.ttl> ;
	//   mf:result    <turtle-subm-12.nt> ;
	//   .

	{`# Tests for - and _ in names, qnames
@prefix ex1: <http://example.org/ex1#> .
@prefix ex-2: <http://example.org/ex2#> .
@prefix ex3_: <http://example.org/ex3#> .
@prefix ex4-: <http://example.org/ex4#> .

ex1:foo-bar ex1:foo_bar "a" .
ex-2:foo-bar ex-2:foo_bar "b" .
ex3_:foo-bar ex3_:foo_bar "c" .
ex4-:foo-bar ex4-:foo_bar "d" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/ex1#foo-bar"},
			Pred: IRI{str: "http://example.org/ex1#foo_bar"},
			Obj:  Literal{str: "a", DataType: xsdString},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ex2#foo-bar"},
			Pred: IRI{str: "http://example.org/ex2#foo_bar"},
			Obj:  Literal{str: "b", DataType: xsdString},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ex3#foo-bar"},
			Pred: IRI{str: "http://example.org/ex3#foo_bar"},
			Obj:  Literal{str: "c", DataType: xsdString},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ex4#foo-bar"},
			Pred: IRI{str: "http://example.org/ex4#foo_bar"},
			Obj:  Literal{str: "d", DataType: xsdString},
		},
	}},

	//<#turtle-subm-13> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-13" ;
	//   rdfs:comment "tests for rdf:_<numbers> and other qnames starting with _" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-13.ttl> ;
	//   mf:result    <turtle-subm-13.nt> ;
	//   .

	{`# Tests for rdf:_<numbers> and other qnames starting with _
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix ex:  <http://example.org/ex#> .
@prefix :    <http://example.org/myprop#> .

ex:foo rdf:_1 "1" .
ex:foo rdf:_2 "2" .
ex:foo :_abc "def" .
ex:foo :_345 "678" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/ex#foo"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#_1"},
			Obj:  Literal{str: "1", DataType: xsdString},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ex#foo"},
			Pred: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#_2"},
			Obj:  Literal{str: "2", DataType: xsdString},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ex#foo"},
			Pred: IRI{str: "http://example.org/myprop#_abc"},
			Obj:  Literal{str: "def", DataType: xsdString},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ex#foo"},
			Pred: IRI{str: "http://example.org/myprop#_345"},
			Obj:  Literal{str: "678", DataType: xsdString},
		},
	}},

	//<#turtle-subm-14> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-14" ;
	//   rdfs:comment "bare : allowed" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-14.ttl> ;
	//   mf:result    <turtle-subm-14.nt> ;
	//   .

	{`# Test for : allowed
@prefix :    <http://example.org/ron> .

[] : [] .

: : : .
`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:b1"},
			Pred: IRI{str: "http://example.org/ron"},
			Obj:  Blank{id: "_:b2"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ron"},
			Pred: IRI{str: "http://example.org/ron"},
			Obj:  IRI{str: "http://example.org/ron"},
		},
	}},

	//<#turtle-subm-15> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-15" ;
	//   rdfs:comment "simple long literal" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-15.ttl> ;
	//   mf:result    <turtle-subm-15.nt> ;
	//   .

	{`# Test long literal
@prefix :  <http://example.org/ex#> .
:a :b """a long
	literal
with
newlines""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/ex#a"},
			Pred: IRI{str: "http://example.org/ex#b"},
			Obj:  Literal{str: "a long\n\tliteral\nwith\nnewlines", DataType: xsdString},
		},
	}},

	//<#turtle-subm-16> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-16" ;
	//   rdfs:comment "long literals with escapes" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-16.ttl> ;
	//   mf:result    <turtle-subm-16.nt> ;
	//   .

	{`@prefix : <http://example.org/foo#> .

## \U00015678 is a not a legal codepoint
## :a :b """\nthis \ris a \U00015678long\t
## literal\uABCD
## """ .
##
## :d :e """\tThis \uABCDis\r \U00015678another\n
## one
## """ .

# \U00015678 is a not a legal codepoint
# \U00012451 in Cuneiform numeric ban 3
:a :b """\nthis \ris a \U00012451long\t
literal\uABCD
""" .

:d :e """\tThis \uABCDis\r \U00012451another\n
one
""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/foo#a"},
			Pred: IRI{str: "http://example.org/foo#b"},
			Obj:  Literal{str: "\nthis \ris a \U00012451long\t\nliteral\uABCD\n", DataType: xsdString},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo#d"},
			Pred: IRI{str: "http://example.org/foo#e"},
			Obj:  Literal{str: "\tThis \uABCDis\r \U00012451another\n\none\n", DataType: xsdString},
		},
	}},

	//<#turtle-subm-17> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-17" ;
	//   rdfs:comment "floating point number" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-17.ttl> ;
	//   mf:result    <turtle-subm-17.nt> ;
	//   .

	{`@prefix : <http://example.org/#> .

:a :b  1.0 .
`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/#a"},
			Pred: IRI{str: "http://example.org/#b"},
			Obj:  Literal{str: "1.0", DataType: xsdDecimal},
		},
	}},

	//<#turtle-subm-18> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-18" ;
	//   rdfs:comment "empty literals, normal and long variant" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-18.ttl> ;
	//   mf:result    <turtle-subm-18.nt> ;
	//   .

	{`@prefix : <http://example.org/#> .

:a :b "" .

:c :d """""" .
`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/#a"},
			Pred: IRI{str: "http://example.org/#b"},
			Obj:  Literal{str: "", DataType: xsdString},
		},
		Triple{
			Subj: IRI{str: "http://example.org/#c"},
			Pred: IRI{str: "http://example.org/#d"},
			Obj:  Literal{str: "", DataType: xsdString},
		},
	}},

	//<#turtle-subm-19> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-19" ;
	//   rdfs:comment "positive integer, decimal and doubles" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-19.ttl> ;
	//   mf:result    <turtle-subm-19.nt> ;
	//   .

	{`@prefix : <http://example.org#> .
:a :b 1.0 .
:c :d 1 .
:e :f 1.0e0 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org#a"},
			Pred: IRI{str: "http://example.org#b"},
			Obj:  Literal{str: "1.0", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org#c"},
			Pred: IRI{str: "http://example.org#d"},
			Obj:  Literal{str: "1", DataType: xsdInteger},
		},
		Triple{
			Subj: IRI{str: "http://example.org#e"},
			Pred: IRI{str: "http://example.org#f"},
			Obj:  Literal{str: "1.0e0", DataType: xsdDouble},
		},
	}},

	//<#turtle-subm-20> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-20" ;
	//   rdfs:comment "negative integer, decimal and doubles" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-20.ttl> ;
	//   mf:result    <turtle-subm-20.nt> ;
	//   .

	{`@prefix : <http://example.org#> .
:a :b -1.0 .
:c :d -1 .
:e :f -1.0e0 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org#a"},
			Pred: IRI{str: "http://example.org#b"},
			Obj:  Literal{str: "-1.0", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org#c"},
			Pred: IRI{str: "http://example.org#d"},
			Obj:  Literal{str: "-1", DataType: xsdInteger},
		},
		Triple{
			Subj: IRI{str: "http://example.org#e"},
			Pred: IRI{str: "http://example.org#f"},
			Obj:  Literal{str: "-1.0e0", DataType: xsdDouble},
		},
	}},

	//<#turtle-subm-21> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-21" ;
	//   rdfs:comment "long literal ending in double quote" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-21.ttl> ;
	//   mf:result    <turtle-subm-21.nt> ;
	//   .

	{`# Test long literal
@prefix :  <http://example.org/ex#> .
:a :b """John said: "Hello World!\"""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/ex#a"},
			Pred: IRI{str: "http://example.org/ex#b"},
			Obj:  Literal{str: `John said: "Hello World!"`, DataType: xsdString},
		},
	}},

	//<#turtle-subm-22> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-22" ;
	//   rdfs:comment "boolean literals" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-22.ttl> ;
	//   mf:result    <turtle-subm-22.nt> ;
	//   .

	{`@prefix : <http://example.org#> .
:a :b true .
:c :d false .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org#a"},
			Pred: IRI{str: "http://example.org#b"},
			Obj:  Literal{str: "true", DataType: xsdBoolean},
		},
		Triple{
			Subj: IRI{str: "http://example.org#c"},
			Pred: IRI{str: "http://example.org#d"},
			Obj:  Literal{str: "false", DataType: xsdBoolean},
		},
	}},

	//<#turtle-subm-23> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-23" ;
	//   rdfs:comment "comments" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-23.ttl> ;
	//   mf:result    <turtle-subm-23.nt> ;
	//   .

	{`# comment test
@prefix : <http://example.org/#> .
:a :b :c . # end of line comment
:d # ignore me
  :e # and me
      :f # and me
        .
:g :h #ignore me
     :i,  # and me
     :j . # and me

:k :l :m ; #ignore me
   :n :o ; # and me
   :p :q . # and me`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/#a"},
			Pred: IRI{str: "http://example.org/#b"},
			Obj:  IRI{str: "http://example.org/#c"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/#d"},
			Pred: IRI{str: "http://example.org/#e"},
			Obj:  IRI{str: "http://example.org/#f"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/#g"},
			Pred: IRI{str: "http://example.org/#h"},
			Obj:  IRI{str: "http://example.org/#i"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/#g"},
			Pred: IRI{str: "http://example.org/#h"},
			Obj:  IRI{str: "http://example.org/#j"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/#k"},
			Pred: IRI{str: "http://example.org/#l"},
			Obj:  IRI{str: "http://example.org/#m"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/#k"},
			Pred: IRI{str: "http://example.org/#n"},
			Obj:  IRI{str: "http://example.org/#o"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/#k"},
			Pred: IRI{str: "http://example.org/#p"},
			Obj:  IRI{str: "http://example.org/#q"},
		},
	}},

	//<#turtle-subm-24> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-24" ;
	//   rdfs:comment "no final mewline" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-24.ttl> ;
	//   mf:result    <turtle-subm-24.nt> ;
	//   .

	{`# comment line with no final newline test
@prefix : <http://example.org/#> .
:a :b :c .
#foo`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/#a"},
			Pred: IRI{str: "http://example.org/#b"},
			Obj:  IRI{str: "http://example.org/#c"},
		},
	}},

	//<#turtle-subm-25> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-25" ;
	//   rdfs:comment "repeating a @prefix changes pname definition" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-25.ttl> ;
	//   mf:result    <turtle-subm-25.nt> ;
	//   .

	{`@prefix foo: <http://example.org/foo#>  .
@prefix foo: <http://example.org/bar#>  .

foo:blah foo:blah foo:blah .
`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/bar#blah"},
			Pred: IRI{str: "http://example.org/bar#blah"},
			Obj:  IRI{str: "http://example.org/bar#blah"},
		},
	}},

	//<#turtle-subm-26> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-26" ;
	//   rdfs:comment "Variations on decimal canonicalization" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-26.ttl> ;
	//   mf:result    <turtle-subm-26.nt> ;
	//   .

	{`<http://example.org/foo> <http://example.org/bar> "2.345"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "1"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "1.0"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "1."^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "1.000000000"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.3"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.234000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.2340000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.23400000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.234000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.2340000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.23400000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.234000000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.2340000000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.23400000000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.234000000000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.2340000000000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.23400000000000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.234000000000000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.2340000000000000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "2.23400000000000000000005"^^<http://www.w3.org/2001/XMLSchema#decimal> .
<http://example.org/foo> <http://example.org/bar> "1.2345678901234567890123457890"^^<http://www.w3.org/2001/XMLSchema#decimal> .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.345", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "1", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "1.0", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "1.", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "1.000000000", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.3", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.234000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.2340000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.23400000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.234000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.2340000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.23400000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.234000000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.2340000000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.23400000000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.234000000000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.2340000000000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.23400000000000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.234000000000000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.2340000000000000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "2.23400000000000000000005", DataType: xsdDecimal},
		},
		Triple{
			Subj: IRI{str: "http://example.org/foo"},
			Pred: IRI{str: "http://example.org/bar"},
			Obj:  Literal{str: "1.2345678901234567890123457890", DataType: xsdDecimal},
		},
	}},

	//<#turtle-subm-27> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "turtle-subm-27" ;
	//   rdfs:comment "Repeating @base changes base for relative IRI lookup" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-subm-27.ttl> ;
	//   mf:result    <turtle-subm-27.nt> ;
	//   .

	{`# In-scope base IRI is <http://www.w3.org/2013/TurtleTests/turtle-subm-27.ttl> at this point
<a1> <b1> <c1> .
@base <http://example.org/ns/> .
# In-scope base IRI is http://example.org/ns/ at this point
<a2> <http://example.org/ns/b2> <c2> .
@base <foo/> .
# In-scope base IRI is http://example.org/ns/foo/ at this point
<a3> <b3> <c3> .
@prefix : <bar#> .
:a4 :b4 :c4 .
@prefix : <http://example.org/ns2#> .
:a5 :b5 :c5 .`, "", []Triple{
		Triple{
			Subj: IRI{str: "a1"},
			Pred: IRI{str: "b1"},
			Obj:  IRI{str: "c1"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ns/a2"},
			Pred: IRI{str: "http://example.org/ns/b2"},
			Obj:  IRI{str: "http://example.org/ns/c2"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ns/foo/a3"},
			Pred: IRI{str: "http://example.org/ns/foo/b3"},
			Obj:  IRI{str: "http://example.org/ns/foo/c3"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ns/foo/bar#a4"},
			Pred: IRI{str: "http://example.org/ns/foo/bar#b4"},
			Obj:  IRI{str: "http://example.org/ns/foo/bar#c4"},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ns2#a5"},
			Pred: IRI{str: "http://example.org/ns2#b5"},
			Obj:  IRI{str: "http://example.org/ns2#c5"},
		},
	}},

	//<#turtle-eval-bad-01> rdf:type rdft:TestTurtleNegativeEval ;
	//   mf:name    "turtle-eval-bad-01" ;
	//   rdfs:comment "Bad IRI : good escape, bad charcater (negative evaluation test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-eval-bad-01.ttl> ;
	//   .

	{`# Bad IRI : good escape, bad charcater
<http://www.w3.org/2013/TurtleTests/\u0020> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`,
		`bad IRI: disallowed character in unicode escape: "\\u0020"`, []Triple{}},

	//<#turtle-eval-bad-02> rdf:type rdft:TestTurtleNegativeEval ;
	//   mf:name    "turtle-eval-bad-02" ;
	//   rdfs:comment "Bad IRI : hex 3C is < (negative evaluation test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-eval-bad-02.ttl> ;
	//   .

	{`# Bad IRI : hex 3C is <
<http://www.w3.org/2013/TurtleTests/\u003C> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`,
		`bad IRI: disallowed character in unicode escape: "\\u003C"`, []Triple{}},

	//<#turtle-eval-bad-03> rdf:type rdft:TestTurtleNegativeEval ;
	//   mf:name    "turtle-eval-bad-03" ;
	//   rdfs:comment "Bad IRI : hex 3E is  (negative evaluation test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-eval-bad-03.ttl> ;
	//   .

	{`# Bad IRI : hex 3E is >
<http://www.w3.org/2013/TurtleTests/\u003E> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`,
		`bad IRI: disallowed character in unicode escape: "\\u003E"`, []Triple{}},

	//<#turtle-eval-bad-04> rdf:type rdft:TestTurtleNegativeEval ;
	//   mf:name    "turtle-eval-bad-04" ;
	//   rdfs:comment "Bad IRI : {abc} (negative evaluation test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-eval-bad-04.ttl> ;
	//   .

	{`# Bad IRI
<http://www.w3.org/2013/TurtleTests/{abc}> <http://www.w3.org/2013/TurtleTests/p> <http://www.w3.org/2013/TurtleTests/o> .`,
		"bad IRI: disallowed character '{'", []Triple{}},

	//# tests requested by Jeremy Carroll
	//# http://www.w3.org/2011/rdf-wg/wiki/Turtle_Candidate_Recommendation_Comments#c35
	//<#comment_following_localName> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "comment_following_localName" ;
	//   rdfs:comment "comment following localName" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <comment_following_localName.ttl> ;
	//   mf:result    <IRI_spo.nt> ;
	//   .

	{`@prefix p: <http://a.example/> .
<http://a.example/s> <http://a.example/p> p:o#comment
.`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o"},
		},
	}},

	//<#number_sign_following_localName> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "number_sign_following_localName" ;
	//   rdfs:comment "number sign following localName" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <number_sign_following_localName.ttl> ;
	//   mf:result    <number_sign_following_localName.nt> ;
	//   .

	{`@prefix p: <http://a.example/> .
<http://a.example/s> <http://a.example/p> p:o\#numbersign
.`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/o#numbersign"},
		},
	}},

	//<#comment_following_PNAME_NS> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "comment_following_PNAME_NS" ;
	//   rdfs:comment "comment following PNAME_NS" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <comment_following_PNAME_NS.ttl> ;
	//   mf:result    <comment_following_PNAME_NS.nt> ;
	//   .

	{`@prefix p: <http://a.example/> .
<http://a.example/s> <http://a.example/p> p:#comment
.`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/"},
		},
	}},

	//<#number_sign_following_PNAME_NS> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "number_sign_following_PNAME_NS" ;
	//   rdfs:comment "number sign following PNAME_NS" ;
	//   rdft:approval rdft:Proposed ;
	//   mf:action    <number_sign_following_PNAME_NS.ttl> ;
	//   mf:result    <number_sign_following_PNAME_NS.nt> ;
	//   .

	{`@prefix p: <http://a.example/>.
<http://a.example/s> <http://a.example/p> p:\#numbersign
.`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://a.example/s"},
			Pred: IRI{str: "http://a.example/p"},
			Obj:  IRI{str: "http://a.example/#numbersign"},
		},
	}},

	//# tests from Dave Beckett
	//# http://www.w3.org/2011/rdf-wg/wiki/Turtle_Candidate_Recommendation_Comments#c28
	//<#LITERAL_LONG2_with_REVERSE_SOLIDUS> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "LITERAL_LONG2_with_REVERSE_SOLIDUS" ;
	//   rdfs:comment "REVERSE SOLIDUS at end of LITERAL_LONG2" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <LITERAL_LONG2_with_REVERSE_SOLIDUS.ttl> ;
	//   mf:result    <LITERAL_LONG2_with_REVERSE_SOLIDUS.nt> ;
	//   .

	{`@prefix : <http://example.org/ns#> .

:s :p1 """test-\\""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/ns#s"},
			Pred: IRI{str: "http://example.org/ns#p1"},
			Obj:  Literal{str: "test-\\", DataType: xsdString},
		},
	}},

	//<#turtle-syntax-bad-LITERAL2_with_langtag_and_datatype> rdf:type rdft:TestTurtleNegativeSyntax ;
	//   mf:name    "turtle-syntax-bad-num-05" ;
	//   rdfs:comment "Bad number format (negative test)" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <turtle-syntax-bad-LITERAL2_with_langtag_and_datatype.ttl> ;
	//   .

	{`<http://example.org/resource> <http://example.org#pred> "value"@en^^<http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral> .`,
		"syntax error: unexpected character: '^'", []Triple{}},

	//<#two_LITERAL_LONG2s> rdf:type rdft:TestTurtleEval ;
	//   mf:name    "two_LITERAL_LONG2s" ;
	//   rdfs:comment "two LITERAL_LONG2s testing quote delimiter overrun" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <two_LITERAL_LONG2s.ttl> ;
	//   mf:result    <two_LITERAL_LONG2s.nt> ;
	//   .

	{`# Test long literal twice to ensure it does not over-quote
@prefix :  <http://example.org/ex#> .
:a :b """first long literal""" .
:c :d """second long literal""" .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/ex#a"},
			Pred: IRI{str: "http://example.org/ex#b"},
			Obj:  Literal{str: "first long literal", DataType: xsdString},
		},
		Triple{
			Subj: IRI{str: "http://example.org/ex#c"},
			Pred: IRI{str: "http://example.org/ex#d"},
			Obj:  Literal{str: "second long literal", DataType: xsdString},
		},
	}},

	//<#langtagged_LONG_with_subtag> rdf:type rdft:TestTurtleEval ;
	//   mf:name      "langtagged_LONG_with_subtag" ;
	//   rdfs:comment "langtagged LONG with subtag \"\"\"Cheers\"\"\"@en-UK" ;
	//   rdft:approval rdft:Approved ;
	//   mf:action    <langtagged_LONG_with_subtag.ttl> ;
	//   mf:result    <langtagged_LONG_with_subtag.nt> ;
	//   .

	{`# Test long literal with lang tag
@prefix :  <http://example.org/ex#> .
:a :b """Cheers"""@en-UK .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://example.org/ex#a"},
			Pred: IRI{str: "http://example.org/ex#b"},
			Obj:  Literal{str: "Cheers", lang: "en-UK", DataType: rdfLangString},
		},
	}},

	//# tests from David Robillard
	//# http://www.w3.org/2011/rdf-wg/wiki/Turtle_Candidate_Recommendation_Comments#c21
	//<#turtle-syntax-bad-blank-label-dot-end>
	//	rdf:type rdft:TestTurtleNegativeSyntax ;
	//	rdfs:comment "Blank node label must not end in dot" ;
	//	mf:name "turtle-syntax-bad-blank-label-dot-end" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-bad-blank-label-dot-end.ttl> .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
_:b1. :p :o .`, "unexpected Dot as predicate", []Triple{}},

	//<#turtle-syntax-bad-number-dot-in-anon>
	//	rdf:type rdft:TestTurtleNegativeSyntax ;
	//	rdfs:comment "Dot delimeter may not appear in anonymous nodes" ;
	//	mf:name "turtle-syntax-bad-number-dot-in-anon" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-bad-number-dot-in-anon.ttl> .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .

:s
	:p [
		:p1 27.
	] .`, "unexpected Property list end as object", []Triple{}},

	//<#turtle-syntax-bad-ln-dash-start>
	//	rdf:type rdft:TestTurtleNegativeSyntax ;
	//	rdfs:comment "Local name must not begin with dash" ;
	//	mf:name "turtle-syntax-bad-ln-dash-start" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-bad-ln-dash-start.ttl> .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p :-o .`, "unexpected character: '-'", []Triple{}},

	//<#turtle-syntax-bad-ln-escape>
	//	rdf:type rdft:TestTurtleNegativeSyntax ;
	//	rdfs:comment "Bad hex escape in local name" ;
	//	mf:name "turtle-syntax-bad-ln-escape" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-bad-ln-escape.ttl> .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p :o%2 .`, "invalid hex escape sequence", []Triple{}},

	//<#turtle-syntax-bad-ln-escape-start>
	//	rdf:type rdft:TestTurtleNegativeSyntax ;
	//	rdfs:comment "Bad hex escape at start of local name" ;
	//	mf:name "turtle-syntax-bad-ln-escape-start" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-bad-ln-escape-start.ttl> .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s :p :%2o .`, "invalid hex escape sequence", []Triple{}},

	//<#turtle-syntax-bad-ns-dot-end>
	//	rdf:type rdft:TestTurtleNegativeSyntax ;
	//	rdfs:comment "Prefix must not end in dot" ;
	//	mf:name "turtle-syntax-bad-ns-dot-end" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-bad-ns-dot-end.ttl> .

	{`@prefix eg. : <http://www.w3.org/2013/TurtleTests/> .
eg.:s eg.:p eg.:o .`, "illegal token: \"eg. \"", []Triple{}},

	//<#turtle-syntax-bad-ns-dot-start>
	//	rdf:type rdft:TestTurtleNegativeSyntax ;
	//	rdfs:comment "Prefix must not start with dot" ;
	//	mf:name "turtle-syntax-bad-ns-dot-start" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-bad-ns-dot-start.ttl> .

	{`@prefix .eg : <http://www.w3.org/2013/TurtleTests/> .
.eg:s .eg:p .eg:o .`, "unexpected character: '.'", []Triple{}},

	//<#turtle-syntax-bad-missing-ns-dot-end>
	//	rdf:type rdft:TestTurtleNegativeSyntax ;
	//	rdfs:comment "Prefix must not end in dot (error in triple, not prefix directive like turtle-syntax-bad-ns-dot-end)" ;
	//	mf:name "turtle-syntax-bad-missing-ns-dot-end" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-bad-missing-ns-dot-end.ttl> .

	{`valid:s valid:p invalid.:o .`,
		"missing namespace for prefix: 'valid'", []Triple{}},

	//<#turtle-syntax-bad-missing-ns-dot-start>
	//	rdf:type rdft:TestTurtleNegativeSyntax ;
	//	rdfs:comment "Prefix must not start with dot (error in triple, not prefix directive like turtle-syntax-bad-ns-dot-end)" ;
	//	mf:name "turtle-syntax-bad-missing-ns-dot-start" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-bad-missing-ns-dot-start.ttl> .

	{`.undefined:s .undefined:p .undefined:o .`,
		"unexpected Dot as subject", []Triple{}},

	//<#turtle-syntax-ln-dots>
	//	rdf:type rdft:TestTurtlePositiveSyntax ;
	//	rdfs:comment "Dots in pname local names" ;
	//	mf:name "turtle-syntax-ln-dots" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-ln-dots.ttl> .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s.1 :p.1 :o.1 .
:s..2 :p..2 :o..2.
:3.s :3.p :3.`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s.1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p.1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o.1"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s..2"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p..2"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o..2"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/3.s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/3.p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/3"},
		},
	}},

	//<#turtle-syntax-ln-colons>
	//	rdf:type rdft:TestTurtlePositiveSyntax ;
	//	rdfs:comment "Colons in pname local names" ;
	//	mf:name "turtle-syntax-ln-colons" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-ln-colons.ttl> .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
:s:1 :p:1 :o:1 .
:s::2 :p::2 :o::2 .
:3:s :3:p :3 .
::s ::p ::o .
::s: ::p: ::o: .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s:1"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p:1"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o:1"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s::2"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p::2"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o::2"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/3:s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/3:p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/3"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/:s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/:p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/:o"},
		},
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/:s:"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/:p:"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/:o:"},
		},
	}},

	//<#turtle-syntax-ns-dots>
	//	rdf:type rdft:TestTurtlePositiveSyntax ;
	//	rdfs:comment "Dots in namespace names" ;
	//	mf:name "turtle-syntax-ns-dots" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-ns-dots.ttl> .

	{`@prefix e.g: <http://www.w3.org/2013/TurtleTests/> .
e.g:s e.g:p e.g:o .`, "", []Triple{
		Triple{
			Subj: IRI{str: "http://www.w3.org/2013/TurtleTests/s"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},

	//<#turtle-syntax-blank-label>
	//	rdf:type rdft:TestTurtlePositiveSyntax ;
	//	rdfs:comment "Characters allowed in blank node labels" ;
	//	mf:name "turtle-syntax-blank-label" ;
	//        rdft:approval rdft:Approved ;
	//	mf:action <turtle-syntax-blank-label.ttl> .

	{`@prefix : <http://www.w3.org/2013/TurtleTests/> .
_:0b :p :o . # Starts with digit
_:_b :p :o . # Starts with underscore
_:b.0 :p :o . # Contains dot, ends with digit`, "", []Triple{
		Triple{
			Subj: Blank{id: "_:0b"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
		Triple{
			Subj: Blank{id: "_:_b"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
		Triple{
			Subj: Blank{id: "_:b.0"},
			Pred: IRI{str: "http://www.w3.org/2013/TurtleTests/p"},
			Obj:  IRI{str: "http://www.w3.org/2013/TurtleTests/o"},
		},
	}},
}
