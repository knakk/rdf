// Package xsd exports IRIs of xsd datatypes.
package xsd

import "github.com/knakk/rdf"

// The XML schema built-in datatypes (xsd):
// https://dvcs.w3.org/hg/rdf/raw-file/default/rdf-concepts/index.html#xsd-datatypes
var (
	// Core types:

	String  = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#string"}
	Boolean = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#boolean"}
	Decimal = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#decimal"}
	Integer = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#integer"}

	// IEEE floating-point numbers:

	Double = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#double"}
	Float  = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#float"}

	// Time and date:

	Date          = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#date"}
	Time          = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#time"}
	DateTime      = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#dateTime"}
	DateTimeStamp = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#dateTimeStamp"}

	// Recurring and partial dates:

	Year              = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#gYear"}
	Month             = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#gMonth"}
	Day               = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#gDay"}
	YearMonth         = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#gYearMonth"}
	Duration          = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#Duration"}
	YearMonthDuration = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#yearMonthDuration"}
	DayTimeDuration   = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#dayTimeDuration"}

	// Limited-range integer numbers

	Byte = rdf.IRI{IRI: "http://www.w3.org/2001/XMLSchema#byte"}
)
