package types

import (
	"fmt"
	"strconv"
)

// BoolValue implementation for boolean flags
type BoolValue struct {
	Bound *bool
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
