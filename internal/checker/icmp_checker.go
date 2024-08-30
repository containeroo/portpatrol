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
	envICMPReadTimeout = "ICMP_READ_TIMEOUT"

	defaultICMPReadTimeout = "1s"
)

// Dialer is an interface that abstracts the DialContext method.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// ICMPChecker implements a basic ICMP ping checker.
type ICMPChecker struct {
	Name        string // The name of the checker.
	Address     string // The address of the target.
	Protocol    ICMPProtocol
	dialer      Dialer        // The dialer to use for the connection.
	ReadTimeout time.Duration // The timeout for reading the ICMP reply.
}

// String returns the name of the checker.
func (c *ICMPChecker) String() string {
	return c.Name
}

// NewICMPChecker initializes a new ICMPChecker with its specific configuration.
func NewICMPChecker(name, address string, dialTimeout time.Duration, getEnv func(string) string) (Checker, error) {
	address = strings.TrimPrefix(address, "icmp://")

	protocol, err := NewProtocol(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP protocol: %w", err)
	}

	// Determine the read timeout
	readTimeoutStr := getEnv(envICMPReadTimeout)
	if readTimeoutStr == "" {
		readTimeoutStr = defaultICMPReadTimeout
	}

	// Parse the duration value for readTimeout
	readTimeout, err := time.ParseDuration(readTimeoutStr)
	if err != nil || readTimeout <= 0 {
		return nil, fmt.Errorf("invalid %s value: %s", envICMPReadTimeout, err)
	}

	dialer := &net.Dialer{
		Timeout: dialTimeout,
	}

	return &ICMPChecker{
		Name:        name,
		Address:     address,
		Protocol:    protocol,
		dialer:      dialer,
		ReadTimeout: readTimeout,
	}, nil
}

// Check sends an ICMP echo request and waits for a reply, respecting the provided context for cancellation.
func (c *ICMPChecker) Check(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, c.Protocol.Network(), c.Address)
	if err != nil {
		return fmt.Errorf("failed to dial ICMP address %s: %w", c.Address, err)
	}
	defer conn.Close()

	// Set the read deadline
	if err := conn.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	identifier := uint16(os.Getpid() & 0xffff)                    // Retrieve the process ID and take the lower 16 bits
	sequence := uint16(atomic.AddUint32(new(uint32), 1) & 0xffff) // Increment the sequence number and take the lower 16 bits

	msg := c.Protocol.MakeRequest(identifier, sequence)

	if _, err := conn.Write(msg); err != nil {
		return fmt.Errorf("failed to send ICMP request to %s: %w", c.Address, err)
	}

	reply := make([]byte, 1500)
	n, err := conn.Read(reply)

	// Check if the context was cancelled
	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled while waiting for ICMP reply from %s: %w", c.Address, ctx.Err())
	}

	if err != nil {
		return fmt.Errorf("failed to read ICMP reply from %s: %w", c.Address, err)
	}

	if err := c.Protocol.ValidateReply(reply[:n], identifier, sequence); err != nil {
		return err
	}

	return nil
}

// calculateChecksum calculates the Internet checksum as defined by RFC 1071.
// This checksum is used in various Internet protocols, including ICMP.
// The checksum is computed over the data in 16-bit words, and any overflow bits are folded back into the sum.
func calculateChecksum(data []byte) uint16 {
	var sum uint32

	// Process each pair of bytes (16 bits) in the data.
	// The checksum is computed by adding these 16-bit values together.
	for i := 0; i < len(data)-1; i += 2 {
		// Combine two adjacent bytes into a 16-bit word and add it to the sum.
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}

	// If there's an odd number of bytes, pad the last byte with zeros to form the final 16-bit word.
	// This is necessary because the checksum is defined over 16-bit words.
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}

	// Fold the carry bits from the upper 16 bits of the sum into the lower 16 bits.
	// This is done by adding the high 16 bits to the low 16 bits until there are no more carry bits.
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)

	// The final checksum is the one's complement of the sum (i.e., invert all bits).
	// This ensures that a valid checksum over the entire message results in a value of 0xFFFF.
	return uint16(^sum)
}
