package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseFlagsResolvesAndValidatesTargetAddresses verifies resolver-backed addresses
// are resolved and validated before they leave the CLI boundary.
func TestParseFlagsResolvesAndValidatesTargetAddresses(t *testing.T) {
	t.Run("HTTP", func(t *testing.T) {
		t.Setenv("NEVER_TEST_HTTP_ADDRESS", "https://example.com/ready")

		cfg, err := ParseFlags([]string{"--http.web.address=env:NEVER_TEST_HTTP_ADDRESS"}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, cfg.Targets, 1)
		assert.Equal(t, "https://example.com/ready", cfg.Targets[0].Address)
	})

	t.Run("TCP", func(t *testing.T) {
		t.Setenv("NEVER_TEST_TCP_ADDRESS", "example.com:443")

		cfg, err := ParseFlags([]string{"--tcp.db.address=env:NEVER_TEST_TCP_ADDRESS"}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, cfg.Targets, 1)
		assert.Equal(t, "example.com:443", cfg.Targets[0].Address)
	})

	t.Run("ICMP", func(t *testing.T) {
		t.Setenv("NEVER_TEST_ICMP_ADDRESS", "example.com")

		cfg, err := ParseFlags([]string{"--icmp.host.address=env:NEVER_TEST_ICMP_ADDRESS"}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, cfg.Targets, 1)
		assert.Equal(t, "example.com", cfg.Targets[0].Address)
	})

	t.Run("invalid resolved address", func(t *testing.T) {
		t.Setenv("NEVER_TEST_ICMP_ADDRESS", "example.com:80")

		_, err := ParseFlags([]string{"--icmp.host.address=env:NEVER_TEST_ICMP_ADDRESS"}, "1.0.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ICMP address must be a hostname or IP without path or port")
	})
}
