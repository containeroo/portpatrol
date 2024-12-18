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
		assert.NotNil(t, df.Config())
		assert.NotNil(t, df.Parsed())
	})
}

func TestDynFlagsGroupManagement(t *testing.T) {
	t.Parallel()

	t.Run("Create new group", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)

		// Create Group
		group := df.Group("group1")
		assert.NotNil(t, group)
		assert.Contains(t, df.Config().Groups(), "group1")
		assert.Equal(t, group, df.Config().Lookup("group1"))
		assert.Equal(t, "group1", group.Name)
		assert.NotNil(t, group.Flags)

		// Get Group again
		group = df.Group("group1")
		assert.NotNil(t, group)
		assert.Contains(t, df.Config().Groups(), "group1")
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

		assert.Empty(t, df.Parsed().Groups())
		assert.Empty(t, df.UnknownArgs())
	})
}

func TestParsedGroupMethods(t *testing.T) {
	t.Parallel()

	t.Run("Retrieve parsed group values", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		df.Group("testGroup").String("flag1", "defaultValue", "Test flag")
		args := []string{"--testGroup.identifier1.flag1", "value1"}
		err := df.Parse(args)
		assert.NoError(t, err)

		parsedGroups := df.Parsed()
		group := parsedGroups.Lookup("testGroup")
		assert.NotNil(t, group)

		identifier := group.Lookup("identifier1")
		assert.NotNil(t, identifier)
		assert.Equal(t, "value1", identifier.Lookup("flag1"))
	})
}

func TestDynFlagsUnknownArgs(t *testing.T) {
	t.Parallel()

	t.Run("Retrieve unparsed arguments", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		args := []string{
			"--unparsable", "value1",
		}
		err := df.Parse(args)
		assert.NoError(t, err)

		unparsedArgs := df.UnknownArgs()
		assert.Contains(t, unparsedArgs, "--unparsable")
	})
}
