package dynflags

import (
	"fmt"
	"strconv"
	"strings"
)

type Float64SlicesValue struct {
	Bound *[]float64
}

func (f *Float64SlicesValue) GetBound() interface{} {
	if f.Bound == nil {
		return nil
	}
	return *f.Bound
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

// Float64Slices defines a float64 slice flag with the specified name, default value, and usage description.
// The flag is added to the group's flag list and returned as a *Flag instance.
func (g *ConfigGroup) Float64Slices(name string, value []float64, usage string) *Flag {
	bound := &value
	defaultValue := make([]string, len(value))
	for i, v := range value {
		defaultValue[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}

	flag := &Flag{
		Type:    FlagTypeFloatSlice,
		Default: strings.Join(defaultValue, ","),
		Usage:   usage,
		value:   &Float64SlicesValue{Bound: bound},
	}
	g.Flags[name] = flag
	g.flagOrder = append(g.flagOrder, name)
	return flag
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
