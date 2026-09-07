package cli

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveHTTPHeaderValues(t *testing.T) {
	t.Setenv("NEVER_TEST_HEADER", "secret")

	headers := http.Header{
		"Authorization": {"env:NEVER_TEST_HEADER"},
		"X-Test":        {"plain", "env:NEVER_TEST_HEADER"},
	}

	err := resolveHTTPHeaderValues(headers)

	require.NoError(t, err)
	assert.Equal(t, "secret", headers.Get("Authorization"))
	assert.Equal(t, []string{"plain", "secret"}, headers.Values("X-Test"))
}
