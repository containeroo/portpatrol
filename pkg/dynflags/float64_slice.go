package dynflags

import (
	"fmt"
	"strconv"
	"strings"
)

// Float64SlicesValue implementation for float64 slice flags
type Float64SlicesValue struct {
	Bound *[]float64
}

func (f *Float64SlicesValue) Parse(value string) (interface{}, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid float64 value: %s, error: %w", value, err)
	}
	return parsed, nil
}

func (f *Float64SlicesValue) Set(value interface{}) error {
	if parsedFloat, ok := value.(float64); ok {
		*f.Bound = append(*f.Bound, parsedFloat)
		return nil
	}
	return fmt.Errorf("invalid value type: expected float64")
}

// Float64SlicesVar defines a float64 slice flag with specified name, default value, and usage string.
// The argument p points to a slice of float64 in which to store the value of the flag.
func (g *ConfigGroup) Float64SlicesVar(p *[]float64, name string, value []float64, usage string) {
	*p = *g.Float64Slices(name, value, usage)
}

// Float64Slices defines a float64 slice flag with specified name, default value, and usage string.
// The return value is the address of a slice of float64 that stores the value of the flag.
func (g *ConfigGroup) Float64Slices(name string, value []float64, usage string) *[]float64 {
	bound := &value
	defaultValue := make([]string, len(value))
	for i, v := range value {
		defaultValue[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}

	g.Flags[name] = &Flag{
		Type:    FlagTypeFloatSlice,
		Default: strings.Join(defaultValue, ","),
		Usage:   usage,
		Value:   &Float64SlicesValue{Bound: bound},
	}
	g.flagOrder = append(g.flagOrder, name)
	return bound
}

// GetFloat64Slices returns the []float64 value of a flag with the given name
func (pg *ParsedGroup) GetFloat64Slices(flagName string) ([]float64, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}

	if slice, ok := value.([]float64); ok {
		return slice, nil
	}

	if f, ok := value.(float64); ok {
		return []float64{f}, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a []float64", flagName)
}
