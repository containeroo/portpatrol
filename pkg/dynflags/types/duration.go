package types

import (
	"time"
	"fmt"
)

// DurationValue implementation for duration flags
type DurationValue struct {
	Bound *time.Duration
}

func (d *DurationValue) Parse(value string) (interface{}, error) {
	return time.ParseDuration(value)
}

func (d *DurationValue) Set(value interface{}) error {
	if dur, ok := value.(time.Duration); ok {
		*d.Bound = dur
		return nil
	}
	return fmt.Errorf("invalid value type: expected duration")
}
