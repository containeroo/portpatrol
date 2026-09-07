package cli

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHTTPStatusCodes(t *testing.T) {
	t.Parallel()

	codes, err := parseHTTPStatusCodes([]string{"200", "204", "300-301"})
	require.NoError(t, err)
	assert.Equal(t, []int{200, 204, 300, 301}, codes)
}

func TestParseHTTPHeaders(t *testing.T) {
	t.Run("nil headers", func(t *testing.T) {
		t.Parallel()

		headers, err := parseHTTPHeaders(nil, false)
		require.NoError(t, err)
		assert.Equal(t, http.Header{}, headers)
	})

	t.Run("duplicate headers allowed", func(t *testing.T) {
		t.Parallel()

		headers, err := parseHTTPHeaders([]string{
			"X-Test=one",
			"X-Test=two",
		}, true)
		require.NoError(t, err)
		assert.Equal(t, []string{"one", "two"}, headers.Values("X-Test"))
	})

	t.Run("header names are case insensitive", func(t *testing.T) {
		t.Parallel()

		_, err := parseHTTPHeaders([]string{"X-Test=one", "x-test=two"}, false)
		require.Error(t, err)

		headers, err := parseHTTPHeaders([]string{"X-Test=one", "x-test=two"}, true)
		require.NoError(t, err)
		assert.Equal(t, []string{"one", "two"}, headers.Values("X-Test"))
	})

	t.Run("resolves values", func(t *testing.T) {
		t.Setenv("NEVER_TEST_HEADER", "secret")

		headers, err := parseHTTPHeaders([]string{"Authorization=env:NEVER_TEST_HEADER"}, false)
		require.NoError(t, err)
		assert.Equal(t, http.Header{"Authorization": {"secret"}}, headers)
	})

	t.Run("whitespace-only key", func(t *testing.T) {
		t.Parallel()

		_, err := parseHTTPHeaders([]string{"   =value"}, false)
		require.Error(t, err)
		assert.EqualError(t, err, `invalid header format: "   =value"`)
	})
}
