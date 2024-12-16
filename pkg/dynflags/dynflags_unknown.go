package dynflags

// UnknownGroup represents a runtime group with parsed values.
type UnknownGroup struct {
	Name   string                 // Identifier for the child group (e.g., "IDENTIFIER1").
	Values map[string]interface{} // Unknown values for the group's flags.
}

// Lookup retrieves the value of a flag in the parsed group.
func (pg *UnknownGroup) Lookup(flagName string) interface{} {
	return pg.Values[flagName]
}

// UnknownGroups represents all parsed groups with lookup and iteration support.
type UnknownGroups struct {
	groups map[string]map[string]*UnknownGroup // Nested map of group name -> identifier -> UnknownGroup.
}

// Lookup retrieves a group by its name.
func (pg *UnknownGroups) Lookup(groupName string) *UnknownIdentifiers {
	if identifiers, exists := pg.groups[groupName]; exists {
		return &UnknownIdentifiers{identifiers: identifiers}
	}
	return nil
}

// Groups returns the underlying map for direct iteration.
func (pg *UnknownGroups) Groups() map[string]map[string]*UnknownGroup {
	return pg.groups
}

// UnknownIdentifiers provides lookup for identifiers within a group.
type UnknownIdentifiers struct {
	identifiers map[string]*UnknownGroup
}

// Lookup retrieves a specific identifier within a group.
func (gi *UnknownIdentifiers) Lookup(identifier string) *UnknownGroup {
	return gi.identifiers[identifier]
}

// Unknown returns a UnknownGroups instance for the dynflags instance.
func (df *DynFlags) Unknown() *UnknownGroups {
	parsed := make(map[string]map[string]*UnknownGroup)
	for groupName, groups := range df.unknownGroups {
		identifierMap := make(map[string]*UnknownGroup)
		for _, group := range groups {
			identifierMap[group.Name] = group
		}
		parsed[groupName] = identifierMap
	}
	return &UnknownGroups{groups: parsed}
}
