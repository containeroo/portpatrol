package dynflags

import (
	"fmt"
	"strconv"
	"strings"
)

// BoolSlicesValue implementation for bool slice flags
type BoolSlicesValue struct {
	Bound *[]bool
}

func (b *BoolSlicesValue) Parse(value string) (interface{}, error) {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("invalid boolean value: %s, error: %w", value, err)
	}
	return parsed, nil
}

func (b *BoolSlicesValue) Set(value interface{}) error {
	if parsedBool, ok := value.(bool); ok {
		*b.Bound = append(*b.Bound, parsedBool)
		return nil
	}
	return fmt.Errorf("invalid value type: expected bool")
}

// BoolSlicesVar defines a bool slice flag with specified name, default value, and usage string.
// The argument p points to a slice of bool in which to store the value of the flag.
func (g *ConfigGroup) BoolSlicesVar(p *[]bool, name string, value []bool, usage string) {
	*p = *g.BoolSlices(name, value, usage)
}

// BoolSlices defines a bool slice flag with specified name, default value, and usage string.
// The return value is the address of a slice of bool that stores the value of the flag.
func (g *ConfigGroup) BoolSlices(name string, value []bool, usage string) *[]bool {
	bound := &value
	defaultValue := make([]string, len(value))
	for i, v := range value {
		defaultValue[i] = strconv.FormatBool(v)
	}

	g.Flags[name] = &Flag{
		Type:    FlagTypeBoolSlice,
		Default: strings.Join(defaultValue, ","),
		Usage:   usage,
		Value:   &BoolSlicesValue{Bound: bound},
	}
	g.flagOrder = append(g.flagOrder, name)
	return bound
}

// GetBoolSlices returns the []bool value of a flag with the given name
func (pg *ParsedGroup) GetBoolSlices(flagName string) ([]bool, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if slice, ok := value.([]bool); ok {
		return slice, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a []bool", flagName)
}
