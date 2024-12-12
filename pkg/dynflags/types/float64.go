package types

import (
	"fmt"
	"strconv"
)

// IntValue implementation for integer flags
type Float64Value struct {
	Bound *float64
}

func (i *Float64Value) Parse(value string) (interface{}, error) {
	return strconv.Atoi(value)
}

func (i *Float64Value) Set(value interface{}) error {
	if num, ok := value.(float64); ok {
		*i.Bound = num
		return nil
	}
	return fmt.Errorf("invalid value type: expected float64")
}
