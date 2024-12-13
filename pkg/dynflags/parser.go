package dynflags

import (
	"fmt"
	"strings"
)

// Parse parses the CLI arguments and populates parsed groups
func (df *DynFlags) Parse(args []string) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return fmt.Errorf("invalid flag format: %s", arg)
		}

		fullKey, value, err := df.extractKeyValue(arg, args, &i)
		if err != nil {
			return err
		}

		// Split the fullKey into group, identifier, and flag name
		parentName, identifier, flagName, err := df.splitKey(fullKey)
		if err != nil {
			return err
		}

		// Handle the parsed group and flags
		parsedGroup := df.createOrGetParsedGroup(parentName, identifier)
		if parsedGroup.Parent == nil {
			if err := df.handleUnknownGroup(parsedGroup, parentName, flagName, value); err != nil {
				return err
			}
			continue
		}

		flag, exists := parsedGroup.Parent.Flags[flagName]
		if !exists {
			if err := df.handleUnknownFlag(parsedGroup, parentName, flagName, value); err != nil {
				return err
			}
			continue
		}

		// Set the parsed flag value
		if err := df.setFlagValue(parsedGroup, flag, flagName, value); err != nil {
			return err
		}
	}
	return nil
}

// extractKeyValue extracts the key and value from a flag argument
func (df *DynFlags) extractKeyValue(arg string, args []string, index *int) (string, string, error) {
	// If the argument contains an equals sign, split it into key and value
	if strings.Contains(arg, "=") {
		parts := strings.SplitN(arg[2:], "=", 2)
		return parts[0], parts[1], nil
	}

	key := arg[2:] // Remove leading '--'
	// Check if the next argument is available and is not another flag
	if *index+1 < len(args) && !strings.HasPrefix(args[*index+1], "--") {
		*index++ // Move to the next argument
		return key, args[*index], nil
	}

	return key, "", fmt.Errorf("missing value for flag: %s", key)
}

// splitKey splits a key into its parent group, identifier, and flag name
func (df *DynFlags) splitKey(fullKey string) (string, string, string, error) {
	parts := strings.Split(fullKey, ".")
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("flag must follow the pattern: --<group>.<identifier>.<flag>=value")
	}
	return parts[0], parts[1], parts[2], nil
}

// handleUnknownGroup handles unknown groups
func (df *DynFlags) handleUnknownGroup(parsedGroup *ParsedGroup, parentName, flagName, value string) error {
	switch df.parseBehavior {
	case ExitOnError:
		return fmt.Errorf("unknown group: '%s'", parentName)
	case ContinueOnError:
		// Log the unknown group for debugging or tracking but continue parsing
		return nil
	case IgnoreUnknown:
		// Store the unknown flag and its value in the parsedGroup
		parsedGroup.unknownValues[flagName] = value
		return nil
	}
	// Default to ContinueOnError if an unsupported behavior is encountered
	return nil
}

// handleUnknownFlag handles unknown flags
func (df *DynFlags) handleUnknownFlag(parsedGroup *ParsedGroup, parentName, flagName, value string) error {
	switch df.parseBehavior {
	case ExitOnError:
		return fmt.Errorf("unknown flag '%s' in group '%s'", flagName, parentName)
	case ContinueOnError:
		return nil
	case IgnoreUnknown:
		parsedGroup.unknownValues[flagName] = value
		return nil
	}
	return nil
}

// setFlagValue sets the value of a flag
func (df *DynFlags) setFlagValue(parsedGroup *ParsedGroup, flag *Flag, flagName, value string) error {
	parsedValue, err := flag.Value.Parse(value)
	if err != nil {
		return fmt.Errorf("failed to parse value for flag '%s': %v", flagName, err)
	}

	if err := flag.Value.Set(parsedValue); err != nil {
		return fmt.Errorf("failed to set value for flag '%s': %v", flagName, err)
	}

	parsedGroup.Values[flagName] = parsedValue
	return nil
}

// createOrGetParsedGroup creates or retrieves a child group for a parent group
func (df *DynFlags) createOrGetParsedGroup(parentName, identifier string) *ParsedGroup {
	parentGroup, exists := df.configGroups[parentName]
	if !exists {
		// Handle unknown groups
		if _, ok := df.unknownGroups[parentName]; !ok {
			df.unknownGroups[parentName] = []*ParsedGroup{}
		}

		for _, group := range df.unknownGroups[parentName] {
			if group.Name == identifier {
				return group
			}
		}

		// Create a new unknown group
		parsedGroup := &ParsedGroup{
			Parent:        nil,
			Name:          identifier,
			Values:        make(map[string]interface{}),
			unknownValues: make(map[string]interface{}),
		}
		df.unknownGroups[parentName] = append(df.unknownGroups[parentName], parsedGroup)
		return parsedGroup
	}

	// Check if the group already exists
	for _, group := range df.parsedGroups[parentName] {
		if group.Name == identifier {
			return group
		}
	}

	// Create a new parsed group
	parsedGroup := &ParsedGroup{
		Parent:        parentGroup,
		Name:          identifier,
		Values:        make(map[string]interface{}),
		unknownValues: make(map[string]interface{}),
	}
	df.parsedGroups[parentName] = append(df.parsedGroups[parentName], parsedGroup)
	return parsedGroup
}
