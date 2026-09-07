package cli

import (
	"testing"

	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertInvalidFlagValueError verifies an enum parse error names the flag, rejected value, and allowed values.
func assertInvalidFlagValueError(t *testing.T, err error, flag string, value string, allowed ...string) {
	t.Helper()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value for flag "+flag)
	assert.Contains(t, err.Error(), "\""+value+"\"")
	assert.Contains(t, err.Error(), "must be one of")
	for _, allowedValue := range allowed {
		assert.Contains(t, err.Error(), allowedValue)
	}
}

func requireHTTPConfig(t *testing.T, target factory.TargetConfig) checker.HTTPConfig {
	t.Helper()
	cfg, ok := target.Config.(checker.HTTPConfig)
	require.True(t, ok)
	return cfg
}

func requireTCPConfig(t *testing.T, target factory.TargetConfig) checker.TCPConfig {
	t.Helper()
	cfg, ok := target.Config.(checker.TCPConfig)
	require.True(t, ok)
	return cfg
}

func requireICMPConfig(t *testing.T, target factory.TargetConfig) checker.ICMPConfig {
	t.Helper()
	cfg, ok := target.Config.(checker.ICMPConfig)
	require.True(t, ok)
	return cfg
}
