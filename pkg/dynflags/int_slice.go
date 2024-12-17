package dynflags

import (
	"fmt"
	"strconv"
	"strings"
)

// IntSlicesValue implementation for int slice flags
type IntSlicesValue struct {
	Bound *[]int
}

func (s *IntSlicesValue) Parse(value string) (interface{}, error) {
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("invalid integer value: %s", value)
	}
	return parsedValue, nil
}

func (s *IntSlicesValue) Set(value interface{}) error {
	if num, ok := value.(int); ok {
		*s.Bound = append(*s.Bound, num)
		return nil
	}
	return fmt.Errorf("invalid value type: expected int")
}

// IntSlicesVar defines an int slice flag with specified name, default value, and usage string.
// The argument p points to a slice of integers in which to store the value of the flag.
func (g *ConfigGroup) IntSlicesVar(p *[]int, name string, value []int, usage string) {
	*p = *g.IntSlices(name, value, usage)
}

// IntSlices defines an int slice flag with specified name, default value, and usage string.
// The return value is the address of a slice of integers that stores the value of the flag.
func (g *ConfigGroup) IntSlices(name string, value []int, usage string) *[]int {
	bound := &value
	defaults := make([]string, len(value))
	for i, v := range value {
		defaults[i] = strconv.Itoa(v)
	}
	g.Flags[name] = &Flag{
		Type:    FlagTypeIntSlice,
		Default: strings.Join(defaults, ","),
		Usage:   usage,
		Value:   &IntSlicesValue{Bound: bound},
	}
	g.flagOrder = append(g.flagOrder, name)
	return bound
}

// GetIntSlices returns the []int value of a flag with the given name
func (pg *ParsedGroup) GetIntSlices(flagName string) ([]int, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}

	if slice, ok := value.([]int); ok {
		return slice, nil
	}

	if i, ok := value.(int); ok {
		return []int{i}, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a []int", flagName)
}
