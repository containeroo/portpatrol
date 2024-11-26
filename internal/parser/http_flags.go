package parser

import (
	"fmt"
	"strconv"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
	"github.com/containeroo/portpatrol/internal/flags"
	"github.com/containeroo/portpatrol/pkg/httputils"
)

const (
	ParamHTTPMethod                  string = "method"
	ParamHTTPHeaders                 string = "headers"
	ParamHTTPAllowDuplicateHeaders   string = "allow-duplicate-headers"
	ParamHTTPExpectedStatusCodes     string = "expected-status-codes"
	ParamHTTPSkipTLSVerify           string = "skip-tls-verify"
	ParamHTTPTimeout                 string = "timeout"
	defaultHTTPAllowDuplicateHeaders bool   = false
	defaultHTTPSkipTLSVerify         bool   = false
)

var httpFlagDocs = []flags.FlagDoc{
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamAddress),
		Description: "The IP address or hostname of the target in the following format: scheme://hostname[:port]",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamName),
		Description: "The name of the target. If not specified, it's derived from the target address.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamType),
		Description: "The type of check to perform. If a scheme (e.g. http://) is specified in --%s.<identifier>.address, this flag can be omitted as the type will be inferred.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamHTTPMethod),
		Description: "The HTTP method to use (e.g., GET, POST). Defaults to \"GET\".",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamHTTPHeaders),
		Description: "A comma-separated list of HTTP headers to include in the request in \"key=value\" format.\n\tExample: Authorization=Bearer token,Content-Type=application/json",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>%s=string", ParamPrefix, ParamHTTPExpectedStatusCodes),
		Description: "A comma-separated list of expected HTTP status codes or ranges. Defaults to 200.\n\tExample: \"200,301,404\" or \"200,300-302\" or \"200,301-302,404,500-502\"",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=bool", ParamPrefix, ParamHTTPSkipTLSVerify),
		Description: "Whether to skip TLS verification. Defaults to false.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamHTTPTimeout),
		Description: "The timeout for the HTTP request (e.g., 5s).",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamInterval),
		Description: "Override the default interval for this target (e.g., 10s).",
	},
}

// parseHTTPCheckerOptions parses HTTP checker-specific options from parameters.
func parseHTTPCheckerOptions(params map[string]string) ([]checker.Option, error) {
	var opts []checker.Option
	unrecognizedParams := trackUnusedParams(params)

	// HTTP Method
	if method, ok := params[ParamHTTPMethod]; ok && method != "" {
		opts = append(opts, checker.WithHTTPMethod(method))
		delete(unrecognizedParams, ParamHTTPMethod)
	}

	// Allow Duplicate Headers
	allowDupHeaders := defaultHTTPAllowDuplicateHeaders
	if allowDupHeadersStr, ok := params[ParamHTTPAllowDuplicateHeaders]; ok && allowDupHeadersStr != "" {
		var err error
		allowDupHeaders, err = strconv.ParseBool(allowDupHeadersStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamHTTPAllowDuplicateHeaders, err)
		}
		delete(unrecognizedParams, ParamHTTPAllowDuplicateHeaders)
	}

	// Headers
	if headersStr, ok := params[ParamHTTPHeaders]; ok && headersStr != "" {
		headers, err := httputils.ParseHeaders(headersStr, allowDupHeaders)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamHTTPHeaders, err)
		}
		opts = append(opts, checker.WithHTTPHeaders(headers))
		delete(unrecognizedParams, ParamHTTPHeaders)
	}

	// Expected Status Codes
	if codesStr, ok := params[ParamHTTPExpectedStatusCodes]; ok && codesStr != "" {
		codes, err := httputils.ParseStatusCodes(codesStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamHTTPExpectedStatusCodes, err)
		}
		opts = append(opts, checker.WithExpectedStatusCodes(codes))
		delete(unrecognizedParams, ParamHTTPExpectedStatusCodes)
	}

	// Skip TLS Verify
	if skipStr, ok := params[ParamHTTPSkipTLSVerify]; ok && skipStr != "" {
		skip, err := strconv.ParseBool(skipStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamHTTPSkipTLSVerify, err)
		}
		opts = append(opts, checker.WithHTTPSkipTLSVerify(skip))
		delete(unrecognizedParams, ParamHTTPSkipTLSVerify)
	}

	// Timeout
	if timeoutStr, ok := params[ParamHTTPTimeout]; ok && timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil || timeout <= 0 {
			return nil, fmt.Errorf("invalid %q: %w", ParamHTTPTimeout, err)
		}
		opts = append(opts, checker.WithHTTPTimeout(timeout))
		delete(unrecognizedParams, ParamHTTPTimeout)
	}

	// Check for unrecognized parameters
	if len(unrecognizedParams) > 0 {
		return nil, fmt.Errorf("unrecognized parameters for HTTP checker: %v", extractMapKeys(unrecognizedParams))
	}

	return opts, nil
}
