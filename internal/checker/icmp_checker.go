package checker

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultICMPReadTimeout  time.Duration = 1 * time.Second
	defaultICMPWriteTimeout time.Duration = 1 * time.Second
)

type ICMPCheckerConfig struct {
	Interval     time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// ICMPChecker implements the Checker interface for ICMP checks.
type ICMPChecker struct {
	name         string        // The name of the checker.
	address      string        // The address of the target.
	readTimeout  time.Duration // The timeout for reading the ICMP reply.
	writeTimeout time.Duration // The timeout for writing the ICMP request.

	protocol Protocol // The protocol (ICMPv4 or ICMPv6) used for the check.
}

// Name returns the name of the checker.
func (c *ICMPChecker) String() string {
	return c.name
}

// NewICMPChecker initializes a new ICMPChecker with the given parameters.
func NewICMPChecker(name, address string, cfg ICMPCheckerConfig) (Checker, error) {
	// The "icmp://" prefix is used to identify the check type and is not needed for further processing,
	// so it must be removed before passing the address to other functions.
	address = strings.TrimPrefix(address, "icmp://")

	readTimeout := cfg.ReadTimeout
	if readTimeout == 0 {
		readTimeout = defaultICMPReadTimeout
	}

	writeTimeout := cfg.WriteTimeout
	if writeTimeout == 0 {
		writeTimeout = defaultICMPWriteTimeout
	}

	protocol, err := newProtocol(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP protocol: %w", err)
	}

	checker := &ICMPChecker{
		name:         name,
		address:      address,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		protocol:     protocol,
	}

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
