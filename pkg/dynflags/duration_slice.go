package dynflags

import (
	"fmt"
	"strings"
	"time"
)

// DurationSlicesValue implementation for duration slice flags
type DurationSlicesValue struct {
	Bound *[]time.Duration
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

// DurationSlicesVar defines a duration slice flag with specified name, default value, and usage string.
// The argument p points to a slice of durations in which to store the value of the flag.
func (g *GroupConfig) DurationSlicesVar(p *[]time.Duration, name string, value []time.Duration, usage string) {
	*p = *g.DurationSlices(name, value, usage)
}

// DurationSlices defines a duration slice flag with specified name, default value, and usage string.
// The return value is the address of a slice of durations that stores the value of the flag.
func (g *GroupConfig) DurationSlices(name string, value []time.Duration, usage string) *[]time.Duration {
	bound := &value
	defaultValue := make([]string, len(value))
	for i, v := range value {
		defaultValue[i] = v.String()
	}

	g.Flags[name] = &Flag{
		Type:    FlagTypeDurationSlice,
		Default: strings.Join(defaultValue, ","),
		Usage:   usage,
		Value:   &DurationSlicesValue{Bound: bound},
	}
	g.flagOrder = append(g.flagOrder, name)
	return bound
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

	return nil, fmt.Errorf("flag '%s' is not a []time.Duration", flagName)
}
