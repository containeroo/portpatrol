package dynflags

import (
	"fmt"
	"strings"
)

// Parse parses the CLI arguments and populates parsed and unknown groups.
func (df *DynFlags) Parse(args []string) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Extract the key and value
		fullKey, value, err := df.extractKeyValue(arg, args, &i)
		if err != nil {
			// Handle unparseable arguments
			if df.parseBehavior == ExitOnError {
				return err
			}
			df.unparsedArgs = append(df.unparsedArgs, arg)
			continue
		}

		// Validate and split the key
		parentName, identifier, flagName, err := df.splitKey(fullKey)
		if err != nil {
			// Handle invalid keys
			if df.parseBehavior == ExitOnError {
				return err
			}
			df.unparsedArgs = append(df.unparsedArgs, arg)
			continue
		}

		// Handle the flag
		if err := df.handleFlag(parentName, identifier, flagName, value); err != nil {
			if df.parseBehavior == ExitOnError {
				return err
			}
			df.unparsedArgs = append(df.unparsedArgs, arg)
		}
	}
	return nil
}

// extractKeyValue extracts the key and value from an argument.
func (df *DynFlags) extractKeyValue(arg string, args []string, index *int) (string, string, error) {
	if !strings.HasPrefix(arg, "--") {
		// Invalid argument format
		return "", "", fmt.Errorf("invalid argument format: %s", arg)
	}

	arg = strings.TrimPrefix(arg, "--")

	// Handle "--key=value" format
	if strings.Contains(arg, "=") {
		parts := strings.SplitN(arg, "=", 2)
		return parts[0], parts[1], nil
	}

	// Handle "--key value" format
	if *index+1 < len(args) && !strings.HasPrefix(args[*index+1], "--") {
		*index++
		return arg, args[*index], nil
	}

	// Missing value for the key
	return "", "", fmt.Errorf("missing value for flag: --%s", arg)
}

// splitKey validates and splits a key into its components.
func (df *DynFlags) splitKey(fullKey string) (string, string, string, error) {
	parts := strings.Split(fullKey, ".")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("flag must follow the pattern: --<group>.<identifier>.<flag>")
	}
	return parts[0], parts[1], parts[2], nil
}

// handleFlag processes a known or unknown flag.
func (df *DynFlags) handleFlag(parentName, identifier, flagName, value string) error {
	if parentGroup, exists := df.configGroups[parentName]; exists {
		if flag := parentGroup.Lookup(flagName); flag != nil {
			// Known flag
			parsedGroup := df.createOrGetParsedGroup(parentGroup, identifier)
			return df.setFlagValue(parsedGroup, flagName, flag, value)
		}
	}

	// Unknown flag
	return df.handleUnknownFlag(parentName, identifier, flagName, value)
}

// setFlagValue sets the value of a known flag in the parsed group.
func (df *DynFlags) setFlagValue(parsedGroup *ParsedGroup, flagName string, flag *Flag, value string) error {
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

// handleUnknownFlag handles unknown flags based on the parse behavior.
func (df *DynFlags) handleUnknownFlag(parentName, identifier, flagName, value string) error {
	switch df.parseBehavior {
	case ExitOnError:
		return fmt.Errorf("unknown flag '%s' in group '%s'", flagName, parentName)
	case ParseUnknown:
		unknownGroup := df.createOrGetUnknownGroup(parentName, identifier)
		unknownGroup.Values[flagName] = value
	}
	return nil
}

// createOrGetParsedGroup retrieves or initializes a parsed group.
func (df *DynFlags) createOrGetParsedGroup(parentGroup *GroupConfig, identifier string) *ParsedGroup {
	for _, group := range df.parsedGroups[parentGroup.Name] {
		if group.Name == identifier {
			return group
		}
	}

	newGroup := &ParsedGroup{
		Parent: parentGroup,
		Name:   identifier,
		Values: make(map[string]interface{}),
	}
	df.parsedGroups[parentGroup.Name] = append(df.parsedGroups[parentGroup.Name], newGroup)
	return newGroup
}

// createOrGetUnknownGroup retrieves or initializes an unknown group.
func (df *DynFlags) createOrGetUnknownGroup(parentName, identifier string) *UnknownGroup {
	for _, group := range df.unknownGroups[parentName] {
		if group.Name == identifier {
			return group
		}
	}

	newGroup := &UnknownGroup{
		Name:   identifier,
		Values: make(map[string]interface{}),
	}
	df.unknownGroups[parentName] = append(df.unknownGroups[parentName], newGroup)
	return newGroup
}
