package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultHTTPTimeout       time.Duration = 1 * time.Second
	defaultHTTPMethod        string        = http.MethodGet
	defaultHTTPSkipTLSVerify bool          = false
)

var defaultHTTPExpectedStatusCodes = []int{200}

// HTTPChecker implements the Checker interface for HTTP checks.
type HTTPChecker struct {
	name                string
	address             string
	method              string
	headers             map[string]string
	expectedStatusCodes []int
	skipTLSVerify       bool
	timeout             time.Duration
	client              *http.Client
}

// Name returns the name of the checker.
func (c *HTTPChecker) Name() string {
	return c.name
}

// newHTTPChecker creates a new HTTPChecker with functional options.
func newHTTPChecker(name, address string, opts ...Option) (*HTTPChecker, error) {
	checker := &HTTPChecker{
		name:                name,
		address:             address,
		method:              defaultHTTPMethod,
		headers:             make(map[string]string),
		expectedStatusCodes: defaultHTTPExpectedStatusCodes,
		skipTLSVerify:       defaultHTTPSkipTLSVerify,
		timeout:             defaultHTTPTimeout,
	}

	// Apply options
	for _, opt := range opts {
		opt.apply(checker)
	}

	// Initialize the HTTP client
	checker.client = &http.Client{
		Timeout: checker.timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: checker.skipTLSVerify,
			},
		},
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

// WithHTTPMethod sets the HTTP method for the HTTPChecker.
func WithHTTPMethod(method string) Option {
	return OptionFunc(func(c Checker) {
		if httpChecker, ok := c.(*HTTPChecker); ok {
			httpChecker.method = method
		}
	})
}

// WithHTTPHeaders sets the HTTP headers for the HTTPChecker.
func WithHTTPHeaders(headers map[string]string) Option {
	return OptionFunc(func(c Checker) {
		if httpChecker, ok := c.(*HTTPChecker); ok {
			for key, value := range headers {
				httpChecker.headers[key] = value
			}
		}
	})
}

// WithExpectedStatusCodes sets the expected status codes for the HTTPChecker.
func WithExpectedStatusCodes(codes []int) Option {
	return OptionFunc(func(c Checker) {
		if httpChecker, ok := c.(*HTTPChecker); ok {
			if len(codes) > 0 {
				httpChecker.expectedStatusCodes = codes
			}
		}
	})
}

// WithHTTPSkipTLSVerify sets the TLS verification flag for the HTTPChecker.
func WithHTTPSkipTLSVerify(skip bool) Option {
	return OptionFunc(func(c Checker) {
		if httpChecker, ok := c.(*HTTPChecker); ok {
			httpChecker.skipTLSVerify = skip
		}
	})
}

// WithHTTPTimeout sets the timeout for the HTTPChecker.
func WithHTTPTimeout(timeout time.Duration) Option {
	return OptionFunc(func(c Checker) {
		if httpChecker, ok := c.(*HTTPChecker); ok {
			if timeout > 0 {
				httpChecker.timeout = timeout
			}
		}
	})
}
