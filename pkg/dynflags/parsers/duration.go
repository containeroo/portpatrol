package parsers

import (
	"fmt"
	"time"
)

type DurationParser struct{}

func (p *DurationParser) Parse(value string) (interface{}, error) {
	parsedValue, err := time.ParseDuration(value)
	if err != nil {
		return nil, fmt.Errorf("invalid duration value: %v", err)
	}
	return parsedValue, nil
}
