package types

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
