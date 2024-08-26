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

func NewTCPChecker(name, address string, timeout time.Duration, getEnv func(string) string) (Checker, error) {
	address = strings.TrimPrefix(address, "tcp://")

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	return &TCPChecker{
		Name:    name,
		Address: address,
		dialer:  dialer,
	}, nil
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