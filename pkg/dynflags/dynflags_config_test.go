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
		flag := group.Lookup("flag1")
		assert.NotNil(t, flag)
		assert.Equal(t, "Test Flag", flag.Usage)
	})

	t.Run("Lookup non-existing flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.GroupConfig{
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

	t.Run("Lookup non-existing group", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)

		groups := df.Config()
		group := groups.Lookup("tcp")
		assert.Nil(t, group)
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
