package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestUnknownGroup(t *testing.T) {
	t.Parallel()

	t.Run("Lookup existing unknown flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.UnknownGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": "value1"},
		}

		value := group.Lookup("flag1")
		assert.Equal(t, "value1", value)
	})

	t.Run("Lookup non-existing unknown flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.UnknownGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}

		value := group.Lookup("flag1")
		assert.Nil(t, value)
	})
}

func TestUnknownGroups(t *testing.T) {
	t.Parallel()

	t.Run("Lookup existing unknown group", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--unknown.identifier1.flag1", "value1",
		}

		df := dynflags.New(dynflags.ParseUnknown)
		err := df.Parse(args)
		assert.NoError(t, err)

		unknownGroups := df.Unknown()

		group := unknownGroups.Lookup("unknown")
		assert.NotNil(t, group)
		assert.Equal(t, "unknown", group.Name)
	})

	t.Run("Lookup non-existing unknown group", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--unknown.identifier1.flag1", "value1",
		}

		df := dynflags.New(dynflags.ContinueOnError)
		err := df.Parse(args)
		assert.NoError(t, err)

		unknownGroups := df.Unknown()
		group := unknownGroups.Lookup("unknown")
		assert.Nil(t, group)
	})
}

func TestDynFlagsUnknown(t *testing.T) {
	t.Parallel()

	t.Run("Combine unknown groups", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--group1.identifier1.flag1", "value1",
			"--group1.identifier2.flag2", "value2",
		}

		df := dynflags.New(dynflags.ParseUnknown)
		err := df.Parse(args)
		assert.NoError(t, err)

		unknownGroups := df.Unknown()

		group := unknownGroups.Lookup("group1")
		assert.NotNil(t, group)
		assert.Equal(t, "group1", group.Name)
		assert.Equal(t, "value1", group.Lookup("identifier1").Lookup("flag1"))
		assert.Equal(t, "value2", group.Lookup("identifier2").Lookup("flag2"))
	})

	t.Run("Handle no unknown groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ParseUnknown)
		unknownGroups := df.Unknown()

		group := unknownGroups.Lookup("nonExistentGroup")
		assert.Nil(t, group)
	})
}

func TestUnknownGroups_Lookup_NilHandling(t *testing.T) {
	t.Parallel()

	t.Run("Lookup on nil UnknownGroups returns nil", func(t *testing.T) {
		t.Parallel()

		var unknownGroups *dynflags.UnknownGroups
		result := unknownGroups.Lookup("http")
		assert.Nil(t, result, "Expected Lookup on nil UnknownGroups to return nil")
	})

	t.Run("Lookup non-existing group returns nil", func(t *testing.T) {
		t.Parallel()

		unknownGroups := &dynflags.UnknownGroups{}

		result := unknownGroups.Lookup("nonExistingGroup")
		assert.Nil(t, result, "Expected Lookup for non-existing group to return nil")
	})
}

func TestUnknownIdentifiers_Lookup_NilHandling(t *testing.T) {
	t.Parallel()

	t.Run("Lookup on nil UnknownIdentifiers returns nil", func(t *testing.T) {
		t.Parallel()

		var unknownIdentifiers *dynflags.UnknownIdentifiers
		result := unknownIdentifiers.Lookup("identifier1")
		assert.Nil(t, result, "Expected Lookup on nil UnknownIdentifiers to return nil")
	})

	t.Run("Lookup non-existing identifier returns nil", func(t *testing.T) {
		t.Parallel()

		unknownIdentifiers := &dynflags.UnknownIdentifiers{}

		result := unknownIdentifiers.Lookup("nonExistingIdentifier")
		assert.Nil(t, result, "Expected Lookup for non-existing identifier to return nil")
	})
}

func TestUnknownGroup_Lookup_NilHandling(t *testing.T) {
	t.Parallel()

	t.Run("Lookup on nil UnknownGroup returns nil", func(t *testing.T) {
		t.Parallel()

		var unknownGroup *dynflags.UnknownGroup
		result := unknownGroup.Lookup("flag1")
		assert.Nil(t, result, "Expected Lookup on nil UnknownGroup to return nil")
	})

	t.Run("Lookup non-existing flag returns nil", func(t *testing.T) {
		t.Parallel()

		unknownGroup := &dynflags.UnknownGroup{
			Name:   "identifier1",
			Values: map[string]interface{}{},
		}

		result := unknownGroup.Lookup("nonExistingFlag")
		assert.Nil(t, result, "Expected Lookup for non-existing flag to return nil")
	})
}
