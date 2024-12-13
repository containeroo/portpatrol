package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestDynFlagsParse(t *testing.T) {
	t.Run("Parse flags with equals sign", func(t *testing.T) {
		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"--test.identifier.name=test name"})
		assert.NoError(t, err)

		parsedGroups := df.Parsed()
		assert.Contains(t, parsedGroups, "test")
		assert.Equal(t, "test name", parsedGroups["test"][0].Values["name"])
	})

	t.Run("Parse flags with space-separated value", func(t *testing.T) {
		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"--test.identifier.name", "test name"})
		assert.NoError(t, err)

		parsedGroups := df.Parsed()
		assert.Contains(t, parsedGroups, "test")
		assert.Equal(t, "test name", parsedGroups["test"][0].Values["name"])
	})

	t.Run("Handle missing value", func(t *testing.T) {
		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"--test.identifier.name"})
		assert.Error(t, err)
		assert.EqualError(t, err, "missing value for flag: test.identifier.name")
	})
}
