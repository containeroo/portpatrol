package dynflags

// UnknownGroup represents a runtime group with unrecognized values.
type UnknownGroup struct {
	Name   string                 // Identifier for the child group (e.g., "IDENTIFIER1").
	Values map[string]interface{} // Unrecognized flags and their parsed values.
}

// Lookup retrieves the value of a flag in the unknown group.
func (ug *UnknownGroup) Lookup(flagName string) interface{} {
	if ug == nil {
		return nil
	}

	return ug.Values[flagName]
}

// UnknownGroups represents all unknown groups with lookup and iteration support.
type UnknownGroups struct {
	groups       map[string]map[string]*UnknownGroup // Nested map of group name -> identifier -> UnknownGroup.
	unparsedArgs []string                            // List of arguments that couldn't be parsed into groups or flags.
}

// Lookup retrieves unknown groups by name.
func (ug *UnknownGroups) Lookup(groupName string) *UnknownIdentifiers {
	if ug == nil {
		return nil
	}

	if identifiers, exists := ug.groups[groupName]; exists {
		return &UnknownIdentifiers{Name: groupName, identifiers: identifiers}
	}
	return nil
}

// Groups returns the underlying map for direct iteration.
func (ug *UnknownGroups) Groups() map[string]map[string]*UnknownGroup {
	return ug.groups
}

// UnknownIdentifiers provides lookup for identifiers within a group.
type UnknownIdentifiers struct {
	Name        string
	identifiers map[string]*UnknownGroup
}

// Lookup retrieves a specific identifier within a group.
func (ui *UnknownIdentifiers) Lookup(identifier string) *UnknownGroup {
	if ui == nil {
		return nil
	}

	return ui.identifiers[identifier]
}

// Unknown returns an UnknownGroups instance for the DynFlags instance.
func (df *DynFlags) Unknown() *UnknownGroups {
	parsed := make(map[string]map[string]*UnknownGroup)
	for groupName, groups := range df.unknownGroups {
		identifierMap := make(map[string]*UnknownGroup)
		for _, group := range groups {
			identifierMap[group.Name] = group
		}
		parsed[groupName] = identifierMap
	}
	return &UnknownGroups{
		groups:       parsed,
		unparsedArgs: df.unparsedArgs,
	}
}

// UnparsedArgs returns the list of unparseable arguments.
func (df *DynFlags) UnparsedArgs() []string {
	return df.unparsedArgs
}
