package checker

import (
	"context"
	"fmt"
	"time"
)

const defaultCheckType = "tcp"

// Checker is an interface that defines methods to perform a check.
type Checker interface {
	Check(ctx context.Context) error // Check performs a check and returns an error if the check fails.
	String() string                  // String returns the name of the checker.
}

// checkerFunc is a factory function that creates a Checker.
type checkerFunc func(name, address string, timeout time.Duration, getEnv func(string) string) (Checker, error)

// CheckerFactory is a map that acts as a registry for checker types.
var CheckerFactory = make(map[string]checkerFunc)

// TypeToSchemes maps checker types to multiple URL schemes.
var TypeToSchemes = make(map[string][]string)

// RegisterChecker registers a new checker factory with a check type and its associated schemes.
func RegisterChecker(checkType string, factory checkerFunc, schemes ...string) {
	CheckerFactory[checkType] = factory
	TypeToSchemes[checkType] = schemes
}

// NewChecker creates a new Checker based on the check type.
func NewChecker(checkType, name, address string, timeout time.Duration, getEnv func(string) string) (Checker, error) {
	factory, found := CheckerFactory[checkType]
	if !found {
		return nil, fmt.Errorf("invalid check type: %s", checkType)
	}
	return factory(name, address, timeout, getEnv)
}

// IsValidCheckType checks if the given check type is supported.
func IsValidCheckType(checkType string) bool {
	_, found := CheckerFactory[checkType]
	return found
}

// InferCheckType infers the check type based on the scheme of the target address.
func InferCheckType(address string) (string, error) {
	scheme, _ := extractScheme(address)

	for checkType, schemes := range TypeToSchemes {
		for _, s := range schemes {
			if scheme == s {
				return checkType, nil
			}
		}
	}

	// Default to TCP if no scheme is provided or recognized
	if scheme == "" {
		return defaultCheckType, nil
	}

	return "", fmt.Errorf("unsupported scheme: %s", scheme)
}
