package dynflags

import (
	"fmt"
	"strconv"
)

// BoolValue implementation for boolean flags
type BoolValue struct {
	Bound *bool
}

func (b *BoolValue) GetBound() interface{} {
	if b.Bound == nil {
		return nil
	}
	return *b.Bound
}

func (b *BoolValue) Parse(value string) (interface{}, error) {
	return strconv.ParseBool(value)
}

func (b *BoolValue) Set(value interface{}) error {
	if val, ok := value.(bool); ok {
		*b.Bound = val
		return nil
	}
	return fmt.Errorf("invalid value type: expected bool")
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func (g *ConfigGroup) Bool(name string, value bool, usage string) *Flag {
	bound := &value
	flag := &Flag{
		Type:    FlagTypeBool,
		Default: value,
		Usage:   usage,
		value:   &BoolValue{Bound: bound},
	}
	g.Flags[name] = flag
	g.flagOrder = append(g.flagOrder, name)
	return flag
}

// GetBool returns the bool value of a flag with the given name
func (pg *ParsedGroup) GetBool(flagName string) (bool, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return false, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if boolVal, ok := value.(bool); ok {
		return boolVal, nil
	}
	return false, fmt.Errorf("flag '%s' is not a bool", flagName)
}
