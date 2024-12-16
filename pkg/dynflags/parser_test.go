package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestDynFlagsParse(t *testing.T) {
	t.Parallel()

	t.Run("Parse flags with equals sign", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"--test.identifier.name=test name"})
		assert.NoError(t, err)

		parsedGroups := df.Parsed()
		assert.Equal(t, "test name", parsedGroups.Lookup("test").Lookup("identifier").Lookup("name"))
	})

	t.Run("Parse flags with space-separated value", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"--test.identifier.name", "test name"})
		assert.NoError(t, err)

		parsedGroups := df.Parsed()
		assert.Equal(t, "test name", parsedGroups.Lookup("test").Lookup("identifier").Lookup("name"))
	})

	t.Run("Handle missing value", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"--test.identifier.name"})
		assert.Error(t, err)
		assert.EqualError(t, err, "missing value for flag: test.identifier.name")
	})

	t.Run("Invalid flag format", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ExitOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"-test.identifier.name"})
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid flag format: -test.identifier.name")
	})

	t.Run("Split key with invalid flag format", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"--test.name", "value"})
		assert.Error(t, err)
		assert.EqualError(t, err, "flag must follow the pattern: --<group>.<identifier>.<flag>=value")
	})

	t.Run("Unknown parent group", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ExitOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"--unknown.identifier.name", "value"})
		assert.Error(t, err)
		assert.EqualError(t, err, "unknown group: 'unknown'")
	})

	t.Run("Unknown flag, exit on error", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ExitOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"--test.identifier.badflag", "value"})
		assert.Error(t, err)
		assert.EqualError(t, err, "unknown flag 'badflag' in group 'test'")
	})

	t.Run("Unknown flag, continue on error", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("test")
		group.String("name", "", "Test name flag")

		err := df.Parse([]string{"--test.identifier.badflag", "value"})
		assert.NoError(t, err)
	})

	t.Run("Set value error", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ExitOnError)
		group := df.Group("test")
		group.Int("name", 0, "Test name flag")

		args := []string{"--test.identifier.name", "value"}
		err := df.Parse(args)
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to parse value for flag 'name': strconv.Atoi: parsing \"value\": invalid syntax")
	})
}
