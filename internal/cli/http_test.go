package cli

import (
	"net/http"
	"testing"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	httpExampleURL     = "http://example.com"
	httpWebAddressFlag = "--http.web.address=" + httpExampleURL
)

// TestParseFlagsHTTPMethod verifies HTTP method parsing uses enum validation.
func TestParseFlagsHTTPMethod(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.method=POST",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		cfg := requireHTTPConfig(t, parsedFlags.Targets[0])
		assert.Equal(t, http.MethodPost, cfg.Method)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.method=INVALID",
		}, "1.0.0")
		assertInvalidFlagValueError(t, err, "--http.web.method", "INVALID", http.MethodGet, http.MethodPost)
	})
}

// TestParseFlagsHTTPTarget verifies HTTP flags are converted into typed target config.
func TestParseFlagsHTTPTarget(t *testing.T) {
	t.Parallel()

	args := []string{
		"--default-interval=5s",
		"--http.web.name=Web",
		httpWebAddressFlag,
		"--http.web.method=POST",
		"--http.web.header=Authorization=Bearer token",
		"--http.web.expected-status-codes=200,204",
		"--http.web.follow-redirects=false",
		"--http.web.max-redirects=3",
		"--http.web.backoff=exponential",
		"--http.web.max-interval=30s",
		"--http.web.max-attempts=3",
	}

	parsedFlags, err := ParseFlags(args, "1.0.0")
	require.NoError(t, err)
	require.Len(t, parsedFlags.Targets, 1)

	target := parsedFlags.Targets[0]
	assert.Equal(t, "web", target.ID)
	assert.Equal(t, "Web", target.Name)
	assert.Equal(t, httpExampleURL, target.Address)
	cfg := requireHTTPConfig(t, target)
	assert.Equal(t, http.MethodPost, cfg.Method)
	assert.Equal(t, http.Header{"Authorization": {"Bearer token"}}, cfg.Headers)
	assert.Equal(t, []int{http.StatusOK, http.StatusNoContent}, cfg.ExpectedStatusCodes)
	assert.False(t, cfg.FollowRedirects)
	assert.Equal(t, 3, cfg.MaxRedirects)
	assert.Equal(t, "never/1.0.0", cfg.UserAgent)
	assert.Equal(t, 3, target.MaxAttempts)
	assert.Equal(t, backoff.ModeExponential, target.Backoff)
	assert.Equal(t, 30*time.Second, target.MaxInterval)
}

// TestParseFlagsHTTPMaxRedirects verifies redirect limit parsing and validation.
func TestParseFlagsHTTPMaxRedirects(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{httpWebAddressFlag}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		cfg := requireHTTPConfig(t, parsedFlags.Targets[0])
		assert.Equal(t, defaultHTTPMaxRedirects, cfg.MaxRedirects)
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.max-redirects=0",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		cfg := requireHTTPConfig(t, parsedFlags.Targets[0])
		assert.Zero(t, cfg.MaxRedirects)
	})

	t.Run("negative", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.max-redirects=-1",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "max-redirects must be non-negative")
	})
}

// TestParseFlagsHTTPFollowRedirects verifies redirect following defaults to enabled and can be disabled.
func TestParseFlagsHTTPFollowRedirects(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{httpWebAddressFlag}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		cfg := requireHTTPConfig(t, parsedFlags.Targets[0])
		assert.True(t, cfg.FollowRedirects)
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.follow-redirects=false",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		cfg := requireHTTPConfig(t, parsedFlags.Targets[0])
		assert.False(t, cfg.FollowRedirects)
	})
}

// TestParseFlagsHTTPBackoff verifies HTTP backoff values use enum validation.
func TestParseFlagsHTTPBackoff(t *testing.T) {
	t.Parallel()

	t.Run("exponential", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.backoff=exponential",
			"--http.web.max-interval=30s",
		}, "1.0.0")
		require.NoError(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.backoff=invalid",
		}, "1.0.0")
		assertInvalidFlagValueError(t, err, "--http.web.backoff", "invalid", backoff.ModeLinear.String(), backoff.ModeExponential.String())
	})
}

// TestParseFlagsHTTPPerTargetMaxAttempts verifies HTTP target max-attempts override parsing.
func TestParseFlagsHTTPPerTargetMaxAttempts(t *testing.T) {
	t.Parallel()

	parsedFlags, err := ParseFlags([]string{
		"--max-attempts=5",
		httpWebAddressFlag,
		"--http.web.max-attempts=2",
	}, "1.0.0")
	require.NoError(t, err)
	require.Len(t, parsedFlags.Targets, 1)
	assert.Equal(t, 5, parsedFlags.MaxAttempts)
	assert.Equal(t, 2, parsedFlags.Targets[0].MaxAttempts)
}

// TestParseFlagsHTTPInputParsing verifies HTTP-specific raw values are parsed at the CLI boundary.
func TestParseFlagsHTTPInputParsing(t *testing.T) {
	t.Parallel()

	t.Run("invalid header", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=InvalidHeader",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, `invalid HTTP header: invalid header format: "InvalidHeader"`)
	})

	t.Run("duplicate header", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=X-Test=one",
			"--http.web.header=X-Test=two",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, `duplicate header: "X-Test=two"`)
	})

	t.Run("duplicate header allowed", func(t *testing.T) {
		t.Parallel()

		cfg, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=X-Test=one",
			"--http.web.header=X-Test=two",
			"--http.web.allow-duplicate-headers=true",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, cfg.Targets, 1)
		httpConfig := requireHTTPConfig(t, cfg.Targets[0])
		assert.Equal(t, []string{"one", "two"}, httpConfig.Headers.Values("X-Test"))
	})

	t.Run("invalid expected status codes", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.expected-status-codes=299-200",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid expected status codes")
	})
}

// TestParseFlagsHTTPRetryValidation verifies retry input is rejected before reaching the factory.
func TestParseFlagsHTTPRetryValidation(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		flag string
		want string
	}{
		{name: "negative interval", flag: "--http.web.interval=-1s", want: "interval must be non-negative"},
		{name: "negative max interval", flag: "--http.web.max-interval=-1s", want: "max-interval must be non-negative"},
		{name: "invalid max attempts", flag: "--http.web.max-attempts=-2", want: "max-attempts must be -1 or positive"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseFlags([]string{httpWebAddressFlag, tt.flag}, "1.0.0")
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.want)
		})
	}
}
