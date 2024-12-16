package dynflags

import (
	"fmt"
	"io"
	"os"
)

// ParseBehavior defines how the parser handles errors
type ParseBehavior int

const (
	// Continue parsing on error
	ContinueOnError ParseBehavior = iota
	// Try to parse unknown flags. Unknown flags can be retrived with the method Unknown() on the DynFlags instance
	ParseUnknown
	// Exit on error
	ExitOnError
)

// DynFlags manages configuration and parsed values
type DynFlags struct {
	configGroups  map[string]*GroupConfig    // Static parent groups
	groupOrder    []string                   // Order of group names
	SortGroups    bool                       // Sort groups in help message
	SortFlags     bool                       // Sort flags in help message
	parsedGroups  map[string][]*ParsedGroup  // Parsed child groups organized by parent group
	unknownGroups map[string][]*UnknownGroup // Unknown parent groups and their parsed values
	parseBehavior ParseBehavior              // Parsing behavior
	unparsedArgs  []string                   // Arguments that couldn't be parsed
	output        io.Writer                  // Output for usage/help
	usage         func()                     // Customizable usage function
	title         string                     // Title in the help message
	description   string                     // Description after the title in the help message
	epilog        string                     // Epilog in the help message
}

// New initializes a new DynFlags instance
func New(behavior ParseBehavior) *DynFlags {
	df := &DynFlags{
		configGroups:  make(map[string]*GroupConfig),
		parsedGroups:  make(map[string][]*ParsedGroup),
		unknownGroups: make(map[string][]*UnknownGroup),
		parseBehavior: behavior,
		output:        os.Stdout,
	}
	df.usage = func() { df.Usage() }
	return df
}

// Title adds a title to the help message
func (df *DynFlags) Title(title string) {
	df.title = title
}

// Description adds a descripton after the Title
func (df *DynFlags) Description(description string) {
	df.description = description
}

// Epilog adds an epilog after the description of the dynamic flags to the help message
func (df *DynFlags) Epilog(epilog string) {
	df.epilog = epilog
}

// Group defines a new group or retrieves an existing one
func (df *DynFlags) Group(name string) *GroupConfig {
	if _, exists := df.configGroups[name]; exists {
		return df.configGroups[name]
	}

	df.groupOrder = append(df.groupOrder, name)

	group := &GroupConfig{
		Name:  name,
		Flags: make(map[string]*Flag),
	}
	df.configGroups[name] = group
	return group
}

// DefaultUsage provides the default usage output
func (df *DynFlags) Usage() {
	fmt.Fprintf(df.output, "Usage: [--<group>.<identifier>.<flag> value]\n\n")
	df.PrintDefaults()
}

// SetOutput sets the output writer
func (df *DynFlags) SetOutput(buf io.Writer) {
	df.output = buf
}
