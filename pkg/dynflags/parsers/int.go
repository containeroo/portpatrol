package parsers

import (
	"fmt"
	"strconv"
)

type IntParser struct{}

func (p *IntParser) Parse(value string) (interface{}, error) {
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("invalid integer value: %v", err)
	}
	return parsedValue, nil
}

func (p *IntParser) Type() string {
	return "int"
}
