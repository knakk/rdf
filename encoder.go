package rdf

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
)

// ErrEncoderClosed is the error returned from Encode() when the Triple/Quad-Encoder is closed
var ErrEncoderClosed = errors.New("Encoder is closed and cannot encode anymore")

// TripleEncoder serializes RDF Triples into one of the following formats:
// N-Triples, Turtle, RDF/XML.
//
// For streaming serialization, use the Encode() method to encode a single Triple
// at a time. Or, if you want to encode multiple triples in one batch, use EncodeAll().
// In either case; when done serializing, Close() must be called, to ensure
// that all writes are persisted, since the Encoder uses buffered IO.
type TripleEncoder struct {
	format             Format            // Serialization format.
	w                  *errWriter        // Buffered writer. Set to nil when Encoder is closed.
	Namespaces         map[string]string // IRI->prefix custom mappings.
	ns                 map[string]string // IRI->prefix mappings.
	nsCount            int               // Counter to generate unique namespace prefixes
	curSubj            Subject           // Keep track of current subject, to enable encoding of predicate lists.
	curPred            Predicate         // Keep track of current subject, to enable encoding of object list.
	OpenStatement      bool              // True when triple statement hasn't been closed (i.e. in a predicate/object list)
	GenerateNamespaces bool              // True to auto generate namespaces, false if you give it some custom namespaces and do not want generated ones
}

// NewTripleEncoder returns a new TripleEncoder capable of serializing into the
// given io.Writer in the given serialization format.
func NewTripleEncoder(w io.Writer, f Format) *TripleEncoder {
	return &TripleEncoder{
		format:             f,
		w:                  &errWriter{w: bufio.NewWriter(w)},
		Namespaces:         make(map[string]string),
		ns:                 make(map[string]string),
		GenerateNamespaces: true,
	}
}

// Encode serializes a single Triple to the io.Writer of the TripleEncoder.
func (e *TripleEncoder) Encode(t Triple) error {
	if e.w == nil {
		return ErrEncoderClosed
	}
	switch e.format {
	case NTriples:
		_, err := e.w.w.Write([]byte(t.Serialize(e.format)))
		if err != nil {
			return err
		}
	case Turtle:
		var s, p, o string

		// object is allways rendered the same
		o = e.prefixify(t.Obj)

		if e.OpenStatement {
			// potentially predicate/object list
			// curSubj and curPred is set
			if TermsEqual(e.curSubj, t.Subj) {
				// In predicate or object list
				if TermsEqual(e.curPred, t.Pred) {
					// in object list
					s = " ,\n\t"
					p = ""
				} else {
					// in predicate list
					p = e.prefixify(t.Pred)

					// check if predicate introduced new prefix directive
					if e.OpenStatement {
						// in predicate list
						s = " ;\n"
						e.curPred = t.Pred
					} else {
						// previous statement closed
						e.curSubj = t.Subj
						s = e.prefixify(t.Subj)
						e.curPred = t.Pred
					}
				}
			} else {
				// not in predicate/ojbect list
				// close previous statement
				e.w.write([]byte(" .\n"))
				e.OpenStatement = false
				p = e.prefixify(t.Pred)
				e.curSubj = t.Subj
				s = e.prefixify(t.Subj)
				e.curPred = t.Pred
			}
		} else {
			// either first statement, or after a prefix directive
			p = e.prefixify(t.Pred)
			s = e.prefixify(t.Subj)
			e.curSubj = t.Subj
			e.curPred = t.Pred
		}

		// allways keep statement open, in case next triple can mean predicate/object list
		e.OpenStatement = true

		e.w.write([]byte(s))
		e.w.write([]byte("\t"))
		e.w.write([]byte(p))
		e.w.write([]byte("\t"))
		e.w.write([]byte(o))

		if e.w.err != nil {
			return e.w.err
		}
	default:
		panic("TODO")
	}
	return nil
}

