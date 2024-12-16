package dynflags

// UnknownGroup represents an unknown group with unrecognized values.
type UnknownGroup struct {
	Name   string                 // Identifier for the child group
	Values map[string]interface{} // Unrecognized flags and their parsed values
}

// Lookup retrieves the value of a flag in the unknown group.
func (ug *UnknownGroup) Lookup(flagName string) interface{} {
	if value, exists := ug.Values[flagName]; exists {
		return value
	}
	return nil
}

// UnknownGroups represents all unknown groups with lookup and iteration support.
type UnknownGroups struct {
	groups map[string][]*UnknownGroup
}

// Lookup retrieves unknown groups by name.
func (ug *UnknownGroups) Lookup(groupName string) []*UnknownGroup {
	if groups, exists := ug.groups[groupName]; exists {
		return groups
	}
	return nil
}

// Groups returns the underlying map for direct iteration.
func (ug *UnknownGroups) Groups() map[string][]*UnknownGroup {
	return ug.groups
}

// Unknown returns an UnknownGroups instance for the dynflags instance.
func (df *DynFlags) Unknown() *UnknownGroups {
	return &UnknownGroups{groups: df.unknownGroups}
}
