package dynflags

import "fmt"

type StringValue struct {
	Bound *string
}

func (s *StringValue) GetBound() interface{} {
	if s.Bound == nil {
		return nil
	}
	return *s.Bound
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

// String defines a string flag with the specified name, default value, and usage description.
// The flag is added to the group's flag list and returned as a *Flag instance.
func (g *ConfigGroup) String(name, value, usage string) *Flag {
	bound := &value
	flag := &Flag{
		Type:    FlagTypeString,
		Default: value,
		Usage:   usage,
		value:   &StringValue{Bound: bound},
	}
	g.Flags[name] = flag
	g.flagOrder = append(g.flagOrder, name)
	return flag
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
