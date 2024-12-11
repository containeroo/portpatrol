package parsers

import (
	"errors"
	"fmt"
	"net/url"
)

type URLParser struct{}

func (p *URLParser) Parse(value string) (interface{}, error) {
	parsedValue, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid URL value: %v", err)
	}
	if parsedValue.Scheme == "" || parsedValue.Host == "" {
		return nil, errors.New("URL must include scheme and host")
	}
	return parsedValue, nil
}

func (p *URLParser) Type() string {
	return "url"
}
