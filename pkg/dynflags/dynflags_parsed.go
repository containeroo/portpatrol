package dynflags

import "fmt"

// ParsedGroup represents a runtime group with parsed values.
// It contains the group identifier, a reference to the parent configuration group,
// parsed flag values, and unrecognized flag values.
type ParsedGroup struct {
	Parent        *GroupConfig           // Reference to the parent static group
	Name          string                 // Identifier for the child group (e.g., "IDENTIFIER1")
	Values        map[string]interface{} // Parsed values for the group's flags
	unknownValues map[string]interface{} // Unrecognized flags and their parsed values
}

// Lookup retrieves the value of a parsed flag by its name.
// Returns the value if it exists, or an error if it doesn't.
func (pg *ParsedGroup) Lookup(flagName string) (interface{}, error) {
	if value, exists := pg.Values[flagName]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("flag '%s' not found in parsed group '%s'", flagName, pg.Name)
}

// ParsedGroups is a wrapper for managing multiple parsed groups.
type ParsedGroups struct {
	groups map[string]*ParsedGroup
}

// Lookup retrieves a `ParsedGroup` by its name.
// Returns the group if it exists, or an error if it doesn't.
func (pg *ParsedGroups) Lookup(groupName string) (*ParsedGroup, error) {
	if group, exists := pg.groups[groupName]; exists {
		return group, nil
	}
	return nil, fmt.Errorf("parsed group '%s' not found", groupName)
}

// Parsed combines all parsed groups and their values into a `ParsedGroups` structure.
// Groups with multiple identifiers are merged into a single `ParsedGroup`.
func (df *DynFlags) Parsed() *ParsedGroups {
	parsed := make(map[string]*ParsedGroup)
	for groupName, groups := range df.parsedGroups {
		if len(groups) > 0 {
			combinedGroup := &ParsedGroup{
				Name:   groupName,
				Values: make(map[string]interface{}),
			}
			for _, group := range groups {
				for k, v := range group.Values {
					combinedGroup.Values[k] = v
				}
			}
			parsed[groupName] = combinedGroup
		}
	}
	return &ParsedGroups{groups: parsed}
}
