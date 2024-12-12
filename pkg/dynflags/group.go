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

// String registers a string flag in the group and binds it to a string pointer
func (g *GroupConfig) String(flagName, defaultValue, description string) *string {
	bound := &defaultValue
	g.Flags[flagName] = &Flag{
		Type:        FlagTypeString,
		Default:     defaultValue,
		Description: description,
		Value:       &types.StringValue{Bound: bound},
	}
	return bound
}

// Int registers an integer flag in the group and binds it to an int pointer
func (g *GroupConfig) Int(flagName string, defaultValue int, description string) *int {
	bound := &defaultValue
	g.Flags[flagName] = &Flag{
		Type:        FlagTypeInt,
		Default:     defaultValue,
		Description: description,
		Value:       &types.IntValue{Bound: bound},
	}
	return bound
}

// Int registers an integer flag in the group and binds it to an int pointer
func (g *GroupConfig) Float64(flagName string, defaultValue float64, description string) *float64 {
	bound := &defaultValue
	g.Flags[flagName] = &Flag{
		Type:        FlagTypeInt,
		Default:     defaultValue,
		Description: description,
		Value:       &types.Float64Value{Bound: bound},
	}
	return bound
}

// Bool registers a boolean flag in the group and binds it to a bool pointer
func (g *GroupConfig) Bool(flagName string, defaultValue bool, description string) *bool {
	bound := &defaultValue
	g.Flags[flagName] = &Flag{
		Type:        FlagTypeBool,
		Default:     defaultValue,
		Description: description,
		Value:       &types.BoolValue{Bound: bound},
	}
	return bound
}

// Duration registers a duration flag in the group and binds it to a time.Duration pointer

func (g *GroupConfig) Duration(flagName string, defaultValue time.Duration, description string) *time.Duration {
	bound := &defaultValue
	g.Flags[flagName] = &Flag{
		Type:        FlagTypeDuration,
		Default:     defaultValue,
		Description: description,
		Value:       &types.DurationValue{Bound: bound},
	}
	return bound
}

// URL registers a URL flag in the group and binds it to a url.URL pointer
func (g *GroupConfig) URL(flagName, defaultValue, description string) *url.URL {
	bound := new(url.URL)
	if defaultValue != "" {
		parsed, err := url.Parse(defaultValue)
		if err != nil {
			panic(fmt.Sprintf("invalid default URL for flag '%s': %s", flagName, err))
		}
		*bound = *parsed // Copy the parsed URL into bound
	}
	g.Flags[flagName] = &Flag{
		Type:        FlagTypeURL,
		Default:     defaultValue,
		Description: description,
		Value:       &types.URLValue{Bound: bound},
	}
	return bound
}
