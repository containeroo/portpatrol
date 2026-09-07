package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckerConstructorsTrustPrevalidatedAddresses documents the checker boundary:
// constructors normalize addresses but do not duplicate CLI validation.
func TestCheckerConstructorsTrustPrevalidatedAddresses(t *testing.T) {
	t.Parallel()

	t.Run("HTTP", func(t *testing.T) {
		t.Parallel()

		check, err := NewHTTPChecker("http", " ://invalid-url ", DefaultHTTPConfig())
		require.NoError(t, err)
		assert.Equal(t, "://invalid-url", check.Address())
	})

	t.Run("TCP", func(t *testing.T) {
		t.Parallel()

		check, err := NewTCPChecker("tcp", " invalid-address ", DefaultTCPConfig())
		require.NoError(t, err)
		assert.Equal(t, "invalid-address", check.Address())
	})

	t.Run("ICMP", func(t *testing.T) {
		t.Parallel()

		check, err := NewICMPChecker("icmp", " icmp://invalid ", DefaultICMPConfig())
		require.NoError(t, err)
		assert.Equal(t, "icmp://invalid", check.Address())
	})
}
