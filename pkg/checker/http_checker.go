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
	envMethod             = "METHOD"
	envHeaders            = "HEADERS"
	envExpectedStatuses   = "EXPECTED_STATUSES"
	defaultExpectedStatus = "200"
)

// HTTPChecker implements the Checker interface for HTTP checks.
type HTTPChecker struct {
	Name                string
	Address             string
	ExpectedStatusCodes []int
	Method              string
	Headers             map[string]string
	client              *http.Client
	DialTimeout         time.Duration
}

// String returns the name of the checker.
func (c *HTTPChecker) String() string {
	return c.Name
}

// NewHTTPChecker initializes a new HTTPChecker with its specific configuration.
func NewHTTPChecker(name, address, method string, headers map[string]string, expectedStatusCodes []int, timeout time.Duration) (*HTTPChecker, error) {
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
		DialTimeout:         timeout,
	}, nil
}

// Check performs an HTTP request and checks the response.
func (c *HTTPChecker) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, c.Method, c.Address, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range c.Headers {
		req.Header.Add(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for _, code := range c.ExpectedStatusCodes {
		if resp.StatusCode == code {
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
			code, err := strconv.Atoi(r)
			if err != nil {
				return nil, fmt.Errorf("invalid status code: %s", r)
			}
			statusCodes = append(statusCodes, code)
			continue
		}

		parts := strings.Split(r, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid status range: %s", r)
		}

		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])
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
func parseHeaders(headers string) map[string]string {
	headerMap := make(map[string]string)
	if headers == "" {
		return headerMap
	}

	pairs := strings.Split(headers, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return headerMap
}
