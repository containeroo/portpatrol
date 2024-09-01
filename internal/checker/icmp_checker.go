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

// ICMPChecker implements a basic ICMP ping checker.
type ICMPChecker struct {
	Name        string        // The name of the checker.
	Address     string        // The address of the target.
	Protocol    Protocol      // The protocol to use for the connection.
	ReadTimeout time.Duration // The timeout for reading the ICMP reply.
}

// String returns the name of the checker.
func (c *ICMPChecker) String() string {
	return c.Name
}

// NewICMPChecker initializes a new ICMPChecker with its specific configuration.
func NewICMPChecker(name, address string, dialTimeout time.Duration, getEnv func(string) string) (Checker, error) {
	address = strings.TrimPrefix(address, "icmp://")
	protocol, err := newProtocol(address)
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP protocol: %w", err)
	}

	readTimeoutStr := getEnv(envICMPReadTimeout)
	if readTimeoutStr == "" {
		readTimeoutStr = defaultICMPReadTimeout
	}

	readTimeout, err := time.ParseDuration(readTimeoutStr)
	if err != nil || readTimeout <= 0 {
		return nil, fmt.Errorf("invalid %s value: %s", envICMPReadTimeout, readTimeoutStr)
	}

	return &ICMPChecker{
		Name:        name,
		Address:     address,
		Protocol:    protocol,
		ReadTimeout: readTimeout,
	}, nil
}

// Check performs a ICMP check on the target.
func (c *ICMPChecker) Check(ctx context.Context) error {
	dst, err := net.ResolveIPAddr(c.Protocol.Network(), c.Address)
	if err != nil {
		return fmt.Errorf("failed to resolve IP address: %w", err)
	}

	conn, err := c.Protocol.ListenPacket(ctx, c.Protocol.Network(), "0.0.0.0")
	if err != nil {
		return fmt.Errorf("failed to listen for ICMP packets: %w", err)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	identifier := uint16(os.Getpid() & 0xffff)
	sequence := uint16(atomic.AddUint32(new(uint32), 1) & 0xffff)

	msg, err := c.Protocol.MakeRequest(identifier, sequence)
	if err != nil {
		return fmt.Errorf("failed to create ICMP request: %w", err)
	}

	if _, err := conn.WriteTo(msg, dst); err != nil {
		return fmt.Errorf("failed to send ICMP request to %s: %w", c.Address, err)
	}

	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		return fmt.Errorf("failed to read ICMP reply from %s: %w", c.Address, err)
	}

	if err := c.Protocol.ValidateReply(reply[:n], identifier, sequence); err != nil {
		return err
	}

	return nil
}