// EncodeAll serializes a slice of Triples to the io.Writer of the TripleEncoder.
// It will ignore duplicate triples.
//
// Note that this function will modify the given slice of triples by sorting it in-place.
func (e *TripleEncoder) EncodeAll(ts []Triple) error {
	if e.w == nil {
		return ErrEncoderClosed
	}
	switch e.format {
	case NTriples:
		for _, t := range ts {
			_, err := e.w.w.Write([]byte(t.Serialize(e.format)))
			if err != nil {
				return err
			}
		}
	case Turtle:
		// Sort triples by Subject, then Predicate, to maximize predicate and object lists.
		sort.Sort(bySubjectThenPred(triples(ts)))

		var s, p, o string

		for i, t := range ts {
			// object is allways rendered the same
			o = e.prefixify(t.Obj)

			if e.OpenStatement {
				// potentially predicate/object list
				// curSubj and curPred is set
				if TermsEqual(e.curSubj, t.Subj) {
					// In predicate or object list
					if TermsEqual(e.curPred, t.Pred) {
						// in object list

						// check if this triple is a duplicate of the preceeding triple
						if i > 0 && TermsEqual(t.Obj, ts[i-1].Obj) {
							continue
						}

						s = " ,\n\t"
						p = ""
					} else {
						// in predicate list
						p = e.prefixify(t.Pred)

						// check if predicate introduced new prefix directive
						if e.OpenStatement {
							// in predicate list
							s = " ;\n"
							e.curPred = t.Pred
						} else {
							// previous statement closed
							e.curSubj = t.Subj
							s = e.prefixify(t.Subj)
							e.curPred = t.Pred
						}
					}
				} else {
					// not in predicate/ojbect list
					// close previous statement
					e.w.write([]byte(" .\n"))
					e.OpenStatement = false
					p = e.prefixify(t.Pred)
					e.curSubj = t.Subj
					s = e.prefixify(t.Subj)
					e.curPred = t.Pred
				}
			} else {
				// either first statement, or after a prefix directive
				p = e.prefixify(t.Pred)
				s = e.prefixify(t.Subj)
				e.curSubj = t.Subj
				e.curPred = t.Pred
			}

			// allways keep statement open, in case next triple can mean predicate/object list
			e.OpenStatement = true

			e.w.write([]byte(s))
			e.w.write([]byte("\t"))
			e.w.write([]byte(p))
			e.w.write([]byte("\t"))
			e.w.write([]byte(o))

			if e.w.err != nil {
				return e.w.err
			}
		}
	default:
		panic("TODO")
	}
	return nil
}

// Close finalizes an encoding session, ensuring that any concluding tokens are
// written should it be needed (eg.g close the root tag for RDF/XML) and
// flushes the underlying buffered writer of the encoder.
//
// The encoder cannot encode anymore when Close() has been called.
func (e *TripleEncoder) Close() error {
	if e.OpenStatement {
		e.w.write([]byte(" .")) // Close final statement
		if e.w.err != nil {
			return e.w.err
		}
	}
	err := e.w.w.Flush()
	e.w = nil
	return err
}

