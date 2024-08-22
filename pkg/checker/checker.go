package checker

import (
	"context"
	"fmt"
	"time"
)

// Checker is an interface that defines methods to perform a check.
type Checker interface {
	Check(ctx context.Context) error // Check performs a check and returns an error if the check fails.
	String() string                  // String returns the name of the checker.
}

// NewChecker creates a new Checker based on the check type.
func NewChecker(checkType, name, address string, timeout time.Duration, getEnv func(string) string) (Checker, error) {
	switch checkType {
	case "http":
		return NewHTTPChecker(name, address, timeout, getEnv)
	case "tcp":
		return NewTCPChecker(name, address, timeout, getEnv)
	default:
		return nil, fmt.Errorf("invalid check type: %s", checkType)
	}
}

// IsValidCheckType validates if the check type is supported.
func IsValidCheckType(checkType string) bool {
	return checkType == "tcp" || checkType == "http"
}

// InferCheckType infers the check type based on the scheme of the target address. If no scheme is provided, it defaults to TCP.
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
