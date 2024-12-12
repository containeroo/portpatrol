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
	title         string                    // Title in the help message
	description   string                    // Description after the title in the help message
	epilog        string                    // Epilog in the help message
}

// ParsedGroup represents a runtime group with parsed values
type ParsedGroup struct {
	Parent *GroupConfig           // Reference to the parent static group
	Name   string                 // Identifier for the child group (e.g., "IDENTIFIER1")
	Values map[string]interface{} // Parsed values for the group's flags
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

// AddTitle adds a title to the help message
func (df *DynFlags) AddTitle(title string) {
	df.title = title
}

// AddDescription adds a descripton after the Title
func (df *DynFlags) AddDescription(description string) {
	df.description = description
}

// AddEpilog adds an epilog after the description of the dynamic flags to the help message
func (df *DynFlags) AddEpilog(epilog string) {
	df.epilog = epilog
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

		parsedValue, err := flag.Value.Parse(value)
		if err != nil {
			return fmt.Errorf("failed to parse value for flag '%s': %v", fullKey, err)
		}

		if err := flag.Value.Set(parsedValue); err != nil {
			return fmt.Errorf("failed to set value for flag '%s': %v", fullKey, err)
		}

		parsedGroup.Values[flagName] = parsedValue
	}
	return nil
}

// createOrGetParsedGroup creates or retrieves a child group for a parent group
func (df *DynFlags) createOrGetParsedGroup(parentName, identifier string) *ParsedGroup {
	parentGroup, exists := df.configGroups[parentName]
	if !exists {
		return nil // Return nil if parent group doesn't exist
	}

	// Check if the group already exists
	for _, group := range df.parsedGroups[parentName] {
		if group.Name == identifier {
			return group
		}
	}

	// Create a new parsed group
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
	fmt.Fprintf(df.output, "Usage: [--<group>.<identifier>.<flag> value]\n\n")
	df.PrintDefaults()
}

// PrintDefaults prints all registered flags
func (df *DynFlags) PrintDefaults() {
	w := tabwriter.NewWriter(df.output, 0, 8, 2, ' ', 0)
	defer w.Flush()

	if df.title != "" {
		fmt.Fprintln(w, df.title)
	}

	if df.description != "" {
		fmt.Fprintln(w, df.description)
	}

	for groupName, group := range df.configGroups {
		fmt.Fprintln(w, strings.ToUpper(groupName))
		fmt.Fprintln(w, "  Flag\tUsage")
		for flagName, flag := range group.Flags {
			usage := flag.Usage
			if flag.Default != nil && flag.Default != "" {
				usage = fmt.Sprintf("%s (default: %v)", flag.Usage, flag.Default)
			}
			fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s %s\t%s\n", groupName, flagName, flag.Type, usage)
		}
		fmt.Fprintln(w, "")
	}

	if df.epilog != "" {
		fmt.Fprintln(w, df.epilog)
	}
}

// SetOutput sets the output writer
func (df *DynFlags) SetOutput(buf io.Writer) {
	df.output = buf
}

// GetAllParsedGroups returns all parsed groups
func (df *DynFlags) GetAllParsedGroups() map[string][]*ParsedGroup {
	return df.parsedGroups
}

// GetValue returns the value of a flag with the given name
func (pg *ParsedGroup) GetValue(flagName string) interface{} {
	return pg.Values[flagName]
}

// GetString returns the string value of a flag with the given name
func (pg *ParsedGroup) GetString(flagName string) (string, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return "", fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("flag '%s' is not a string", flagName)
}

// GetBool returns the bool value of a flag with the given name
func (pg *ParsedGroup) GetBool(flagName string) (bool, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return false, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if boolVal, ok := value.(bool); ok {
		return boolVal, nil
	}
	return false, fmt.Errorf("flag '%s' is not a bool", flagName)
}

// GetInt returns the int value of a flag with the given name
func (pg *ParsedGroup) GetInt(flagName string) (int, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return 0, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if intVal, ok := value.(int); ok {
		return intVal, nil
	}
	return 0, fmt.Errorf("flag '%s' is not an int", flagName)
}

// GetFloat64 returns the float64 value of a flag with the given name
func (pg *ParsedGroup) GetFloat64(flagName string) (float64, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return 0, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if floatVal, ok := value.(float64); ok {
		return floatVal, nil
	}
	return 0, fmt.Errorf("flag '%s' is not a float64", flagName)
}

// GetDuration returns the time.Duration value of a flag with the given name
func (pg *ParsedGroup) GetDuration(flagName string) (time.Duration, error) {
	vaue, exists := pg.Values[flagName]
	if !exists {
		return 0, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if durationVal, ok := vaue.(time.Duration); ok {
		return durationVal, nil
	}
	return 0, fmt.Errorf("flag '%s' is not a time.Duration", flagName)
}
