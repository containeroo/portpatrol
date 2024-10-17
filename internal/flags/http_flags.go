package flags

import (
	"fmt"
	"strconv"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
	"github.com/containeroo/portpatrol/pkg/httputils"
)

const (
	paramHTTPMethod                  string = "method"
	paramHTTPHeaders                 string = "headers"
	paramHTTPAllowDuplicateHeaders   string = "allow-duplicate-headers"
	paramHTTPExpectedStatusCodes     string = "expected-status-codes"
	paramHTTPSkipTLSVerify           string = "skip-tls-verify"
	paramHTTPTimeout                 string = "timeout"
	defaultHTTPAllowDuplicateHeaders bool   = false
	defaultHTTPSkipTLSVerify         bool   = false
)

// parseHTTPCheckerOptions parses HTTP checker specific options from parameters.
func parseHTTPCheckerOptions(params map[string]string) ([]checker.Option, error) {
	var opts []checker.Option

	// Track unrecognized parameters
	unrecognizedParams := make(map[string]struct{})
	for key := range params {
		unrecognizedParams[key] = struct{}{}
	}

	// HTTP Method
	if method, ok := params[paramHTTPMethod]; ok && method != "" {
		opts = append(opts, checker.WithHTTPMethod(method))
		delete(unrecognizedParams, paramHTTPMethod)
	}

	// Allow Duplicate Headers
	allowDupHeaders := defaultHTTPAllowDuplicateHeaders
	if allowDupHeadersStr, ok := params[paramHTTPAllowDuplicateHeaders]; ok && allowDupHeadersStr != "" {
		var err error
		allowDupHeaders, err = strconv.ParseBool(allowDupHeadersStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramHTTPAllowDuplicateHeaders, err)
		}
		delete(unrecognizedParams, paramHTTPAllowDuplicateHeaders)
	}

	// Headers
	if headersStr, ok := params[paramHTTPHeaders]; ok && headersStr != "" {
		headers, err := httputils.ParseHeaders(headersStr, allowDupHeaders)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramHTTPHeaders, err)
		}
		opts = append(opts, checker.WithHTTPHeaders(headers))
		delete(unrecognizedParams, paramHTTPHeaders)
	}

	// Expected Status Codes
	if codesStr, ok := params[paramHTTPExpectedStatusCodes]; ok && codesStr != "" {
		codes, err := httputils.ParseStatusCodes(codesStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramHTTPExpectedStatusCodes, err)
		}
		opts = append(opts, checker.WithExpectedStatusCodes(codes))
		delete(unrecognizedParams, paramHTTPExpectedStatusCodes)
	}

	// Skip TLS Verify
	if skipStr, ok := params[paramHTTPSkipTLSVerify]; ok && skipStr != "" {
		skip, err := strconv.ParseBool(skipStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", paramHTTPSkipTLSVerify, err)
		}
		opts = append(opts, checker.WithHTTPSkipTLSVerify(skip))
		delete(unrecognizedParams, paramHTTPSkipTLSVerify)
	}

	// Timeout
	if timeoutStr, ok := params[paramHTTPTimeout]; ok && timeoutStr != "" {
		t, err := time.ParseDuration(timeoutStr)
		if err != nil || t <= 0 {
			return nil, fmt.Errorf("invalid %q: %w", paramHTTPTimeout, err)
		}
		opts = append(opts, checker.WithHTTPTimeout(t))
		delete(unrecognizedParams, paramHTTPTimeout)
	}

	// Check for unrecognized parameters
	if len(unrecognizedParams) > 0 {
		var unknownKeys []string
		for key := range unrecognizedParams {
			unknownKeys = append(unknownKeys, key)
		}
		return nil, fmt.Errorf("unrecognized parameters for HTTP checker: %v", unknownKeys)
	}

	return opts, nil
}
