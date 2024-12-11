package dynflags

import (
	"fmt"
	"strings"
)

// DynFlags manages all groups and flags
type DynFlags struct {
	Groups          map[string]map[string]*Group
	ContinueOnError bool
}

// New initializes a new DynFlags instance
func New(continueOnError bool) *DynFlags {
	return &DynFlags{
		Groups:          make(map[string]map[string]*Group),
		ContinueOnError: continueOnError,
	}
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

// Parse parses command-line arguments and sets flag values
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
			return fmt.Errorf("flag must follow the pattern: --group.identifier.key=value")
		}
		groupName, identifier, flagName := keyParts[0], keyParts[1], keyParts[2]

		if _, exists := df.Groups[groupName]; !exists {
			df.Groups[groupName] = make(map[string]*Group)
		}
		if _, exists := df.Groups[groupName][identifier]; !exists {
			df.Groups[groupName][identifier] = &Group{
				Name:  identifier,
				Flags: make(map[string]*Flag),
			}
		}

		group := df.Groups[groupName][identifier]
		flag, exists := group.Flags[flagName]
		if !exists {
			if !df.ContinueOnError {
				return fmt.Errorf("unknown flag '%s' in group '%s.%s'", flagName, groupName, identifier)
			}
			continue
		}

		parsedValue, err := flag.Parser.Parse(value)
		if err != nil {
			return fmt.Errorf("failed to parse flag '%s': %v", fullKey, err)
		}
		flag.Value = parsedValue
	}
	return nil
}

// GroupFlags retrieves all flags for a specific group
func (df *DynFlags) GroupFlags(groupPrefix string) map[string]*Group {
	return df.Groups[groupPrefix]
}
