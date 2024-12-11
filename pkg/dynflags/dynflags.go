package dynflags

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"
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
	parsedGroups  map[string][]*ParsedGroup // Parsed child groups organized by parent group
	parseBehavior ParseBehavior             // Parsing behavior
	output        io.Writer                 // Output for usage/help
	usage         func()                    // Customizable usage function
}

// New initializes a new DynFlags instance
func New(behavior ParseBehavior) *DynFlags {
	df := &DynFlags{
		configGroups:  make(map[string]*GroupConfig),
		parsedGroups:  make(map[string][]*ParsedGroup),
		parseBehavior: behavior,
		output:        os.Stdout,
	}
	df.usage = func() { df.Usage() }
	return df
}

// Group defines a new static group or retrieves an existing one
func (df *DynFlags) Group(name string) (*GroupConfig, error) {
	if _, exists := df.configGroups[name]; exists {
		return nil, fmt.Errorf("group '%s' already exists", name)
	}
	group := &GroupConfig{
		Name:  name,
		Flags: make(map[string]*Flag),
	}
	df.configGroups[name] = group
	return group, nil
}

// GetParsedGroups retrieves all parsed child groups for a parent group
func (df *DynFlags) GetParsedGroups(parentName string) ([]*ParsedGroup, error) {
	groups, exists := df.parsedGroups[parentName]
	if !exists {
		return nil, fmt.Errorf("no parsed groups found for parent '%s'", parentName)
	}
	return groups, nil
}

// Parse parses the CLI arguments and populates parsed groups
func (df *DynFlags) Parse(args []string) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return fmt.Errorf("invalid flag format: %s", arg)
		}

		var fullKey, value string
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg[2:], "=", 2)
			fullKey, value = parts[0], parts[1]
		} else {
			fullKey = arg[2:]
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				value = args[i+1]
				i++
			} else {
				return fmt.Errorf("missing value for flag: %s", fullKey)
			}
		}

		keyParts := strings.Split(fullKey, ".")
		if len(keyParts) < 3 {
			return fmt.Errorf("flag must follow the pattern: --<group>.<identifier>.<flag>=value")
		}
		parentName := keyParts[0]
		identifier := keyParts[1]
		flagName := keyParts[2]

		parsedGroup := df.createOrGetParsedGroup(parentName, identifier)
		if parsedGroup == nil {
			return fmt.Errorf("unknown parent group: '%s'", parentName)
		}

		flag, exists := parsedGroup.Parent.Flags[flagName]
		if !exists {
			switch df.parseBehavior {
			case ExitOnError:
				return fmt.Errorf("unknown flag '%s' in group '%s'", flagName, parentName)
			case ContinueOnError, IgnoreUnknown:
				continue
			}
		}

		parsedValue, err := flag.Parser.Parse(value)
		if err != nil {
			return fmt.Errorf("failed to parse value for flag '%s': %v", fullKey, err)
		}

		parsedGroup.Values[flagName] = parsedValue

		// Update the bound variable if applicable
		switch flag.Type {
		case "string":
			if ptr, ok := flag.Value.(*string); ok {
				*ptr = parsedValue.(string)
			}
		case "int":
			if ptr, ok := flag.Value.(*int); ok {
				*ptr = parsedValue.(int)
			}
		case "bool":
			if ptr, ok := flag.Value.(*bool); ok {
				*ptr = parsedValue.(bool)
			}
		case "duration":
			if ptr, ok := flag.Value.(*time.Duration); ok {
				*ptr = parsedValue.(time.Duration)
			}
		case "float":
			if ptr, ok := flag.Value.(*float64); ok {
				*ptr = parsedValue.(float64)
			}
		}
	}
	return nil
}

// createOrGetParsedGroup creates or retrieves a child group for a parent group
func (df *DynFlags) createOrGetParsedGroup(parentName, identifier string) *ParsedGroup {
	parentGroup, exists := df.configGroups[parentName]
	if !exists {
		return nil
	}

	for _, group := range df.parsedGroups[parentName] {
		if group.Name == identifier {
			return group
		}
	}

	parsedGroup := &ParsedGroup{
		Parent: parentGroup,
		Name:   identifier,
		Values: make(map[string]interface{}),
	}
	df.parsedGroups[parentName] = append(df.parsedGroups[parentName], parsedGroup)
	return parsedGroup
}

// DefaultUsage provides the default usage output
func (df *DynFlags) Usage() {
	fmt.Fprintf(df.output, "Usage: [OPTIONS] [--<group>.<identifier>.<flag> value]\n\n")
	df.PrintDefaults()
}

// PrintDefaults prints all registered flags
func (df *DynFlags) PrintDefaults() {
	w := tabwriter.NewWriter(df.output, 0, 8, 2, ' ', 0)
	defer w.Flush()
	fmt.Fprintln(w, "FLAG\tDESCRIPTION")
	for groupName, group := range df.configGroups {
		for flagName, flag := range group.Flags {
			description := flag.Description
			if flag.Default != nil && flag.Default != "" {
				description = fmt.Sprintf("%s (defaults to %v)", flag.Description, flag.Default)
			}
			fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s %s\t%s\n", groupName, flagName, strings.ToUpper(flag.Type), description)
		}
	}
}
