package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
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
