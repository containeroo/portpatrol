package dynflags

// ParsedGroup represents a runtime group with parsed values
type ParsedGroup struct {
	Parent *GroupConfig           // Reference to the parent static group
	Name   string                 // Identifier for the child group (e.g., "IDENTIFIER1")
	Values map[string]interface{} // Parsed values for the group's flags
}

// Lookup retrieves the value of a flag in the parsed group.
func (pg *ParsedGroup) Lookup(flagName string) interface{} {
	if value, exists := pg.Values[flagName]; exists {
		return value
	}
	return nil
}

// ParsedGroups represents all parsed groups with lookup and iteration support.
type ParsedGroups struct {
	groups map[string][]*ParsedGroup
}

// Lookup retrieves parsed groups by name.
func (pg *ParsedGroups) Lookup(groupName string) []*ParsedGroup {
	if groups, exists := pg.groups[groupName]; exists {
		return groups
	}
	return nil
}

// Groups returns the underlying map for direct iteration.
func (pg *ParsedGroups) Groups() map[string][]*ParsedGroup {
	return pg.groups
}

// Parsed returns a ParsedGroups instance for the dynflags instance.
func (df *DynFlags) Parsed() *ParsedGroups {
	return &ParsedGroups{groups: df.parsedGroups}
}
