package checker

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SupportedCheckTypes maps check types to their supported schemes.
var SupportedCheckTypes = map[string][]string{
	"http": {"http", "https"},
	"tcp":  {"tcp"},
	"icmp": {"icmp"},
}

// Checker is an interface that defines methods to perform a check.
type Checker interface {
	Check(ctx context.Context) error // Check performs a check and returns an error if the check fails.
	String() string                  // String returns the name of the checker.
}

// Factory function that returns the appropriate Checker based on checkType
func NewChecker(checkType, name, address string, timeout time.Duration, getEnv func(string) string) (Checker, error) {
	switch checkType {
	case "http", "https":
		// HTTP and HTTPS checkers may need environment variables for proxy settings, etc.
		return NewHTTPChecker(name, address, timeout, getEnv)
	case "tcp":
		// TCP checkers may not need environment variables
		return NewTCPChecker(name, address, timeout)
	case "icmp":
		// ICMP checkers may have a different timeout logic
		return NewICMPChecker(name, address, timeout, getEnv)
	default:
		return nil, fmt.Errorf("unsupported check type: %s", checkType)
	}
}

// IsValidCheckType validates if the check type is supported.
func IsValidCheckType(checkType string) bool {
	_, exists := SupportedCheckTypes[checkType]

	return exists
}

// InferCheckType infers the check type based on the scheme of the target address.
// It returns an empty string and no error if no scheme is provided.
// If an unsupported scheme is provided, it returns an error.
func InferCheckType(address string) (string, error) {
	scheme, _ := extractScheme(address)
	if scheme == "" {
		return "", nil
	}

	scheme = strings.ToLower(scheme) // Normalize the scheme to lowercase

	for checkType, schemes := range SupportedCheckTypes {
		for _, s := range schemes {
			if s == scheme {
				return checkType, nil
			}
		}
	}

	return "", fmt.Errorf("unsupported scheme: %s", scheme)
}
