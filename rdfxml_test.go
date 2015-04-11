package rdf

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func BenchmarkDecodeRDFXML(b *testing.B) {
	input := `<?xml version="1.0" encoding="utf-8"?>
<rdf:RDF xml:base="http://www.gutenberg.org/"
  xmlns:cc="http://web.resource.org/cc/"
  xmlns:pgterms="http://www.gutenberg.org/2009/pgterms/"
  xmlns:dcam="http://purl.org/dc/dcam/"
  xmlns:dcterms="http://purl.org/dc/terms/"
  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
>
  <pgterms:ebook rdf:about="ebooks/1401">
    <dcterms:license rdf:resource="license"/>
    <dcterms:publisher>Project Gutenberg</dcterms:publisher>
    <dcterms:hasFormat>
      <pgterms:file rdf:about="http://www.gutenberg.org/ebooks/1401.epub.noimages">
        <dcterms:isFormatOf rdf:resource="ebooks/1401"/>
        <dcterms:extent rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">245821</dcterms:extent>
        <dcterms:format>
          <rdf:Description rdf:nodeID="Nb46390c65b94400cb3c45afb7036fb76">
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">application/epub+zip</rdf:value>
          </rdf:Description>
        </dcterms:format>
        <dcterms:modified rdf:datatype="http://www.w3.org/2001/XMLSchema#dateTime">2012-07-30T09:25:34.452918</dcterms:modified>
      </pgterms:file>
    </dcterms:hasFormat>
    <dcterms:hasFormat>
      <pgterms:file rdf:about="http://www.gutenberg.org/files/1401/1401-h/1401-h.htm">
        <dcterms:isFormatOf rdf:resource="ebooks/1401"/>
        <dcterms:format>
          <rdf:Description rdf:nodeID="Nf1374551f7d7427eb919455f56f9ebf3">
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">text/html; charset=iso-8859-1</rdf:value>
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
          </rdf:Description>
        </dcterms:format>
        <dcterms:modified rdf:datatype="http://www.w3.org/2001/XMLSchema#dateTime">2012-07-29T09:31:34</dcterms:modified>
        <dcterms:extent rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">661872</dcterms:extent>
      </pgterms:file>
    </dcterms:hasFormat>
    <pgterms:downloads rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">183</pgterms:downloads>
    <dcterms:subject>
      <rdf:Description rdf:nodeID="Nb52def5a42284aeb89f3b9353b37f44b">
        <dcam:memberOf rdf:resource="http://purl.org/dc/terms/LCSH"/>
        <rdf:value>Tarzan (Fictitious character) -- Fiction</rdf:value>
      </rdf:Description>
    </dcterms:subject>
    <dcterms:type>
      <rdf:Description rdf:nodeID="Nb9fcaa2b8fdf431fb0bc5a2c369e4680">
        <rdf:value>Text</rdf:value>
        <dcam:memberOf rdf:resource="http://purl.org/dc/terms/DCMIType"/>
      </rdf:Description>
    </dcterms:type>
    <dcterms:subject>
      <rdf:Description rdf:nodeID="Nb307a550edb3499ba2eb148af1f11280">
        <rdf:value>Fantasy fiction</rdf:value>
        <dcam:memberOf rdf:resource="http://purl.org/dc/terms/LCSH"/>
      </rdf:Description>
    </dcterms:subject>
    <dcterms:hasFormat>
      <pgterms:file rdf:about="http://www.gutenberg.org/ebooks/1401.plucker">
        <dcterms:isFormatOf rdf:resource="ebooks/1401"/>
        <dcterms:format>
          <rdf:Description rdf:nodeID="N001c92dd22ee4517a16e0cf46137df26">
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">application/prs.plucker</rdf:value>
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
          </rdf:Description>
        </dcterms:format>
        <dcterms:modified rdf:datatype="http://www.w3.org/2001/XMLSchema#dateTime">2012-07-30T09:25:42.164430</dcterms:modified>
        <dcterms:extent rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">361242</dcterms:extent>
      </pgterms:file>
    </dcterms:hasFormat>
    <dcterms:hasFormat>
      <pgterms:file rdf:about="http://www.gutenberg.org/files/1401/1401.txt">
        <dcterms:modified rdf:datatype="http://www.w3.org/2001/XMLSchema#dateTime">2012-07-29T09:31:40</dcterms:modified>
        <dcterms:isFormatOf rdf:resource="ebooks/1401"/>
        <dcterms:extent rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">636082</dcterms:extent>
        <dcterms:format>
          <rdf:Description rdf:nodeID="N278115d13cca4a2ca64950e9c20fdf48">
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">text/plain; charset=us-ascii</rdf:value>
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
          </rdf:Description>
        </dcterms:format>
      </pgterms:file>
    </dcterms:hasFormat>
    <dcterms:hasFormat>
      <pgterms:file rdf:about="http://www.gutenberg.org/ebooks/1401.qioo">
        <dcterms:isFormatOf rdf:resource="ebooks/1401"/>
        <dcterms:modified rdf:datatype="http://www.w3.org/2001/XMLSchema#dateTime">2012-07-30T09:25:33.573357</dcterms:modified>
        <dcterms:extent rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">294903</dcterms:extent>
        <dcterms:format>
          <rdf:Description rdf:nodeID="Nb6f86e1bda1c4c6b9d04b1b736240fa7">
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">application/x-qioo-ebook</rdf:value>
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
          </rdf:Description>
        </dcterms:format>
      </pgterms:file>
    </dcterms:hasFormat>
    <dcterms:subject>
      <rdf:Description rdf:nodeID="N62d27de7b67d4bb2a7deaaf573728d88">
        <rdf:value>Adventure stories</rdf:value>
        <dcam:memberOf rdf:resource="http://purl.org/dc/terms/LCSH"/>
      </rdf:Description>
    </dcterms:subject>
    <dcterms:title>Tarzan the Untamed</dcterms:title>
    <dcterms:hasFormat>
      <pgterms:file rdf:about="http://www.gutenberg.org/files/1401/1401.zip">
        <dcterms:format>
          <rdf:Description rdf:nodeID="N376fc5f5c83442fdbb282837c5a6f548">
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">application/zip</rdf:value>
          </rdf:Description>
        </dcterms:format>
        <dcterms:isFormatOf rdf:resource="ebooks/1401"/>
        <dcterms:extent rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">232967</dcterms:extent>
        <dcterms:format>
          <rdf:Description rdf:nodeID="N56bb3cef5fce45aabf0dfd6ea08e6bd2">
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">text/plain; charset=us-ascii</rdf:value>
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
          </rdf:Description>
        </dcterms:format>
        <dcterms:modified rdf:datatype="http://www.w3.org/2001/XMLSchema#dateTime">2012-07-29T09:32:04</dcterms:modified>
      </pgterms:file>
    </dcterms:hasFormat>
    <dcterms:creator>
      <pgterms:agent rdf:about="2009/agents/48">
        <pgterms:name>Burroughs, Edgar Rice</pgterms:name>
        <pgterms:birthdate rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">1875</pgterms:birthdate>
        <pgterms:deathdate rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">1950</pgterms:deathdate>
        <pgterms:webpage rdf:resource="http://en.wikipedia.org/wiki/Edgar_Rice_Burroughs"/>
      </pgterms:agent>
    </dcterms:creator>
    <dcterms:subject>
      <rdf:Description rdf:nodeID="N50cc1255b7ef484cbe22f3d63a00fdaf">
        <dcam:memberOf rdf:resource="http://purl.org/dc/terms/LCC"/>
        <rdf:value>PS</rdf:value>
      </rdf:Description>
    </dcterms:subject>
    <dcterms:language>
      <rdf:Description rdf:nodeID="N891b552e92d24ac5b201c7354e5cc5c8">
        <rdf:value rdf:datatype="http://purl.org/dc/terms/RFC4646">en</rdf:value>
      </rdf:Description>
    </dcterms:language>
    <dcterms:rights>Public domain in the USA.</dcterms:rights>
    <dcterms:hasFormat>
      <pgterms:file rdf:about="http://www.gutenberg.org/ebooks/1401.txt.utf-8">
        <dcterms:isFormatOf rdf:resource="ebooks/1401"/>
        <dcterms:extent rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">636054</dcterms:extent>
        <dcterms:format>
          <rdf:Description rdf:nodeID="Nbd9fc4e881e24448b5595cf9069a3d1e">
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">text/plain</rdf:value>
          </rdf:Description>
        </dcterms:format>
        <dcterms:modified rdf:datatype="http://www.w3.org/2001/XMLSchema#dateTime">2012-07-30T09:25:32.973525</dcterms:modified>
      </pgterms:file>
    </dcterms:hasFormat>
    <dcterms:hasFormat>
      <pgterms:file rdf:about="http://www.gutenberg.org/ebooks/1401.kindle.noimages">
        <dcterms:modified rdf:datatype="http://www.w3.org/2001/XMLSchema#dateTime">2012-07-30T09:25:38.741381</dcterms:modified>
        <dcterms:extent rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">1014124</dcterms:extent>
        <dcterms:isFormatOf rdf:resource="ebooks/1401"/>
        <dcterms:format>
          <rdf:Description rdf:nodeID="N44904e6d853f43b9926ddc60463d45ad">
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">application/x-mobipocket-ebook</rdf:value>
          </rdf:Description>
        </dcterms:format>
      </pgterms:file>
    </dcterms:hasFormat>
    <dcterms:hasFormat>
      <pgterms:file rdf:about="http://www.gutenberg.org/files/1401/1401-h.zip">
        <dcterms:modified rdf:datatype="http://www.w3.org/2001/XMLSchema#dateTime">2012-07-29T09:32:04</dcterms:modified>
        <dcterms:format>
          <rdf:Description rdf:nodeID="N3339d80d716845ebb8a1ed3c977f22da">
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">text/html; charset=iso-8859-1</rdf:value>
          </rdf:Description>
        </dcterms:format>
        <dcterms:extent rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">235966</dcterms:extent>
        <dcterms:isFormatOf rdf:resource="ebooks/1401"/>
        <dcterms:format>
          <rdf:Description rdf:nodeID="Ncc2c14b838514a1ca95b06778018e522">
            <dcam:memberOf rdf:resource="http://purl.org/dc/terms/IMT"/>
            <rdf:value rdf:datatype="http://purl.org/dc/terms/IMT">application/zip</rdf:value>
          </rdf:Description>
        </dcterms:format>
      </pgterms:file>
    </dcterms:hasFormat>
    <dcterms:issued rdf:datatype="http://www.w3.org/2001/XMLSchema#date">1998-07-01</dcterms:issued>
  </pgterms:ebook>
  <cc:Work rdf:about="">
    <cc:license rdf:resource="http://www.gnu.org/licenses/gpl.html"/>
  </cc:Work>
  <rdf:Description rdf:about="http://en.wikipedia.org/wiki/Edgar_Rice_Burroughs">
    <dcterms:description>Wikipedia</dcterms:description>
  </rdf:Description>
</rdf:RDF>`
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dec := NewTripleDecoder(bytes.NewBufferString(input), RDFXML)
		for _, err := dec.Decode(); err != io.EOF; _, err = dec.Decode() {
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	b.SetBytes(int64(len(input)))
}

func TestRDFXMLExamples(t *testing.T) {
	for i, test := range rdfxmlExamples {
		dec := NewTripleDecoder(bytes.NewBufferString(test.rdfxml), RDFXML)
		dec.SetOption(Base, IRI{str: "http://www.w3.org/2013/RDFXMLTests/" + test.file})
		ts, err := dec.DecodeAll()
		if err != nil {
			t.Fatalf("[%d] parseRDFXML(%s).Serialize(NTriples) => %v, want %q", i, test.rdfxml, err, test.nt)
			continue
		}

		var b bytes.Buffer
		enc := NewTripleEncoder(&b, NTriples)
		err = enc.EncodeAll(ts)
		enc.Close()
		if err != nil {
			t.Fatalf("[%d] parseRDFXML(%s).Serialize(NTriples) => %v, want %q", i, test.rdfxml, err, test.nt)
		}
		if b.String() != test.nt {
			t.Fatalf("[%d] parseRDFXML(%s).Serialize(NTriples) => %v, want %v", i, test.rdfxml, b.String(), test.nt)
		}
	}
}

func TestRDFXML(t *testing.T) {
	for i, test := range rdfxmlTestSuite {
		dec := NewTripleDecoder(bytes.NewBufferString(test.rdfxml), RDFXML)
		dec.SetOption(Base, IRI{str: "http://www.w3.org/2013/RDFXMLTests/" + test.file})
		ts, err := dec.DecodeAll()
		if test.err == "TODO" {
			continue
		}
		if test.err != "" && err == nil {
			t.Fatalf("[%d] parseRDFXML(%s).Serialize(NTriples) => <no error>, want %q", i, test.rdfxml, test.err)
			continue
		}

		if test.err != "" && err != nil {
			if !strings.HasSuffix(err.Error(), test.err) {
				t.Fatalf("[%d] parseRDFXML(%s).Serialize(NTriples) => %s, want %q", i, test.rdfxml, err, test.err)
			}
			continue
		}

		if test.err == "" && err != nil {
			t.Fatalf("[%d] parseRDFXML(%s).Serialize(NTriples) => %v, want %q", i, test.rdfxml, err, test.nt)
			continue
		}

		var b bytes.Buffer
		enc := NewTripleEncoder(&b, NTriples)
		err = enc.EncodeAll(ts)
		enc.Close()
		if err != nil {
			t.Fatalf("[%d] parseRDFXML(%s).Serialize(NTriples) => %v, want %q", i, test.rdfxml, err, test.nt)
		}
		if b.String() != test.nt {
			t.Fatalf("[%d] parseRDFXML(%s).Serialize(NTriples) => %v, want %v", i, test.rdfxml, b.String(), test.nt)
		}
	}
}

var rdfxmlExamples = []struct {
	file   string
	rdfxml string
	nt     string
}{
	{
		// [0]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example07.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:dc="http://purl.org/dc/elements/1.1/"
         xmlns:ex="http://example.org/stuff/1.0/">
  <rdf:Description rdf:about="http://www.w3.org/TR/rdf-syntax-grammar"
		   dc:title="RDF/XML Syntax Specification (Revised)">
    <ex:editor>
      <rdf:Description ex:fullName="Dave Beckett">
	<ex:homePage rdf:resource="http://purl.org/net/dajobe/" />
      </rdf:Description>
    </ex:editor>
  </rdf:Description>
</rdf:RDF>
`,
		`<http://www.w3.org/TR/rdf-syntax-grammar> <http://purl.org/dc/elements/1.1/title> "RDF/XML Syntax Specification (Revised)" .
<http://www.w3.org/TR/rdf-syntax-grammar> <http://example.org/stuff/1.0/editor> _:b0 .
_:b0 <http://example.org/stuff/1.0/fullName> "Dave Beckett" .
_:b0 <http://example.org/stuff/1.0/homePage> <http://purl.org/net/dajobe/> .
`,
	},
	{
		// [1]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example08.rdf",
		`<?xml version="1.0" encoding="utf-8"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:dc="http://purl.org/dc/elements/1.1/">
  <rdf:Description rdf:about="http://www.w3.org/TR/rdf-syntax-grammar">
    <dc:title>RDF/XML Syntax Specification (Revised)</dc:title>
    <dc:title xml:lang="en">RDF/XML Syntax Specification (Revised)</dc:title>
    <dc:title xml:lang="en-US">RDF/XML Syntax Specification (Revised)</dc:title>
  </rdf:Description>

  <rdf:Description rdf:about="http://example.org/buecher/baum" xml:lang="de">
    <dc:title>Der Baum</dc:title>
    <dc:description>Das Buch ist außergewöhnlich</dc:description>
    <dc:title xml:lang="en">The Tree</dc:title>
  </rdf:Description>
</rdf:RDF>
`,
		`<http://www.w3.org/TR/rdf-syntax-grammar> <http://purl.org/dc/elements/1.1/title> "RDF/XML Syntax Specification (Revised)" .
<http://www.w3.org/TR/rdf-syntax-grammar> <http://purl.org/dc/elements/1.1/title> "RDF/XML Syntax Specification (Revised)"@en .
<http://www.w3.org/TR/rdf-syntax-grammar> <http://purl.org/dc/elements/1.1/title> "RDF/XML Syntax Specification (Revised)"@en-US .
<http://example.org/buecher/baum> <http://purl.org/dc/elements/1.1/title> "Der Baum"@de .
<http://example.org/buecher/baum> <http://purl.org/dc/elements/1.1/description> "Das Buch ist außergewöhnlich"@de .
<http://example.org/buecher/baum> <http://purl.org/dc/elements/1.1/title> "The Tree"@en .
`,
	},
	{
		// [2]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example09.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:ex="http://example.org/stuff/1.0/">
  <rdf:Description rdf:about="http://example.org/item01">
    <ex:prop rdf:parseType="Literal"
             xmlns:a="http://example.org/a#"><a:Box required="true">
         <a:widget size="10" />
         <a:grommit id="23" /></a:Box>
    </ex:prop>
  </rdf:Description>
</rdf:RDF>
`,
		`<http://example.org/item01> <http://example.org/stuff/1.0/prop> "<a:Box xmlns:a=\"http://example.org/a#\" required=\"true\">\n         <a:widget size=\"10\"></a:widget>\n         <a:grommit id=\"23\"></a:grommit></a:Box>\n    "^^<http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral> .
`,
	},
	{
		// [3]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example10.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:ex="http://example.org/stuff/1.0/">
  <rdf:Description rdf:about="http://example.org/item01">
    <ex:size rdf:datatype="http://www.w3.org/2001/XMLSchema#int">123</ex:size>
  </rdf:Description>
</rdf:RDF>
`,
		`<http://example.org/item01> <http://example.org/stuff/1.0/size> "123"^^<http://www.w3.org/2001/XMLSchema#int> .
`,
	},
	{
		// [4]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example11.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:dc="http://purl.org/dc/elements/1.1/"
         xmlns:ex="http://example.org/stuff/1.0/">
  <rdf:Description rdf:about="http://www.w3.org/TR/rdf-syntax-grammar"
		   dc:title="RDF/XML Syntax Specification (Revised)">
    <ex:editor rdf:nodeID="abc"/>
  </rdf:Description>

  <rdf:Description rdf:nodeID="abc"
                   ex:fullName="Dave Beckett">
    <ex:homePage rdf:resource="http://purl.org/net/dajobe/"/>
  </rdf:Description>
</rdf:RDF>
`,
		`<http://www.w3.org/TR/rdf-syntax-grammar> <http://purl.org/dc/elements/1.1/title> "RDF/XML Syntax Specification (Revised)" .
<http://www.w3.org/TR/rdf-syntax-grammar> <http://example.org/stuff/1.0/editor> _:abc .
_:abc <http://example.org/stuff/1.0/fullName> "Dave Beckett" .
_:abc <http://example.org/stuff/1.0/homePage> <http://purl.org/net/dajobe/> .
`,
	},
	{
		// [5]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example12.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:dc="http://purl.org/dc/elements/1.1/"
         xmlns:ex="http://example.org/stuff/1.0/">
  <rdf:Description rdf:about="http://www.w3.org/TR/rdf-syntax-grammar"
		   dc:title="RDF/XML Syntax Specification (Revised)">
    <ex:editor rdf:parseType="Resource">
      <ex:fullName>Dave Beckett</ex:fullName>
      <ex:homePage rdf:resource="http://purl.org/net/dajobe/"/>
    </ex:editor>
  </rdf:Description>
</rdf:RDF>
`,
		`<http://www.w3.org/TR/rdf-syntax-grammar> <http://purl.org/dc/elements/1.1/title> "RDF/XML Syntax Specification (Revised)" .
<http://www.w3.org/TR/rdf-syntax-grammar> <http://example.org/stuff/1.0/editor> _:b0 .
_:b0 <http://example.org/stuff/1.0/fullName> "Dave Beckett" .
_:b0 <http://example.org/stuff/1.0/homePage> <http://purl.org/net/dajobe/> .
`,
	},
	{
		// [6]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example13.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:dc="http://purl.org/dc/elements/1.1/"
         xmlns:ex="http://example.org/stuff/1.0/">
  <rdf:Description rdf:about="http://www.w3.org/TR/rdf-syntax-grammar"
		   dc:title="RDF/XML Syntax Specification (Revised)">
    <ex:editor ex:fullName="Dave Beckett" />
    <!-- Note the ex:homePage property has been ignored for this example -->
  </rdf:Description>
</rdf:RDF>
`,
		`<http://www.w3.org/TR/rdf-syntax-grammar> <http://purl.org/dc/elements/1.1/title> "RDF/XML Syntax Specification (Revised)" .
<http://www.w3.org/TR/rdf-syntax-grammar> <http://example.org/stuff/1.0/editor> _:b0 .
_:b0 <http://example.org/stuff/1.0/fullName> "Dave Beckett" .
`,
	},
	{
		// [7]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example14.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:dc="http://purl.org/dc/elements/1.1/"
         xmlns:ex="http://example.org/stuff/1.0/">
  <rdf:Description rdf:about="http://example.org/thing">
    <rdf:type rdf:resource="http://example.org/stuff/1.0/Document"/>
    <dc:title>A marvelous thing</dc:title>
  </rdf:Description>
</rdf:RDF>
`,
		`<http://example.org/thing> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/stuff/1.0/Document> .
<http://example.org/thing> <http://purl.org/dc/elements/1.1/title> "A marvelous thing" .
`,
	},
	{
		// [8]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example15.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:dc="http://purl.org/dc/elements/1.1/"
         xmlns:ex="http://example.org/stuff/1.0/">
  <ex:Document rdf:about="http://example.org/thing">
    <dc:title>A marvelous thing</dc:title>
  </ex:Document>
</rdf:RDF>
`,
		`<http://example.org/thing> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/stuff/1.0/Document> .
<http://example.org/thing> <http://purl.org/dc/elements/1.1/title> "A marvelous thing" .
`,
	},
	{
		// [9]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example16.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:ex="http://example.org/stuff/1.0/"
         xml:base="http://example.org/here/">
  <rdf:Description rdf:ID="snack">
    <ex:prop rdf:resource="fruit/apple"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/here/#snack> <http://example.org/stuff/1.0/prop> <http://example.org/here/fruit/apple> .
`,
	},
	{
		// [10]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example17.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Seq rdf:about="http://example.org/favourite-fruit">
    <rdf:_1 rdf:resource="http://example.org/banana"/>
    <rdf:_2 rdf:resource="http://example.org/apple"/>
    <rdf:_3 rdf:resource="http://example.org/pear"/>
  </rdf:Seq>
</rdf:RDF>`,
		`<http://example.org/favourite-fruit> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq> .
<http://example.org/favourite-fruit> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> <http://example.org/banana> .
<http://example.org/favourite-fruit> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_2> <http://example.org/apple> .
<http://example.org/favourite-fruit> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_3> <http://example.org/pear> .
`,
	},
	{
		// [11]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example18.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Seq rdf:about="http://example.org/favourite-fruit">
    <rdf:li rdf:resource="http://example.org/banana"/>
    <rdf:li rdf:resource="http://example.org/apple"/>
    <rdf:li rdf:resource="http://example.org/pear"/>
  </rdf:Seq>
</rdf:RDF>`,
		`<http://example.org/favourite-fruit> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq> .
<http://example.org/favourite-fruit> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> <http://example.org/banana> .
<http://example.org/favourite-fruit> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_2> <http://example.org/apple> .
<http://example.org/favourite-fruit> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_3> <http://example.org/pear> .
`,
	},
	{
		// [12]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example19.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:ex="http://example.org/stuff/1.0/">
  <rdf:Description rdf:about="http://example.org/basket">
    <ex:hasFruit rdf:parseType="Collection">
      <rdf:Description rdf:about="http://example.org/banana"/>
      <rdf:Description rdf:about="http://example.org/apple"/>
      <rdf:Description rdf:about="http://example.org/pear"/>
    </ex:hasFruit>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/basket> <http://example.org/stuff/1.0/hasFruit> _:b0 .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#first> <http://example.org/banana> .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#rest> _:b1 .
_:b1 <http://www.w3.org/1999/02/22-rdf-syntax-ns#first> <http://example.org/apple> .
_:b1 <http://www.w3.org/1999/02/22-rdf-syntax-ns#rest> _:b2 .
_:b2 <http://www.w3.org/1999/02/22-rdf-syntax-ns#first> <http://example.org/pear> .
_:b2 <http://www.w3.org/1999/02/22-rdf-syntax-ns#rest> <http://www.w3.org/1999/02/22-rdf-syntax-ns#nil> .
`,
	},
	{
		// [13]
		"http://www.w3.org/TR/2004/REC-rdf-syntax-grammar-20040210/example20.rdf",
		`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:ex="http://example.org/stuff/1.0/"
         xml:base="http://example.org/triples/">
  <rdf:Description rdf:about="http://example.org/">
    <ex:prop rdf:ID="triple1">blah</ex:prop>
  </rdf:Description>
</rdf:RDF>
`,
		`<http://example.org/> <http://example.org/stuff/1.0/prop> "blah" .
<http://example.org/triples/#triple1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://example.org/triples/#triple1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> <http://example.org/> .
<http://example.org/triples/#triple1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://example.org/stuff/1.0/prop> .
<http://example.org/triples/#triple1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> "blah" .
`,
	},
}

var rdfxmlTestSuite = []struct {
	file   string
	rdfxml string
	nt     string
	err    string
}{

	{
		// [0] #amp-in-url-test001
		//
		// Description: the purpose of this test case is to show how one
		// of XML's Predefined Entities - in this case the ampersand - is
		// represented when it is used in the value of an rdf:about
		// attribute. The ampersand is represented by its numeric
		// character reference as specified in:
		// http://www.w3.org/TR/REC-xml#sec-predefined-ent In the
		// associated N-Triples file, the ampersand will be represented
		// with a single ampersand character (and not the ampersand's
		// numeric character reference). Note: when a XML/HTML browser is
		// used to display this file, a single ampersand character may be
		// displayed and not the ampersand's numeric character reference.
		// In this case, the browser may provide an alternate way to view
		// the file (such as viewing the file's source or saving to a
		// file).
		//
		"amp-in-url/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

  <rdf:Description rdf:about="http://example/q?abc=1&#38;def=2">
    <rdf:value>xxx</rdf:value>
  </rdf:Description>

</rdf:RDF>`,
		`<http://example/q?abc=1&def=2> <http://www.w3.org/1999/02/22-rdf-syntax-ns#value> "xxx" .
`,
		"",
	},
	{
		// [1] #datatypes-test001
		//
		// A simple datatype production; a language+datatype production.
		//
		"datatypes/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Description rdf:about="http://example.org/foo">
   <eg:bar rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">10</eg:bar>
   <eg:baz rdf:datatype="http://www.w3.org/2001/XMLSchema#integer" xml:lang="fr">10</eg:baz>
 </rdf:Description>

</rdf:RDF>`,
		`<http://example.org/foo> <http://example.org/bar> "10"^^<http://www.w3.org/2001/XMLSchema#integer> .
<http://example.org/foo> <http://example.org/baz> "10"^^<http://www.w3.org/2001/XMLSchema#integer> .
`,
		"",
	},
	{
		// [2] #datatypes-test002
		//
		// A parser is not required to know about well-formed datatyped
		// literals.
		//
		"datatypes/test002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Description rdf:about="http://example.org/foo">
   <eg:bar rdf:datatype="http://www.w3.org/2001/XMLSchema#integer">flargh</eg:bar>
 </rdf:Description>

</rdf:RDF>`,
		`<http://example.org/foo> <http://example.org/bar> "flargh"^^<http://www.w3.org/2001/XMLSchema#integer> .
`,
		"",
	},
	{
		// [3] #rdf-charmod-literals-test001
		//
		// Does the treatment of literals conform to charmod ? Test for
		// success of legal Normal Form C literal
		//
		"rdf-charmod-literals/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">
   <!-- Dürst registers himself as a creator of the Charmod WD. -->

   <rdf:Description rdf:about="http://www.w3.org/TR/2002/WD-charmod-20020220">

   <!-- The ü below is a single character #xFC in NFC
        (encoded as two UTF-8 octets #xC3 #xBC)  -->
      <eg:Creator eg:named="Dürst"/>

   </rdf:Description>
</rdf:RDF>`,
		`<http://www.w3.org/TR/2002/WD-charmod-20020220> <http://example.org/Creator> _:b0 .
_:b0 <http://example.org/named> "Dürst" .
`,
		"",
	},
	{
		// [4] #rdf-charmod-uris-test001
		//
		// A uriref is allowed to match non-US ASCII forms conforming to
		// Unicode Normal Form C. No escaping algorithm is applied.
		//
		"rdf-charmod-uris/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/#">

  <!-- The é below is a single Unicode character #xE9 in
       Unicode Normal Form C, NFC (here encoded as
       two UTF-8 octets #C3,#A9) -->

   <rdf:Description rdf:about="http://example.org/#André">
      <eg:owes>2000</eg:owes>
   </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/#André> <http://example.org/#owes> "2000" .
`,
		"",
	},
	{
		// [5] #rdf-charmod-uris-test002
		//
		// A uriref which already has % escaping is permitted. No
		// unescaping algorithm is applied.
		//
		"rdf-charmod-uris/test002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/#">
 
  <!-- The %C3%A9 below corresponds to é under the standard
        %-escaping algorithm for URIs. -->

   <rdf:Description rdf:about="http://example.org/#Andr%C3%A9">
      <eg:owes>2000</eg:owes>
   </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/#Andr%C3%A9> <http://example.org/#owes> "2000" .
`,
		"",
	},
	{
		// [6] #rdf-containers-syntax-vs-schema-error001
		//
		// rdf:li is not allowed as as an attribute
		//
		"rdf-containers-syntax-vs-schema/error001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:foo="http://foo/">

  <foo:bar rdf:li="1"/>
</rdf:RDF>`,
		"",

		"unexpected as attribute: rdf:li",
	},
	{
		// [7] #rdf-containers-syntax-vs-schema-error002
		//
		// rdf:li elements as typed nodes - a bizarre case As specified
		// in
		// http://lists.w3.org/Archives/Public/w3c-rdfcore-wg/2001Nov/0651.html
		// is not an error.
		//
		"rdf-containers-syntax-vs-schema/error002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:foo="http://foo/">
  <rdf:li/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:li",
	},
	{
		// [8] #rdf-containers-syntax-vs-schema-test001
		//
		// Simple container
		//
		"rdf-containers-syntax-vs-schema/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

  <rdf:Bag> 
    <rdf:li>1</rdf:li>
    <rdf:li>2</rdf:li>
  </rdf:Bag>
</rdf:RDF>`,
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag> .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "1" .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_2> "2" .
`,
		"",
	},
	{
		// [9] #rdf-containers-syntax-vs-schema-test002
		//
		// rdf:li is unaffected by other rdf:_nnn properties. This test
		// case is concerned only with defining the triples that this
		// particular example RDF/XML represents. It is not concerned
		// with whether that collection of triples violates any other
		// constraints, e.g. restrictions on the number of rdf:_1
		// properties that may be defined for a resource.
		//
		"rdf-containers-syntax-vs-schema/test002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:foo="http://foo/">

  <foo:Bar>
    <rdf:_1>_1</rdf:_1>
    <rdf:li>1</rdf:li>
    <rdf:_3>_3</rdf:_3>
    <rdf:li>2</rdf:li>
  </foo:Bar>
</rdf:RDF>`,
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://foo/Bar> .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "_1" .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "1" .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_3> "_3" .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_2> "2" .
`,
		"",
	},
	{
		// [10] #rdf-containers-syntax-vs-schema-test003
		//
		// rdf:li elements can exist in any description element
		//
		"rdf-containers-syntax-vs-schema/test003.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:foo="http://foo/">

  <foo:Bar>
    <rdf:li>1</rdf:li>
    <rdf:li>2</rdf:li>
  </foo:Bar>
</rdf:RDF>`,
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://foo/Bar> .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "1" .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_2> "2" .
`,
		"",
	},
	{
		// [11] #rdf-containers-syntax-vs-schema-test004
		//
		// rdf:li elements match any of the property element productions
		//
		"rdf-containers-syntax-vs-schema/test004.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:foo="http://foo/">

  <foo:Bar>
    <rdf:li rdf:ID="e1">1</rdf:li>
    <rdf:li rdf:parseType="Literal">2</rdf:li>
    <rdf:li rdf:parseType="Resource">
      <rdf:type rdf:resource="http://foo/Bar"/>
    </rdf:li>
    <rdf:li rdf:ID="e4" foo:bar="foobar"/>
  </foo:Bar>
</rdf:RDF>`,
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://foo/Bar> .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "1" .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test004.rdf#e1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test004.rdf#e1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> _:b0 .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test004.rdf#e1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test004.rdf#e1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> "1" .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_2> "2"^^<http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral> .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_3> _:b1 .
_:b1 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://foo/Bar> .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_4> _:b2 .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test004.rdf#e4> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test004.rdf#e4> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> _:b0 .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test004.rdf#e4> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_4> .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test004.rdf#e4> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> _:b2 .
_:b2 <http://foo/bar> "foobar" .
`,
		"",
	},
	{
		// [12] #rdf-containers-syntax-vs-schema-test006
		//
		// containers match the typed node production
		//
		"rdf-containers-syntax-vs-schema/test006.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:foo="http://foo/">

  <rdf:Seq rdf:ID="e1" rdf:_3="3" rdf:value="foobar"/>
  <rdf:Alt rdf:about="#e2" rdf:_2="2" rdf:value="foobar">
    <rdf:value>barfoo</rdf:value>
  </rdf:Alt>
  <rdf:Bag />
</rdf:RDF>`,
		`<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test006.rdf#e1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq> .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test006.rdf#e1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_3> "3" .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test006.rdf#e1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#value> "foobar" .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test006.rdf#e2> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Alt> .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test006.rdf#e2> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_2> "2" .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test006.rdf#e2> <http://www.w3.org/1999/02/22-rdf-syntax-ns#value> "foobar" .
<http://www.w3.org/2013/RDFXMLTests/rdf-containers-syntax-vs-schema/test006.rdf#e2> <http://www.w3.org/1999/02/22-rdf-syntax-ns#value> "barfoo" .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag> .
`,
		"",
	},
	{
		// [13] #rdf-containers-syntax-vs-schema-test007
		//
		// rdf:li processing within each element is independent
		//
		"rdf-containers-syntax-vs-schema/test007.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:foo="http://foo/">

  <rdf:Description>
    <rdf:li>
      <rdf:Description>
        <rdf:li>1</rdf:li>
        <rdf:li>2</rdf:li>
      </rdf:Description>
    </rdf:li>
    <rdf:li>2</rdf:li>
  </rdf:Description>
</rdf:RDF>`,
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> _:b1 .
_:b1 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "1" .
_:b1 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_2> "2" .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_2> "2" .
`,
		"",
	},
	{
		// [14] #rdf-containers-syntax-vs-schema-test008
		//
		// rdf:li processing is per element, not per resource.
		//
		"rdf-containers-syntax-vs-schema/test008.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

  <rdf:Description rdf:about="http://desc"> 
    <rdf:li>1</rdf:li>
  </rdf:Description>

  <rdf:Description rdf:about="http://desc"> 
    <rdf:li>1-again</rdf:li>
  </rdf:Description>
</rdf:RDF>`,
		`<http://desc> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "1" .
<http://desc> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "1-again" .
`,
		"",
	},
	{
		// [15] #rdf-element-not-mandatory-test001
		//
		// A surrounding rdf:RDF element is no longer mandatory.
		//
		"rdf-element-not-mandatory/test001.rdf",
		`<Book xmlns="http://example.org/terms#">
  <title>Dogs in Hats</title>
</Book>`,
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/terms#Book> .
_:b0 <http://example.org/terms#title> "Dogs in Hats" .
`,
		"",
	},
	{
		// [16] #rdf-ns-prefix-confusion-test0001
		//
		// RDF attributes that are required to have an rdf: prefix about
		// aboutEach ID bagID type resource parseType
		//
		"rdf-ns-prefix-confusion/test0001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  List of RDF attributes that are required to have an rdf: prefix
    about aboutEach 
    ID bagID type resource parseType 

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Test rdf:about attribute - expect 1 triple -->

  <!-- 6.3 description, part 2; 6.7 aboutAttr -->
  <rdf:Description rdf:about="http://example.org/resource1/">
    <eg:property>bar</eg:property>
  </rdf:Description>
   
</rdf:RDF>`,
		`<http://example.org/resource1/> <http://example.org/property> "bar" .
`,
		"",
	},
	{
		// [17] #rdf-ns-prefix-confusion-test0003
		//
		// RDF attributes that are required to have an rdf: prefix about
		// aboutEach ID bagID type resource parseType
		//
		"rdf-ns-prefix-confusion/test0003.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">
 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  List of RDF attributes that are required to have an rdf: prefix
    about aboutEach 
    ID bagID type resource parseType 

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Test rdf:resource - expect 1 triple -->

  <!-- 6.3 description, part 2 -->
  <rdf:Description rdf:about="http://example.org/resource1/">
    <!-- 6.12 propertyElt part 4; 6.16 idRefAttr; 6.18 resourceAttr -->
    <eg:property rdf:resource="http://example.org/resource2/"/>
   
 </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/resource1/> <http://example.org/property> <http://example.org/resource2/> .
`,
		"",
	},
	{
		// [18] #rdf-ns-prefix-confusion-test0004
		//
		// RDF attributes that are required to have an rdf: prefix about
		// aboutEach ID bagID type resource parseType
		//
		"rdf-ns-prefix-confusion/test0004.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">
 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  List of RDF attributes that are required to have an rdf: prefix
    about aboutEach 
    ID bagID type resource parseType 

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Test rdf:ID - expect 1 triple  -->

  <!-- 6.3 description, part 2; 6.5 idAboutAttr; 6.6 idAttr -->
  <rdf:Description rdf:ID="foo">
    <eg:property>bar</eg:property>
  </rdf:Description>
  
</rdf:RDF>`,
		`<http://www.w3.org/2013/RDFXMLTests/rdf-ns-prefix-confusion/test0004.rdf#foo> <http://example.org/property> "bar" .
`,
		"",
	},
	{
		// [19] #rdf-ns-prefix-confusion-test0005
		//
		// RDF attributes that are required to have an rdf: prefix about
		// aboutEach ID bagID type resource parseType
		//
		"rdf-ns-prefix-confusion/test0005.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">
 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  List of RDF attributes that are required to have an rdf: prefix
    about aboutEach 
    ID bagID type resource parseType 

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Test rdf:parseType - expect 2 triples -->

  <!-- 6.3 description, part 2; 6.5 idAboutAttr; 6.7 aboutAbout -->
  <rdf:Description rdf:about="http://example.org/resource1/">

    <!-- 6.12 propertyElt, part 3; 6.33 parseResource -->
    <eg:property rdf:parseType="Resource">

       <!-- 6.12 propertyElt, part 1 -->
       <eg:property2>bar</eg:property2>
    </eg:property>
  </rdf:Description>
  
</rdf:RDF>`,
		`<http://example.org/resource1/> <http://example.org/property> _:b0 .
_:b0 <http://example.org/property2> "bar" .
`,
		"",
	},
	{
		// [20] #rdf-ns-prefix-confusion-test0006
		//
		// RDF attributes that are required to have an rdf: prefix about
		// aboutEach ID bagID type resource parseType
		//
		"rdf-ns-prefix-confusion/test0006.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  List of RDF attributes that are required to have an rdf: prefix
    about aboutEach 
    ID bagID type resource parseType 

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Test rdf:type attribute - expect 1 triple -->

  <!-- 6.3 description, part 1; 6.10 propAttr, part 1; 6.11 typeAttr -->
  <rdf:Description rdf:about="http://example.org/resource/"
                   rdf:type="http://example.org/class/"/>
  
</rdf:RDF>`,
		`<http://example.org/resource/> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/class/> .
`,
		"",
	},
	{
		// [21] #rdf-ns-prefix-confusion-test0009
		//
		// Namespace qualification MUST be used for all property
		// attributes.
		//
		"rdf-ns-prefix-confusion/test0009.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  Namespace qualification MUST be used for all property attributes.

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Test namespace-qualified property attribute - expect 1 triple -->

  <!-- 6.3 description, part 1; 6.10 propAttr; 6.14 propName; 6.19 Qname -->

  <rdf:Description rdf:about="http://example.org/resource/" eg:property="bar" />

</rdf:RDF>`,
		`<http://example.org/resource/> <http://example.org/property> "bar" .
`,
		"",
	},
	{
		// [22] #rdf-ns-prefix-confusion-test0010
		//
		// Non-prefixed RDF elements (NOT attributes) are allowed when a
		// default XML element namespace is defined with an
		// xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#" attribute.
		//
		"rdf-ns-prefix-confusion/test0010.rdf",
		`<RDF xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:eg="http://example.org/">

 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  Non-prefixed RDF elements (NOT attributes) are allowed when a
  default XML element namespace is defined with an
  xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#" attribute.

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Testing outer bare RDF element (using default namespace) -->

  <!-- Testing bare Description element (using default namespace) 
       - expect 1 triple -->

  <!-- 6.3 description, part 1; 6.10 propAttr; 6.14 propName; 6.19 Qname -->

  <Description rdf:about="http://example.org/resource/" eg:property="bar" />

</RDF>`,
		`<http://example.org/resource/> <http://example.org/property> "bar" .
`,
		"",
	},
	{
		// [23] #rdf-ns-prefix-confusion-test0011
		//
		// Non-prefixed RDF elements (NOT attributes) are allowed when a
		// default XML element namespace is defined with an
		// xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#" attribute.
		//
		"rdf-ns-prefix-confusion/test0011.rdf",
		`<RDF xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:eg="http://example.org/">

 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  Non-prefixed RDF elements (NOT attributes) are allowed when a
  default XML element namespace is defined with an
  xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#" attribute.

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Testing outer bare RDF element (using default namespace) -->

  <!-- Testing bare Seq element (using default namespace)
       - expect 2 triples  -->

  <!-- 6.2 obj; 6.4 container; 6.25 sequence, part 1; idAttr; --> 
  <Seq rdf:ID="container">
    <!-- 6.28 member; 6.29 inlineItem, part 1 -->
    <rdf:li>bar</rdf:li>
  </Seq>

</RDF>`,
		`<http://www.w3.org/2013/RDFXMLTests/rdf-ns-prefix-confusion/test0011.rdf#container> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq> .
<http://www.w3.org/2013/RDFXMLTests/rdf-ns-prefix-confusion/test0011.rdf#container> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "bar" .
`,
		"",
	},
	{
		// [24] #rdf-ns-prefix-confusion-test0012
		//
		// Non-prefixed RDF elements (NOT attributes) are allowed when a
		// default XML element namespace is defined with an
		// xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#" attribute.
		//
		"rdf-ns-prefix-confusion/test0012.rdf",
		`<RDF xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:eg="http://example.org/">

 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  Non-prefixed RDF elements (NOT attributes) are allowed when a
  default XML element namespace is defined with an
  xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#" attribute.

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Testing outer bare RDF element (using default namespace) -->

  <!-- Testing bare Bag element (using default namespace)
       - expect 2 triples  -->

  <!-- 6.2 obj; 6.4 container; 6.26 bag, part 1; idAttr; --> 
  <Bag rdf:ID="container">
    <!-- 6.28 member; 6.29 inlineItem, part 1 -->
    <rdf:li>bar</rdf:li>
  </Bag>

</RDF>`,
		`<http://www.w3.org/2013/RDFXMLTests/rdf-ns-prefix-confusion/test0012.rdf#container> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag> .
<http://www.w3.org/2013/RDFXMLTests/rdf-ns-prefix-confusion/test0012.rdf#container> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "bar" .
`,
		"",
	},
	{
		// [25] #rdf-ns-prefix-confusion-test0013
		//
		// Non-prefixed RDF elements (NOT attributes) are allowed when a
		// default XML element namespace is defined with an
		// xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#" attribute.
		//
		"rdf-ns-prefix-confusion/test0013.rdf",
		`<RDF xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:eg="http://example.org/">

 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  Non-prefixed RDF elements (NOT attributes) are allowed when a
  default XML element namespace is defined with an
  xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#" attribute.

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Testing outer bare RDF element (using default namespace) -->

  <!-- Testing bare Alt element (using default namespace)
       - expect 2 triples  -->

  <!-- 6.2 obj; 6.4 container; 6.27 alternative, part 1; idAttr; --> 
  <Alt rdf:ID="container">
    <!-- 6.28 member; 6.29 inlineItem, part 1 -->
    <rdf:li>bar</rdf:li>
  </Alt>

</RDF>`,
		`<http://www.w3.org/2013/RDFXMLTests/rdf-ns-prefix-confusion/test0013.rdf#container> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Alt> .
<http://www.w3.org/2013/RDFXMLTests/rdf-ns-prefix-confusion/test0013.rdf#container> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "bar" .
`,
		"",
	},
	{
		// [26] #rdf-ns-prefix-confusion-test0014
		//
		// Non-prefixed RDF elements (NOT attributes) are allowed when a
		// default XML element namespace is defined with an
		// xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#" attribute.
		//
		"rdf-ns-prefix-confusion/test0014.rdf",
		`<RDF xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
     xmlns:eg="http://example.org/">

 <!-- 
  Test case for
  Issue http://www.w3.org/2000/03/rdf-tracking/#rdf-ns-prefix-confusion

  Non-prefixed RDF elements (NOT attributes) are allowed when a
  default XML element namespace is defined with an
  xmlns="http://www.w3.org/1999/02/22-rdf-syntax-ns#" attribute.

  Dave Beckett - http://purl.org/net/dajobe/

 -->

  <!-- Testing outer bare RDF element (using default namespace) -->

  <!-- Testing bare Seq element (using default namespace) -->

  <!-- Testing bare li element (using default namespace) 
       - expect 2 triples -->

  <!-- 6.2 obj; 6.4 container; 6.25 sequence, part 1; idAttr; --> 
  <Seq rdf:ID="container">
    <!-- 6.28 member; 6.29 inlineItem, part 1 -->
    <li>bar</li>
  </Seq>

</RDF>`,
		`<http://www.w3.org/2013/RDFXMLTests/rdf-ns-prefix-confusion/test0014.rdf#container> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq> .
<http://www.w3.org/2013/RDFXMLTests/rdf-ns-prefix-confusion/test0014.rdf#container> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "bar" .
`,
		"",
	},
	{
		// [27] #rdfms-abouteach-error001
		//
		// aboutEach removed from the RDF specifications. See URI above
		// for further details.
		//
		"rdfms-abouteach/error001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

  <rdf:Bag rdf:ID="node">
    <rdf:li rdf:resource="http://example.org/node2"/>
  </rdf:Bag>

  <rdf:Description rdf:aboutEach="#node">
    <dc:rights xmlns:dc="http://purl.org/dc/elements/1.1/">me</dc:rights>
  </rdf:Description>

</rdf:RDF>`,
		"",

		"deprecated: rdf:aboutEach",
	},
	{
		// [28] #rdfms-abouteach-error002
		//
		// aboutEach removed from the RDF specifications. See URI above
		// for further details.
		//
		"rdfms-abouteach/error002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

  <rdf:Description rdf:about="http://example.org/node">
    <eg:property>foo</eg:property>
  </rdf:Description>

  <rdf:Description rdf:aboutEachPrefix="http://example.org/">
    <dc:creator xmlns:dc="http://purl.org/dc/elements/1.1/">me</dc:creator>
  </rdf:Description>

</rdf:RDF>`,
		"",

		"deprecated: rdf:aboutEachPrefix",
	},
	{
		// [29] #rdfms-difference-between-ID-and-about-error1
		//
		// two elements cannot use the same ID
		//
		"rdfms-difference-between-ID-and-about/error1.rdf",
		`<!--
	Base URI: http://www.w3.org/2013/RDFXMLTests/rdfms-difference-between-ID-and-about/error1.rdf

	This is illegal RDF: two elements cannot use the same ID.
	-->
	<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
	<rdf:Description rdf:ID="foo">
	  <rdf:value>abc</rdf:value>
	</rdf:Description>
	<rdf:Description rdf:ID="foo">
	  <rdf:value>abc</rdf:value>
	</rdf:Description>
	</rdf:RDF>`,
		"",

		"TODO",
	},
	{
		// [30] #rdfms-difference-between-ID-and-about-test1
		//
		// A statement with an rdf:ID creates a regular triple.
		//
		"rdfms-difference-between-ID-and-about/test1.rdf",
		`<!--  
Base URI: http://www.w3.org/2013/RDFXMLTests/rdfms-difference-between-ID-and-about/test1.rdf

A statement with an rdf:ID creates a regular triple.
--> 
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description rdf:ID="foo">
  <rdf:value>abc</rdf:value>
</rdf:Description>
</rdf:RDF>`,
		`<http://www.w3.org/2013/RDFXMLTests/rdfms-difference-between-ID-and-about/test1.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#value> "abc" .
`,
		"",
	},
	{
		// [31] #rdfms-difference-between-ID-and-about-test2
		//
		// This test shows the treatment of non-ASCII characters in the
		// value of rdf:ID attribute.
		//
		"rdfms-difference-between-ID-and-about/test2.rdf",
		`<!--  
Base URI: http://www.w3.org/2013/RDFXMLTests/rdfms-difference-between-ID-and-about/test2.rdf

Non-ASCII characters in IDs are not converted.
-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description rdf:ID="D&#xFC;rst">
  <rdf:value>abc</rdf:value>
</rdf:Description>
</rdf:RDF>`,
		`<http://www.w3.org/2013/RDFXMLTests/rdfms-difference-between-ID-and-about/test2.rdf#Dürst> <http://www.w3.org/1999/02/22-rdf-syntax-ns#value> "abc" .
`,
		"",
	},
	{
		// [32] #rdfms-difference-between-ID-and-about-test3
		//
		// This test shows the treatment of non-ASCII characters in the
		// value of rdf:about attribute.
		//
		"rdfms-difference-between-ID-and-about/test3.rdf",
		`<!--  
Base URI: http://www.w3.org/2013/RDFXMLTests/rdfms-difference-between-ID-and-about/test3.rdf

Non-ASCII characters in URIs are not converted.
-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
<rdf:Description rdf:about="#D&#xFC;rst">
  <rdf:value>abc</rdf:value>
</rdf:Description>
</rdf:RDF>`,
		`<http://www.w3.org/2013/RDFXMLTests/rdfms-difference-between-ID-and-about/test3.rdf#Dürst> <http://www.w3.org/1999/02/22-rdf-syntax-ns#value> "abc" .
`,
		"",
	},
	{
		// [33] #rdfms-duplicate-member-props-test001
		//
		// The question posed to the RDF WG was: should an RDF document
		// containing multiple rdf:_n properties (with the same n) on an
		// element be rejected as illegal? The WG decided that a parser
		// should accept that case as legal RDF.
		//
		"rdfms-duplicate-member-props/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Bag rdf:about="http://example.org/foo">
     <rdf:_1 rdf:resource="http://example.org/a" />
     <rdf:_1 rdf:resource="http://example.org/b" />
  </rdf:Bag>
</rdf:RDF>`,
		`<http://example.org/foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag> .
<http://example.org/foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> <http://example.org/a> .
<http://example.org/foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> <http://example.org/b> .
`,
		"",
	},
	{
		// [34] #rdfms-empty-property-elements-error001
		//
		// This is not legal RDF; specifying an rdf:parseType of
		// "Literal" and an rdf:resource attribute at the same time is an
		// error.
		//
		"rdfms-empty-property-elements/error001.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/error001.nrdf

 Description:

 This is not legal RDF; specifying an rdf:parseType of "Literal" and an
 rdf:resource attribute at the same time is an error.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">

<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty rdf:parseType="Literal"
    rdf:resource="http://random.ioctl.org/#foo" />
</rdf:Description>

</rdf:RDF>`,
		"",

		"cannot have both rdf:parseType=\"Literal\" and rdf:resource",
	},
	{
		// [35] #rdfms-empty-property-elements-error002
		//
		// This is not legal RDF; specifying an rdf:parseType of
		// "Literal" and an rdf:resource attribute at the same time is an
		// error.
		//
		"rdfms-empty-property-elements/error002.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/error002.nrdf

 Description:

 This is not legal RDF; specifying an rdf:parseType of "Literal" and an
 rdf:resource attribute at the same time is an error.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">

<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty rdf:parseType="Literal"
    rdf:resource="http://random.ioctl.org/#foo"></random:someProperty>
</rdf:Description>

</rdf:RDF>`,
		"",

		"cannot have both rdf:parseType=\"Literal\" and rdf:resource",
	},
	{
		// [36] #rdfms-empty-property-elements-test001
		//
		// The rdf:resource attribute means that the value of this
		// property element is a resource.
		//
		"rdfms-empty-property-elements/test001.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test001.rdf

 Description:

 The rdf:resource attribute means that the value of this property element
 is a resource.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">

<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty rdf:resource="http://random.ioctl.org/#foo" />
</rdf:Description>

</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> <http://random.ioctl.org/#foo> .
`,
		"",
	},
	{
		// [37] #rdfms-empty-property-elements-test002
		//
		// The basic case. An empty property element just gives an empty
		// literal.
		//
		"rdfms-empty-property-elements/test002.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test002.rdf

 Description:

 The basic case. An empty property element just gives an empty literal.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">

<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty />
</rdf:Description>

</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> "" .
`,
		"",
	},
	{
		// [38] #rdfms-empty-property-elements-test004
		//
		// If the parseType indicates the value is a resource, we must
		// create one. With no additional information, the resource is
		// anonymous.
		//
		"rdfms-empty-property-elements/test004.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test004.rdf

 Description:

 If the parseType indicates the value is a resource, we must create one. With
 no additional information, the resource is anonymous.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">

<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty rdf:parseType="Resource" />
</rdf:Description>

</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> _:b0 .
`,
		"",
	},
	{
		// [39] #rdfms-empty-property-elements-test005
		//
		// An empty property element just gives an empty literal. We
		// reify the statement at the same time.
		//
		"rdfms-empty-property-elements/test005.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test005.rdf

 Description:

 An empty property element just gives an empty literal. We reify the statement
 at the same time.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
   xmlns:random="http://random.ioctl.org/#">
 
 <rdf:Description rdf:about="http://random.ioctl.org/#bar">
   <random:someProperty rdf:ID="foo" />
 </rdf:Description>

</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> "" .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test005.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test005.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> <http://random.ioctl.org/#bar> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test005.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://random.ioctl.org/#someProperty> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test005.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> "" .
`,
		"",
	},
	{
		// [40] #rdfms-empty-property-elements-test006
		//
		// Here the parseType indicates that we should create a resource.
		// We also reify the generated statement.
		//
		"rdfms-empty-property-elements/test006.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test006.rdf

 Description:

 Here the parseType indicates that we should create a resource. We also
 reify the generated statement.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
   xmlns:random="http://random.ioctl.org/#">
 
 <rdf:Description rdf:about="http://random.ioctl.org/#bar">
   <random:someProperty rdf:ID="foo" rdf:parseType="Resource" />
 </rdf:Description>

</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> _:b0 .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test006.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test006.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> <http://random.ioctl.org/#bar> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test006.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://random.ioctl.org/#someProperty> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test006.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> _:b0 .
`,
		"",
	},
	{
		// [41] #rdfms-empty-property-elements-test007
		//
		// As test001.rdf; this uses an explicit closing tag.
		//
		"rdfms-empty-property-elements/test007.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test007.rdf

 Description:

 As test001.rdf; this uses an explicit closing tag.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">

<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty rdf:resource="http://random.ioctl.org/#foo"></random:someProperty>
</rdf:Description>

</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> <http://random.ioctl.org/#foo> .
`,
		"",
	},
	{
		// [42] #rdfms-empty-property-elements-test008
		//
		// As test002.rdf; this uses an explicit closing tag.
		//
		"rdfms-empty-property-elements/test008.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test008.rdf

 Description:

 As test002.rdf; this uses an explicit closing tag.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">

<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty></random:someProperty>
</rdf:Description>

</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> "" .
`,
		"",
	},
	{
		// [43] #rdfms-empty-property-elements-test010
		//
		// As test004.rdf; this uses an explicit closing tag.
		//
		"rdfms-empty-property-elements/test010.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test010.rdf

 Description:

 As test004.rdf; this uses an explicit closing tag.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">

<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty rdf:parseType="Resource"></random:someProperty>
</rdf:Description>

</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> _:b0 .
`,
		"",
	},
	{
		// [44] #rdfms-empty-property-elements-test011
		//
		// As test005.rdf; this uses an explicit closing tag.
		//
		"rdfms-empty-property-elements/test011.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test011.rdf

 Description:

 As test005.rdf; this uses an explicit closing tag.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
   xmlns:random="http://random.ioctl.org/#">
 
 <rdf:Description rdf:about="http://random.ioctl.org/#bar">
   <random:someProperty rdf:ID="foo"></random:someProperty>
 </rdf:Description>
</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> "" .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test011.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test011.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> <http://random.ioctl.org/#bar> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test011.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://random.ioctl.org/#someProperty> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test011.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> "" .
`,
		"",
	},
	{
		// [45] #rdfms-empty-property-elements-test012
		//
		// As test006.rdf; this uses an explicit closing tag.
		//
		"rdfms-empty-property-elements/test012.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test012.rdf

 Description:

 As test006.rdf; this uses an explicit closing tag.

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
   xmlns:random="http://random.ioctl.org/#">
 
 <rdf:Description rdf:about="http://random.ioctl.org/#bar">
   <random:someProperty rdf:ID="foo" rdf:parseType="Resource"></random:someProperty>
 </rdf:Description>
</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> _:b0 .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test012.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test012.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> <http://random.ioctl.org/#bar> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test012.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://random.ioctl.org/#someProperty> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test012.rdf#foo> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> _:b0 .
`,
		"",
	},
	{
		// [46] #rdfms-empty-property-elements-test013
		//
		// Test of the last alternative for production [6.12],
		// interpreted according to RDFMS paragraphs 229-234:
		// http://lists.w3.org/Archives/Public/www-archive/2001Jun/att-0021/00-part#229
		//
		"rdfms-empty-property-elements/test013.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test013.rdf

 Description:

 Test of the last alternative for production [6.12],
 interpreted according to RDFMS paragraphs 229-234:
http://lists.w3.org/Archives/Public/www-archive/2001Jun/att-0021/00-part#229

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">
 
<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty rdf:resource="http://random.ioctl.org/#foo"
        random:prop2="baz" />
</rdf:Description>
</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> <http://random.ioctl.org/#foo> .
<http://random.ioctl.org/#foo> <http://random.ioctl.org/#prop2> "baz" .
`,
		"",
	},
	{
		// [47] #rdfms-empty-property-elements-test014
		//
		// Test of the last alternative for production [6.12],
		// interpreted according to RDFMS paragraphs 229-234:
		// http://lists.w3.org/Archives/Public/www-archive/2001Jun/att-0021/00-part#229
		//
		"rdfms-empty-property-elements/test014.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test014.rdf

 Description:

 Test of the last alternative for production [6.12],
 interpreted according to RDFMS paragraphs 229-234:
http://lists.w3.org/Archives/Public/www-archive/2001Jun/att-0021/00-part#229

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">
 
<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty random:prop2="baz" />
</rdf:Description>
</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> _:b0 .
_:b0 <http://random.ioctl.org/#prop2> "baz" .
`,
		"",
	},
	{
		// [48] #rdfms-empty-property-elements-test015
		//
		// Test of the last alternative for production [6.12],
		// interpreted according to RDFMS paragraphs 229-234:
		// http://lists.w3.org/Archives/Public/www-archive/2001Jun/att-0021/00-part#229
		// Here we have an explicit closing tag. This does not match any
		// of the productions in the original document, but is
		// indistinguishable from test014 as far as XML is concerned.
		//
		"rdfms-empty-property-elements/test015.rdf",
		`<!--

 Assumed base URI:

http://www.w3.org/2013/RDFXMLTests/rdfms-empty-property-elements/test015.rdf

 Description:

 Test of the last alternative for production [6.12],
 interpreted according to RDFMS paragraphs 229-234:
http://lists.w3.org/Archives/Public/www-archive/2001Jun/att-0021/00-part#229
 Here we have an explicit closing tag. This does not match any
 of the productions in the original document, but is indistinguishable
 from test014 as far as XML is concerned.

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">
 
<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty random:prop2="baz"></random:someProperty>
</rdf:Description>
</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> _:b0 .
_:b0 <http://random.ioctl.org/#prop2> "baz" .
`,
		"",
	},
	{
		// [49] #rdfms-empty-property-elements-test016
		//
		// Like rdfms-empty-property-elements/test001.rdf but with a
		// processing instruction as the only content of the otherwise
		// empty element.
		//
		"rdfms-empty-property-elements/test016.rdf",
		`<!--

 Description:
 Like test001.rdf but with a processing instruction 
 as the only content of the otherwise empty element.

 Author: Jeremy Carroll

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">

<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty rdf:resource="http://random.ioctl.org/#foo"><?a 
       processing    instruction?></random:someProperty>
</rdf:Description>

</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> <http://random.ioctl.org/#foo> .
`,
		"",
	},
	{
		// [50] #rdfms-empty-property-elements-test017
		//
		// Like rdfms-empty-property-elements/test001.rdf but with a
		// comment as the only content of the otherwise empty element.
		//
		"rdfms-empty-property-elements/test017.rdf",
		`<!--

 Description:
 Like test001.rdf but with a comment 
 as the only content of the otherwise empty element.

 Author: Jeremy Carroll

-->
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns:random="http://random.ioctl.org/#">

<rdf:Description rdf:about="http://random.ioctl.org/#bar">
  <random:someProperty rdf:resource="http://random.ioctl.org/#foo"><!--
      A comment

 Even with a comment or a processing instruction within an empty
 property element, it is still empty because an RDF Parser ignores
 the processing instruction and comment nodes when not within an 
 XML Literal.

--></random:someProperty>
</rdf:Description>

</rdf:RDF>`,
		`<http://random.ioctl.org/#bar> <http://random.ioctl.org/#someProperty> <http://random.ioctl.org/#foo> .
`,
		"",
	},
	{
		// [51] #rdfms-identity-anon-resources-test001
		//
		// a RDF Description with no ID or about attribute describes an
		// un-named resource, aka a bNode.
		//
		"rdfms-identity-anon-resources/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Description>
   <eg:property>property value</eg:property>
 </rdf:Description>

</rdf:RDF>`,
		`_:b0 <http://example.org/property> "property value" .
`,
		"",
	},
	{
		// [52] #rdfms-identity-anon-resources-test002
		//
		// a RDF Description with no ID or about attribute describes an
		// un-named resource, aka a bNode.
		//
		"rdfms-identity-anon-resources/test002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <eg:node>
   <eg:property>property value</eg:property>
 </eg:node>

</rdf:RDF>`,
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/node> .
_:b0 <http://example.org/property> "property value" .
`,
		"",
	},
	{
		// [53] #rdfms-identity-anon-resources-test003
		//
		// a RDF container (in this case a Bag) without an ID attribute
		// describes an un-named resource, aka a bNode.
		//
		"rdfms-identity-anon-resources/test003.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Bag/>

</rdf:RDF>`,
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag> .
`,
		"",
	},
	{
		// [54] #rdfms-identity-anon-resources-test004
		//
		// a RDF container (in this case an Alt) without an ID attribute
		// describes an un-named resource, aka a bNode.
		//
		"rdfms-identity-anon-resources/test004.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Alt>
  <rdf:li>some value</rdf:li>
 </rdf:Alt>

</rdf:RDF>`,
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Alt> .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> "some value" .
`,
		"",
	},
	{
		// [55] #rdfms-identity-anon-resources-test005
		//
		// a RDF container (in this case an Seq) without an ID attribute
		// describes an un-named resource, aka a bNode.
		//
		"rdfms-identity-anon-resources/test005.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Seq/>

</rdf:RDF>`,
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq> .
`,
		"",
	},
	{
		// [56] #rdfms-not-id-and-resource-attr-test001
		//
		// rdf:ID on an empty property element indicates reification.
		//
		"rdfms-not-id-and-resource-attr/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

  <rdf:Description>
    <eg:prop1  rdf:ID="reify" eg:prop2="val"></eg:prop1>
  </rdf:Description>
</rdf:RDF>`,
		`_:b0 <http://example.org/prop1> _:b1 .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test001.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test001.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> _:b0 .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test001.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://example.org/prop1> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test001.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> _:b1 .
_:b1 <http://example.org/prop2> "val" .
`,
		"",
	},
	{
		// [57] #rdfms-not-id-and-resource-attr-test002
		//
		// rdf:reource on an empty property element indicates the URI of
		// the object.
		//
		"rdfms-not-id-and-resource-attr/test002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

  <rdf:Description>
    <eg:prop1  rdf:resource="http://example.org/object#uriRef" eg:prop2="val"></eg:prop1>
  </rdf:Description>
</rdf:RDF>`,
		`_:b0 <http://example.org/prop1> <http://example.org/object#uriRef> .
<http://example.org/object#uriRef> <http://example.org/prop2> "val" .
`,
		"",
	},
	{
		// [58] #rdfms-not-id-and-resource-attr-test004
		//
		// rdf:ID and rdf:resource are allowed together on empty property
		// element.
		//
		"rdfms-not-id-and-resource-attr/test004.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

  <rdf:Description>
    <eg:prop1  rdf:ID="reify" rdf:resource="http://example.org/object"/>
  </rdf:Description>
</rdf:RDF>`,
		`_:b0 <http://example.org/prop1> <http://example.org/object> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test004.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test004.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> _:b0 .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test004.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://example.org/prop1> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test004.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> <http://example.org/object> .
`,
		"",
	},
	{
		// [59] #rdfms-not-id-and-resource-attr-test005
		//
		// rdf:ID and rdf:resource are allowed together on empty property
		// element.
		//
		"rdfms-not-id-and-resource-attr/test005.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

  <rdf:Description>
    <eg:prop1  rdf:resource="http://example.org/object" rdf:ID="reify" eg:prop2="val"/>
  </rdf:Description>
</rdf:RDF>`,
		`_:b0 <http://example.org/prop1> <http://example.org/object> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test005.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test005.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> _:b0 .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test005.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://example.org/prop1> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-not-id-and-resource-attr/test005.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> <http://example.org/object> .
<http://example.org/object> <http://example.org/prop2> "val" .
`,
		"",
	},
	{
		// [60] #rdfms-para196-test001
		//
		// test case showing that the 2nd URI in M Paragraph 196 is
		// permitted as a namespace URI (and any namespace URI starting
		// with that URI)
		//
		"rdfms-para196/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:a="http://www.w3.org/TR/REC-rdf-syntax"
         xmlns:b="http://www.w3.org/TR/REC-rdf-syntax-blah-blah"
         xmlns:c="http://www.w3.org/TR/REC-rdf-syntax#">
  <rdf:Description rdf:about="http://example.org/">
     <a:foo>permitted</a:foo>
     <b:bar>also permitted</b:bar>
     <c:baz>this one also permitted</c:baz>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/> <http://www.w3.org/TR/REC-rdf-syntaxfoo> "permitted" .
<http://example.org/> <http://www.w3.org/TR/REC-rdf-syntax-blah-blahbar> "also permitted" .
<http://example.org/> <http://www.w3.org/TR/REC-rdf-syntax#baz> "this one also permitted" .
`,
		"",
	},
	{
		// [61] #rdfms-rdf-id-error001
		//
		// The value of rdf:ID must match the XML Name production, (as
		// modified by XML Namespaces).
		//
		"rdfms-rdf-id/error001.rdf",
		`<!--

  The value of rdf:ID must match the XML Name production,
  (as modified by XML Namespaces). 
  $Id: error001.rdf,v 1.1 2002/07/30 09:45:51 jcarroll Exp $

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

 <rdf:Description rdf:ID='333-555-666' />

</rdf:RDF>`,
		"",

		"rdf:ID is not a valid XML NCName: \"333-555-666\"",
	},
	{
		// [62] #rdfms-rdf-id-error002
		//
		// The value of rdf:ID must match the XML Name production, (as
		// modified by XML Namespaces).
		//
		"rdfms-rdf-id/error002.rdf",
		`<!--

  The value of rdf:ID must match the XML Name production,
  (as modified by XML Namespaces). 
  $Id: error002.rdf,v 1.1 2002/07/30 09:45:51 jcarroll Exp $

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

 <rdf:Description rdf:ID="_:xx" />

</rdf:RDF>`,
		"",

		"rdf:ID is not a valid XML NCName: \"_:xx\"",
	},
	{
		// [63] #rdfms-rdf-id-error003
		//
		// The value of rdf:ID must match the XML Name production, (as
		// modified by XML Namespaces).
		//
		"rdfms-rdf-id/error003.rdf",
		`<!--

  The value of rdf:ID must match the XML Name production,
  (as modified by XML Namespaces). 
  $Id: error003.rdf,v 1.1 2002/07/30 09:45:51 jcarroll Exp $

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Description>
   <eg:prop rdf:ID="q:name" />
 </rdf:Description>

</rdf:RDF>`,
		"",

		"rdf:ID is not a valid XML NCName: \"q:name\"",
	},
	{
		// [64] #rdfms-rdf-id-error004
		//
		// The value of rdf:ID must match the XML Name production, (as
		// modified by XML Namespaces).
		//
		"rdfms-rdf-id/error004.rdf",
		`<!--

  The value of rdf:ID must match the XML Name production,
  (as modified by XML Namespaces). 
  $Id: error004.rdf,v 1.1 2002/07/30 09:45:51 jcarroll Exp $

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Description rdf:ID="a/b" eg:prop="val" />

</rdf:RDF>`,
		"",

		"rdf:ID is not a valid XML NCName: \"a/b\"",
	},
	{
		// [65] #rdfms-rdf-id-error005
		//
		// The value of rdf:ID must match the XML Name production, (as
		// modified by XML Namespaces).
		//
		"rdfms-rdf-id/error005.rdf",
		`<!--

  The value of rdf:ID must match the XML Name production,
  (as modified by XML Namespaces). 
  $Id: error005.rdf,v 1.1 2002/07/30 09:45:51 jcarroll Exp $

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <!-- &#x301; is a non-spacing acute accent.
      It is legal within an XML Name, but not as the first
      character.     -->

 <rdf:Description rdf:ID="&#x301;bb" eg:prop="val" />

</rdf:RDF>`,
		"",

		"rdf:ID is not a valid XML NCName: \"\u0301bb\"",
	},
	{
		// [66] #rdfms-rdf-id-error006
		//
		// The value of rdf:bagID must match the XML Name production, (as
		// modified by XML Namespaces).
		//
		"rdfms-rdf-id/error006.rdf",
		`<!--

  The value of rdf:bagID must match the XML Name production,
  (as modified by XML Namespaces). 
  $Id: error006.rdf,v 1.1 2002/07/30 09:45:51 jcarroll Exp $

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

 <rdf:Description rdf:bagID='333-555-666' />

</rdf:RDF>`,
		"",

		"deprecated: rdf:bagID",
	},
	{
		// [67] #rdfms-rdf-id-error007
		//
		// The value of rdf:bagID must match the XML Name production, (as
		// modified by XML Namespaces).
		//
		"rdfms-rdf-id/error007.rdf",
		`<!--

  The value of rdf:bagID must match the XML Name production,
  (as modified by XML Namespaces). 
  $Id: error007.rdf,v 1.1 2002/07/30 09:45:51 jcarroll Exp $

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Description>
   <eg:prop rdf:bagID="q:name" />
 </rdf:Description>

</rdf:RDF>`,
		"",

		"deprecated: rdf:bagID",
	},
	{
		// [68] #rdfms-rdf-names-use-error-001
		//
		// RDF is forbidden as a node element name.
		//
		"rdfms-rdf-names-use/error-001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:RDF/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:RDF",
	},
	{
		// [69] #rdfms-rdf-names-use-error-002
		//
		// ID is forbidden as a node element name.
		//
		"rdfms-rdf-names-use/error-002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:ID/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:ID",
	},
	{
		// [70] #rdfms-rdf-names-use-error-003
		//
		// about is forbidden as a node element name.
		//
		"rdfms-rdf-names-use/error-003.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:about/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:about",
	},
	{
		// [71] #rdfms-rdf-names-use-error-004
		//
		// bagID is forbidden as a node element name.
		//
		"rdfms-rdf-names-use/error-004.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:bagID/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:bagID",
	},
	{
		// [72] #rdfms-rdf-names-use-error-005
		//
		// parseType is forbidden as a node element name.
		//
		"rdfms-rdf-names-use/error-005.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:parseType/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:parseType",
	},
	{
		// [73] #rdfms-rdf-names-use-error-006
		//
		// resource is forbidden as a node element name.
		//
		"rdfms-rdf-names-use/error-006.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:resource/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:resource",
	},
	{
		// [74] #rdfms-rdf-names-use-error-007
		//
		// nodeID is forbidden as a node element name.
		//
		"rdfms-rdf-names-use/error-007.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:nodeID/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:nodeID",
	},
	{
		// [75] #rdfms-rdf-names-use-error-008
		//
		// li is forbidden as a node element name.
		//
		"rdfms-rdf-names-use/error-008.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:li/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:li",
	},
	{
		// [76] #rdfms-rdf-names-use-error-009
		//
		// aboutEach is forbidden as a node element name.
		//
		"rdfms-rdf-names-use/error-009.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:aboutEach/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:aboutEach",
	},
	{
		// [77] #rdfms-rdf-names-use-error-010
		//
		// aboutEachPrefix is forbidden as a node element name.
		//
		"rdfms-rdf-names-use/error-010.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:aboutEachPrefix/>
</rdf:RDF>`,
		"",

		"disallowed as node element name: rdf:aboutEachPrefix",
	},
	{
		// [78] #rdfms-rdf-names-use-error-011
		//
		// Description is forbidden as a property element name.
		//
		"rdfms-rdf-names-use/error-011.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:Description rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		"",

		"disallowed as property element name: rdf:Description",
	},
	{
		// [79] #rdfms-rdf-names-use-error-012
		//
		// RDF is forbidden as a property element name.
		//
		"rdfms-rdf-names-use/error-012.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:RDF rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		"",

		"disallowed as property element name: rdf:RDF",
	},
	{
		// [80] #rdfms-rdf-names-use-error-013
		//
		// ID is forbidden as a property element name.
		//
		"rdfms-rdf-names-use/error-013.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:ID rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		"",

		"disallowed as property element name: rdf:ID",
	},
	{
		// [81] #rdfms-rdf-names-use-error-014
		//
		// about is forbidden as a property element name.
		//
		"rdfms-rdf-names-use/error-014.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:about rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		"",

		"disallowed as property element name: rdf:about",
	},
	{
		// [82] #rdfms-rdf-names-use-error-015
		//
		// bagID is forbidden as a property element name.
		//
		"rdfms-rdf-names-use/error-015.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:bagID rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		"",

		"disallowed as property element name: rdf:bagID",
	},
	{
		// [83] #rdfms-rdf-names-use-error-016
		//
		// parseType is forbidden as a property element name.
		//
		"rdfms-rdf-names-use/error-016.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:parseType rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		"",

		"disallowed as property element name: rdf:parseType",
	},
	{
		// [84] #rdfms-rdf-names-use-error-017
		//
		// resource is forbidden as a property element name.
		//
		"rdfms-rdf-names-use/error-017.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:resource rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		"",

		"disallowed as property element name: rdf:resource",
	},
	{
		// [85] #rdfms-rdf-names-use-error-018
		//
		// nodeID is forbidden as a property element name.
		//
		"rdfms-rdf-names-use/error-018.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:nodeID rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		"",

		"disallowed as property element name: rdf:nodeID",
	},
	{
		// [86] #rdfms-rdf-names-use-error-019
		//
		// aboutEach is forbidden as a property element name.
		//
		"rdfms-rdf-names-use/error-019.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:aboutEach rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		"",

		"disallowed as property element name: rdf:aboutEach",
	},
	{
		// [87] #rdfms-rdf-names-use-error-020
		//
		// aboutEachPrefix is forbidden as a property element name.
		//
		"rdfms-rdf-names-use/error-020.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:aboutEachPrefix rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		"",

		"disallowed as property element name: rdf:aboutEachPrefix",
	},
	{
		// [88] #rdfms-rdf-names-use-test-001
		//
		// Description is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		``, // valid, but no triples generated
		"",
	},
	{
		// [89] #rdfms-rdf-names-use-test-002
		//
		// Seq is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Seq rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq> .
`,
		"",
	},
	{
		// [90] #rdfms-rdf-names-use-test-003
		//
		// Bag is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-003.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Bag rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag> .
`,
		"",
	},
	{
		// [91] #rdfms-rdf-names-use-test-004
		//
		// Alt is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-004.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Alt rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Alt> .
`,
		"",
	},
	{
		// [92] #rdfms-rdf-names-use-test-005
		//
		// Statement is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-005.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Statement rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
`,
		"",
	},
	{
		// [93] #rdfms-rdf-names-use-test-006
		//
		// Property is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-006.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Property rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Property> .
`,
		"",
	},
	{
		// [94] #rdfms-rdf-names-use-test-007
		//
		// List is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-007.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:List rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#List> .
`,
		"",
	},
	{
		// [95] #rdfms-rdf-names-use-test-008
		//
		// subject is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-008.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:subject rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> .
`,
		"",
	},
	{
		// [96] #rdfms-rdf-names-use-test-009
		//
		// predicate is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-009.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:predicate rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> .
`,
		"",
	},
	{
		// [97] #rdfms-rdf-names-use-test-010
		//
		// object is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-010.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:object rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> .
`,
		"",
	},
	{
		// [98] #rdfms-rdf-names-use-test-011
		//
		// type is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-011.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:type rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> .
`,
		"",
	},
	{
		// [99] #rdfms-rdf-names-use-test-012
		//
		// value is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-012.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:value rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#value> .
`,
		"",
	},
	{
		// [100] #rdfms-rdf-names-use-test-013
		//
		// first is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-013.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:first rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#first> .
`,
		"",
	},
	{
		// [101] #rdfms-rdf-names-use-test-014
		//
		// rest is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-014.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:rest rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#rest> .
`,
		"",
	},
	{
		// [102] #rdfms-rdf-names-use-test-015
		//
		// _1 is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-015.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:_1 rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> .
`,
		"",
	},
	{
		// [103] #rdfms-rdf-names-use-test-016
		//
		// nil is allowed as a node element name.
		//
		"rdfms-rdf-names-use/test-016.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:nil rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#nil> .
`,
		"",
	},
	{
		// [104] #rdfms-rdf-names-use-test-017
		//
		// Seq is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-017.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:Seq rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq> <http://example.org/node2> .
`,
		"",
	},
	{
		// [105] #rdfms-rdf-names-use-test-018
		//
		// Bag is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-018.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:Bag rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag> <http://example.org/node2> .
`,
		"",
	},
	{
		// [106] #rdfms-rdf-names-use-test-019
		//
		// Alt is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-019.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:Alt rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Alt> <http://example.org/node2> .
`,
		"",
	},
	{
		// [107] #rdfms-rdf-names-use-test-020
		//
		// Statement is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-020.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:Statement rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> <http://example.org/node2> .
`,
		"",
	},
	{
		// [108] #rdfms-rdf-names-use-test-021
		//
		// Property is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-021.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:Property rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Property> <http://example.org/node2> .
`,
		"",
	},
	{
		// [109] #rdfms-rdf-names-use-test-022
		//
		// List is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-022.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:List rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#List> <http://example.org/node2> .
`,
		"",
	},
	{
		// [110] #rdfms-rdf-names-use-test-023
		//
		// subject is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-023.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:subject rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> <http://example.org/node2> .
`,
		"",
	},
	{
		// [111] #rdfms-rdf-names-use-test-024
		//
		// predicate is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-024.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:predicate rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://example.org/node2> .
`,
		"",
	},
	{
		// [112] #rdfms-rdf-names-use-test-025
		//
		// object is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-025.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:object rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> <http://example.org/node2> .
`,
		"",
	},
	{
		// [113] #rdfms-rdf-names-use-test-026
		//
		// type is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-026.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:type rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/node2> .
`,
		"",
	},
	{
		// [114] #rdfms-rdf-names-use-test-027
		//
		// value is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-027.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:value rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#value> <http://example.org/node2> .
`,
		"",
	},
	{
		// [115] #rdfms-rdf-names-use-test-028
		//
		// first is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-028.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:first rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#first> <http://example.org/node2> .
`,
		"",
	},
	{
		// [116] #rdfms-rdf-names-use-test-029
		//
		// rest is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-029.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:rest rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#rest> <http://example.org/node2> .
`,
		"",
	},
	{
		// [117] #rdfms-rdf-names-use-test-030
		//
		// _1 is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-030.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:_1 rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> <http://example.org/node2> .
`,
		"",
	},
	{
		// [118] #rdfms-rdf-names-use-test-031
		//
		// li is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-031.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:li rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#_1> <http://example.org/node2> .
`,
		"",
	},
	{
		// [119] #rdfms-rdf-names-use-test-032
		//
		// Seq is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-032.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1"
    rdf:Seq="string" />
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq> "string" .
`,
		"",
	},
	{
		// [120] #rdfms-rdf-names-use-test-033
		//
		// Bag is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-033.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1"
    rdf:Bag="string" />
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag> "string" .
`,
		"",
	},
	{
		// [121] #rdfms-rdf-names-use-test-034
		//
		// Alt is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-034.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1"
    rdf:Alt="string" />
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Alt> "string" .
`,
		"",
	},
	{
		// [122] #rdfms-rdf-names-use-test-035
		//
		// Statement is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-035.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1"
    rdf:Statement="string" />
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> "string" .
`,
		"",
	},
	{
		// [123] #rdfms-rdf-names-use-test-036
		//
		// Property is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-036.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1"
    rdf:Property="string" />
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Property> "string" .
`,
		"",
	},
	{
		// [124] #rdfms-rdf-names-use-test-037
		//
		// List is allowed as a property element name.
		//
		"rdfms-rdf-names-use/test-037.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1"
    rdf:List="string" />
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#List> "string" .
`,
		"",
	},
	{
		// [125] #rdfms-rdf-names-use-warn-001
		//
		// foo is allowed with warnings as a node element name.
		//
		"rdfms-rdf-names-use/warn-001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:foo rdf:about="http://example.org/node"/>
</rdf:RDF>`,
		`<http://example.org/node> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#foo> .
`,
		"",
	},
	{
		// [126] #rdfms-rdf-names-use-warn-002
		//
		// foo is allowed with warnings as a property element name.
		//
		"rdfms-rdf-names-use/warn-002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1">
    <rdf:foo rdf:resource="http://example.org/node2"/>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#foo> <http://example.org/node2> .
`,
		"",
	},
	{
		// [127] #rdfms-rdf-names-use-warn-003
		//
		// foo is allowed with warnings as a property attribute name.
		//
		"rdfms-rdf-names-use/warn-003.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/node1"
    rdf:foo="string" />
</rdf:RDF>`,
		`<http://example.org/node1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#foo> "string" .
`,
		"",
	},
	{
		// [128] #rdfms-reification-required-test001
		//
		// A parser is not required to generate a bag of reified
		// statements for all description elements.
		//
		"rdfms-reification-required/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org#">
  <rdf:Description rdf:about="http://example.org/" eg:prop="10"/>
</rdf:RDF>`,
		`<http://example.org/> <http://example.org#prop> "10" .
`,
		"",
	},
	{
		// [129] #rdfms-seq-representation-test001
		//
		// rdf:parseType="Collection" is parsed like the nonstandard
		// daml:collection.
		//
		"rdfms-seq-representation/test001.rdf",
		`<rdf:RDF
    xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
    xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#"
    xmlns:eg="http://example.org/eg#">

    <rdf:Description rdf:about="http://example.org/eg#eric">
        <rdf:type rdf:parseType="Resource">
            <eg:intersectionOf rdf:parseType="Collection">
                <rdf:Description rdf:about="http://example.org/eg#Person"/>
                <rdf:Description rdf:about="http://example.org/eg#Male"/>
            </eg:intersectionOf>
        </rdf:type>
    </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/eg#eric> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> _:b0 .
_:b0 <http://example.org/eg#intersectionOf> _:b1 .
_:b1 <http://www.w3.org/1999/02/22-rdf-syntax-ns#first> <http://example.org/eg#Person> .
_:b1 <http://www.w3.org/1999/02/22-rdf-syntax-ns#rest> _:b2 .
_:b2 <http://www.w3.org/1999/02/22-rdf-syntax-ns#first> <http://example.org/eg#Male> .
_:b2 <http://www.w3.org/1999/02/22-rdf-syntax-ns#rest> <http://www.w3.org/1999/02/22-rdf-syntax-ns#nil> .
`,
		"",
	},
	{
		// [130] #rdfms-syntax-incomplete-test001
		//
		// rdf:nodeID can be used to label a blank node.
		//
		"rdfms-syntax-incomplete/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Description rdf:nodeID="a">
   <eg:property rdf:nodeID="a" />
 </rdf:Description>

</rdf:RDF>`,
		`_:a <http://example.org/property> _:a .
`,
		"",
	},
	{
		// [131] #rdfms-syntax-incomplete-test002
		//
		// rdf:nodeID can be used to label a blank node. These have file
		// scope and are distinct from any unlabelled blank nodes.
		//
		"rdfms-syntax-incomplete/test002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Description rdf:nodeID="a">
   <eg:property1 rdf:nodeID="a" />
 </rdf:Description>
 <rdf:Description>
   <eg:property2>
<!-- Note the rdf:nodeID="b" is redundant. -->
      <rdf:Description rdf:nodeID="b">
            <eg:property3 rdf:nodeID="a" />
      </rdf:Description>
   </eg:property2>
 </rdf:Description>

</rdf:RDF>`,
		`_:a <http://example.org/property1> _:a .
_:b0 <http://example.org/property2> _:b .
_:b <http://example.org/property3> _:a .
`,
		"",
	},
	{
		// [132] #rdfms-syntax-incomplete-test003
		//
		// On an rdf:Description or typed node rdf:nodeID behaves
		// similarly to an rdf:about.
		//
		"rdfms-syntax-incomplete/test003.rdf",
		`<!--

  On an rdf:Description or typed node rdf:nodeID behaves
  similarly to an rdf:about.
  $Id: test003.rdf,v 1.2 2003/07/24 15:51:06 jcarroll Exp $

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <!-- In this example the rdf:nodeID is redundant. -->
 <rdf:Description rdf:nodeID="a" eg:property1="value" />

</rdf:RDF>`,
		`_:a <http://example.org/property1> "value" .
`,
		"",
	},
	{
		// [133] #rdfms-syntax-incomplete-test004
		//
		// On a property element rdf:nodeID behaves similarly to
		// rdf:resource.
		//
		"rdfms-syntax-incomplete/test004.rdf",
		`<!--

  On a property element rdf:nodeID behaves
  similarly to rdf:resource.
  $Id: test004.rdf,v 1.1 2002/07/30 09:46:05 jcarroll Exp $

-->


<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <!-- The rdf:nodeID allows two references to the same node
      as an object of triples in the graph. -->
 <rdf:Description >
   <eg:property1 rdf:ID="reify" rdf:nodeID="a" />
 </rdf:Description>

 <rdf:Description>
   <eg:property2 rdf:nodeID="a"/>
 </rdf:Description>

</rdf:RDF>`,
		`_:b0 <http://example.org/property1> _:a .
<http://www.w3.org/2013/RDFXMLTests/rdfms-syntax-incomplete/test004.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-syntax-incomplete/test004.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> _:b0 .
<http://www.w3.org/2013/RDFXMLTests/rdfms-syntax-incomplete/test004.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://example.org/property1> .
<http://www.w3.org/2013/RDFXMLTests/rdfms-syntax-incomplete/test004.rdf#reify> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> _:a .
_:b1 <http://example.org/property2> _:a .
`,
		"",
	},
	{
		// [134] #rdfms-syntax-incomplete-error001
		//
		// The value of rdf:nodeID must match the XML Name production,
		// (as modified by XML Namespaces).
		//
		"rdfms-syntax-incomplete/error001.rdf",
		`<!--

  The value of rdf:nodeID must match the XML Name production,
  (as modified by XML Namespaces). 
  $Id: error001.rdf,v 1.1 2002/07/30 09:45:51 jcarroll Exp $

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

 <rdf:Description rdf:nodeID='333-555-666' />

</rdf:RDF>`,
		"",

		"rdf:nodeID is not a valid XML NCName: \"333-555-666\"",
	},
	{
		// [135] #rdfms-syntax-incomplete-error002
		//
		// The value of rdf:nodeID must match the XML Name production,
		// (as modified by XML Namespaces).
		//
		"rdfms-syntax-incomplete/error002.rdf",
		`<!--

  The value of rdf:nodeID must match the XML Name production,
  (as modified by XML Namespaces). 
  $Id: error002.rdf,v 1.1 2002/07/30 09:45:51 jcarroll Exp $

-->

<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

 <rdf:Description rdf:nodeID="_:bnode" />

</rdf:RDF>`,
		"",

		"rdf:nodeID is not a valid XML NCName: \"_:bnode\"",
	},
	{
		// [136] #rdfms-syntax-incomplete-error003
		//
		// The value of rdf:nodeID must match the XML Name production,
		// (as modified by XML Namespaces).
		//
		"rdfms-syntax-incomplete/error003.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Description>
   <eg:prop rdf:nodeID="q:name" />
 </rdf:Description>

</rdf:RDF>`,
		"",

		"rdf:nodeID is not a valid XML NCName: \"q:name\"",
	},
	{
		// [137] #rdfms-syntax-incomplete-error004
		//
		// Cannot have rdf:nodeID and rdf:ID.
		//
		"rdfms-syntax-incomplete/error004.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

 <rdf:Description rdf:nodeID='a' rdf:ID='b'/>

</rdf:RDF>`,
		"",

		"A node element cannot have both rdf:ID and rdf:nodeID",
	},
	{
		// [138] #rdfms-syntax-incomplete-error005
		//
		// Cannot have rdf:nodeID and rdf:about.
		//
		"rdfms-syntax-incomplete/error005.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

 <rdf:Description rdf:nodeID="a" rdf:about="http://example.org/"/>

</rdf:RDF>`,
		"",

		"A node element cannot have both rdf:about and rdf:nodeID",
	},
	{
		// [139] #rdfms-syntax-incomplete-error006
		//
		// Cannot have rdf:nodeID and rdf:resource.
		//
		"rdfms-syntax-incomplete/error006.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

 <rdf:Description>
   <eg:prop rdf:nodeID="a" rdf:resource="http://www.example.org/" />
 </rdf:Description>

</rdf:RDF>`,
		"",

		"A property element cannot have both rdf:resource and rdf:nodeID",
	},
	{
		// [140] #rdfms-uri-substructure-test001
		//
		// Demonstrates the Recommended partitioning of a URI into a
		// namespace part and a localname part
		//
		"rdfms-uri-substructure/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

<rdf:Description>
  <eg:property>10</eg:property>
</rdf:Description>

</rdf:RDF>`,
		`_:b0 <http://example.org/property> "10" .
`,
		"",
	},
	{
		// [141] #rdfms-xmllang-test003
		//
		// In-scope xml:lang applies to element content literal values
		//
		"rdfms-xmllang/test003.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

  <rdf:Description rdf:about="http://example.org/node">
     <eg:property>chat</eg:property>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node> <http://example.org/property> "chat" .
`,
		"",
	},
	{
		// [142] #rdfms-xmllang-test004
		//
		// In-scope xml:lang applies to element content literal values
		//
		"rdfms-xmllang/test004.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

  <rdf:Description rdf:about="http://example.org/node">
     <eg:property xml:lang="fr">chat</eg:property>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/node> <http://example.org/property> "chat"@fr .
`,
		"",
	},
	{
		// [143] #rdfms-xmllang-test005
		//
		// In-scope xml:lang applies to element content literal values
		//
		"rdfms-xmllang/test005.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

  <rdf:Description rdf:about="http://example.org/node"
                   eg:property="chat" />

</rdf:RDF>`,
		`<http://example.org/node> <http://example.org/property> "chat" .
`,
		"",
	},
	{
		// [144] #rdfms-xmllang-test006
		//
		// In-scope xml:lang applies to element content literal values
		//
		"rdfms-xmllang/test006.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">

  <rdf:Description rdf:about="http://example.org/node"
                   xml:lang="fr"
                   eg:property="chat" />

</rdf:RDF>`,
		`<http://example.org/node> <http://example.org/property> "chat"@fr .
`,
		"",
	},
	{
		// [145] #rdfs-domain-and-range-test001
		//
		// a RDF Property may have more than one domain property
		//
		"rdfs-domain-and-range/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#">

  <rdf:Property rdf:about="http://example.org/bar">
    <rdfs:domain rdf:resource="http://example.org/Domain1"/>
    <rdfs:domain rdf:resource="http://example.org/Domain2"/>
  </rdf:Property>

</rdf:RDF>`,
		`<http://example.org/bar> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Property> .
<http://example.org/bar> <http://www.w3.org/2000/01/rdf-schema#domain> <http://example.org/Domain1> .
<http://example.org/bar> <http://www.w3.org/2000/01/rdf-schema#domain> <http://example.org/Domain2> .
`,
		"",
	},
	{
		// [146] #rdfs-domain-and-range-test002
		//
		// a RDF Property may have more than one domain property
		//
		"rdfs-domain-and-range/test002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#">

  <rdf:Property rdf:about="http://example.org/bar">
    <rdfs:range rdf:resource="http://example.org/Range1"/>
    <rdfs:range rdf:resource="http://example.org/Range2"/>
  </rdf:Property>

</rdf:RDF>`,
		`<http://example.org/bar> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Property> .
<http://example.org/bar> <http://www.w3.org/2000/01/rdf-schema#range> <http://example.org/Range1> .
<http://example.org/bar> <http://www.w3.org/2000/01/rdf-schema#range> <http://example.org/Range2> .
`,
		"",
	},
	{
		// [147] #unrecognised-xml-attributes-test001
		//
		// Unrecognized attributes in the xml namespace should be
		// ignored.
		//
		"unrecognised-xml-attributes/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:ex="http://example.org/schema#">
  <rdf:Description rdf:about="http://example.org/thing">
    <ex:prop1 xml:space="default">blah</ex:prop1>
    <ex:prop2 xml:foo="anything">more</ex:prop2>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/thing> <http://example.org/schema#prop1> "blah" .
<http://example.org/thing> <http://example.org/schema#prop2> "more" .
`,
		"",
	},
	{
		// [148] #unrecognised-xml-attributes-test002
		//
		// Unrecognized attributes in the xml namespace should be
		// ignored.
		//
		"unrecognised-xml-attributes/test002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:ex="http://example.org/schema#">
  <rdf:Description rdf:about="http://example.org/thing">
    <ex:prop1 xmlnewthing="anything">stuff</ex:prop1>
  </rdf:Description>
</rdf:RDF>`,
		`<http://example.org/thing> <http://example.org/schema#prop1> "stuff" .
`,
		"",
	},
	{
		// [149] #xml-canon-test001
		//
		// Demonstrating the canonicalisation of XMLLiterals.
		//
		"xml-canon/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/">


  <rdf:Description rdf:about="http://www.example.org/a">
    <eg:prop rdf:parseType="Literal"><br /></eg:prop>
  </rdf:Description>

</rdf:RDF>`,
		`<http://www.example.org/a> <http://example.org/prop> "<br></br>"^^<http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral> .
`,
		"",
	},
	{
		// [150] #xmlbase-test001
		//
		// xml:base applies to an rdf:ID on an rdf:Description element.
		//
		"xmlbase/test001.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org/dir/file">

 <rdf:Description rdf:ID="frag" eg:value="v" />

</rdf:RDF>`,
		`<http://example.org/dir/file#frag> <http://example.org/value> "v" .
`,
		"",
	},
	{
		// [151] #xmlbase-test002
		//
		// xml:base applies to an rdf:resource attribute.
		//
		"xmlbase/test002.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org/dir/file">

 <rdf:Description>
   <eg:value rdf:resource="relFile" />
 </rdf:Description>

</rdf:RDF>`,
		`_:b0 <http://example.org/value> <http://example.org/dir/relFile> .
`,
		"",
	},
	{
		// [152] #xmlbase-test003
		//
		// xml:base applies to an rdf:about attribute.
		//
		"xmlbase/test003.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org/dir/file">

 <eg:type rdf:about="relfile" />

</rdf:RDF>`,
		`<http://example.org/dir/relfile> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/type> .
`,
		"",
	},
	{
		// [153] #xmlbase-test004
		//
		// xml:base applies to an rdf:ID on a property element.
		//
		"xmlbase/test004.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org/dir/file">

 <rdf:Description>
  <eg:value rdf:ID="frag">v</eg:value>
 </rdf:Description>

</rdf:RDF>`,
		`_:b0 <http://example.org/value> "v" .
<http://example.org/dir/file#frag> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/1999/02/22-rdf-syntax-ns#Statement> .
<http://example.org/dir/file#frag> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> _:b0 .
<http://example.org/dir/file#frag> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> <http://example.org/value> .
<http://example.org/dir/file#frag> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> "v" .
`,
		"",
	},
	{
		// [154] #xmlbase-test006
		//
		// xml:base scoping.
		//
		"xmlbase/test006.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org/dir/file">

 <rdf:Description rdf:ID="frag" eg:value="v" xml:base="http://example.org/file2"/>
 <eg:type rdf:about="relFile" />

</rdf:RDF>`,
		`<http://example.org/file2#frag> <http://example.org/value> "v" .
<http://example.org/dir/relFile> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/type> .
`,
		"",
	},
	{
		// [155] #xmlbase-test007
		//
		// example of relative URI resolution.
		//
		"xmlbase/test007.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org/dir/file">

 <eg:type rdf:about="../relfile" />

</rdf:RDF>`,
		`<http://example.org/relfile> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/type> .
`,
		"",
	},
	{
		// [156] #xmlbase-test008
		//
		// example of empty same document ref resolution.
		//
		"xmlbase/test008.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org/dir/file">

 <eg:type rdf:about="" />

</rdf:RDF>`,
		`<http://example.org/dir/file> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/type> .
`,
		"",
	},
	{
		// [157] #xmlbase-test009
		//
		// Example of relative uri with absolute path resolution.
		//
		"xmlbase/test009.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org/dir/file">

 <eg:type rdf:about="/absfile" />

</rdf:RDF>`,
		`<http://example.org/absfile> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/type> .
`,
		"",
	},
	{
		// [158] #xmlbase-test010
		//
		// Example of relative uri with net path resolution.
		//
		"xmlbase/test010.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org/dir/file">

 <eg:type rdf:about="//another.example.org/absfile" />

</rdf:RDF>`,
		`<http://another.example.org/absfile> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/type> .
`,
		"",
	},
	{
		// [159] #xmlbase-test011
		//
		// Example of xml:base with no path component.
		//
		"xmlbase/test011.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org">

 <eg:type rdf:about="relfile" />

</rdf:RDF>`,
		`<http://example.org/relfile> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/type> .
`,
		"",
	},
	{
		// [160] #xmlbase-test013
		//
		// With an xml:base with fragment the fragment is ignored.
		//
		"xmlbase/test013.rdf",
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:eg="http://example.org/"
         xml:base="http://example.org/dir/file#frag">

 <eg:type rdf:about="" />
 <rdf:Description rdf:ID="foo" >
   <eg:value rdf:resource="relpath" />
 </rdf:Description>

</rdf:RDF>`,
		`<http://example.org/dir/file> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.org/type> .
<http://example.org/dir/file#foo> <http://example.org/value> <http://example.org/dir/relpath> .
`,
		"",
	},
}
