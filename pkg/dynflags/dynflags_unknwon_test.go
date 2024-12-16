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
		assert.Equal(t, "testGroup", group.Name)
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
		}

		df := dynflags.New(dynflags.ParseUnknown)
		df.Parse(args)

		unknownGroups := df.Unknown()

		group := unknownGroups.Lookup("group1")
		assert.NotNil(t, group)
		assert.Equal(t, "group1", group.Name)
		assert.Equal(t, "value1", group.Lookup("flag1"))
		assert.Equal(t, "value2", group.Lookup("flag2"))
	})

	t.Run("Handle no unknown groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ParseUnknown)
		unknownGroups := df.Unknown()

		group := unknownGroups.Lookup("nonExistentGroup")
		assert.Nil(t, group)
	})
}
