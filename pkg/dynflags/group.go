package dynflags

import (
	"net/url"
	"time"

	"github.com/containeroo/portpatrol/pkg/dynflags/parsers"
)

// GroupConfig represents the static configuration for a group
type GroupConfig struct {
	Name  string           // Name of the group
	Flags map[string]*Flag // Flags within the group
}

// ParsedGroup represents a runtime group with parsed values
type ParsedGroup struct {
	Parent *GroupConfig           // Reference to the parent static group
	Name   string                 // Identifier for the child group
	Values map[string]interface{} // Parsed values for the group's flags
}

// String registers a string flag in the group and binds it to the provided variable
func (g *GroupConfig) String(flagName, defaultValue, description string) *string {
	parsedValue := new(string)
	*parsedValue = defaultValue // Initialize with the default value

	flag := &Flag{
		Type:        "string",
		Default:     defaultValue,
		Description: description,
		Value:       parsedValue, // Store the pointer in the flag
		Parser:      &parsers.StringParser{},
	}
	g.Flags[flagName] = flag
	return parsedValue // Return the pointer
}

// Int registers an integer flag in the group
func (g *GroupConfig) Int(flagName string, defaultValue int, description string) *int {
	parsedValue := new(int)
	*parsedValue = defaultValue

	flag := &Flag{
		Type:        "int",
		Default:     defaultValue,
		Description: description,
		Value:       parsedValue,
		Parser:      &parsers.IntParser{},
	}
	g.Flags[flagName] = flag
	return parsedValue
}

// Bool registers a boolean flag in the group
func (g *GroupConfig) Bool(flagName string, defaultValue bool, description string) *bool {
	parsedValue := new(bool)
	*parsedValue = defaultValue

	flag := &Flag{
		Type:        "bool",
		Default:     defaultValue,
		Description: description,
		Value:       parsedValue,
		Parser:      &parsers.BoolParser{},
	}
	g.Flags[flagName] = flag
	return parsedValue
}

// Duration registers a duration flag in the group
func (g *GroupConfig) Duration(flagName string, defaultValue time.Duration, description string) *time.Duration {
	parsedValue := new(time.Duration)
	*parsedValue = defaultValue

	flag := &Flag{
		Type:        "duration",
		Default:     defaultValue,
		Description: description,
		Value:       defaultValue,
		Parser:      &parsers.DurationParser{},
	}
	g.Flags[flagName] = flag
	return parsedValue
}

// Float registers a float flag in the group
func (g *GroupConfig) Float(flagName string, defaultValue float64, description string) *float64 {
	parsedValue := new(float64)
	*parsedValue = defaultValue

	flag := &Flag{
		Type:        "float",
		Default:     defaultValue,
		Description: description,
		Value:       defaultValue,
		Parser:      &parsers.FloatParser{},
	}
	g.Flags[flagName] = flag
	return parsedValue
}

// URL registers a URL flag in the group and binds it to a `url.URL` pointer
func (g *GroupConfig) URL(flagName, defaultValue, description string) *url.URL {
	parsedValue := new(url.URL)

	if defaultValue != "" {
		parsedURL, err := url.Parse(defaultValue)
		if err == nil {
			*parsedValue = *parsedURL
		}
	}

	flag := &Flag{
		Type:        "url",
		Default:     defaultValue,
		Description: description,
		Value:       parsedValue,
		Parser:      &parsers.URLParser{},
	}
	g.Flags[flagName] = flag
	return parsedValue
}
