package checker

import (
	"encoding/binary"
	"testing"
)

func TestNewProtocol(t *testing.T) {
	t.Parallel()

	t.Run("Valid IPv4 address", func(t *testing.T) {
		t.Parallel()

		address := "192.168.1.1"
		protocol, err := NewProtocol(address)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		p := protocol.(*ICMPv4)

		expected := "ip4:icmp"
		if p.Network() != expected {
			t.Errorf("expected network %s, got %s", expected, p.Network())
		}
	})

	t.Run("Valid IPv6 address", func(t *testing.T) {
		t.Parallel()

		address := "2001:db8::1"
		protocol, err := NewProtocol(address)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		p := protocol.(*ICMPv6)

		expected := "ip6:ipv6-icmp"
		if p.Network() != expected {
			t.Fatalf("expected network %s, got %s", expected, p.Network())
		}
	})

	t.Run("Invalid IP address", func(t *testing.T) {
		t.Parallel()

		address := "invalid-ip"
		_, err := NewProtocol(address)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		expectedError := "invalid IP address: invalid-ip"
		if err.Error() != expectedError {
			t.Errorf("expected error %q, got %q", expectedError, err.Error())
		}
	})
}

func TestICMPv4_MakeRequest(t *testing.T) {
	t.Parallel()

	identifier := uint16(12345)
	sequence := uint16(1)

	p := &ICMPv4{}
	packet := p.MakeRequest(identifier, sequence)

	if len(packet) != 8 {
		t.Fatalf("expected packet length 8, got %d", len(packet))
	}

	if packet[0] != 8 {
		t.Errorf("expected ICMP type 8 (Echo Request), got %d", packet[0])
	}

	id := binary.BigEndian.Uint16(packet[4:6])
	if id != identifier {
		t.Errorf("expected identifier %d, got %d", identifier, id)
	}

	seq := binary.BigEndian.Uint16(packet[6:8])
	if seq != sequence {
		t.Errorf("expected sequence number %d, got %d", sequence, seq)
	}

	actualChecksum := binary.BigEndian.Uint16(packet[2:4])
	binary.BigEndian.PutUint16(packet[2:], 0)
	expectedChecksum := calculateChecksum(packet)

	if actualChecksum != expectedChecksum {
		t.Errorf("expected checksum %d, got %d", expectedChecksum, actualChecksum)
	}
}

