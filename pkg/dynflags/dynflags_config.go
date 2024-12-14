package dynflags

import "fmt"

// GroupConfig represents the static configuration for a group.
// It contains the group name, usage information, flags, and their order.
type GroupConfig struct {
	Name      string           // Name of the group
	usage     string           // Title for usage
	Flags     map[string]*Flag // Flags within the group
	flagOrder []string         // Order of flags
}

// Lookup retrieves a flag definition within the group by its name.
// Returns the flag if it exists, or an error if it doesn't.
func (gc *GroupConfig) Lookup(flagName string) (*Flag, error) {
	if flag, exists := gc.Flags[flagName]; exists {
		return flag, nil
	}
	return nil, fmt.Errorf("flag '%s' not found in config group '%s'", flagName, gc.Name)
}

// Groups returns a ConfigGroups wrapper around all registered groups
// for managing and retrieving group configurations.
func (df *DynFlags) Groups() *ConfigGroups {
	return &ConfigGroups{groups: df.configGroups}
}

// ConfigGroups is a wrapper around a map of `GroupConfig`
// providing methods to lookup and iterate over groups.
type ConfigGroups struct {
	groups map[string]*GroupConfig
}

// Lookup retrieves a `GroupConfig` by its name.
// Returns the group if it exists, or an error if it doesn't.
func (cg *ConfigGroups) Lookup(groupName string) (*GroupConfig, error) {
	if group, exists := cg.groups[groupName]; exists {
		return group, nil
	}
	return nil, fmt.Errorf("config group '%s' not found", groupName)
}

// Iterate provides access to all groups for iteration.
func (cg *ConfigGroups) Iterate() map[string]*GroupConfig {
	return cg.groups
}
