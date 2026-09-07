package checker

import (
	"github.com/containeroo/never/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

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

	t.Run("trims whitespace", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateHTTPAddress(" https://example.com "))
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

	t.Run("IPv4", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress(testutils.LocalhostIPv4))
	})

	t.Run("IPv6", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress("2001:db8::1"))
	})

	t.Run("hostname", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress("example.com"))
	})

	t.Run("localhost", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress("localhost"))
	})

	t.Run("trims whitespace", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateICMPAddress(" example.com "))
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		assertExactValidationError(t, validateICMPAddress(""), "ICMP address cannot be empty")
	})

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

	t.Run("IPv4 with port", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTCPAddress(testutils.LocalhostAddr("80")))
	})

	t.Run("hostname with port", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTCPAddress("example.com:443"))
	})

	t.Run("IPv6 with port", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTCPAddress("[2001:db8::1]:443"))
	})

	t.Run("trims whitespace", func(t *testing.T) {
		t.Parallel()
		assertNoValidationError(t, validateTCPAddress(" example.com:443 "))
	})

	t.Run("missing port", func(t *testing.T) {
		t.Parallel()
		assertValidationErrorContains(t, validateTCPAddress("example.com"), "TCP address must be host:port")
	})

	t.Run("scheme", func(t *testing.T) {
		t.Parallel()
		assertValidationErrorContains(t, validateTCPAddress("tcp://example.com:80"), "TCP address must be host:port")
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