func TestICMPv4_ValidateReply(t *testing.T) {
	t.Parallel()

	t.Run("Valid ICMPv4 Echo Reply", func(t *testing.T) {
		t.Parallel()

		p := &ICMPv4{}
		identifier := uint16(12345)
		sequence := uint16(1)

		reply := make([]byte, 28)
		reply[20] = 0 // ICMP Echo Reply type

		binary.BigEndian.PutUint16(reply[24:], identifier)
		binary.BigEndian.PutUint16(reply[26:], sequence)

		err := p.ValidateReply(reply, identifier, sequence)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("Invalid ICMPv4 Echo Reply (invalid type)", func(t *testing.T) {
		t.Parallel()

		p := &ICMPv4{}
		identifier := uint16(12345)
		sequence := uint16(1)

		reply := make([]byte, 28)
		reply[20] = 1 // Invalid ICMP type

		binary.BigEndian.PutUint16(reply[24:], identifier)
		binary.BigEndian.PutUint16(reply[26:], sequence)

		err := p.ValidateReply(reply, identifier, sequence)
		if err == nil || err.Error() != "unexpected ICMP reply type: 1" {
			t.Errorf("expected error 'unexpected ICMP reply type: 1', got %v", err)
		}
	})

	t.Run("Invalid ICMPv4 Echo Reply (identifier mismatch)", func(t *testing.T) {
		t.Parallel()

		p := &ICMPv4{}
		identifier := uint16(12345)
		sequence := uint16(1)

		reply := make([]byte, 28)
		reply[20] = 0 // ICMP Echo Reply type

		binary.BigEndian.PutUint16(reply[24:], identifier) // Actual identifier
		binary.BigEndian.PutUint16(reply[26:], sequence)   // Sequence number

		err := p.ValidateReply(reply, identifier+1, sequence) // Expected different identifier
		if err == nil || err.Error() != "identifier or sequence number mismatch: got id=12345 seq=1, expected id=12346 seq=1" {
			t.Errorf("expected identifier mismatch error, got %v", err)
		}
	})

	t.Run("Invalid ICMPv4 Echo Reply (sequence mismatch)", func(t *testing.T) {
		t.Parallel()

		p := &ICMPv4{}
		identifier := uint16(12345)
		sequence := uint16(1)

		reply := make([]byte, 28)
		reply[20] = 0 // ICMP Echo Reply type

		binary.BigEndian.PutUint16(reply[24:], identifier) // Identifier
		binary.BigEndian.PutUint16(reply[26:], sequence)   // Actual sequence

		err := p.ValidateReply(reply, identifier, sequence+1) // Expected different sequence
		if err == nil || err.Error() != "identifier or sequence number mismatch: got id=12345 seq=1, expected id=12345 seq=2" {
			t.Errorf("expected sequence mismatch error, got %v", err)
		}
	})

	t.Run("Invalid ICMPv4 Echo Reply (to short)", func(t *testing.T) {
		t.Parallel()

		p := &ICMPv4{}
		identifier := uint16(12345)
		sequence := uint16(1)

		// Create a reply that is too short (less than 28 bytes, which is the minimum for ICMPv4: 20 bytes IP header + 8 bytes ICMP header)
		shortReply := make([]byte, 24) // 24 bytes, which is shorter than the required 28 bytes

		err := p.ValidateReply(shortReply, identifier, sequence)
		if err == nil || err.Error() != "reply too short, not a valid ICMP echo reply" {
			t.Errorf("expected error 'reply too short, not a valid ICMP echo reply', got %v", err)
		}
	})
}

func TestICMPv6_MakeRequest(t *testing.T) {
	t.Parallel()

	identifier := uint16(12345)
	sequence := uint16(1)

	p := &ICMPv6{}
	packet := p.MakeRequest(identifier, sequence)

	if len(packet) != 8 {
		t.Fatalf("expected packet length 8, got %d", len(packet))
	}

	if packet[0] != 128 {
		t.Errorf("expected ICMPv6 type 128 (Echo Request), got %d", packet[0])
	}

	id := binary.BigEndian.Uint16(packet[4:6])
	if id != identifier {
		t.Errorf("expected identifier %d, got %d", identifier, id)
	}

	seq := binary.BigEndian.Uint16(packet[6:8])
	if seq != sequence {
		t.Errorf("expected sequence number %d, got %d", sequence, seq)
	}

	actualChecksum := binary.BigEndian.Uint16(packet[2:4])
	binary.BigEndian.PutUint16(packet[2:], 0)
	expectedChecksum := calculateChecksum(packet)

	if actualChecksum != expectedChecksum {
		t.Errorf("expected checksum %d, got %d", expectedChecksum, actualChecksum)
	}
}

func TestICMPv6_ValidateReply(t *testing.T) {
	t.Parallel()

	t.Run("Valid ICMPv6 Echo Reply", func(t *testing.T) {
		t.Parallel()

		p := &ICMPv6{}
		identifier := uint16(12345)
		sequence := uint16(1)

		reply := make([]byte, 8)
		reply[0] = 129 // ICMPv6 Echo Reply type
		binary.BigEndian.PutUint16(reply[4:], identifier)
		binary.BigEndian.PutUint16(reply[6:], sequence)

		err := p.ValidateReply(reply, identifier, sequence)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("Invalid ICMPv6 Echo Reply (invalid type)", func(t *testing.T) {
		t.Parallel()

		p := &ICMPv6{}
		identifier := uint16(12345)
		sequence := uint16(1)

		reply := make([]byte, 8)
		reply[0] = 1 // Invalid ICMPv6 type

		binary.BigEndian.PutUint16(reply[4:], identifier)
		binary.BigEndian.PutUint16(reply[6:], sequence)

		err := p.ValidateReply(reply, identifier, sequence)
		if err == nil || err.Error() != "unexpected ICMPv6 reply type: 1" {
			t.Errorf("expected error 'unexpected ICMPv6 reply type: 1', got %v", err)
		}
	})

	t.Run("Invalid ICMPv6 Echo Reply (identifier mismatch)", func(t *testing.T) {
		t.Parallel()

		p := &ICMPv6{}
		identifier := uint16(12345)
		sequence := uint16(1)

		reply := make([]byte, 8)
		reply[0] = 129 // ICMPv6 Echo Reply type

		binary.BigEndian.PutUint16(reply[4:], identifier) // Actual identifier
		binary.BigEndian.PutUint16(reply[6:], sequence)   // Sequence number

		err := p.ValidateReply(reply, identifier+1, sequence) // Expected different identifier
		if err == nil || err.Error() != "identifier or sequence number mismatch: got id=12345 seq=1, expected id=12346 seq=1" {
			t.Errorf("expected identifier mismatch error, got %v", err)
		}
	})

	t.Run("Invalid ICMPv6 Echo Reply (sequence mismatch)", func(t *testing.T) {
		t.Parallel()

		p := &ICMPv6{}
		identifier := uint16(12345)
		sequence := uint16(1)

		reply := make([]byte, 8)
		reply[0] = 129 // ICMPv6 Echo Reply type

		binary.BigEndian.PutUint16(reply[4:], identifier) // Identifier
		binary.BigEndian.PutUint16(reply[6:], sequence)   // Actual sequence

		err := p.ValidateReply(reply, identifier, sequence+1) // Expected different sequence
		if err == nil || err.Error() != "identifier or sequence number mismatch: got id=12345 seq=1, expected id=12345 seq=2" {
			t.Errorf("expected sequence mismatch error, got %v", err)
		}
	})

	t.Run("Invalid ICMPv6 Echo Reply (to short)", func(t *testing.T) {
		t.Parallel()

		p := &ICMPv6{}
		identifier := uint16(12345)
		sequence := uint16(1)

		// Create a reply that is too short (less than 8 bytes)
		shortReply := make([]byte, 4) // 4 bytes, which is shorter than the 8-byte ICMPv6 header

		err := p.ValidateReply(shortReply, identifier, sequence)
		if err == nil || err.Error() != "reply too short, not a valid ICMPv6 echo reply" {
			t.Errorf("expected error 'reply too short, not a valid ICMPv6 echo reply', got %v", err)
		}
	})
}
