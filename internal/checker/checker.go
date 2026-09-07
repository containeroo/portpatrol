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

// Checker defines an interface for performing various types of checks, such as TCP, HTTP, or ICMP.
// It provides methods for executing the check and obtaining a string representation of the checker.
type Checker interface {
	Check(ctx context.Context) error // Check performs a check and returns an error if the check fails.
	Name() string                    // Name returns the name of the checker.
	Type() string                    // Type returns the type of the checker.
	Address() string                 // Address returns the address of the checker.
}

// addressOverride changes the address exposed for logging without changing checker behavior.
type addressOverride struct {
	Checker
	address string
}

// Address returns the overridden checker address.
func (c addressOverride) Address() string { return c.address }

// OverrideAddress returns a checker that exposes address through Address while delegating checks to checker.
func OverrideAddress(c Checker, address string) Checker {
	return addressOverride{Checker: c, address: address}
}

// ParseCheckType converts a string to a CheckType enum.
func ParseCheckType(typeStr string) (CheckType, error) {
	switch strings.ToLower(typeStr) {
	case "http", "https":
		return HTTP, nil
	case "tcp":
		return TCP, nil
	case "icmp":
		return ICMP, nil
	default:
		return "", fmt.Errorf("unsupported check type: %s", typeStr)
	}
}
