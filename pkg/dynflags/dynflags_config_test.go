package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestGroupConfig(t *testing.T) {
	t.Parallel()

	t.Run("Lookup existing flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.GroupConfig{
			Name:  "testGroup",
			Flags: map[string]*dynflags.Flag{"flag1": {Usage: "Test Flag"}},
		}
		flag, err := group.Lookup("flag1")
		assert.NoError(t, err)
		assert.Equal(t, "Test Flag", flag.Usage)
	})

	t.Run("Lookup non-existing flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.GroupConfig{
			Name:  "testGroup",
			Flags: map[string]*dynflags.Flag{},
		}
		flag, err := group.Lookup("flag1")
		assert.Error(t, err)
		assert.Nil(t, flag)
		assert.EqualError(t, err, "flag 'flag1' not found in config group 'testGroup'")
	})
}

func TestConfigGroups(t *testing.T) {
	t.Parallel()

	t.Run("Lookup existing group", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		df.Group("http")

		groups := df.Groups()
		group, err := groups.Lookup("http")
		assert.NoError(t, err)
		assert.Equal(t, "http", group.Name)
	})

	t.Run("Lookup non-existing group", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)

		groups := df.Groups()
		group, err := groups.Lookup("tcp")
		assert.Error(t, err)
		assert.Nil(t, group)
		assert.EqualError(t, err, "config group 'tcp' not found")
	})

	t.Run("Iterate over groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		df.Group("http")
		df.Group("tcp")

		groups := df.Groups().Iterate()

		assert.Contains(t, groups, "http")
		assert.Contains(t, groups, "tcp")
		assert.Equal(t, "http", groups["http"].Name)
		assert.Equal(t, "tcp", groups["tcp"].Name)
	})
}
