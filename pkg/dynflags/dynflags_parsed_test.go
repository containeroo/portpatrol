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
		df.Parse(args)

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
		df.Parse(args)

		parsedGroups := df.Parsed()

		group := parsedGroups.Lookup("group1")
		assert.NotNil(t, group)
		assert.Equal(t, "group1", group.Name)
		assert.Equal(t, "value1", group.Lookup("flag1"))
		assert.Equal(t, "value2", group.Lookup("flag2"))
	})

	t.Run("Handle no parsed groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		parsedGroups := df.Parsed()

		group := parsedGroups.Lookup("nonExistentGroup")
		assert.Nil(t, group)
	})
}
