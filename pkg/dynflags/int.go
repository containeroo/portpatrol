package dynflags

import (
	"fmt"
	"strconv"
)

// IntValue implementation for integer flags
type IntValue struct {
	Bound *int
}

func (i *IntValue) Parse(value string) (interface{}, error) {
	return strconv.Atoi(value)
}

func (i *IntValue) Set(value interface{}) error {
	if num, ok := value.(int); ok {
		*i.Bound = num
		return nil
	}
	return fmt.Errorf("invalid value type: expected int")
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (g *ConfigGroup) IntVar(p *int, name string, value int, usage string) {
	*p = *g.Int(name, value, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (g *ConfigGroup) Int(name string, value int, usage string) *int {
	bound := &value
	g.Flags[name] = &Flag{
		Type:    FlagTypeInt,
		Default: value,
		Usage:   usage,
		Value:   &IntValue{Bound: bound},
	}
	g.flagOrder = append(g.flagOrder, name)
	return bound
}

// GetInt returns the int value of a flag with the given name
func (pg *ParsedGroup) GetInt(flagName string) (int, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return 0, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if intVal, ok := value.(int); ok {
		return intVal, nil
	}
	return 0, fmt.Errorf("flag '%s' is not an int", flagName)
}
