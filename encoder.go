package rdf

import (
	"bufio"
	"errors"
	"io"
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
	format  Format            // Serialization format.
	w       *bufio.Writer     // Buffered writer. Set to nil when Encoder is closed.
	ns      map[string]string // IRI->prefix mappings.
	curSubj Subject           // Keep track of current subject, to enable encoding of predicate lists.
	curPred Predicate         // Keep track of current subject, to enable encoding of object list.
}

// NewTripleEncoder returns a new TripleEncoder capable of serializing into the
// given io.Writer in the given serialization format.
func NewTripleEncoder(w io.Writer, f Format) *TripleEncoder {
	return &TripleEncoder{
		format: f,
		w:      bufio.NewWriter(w),
		ns:     make(map[string]string),
	}
}

// Encode serializes a single Triple to the io.Writer of the TripleEncoder.
func (e *TripleEncoder) Encode(t Triple) error {
	if e.w == nil {
		return ErrEncoderClosed
	}
	switch e.format {
	case FormatNT:
		_, err := e.w.Write([]byte(t.Serialize(e.format)))
		if err != nil {
			return err
		}
	default:
		panic("TODO")
	}
	return nil
}

// EncodeAll serializes a slice of Triples to the io.Writer of the TripleEncoder.
func (e *TripleEncoder) EncodeAll(ts []Triple) error {
	if e.w == nil {
		return ErrEncoderClosed
	}
	switch e.format {
	case FormatNT:
		for _, t := range ts {
			_, err := e.w.Write([]byte(t.Serialize(e.format)))
			if err != nil {
				return err
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
	err := e.w.Flush()
	e.w = nil
	return err
}

/*
func (e *QuadEncoder) Encode(q Quad) err {
	return nil
}
func (e *QuadEncoder) EncodeAll(qs []Quad) err {
	return nil
}

func (e *QuadEncoder) Close() err {
	return nil
}
*/
