package dynflags

import (
	"fmt"
	"strings"
)

// StringSlicesValue implementation for string slice flags
type StringSlicesValue struct {
	Bound *[]string
}

func (s *StringSlicesValue) Parse(value string) (interface{}, error) {
	return value, nil
}

func (s *StringSlicesValue) Set(value interface{}) error {
	if str, ok := value.(string); ok {
		*s.Bound = append(*s.Bound, str)
		return nil
	}
	return fmt.Errorf("invalid value type: expected string")
}

// StringSlicesVar defines a string slice flag with specified name, default value, and usage string.
// The argument p points to a slice of strings in which to store the value of the flag.
func (g *ConfigGroup) StringSlicesVar(p *[]string, name string, value []string, usage string) {
	*p = *g.StringSlices(name, value, usage)
}

// StringSlices defines a string slice flag with specified name, default value, and usage string.
// The return value is the address of a slice of strings that stores the value of the flag.
func (g *ConfigGroup) StringSlices(name string, value []string, usage string) *[]string {
	bound := &value
	g.Flags[name] = &Flag{
		Type:    FlagTypeStringSlice,
		Default: strings.Join(value, ","),
		Usage:   usage,
		Value:   &StringSlicesValue{Bound: bound},
	}
	g.flagOrder = append(g.flagOrder, name)
	return bound
}

// GetStringSlices returns the []string value of a flag with the given name
func (pg *ParsedGroup) GetStringSlices(flagName string) ([]string, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if slice, ok := value.([]string); ok {
		return slice, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a []string", flagName)
}
