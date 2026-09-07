package cli

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/containeroo/never/internal/checker"

	"github.com/containeroo/httputils"
	"github.com/containeroo/resolver"
	"github.com/containeroo/tinyflags"
)

const httpUserAgentPrefix = "never/"

// parseHTTPConfig converts HTTP flag values into the typed checker config.
// Raw header and status-code syntax is parsed here and never leaves the CLI package.
func parseHTTPConfig(
	group *tinyflags.DynamicGroup,
	id string,
	version string,
	showPath bool,
) (checker.HTTPConfig, error) {
	allowDuplicateHeaders := tinyflags.GetOrDefaultDynamic[bool](group, id, "allow-duplicate-headers")
	headers, err := parseHTTPHeaders(
		tinyflags.GetOrDefaultDynamic[[]string](group, id, "header"),
		allowDuplicateHeaders,
	)
	if err != nil {
		return checker.HTTPConfig{}, fmt.Errorf("invalid HTTP header: %w", err)
	}

	expectedStatusCodes, err := parseHTTPStatusCodes(
		tinyflags.GetOrDefaultDynamic[[]string](group, id, "expected-status-codes"),
	)
	if err != nil {
		return checker.HTTPConfig{}, fmt.Errorf("invalid expected status codes: %w", err)
	}

	return checker.HTTPConfig{
		Method:              tinyflags.GetOrDefaultDynamic[string](group, id, "method"),
		Headers:             headers,
		ExpectedStatusCodes: expectedStatusCodes,
		FollowRedirects:     tinyflags.GetOrDefaultDynamic[bool](group, id, "follow-redirects"),
		MaxRedirects:        tinyflags.GetOrDefaultDynamic[int](group, id, "max-redirects"),
		SkipTLSVerify:       tinyflags.GetOrDefaultDynamic[bool](group, id, "skip-tls-verify"),
		Timeout:             tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout"),
		UserAgent:           httpUserAgentPrefix + version,
		ShowPath:            showPath,
	}, nil
}

// parseHTTPStatusCodes parses HTTP status-code flag values once at the CLI boundary.
func parseHTTPStatusCodes(values []string) ([]int, error) {
	if len(values) == 0 {
		return nil, nil
	}

	return httputils.ParseStatusCodes(strings.Join(values, ","))
}

// parseHTTPHeaders parses and resolves HTTP header flag values once at the CLI boundary.
func parseHTTPHeaders(values []string, allowDuplicates bool) (http.Header, error) {
	headers := make(http.Header)

	for _, raw := range values {
		key, value, ok := strings.Cut(raw, "=")
		key = http.CanonicalHeaderKey(strings.TrimSpace(key))
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid header format: %q", raw)
		}

		resolved, err := resolver.ResolveVariable(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve variable in header: %w", err)
		}

		if _, exists := headers[key]; exists && !allowDuplicates {
			return nil, fmt.Errorf("duplicate header: %q", raw)
		}

		if allowDuplicates {
			headers.Add(key, resolved)
			continue
		}
		headers.Set(key, resolved)
	}

	return headers, nil
}
