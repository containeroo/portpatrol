package cli

import (
	"testing"
	"time"

	"github.com/containeroo/never/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// TestValidateHTTPAddress verifies HTTP address validation accepts supported inputs.
func TestValidateHTTPAddress(t *testing.T) {
	t.Parallel()

	t.Run("http URL", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPAddress("http://example.com"))
	})

	t.Run("https URL", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPAddress("https://example.com/ready"))
	})

	t.Run("resolver reference", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPAddress("env:TARGET_URL"))
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateHTTPAddress(""), "invalid HTTP URL")
	})

	t.Run("missing host", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateHTTPAddress("http://"), "invalid HTTP URL")
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateHTTPAddress("ftp://example.com"), `unsupported scheme: "ftp"`)
	})
}

// TestValidateICMPAddress verifies ICMP address validation accepts supported inputs.
func TestValidateICMPAddress(t *testing.T) {
	t.Parallel()

	for _, address := range []string{testutils.LocalhostIPv4, "2001:db8::1", "example.com", "localhost", "env:TARGET_HOST"} {
		address := address
		t.Run(address, func(t *testing.T) {
			t.Parallel()
			assertNoValidationError(t, validateICMPAddress(address))
		})
	}

	t.Run("scheme", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateICMPAddress("icmp://example.com"), "ICMP check cannot have a scheme")
	})

	t.Run("path", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateICMPAddress("example.com/ready"), "ICMP address must be a hostname or IP without path or port")
	})

	t.Run("port", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateICMPAddress("example.com:80"), "ICMP address must be a hostname or IP without path or port")
	})

	t.Run("invalid hostname", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateICMPAddress("exa_mple.com"), `invalid hostname: "exa_mple.com"`)
	})
}

// TestValidateTCPAddress verifies TCP address validation accepts supported inputs.
func TestValidateTCPAddress(t *testing.T) {
	t.Parallel()

	for _, address := range []string{testutils.LocalhostAddr("80"), "example.com:443", "[2001:db8::1]:443", "env:TARGET_ADDRESS"} {
		address := address
		t.Run(address, func(t *testing.T) {
			t.Parallel()
			assertNoValidationError(t, validateTCPAddress(address))
		})
	}

	t.Run("missing port", func(t *testing.T) {
		t.Parallel()
		assertValidationErrorContains(t, validateTCPAddress("example.com"), "TCP address must be host:port")
	})

	t.Run("scheme", func(t *testing.T) {
		t.Parallel()
		assertValidationErrorContains(t, validateTCPAddress("tcp://example.com:80"), "TCP address must be host:port")
	})
}

// TestIsResolvableValue verifies supported resolver prefixes are recognized.
func TestIsResolvableValue(t *testing.T) {
	t.Parallel()

	for _, value := range []string{
		"env:TARGET",
		"file:/config/app.txt//Target",
		"json:/config/app.json//target",
		"yaml:/config/app.yaml//target",
		"ini:/config/app.ini//Target.Address",
	} {
		assert.True(t, isResolvableValue(value), value)
	}

	assert.False(t, isResolvableValue("http://example.com"))
}

// TestIsHostnameLike verifies hostname validation edge cases.
func TestIsHostnameLike(t *testing.T) {
	t.Parallel()

	for _, hostname := range []string{"example.com", "sub.domain.local", "localhost", "a-b.c"} {
		assert.True(t, isHostnameLike(hostname), hostname)
	}

	for _, hostname := range []string{"", "-bad.example", "bad-.example", "bad..example", "example.com/", "example.com:80", "exa_mple.com"} {
		assert.False(t, isHostnameLike(hostname), hostname)
	}
}

// TestIsAlphaNum verifies ASCII hostname characters.
func TestIsAlphaNum(t *testing.T) {
	t.Parallel()

	for _, ch := range []byte{'a', 'A', '0', '9'} {
		assert.True(t, isAlphaNum(ch), string(ch))
	}
	for _, ch := range []byte{'-', '_', '!', '$', '%', '*', '/', ':', '@', '[', '\\', ']', '^'} {
		assert.False(t, isAlphaNum(ch), string(ch))
	}
}
