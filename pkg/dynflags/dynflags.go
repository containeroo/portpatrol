package dynflags

import (
	"fmt"
	"io"
	"os"
)

// ParseBehavior defines how the parser handles errors
type ParseBehavior int

const (
	ContinueOnError ParseBehavior = iota
	ExitOnError
	IgnoreUnknown
)

// DynFlags manages configuration and parsed values
type DynFlags struct {
	configGroups  map[string]*GroupConfig   // Static parent groups
	groupOrder    []string                  // Order of group names
	SortGroups    bool                      // Sort groups in help message
	SortFlags     bool                      // Sort flags in help message
	parsedGroups  map[string][]*ParsedGroup // Parsed child groups organized by parent group
	unknownGroups map[string][]*ParsedGroup // Unknown parent groups and their parsed values
	parseBehavior ParseBehavior             // Parsing behavior
	output        io.Writer                 // Output for usage/help
	usage         func()                    // Customizable usage function
	title         string                    // Title in the help message
	description   string                    // Description after the title in the help message
	epilog        string                    // Epilog in the help message
}

// GroupConfig represents the static configuration for a group
type GroupConfig struct {
	Name      string           // Name of the group
	usage     string           // Title for usage
	Flags     map[string]*Flag // Flags within the group
	flagOrder []string         // Order of flags
}

// ParsedGroup represents a runtime group with parsed values
type ParsedGroup struct {
	Parent        *GroupConfig           // Reference to the parent static group
	Name          string                 // Identifier for the child group (e.g., "IDENTIFIER1")
	Values        map[string]interface{} // Parsed values for the group's flags
	unknownValues map[string]interface{} // Unrecognized flags and their parsed values
}

// New initializes a new DynFlags instance
func New(behavior ParseBehavior) *DynFlags {
	df := &DynFlags{
		configGroups:  make(map[string]*GroupConfig),
		parsedGroups:  make(map[string][]*ParsedGroup),
		unknownGroups: make(map[string][]*ParsedGroup),
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
		panic(fmt.Sprintf("group '%s' already exists", name))
	}

	df.groupOrder = append(df.groupOrder, name)

	group := &GroupConfig{
		Name:  name,
		Flags: make(map[string]*Flag),
	}
	df.configGroups[name] = group
	return group
}

// Groups returns all configured Groups
func (df *DynFlags) Groups() map[string]*GroupConfig {
	return df.configGroups
}

// Parsed returns all parsed groups
func (df *DynFlags) Parsed() map[string][]*ParsedGroup {
	return df.parsedGroups
}

// Unknown returns alls Unknown parsed groups
func (df *DynFlags) Unknown() map[string][]*ParsedGroup {
	return df.unknownGroups
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

// GetUnknownValues returns all unrecognized flags in a group
func (pg *ParsedGroup) GetUnknownValues() map[string]interface{} {
	return pg.unknownValues
}

// GetValue returns the value of a flag with the given name
func (pg *ParsedGroup) GetValue(name string) interface{} {
	return pg.Values[name]
}
