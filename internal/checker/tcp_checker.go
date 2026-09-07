package checker

import (
	"context"
	"net"
	"strings"
	"time"
)

const defaultTCPTimeout time.Duration = 2 * time.Second

// TCPChecker implements the Checker interface for TCP checks.
type TCPChecker struct {
	name    string
	address string
	dialer  *net.Dialer
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

// NewChecker constructs a TCP checker from the config.
func (c TCPConfig) NewChecker(name, address string) (Checker, error) {
	return &TCPChecker{
		name:    name,
		address: strings.TrimSpace(address),
		dialer:  &net.Dialer{Timeout: c.Timeout},
	}, nil
}

func DefaultTCPConfig() TCPConfig { return TCPConfig{Timeout: defaultTCPTimeout} }
