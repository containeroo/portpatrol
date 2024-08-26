package checker

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	envHTTPMethod              = "HTTP_METHOD"
	envHTTPHeaders             = "HTTP_HEADERS"
	envHTTPExpectedStatusCodes = "HTTP_EXPECTED_STATUS_CODES"

	defaultHTTPExpectedStatus = "200"
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
	// Parse method
	method := getEnv(envHTTPMethod)
	if method == "" {
		method = http.MethodGet
	}

	// Parse headers
	headers, err := parseHeaders(getEnv(envHTTPHeaders))
	if err != nil {
		return nil, fmt.Errorf("invalid %s value: %w", envHTTPHeaders, err)
	}

	// Parse expected status codes
	statusCodes := getEnv(envHTTPExpectedStatusCodes)
	if statusCodes == "" {
		statusCodes = defaultHTTPExpectedStatus
	}
	expectedStatusCodes, err := parseExpectedStatuses(statusCodes)
	if err != nil {
		return nil, fmt.Errorf("invalid %s value: %w", envHTTPExpectedStatusCodes, err)
	}

	// Create the HTTP client
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	return &HTTPChecker{
		Name:                name,
		Address:             address,
		ExpectedStatusCodes: expectedStatusCodes,
		Method:              method,
		Headers:             headers,
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

// parseExpectedStatuses parses a string of expected statuses into a slice of acceptable status codes.
func parseExpectedStatuses(statuses string) ([]int, error) {
	var statusCodes []int

	ranges := strings.Split(statuses, ",")
	for _, r := range ranges {
		// Check if the range is a single status code
		if !strings.Contains(r, "-") {
			code, err := strconv.Atoi(r) // Parse the status code
			if err != nil {
				return nil, fmt.Errorf("invalid status code: %s", r)
			}
			statusCodes = append(statusCodes, code)
			continue
		}

		// Split the range into start and end
		parts := strings.Split(r, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid status range: %s", r)
		}

		// Parse the start and end status codes
		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])
		// Check if parsing failed or if the start is greater than the end
		if err1 != nil || err2 != nil || start > end {
			return nil, fmt.Errorf("invalid status range: %s", r)
		}

		// Generate a slice of status codes in the range
		for i := start; i <= end; i++ {
			statusCodes = append(statusCodes, i)
		}

	}
	return statusCodes, nil
}

// parseHeaders parses the HTTP headers from a comma-separated key=value list.
func parseHeaders(headers string) (map[string]string, error) {
	headerMap := make(map[string]string)
	if headers == "" {
		return headerMap, nil
	}

	// Split the headers into key=value pairs
	pairs := strings.Split(headers, ",")
	for _, pair := range pairs {
		if strings.TrimSpace(pair) == "" {
			continue // Skip any empty parts resulting from trailing commas
		}

		// Split the pair into key and value
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: %s", pair)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("header key cannot be empty: %s", pair)
		}
		headerMap[key] = value
	}

	return headerMap, nil
}
