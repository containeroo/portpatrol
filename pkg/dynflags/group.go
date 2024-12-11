package dynflags

import (
	"time"

	"github.com/containeroo/portpatrol/pkg/dynflags/parsers"
)

// Group represents a collection of flags under a shared identifier
type Group struct {
	Name  string
	Flags map[string]*Flag
}

// String registers a string flag in a group
func (g *Group) String(flagName, defaultValue, description string) *Flag {
	flag := &Flag{
		Parser:      &parsers.StringParser{},
		Default:     defaultValue,
		Description: description,
		Value:       defaultValue,
	}
	g.Flags[flagName] = flag
	return flag
}

func (g *Group) Int(flagName string, defaultValue int, description string) *Flag {
	flag := &Flag{
		Parser:      &parsers.IntParser{},
		Default:     defaultValue,
		Description: description,
		Value:       defaultValue,
	}
	g.Flags[flagName] = flag
	return flag
}

func (g *Group) Bool(flagName string, defaultValue bool, description string) *Flag {
	flag := &Flag{
		Parser:      &parsers.BoolParser{},
		Default:     defaultValue,
		Description: description,
		Value:       defaultValue,
	}
	g.Flags[flagName] = flag
	return flag
}

func (g *Group) Duration(flagName string, defaultValue time.Duration, description string) *Flag {
	flag := &Flag{
		Parser:      &parsers.DurationParser{},
		Default:     defaultValue,
		Description: description,
		Value:       defaultValue,
	}
	g.Flags[flagName] = flag
	return flag
}

func (g *Group) Float(flagName string, defaultValue float64, description string) *Flag {
	flag := &Flag{
		Parser:      &parsers.FloatParser{},
		Default:     defaultValue,
		Description: description,
		Value:       defaultValue,
	}
	g.Flags[flagName] = flag
	return flag
}

func (g *Group) URL(flagName string, defaultValue string, description string) *Flag {
	flag := &Flag{
		Parser:      &parsers.URLParser{},
		Default:     defaultValue,
		Description: description,
		Value:       defaultValue,
	}
	g.Flags[flagName] = flag
	return flag
}
