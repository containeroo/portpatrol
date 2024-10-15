package checker

import (
	"context"
	"fmt"
	"strings"
)

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

type CheckerConfig interface {
	// Marker interface; no methods required
}

// Checker defines an interface for performing various types of checks, such as TCP, HTTP, or ICMP.
// It provides methods for executing the check and obtaining a string representation of the checker.
type Checker interface {
	// Check performs a check and returns an error if the check fails.
	Check(ctx context.Context) error

	// String returns the name of the checker.
	String() string
}

func NewChecker(checkType CheckType, name, address string, config CheckerConfig) (Checker, error) {
	switch checkType {
	case HTTP:
		httpConfig, ok := config.(HTTPCheckerConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config for HTTP checker")
		}
		return NewHTTPChecker(name, address, httpConfig)
	case TCP:
		tcpConfig, ok := config.(TCPCheckerConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config for TCP checker")
		}
		return NewTCPChecker(name, address, tcpConfig)
	case ICMP:
		icmpConfig, ok := config.(ICMPCheckerConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config for ICMP checker")
		}
		return NewICMPChecker(name, address, icmpConfig)
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
