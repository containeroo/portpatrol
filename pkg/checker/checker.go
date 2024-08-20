package checker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Checker is an interface that defines methods to perform a check.
type Checker interface {
	Check(ctx context.Context) error
	String() string
}

// NewChecker creates a new Checker based on the check type
func NewChecker(ctx context.Context, checkType, name, address string, timeout time.Duration, getEnv func(string) string) (Checker, error) {
	var checker Checker
	var err error

	switch checkType {
	case "http":
		name, err = validateHTTPChecker(name, address)
		if err != nil {
			return nil, err
		}

		statusCodes := getEnv(envExpectedStatuses)
		if statusCodes == "" {
			statusCodes = defaultExpectedStatus
		}
		expectedStatusCodes, err := parseExpectedStatuses(statusCodes)
		if err != nil {
			return nil, fmt.Errorf("invalid EXPECTED_STATUSES value: %w", err)
		}

		method := getEnv(envMethod)
		if method == "" {
			method = http.MethodGet
		}

		headers := parseHeaders(getEnv(envHeaders))

		checker, err = NewHTTPChecker(name, address, method, headers, expectedStatusCodes, timeout)
		if err != nil {
			return nil, err
		}
	case "tcp":
		name, err = validateTCPChecker(name, address)
		if err != nil {
			return nil, err
		}

		checker, err = NewTCPChecker(name, address, timeout, getEnv)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid check type: %s", checkType)
	}

	return checker, nil
}

// InferCheckType infers the check type based on the scheme of the target address.
func InferCheckType(address string) (string, error) {
	scheme, _ := extractScheme(address)

	switch scheme {
	case "http", "https":
		return "http", nil
	case "tcp":
		return "tcp", nil
	case "":
		// If no scheme is provided, default to TCP.
		return "tcp", nil
	default:
		return "", fmt.Errorf("unsupported scheme: %s", scheme)
	}
}

// validateHTTPChecker validates the configuration for an HTTP checker.
func validateHTTPChecker(name, address string) (string, error) {
	url, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("invalid HTTP address: %w", err)
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return "", fmt.Errorf("HTTP address must include a valid scheme (http or https)")
	}

	if url.Hostname() == "" {
		return "", fmt.Errorf("invalid HTTP address: missing hostname")
	}

	if name == "" {
		name = url.Hostname()
	}

	return name, nil
}

// validateTCPChecker validates the configuration for a TCP checker.
func validateTCPChecker(name, address string) (string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", fmt.Errorf("invalid TCP address: %w", err)
	}

	if host == "" {
		return "", fmt.Errorf("invalid TCP address: missing host")
	}

	if port == "" {
		return "", fmt.Errorf("invalid TCP address: missing port")
	}

	if name == "" {
		name = host
	}

	return name, nil
}