func (e *TripleEncoder) prefixify(t Term) string {
	if t.Type() == TermIRI {
		if t.(IRI).str == "http://www.w3.org/1999/02/22-rdf-syntax-ns#type" {
			return "a"
		}
		first, rest := t.(IRI).Split()
		if first == "" {
			// cannot split into prefix and namespace
			return t.Serialize(Turtle)
		}

		prefix, ok := e.ns[first]
		if !ok {
			custom, ok := e.Namespaces[first]
			if ok {
				//we have a custom namespace specified, use that
				prefix = custom
			} else {
				if !e.GenerateNamespaces {
					return t.Serialize(Turtle)
				}
				prefix = fmt.Sprintf("ns%d", e.nsCount)
				e.nsCount++
			}
			e.ns[first] = prefix
			if e.OpenStatement {
				e.w.write([]byte(" .\n"))
			}
			e.w.write([]byte(fmt.Sprintf("@prefix %s:\t<%s> .\n", prefix, first)))
			e.OpenStatement = false
		}
		return fmt.Sprintf("%s:%s", prefix, rest)
	}
	if t.Type() == TermLiteral {
		switch t.(Literal).DataType {
		case xsdString, xsdInteger, xsdBoolean, xsdDouble, xsdDecimal, rdfLangString:
			// serialize normally in Literal.Serialize method
			break
		default:
			first, rest := t.(Literal).DataType.Split()
			if first == "" {
				return t.Serialize(Turtle)
			}

			prefix, ok := e.ns[first]
			if !ok {
				custom, ok := e.Namespaces[first]
				if ok {
					//we have a custom namespace specified, use that
					prefix = custom
				} else {
					if !e.GenerateNamespaces {
						return t.Serialize(Turtle)
					}
					prefix = fmt.Sprintf("ns%d", e.nsCount)
					e.nsCount++
				}
				e.ns[first] = prefix
				if e.OpenStatement {
					e.w.write([]byte(" .\n"))
				}
				e.w.write([]byte(fmt.Sprintf("@prefix %s:\t<%s> .\n", prefix, first)))
				e.OpenStatement = false
			}
			return fmt.Sprintf("\"%s\"^^%s:%s", t.Serialize(formatInternal), prefix, rest)
		}
	}
	return t.Serialize(Turtle)
}

func escapeLocal(rest string) string {
	// escape rest according to PN_LOCAL
	// http://www.w3.org/TR/turtle/#reserved
	var b bytes.Buffer
	for _, r := range rest {
		if int(r) <= 126 && int(r) >= 33 {
			// only bother to check if rune is in range
			switch r {
			case '_', '~', '.', '-', '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', '/', '?', '#', '@', '%':
				b.WriteRune('\\')
				b.WriteRune(r)
			default:
				b.WriteRune(r)
			}
		} else {
			b.WriteRune(r)
		}
	}
	// TODO should also ensure that last character is not '.'
	return b.String()
}

type triples []Triple

type bySubjectThenPred triples

func (t bySubjectThenPred) Len() int {
	return len(t)
}

func (t bySubjectThenPred) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t bySubjectThenPred) Less(i, j int) bool {
	// todo implement custom comparestring function wich returns -1 0 1 for less, equal, greater
	// https://groups.google.com/forum/#!topic/golang-nuts/5mMdKvkxWxo
	// see also bytes.Compare
	p, q := t[i].Subj.Serialize(NTriples), t[j].Subj.Serialize(NTriples)
	switch {
	case p < q:
		return true
	case q < p:
		return false
	default:
		// subjects are equal, continue by comparing predicates
		return t[i].Pred.Serialize(NTriples) < t[j].Pred.Serialize(NTriples)
	}
}

type errWriter struct {
	w   *bufio.Writer
	err error
}

func (ew *errWriter) write(buf []byte) {
	if ew.err != nil {
		return
	}
	_, ew.err = ew.w.Write(buf)
}

// QuadEncoder serializes RDF Quads. Currently only supports N-Quads.
type QuadEncoder struct {
	w *errWriter
}

// NewQuadEncoder returns a new QuadEncoder on the given writer. The only supported
// format is NQuads.
func NewQuadEncoder(w io.Writer, f Format) *QuadEncoder {
	if f != NQuads {
		panic("NewQuadEncoder: only N-Quads format supported ATM")
	}
	return &QuadEncoder{
		w: &errWriter{w: bufio.NewWriter(w)},
	}
}

// Encode encodes a Quad.
func (e *QuadEncoder) Encode(q Quad) error {
	_, err := e.w.w.Write([]byte(q.Serialize(NQuads)))
	if err != nil {
		return err
	}
	return nil
}

// EncodeAll encodes all quads.
func (e *QuadEncoder) EncodeAll(qs []Quad) error {
	if e.w == nil {
		return ErrEncoderClosed
	}
	for _, q := range qs {
		_, err := e.w.w.Write([]byte(q.Serialize(NQuads)))
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the encoder and flushes the underlying buffering writer.
func (e *QuadEncoder) Close() error {
	err := e.w.w.Flush()
	e.w = nil
	return err
}
