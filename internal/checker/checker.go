package checker

import (
	"context"
	"fmt"
	"strings"
)

// CheckType represents the type of check to perform.
type CheckType string

const (
	TCP  CheckType = "TCP" // TCP represents a check over the TCP protocol.
	HTTP CheckType = "HTTP"
	ICMP CheckType = "ICMP"
)

// String returns the string representation of the CheckType.
func (c CheckType) String() string {
	return string(c)
}

// Option defines a functional option for configuring a Checker.
type Option interface {
	apply(Checker)
}

// OptionFunc is a function that applies an Option to a Checker.
type OptionFunc func(Checker)

// apply calls the OptionFunc with the given Checker.
func (f OptionFunc) apply(c Checker) {
	f(c)
}

// Checker defines an interface for performing various types of checks, such as TCP, HTTP, or ICMP.
// It provides methods for executing the check and obtaining a string representation of the checker.
type Checker interface {
	// Check performs a check and returns an error if the check fails.
	Check(ctx context.Context) error

	// GetName returns the name of the checker.
	GetName() string

	// GetType returns the type of the checker.
	GetType() string

	// GetAddress returns the address of the checker.
	GetAddress() string
}

// GetCheckTypeFromString converts a string to a CheckType enum.
func GetCheckTypeFromString(checkTypeStr string) (CheckType, error) {
	switch strings.ToLower(checkTypeStr) {
	case "http", "https":
		return HTTP, nil
	case "tcp":
		return TCP, nil
	case "icmp":
		return ICMP, nil
	default:
		return "", fmt.Errorf("unsupported check type: %s", checkTypeStr)
	}
}

// NewChecker creates a new Checker based on the specified CheckType, name, address, and options.
func NewChecker(checkType CheckType, name, address string, opts ...Option) (Checker, error) {
	// Create the appropriate checker based on the type
	switch checkType {
	case HTTP:
		return newHTTPChecker(name, address, opts...)
	case TCP:
		return newTCPChecker(name, address, opts...)
	case ICMP:
		return newICMPChecker(name, address, opts...)
	default:
		return nil, fmt.Errorf("unsupported check type: %s", checkType)
	}
}
