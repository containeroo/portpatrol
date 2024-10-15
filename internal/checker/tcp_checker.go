package checker

import (
	"context"
	"net"
	"strings"
	"time"
)

const defaultTCPTimeout = 1 * time.Second

// TCPChecker implements the Checker interface for TCP checks.
type TCPChecker struct {
	name    string        // The name of the checker.
	address string        // The address of the target.
	timeout time.Duration // The timeout for dialing the target.

	dialer *net.Dialer // The dialer to use for the TCP connection.
}

type TCPCheckerConfig struct {
	Interval time.Duration
	Timeout  time.Duration
}

// Name returns the name of the checker.
func (c *TCPChecker) String() string {
	return c.name
}

// NewTCPChecker creates a new TCPChecker with the given parameters.
func NewTCPChecker(name, address string, cfg TCPCheckerConfig) (Checker, error) {
	// The "tcp://" prefix is used to identify the check type and is not needed for further processing,
	// so it must be removed before passing the address to other functions.
	address = strings.TrimPrefix(address, "tcp://")

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTCPTimeout
	}

	checker := &TCPChecker{
		name:    name,
		address: address,
		timeout: timeout,
		dialer: &net.Dialer{
			Timeout: timeout,
		},
	}

	return checker, nil
}

// Check performs a TCP connection check.
func (c *TCPChecker) Check(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
