package factory_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCheckersRejectsInvalidRetrySettings verifies retry validation remains
// readable and centralized without changing its external behavior.
func TestBuildCheckersRejectsInvalidRetrySettings(t *testing.T) {
	t.Parallel()

	baseTarget := func() factory.TargetConfig {
		return factory.TargetConfig{
			ID:      targetID,
			Address: testHTTPAddress,
			HTTP:    &factory.HTTPConfig{Method: http.MethodGet},
		}
	}

	t.Run("negative default interval", func(t *testing.T) {
		t.Parallel()

		_, err := factory.BuildCheckers([]factory.TargetConfig{baseTarget()}, -time.Second, -1, testVersion, false)
		require.Error(t, err)
		assert.EqualError(t, err, "default interval must be non-negative")
	})

	for _, tt := range []struct {
		name        string
		maxAttempts int
	}{
		{name: "zero global max attempts", maxAttempts: 0},
		{name: "global max attempts below endless", maxAttempts: -2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := factory.BuildCheckers([]factory.TargetConfig{baseTarget()}, time.Second, tt.maxAttempts, testVersion, false)
			require.Error(t, err)
			assert.EqualError(t, err, "max attempts must be -1 or positive")
		})
	}

	t.Run("negative target interval", func(t *testing.T) {
		t.Parallel()

		target := baseTarget()
		target.Interval = -time.Second

		_, err := factory.BuildCheckers([]factory.TargetConfig{target}, time.Second, -1, testVersion, false)
		require.Error(t, err)
		assert.EqualError(t, err, "invalid retry settings for "+targetID)
	})

	t.Run("negative target max interval", func(t *testing.T) {
		t.Parallel()

		target := baseTarget()
		target.MaxInterval = -time.Second

		_, err := factory.BuildCheckers([]factory.TargetConfig{target}, time.Second, -1, testVersion, false)
		require.Error(t, err)
		assert.EqualError(t, err, "invalid retry settings for "+targetID)
	})

	t.Run("target max attempts below endless", func(t *testing.T) {
		t.Parallel()

		target := baseTarget()
		target.MaxAttempts = -2

		_, err := factory.BuildCheckers([]factory.TargetConfig{target}, time.Second, -1, testVersion, false)
		require.Error(t, err)
		assert.EqualError(t, err, "invalid retry settings for "+targetID)
	})

	t.Run("unsupported backoff", func(t *testing.T) {
		t.Parallel()

		target := baseTarget()
		target.Backoff = backoff.Mode("invalid")

		_, err := factory.BuildCheckers([]factory.TargetConfig{target}, time.Second, -1, testVersion, false)
		require.Error(t, err)
		assert.EqualError(t, err, "invalid backoff for "+targetID)
	})
}
