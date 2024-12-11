package parsers

type StringParser struct{}

func (p *StringParser) Parse(value string) (interface{}, error) {
	return value, nil
}

func (p *StringParser) Type() string {
	return "string"
}
