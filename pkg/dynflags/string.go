package dynflags

import "fmt"

// StringValue implementation for string flags
type StringValue struct {
	Bound *string
}

func (s *StringValue) Parse(value string) (interface{}, error) {
	return value, nil
}

func (s *StringValue) Set(value interface{}) error {
	if str, ok := value.(string); ok {
		*s.Bound = str
		return nil
	}
	return fmt.Errorf("invalid value type: expected string")
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (g *GroupConfig) StringVar(p *string, name, value, usage string) {
	*p = *g.String(name, value, usage)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (g *GroupConfig) String(name, value, usage string) *string {
	bound := &value
	g.Flags[name] = &Flag{
		Type:    FlagTypeString,
		Default: value,
		Usage:   usage,
		Value:   &StringValue{Bound: bound},
	}
	return bound
}

// GetString returns the string value of a flag with the given name
func (pg *ParsedGroup) GetString(flagName string) (string, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return "", fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("flag '%s' is not a string", flagName)
}
