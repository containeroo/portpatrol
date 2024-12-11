package dynflags

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// ParseBehavior defines how the parser handles errors
type ParseBehavior int

const (
	// ContinueOnError skips unregistered flags but continues parsing
	ContinueOnError ParseBehavior = iota
	// ExitOnError stops parsing and exits on encountering an unregistered flag
	ExitOnError
	// IgnoreUnknown silently ignores unregistered flags
	IgnoreUnknown
)

// DynFlags manages all groups and flags
type DynFlags struct {
	Groups        map[string]map[string]*Group
	ParseBehavior ParseBehavior
	Output        io.Writer
	Usage         func()
}

// New initializes a new DynFlags instance with a specific parsing behavior
func New(behavior ParseBehavior) *DynFlags {
	df := &DynFlags{
		Groups:        make(map[string]map[string]*Group),
		ParseBehavior: behavior,
		Output:        os.Stdout, // Default output to stdout
	}
	df.Usage = func() { df.DefaultUsage() }
	return df
}

// Group retrieves or creates a new group under the given prefix
func (df *DynFlags) Group(prefix string) *Group {
	if _, exists := df.Groups[prefix]; !exists {
		df.Groups[prefix] = make(map[string]*Group)
	}
	return &Group{
		Name:  prefix,
		Flags: make(map[string]*Flag),
	}
}

// GroupFlags retrieves all flags for a specific group
func (df *DynFlags) GroupFlags(groupPrefix string) map[string]*Group {
	return df.Groups[groupPrefix]
}

// SetOutput sets the output destination for the Usage function
func (df *DynFlags) SetOutput(output io.Writer) {
	df.Output = output
}

// PrintDefaults prints all registered flags in a tabular format
func (df *DynFlags) PrintDefaults() {
	w := tabwriter.NewWriter(df.Output, 0, 8, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "FLAG\tTYPE\tDEFAULT\tDESCRIPTION")
	fmt.Fprintln(w, df.Groups)
	for groupName, group := range df.Groups {
		for _, identifierGroup := range group {
			for flagName, flag := range identifierGroup.Flags {
				fmt.Fprintf(w, "--%s.%s.%s\t%s\t%v\t%s\n",
					groupName, flagName, identifierGroup.Name, flag.Type, flag.Default, flag.Description)
			}
		}
	}
}

// DefaultUsage provides the default usage output
func (df *DynFlags) DefaultUsage() {
	fmt.Fprintf(df.Output, "Usage: [OPTIONS] [--<group>.<identifier>.<property> value]\n\n")
	df.PrintDefaults()
}
