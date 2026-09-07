package checker

import "context"

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

// Config constructs one concrete checker implementation.
type Config interface {
	NewChecker(name, address string) (Checker, error)
}
