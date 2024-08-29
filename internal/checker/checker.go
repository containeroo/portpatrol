package checker

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CheckerConstructor is a function type for creating Checkers.
type CheckerConstructor func(name, address string, timeout time.Duration, getEnv func(string) string) (Checker, error)

// CheckerFactory is a map that stores the different Checker factories.
var CheckerFactory = map[string]CheckerConstructor{
	"http":  NewHTTPChecker,
	"https": NewHTTPChecker,
	"tcp":   NewTCPChecker,
	"icmp":  NewICMPChecker,
}

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

// NewChecker creates a Checker based on the provided check type.
func NewChecker(checkType, name, address string, timeout time.Duration, getEnv func(string) string) (Checker, error) {
	factory, found := CheckerFactory[checkType]
	if !found {
		return nil, fmt.Errorf("unknown check type: %s", checkType)
	}

	return factory(name, address, timeout, getEnv)
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
		return "", nil // No scheme provided, return an empty string and no error
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
