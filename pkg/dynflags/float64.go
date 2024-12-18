package dynflags

import (
	"fmt"
	"strconv"
)

type Float64Value struct {
	Bound *float64
}

func (f *Float64Value) GetBound() interface{} {
	if f.Bound == nil {
		return nil
	}
	return *f.Bound
}

func (i *Float64Value) Parse(value string) (interface{}, error) {
	return strconv.ParseFloat(value, 64)
}

func (i *Float64Value) Set(value interface{}) error {
	if num, ok := value.(float64); ok {
		*i.Bound = num
		return nil
	}
	return fmt.Errorf("invalid value type: expected float64")
}

// Float64 defines a float64 flag with the specified name, default value, and usage description.
// The flag is added to the group's flag list and returned as a *Flag instance.
func (g *ConfigGroup) Float64(name string, value float64, usage string) *Flag {
	bound := &value
	flag := &Flag{
		Type:    FlagTypeInt,
		Default: value,
		Usage:   usage,
		value:   &Float64Value{Bound: bound},
	}
	g.Flags[name] = flag
	g.flagOrder = append(g.flagOrder, name)
	return flag
}

// GetFloat64 returns the float64 value of a flag with the given name
func (pg *ParsedGroup) GetFloat64(flagName string) (float64, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return 0, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if floatVal, ok := value.(float64); ok {
		return floatVal, nil
	}
	return 0, fmt.Errorf("flag '%s' is not a float64", flagName)
}
