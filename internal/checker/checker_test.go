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

		check, err := NewHTTPChecker("example", "http://example.com", DefaultHTTPConfig())

		require.NoError(t, err)
		assert.Equal(t, check.Name(), "example")
		assert.Equal(t, check.Type(), "HTTP")
	})

	t.Run("Valid TCP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewTCPChecker("example", "example.com:80", DefaultTCPConfig())

		require.NoError(t, err)
		assert.Equal(t, check.Name(), "example")
		assert.Equal(t, check.Type(), "TCP")
	})

	t.Run("Valid ICMP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewICMPChecker("example", "example.com", DefaultICMPConfig())

		require.NoError(t, err)
		assert.Equal(t, check.Name(), "example")
		assert.Equal(t, check.Type(), "ICMP")
	})

}

// TestParseCheckType verifies the expected behavior.
func TestParseCheckType(t *testing.T) {
	t.Parallel()

	t.Run("Check type HTTP", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("HTTP")

		require.NoError(t, err)
		assert.Equal(t, result, HTTP)
	})

	t.Run("Check type http", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("http")

		require.NoError(t, err)
		assert.Equal(t, result, HTTP)
	})

	t.Run("Check type TCP", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("tcp")

		require.NoError(t, err)
		assert.Equal(t, result, TCP)
	})

	t.Run("Check type tcp", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("tcp")

		require.NoError(t, err)
		assert.Equal(t, result, TCP)
	})

	t.Run("Check type ICMP", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("ICMP")

		require.NoError(t, err)
		assert.Equal(t, result, ICMP)
	})

	t.Run("Check type icmp", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("icmp")

		require.NoError(t, err)
		assert.Equal(t, result, ICMP)
	})

	t.Run("Invalid check type", func(t *testing.T) {
		t.Parallel()

		_, err := ParseCheckType("invalid")

		require.Error(t, err)
		assert.EqualError(t, err, "unsupported check type: invalid")
	})
}

// TestOverrideAddress verifies logging can expose a different address without changing checker behavior.
func TestOverrideAddress(t *testing.T) {
	t.Parallel()

	check, err := NewHTTPChecker("example", "http://example.com/private", DefaultHTTPConfig())
	require.NoError(t, err)

	overridden := OverrideAddress(check, "http://example.com")
	assert.Equal(t, "http://example.com", overridden.Address())
	assert.Equal(t, check.Name(), overridden.Name())
	assert.Equal(t, check.Type(), overridden.Type())
}
