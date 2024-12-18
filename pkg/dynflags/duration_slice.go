package dynflags

import (
	"fmt"
	"strings"
	"time"
)

type DurationSlicesValue struct {
	Bound *[]time.Duration
}

func (d *DurationSlicesValue) GetBound() interface{} {
	if d.Bound == nil {
		return nil
	}
	return *d.Bound
}

func (d *DurationSlicesValue) Parse(value string) (interface{}, error) {
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return nil, fmt.Errorf("invalid duration value: %s, error: %w", value, err)
	}
	return parsed, nil
}

func (d *DurationSlicesValue) Set(value interface{}) error {
	if parsedDuration, ok := value.(time.Duration); ok {
		*d.Bound = append(*d.Bound, parsedDuration)
		return nil
	}
	return fmt.Errorf("invalid value type: expected time.Duration")
}

// DurationSlices defines a duration slice flag with the specified name, default value, and usage description.
// The flag is added to the group's flag list and returned as a *Flag instance.
func (g *ConfigGroup) DurationSlices(name string, value []time.Duration, usage string) *Flag {
	bound := &value
	defaultValue := make([]string, len(value))
	for i, v := range value {
		defaultValue[i] = v.String()
	}

	flag := &Flag{
		Type:    FlagTypeDurationSlice,
		Default: strings.Join(defaultValue, ","),
		Usage:   usage,
		value:   &DurationSlicesValue{Bound: bound},
	}
	g.Flags[name] = flag
	g.flagOrder = append(g.flagOrder, name)
	return flag
}

// GetDurationSlices returns the []time.Duration value of a flag with the given name
func (pg *ParsedGroup) GetDurationSlices(flagName string) ([]time.Duration, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}

	if slice, ok := value.([]time.Duration); ok {
		return slice, nil
	}

	if d, ok := value.(time.Duration); ok {
		return []time.Duration{d}, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a []time.Duration", flagName)
}
