package types

import (
	"fmt"
	"net/url"
)

// URLValue implementation for URL flags
type URLValue struct {
	Bound *url.URL
}

func (u *URLValue) Parse(value string) (interface{}, error) {
	return url.Parse(value)
}

func (u *URLValue) Set(value interface{}) error {
	if parsedURL, ok := value.(*url.URL); ok {
		*u.Bound = *parsedURL
		return nil
	}
	return fmt.Errorf("invalid value type: expected URL")
}
