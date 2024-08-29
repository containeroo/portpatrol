package checker

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
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
	Name        string        // The name of the checker.
	Address     string        // The address of the target.
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
		dialer:      dialer,
		ReadTimeout: readTimeout,
	}, nil
}

// Check sends an ICMP echo request and waits for a reply, respecting the provided context for cancellation.
func (c *ICMPChecker) Check(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, "ip4:icmp", c.Address)
	if err != nil {
		return fmt.Errorf("failed to dial ICMP address %s: %w", c.Address, err)
	}
	defer conn.Close()

	// Set the read deadline separately
	if err := conn.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Generate identifier and sequence number
	identifier := uint16(os.Getpid() & 0xffff)
	sequence := uint16(time.Now().UnixNano() & 0xffff) // Use a timestamp-based sequence number

	// Create ICMP echo request packet
	msg := makeICMPEchoRequest(identifier, sequence)

	// Create a channel to receive the result of the goroutine
	done := make(chan error, 1)

	// Run the ping in a separate goroutine
	go func() {
		if _, err := conn.Write(msg); err != nil {
			done <- fmt.Errorf("failed to send ICMP request to %s: %w", c.Address, err)
			return
		}

		// Wait for a reply
		reply := make([]byte, 64) // Reduced buffer size for reply because we only need the first 64 bytes
		n, err := conn.Read(reply)
		if err != nil {
			done <- fmt.Errorf("failed to read ICMP reply from %s: %w", c.Address, err)
			return
		}

		// Check if the reply is an ICMP Echo Reply and matches identifier and sequence number
		if err := validateICMPEchoReply(reply[:n], identifier, sequence); err != nil {
			done <- err
			return
		}

		done <- nil
	}()

	select {
	case <-ctx.Done():
		// If context is canceled, return the context's error
		return fmt.Errorf("context cancelled while waiting for ICMP reply from %s: %w", c.Address, ctx.Err())
	case err := <-done:
		// If the goroutine finishes before the context is canceled, return the goroutine's result
		return err
	}
}

// makeICMPEchoRequest creates an ICMP Echo Request packet.
func makeICMPEchoRequest(identifier, sequence uint16) []byte {
	// ICMP header (8 bytes)
	header := make([]byte, 8)
	header[0] = 8                                      // ICMP Echo Request
	header[1] = 0                                      // Code 0
	binary.BigEndian.PutUint16(header[4:], identifier) // Identifier
	binary.BigEndian.PutUint16(header[6:], sequence)   // Sequence number

	// Calculate checksum
	checksum := calculateChecksum(header)
	binary.BigEndian.PutUint16(header[2:], checksum) // Set checksum in header

	return header
}

// validateICMPEchoReply checks if the reply is an ICMP Echo Reply and validates it.
func validateICMPEchoReply(reply []byte, identifier, sequence uint16) error {
	if len(reply) < 20+8 { // 20 bytes for IP header + 8 bytes for ICMP header
		return fmt.Errorf("reply too short, not a valid ICMP echo reply")
	}

	if reply[20] != 0 { // ICMP Echo Reply type
		return fmt.Errorf("unexpected ICMP reply type: %d", reply[20])
	}

	recvIdentifier := binary.BigEndian.Uint16(reply[24:26])
	recvSequence := binary.BigEndian.Uint16(reply[26:28])

	if recvIdentifier != identifier || recvSequence != sequence {
		return fmt.Errorf("identifier or sequence number mismatch: got id=%d seq=%d, expected id=%d seq=%d",
			recvIdentifier, recvSequence, identifier, sequence)
	}

	return nil
}

// calculateChecksum calculates the checksum of an ICMP packet.
func calculateChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}

	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	return uint16(^sum)
}
