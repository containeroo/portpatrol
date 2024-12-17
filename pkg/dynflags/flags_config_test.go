package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestConfigGroup(t *testing.T) {
	t.Parallel()

	t.Run("Lookup existing flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{
			Name:  "testGroup",
			Flags: map[string]*dynflags.Flag{"flag1": {Usage: "Test Flag"}},
		}
		flag := group.Lookup("flag1")
		assert.NotNil(t, flag)
		assert.Equal(t, "Test Flag", flag.Usage)
	})

	t.Run("Lookup non-existing flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{
			Name:  "testGroup",
			Flags: map[string]*dynflags.Flag{},
		}
		flag := group.Lookup("flag1")
		assert.Nil(t, flag)
	})
}

func TestConfigGroups(t *testing.T) {
	t.Parallel()

	t.Run("Lookup existing group", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		df.Group("http")

		groups := df.Config()
		group := groups.Lookup("http")

		assert.NotNil(t, group)
		assert.Equal(t, "http", group.Name)
	})

	t.Run("Iterate over groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		df.Group("http")
		df.Group("tcp")

		groups := df.Config().Groups()

		assert.Contains(t, groups, "http")
		assert.Contains(t, groups, "tcp")
		assert.Equal(t, "http", groups["http"].Name)
		assert.Equal(t, "tcp", groups["tcp"].Name)
	})
}

func TestConfigGroup_Lookup_NilHandling(t *testing.T) {
	t.Parallel()

	t.Run("Lookup on nil ConfigGroup returns nil", func(t *testing.T) {
		t.Parallel()

		var groupConfig *dynflags.ConfigGroup
		result := groupConfig.Lookup("flag1")
		assert.Nil(t, result, "Expected Lookup on nil ConfigGroup to return nil")
	})

	t.Run("Lookup non-existing flag returns nil", func(t *testing.T) {
		t.Parallel()

		groupConfig := &dynflags.ConfigGroup{
			Name:  "testGroup",
			Flags: map[string]*dynflags.Flag{},
		}

		result := groupConfig.Lookup("nonExistingFlag")
		assert.Nil(t, result, "Expected Lookup for non-existing flag to return nil")
	})
}

func TestConfigGroups_Lookup_NilHandling(t *testing.T) {
	t.Parallel()

	t.Run("Lookup on nil ConfigGroups returns nil", func(t *testing.T) {
		t.Parallel()

		var configGroups *dynflags.ConfigGroups
		result := configGroups.Lookup("group1")

		assert.Nil(t, result, "Expected Lookup on nil ConfigGroups to return nil")
	})

	t.Run("Lookup non-existing group returns nil", func(t *testing.T) {
		t.Parallel()

		configGroups := &dynflags.ConfigGroups{}
		result := configGroups.Lookup("nonExistingGroup")

		assert.Nil(t, result, "Expected Lookup for non-existing group to return nil")
	})
}

func TestConfigGroups_Groups_NilHandling(t *testing.T) {
	t.Parallel()

	t.Run("Groups on nil ConfigGroups returns nil", func(t *testing.T) {
		var configGroups *dynflags.ConfigGroups
		result := configGroups.Groups()

		assert.Nil(t, result, "Expected Groups on nil ConfigGroups to return nil")
	})
}

func TestDynFlags_Config_NilHandling(t *testing.T) {
	t.Parallel()

	t.Run("Config on nil DynFlags returns nil", func(t *testing.T) {
		var dynFlags *dynflags.DynFlags
		result := dynFlags.Config()

		assert.Nil(t, result, "Expected Config on nil DynFlags to return nil")
	})
}
