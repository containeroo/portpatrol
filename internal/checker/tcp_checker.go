package checker

import (
	"context"
	"net"
	"time"
)

const defaultTCPTimeout time.Duration = 2 * time.Second

// TCPChecker implements the Checker interface for TCP checks.
type TCPChecker struct {
	name    string
	address string
	dialer  *net.Dialer
}

// NewTCPChecker constructs a TCP checker from explicit protocol settings.
func NewTCPChecker(name, address string, cfg TCPConfig) (*TCPChecker, error) {
	address = normalizeAddress(address)
	return &TCPChecker{
		name:    name,
		address: address,
		dialer:  &net.Dialer{Timeout: cfg.Timeout},
	}, nil
}

// Address returns the checker address.
func (c *TCPChecker) Address() string { return c.address }

// Name returns the checker name.
func (c *TCPChecker) Name() string { return c.name }

// Type returns the checker type.
func (c *TCPChecker) Type() string { return TCP.String() }

// Check performs the checker operation.
func (c *TCPChecker) Check(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return err
	}
	defer conn.Close() // nolint:errcheck
	return nil
}

// TCPConfig contains TCP connection settings.
type TCPConfig struct{ Timeout time.Duration }

func DefaultTCPConfig() TCPConfig { return TCPConfig{Timeout: defaultTCPTimeout} }
