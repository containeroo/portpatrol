package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestParsedGroup(t *testing.T) {
	t.Parallel()

	t.Run("Lookup existing parsed flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": "value1"},
		}

		value := group.Lookup("flag1")
		assert.Equal(t, "value1", value)
	})

	t.Run("Lookup non-existing parsed flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}

		value := group.Lookup("flag1")
		assert.Nil(t, value)
	})
}

func TestParsedGroups(t *testing.T) {
	t.Parallel()

	t.Run("Lookup existing parsed group", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		args := []string{"--testgroup.identifier1.flag1", "value1"}
		err := df.Parse(args)
		assert.NoError(t, err)

		group := df.Group("testGroup")
		assert.NotNil(t, group)
		assert.Equal(t, "testGroup", group.Name)
	})

	t.Run("Lookup non-existing parsed group", func(t *testing.T) {
		t.Parallel()

		parsedGroups := &dynflags.ParsedGroups{}

		group := parsedGroups.Lookup("nonExistentGroup")
		assert.Nil(t, group)
	})
}

func TestDynFlagsParsed(t *testing.T) {
	t.Parallel()

	t.Run("Combine parsed groups", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--group1.identifier1.flag1", "value1",
			"--group1.identifier2.flag2", "value2",
		}

		df := dynflags.New(dynflags.ContinueOnError)
		g1 := df.Group("group1")
		g1.String("flag1", "", "Description flag1")
		g1.String("flag2", "", "Description flag2")

		err := df.Parse(args)
		assert.NoError(t, err)

		parsedGroups := df.Parsed()

		group := parsedGroups.Lookup("group1")
		assert.NotNil(t, group)
		assert.Equal(t, "group1", group.Name)
		assert.Equal(t, "value1", group.Lookup("identifier1").Lookup("flag1"))
		assert.Equal(t, "value2", group.Lookup("identifier2").Lookup("flag2"))
	})

	t.Run("Handle no parsed groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		parsedGroups := df.Parsed()

		group := parsedGroups.Lookup("nonExistentGroup")
		assert.Nil(t, group)
	})
}

func TestParsedGroups_Lookup_NilHandling(t *testing.T) {
	t.Parallel()

	t.Run("Lookup on nil ParsedGroups returns nil", func(t *testing.T) {
		t.Parallel()

		var parsedGroups *dynflags.ParsedGroups
		result := parsedGroups.Lookup("http")
		assert.Nil(t, result, "Expected Lookup on nil ParsedGroups to return nil")
	})

	t.Run("Lookup non-existing group returns nil", func(t *testing.T) {
		t.Parallel()

		parsedGroups := &dynflags.ParsedGroups{}

		result := parsedGroups.Lookup("nonExistingGroup")
		assert.Nil(t, result, "Expected Lookup for non-existing group to return nil")
	})
}

func TestParsedIdentifiers_Lookup_NilHandling(t *testing.T) {
	t.Parallel()

	t.Run("Lookup on nil ParsedIdentifiers returns nil", func(t *testing.T) {
		t.Parallel()

		var parsedIdentifiers *dynflags.ParsedIdentifiers
		result := parsedIdentifiers.Lookup("identifier1")
		assert.Nil(t, result, "Expected Lookup on nil ParsedIdentifiers to return nil")
	})

	t.Run("Lookup non-existing identifier returns nil", func(t *testing.T) {
		t.Parallel()

		parsedIdentifiers := &dynflags.ParsedIdentifiers{}

		result := parsedIdentifiers.Lookup("nonExistingIdentifier")
		assert.Nil(t, result, "Expected Lookup for non-existing identifier to return nil")
	})
}

func TestParsedGroup_Lookup_NilHandling(t *testing.T) {
	t.Parallel()

	t.Run("Lookup on nil ParsedGroup returns nil", func(t *testing.T) {
		t.Parallel()

		var parsedGroup *dynflags.ParsedGroup
		result := parsedGroup.Lookup("flag1")
		assert.Nil(t, result, "Expected Lookup on nil ParsedGroup to return nil")
	})

	t.Run("Lookup non-existing flag returns nil", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "identifier1",
			Values: map[string]interface{}{},
		}

		result := parsedGroup.Lookup("nonExistingFlag")
		assert.Nil(t, result, "Expected Lookup for non-existing flag to return nil")
	})
}
