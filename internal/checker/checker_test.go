package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewChecker verifies the expected behavior.
func TestNewChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP checker", func(t *testing.T) {
		t.Parallel()

		check, err := DefaultHTTPConfig().NewChecker("example", "http://example.com")

		require.NoError(t, err)
		assert.Equal(t, check.Name(), "example")
		assert.Equal(t, check.Type(), "HTTP")
	})

	t.Run("Valid TCP checker", func(t *testing.T) {
		t.Parallel()

		check, err := DefaultTCPConfig().NewChecker("example", "example.com:80")

		require.NoError(t, err)
		assert.Equal(t, check.Name(), "example")
		assert.Equal(t, check.Type(), "TCP")
	})

	t.Run("Valid ICMP checker", func(t *testing.T) {
		t.Parallel()

		check, err := DefaultICMPConfig().NewChecker("example", "example.com")

		require.NoError(t, err)
		assert.Equal(t, check.Name(), "example")
		assert.Equal(t, check.Type(), "ICMP")
	})

}
