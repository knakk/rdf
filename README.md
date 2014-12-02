# rdf

This package introduces data structures for working with RDF in Go.

The main use case is to represent data coming from or going to a triple/quad-store via the SPARQL protocol. It does not provide any means of working with the data as a graph, i.e no in-memory graph traversal or querying, as this is better handeled by a dedicated triple store.

For complete documentation see [godoc](:http://godoc.org/github.com/knakk/rdf).

## Types and interfaces

_Note_: What is named `URI` in this library, is actually a `IRI reference` (a generalization of `URI` supporting the full unicode range, except the characters `<>'"{}|^\`' and space). The choice was made because `URI` is so widely used.

The `Term` interface represents a RDF term, and is implemented by the three differend kinds: URIs, Literals and Blank nodes.

```go
type Term interface {
	// String should return the string representation of a RDF term, in a
	// form suitable for insertion into a SPARQL query.
	String() string

	// Value returns the typed value of a RDF term, boxed in an empty interface.
	// For URIs and Blank nodes this would return the uri and blank label as strings.
	Value() interface{}

	// Eq tests for equality with another RDF term.
	Eq(other Term) bool

	// Type returns the RDF term type.
	Type() TermType
}
```

The 3 term types:

```
// Blank represents a RDF blank node; an unqualified URI with an ID.
type Blank struct {
	ID string
}
```

```go
// URI represents a RDF URI resource.
type URI struct {
	URI string
}
```

```go
// Literal represents a RDF literal; a value with a datatype and
// (optionally) an associated language tag for strings.
type Literal struct {
	// Val represents the typed value of a RDF Literal, boxed in an empty interface.
	// A type assertion is needed to get the value in the corresponding Go type.
	Val interface{}

	// Lang, if not empty, represents the language tag of a string.
	Lang string

	// The datatype of the Literal.
	DataType *URI
}
```

A RDF triple is composed of 3 terms forming the subject, predicate and object of a statement.

```go
type Triple struct {
	Subj, Pred, Obj Term
}
```

A quad is a triple with a named graph:
```go
type Quad struct {
	Subj  Term
	Pred  Term
	Obj   Term
	Graph URI
}
```

## Usage

This package exist mainly to declare datastructures for working with RDF data. The only functions provided are constructors for creating structs of the different RDF terms.

There are two constructors for each term type - one which validates that input conforms to the RDF standards, and another with the suffix `Unsafe` which doesn't:

```go
u, err := NewURI("an invalid uri")
if err != nil {
  // handle error on invalid input
}

u2 := NewURIUnsafe("http://my.resource/nr/123") // no validation on input
```
