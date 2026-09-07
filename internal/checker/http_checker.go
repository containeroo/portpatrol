package checker

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout         time.Duration = 2 * time.Second
	defaultHTTPMethod          string        = http.MethodGet
	defaultHTTPFollowRedirects bool          = true
	defaultHTTPMaxRedirects    int           = 10
)

// HTTPChecker implements the Checker interface for HTTP checks.
type HTTPChecker struct {
	name                string
	address             string
	method              string
	headers             http.Header
	expectedStatusCodes []int
	followRedirects     bool
	maxRedirects        int
	skipTLSVerify       bool
	timeout             time.Duration
	client              *http.Client
}

// NewHTTPChecker constructs an HTTP checker from explicit protocol settings.
func NewHTTPChecker(name, address string, cfg HTTPConfig) (*HTTPChecker, error) {
	address = strings.TrimSpace(address)
	if err := validateHTTPAddress(address); err != nil {
		return nil, err
	}
	checker := &HTTPChecker{
		name:                name,
		address:             address,
		method:              cfg.Method,
		headers:             cfg.Headers.Clone(),
		expectedStatusCodes: slices.Clone(cfg.ExpectedStatusCodes),
		followRedirects:     cfg.FollowRedirects,
		maxRedirects:        cfg.MaxRedirects,
		skipTLSVerify:       cfg.SkipTLSVerify,
		timeout:             cfg.Timeout,
	}

	checker.client = &http.Client{
		Timeout: checker.timeout,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if !checker.followRedirects || checker.maxRedirects == 0 {
				return http.ErrUseLastResponse
			}
			if len(via) > checker.maxRedirects {
				return fmt.Errorf("stopped after %d redirects", checker.maxRedirects)
			}

			return nil
		},
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: checker.skipTLSVerify,
			},
		},
	}

	return checker, nil
}

// Address returns the checker address.
func (c *HTTPChecker) Address() string { return c.address }

// Name returns the checker name.
func (c *HTTPChecker) Name() string { return c.name }

// Type returns the checker type.
func (c *HTTPChecker) Type() string { return HTTP.String() }

// Check performs the checker operation.
func (c *HTTPChecker) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, c.method, c.address, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	if slices.Contains(c.expectedStatusCodes, resp.StatusCode) {
		return nil
	}

	return fmt.Errorf("unexpected status code: got %d, expected one of %v", resp.StatusCode, c.expectedStatusCodes)
}

// HTTPConfig contains only HTTP-specific settings.
type HTTPConfig struct {
	Method              string
	Headers             http.Header
	ExpectedStatusCodes []int
	FollowRedirects     bool
	MaxRedirects        int
	SkipTLSVerify       bool
	Timeout             time.Duration
}

// DefaultHTTPConfig returns an independent configuration with the application defaults.
func DefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Method:              defaultHTTPMethod,
		Headers:             make(http.Header),
		ExpectedStatusCodes: []int{200},
		FollowRedirects:     defaultHTTPFollowRedirects,
		MaxRedirects:        defaultHTTPMaxRedirects,
		Timeout:             defaultHTTPTimeout,
	}
}
