package checker

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"
)

const (
	defaultICMPReadTimeout  time.Duration = 1 * time.Second
	defaultICMPWriteTimeout time.Duration = 1 * time.Second
)

// ICMPChecker implements the Checker interface for ICMP checks.
type ICMPChecker struct {
	name         string
	address      string
	readTimeout  time.Duration
	writeTimeout time.Duration
	protocol     Protocol
}

// Name returns the address of the checker.
func (c *ICMPChecker) GetAddress() string {
	return c.address
}

// Name returns the name of the checker.
func (c *ICMPChecker) GetName() string {
	return c.name
}

// Name returns the type of the checker.
func (c *ICMPChecker) GetType() string {
	return ICMP.String()
}

// newICMPChecker initializes a new ICMPChecker with functional options.
func newICMPChecker(name, address string, opts ...Option) (*ICMPChecker, error) {
	checker := &ICMPChecker{
		name:         name,
		address:      address,
		readTimeout:  defaultICMPReadTimeout,
		writeTimeout: defaultICMPWriteTimeout,
	}

	// Apply options
	for _, opt := range opts {
		opt.apply(checker)
	}

	// Initialize protocol based on address
	protocol, err := newProtocol(checker.address)
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP protocol: %w", err)
	}
	checker.protocol = protocol

	return checker, nil
}

// Check performs an ICMP check on the target.
func (c *ICMPChecker) Check(ctx context.Context) error {
	// Resolve the IP address
	dst, err := net.ResolveIPAddr(c.protocol.Network(), c.address)
	if err != nil {
		return fmt.Errorf("failed to resolve IP address '%s': %w", c.address, err)
	}

	// Listen for ICMP packets
	conn, err := c.protocol.ListenPacket(ctx, c.protocol.Network(), "")
	if err != nil {
		return fmt.Errorf("failed to listen for ICMP packets: %w", err)
	}
	defer conn.Close()

	identifier := uint16(os.Getpid() & 0xffff)                    // Create a unique identifier
	sequence := uint16(atomic.AddUint32(new(uint32), 1) & 0xffff) // Create a unique sequence number

	// Make the ICMP request
	msg, err := c.protocol.MakeRequest(identifier, sequence)
	if err != nil {
		return fmt.Errorf("failed to create ICMP request: %w", err)
	}

	// Set write deadline
	if err := conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Send the ICMP request
	if _, err := conn.WriteTo(msg, dst); err != nil {
		return fmt.Errorf("failed to send ICMP request: %w", err)
	}

	// Set read deadline
	if err := conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read the ICMP reply
	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		return fmt.Errorf("failed to read ICMP reply: %w", err)
	}

	// Validate the ICMP reply
	if err := c.protocol.ValidateReply(reply[:n], identifier, sequence); err != nil {
		return fmt.Errorf("failed to validate ICMP reply: %w", err)
	}

	return nil
}

// WithICMPReadTimeout sets the read timeout for the ICMPChecker.
func WithICMPReadTimeout(timeout time.Duration) Option {
	return OptionFunc(func(c Checker) {
		if icmpChecker, ok := c.(*ICMPChecker); ok {
			icmpChecker.readTimeout = timeout
		}
	})
}

// WithICMPWriteTimeout sets the write timeout for the ICMPChecker.
func WithICMPWriteTimeout(timeout time.Duration) Option {
	return OptionFunc(func(c Checker) {
		if icmpChecker, ok := c.(*ICMPChecker); ok {
			icmpChecker.writeTimeout = timeout
		}
	})
}
