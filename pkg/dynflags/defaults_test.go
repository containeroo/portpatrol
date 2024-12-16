package dynflags_test

import (
	"bytes"
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestPrintDefaults(t *testing.T) {
	t.Parallel()

	t.Run("No groups, title, description, or epilog", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)

		df.PrintDefaults()

		output := buf.String()
		assert.Empty(t, output)
	})

	t.Run("Only title is present", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)
		df.Title("Test Title")

		df.PrintDefaults()

		output := buf.String()
		assert.Contains(t, output, "Test Title")
		assert.NotContains(t, output, "Usage:")
	})

	t.Run("Only description is present", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)
		df.Description("Test Description")

		df.PrintDefaults()

		output := buf.String()
		assert.Contains(t, output, "Test Description")
		assert.NotContains(t, output, "Usage:")
	})

	t.Run("Only epilog is present", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)
		df.Epilog("Test Epilog")

		df.PrintDefaults()

		output := buf.String()
		assert.Contains(t, output, "Test Epilog")
	})

	t.Run("Title, description, and epilog are all present", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)
		df.Title("Test Title")
		df.Description("Test Description")
		df.Epilog("Test Epilog")

		df.PrintDefaults()

		output := buf.String()
		assert.Contains(t, output, "Test Title")
		assert.Contains(t, output, "Test Description")
		assert.Contains(t, output, "Test Epilog")
	})

	t.Run("Single group with unsorted flags", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)
		group := df.Group("test")
		group.String("flag2", "", "Second flag")
		group.String("flag1", "", "First flag")

		df.PrintDefaults()

		output := buf.String()
		assert.Contains(t, output, "TEST")
		assert.Contains(t, output, "--test.<IDENTIFIER>.flag2")
		assert.Contains(t, output, "--test.<IDENTIFIER>.flag1")
	})

	t.Run("Multiple groups with sorted flags", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)
		df.SortFlags = true
		df.SortGroups = true
		group1 := df.Group("test1")
		group1.String("flagA", "", "Flag A")
		group1.String("flagB", "", "Flag B")

		group2 := df.Group("test2")
		group2.String("flagX", "", "Flag X")

		df.PrintDefaults()

		output := buf.String()
		assert.Contains(t, output, "TEST1")
		assert.Contains(t, output, "--test1.<IDENTIFIER>.flagA")
		assert.Contains(t, output, "--test1.<IDENTIFIER>.flagB")
		assert.Contains(t, output, "TEST2")
		assert.Contains(t, output, "--test2.<IDENTIFIER>.flagX")
	})

	t.Run("Group with usage text", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)
		group := df.Group("test")
		group.Usage = "Test Group Usage"
		group.String("flag", "", "Test flag")

		df.PrintDefaults()

		output := buf.String()
		assert.Contains(t, output, "Test Group Usage")
		assert.Contains(t, output, "--test.<IDENTIFIER>.flag")
	})

	t.Run("Flags with and without default values", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)
		group := df.Group("test")
		group.String("flag1", "default1", "Flag with default")
		group.String("flag2", "", "Flag without default")

		df.PrintDefaults()

		output := buf.String()
		assert.Contains(t, output, "--test.<IDENTIFIER>.flag1")
		assert.Contains(t, output, "(default: default1)")
		assert.Contains(t, output, "--test.<IDENTIFIER>.flag2")
		assert.NotContains(t, output, "(default: )")
	})

	t.Run("Empty group with no flags", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		df := dynflags.New(dynflags.ContinueOnError)
		df.SetOutput(&buf)
		df.Group("test")

		df.PrintDefaults()

		output := buf.String()
		assert.Contains(t, output, "TEST")
		assert.NotContains(t, output, "Flag\tUsage")
	})
}
