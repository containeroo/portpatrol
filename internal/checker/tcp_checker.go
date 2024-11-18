package checker

import (
	"context"
	"net"
	"time"
)

const defaultTCPTimeout time.Duration = 1 * time.Second

// TCPChecker implements the Checker interface for TCP checks.
type TCPChecker struct {
	name    string
	address string
	timeout time.Duration
	dialer  *net.Dialer
}

func (c *TCPChecker) GetAddress() string { return c.address }
func (c *TCPChecker) GetName() string    { return c.name }
func (c *TCPChecker) GetType() string    { return TCP.String() }
func (c *TCPChecker) Check(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// newTCPChecker creates a new TCPChecker with functional options.
func newTCPChecker(name, address string, opts ...Option) (*TCPChecker, error) {
	checker := &TCPChecker{
		name:    name,
		address: address,
		timeout: defaultTCPTimeout,
		dialer: &net.Dialer{
			Timeout: defaultTCPTimeout,
		},
	}

	for _, opt := range opts {
		opt.apply(checker)
	}

	return checker, nil
}

// WithTCPTimeout sets the timeout for the TCPChecker.
func WithTCPTimeout(timeout time.Duration) Option {
	return OptionFunc(func(c Checker) {
		if tcpChecker, ok := c.(*TCPChecker); ok {
			tcpChecker.timeout = timeout
		}
	})
}
