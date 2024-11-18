package checks

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CheckType represents the type of check to perform.
type CheckType int

const (
	TCP  CheckType = iota // TCP represents a check over the TCP protocol.
	HTTP                  // HTTP represents a check over the HTTP protocol.
	ICMP                  // ICMP represents a check using the ICMP protocol (ping).
)

const defaultCheckInterval time.Duration = 1 * time.Second

// String returns the string representation of the CheckType.
func (c CheckType) String() string {
	return [...]string{"TCP", "HTTP", "ICMP"}[c]
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
		return -1, fmt.Errorf("unsupported check type: %s", checkTypeStr)
	}
}

// NewChecker creates a new Checker based on the specified CheckType, name, address, and options.
func NewChecker(checkType CheckType, name, address string, opts ...Option) (Checker, error) {
	// Create the appropriate checker based on the type
	switch checkType {
	case HTTP:
		return newHTTPChecker(name, address, opts...)
	case TCP:
		// The "tcp://" prefix is used to identify the check type and is not needed for further processing,
		// so it must be removed before passing the address to other functions.
		address = strings.TrimPrefix(address, "tcp://")
		return newTCPChecker(name, address, opts...)
	case ICMP:
		// The "icmp://" prefix is used to identify the check type and is not needed for further processing,
		// so it must be removed before passing the address to other functions.
		address = strings.TrimPrefix(address, "icmp://")
		return newICMPChecker(name, address, opts...)
	default:
		return nil, fmt.Errorf("unsupported check type: %d", checkType)
	}
}
