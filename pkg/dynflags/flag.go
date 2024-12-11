package dynflags

import "github.com/containeroo/portpatrol/pkg/dynflags/parsers"

// Flag represents a single command-line flag
type Flag struct {
	Parser      parsers.Parser
	Description string
	Default     interface{}
	Value       interface{}
}
