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

		value, err := group.Lookup("flag1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", value)
	})

	t.Run("Lookup non-existing parsed flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}

		value, err := group.Lookup("flag1")
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.EqualError(t, err, "flag 'flag1' not found in parsed group 'testGroup'")
	})
}

func TestParsedGroups(t *testing.T) {
	t.Parallel()

	t.Run("Lookup existing parsed group", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		args := []string{"--testgroup.identifier1.flag1", "value1"}
		df.Parse(args)

		group := df.Group("testGroup")
		assert.NotNil(t, group)
		assert.Equal(t, "testGroup", group.Name)
	})

	t.Run("Lookup non-existing parsed group", func(t *testing.T) {
		t.Parallel()

		parsedGroups := &dynflags.ParsedGroups{
			groups: map[string]*dynflags.ParsedGroup{},
		}

		group, err := parsedGroups.Lookup("nonExistentGroup")
		assert.Error(t, err)
		assert.Nil(t, group)
		assert.EqualError(t, err, "parsed group 'nonExistentGroup' not found")
	})
}

func TestDynFlagsParsed(t *testing.T) {
	t.Parallel()

	t.Run("Combine parsed groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		df.parsedGroups = map[string][]*dynflags.ParsedGroup{
			"group1": {
				{
					Name:   "identifier1",
					Values: map[string]interface{}{"flag1": "value1"},
				},
				{
					Name:   "identifier2",
					Values: map[string]interface{}{"flag2": "value2"},
				},
			},
		}

		parsedGroups := df.Parsed()

		group, err := parsedGroups.Lookup("group1")
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, "group1", group.Name)
		assert.Equal(t, "value1", group.Values["flag1"])
		assert.Equal(t, "value2", group.Values["flag2"])
	})

	t.Run("Handle no parsed groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		parsedGroups := df.Parsed()

		group, err := parsedGroups.Lookup("nonExistentGroup")
		assert.Error(t, err)
		assert.Nil(t, group)
		assert.EqualError(t, err, "parsed group 'nonExistentGroup' not found")
	})
}
