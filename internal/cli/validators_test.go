package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateHTTPStatusCodes verifies HTTP status code validation accepts supported expressions.
func TestValidateHTTPStatusCodes(t *testing.T) {
	t.Parallel()

	t.Run("single code", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPStatusCodes("200"))
	})

	t.Run("comma list", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPStatusCodes("200,204,301"))
	})

	t.Run("range", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPStatusCodes("200-299"))
	})

	t.Run("descending range", func(t *testing.T) {
		t.Parallel()
		assertValidationErrorContains(t, validateHTTPStatusCodes("299-200"), "invalid status code")
	})

	t.Run("not numeric", func(t *testing.T) {
		t.Parallel()
		assertValidationErrorContains(t, validateHTTPStatusCodes("ok"), "invalid status code")
	})
}

// TestValidateMaxAttempts verifies global max-attempts validation.
func TestValidateMaxAttempts(t *testing.T) {
	t.Parallel()

	t.Run("endless", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateMaxAttempts(-1))
	})

	t.Run("positive", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateMaxAttempts(1))
	})

	t.Run("zero", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateMaxAttempts(0), "max-attempts must be -1 or positive")
	})

	t.Run("below endless", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateMaxAttempts(-2), "max-attempts must be -1 or positive")
	})
}

// TestValidateOptionalMaxAttempts verifies per-target max-attempts validation.
func TestValidateOptionalMaxAttempts(t *testing.T) {
	t.Parallel()

	t.Run("inherits global", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateOptionalMaxAttempts(0))
	})

	t.Run("endless", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateOptionalMaxAttempts(-1))
	})

	t.Run("positive", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateOptionalMaxAttempts(3))
	})

	t.Run("below endless", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateOptionalMaxAttempts(-2), "max-attempts must be -1 or positive")
	})
}

// TestValidatePositiveDuration verifies positive duration validation.
func TestValidatePositiveDuration(t *testing.T) {
	t.Parallel()

	validateTimeout := validateTimeoutDuration()

	t.Run("positive", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTimeout(time.Nanosecond))
	})

	t.Run("zero", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateTimeout(0), "timeout must be positive")
	})

	t.Run("negative", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateTimeout(-time.Second), "timeout must be positive")
	})
}

// TestValidateNonNegativeDuration verifies non-negative duration validation.
func TestValidateNonNegativeDuration(t *testing.T) {
	t.Parallel()

	validateInterval := validateNonNegativeDuration("interval")

	t.Run("zero", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateInterval(0))
	})

	t.Run("positive", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateInterval(time.Second))
	})

	t.Run("negative", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateInterval(-time.Second), "interval must be non-negative")
	})
}

// assertNoValidationError verifies a validator accepted the value.
func assertNoValidationError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}

// assertExactValidationError verifies a validator returned the exact expected error.
func assertExactValidationError(t *testing.T, err error, want string) {
	t.Helper()
	require.Error(t, err)
	assert.EqualError(t, err, want)
}

// assertValidationErrorContains verifies a validator returned an error containing the expected text.
func assertValidationErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	require.Error(t, err)
	assert.Contains(t, err.Error(), want)
}
