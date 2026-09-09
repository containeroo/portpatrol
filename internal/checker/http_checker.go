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
	defaultHTTPFollowRedirects bool          = true
	defaultHTTPMaxRedirects    int           = 10
	defaultHTTPMethod          string        = http.MethodGet
	defaultHTTPSTatusCode      int           = 200
	defaultHTTPTimeout         time.Duration = 2 * time.Second
)

// HTTPChecker implements the Checker interface for HTTP checks.
type HTTPChecker struct {
	name                string
	requestURL          string
	displayAddress      string
	method              string
	headers             http.Header
	expectedStatusCodes []int
	followRedirects     bool
	maxRedirects        int
	skipTLSVerify       bool
	timeout             time.Duration
	client              *http.Client
}

// NewChecker constructs an HTTP checker from the config.
func (c HTTPConfig) NewChecker(name, address string) (Checker, error) {
	requestURL := strings.TrimSpace(address)
	displayAddress, err := httpDisplayAddress(requestURL, c.AddressDetail)
	if err != nil {
		return nil, err
	}

	headers := c.Headers.Clone()
	if headers == nil {
		headers = make(http.Header)
	}
	if c.UserAgent != "" && headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", c.UserAgent)
	}

	checker := &HTTPChecker{
		name:                name,
		requestURL:          requestURL,
		displayAddress:      displayAddress,
		method:              c.Method,
		headers:             headers,
		expectedStatusCodes: slices.Clone(c.ExpectedStatusCodes),
		followRedirects:     c.FollowRedirects,
		maxRedirects:        c.MaxRedirects,
		skipTLSVerify:       c.SkipTLSVerify,
		timeout:             c.Timeout,
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

// Address returns the configured HTTP address representation used in logs.
func (c *HTTPChecker) Address() string { return c.displayAddress }

// Name returns the checker name.
func (c *HTTPChecker) Name() string { return c.name }

// Type returns the checker type.
func (c *HTTPChecker) Type() string { return HTTP.String() }

// Check performs the checker operation.
func (c *HTTPChecker) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, c.method, c.requestURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header = c.headers.Clone()

	resp, err := c.client.Do(req)
	if err != nil {
		if urlErr, ok := errors.AsType[*url.Error](err); ok {
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
	UserAgent           string
	AddressDetail       HTTPAddressDetail
}

// DefaultHTTPConfig returns an independent configuration with the application defaults.
func DefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Method:              defaultHTTPMethod,
		Headers:             make(http.Header),
		ExpectedStatusCodes: []int{defaultHTTPSTatusCode},
		FollowRedirects:     defaultHTTPFollowRedirects,
		MaxRedirects:        defaultHTTPMaxRedirects,
		Timeout:             defaultHTTPTimeout,
		AddressDetail:       HTTPAddressOrigin,
	}
}

// HTTPAddressDetail controls how much of an HTTP target address is exposed through Checker.Address.
type HTTPAddressDetail string

const (
	HTTPAddressOrigin HTTPAddressDetail = "origin"
	HTTPAddressPath   HTTPAddressDetail = "path"
	HTTPAddressQuery  HTTPAddressDetail = "query"
	HTTPAddressFull   HTTPAddressDetail = "full"
)

// String returns the configured HTTP address detail.
func (d HTTPAddressDetail) String() string { return string(d) }

// httpDisplayAddress returns the HTTP address exposed through Checker.Address.
func httpDisplayAddress(address string, detail HTTPAddressDetail) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("invalid HTTP URL: %w", err)
	}

	switch detail {
	case HTTPAddressOrigin:
		u.User = nil
		u.Path = ""
		u.RawPath = ""
		u.RawQuery = ""
		u.ForceQuery = false
		u.Fragment = ""
		u.RawFragment = ""
	case HTTPAddressPath:
		u.User = nil
		u.RawQuery = ""
		u.ForceQuery = false
		u.Fragment = ""
		u.RawFragment = ""
	case HTTPAddressQuery:
		u.User = nil
		u.Fragment = ""
		u.RawFragment = ""
	case HTTPAddressFull:
		// Preserve the complete URL.
	default:
		return "", fmt.Errorf("invalid HTTP address detail: %q", detail)
	}

	return u.String(), nil
}
