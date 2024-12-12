package dynflags

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

// URLVar defines a URL flag with specified name, default value, and usage string.
// The argument p points to a url.URL variable in which to store the value of the flag.
func (g *GroupConfig) URLVar(p *url.URL, name, value, usage string) {
	*p = *g.URL(name, value, usage)
}

// URL defines a URL flag with specified name, default value, and usage string.
// The return value is the address of a url.URL variable that stores the value of the flag.
func (g *GroupConfig) URL(name, value, usage string) *url.URL {
	bound := new(url.URL)
	if value != "" {
		parsed, err := url.Parse(value)
		if err != nil {
			panic(fmt.Sprintf("invalid default URL for flag '%s': %s", name, err))
		}
		*bound = *parsed // Copy the parsed URL into bound
	}
	g.Flags[name] = &Flag{
		Type:    FlagTypeURL,
		Default: value,
		Usage:   usage,
		Value:   &URLValue{Bound: bound},
	}
	return bound
}

// GetURL returns the url.URL value of a flag with the given name
func (pg *ParsedGroup) GetURL(flagName string) (*url.URL, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if url, ok := value.(url.URL); ok {
		return &url, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a URL", flagName)
}
