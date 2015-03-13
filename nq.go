package rdf

import "io"

// parseNQ parses a line of N-Quads and returns a valid quad or an error.
func (d *QuadDecoder) parseNQ() (q Quad, err error) {
	defer d.recover(&err)

	for d.peek().typ == tokenEOL {
		d.next()
	}
	if d.peek().typ == tokenEOF {
		return q, io.EOF
	}

	// Set Quad context to default graph
	q.Ctx = d.DefaultGraph

	// parse quad subject
	tok := d.expectAs("subject", tokenIRIAbs, tokenBNode)
	if tok.typ == tokenIRIAbs {
		q.Subj = IRI{str: tok.text}
	} else {
		q.Subj = Blank{id: tok.text}
	}

	// parse quad predicate
	tok = d.expect1As("predicate", tokenIRIAbs)
	q.Pred = IRI{str: tok.text}

	// parse quad object
	tok = d.expectAs("object", tokenIRIAbs, tokenBNode, tokenLiteral)

	switch tok.typ {
	case tokenBNode:
		q.Obj = Blank{id: tok.text}
	case tokenLiteral:
		val := tok.text
		l := Literal{
			str:      val,
			DataType: xsdString,
		}
		p := d.peek()
		switch p.typ {
		case tokenLangMarker:
			d.next() // consume peeked token
			tok = d.expect1As("literal language", tokenLang)
			l.lang = tok.text
			l.DataType = rdfLangString
		case tokenDataTypeMarker:
			d.next() // consume peeked token
			tok = d.expect1As("literal datatype", tokenIRIAbs)
			l.DataType = IRI{str: tok.text}
		}
		q.Obj = l
	case tokenIRIAbs:
		q.Obj = IRI{str: tok.text}
	}

	// parse optional graph
	p := d.peek()
	switch p.typ {
	case tokenIRIAbs:
		tok = d.next() // consume peeked token
		q.Ctx = IRI{str: tok.text}
	case tokenBNode:
		tok = d.next() // consume peeked token
		q.Ctx = Blank{id: tok.text}
	case tokenDot:
		break
	default:
		d.expectAs("graph", tokenIRIAbs, tokenBNode)
	}

	// parse final dot
	d.expect1As("dot (.)", tokenDot)

	// check for extra tokens, assert we reached end of line
	d.expect1As("end of line", tokenEOL)

	if d.peek().typ == tokenEOF {
		// drain lexer
		d.next()
	}
	return q, err
}
