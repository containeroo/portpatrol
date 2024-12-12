package types

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
