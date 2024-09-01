package checker

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// Protocol defines the interface for an ICMP protocol.
type Protocol interface {
	MakeRequest(identifier, sequence uint16) ([]byte, error)
	ValidateReply(reply []byte, identifier, sequence uint16) error
	Network() string
	ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error)
	SetDeadline(t time.Time) error
}

// newProtocol creates a new ICMP protocol based on the given address.
func newProtocol(address string) (Protocol, error) {
	ip := net.ParseIP(address)
	if ip == nil {
		// If the address is not an IP, try resolving it as a domain name
		ips, err := net.LookupIP(address)
		if err != nil || len(ips) == 0 {
			return nil, fmt.Errorf("invalid or unresolvable address: %s", address)
		}
		ip = ips[0] // Use the first resolved IP address
	}

	if ip.To16() != nil && ip.To4() == nil {
		return &ICMPv6{}, nil
	}

	return &ICMPv4{}, nil
}

// ICMPv4 implements the ICMP protocol for IPv4.
type ICMPv4 struct {
	conn net.PacketConn
}

// MakeRequest creates an ICMP echo request message.
func (p *ICMPv4) MakeRequest(identifier, sequence uint16) ([]byte, error) {
	body := &icmp.Echo{
		ID:   int(identifier),
		Seq:  int(sequence),
		Data: []byte("HELLO-R-U-THERE"),
	}
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: body,
	}

	return msg.Marshal(nil)
}

// ValidateReply validates an ICMP echo reply message.
func (p *ICMPv4) ValidateReply(reply []byte, identifier, sequence uint16) error {
	parsedMsg, err := icmp.ParseMessage(1, reply)
	if err != nil {
		return fmt.Errorf("failed to parse ICMPv4 message: %w", err)
	}

	if parsedMsg.Type != ipv4.ICMPTypeEchoReply {
		return fmt.Errorf("unexpected ICMPv4 message type: %v", parsedMsg.Type)
	}

	body, ok := parsedMsg.Body.(*icmp.Echo)
	if !ok || body.ID != int(identifier) || body.Seq != int(sequence) {
		return fmt.Errorf("identifier or sequence mismatch")
	}

	return nil
}

// Network returns the network type for the ICMP protocol.
func (p *ICMPv4) Network() string {
	return "ip4:icmp"
}

// ListenPacket creates a new ICMPv4 packet connection.
func (p *ICMPv4) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	var lc net.ListenConfig
	conn, err := lc.ListenPacket(ctx, network, address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen for ICMP packets: %w", err)
	}
	p.conn = conn

	return p.conn, nil
}

// SetDeadline sets the deadline for the ICMPv4 packet connection.
func (p *ICMPv4) SetDeadline(t time.Time) error {
	return p.conn.SetDeadline(t)
}

// ICMPv6 implements the ICMP protocol for IPv6.
type ICMPv6 struct {
	conn net.PacketConn
}

// MakeRequest creates an ICMP echo request message.
func (p *ICMPv6) MakeRequest(identifier, sequence uint16) ([]byte, error) {
	body := &icmp.Echo{
		ID:   int(identifier),
		Seq:  int(sequence),
		Data: []byte("HELLO-R-U-THERE"),
	}
	msg := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Code: 0,
		Body: body,
	}

	return msg.Marshal(nil)
}

// ValidateReply validates an ICMP echo reply message.
func (p *ICMPv6) ValidateReply(reply []byte, identifier, sequence uint16) error {
	parsedMsg, err := icmp.ParseMessage(58, reply)
	if err != nil {
		return fmt.Errorf("failed to parse ICMPv6 message: %w", err)
	}

	if parsedMsg.Type != ipv6.ICMPTypeEchoReply {
		return fmt.Errorf("unexpected ICMPv6 message type: %v", parsedMsg.Type)
	}

	body, ok := parsedMsg.Body.(*icmp.Echo)
	if !ok || body.ID != int(identifier) || body.Seq != int(sequence) {
		return fmt.Errorf("identifier or sequence mismatch")
	}

	return nil
}

// Network returns the network type for the ICMP protocol.
func (p *ICMPv6) Network() string {
	return "ip6:ipv6-icmp"
}

// ListenPacket creates a new ICMPv6 packet connection.
func (p *ICMPv6) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	var lc net.ListenConfig
	conn, err := lc.ListenPacket(ctx, network, address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen for ICMP packets: %w", err)
	}
	p.conn = conn

	return p.conn, nil
}

// SetDeadline sets the deadline for the ICMPv6 packet connection.
func (p *ICMPv6) SetDeadline(t time.Time) error {
	return p.conn.SetDeadline(t)
}
