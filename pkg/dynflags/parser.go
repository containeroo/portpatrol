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

		// Extract the key and value from the argument
		fullKey, value, err := df.extractKeyValue(arg, args, &i)
		if err != nil {
			return err
		}

		// Split the fullKey into group, identifier, and flag name
		parentName, identifier, flagName, err := df.splitKey(fullKey)
		if err != nil {
			return err
		}

		// Process groups and flags
		if err := df.processFlag(parentName, identifier, flagName, value); err != nil {
			return err
		}
	}
	return nil
}

// extractKeyValue extracts the key and value from a flag argument
func (df *DynFlags) extractKeyValue(arg string, args []string, index *int) (string, string, error) {
	if strings.Contains(arg, "=") {
		parts := strings.SplitN(arg[2:], "=", 2)
		return parts[0], parts[1], nil
	}

	key := arg[2:] // Remove leading '--'
	if *index+1 < len(args) && !strings.HasPrefix(args[*index+1], "--") {
		*index++
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

// processFlag handles the logic for parsing flags and updating the appropriate groups
func (df *DynFlags) processFlag(parentName, identifier, flagName, value string) error {
	parsedGroup := df.createOrGetParsedGroup(parentName, identifier)
	if parsedGroup.Parent == nil {
		// Handle unknown groups
		return df.handleUnknownFlag(parentName, identifier, flagName, value)
	}

	flag, exists := parsedGroup.Parent.Flags[flagName]
	if !exists {
		// Handle unknown flags in a known group
		return df.handleUnknownFlag(parentName, identifier, flagName, value)
	}

	return df.setFlagValue(parsedGroup, flag, flagName, value)
}

// handleUnknownFlag processes unknown flags
func (df *DynFlags) handleUnknownFlag(parentName, identifier, flagName, value string) error {
	switch df.parseBehavior {
	case ExitOnError:
		return fmt.Errorf("unknown flag '%s' in group '%s'", flagName, parentName)
	case ContinueOnError:
		return nil
	case IgnoreUnknown:
		unknownGroup := df.createOrGetUnknownGroup(parentName, identifier)
		unknownGroup.Values[flagName] = value
		return nil
	}
	return nil
}

// setFlagValue sets the value of a known flag
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

// createOrGetParsedGroup retrieves or initializes a parsed group
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

// createOrGetUnknownGroup retrieves or initializes an unknown group
func (df *DynFlags) createOrGetUnknownGroup(parentName, identifier string) *UnknownGroup {
	for _, group := range df.unknownGroups[parentName] {
		if group.Name == identifier {
			return group
		}
	}

	unknownGroup := &UnknownGroup{
		Name:   identifier,
		Values: make(map[string]interface{}),
	}
	df.unknownGroups[parentName] = append(df.unknownGroups[parentName], unknownGroup)
	return unknownGroup
}
