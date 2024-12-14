package dynflags_test

import (
	"bytes"
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestDynFlagsInitialization(t *testing.T) {
	t.Parallel()

	t.Run("New initializes correctly", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		assert.NotNil(t, df)
		assert.NotNil(t, df.Groups())
		assert.NotNil(t, df.Parsed())
		assert.NotNil(t, df.Unknown())
	})
}

func TestDynFlagsGroupManagement(t *testing.T) {
	t.Parallel()

	t.Run("Create new group", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("group1")

		assert.NotNil(t, group)
		assert.Contains(t, df.Groups(), "group1")
		assert.Equal(t, group, df.Groups()["group1"])
		assert.Equal(t, "group1", group.Name)
		assert.NotNil(t, group.Flags)
	})

	t.Run("Duplicate group panics", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		df.Group("group1")

		assert.Panics(t, func() {
			df.Group("group1")
		}, "Expected panic for duplicate group creation")
	})
}

func TestDynFlagsUsageOutput(t *testing.T) {
	t.Parallel()

	t.Run("Generate usage with title, description, and epilog", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)

		df.Title("Test Application")
		df.Description("This application demonstrates usage of dynamic flags.")
		df.Epilog("For more information, visit https://example.com.")

		df.Usage()

		output := buf.String()
		assert.Contains(t, output, "Test Application")
		assert.Contains(t, output, "This application demonstrates usage of dynamic flags.")
		assert.Contains(t, output, "For more information, visit https://example.com.")
	})
}

func TestDynFlagsParsedAndUnknown(t *testing.T) {
	t.Parallel()

	t.Run("Empty parsed and unknown groups", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)

		assert.Empty(t, df.Parsed())
		assert.Empty(t, df.Unknown())
	})
}

func TestParsedGroupMethods(t *testing.T) {
	t.Parallel()

	t.Run("Get unknown values", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.IgnoreUnknown)
		df.Group("known")
		args := []string{"--unknown.identifier.value", "value 1"}
		err := df.Parse(args)
		assert.NoError(t, err)

		// Retrieve the unknown value
		unknownGroups := df.Unknown()
		u := unknownGroups.Lookup("unknown")
		v := u.Lookup("value")
		assert.NoError(t, err)
		assert.Equal(t, "value 1", v)
	})

	t.Run("Get flag value", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"flag1": "value1",
			},
		}

		value := parsedGroup.GetValue("flag1")
		assert.Equal(t, "value1", value)
	})

	t.Run("Get non-existent flag value", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Values: map[string]interface{}{},
		}

		value := parsedGroup.GetValue("nonExistentFlag")
		assert.Nil(t, value)
	})
}
