package cli

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/containeroo/never/internal/checker"

	"github.com/containeroo/httputils"
	"github.com/containeroo/resolver"
	"github.com/containeroo/tinyflags"
)

const httpUserAgentPrefix = "never/"

// registerHTTPFlags registers HTTP-related flags.
func registerHTTPFlags(tf *tinyflags.FlagSet) {
	httpGroup := tf.DynamicGroup("http").Title("HTTP")
	httpGroup.String("name", "", "Name of the HTTP checker. Defaults to <ID>.")
	tinyflags.DynamicEnum(
		httpGroup, "method", checker.DefaultHTTPConfig().Method, "HTTP method to use.",
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	).
		Placeholder("METHOD")
	httpGroup.String("address", "", "HTTP target URL").
		Validate(validateHTTPAddress).
		Required()
	tinyflags.DynamicEnum(
		httpGroup,
		"address-detail",
		checker.HTTPAddressOrigin,
		"HTTP address detail shown in logs (full may expose secrets).",
		checker.HTTPAddressOrigin,
		checker.HTTPAddressPath,
		checker.HTTPAddressQuery,
		checker.HTTPAddressFull,
	).
		Placeholder("DETAIL")
	httpGroup.Duration("interval", 0*time.Second, "Time between HTTP requests. Defaults to --default-interval when unset or 0.").
		Validate(validateNonNegativeDuration("interval")).
		Placeholder("DURATION")
	httpGroup.Int("max-attempts", 0, "Maximum attempts before giving up. Defaults to --max-attempts when unset or 0.").
		Validate(validateOptionalMaxAttempts).
		Placeholder("N")
	registerRetryFlags(httpGroup)
	// Headers are repeated flags; a newline delimiter preserves commas inside header values.
	httpGroup.StringSlice("header", []string{}, "HTTP headers to send").
		Delimiter("\n").
		Placeholder("KEY=VALUE")
	httpGroup.Bool("allow-duplicate-headers", defaultHTTPAllowDuplicateHeaders, "Allow duplicate HTTP headers")
	tinyflags.DynamicSlice(
		httpGroup,
		"expected-status-codes",
		[]int{http.StatusOK},
		"Expected HTTP status codes. Comma-separated list of status codes, ranges possible (eg \"200-299\", \"300,301\")",
		httputils.ParseStatusCodes,
		strconv.Itoa,
	).
		Placeholder("CODES...")
	httpGroup.Bool("follow-redirects", defaultHTTPFollowRedirects, "Follow HTTP redirects").Strict()
	httpGroup.Int("max-redirects", defaultHTTPMaxRedirects, "Maximum number of redirects to follow. Set to 0 to disable redirects.").
		Validate(validateNonNegativeInt("max-redirects")).
		Placeholder("N")

	httpGroup.Bool("skip-tls-verify", defaultHTTPSkipTLSVerify, "Skip TLS verification")
	httpGroup.Duration("timeout", checker.DefaultHTTPConfig().Timeout, "Request timeout").
		Validate(validateTimeoutDuration()).
		Placeholder("DURATION")
}

// resolveHTTPHeaderValues resolves configured variables in parsed HTTP header values.
func resolveHTTPHeaderValues(headers http.Header) error {
	for _, values := range headers {
		for i, value := range values {
			resolved, err := resolver.ResolveVariable(value)
			if err != nil {
				return fmt.Errorf("failed to resolve variable in header: %w", err)
			}
			values[i] = resolved
		}
	}

	return nil
}
