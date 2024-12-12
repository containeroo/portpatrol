package dynflags

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type FlagType string

const (
	FlagTypeString   FlagType = "STRING"
	FlagTypeInt      FlagType = "INT"
	FlagTypeBool     FlagType = "BOOL"
	FlagTypeDuration FlagType = "DURATION"
	FlagTypeFloat    FlagType = "FLOAT"
	FlagTypeURL      FlagType = "URL"
)

// Flag represents a single configuration flag
type Flag struct {
	Default     interface{} // Default value for the flag
	Type        FlagType    // Type of the flag
	Description string      // Description for usage
	Value       FlagValue   // Encapsulated parsing and value-setting logic
}

// FlagValue interface encapsulates parsing and value-setting logic
type FlagValue interface {
	Parse(value string) (interface{}, error)
	Set(value interface{}) error
}

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

// URLValue implementation for URL flags
type URLValue struct {
	Bound *url.URL
}

func (u *URLValue) Parse(value string) (interface{}, error) {
	return url.Parse(value)
}

func (u *URLValue) Set(value interface{}) error {
	if parsedURL, ok := value.(*url.URL); ok {
		*u.Bound = *parsedURL
		return nil
	}
	return fmt.Errorf("invalid value type: expected URL")
}
