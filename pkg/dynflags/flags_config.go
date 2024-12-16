package dynflags

// GroupConfig represents the static configuration for a group.
type GroupConfig struct {
	Name      string           // Name of the group.
	Usage     string           // Title for usage. If not set it takes the name of the group in Uppercase.
	Flags     map[string]*Flag // Flags within the group.
	flagOrder []string         // Order of flags.
}

// Lookup retrieves a flag in the group by its name.
func (gc *GroupConfig) Lookup(flagName string) *Flag {
	return gc.Flags[flagName]
}

// ConfigGroups represents all configuration groups with lookup and iteration support.
type ConfigGroups struct {
	groups map[string]*GroupConfig
}

// Lookup retrieves a configuration group by its name.
func (cg *ConfigGroups) Lookup(groupName string) *GroupConfig {
	return cg.groups[groupName]
}

// Groups returns the underlying map for direct iteration.
func (cg *ConfigGroups) Groups() map[string]*GroupConfig {
	return cg.groups
}

// Config returns a ConfigGroups instance for the dynflags instance.
func (df *DynFlags) Config() *ConfigGroups {
	return &ConfigGroups{groups: df.configGroups}
}
