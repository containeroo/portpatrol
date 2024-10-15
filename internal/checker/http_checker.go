package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultHTTPDialTimeout time.Duration = 1 * time.Second
	defaultHTTPMethod                    = http.MethodGet
)

var defaultHTTPExpectedStatusCodes = []int{200} // Slices cannot be constants

// HTTPChecker implements the Checker interface for HTTP checks.
type HTTPChecker struct {
	name                  string            // The name of the checker.
	address               string            // The address of the target.
	expectedStatusCodes   []int             // The expected status codes.
	method                string            // The HTTP method to use.
	headers               map[string]string // The HTTP headers to include in the request.
	allowDuplicateHeaders bool              // Whether to allow duplicate headers.
	skipTLSVerify         bool              // Whether to skip TLS verification.
	dialTimeout           time.Duration     // The timeout for dialing the target.

	client *http.Client // The HTTP client to use for the request.
}

type HTTPCheckerConfig struct {
	Interval            time.Duration
	Method              string
	Headers             map[string]string
	ExpectedStatusCodes []int
	SkipTLSVerify       bool
	Timeout             time.Duration
}

// String returns the name of the checker.
func (c *HTTPChecker) String() string {
	return c.name
}

// NewHTTPChecker creates a new HTTPChecker with default values and applies any provided options.
func NewHTTPChecker(name, address string, cfg HTTPCheckerConfig) (Checker, error) {
	// Set defaults if necessary
	method := cfg.Method
	if method == "" {
		method = defaultHTTPMethod
	}

	expectedStatusCodes := cfg.ExpectedStatusCodes
	if len(expectedStatusCodes) == 0 {
		expectedStatusCodes = defaultHTTPExpectedStatusCodes
	}

	headers := cfg.Headers
	if headers == nil {
		headers = make(map[string]string)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultHTTPDialTimeout
	}

	// Initialize the HTTP client
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipTLSVerify,
			},
		},
	}

	checker := &HTTPChecker{
		name:                name,
		address:             address,
		method:              method,
		expectedStatusCodes: expectedStatusCodes,
		headers:             headers,
		skipTLSVerify:       cfg.SkipTLSVerify,
		dialTimeout:         timeout,
		client:              client,
	}

	return checker, nil
}

// Check performs an HTTP request and checks the response.
func (c *HTTPChecker) Check(ctx context.Context) error {
	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, c.method, c.address, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to the request
	for key, value := range c.headers {
		req.Header.Add(key, value)
	}

	// Perform the HTTP request
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	for _, code := range c.expectedStatusCodes {
		if resp.StatusCode == code {
			return nil // Return nil if the status code matches
		}
	}

	return fmt.Errorf("unexpected status code: got %d, expected one of %v", resp.StatusCode, c.expectedStatusCodes)
}
