package dynflags

import (
	"fmt"
	"strings"
)

// StringSlicesValue implementation for string slice flags
type StringSlicesValue struct {
	Bound *[]string
}

func (s *StringSlicesValue) GetBound() interface{} {
	if s.Bound == nil {
		return nil
	}
	return *s.Bound
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

// StringSlices defines a string slice flag with specified name, default value, and usage string.
// The return value is the address of a slice of strings that stores the value of the flag.
func (g *ConfigGroup) StringSlices(name string, value []string, usage string) *Flag {
	bound := &value
	flag := &Flag{
		Type:    FlagTypeStringSlice,
		Default: strings.Join(value, ","),
		Usage:   usage,
		value:   &StringSlicesValue{Bound: bound},
	}
	g.Flags[name] = flag
	g.flagOrder = append(g.flagOrder, name)
	return flag
}

// GetStringSlices returns the []string value of a flag with the given name
func (pg *ParsedGroup) GetStringSlices(flagName string) ([]string, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}

	if strSlice, ok := value.([]string); ok {
		return strSlice, nil
	}

	if str, ok := value.(string); ok {
		return []string{str}, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a []string", flagName)
}
