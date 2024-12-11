package parsers

import (
	"fmt"
	"strconv"
)

type BoolParser struct{}

func (p *BoolParser) Parse(value string) (interface{}, error) {
	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("invalid boolean value: %v", err)
	}
	return parsedValue, nil
}
