package cli

import (
	"net/http"
	"time"

	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/tinyflags"
)

// registerHTTPFlags registers HTTP-related flags and binds them to cfg.
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
		Required()
	httpGroup.Duration("interval", 0*time.Second, "Time between HTTP requests. Defaults to --default-interval when unset or 0.").
		Validate(validateNonNegativeDuration("interval")).
		Placeholder("DURATION")
	httpGroup.Int("max-attempts", 0, "Maximum attempts before giving up. Defaults to --max-attempts when unset or 0.").
		Validate(validateOptionalMaxAttempts).
		Placeholder("N")
	registerRetryFlags(httpGroup)
	httpGroup.StringSlice("header", []string{}, "HTTP headers to send").
		Placeholder("KEY=VALUE)")
	httpGroup.Bool("allow-duplicate-headers", defaultHTTPAllowDuplicateHeaders, "Allow duplicate HTTP headers")
	httpGroup.StringSlice(
		"expected-status-codes",
		[]string{"200"},
		"Expected HTTP status codes. Comma-separated list of status codes, ranges possible (eg \"200-299\", \"300,301\")",
	).
		Validate(validateHTTPStatusCodes).
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
