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
	defaultICMPReadTimeout  time.Duration = 2 * time.Second
	defaultICMPWriteTimeout time.Duration = 2 * time.Second
)

var icmpSeq uint32

// ICMPChecker implements the Checker interface for ICMP checks.
type ICMPChecker struct {
	name         string
	address      string
	readTimeout  time.Duration
	writeTimeout time.Duration
	protocol     Protocol
	lookupIP     func(context.Context, string, string) ([]net.IP, error)
}

// NewICMPChecker constructs an ICMP checker without resolving its address.
func NewICMPChecker(name, address string, cfg ICMPConfig) (*ICMPChecker, error) {
	address = normalizeAddress(address)
	return &ICMPChecker{
		name:         name,
		address:      address,
		readTimeout:  cfg.ReadTimeout,
		writeTimeout: cfg.WriteTimeout,
	}, nil
}

// Address returns the checker address.
func (c *ICMPChecker) Address() string { return c.address }

// Name returns the checker name.
func (c *ICMPChecker) Name() string { return c.name }

// Type returns the checker type.
func (c *ICMPChecker) Type() string { return ICMP.String() }

// Check performs the checker operation.
func (c *ICMPChecker) Check(ctx context.Context) (result error) {
	// Bound DNS and I/O together; the phase-specific deadlines can be shorter.
	ctx, cancel := context.WithTimeout(ctx, c.readTimeout+c.writeTimeout)
	defer cancel()
	defer func() {
		if ctx.Err() != nil {
			result = ctx.Err()
		}
	}()
	if err := ctx.Err(); err != nil {
		return err
	}
	lookup := c.lookupIP
	if lookup == nil {
		lookup = net.DefaultResolver.LookupIP
	}
	ips, err := lookup(ctx, "ip", c.address)
	if err != nil {
		return fmt.Errorf("failed to resolve IP address '%s': %w", c.address, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("no IP addresses for %s", c.address)
	}
	dst := &net.IPAddr{IP: ips[0]}
	protocol := c.protocol
	if protocol == nil {
		protocol, err = newProtocol(dst.IP.String())
		if err != nil {
			return err
		}
	}
	conn, err := protocol.ListenPacket(ctx, protocol.Network(), "")
	if err != nil {
		return fmt.Errorf("failed to listen for ICMP packets: %w", err)
	}
	defer conn.Close() // nolint:errcheck
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	id := uint16(os.Getpid() & 0xffff)                    // Process-scoped identifier
	seq := uint16(atomic.AddUint32(&icmpSeq, 1) & 0xffff) // Monotonic sequence number

	msg, err := protocol.MakeRequest(id, seq)
	if err != nil {
		return fmt.Errorf("failed to create ICMP request: %w", err)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	if _, err := conn.WriteTo(msg, dst); err != nil {
		return fmt.Errorf("failed to send ICMP request: %w", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	reply := make([]byte, 1500)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, peer, err := conn.ReadFrom(reply)
		if err != nil {
			return fmt.Errorf("failed to read ICMP reply: %w", err)
		}
		source, ok := peer.(*net.IPAddr)
		if !ok || !source.IP.Equal(dst.IP) {
			continue
		}
		if err := protocol.ValidateReply(reply[:n], id, seq); err != nil {
			continue
		}
		return nil
	}
}

// ICMPConfig contains ICMP phase timeouts.
type ICMPConfig struct{ ReadTimeout, WriteTimeout time.Duration }

func DefaultICMPConfig() ICMPConfig {
	return ICMPConfig{ReadTimeout: defaultICMPReadTimeout, WriteTimeout: defaultICMPWriteTimeout}
}
