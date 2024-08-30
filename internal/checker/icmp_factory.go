package checker

import (
	"encoding/binary"
	"fmt"
	"net"
)

// ICMPProtocol is an interface that abstracts the MakeRequest and ValidateReply methods.
type ICMPProtocol interface {
	MakeRequest(identifier, sequence uint16) []byte                // MakeRequest creates an ICMP Echo Request packet.
	ValidateReply(reply []byte, identifier, sequence uint16) error // ValidateReply checks if the reply is an ICMP Echo Reply and validates it.
	Network() string                                               // Network returns the network type.
}

// NewProtocol initializes a new ICMPProtocol based on the provided address.
func NewProtocol(address string) (ICMPProtocol, error) {
	var protocol ICMPProtocol

	ip := net.ParseIP(address)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", address)
	}

	switch {
	case ip.To4() != nil:
		protocol = &ICMPv4{}
	case ip.To16() != nil && ip.To4() == nil: // Check for IPv6 and ensure it's not IPv4-mapped
		protocol = &ICMPv6{}
	}

	return protocol, nil
}

// ICMPv4 implements the ICMPProtocol interface for IPv4.
type ICMPv4 struct{}

// MakeRequest creates an ICMP Echo Request packet.
func (p *ICMPv4) MakeRequest(identifier, sequence uint16) []byte {
	header := make([]byte, 8)
	header[0] = 8                                      // ICMP Echo Request
	header[1] = 0                                      // Code 0
	binary.BigEndian.PutUint16(header[4:], identifier) // Identifier
	binary.BigEndian.PutUint16(header[6:], sequence)   // Sequence number

	checksum := calculateChecksum(header)
	binary.BigEndian.PutUint16(header[2:], checksum)

	return header
}

// ValidateReply checks if the reply is an ICMP Echo Reply and validates it.
func (p *ICMPv4) ValidateReply(reply []byte, identifier, sequence uint16) error {
	if len(reply) < 28 { // 20 bytes for IP header + 8 bytes for ICMP header
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

// Network returns the network type for IPv4.
func (p *ICMPv4) Network() string {
	return "ip4:icmp"
}

// ICMPv6 implements the ICMPProtocol interface for IPv6.
type ICMPv6 struct{}

// MakeRequest creates an ICMPv6 Echo Request packet.
func (p *ICMPv6) MakeRequest(identifier, sequence uint16) []byte {
	header := make([]byte, 8)
	header[0] = 128                                    // ICMPv6 Echo Request
	header[1] = 0                                      // Code 0
	binary.BigEndian.PutUint16(header[4:], identifier) // Identifier
	binary.BigEndian.PutUint16(header[6:], sequence)   // Sequence number

	checksum := calculateChecksum(header)
	binary.BigEndian.PutUint16(header[2:], checksum)

	return header
}

// ValidateReply checks if the reply is an ICMPv
func (p *ICMPv6) ValidateReply(reply []byte, identifier, sequence uint16) error {
	if len(reply) < 8 { // 8 bytes for ICMPv6 header
		return fmt.Errorf("reply too short, not a valid ICMPv6 echo reply")
	}

	if reply[0] != 129 { // ICMPv6 Echo Reply type
		return fmt.Errorf("unexpected ICMPv6 reply type: %d", reply[0])
	}

	recvIdentifier := binary.BigEndian.Uint16(reply[4:6])
	recvSequence := binary.BigEndian.Uint16(reply[6:8])

	if recvIdentifier != identifier || recvSequence != sequence {
		return fmt.Errorf("identifier or sequence number mismatch: got id=%d seq=%d, expected id=%d seq=%d",
			recvIdentifier, recvSequence, identifier, sequence)
	}

	return nil
}

// Network returns the network type for IPv6.
func (p *ICMPv6) Network() string {
	return "ip6:ipv6-icmp"
}
