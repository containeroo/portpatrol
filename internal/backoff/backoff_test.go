package backoff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNextIntervalLinearKeepsBaseInterval verifies the expected behavior.
func TestNextIntervalLinearKeepsBaseInterval(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 100*time.Millisecond, NextInterval(ModeLinear, 100*time.Millisecond, 5, 0))
}

// TestNextIntervalExponentialDoublesByAttempt verifies the expected behavior.
func TestNextIntervalExponentialDoublesByAttempt(t *testing.T) {
	t.Parallel()

	t.Run("first failure uses base interval", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 100*time.Millisecond, NextInterval(ModeExponential, 100*time.Millisecond, 1, 0))
	})

	t.Run("second failure doubles interval", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 200*time.Millisecond, NextInterval(ModeExponential, 100*time.Millisecond, 2, 0))
	})

	t.Run("third failure doubles again", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 400*time.Millisecond, NextInterval(ModeExponential, 100*time.Millisecond, 3, 0))
	})
}

// TestNextIntervalExponentialCapsAtMaxInterval verifies the expected behavior.
func TestNextIntervalExponentialCapsAtMaxInterval(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 500*time.Millisecond, NextInterval(ModeExponential, 100*time.Millisecond, 10, 500*time.Millisecond))
}
