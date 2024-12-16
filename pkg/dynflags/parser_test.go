package dynflags_test

import (
	"testing"
	"time"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestDynFlagsParse(t *testing.T) {
	t.Parallel()

	t.Run("Parse valid arguments", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("http")
		group.String("method", "GET", "HTTP method to use")
		group.String("url", "", "Target URL")

		args := []string{
			"--http.identifier1.method", "POST",
			"--http.identifier1.url=https://example.com",
		}
		err := df.Parse(args)
		assert.NoError(t, err)

		parsedGroups := df.Parsed()
		httpGroup := parsedGroups.Lookup("http")
		assert.NotNil(t, httpGroup)

		identifier1 := httpGroup.Lookup("identifier1")
		assert.NotNil(t, identifier1)
		assert.Equal(t, "POST", identifier1.Lookup("method"))
		assert.Equal(t, "https://example.com", identifier1.Lookup("url"))
	})

	t.Run("Exit on missing key", func(t *testing.T) {
		df := dynflags.New(dynflags.ExitOnError)
		group := df.Group("http")
		group.String("method", "GET", "HTTP method to use")

		args := []string{
			"-http.identifier1", "https://example.com",
		}
		err := df.Parse(args)
		assert.Error(t, err)
	})

	t.Run("Parse with missing value", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("http")
		group.String("method", "GET", "HTTP method to use")

		args := []string{
			"--http.identifier1.method",
		}
		err := df.Parse(args)
		assert.NoError(t, err)

		unparsedArgs := df.UnparsedArgs()
		assert.Contains(t, unparsedArgs, "--http.identifier1.method")
	})

	t.Run("Parse with wrong value type and continue", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("http")
		group.Duration("timeout", 10*time.Second, "HTTP timeout")

		args := []string{
			"--http.identifier1.timeout", "1",
		}
		err := df.Parse(args)
		assert.NoError(t, err)

		unparsedArgs := df.UnparsedArgs()
		assert.Contains(t, unparsedArgs, "--http.identifier1.timeout")
	})

	t.Run("Parse with invalid flag format", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)

		args := []string{
			"-invalidFlag",
		}
		err := df.Parse(args)
		assert.NoError(t, err)

		unparsedArgs := df.UnparsedArgs()
		assert.Contains(t, unparsedArgs, "-invalidFlag")
	})

	t.Run("Parse with no identifier and exit", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ExitOnError)
		group1 := df.Group("http")
		group1.Duration("timeout", 10*time.Second, "HTTP timeout")

		args := []string{
			"--http.duration", "10s",
		}
		err := df.Parse(args)
		assert.Error(t, err)
	})

	t.Run("Parse with unknown group and continue on error", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ParseUnknown)

		args := []string{
			"--unknown.identifier1.flag1", "value1",
		}
		err := df.Parse(args)
		assert.NoError(t, err)

		unknownGroups := df.Unknown()
		unknownGroup := unknownGroups.Lookup("unknown")
		assert.NotNil(t, unknownGroup)

		identifier1 := unknownGroup.Lookup("identifier1")
		assert.NotNil(t, identifier1)
		assert.Equal(t, "value1", identifier1.Lookup("flag1"))
	})

	t.Run("Parse with unknown group and parse unknown behavior", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ParseUnknown)

		args := []string{
			"--unknown.identifier1.flag1", "value1",
		}
		err := df.Parse(args)
		assert.NoError(t, err)

		unknownGroups := df.Unknown()
		unknownGroup := unknownGroups.Lookup("unknown")
		assert.NotNil(t, unknownGroup)

		identifier1 := unknownGroup.Lookup("identifier1")
		assert.NotNil(t, identifier1)
		assert.Equal(t, "value1", identifier1.Lookup("flag1"))
	})

	t.Run("Parse with unknown group and exit on error", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ExitOnError)

		args := []string{
			"--unknown.identifier1.flag1", "value1",
		}
		err := df.Parse(args)
		assert.Error(t, err)
		assert.EqualError(t, err, "unknown flag 'flag1' in group 'unknown'")
	})

	t.Run("Handle invalid key format", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)

		args := []string{
			"--invalidformat",
		}
		err := df.Parse(args)
		assert.NoError(t, err)

		unparsedArgs := df.UnparsedArgs()
		assert.Contains(t, unparsedArgs, "--invalidformat")
	})

	t.Run("Handle missing flag value", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		group := df.Group("http")
		group.String("method", "GET", "HTTP method to use")

		args := []string{
			"--http.identifier1.method",
		}
		err := df.Parse(args)
		assert.NoError(t, err)

		unparsedArgs := df.UnparsedArgs()
		assert.Contains(t, unparsedArgs, "--http.identifier1.method")
	})
}
