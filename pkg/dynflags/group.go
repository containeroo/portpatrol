package dynflags

import (
	"fmt"
	"net/url"
	"time"

	"github.com/containeroo/portpatrol/pkg/dynflags/types"
)

// GroupConfig represents the static configuration for a group
type GroupConfig struct {
	Name  string           // Name of the group
	Flags map[string]*Flag // Flags within the group
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
		Value:   &types.StringValue{Bound: bound},
	}
	return bound
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (g *GroupConfig) IntVar(p *int, name string, value int, usage string) {
	*p = *g.Int(name, value, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (g *GroupConfig) Int(name string, value int, usage string) *int {
	bound := &value
	g.Flags[name] = &Flag{
		Type:    FlagTypeInt,
		Default: value,
		Usage:   usage,
		Value:   &types.IntValue{Bound: bound},
	}
	return bound
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (g *GroupConfig) Float64Var(p *float64, name string, value float64, usage string) {
	*p = *g.Float64(name, value, usage)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func (g *GroupConfig) Float64(name string, value float64, usage string) *float64 {
	bound := &value
	g.Flags[name] = &Flag{
		Type:    FlagTypeInt,
		Default: value,
		Usage:   usage,
		Value:   &types.Float64Value{Bound: bound},
	}
	return bound
}

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func (g *GroupConfig) BoolVar(p *bool, name string, value bool, usage string) {
	*p = *g.Bool(name, value, usage)
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func (g *GroupConfig) Bool(name string, value bool, usage string) *bool {
	bound := &value
	g.Flags[name] = &Flag{
		Type:    FlagTypeBool,
		Default: value,
		Usage:   usage,
		Value:   &types.BoolValue{Bound: bound},
	}
	return bound
}

// DurationVar defines a duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func (g *GroupConfig) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	*p = *g.Duration(name, value, usage)
}

// Duration defines a duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func (g *GroupConfig) Duration(name string, value time.Duration, usage string) *time.Duration {
	bound := &value
	g.Flags[name] = &Flag{
		Type:    FlagTypeDuration,
		Default: value,
		Usage:   usage,
		Value:   &types.DurationValue{Bound: bound},
	}
	return bound
}

// URLVar defines a URL flag with specified name, default value, and usage string.
// The argument p points to a url.URL variable in which to store the value of the flag.
func (g *GroupConfig) URLVar(p *url.URL, name, value, usage string) {
	*p = *g.URL(name, value, usage)
}

// URL defines a URL flag with specified name, default value, and usage string.
// The return value is the address of a url.URL variable that stores the value of the flag.
func (g *GroupConfig) URL(name, value, usage string) *url.URL {
	bound := new(url.URL)
	if value != "" {
		parsed, err := url.Parse(value)
		if err != nil {
			panic(fmt.Sprintf("invalid default URL for flag '%s': %s", name, err))
		}
		*bound = *parsed // Copy the parsed URL into bound
	}
	g.Flags[name] = &Flag{
		Type:    FlagTypeURL,
		Default: value,
		Usage:   usage,
		Value:   &types.URLValue{Bound: bound},
	}
	return bound
}
