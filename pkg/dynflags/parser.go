package dynflags

import (
	"fmt"
	"strings"
)

// Parse parses the CLI arguments and populates parsed groups.
func (df *DynFlags) Parse(args []string) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		// Extract the key and value from the argument
		fullKey, value, err := df.extractKeyValue(arg, args, &i)
		if err != nil {
			if err := df.handleUnknownFlag("", "", "", arg); err != nil {
				return err
			}
			continue
		}

		// Split the fullKey into group, identifier, and flag name
		parentName, identifier, flagName, err := df.splitKey(fullKey)
		if err != nil {
			if err := df.handleUnknownFlag("", "", "", arg); err != nil {
				return err
			}

			continue
		}

		// Process groups and flags
		if err := df.processFlag(parentName, identifier, flagName, value); err != nil {
			return err
		}
	}
	return nil
}

// extractKeyValue extracts the key and value from a flag argument.
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

	// If the argument cannot be parsed, add it directly to unparsedArgs
	df.unparsedArgs = append(df.unparsedArgs, arg)
	return "", "", fmt.Errorf("missing value for flag: %s", arg)
}

// splitKey splits a key into its parent group, identifier, and flag name.
func (df *DynFlags) splitKey(fullKey string) (string, string, string, error) {
	parts := strings.Split(fullKey, ".")
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("flag must follow the pattern: --<group>.<identifier>.<flag>=value")
	}
	return parts[0], parts[1], parts[2], nil
}

// processFlag handles the logic for parsing flags and updating the appropriate groups.
func (df *DynFlags) processFlag(parentName, identifier, flagName, value string) error {
	parsedGroup := df.createOrGetParsedGroup(parentName, identifier)
	if parsedGroup == nil {
		// Handle unknown groups
		return df.handleUnknownFlag(parentName, identifier, flagName, value)
	}

	flag := parsedGroup.Parent.Lookup(flagName)
	if flag == nil {
		// Handle unknown flags in a known group
		return df.handleUnknownFlag(parentName, identifier, flagName, value)
	}

	return df.setFlagValue(parsedGroup, flag, flagName, value)
}

// handleUnknownFlag processes unknown flags.
func (df *DynFlags) handleUnknownFlag(parentName, identifier, flagName, value string) error {
	switch df.parseBehavior {
	case ExitOnError:
		return fmt.Errorf("unknown flag '%s' in group '%s'", flagName, parentName)
	case ParseUnknown:
		unknownGroup := df.createOrGetUnknownGroup(parentName, identifier)
		unknownGroup.Values[flagName] = value
		return nil
	case ContinueOnError:
		if parentName == "" && identifier == "" && flagName == "" && value != "" {
			df.unparsedArgs = append(df.unparsedArgs, value) // Append the original unparseable argument
		} else {
			df.unparsedArgs = append(df.unparsedArgs, fmt.Sprintf("--%s.%s.%s=%s", parentName, identifier, flagName, value))
		}
		return nil
	}
	return nil
}

// setFlagValue sets the value of a known flag.
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

// createOrGetParsedGroup retrieves or initializes a parsed group.
func (df *DynFlags) createOrGetParsedGroup(parentName, identifier string) *ParsedGroup {
	parentGroup, exists := df.configGroups[parentName]
	if !exists {
		return nil
	}

	if _, exists := df.parsedGroups[parentName]; !exists {
		df.parsedGroups[parentName] = []*ParsedGroup{}
	}

	for _, group := range df.parsedGroups[parentName] {
		if group.Name == identifier {
			return group
		}
	}

	newGroup := &ParsedGroup{
		Parent: parentGroup,
		Name:   identifier,
		Values: make(map[string]interface{}),
	}
	df.parsedGroups[parentName] = append(df.parsedGroups[parentName], newGroup)
	return newGroup
}

// createOrGetUnknownGroup retrieves or initializes an unknown group.
func (df *DynFlags) createOrGetUnknownGroup(parentName, identifier string) *UnknownGroup {
	if _, exists := df.unknownGroups[parentName]; !exists {
		df.unknownGroups[parentName] = []*UnknownGroup{}
	}

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
