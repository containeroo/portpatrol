package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	envHTTPMethod                = "HTTP_METHOD"
	envHTTPHeaders               = "HTTP_HEADERS"
	envHTTPAllowDuplicateHeaders = "HTTP_ALLOW_DUPLICATE_HEADERS"
	envHTTPExpectedStatusCodes   = "HTTP_EXPECTED_STATUS_CODES"
	envHTTPSkipTLSVerify         = "HTTP_SKIP_TLS_VERIFY"

	defaultHTTPExpectedStatus        = "200"
	defaultHTTPAllowDuplicateHeaders = "false"
	defaultHTTPSkipTLSVerify         = "false"
)

// HTTPChecker implements the Checker interface for HTTP checks.
type HTTPChecker struct {
	Name                string            // The name of the checker.
	Address             string            // The address of the target.
	ExpectedStatusCodes []int             // The expected status codes.
	Method              string            // The HTTP method to use.
	Headers             map[string]string // The HTTP headers to include in the request.
	client              *http.Client      // The HTTP client to use for the request.
	DialTimeout         time.Duration     // The timeout for dialing the target.
}

// String returns the name of the checker.
func (c *HTTPChecker) String() string {
	return c.Name
}

// NewHTTPChecker creates a new HTTPChecker.
func NewHTTPChecker(name, address string, timeout time.Duration, getEnv func(string) string) (Checker, error) {
	// Determine the HTTP method
	httpMethod := getEnv(envHTTPMethod)
	if httpMethod == "" {
		httpMethod = http.MethodGet
	}

	// Determine if duplicate headers are allowed
	allowDupHeaderStr := getEnv(envHTTPAllowDuplicateHeaders)
	if allowDupHeaderStr == "" {
		allowDupHeaderStr = defaultHTTPAllowDuplicateHeaders
	}

	// Parse the boolean value for allowDuplicates
	allowDupHeaders, err := strconv.ParseBool(allowDupHeaderStr)
	if err != nil {
		return nil, fmt.Errorf("invalid %s value: %w", envHTTPAllowDuplicateHeaders, err)
	}

	// Parse headers
	parsedHeaders, err := parseHTTPHeaders(getEnv(envHTTPHeaders), allowDupHeaders)
	if err != nil {
		return nil, fmt.Errorf("invalid %s value: %w", envHTTPHeaders, err)
	}

	// Determine the expected status codes
	expectedStatusStr := getEnv(envHTTPExpectedStatusCodes)
	if expectedStatusStr == "" {
		expectedStatusStr = defaultHTTPExpectedStatus
	}
	expectedStatusCodes, err := parseHTTPStatusCodes(expectedStatusStr)
	if err != nil {
		return nil, fmt.Errorf("invalid %s value: %w", envHTTPExpectedStatusCodes, err)
	}

	// Determine if TLS verification should be skipped
	skipTLSVerifyStr := getEnv(envHTTPSkipTLSVerify)
	if skipTLSVerifyStr == "" {
		skipTLSVerifyStr = defaultHTTPSkipTLSVerify
	}

	// Parse the boolean value for skipTLSVerify
	skipTLSVerify, err := strconv.ParseBool(skipTLSVerifyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid %s value: %w", envHTTPSkipTLSVerify, err)
	}

	// Create the HTTP client with the given timeout and TLS configuration
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipTLSVerify,
			},
		},
	}

	return &HTTPChecker{
		Name:                name,
		Address:             address,
		ExpectedStatusCodes: expectedStatusCodes,
		Method:              httpMethod,
		Headers:             parsedHeaders,
		client:              client,
	}, nil
}

// Check performs an HTTP request and checks the response.
func (c *HTTPChecker) Check(ctx context.Context) error {
	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, c.Method, c.Address, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to the request
	for key, value := range c.Headers {
		req.Header.Add(key, value)
	}

	// Perform the HTTP request
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	for _, code := range c.ExpectedStatusCodes {
		if resp.StatusCode == code {
			// Return nil if the status code matches
			return nil
		}
	}

	return fmt.Errorf("unexpected status code: got %d, expected one of %v", resp.StatusCode, c.ExpectedStatusCodes)
}

// parseHTTPStatusCodes parses a comma-separated string of HTTP status codes and ranges
// into a slice of individual status codes. It supports combinations of single codes
// (e.g., "200") and ranges (e.g., "200-204"), including mixed combinations like "200,300-301,404".
// Returns an error if any code or range is invalid.
func parseHTTPStatusCodes(statusRanges string) ([]int, error) {
	var statusCodes []int

	ranges := strings.Split(statusRanges, ",")
	for _, r := range ranges {
		trimmed := strings.TrimSpace(r)

		if !strings.Contains(trimmed, "-") {
			// Handle individual status codes like "200"
			code, err := strconv.Atoi(trimmed)
			if err != nil {
				return nil, fmt.Errorf("invalid status code: %s", trimmed)
			}
			statusCodes = append(statusCodes, code)
			continue
		}

		// Handle ranges like "200-204"
		parts := strings.Split(trimmed, "-") // Split the range into start and end
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid status range: %s", trimmed)
		}

		// Parse the start and end status codes
		start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

		// Check if parsing failed or if the start is greater than the end
		if err1 != nil || err2 != nil || start > end {
			return nil, fmt.Errorf("invalid status range: %s", trimmed)
		}

		// Generate a slice of status codes in the range
		for i := start; i <= end; i++ {
			statusCodes = append(statusCodes, i)
		}
	}

	return statusCodes, nil
}

// parseHTTPHeaders parses a comma-separated list of HTTP headers in key=value format
// and returns a map of the headers. It supports multiple headers, including combinations
// like "Content-Type=application/json, Authorization=Bearer token". The value can be
// empty (e.g., "X-Empty-Header="), but the key must not be empty.
// If allowDuplicates is false, the function will return an error if a duplicate header is encountered.
// If allowDuplicates is true, the function will override the previous value with the new one.
func parseHTTPHeaders(headers string, allowDuplicates bool) (map[string]string, error) {
	headerMap := make(map[string]string)
	if headers == "" {
		return headerMap, nil
	}

	// Split the headers into key=value pairs
	pairs := strings.Split(headers, ",")
	for _, pair := range pairs {
		trimmedPair := strings.TrimSpace(pair)
		if trimmedPair == "" {
			continue // Skip any empty parts resulting from trailing commas
		}

		// Split the pair into key and value
		parts := strings.SplitN(trimmedPair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: %s", pair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("header key cannot be empty: %s", pair)
		}

		if _, exists := headerMap[key]; exists && !allowDuplicates {
			return nil, fmt.Errorf("duplicate header key found: %s", key)
		}

		headerMap[key] = value
	}

	return headerMap, nil
}
