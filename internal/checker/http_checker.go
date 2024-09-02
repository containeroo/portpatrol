package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/containeroo/portpatrol/pkg/httputils"
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

	// Parse the headers string into a headers map
	parsedHeaders, err := httputils.ParseHeaders(getEnv(envHTTPHeaders), allowDupHeaders)
	if err != nil {
		return nil, fmt.Errorf("invalid %s value: %w", envHTTPHeaders, err)
	}

	// Determine the expected status codes
	expectedStatusStr := getEnv(envHTTPExpectedStatusCodes)
	if expectedStatusStr == "" {
		expectedStatusStr = defaultHTTPExpectedStatus
	}

	// Parse the expected status codes string into a slice of status codes
	expectedStatusCodes, err := httputils.ParseStatusCodes(expectedStatusStr)
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
			return nil // Return nil if the status code matches
		}
	}

	return fmt.Errorf("unexpected status code: got %d, expected one of %v", resp.StatusCode, c.ExpectedStatusCodes)
}
