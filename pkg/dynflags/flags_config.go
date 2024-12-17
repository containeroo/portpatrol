package dynflags

// ConfigGroup represents the static configuration for a group.
type ConfigGroup struct {
	Name      string           // Name of the group.
	Usage     string           // Title for usage. If not set it takes the name of the group in Uppercase.
	Flags     map[string]*Flag // Flags within the group.
	flagOrder []string         // Order of flags.
}

// Lookup retrieves a flag in the group by its name.
func (gc *ConfigGroup) Lookup(flagName string) *Flag {
	if gc == nil {
		return nil
	}

	return gc.Flags[flagName]
}

// ConfigGroups represents all configuration groups with lookup and iteration support.
type ConfigGroups struct {
	groups map[string]*ConfigGroup
}

// Lookup retrieves a configuration group by its name.
func (cg *ConfigGroups) Lookup(groupName string) *ConfigGroup {
	if cg == nil {
		return nil
	}

	return cg.groups[groupName]
}

// Groups returns the underlying map for direct iteration.
func (cg *ConfigGroups) Groups() map[string]*ConfigGroup {
	if cg == nil {
		return nil
	}

	return cg.groups
}

// Config returns a ConfigGroups instance for the dynflags instance.
func (df *DynFlags) Config() *ConfigGroups {
	if df == nil {
		return nil
	}

	return &ConfigGroups{groups: df.configGroups}
}
