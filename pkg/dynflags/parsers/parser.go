package parsers

// Parser interface for parsing values
type Parser interface {
	Parse(value string) (interface{}, error)
}
