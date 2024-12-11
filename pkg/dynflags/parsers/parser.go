package parsers

// Parser interface for parsing values
type Parser interface {
	// Parse parses a value
	Parse(value string) (interface{}, error)
	// Type returns the type of the parser
	Type() string
}
