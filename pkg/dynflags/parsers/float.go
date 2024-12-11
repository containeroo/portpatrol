package parsers

import (
	"fmt"
	"strconv"
)

type FloatParser struct{}

func (p *FloatParser) Parse(value string) (interface{}, error) {
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid float value: %v", err)
	}
	return parsedValue, nil
}
