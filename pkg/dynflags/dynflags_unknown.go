package dynflags

import "fmt"

type UnknownGroup struct {
	Name   string                 // Identifier for the child group
	Values map[string]interface{} // Parsed values for the group's flags
}

// Lookup retrieves the value of a flag in the unknown group.
func (ug *UnknownGroup) Lookup(flagName string) (interface{}, error) {
	if value, exists := ug.Values[flagName]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("flag '%s' not found in unknown group '%s'", flagName, ug.Name)
}

type UnknownGroups struct {
	groups map[string]*UnknownGroup
}

// Lookup retrieves an `UnknownGroup` by its name.
func (ug *UnknownGroups) Lookup(groupName string) (*UnknownGroup, error) {
	if group, exists := ug.groups[groupName]; exists {
		return group, nil
	}
	return nil, fmt.Errorf("unknown group '%s' not found", groupName)
}

// Unknown returns all unknown groups as a single `UnknownGroups` instance.
func (df *DynFlags) Unknown() *UnknownGroups {
	unknown := make(map[string]*UnknownGroup)
	for groupName, groups := range df.unknownGroups {
		if len(groups) > 0 {
			combinedGroup := &UnknownGroup{
				Name:   groupName,
				Values: make(map[string]interface{}),
			}
			for _, group := range groups {
				for k, v := range group.Values {
					combinedGroup.Values[k] = v
				}
			}
			unknown[groupName] = combinedGroup
		}
	}
	return &UnknownGroups{groups: unknown}
}
