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

		value, err := group.Lookup("flag1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", value)
	})

	t.Run("Lookup non-existing unknown flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.UnknownGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}

		value, err := group.Lookup("flag1")
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.EqualError(t, err, "flag 'flag1' not found in unknown group 'testGroup'")
	})
}

func TestUnknownGroups(t *testing.T) {
	t.Parallel()

	t.Run("Lookup existing unknown group", func(t *testing.T) {
		t.Parallel()

		unknownGroups := &dynflags.UnknownGroups{
			groups: map[string]*dynflags.UnknownGroup{
				"testGroup": {
					Name:   "testGroup",
					Values: map[string]interface{}{"flag1": "value1"},
				},
			},
		}

		group, err := unknownGroups.Lookup("testGroup")
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, "testGroup", group.Name)
	})

	t.Run("Lookup non-existing unknown group", func(t *testing.T) {
		t.Parallel()

		unknownGroups := &dynflags.UnknownGroups{
			groups: map[string]*dynflags.UnknownGroup{},
		}

		group, err := unknownGroups.Lookup("nonExistentGroup")
		assert.Error(t, err)
		assert.Nil(t, group)
		assert.EqualError(t, err, "unknown group 'nonExistentGroup' not found")
	})
}

func TestDynFlagsUnknown(t *testing.T) {
	t.Parallel()

	t.Run("Combine unknown groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.IgnoreUnknown)
		df.unknownGroups = map[string][]*dynflags.UnknownGroup{
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

		unknownGroups := df.Unknown()

		group, err := unknownGroups.Lookup("group1")
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, "group1", group.Name)
		assert.Equal(t, "value1", group.Values["flag1"])
		assert.Equal(t, "value2", group.Values["flag2"])
	})

	t.Run("Handle no unknown groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.IgnoreUnknown)
		unknownGroups := df.Unknown()

		group, err := unknownGroups.Lookup("nonExistentGroup")
		assert.Error(t, err)
		assert.Nil(t, group)
		assert.EqualError(t, err, "unknown group 'nonExistentGroup' not found")
	})
}
