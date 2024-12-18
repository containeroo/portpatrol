package dynflags

import (
	"fmt"
	"net/url"
)

type URLValue struct {
	Bound *url.URL
}

func (u *URLValue) GetBound() interface{} {
	if u.Bound == nil {
		return nil
	}
	return *u.Bound
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

// URL defines a URL flag with the specified name, default value, and usage description.
// The flag is added to the group's flag list and returned as a *Flag instance.
func (g *ConfigGroup) URL(name, value, usage string) *Flag {
	bound := new(url.URL)
	if value != "" {
		parsed, err := url.Parse(value)
		if err != nil {
			panic(fmt.Sprintf("invalid default URL for flag '%s': %s", name, err))
		}
		*bound = *parsed // Copy the parsed URL into bound
	}
	flag := &Flag{
		Type:    FlagTypeURL,
		Default: value,
		Usage:   usage,
		value:   &URLValue{Bound: bound},
	}
	g.Flags[name] = flag
	g.flagOrder = append(g.flagOrder, name)
	return flag
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
