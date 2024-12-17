package dynflags

import (
	"fmt"
	"net/url"
	"strings"
)

// URLSlicesValue implementation for URL slice flags
type URLSlicesValue struct {
	Bound *[]*url.URL
}

func (u *URLSlicesValue) Parse(value string) (interface{}, error) {
	parsedURL, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %s, error: %w", value, err)
	}
	return parsedURL, nil
}

func (u *URLSlicesValue) Set(value interface{}) error {
	if parsedURL, ok := value.(*url.URL); ok {
		*u.Bound = append(*u.Bound, parsedURL)
		return nil
	}
	return fmt.Errorf("invalid value type: expected *url.URL")
}

// URLSlicesVar defines a URL slice flag with specified name, default value, and usage string.
// The argument p points to a slice of URLs in which to store the value of the flag.
func (g *ConfigGroup) URLSlicesVar(p *[]*url.URL, name string, value []*url.URL, usage string) {
	*p = *g.URLSlices(name, value, usage)
}

// URLSlices defines a URL slice flag with specified name, default value, and usage string.
// The return value is the address of a slice of URLs that stores the value of the flag.
func (g *ConfigGroup) URLSlices(name string, value []*url.URL, usage string) *[]*url.URL {
	bound := &value
	defaultValue := make([]string, len(value))
	for i, u := range value {
		defaultValue[i] = u.String()
	}

	g.Flags[name] = &Flag{
		Type:    FlagTypeURLSlice,
		Default: strings.Join(defaultValue, ","),
		Usage:   usage,
		Value:   &URLSlicesValue{Bound: bound},
	}
	g.flagOrder = append(g.flagOrder, name)
	return bound
}

// GetURLSlices returns the []*url.URL value of a flag with the given name
func (pg *ParsedGroup) GetURLSlices(flagName string) ([]*url.URL, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if slice, ok := value.([]*url.URL); ok {
		return slice, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a []*url.URL", flagName)
}
