package checker

import (
	"context"
	"net"
	"strings"
	"time"
)

// TCPChecker implements the Checker interface for TCP checks.
type TCPChecker struct {
	Name    string      // The name of the checker.
	Address string      // The address of the target.
	dialer  *net.Dialer // The dialer to use for the connection.
}

// String returns the name of the checker.
func (c *TCPChecker) String() string {
	return c.Name
}

// NewTCPChecker creates a new TCPChecker.
func NewTCPChecker(name, address string, timeout time.Duration) (Checker, error) {
	// The "tcp://" prefix is used to identify the check type and is not needed for further processing,
	// so it must be removed before passing the address to other functions.
	address = strings.TrimPrefix(address, "tcp://")

	checker := TCPChecker{
		Address: address,
		Name:    name,
		dialer: &net.Dialer{
			Timeout: timeout,
		},
	}

	return &checker, nil
}

// Check performs a TCP connection check.
func (c *TCPChecker) Check(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, "tcp", c.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}
