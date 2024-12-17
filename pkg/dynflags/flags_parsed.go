package dynflags

// ParsedGroup represents a runtime group with parsed values.
type ParsedGroup struct {
	Parent *ConfigGroup           // Reference to the parent static group.
	Name   string                 // Identifier for the child group (e.g., "IDENTIFIER1").
	Values map[string]interface{} // Parsed values for the group's flags.
}

// Lookup retrieves the value of a flag in the parsed group.
func (pg *ParsedGroup) Lookup(flagName string) interface{} {
	if pg == nil {
		return nil
	}

	return pg.Values[flagName]
}

// ParsedGroups represents all parsed groups with lookup and iteration support.
type ParsedGroups struct {
	groups map[string]map[string]*ParsedGroup // Nested map of group name -> identifier -> ParsedGroup.
}

// Lookup retrieves a group by its name.
func (pg *ParsedGroups) Lookup(groupName string) *ParsedIdentifiers {
	if pg == nil {
		return nil
	}
	if identifiers, exists := pg.groups[groupName]; exists {
		return &ParsedIdentifiers{Name: groupName, identifiers: identifiers}
	}
	return nil
}

// Groups returns the underlying map for direct iteration.
func (pg *ParsedGroups) Groups() map[string]map[string]*ParsedGroup {
	return pg.groups
}

// ParsedIdentifiers provides lookup for identifiers within a group.
type ParsedIdentifiers struct {
	Name        string
	identifiers map[string]*ParsedGroup
}

// Lookup retrieves a specific identifier within a group.
func (gi *ParsedIdentifiers) Lookup(identifier string) *ParsedGroup {
	if gi == nil {
		return nil
	}

	return gi.identifiers[identifier]
}

// Parsed returns a ParsedGroups instance for the dynflags instance.
func (df *DynFlags) Parsed() *ParsedGroups {
	parsed := make(map[string]map[string]*ParsedGroup)
	for groupName, groups := range df.parsedGroups {
		identifierMap := make(map[string]*ParsedGroup)
		for _, group := range groups {
			identifierMap[group.Name] = group
		}
		parsed[groupName] = identifierMap
	}
	return &ParsedGroups{groups: parsed}
}
