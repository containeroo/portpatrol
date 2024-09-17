package checker

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CheckType is an enumeration that represents the type of check being performed.
type CheckType int

const (
	TCP  CheckType = iota // TCP represents a check over the TCP protocol.
	HTTP                  // HTTP represents a check over the HTTP protocol.
	ICMP                  // ICMP represents a check using the ICMP protocol (ping).
)

// String returns the string representation of the CheckType.
func (c CheckType) String() string {
	return [...]string{"TCP", "HTTP", "ICMP"}[c]
}

// Checker is an interface that defines methods to perform a check.
type Checker interface {
	Check(ctx context.Context) error // Check performs a check and returns an error if the check fails.
	String() string                  // String returns the name of the checker.
}

// Factory function that returns the appropriate Checker based on checkType
func NewChecker(checkType CheckType, name, address string, timeout time.Duration, getEnv func(string) string) (Checker, error) {
	switch checkType {
	case HTTP: // HTTP and HTTPS checkers may need environment variables for proxy settings, etc.
		return NewHTTPChecker(name, address, timeout, getEnv)
	case TCP: // TCP checkers may not need environment variables
		return NewTCPChecker(name, address, timeout)
	case ICMP: // ICMP checkers may have a different timeout logic
		return NewICMPChecker(name, address, timeout, getEnv)
	default:
		return nil, fmt.Errorf("unsupported check type: %d", checkType)
	}
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
