package factory_test

import (
	"testing"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCheckersRetrySettings verifies the factory preserves already-resolved retry settings.
func TestBuildCheckersRetrySettings(t *testing.T) {
	t.Parallel()

	t.Run("endless retries", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:          targetID,
				Address:     testHTTPAddress,
				MaxAttempts: 0,
				Config:      checker.DefaultHTTPConfig(),
			},
		}, 3*time.Second)
		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, 3*time.Second, checkers[0].Interval)
		assert.Zero(t, checkers[0].MaxAttempts)
		assert.Equal(t, backoff.ModeLinear, checkers[0].Backoff)
		assert.Zero(t, checkers[0].MaxInterval)
	})

	t.Run("finite retries", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:          targetID,
				Address:     testHTTPAddress,
				Interval:    2 * time.Second,
				MaxAttempts: 3,
				Backoff:     backoff.ModeExponential,
				MaxInterval: 30 * time.Second,
				Config:      checker.DefaultHTTPConfig(),
			},
		}, 3*time.Second)
		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, 2*time.Second, checkers[0].Interval)
		assert.Equal(t, 3, checkers[0].MaxAttempts)
		assert.Equal(t, backoff.ModeExponential, checkers[0].Backoff)
		assert.Equal(t, 30*time.Second, checkers[0].MaxInterval)
	})
}
