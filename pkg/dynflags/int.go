package dynflags

import (
	"fmt"
	"strconv"
)

type IntValue struct {
	Bound *int
}

func (i *IntValue) GetBound() interface{} {
	if i.Bound == nil {
		return nil
	}
	return *i.Bound
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

// Int defines an integer flag with the specified name, default value, and usage description.
// The flag is added to the group's flag list and returned as a *Flag instance.
func (g *ConfigGroup) Int(name string, value int, usage string) *Flag {
	bound := &value
	flag := &Flag{
		Type:    FlagTypeInt,
		Default: value,
		Usage:   usage,
		value:   &IntValue{Bound: bound},
	}
	g.Flags[name] = flag
	g.flagOrder = append(g.flagOrder, name)
	return flag
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
