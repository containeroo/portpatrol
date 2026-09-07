package factory_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	targetID        = "mygroup"
	testHTTPAddress = "http://example.com"
	testVersion     = "0.0.0"
)

// TestBuildCheckers verifies the expected behavior.
func TestBuildCheckers(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP Checker", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:          targetID,
				Name:        targetID,
				Address:     testHTTPAddress,
				Interval:    5 * time.Second,
				MaxAttempts: 3,
				HTTP: &factory.HTTPConfig{
					Method:                http.MethodGet,
					Headers:               []string{"Content-Type=application/json"},
					AllowDuplicateHeaders: true,
					ExpectedStatusCodes:   []string{"200"},
					FollowRedirects:       true,
					MaxRedirects:          4,
					SkipTLSVerify:         true,
					Timeout:               33 * time.Second,
				},
			},
		}, 9*time.Second, -1, testVersion, false)

		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, testHTTPAddress, checkers[0].Checker.Address())
		assert.Equal(t, 5*time.Second, checkers[0].Interval)
		assert.Equal(t, 3, checkers[0].MaxAttempts)
	})

	t.Run("Rejects missing checker config", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{ID: targetID, Address: testHTTPAddress},
		}, 2*time.Second, -1, testVersion, false)

		assert.Nil(t, checkers)
		assert.EqualError(t, err, `target "mygroup" must configure exactly one checker`)
	})

	t.Run("Rejects multiple checker configs", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      targetID,
				Address: testHTTPAddress,
				HTTP:    &factory.HTTPConfig{},
				TCP:     &factory.TCPConfig{},
			},
		}, 2*time.Second, -1, testVersion, false)

		assert.Nil(t, checkers)
		assert.EqualError(t, err, `target "mygroup" must configure exactly one checker`)
	})

	t.Run("Invalid Header Parsing", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      targetID,
				Address: testHTTPAddress,
				HTTP: &factory.HTTPConfig{
					Method:  http.MethodGet,
					Headers: []string{"InvalidHeaderFormat"},
				},
			},
		}, 2*time.Second, -1, testVersion, false)

		require.Error(t, err)
		assert.Nil(t, checkers)
		assert.EqualError(t, err, `target "mygroup": failed to create HTTP checker: invalid HTTP header: invalid header format: "InvalidHeaderFormat"`)
	})

	t.Run("Invalid HTTP Status codes", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      "myid",
				Address: testHTTPAddress,
				HTTP: &factory.HTTPConfig{
					Method:              http.MethodGet,
					ExpectedStatusCodes: []string{"201-200"},
				},
			},
		}, 2*time.Second, -1, testVersion, false)

		require.Error(t, err)
		assert.Empty(t, checkers)
	})

	t.Run("Valid HTTP Status codes", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      targetID,
				Address: testHTTPAddress,
				HTTP: &factory.HTTPConfig{
					Method:              http.MethodGet,
					ExpectedStatusCodes: []string{"200,201"},
				},
			},
		}, 2*time.Second, -1, testVersion, false)

		require.NoError(t, err)
		assert.Len(t, checkers, 1)
	})

	t.Run("HTTP Backoff", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:          targetID,
				Address:     testHTTPAddress,
				Backoff:     backoff.ModeExponential,
				MaxInterval: 30 * time.Second,
				HTTP:        &factory.HTTPConfig{Method: http.MethodGet},
			},
		}, 2*time.Second, -1, testVersion, false)

		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, backoff.ModeExponential, checkers[0].Backoff)
		assert.Equal(t, 30*time.Second, checkers[0].MaxInterval)
	})

	t.Run("Valid TCP Checker", func(t *testing.T) {
		t.Parallel()

		address := testutils.LocalhostAddr("8080")
		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      targetID,
				Address: address,
				TCP:     &factory.TCPConfig{Timeout: 3 * time.Second},
			},
		}, 2*time.Second, -1, testVersion, false)

		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, address, checkers[0].Checker.Address())
	})

	t.Run("Valid ICMP Checker", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      targetID,
				Address: testutils.LocalhostIPv4,
				ICMP: &factory.ICMPConfig{
					Timeout:      2 * time.Second,
					ReadTimeout:  2 * time.Second,
					WriteTimeout: 2 * time.Second,
				},
			},
		}, 2*time.Second, -1, testVersion, false)

		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, testutils.LocalhostIPv4, checkers[0].Checker.Address())
	})

	t.Run("Invalid ICMP Checker", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      targetID,
				Address: "://invalid-url",
				ICMP:    &factory.ICMPConfig{},
			},
		}, 2*time.Second, -1, testVersion, false)

		assert.Nil(t, checkers)
		require.Error(t, err)
	})
}

func TestBuildCheckersHTTPLogAddress(t *testing.T) {
	t.Parallel()

	const address = "https://user:password@example.com/private?q=secret#fragment"

	for _, tt := range []struct {
		name     string
		showPath bool
		want     string
	}{
		{name: "hidden", want: "https://example.com"},
		{name: "visible", showPath: true, want: "https://example.com/private?q=secret#fragment"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checkers, err := factory.BuildCheckers([]factory.TargetConfig{
				{
					ID:      targetID,
					Address: address,
					HTTP: &factory.HTTPConfig{
						FollowRedirects: true,
						MaxRedirects:    10,
					},
				},
			}, 2*time.Second, -1, testVersion, tt.showPath)

			require.NoError(t, err)
			require.Len(t, checkers, 1)
			assert.Equal(t, tt.want, checkers[0].Checker.Address())
		})
	}
}
