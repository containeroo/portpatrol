package checker

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestICMPChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid ICMP config", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{
				envICMPReadTimeout: "1s",
			}
			return env[key]
		}

		checker, err := NewICMPChecker("example", "127.0.0.1", 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		expected := "example"
		if fmt.Sprintf("%v", checker) != expected {
			t.Errorf("expected Name to be '%s', got %v", expected, checker)
		}

		c := checker.(*ICMPChecker) // Type assertion to *ICMPChecker

		expected = "example"
		if c.Name != expected {
			t.Errorf("expected Name to be '%s', got %v", expected, c.Name)
		}

		expected = "127.0.0.1"
		if c.Address != expected {
			t.Errorf("expected Address to be '%s', got %v", expected, c.Address)
		}

		expectedReadTimeout := 1 * time.Second
		if c.ReadTimeout != expectedReadTimeout {
			t.Errorf("expected Method to be '%s', got %v", expected, c.ReadTimeout)
		}
	})

	t.Run("Valid ICMP check", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			env := map[string]string{}
			return env[key]
		}

		mockDialer := &mockDialer{
			dialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return &mockConn{}, nil // Simulate a successful ICMP connection
			},
		}

		checker, err := NewICMPChecker("example", "127.0.0.1", 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("failed to create HTTPChecker: %q", err)
		}

		// cast the checker to ICMPChecker to update the dialer
		c := checker.(*ICMPChecker)
		c.dialer = mockDialer

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		// Perform the check
		err = c.Check(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}
	})

	t.Run("Invalid ICMP check (malformed read timeout)", func(t *testing.T) {
		mockEnv := func(key string) string {
			env := map[string]string{
				envICMPReadTimeout: "invalid",
			}
			return env[key]
		}

		_, err := NewICMPChecker("example", "127.0.0.1", 1*time.Second, mockEnv)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: time: invalid duration \"invalid\"", envICMPReadTimeout)
		if err.Error() != expected {
			t.Fatalf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Failed ICMP check - connection refused", func(t *testing.T) {
		t.Parallel()

		mockDialer := &mockDialer{
			dialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				// Simulate a connection failure
				return nil, errors.New("connection refused")
			},
		}

		checker := &ICMPChecker{
			Name:        "example",
			Address:     "127.0.0.1",
			dialer:      mockDialer,
			ReadTimeout: 1 * time.Second,
		}

		err := checker.Check(context.Background())
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to dial ICMP address 127.0.0.1: connection refused"
		if err.Error() != expected {
			t.Errorf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Failed ICMP check - invalid response", func(t *testing.T) {
		t.Parallel()

		mockDialer := &mockDialer{
			dialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return &mockConn{reply: []byte{1, 2, 3, 4}}, nil // Simulate an invalid ICMP response
			},
		}

		checker := &ICMPChecker{
			Name:        "example",
			Address:     "127.0.0.1",
			dialer:      mockDialer,
			ReadTimeout: 1 * time.Second,
		}

		err := checker.Check(context.Background())
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "reply too short, not a valid ICMP echo reply"
		if err.Error() != expected {
			t.Errorf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Failed ICMP check - deadline error", func(t *testing.T) {
		t.Parallel()

		mockDialer := &mockDialer{
			dialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return &mockConnWithDeadlineError{}, nil
			},
		}

		checker := &ICMPChecker{
			Name:        "example",
			Address:     "127.0.0.1",
			dialer:      mockDialer,
			ReadTimeout: 1 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to set read deadline: mock deadline error"
		if err.Error() != expected {
			t.Errorf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Failed ICMP check - write error", func(t *testing.T) {
		t.Parallel()

		mockDialer := &mockDialer{
			dialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return &mockConnWithWriteError{}, nil
			},
		}

		checker := &ICMPChecker{
			Name:        "example",
			Address:     "127.0.0.1",
			dialer:      mockDialer,
			ReadTimeout: 1 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to send ICMP request to 127.0.0.1: mock write error"
		if err.Error() != expected {
			t.Errorf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Failed ICMP check - read error", func(t *testing.T) {
		t.Parallel()

		mockDialer := &mockDialer{
			dialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return &mockConnWithReadError{}, nil
			},
		}

		checker := &ICMPChecker{
			Name:        "example",
			Address:     "127.0.0.1",
			dialer:      mockDialer,
			ReadTimeout: 1 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to read ICMP reply from 127.0.0.1: mock read error"
		if err.Error() != expected {
			t.Errorf("expected error containing %q, got %q", expected, err)
		}
	})

	t.Run("Failed ICMP check - context canceled", func(t *testing.T) {
		t.Parallel()

		mockDialer := &mockDialer{
			dialContextFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return &mockConnWithDelay{delay: 2 * time.Second}, nil
			},
		}

		checker := &ICMPChecker{
			Name:        "example",
			Address:     "127.0.0.1",
			dialer:      mockDialer,
			ReadTimeout: 5 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "context cancelled while waiting for ICMP reply from 127.0.0.1: context deadline exceeded"
		if err.Error() != expected {
			t.Errorf("expected error containing %q, got %q", expected, err)
		}
	})
}

func TestMakeICMPEchoRequest(t *testing.T) {
	t.Parallel()

	t.Run("Valid ICMP Echo Request", func(t *testing.T) {
		t.Parallel()

		identifier := uint16(12345)
		sequence := uint16(1)

		// Generate the ICMP echo request packet
		packet := makeICMPEchoRequest(identifier, sequence)

		if len(packet) != 8 {
			t.Fatalf("expected packet length 8, got %d", len(packet))
		}

		// Verify the ICMP type field
		if packet[0] != 8 {
			t.Errorf("expected ICMP type 8 (Echo Request), got %d", packet[0])
		}

		// Verify the identifier
		id := binary.BigEndian.Uint16(packet[4:6])
		if id != identifier {
			t.Errorf("expected identifier %d, got %d", identifier, id)
		}

		// Verify the sequence number
		seq := binary.BigEndian.Uint16(packet[6:8])
		if seq != sequence {
			t.Errorf("expected sequence number %d, got %d", sequence, seq)
		}

		// Extract the checksum from the packet
		actualChecksum := binary.BigEndian.Uint16(packet[2:4])

		// Temporarily set the checksum to 0 to recalculate it
		binary.BigEndian.PutUint16(packet[2:], 0)
		expectedChecksum := calculateChecksum(packet)

		// Verify the checksum
		if actualChecksum != expectedChecksum {
			t.Errorf("expected checksum %d, got %d", expectedChecksum, actualChecksum)
		}
	})
}

func TestValidateICMPEchoReply(t *testing.T) {
	t.Parallel()

	t.Run("Valid ICMP Echo Reply", func(t *testing.T) {
		t.Parallel()

		identifier := uint16(12345)
		sequence := uint16(1)

		// Create a valid ICMP Echo Reply
		reply := make([]byte, 28)                          // 20 bytes for IP header, 8 bytes for ICMP header
		reply[20] = 0                                      // ICMP Echo Reply type
		binary.BigEndian.PutUint16(reply[24:], identifier) // Identifier
		binary.BigEndian.PutUint16(reply[26:], sequence)   // Sequence number

		// Test valid reply
		err := validateICMPEchoReply(reply, identifier, sequence)
		if err != nil {
			t.Fatalf("expected no error, got %q", err)
		}

		// Test with wrong identifier
		err = validateICMPEchoReply(reply, identifier+1, sequence)
		if err == nil || err.Error() != "identifier or sequence number mismatch: got id=12345 seq=1, expected id=12346 seq=1" {
			t.Fatalf("expected identifier mismatch error, got %q", err)
		}

		// Test with wrong sequence number
		err = validateICMPEchoReply(reply, identifier, sequence+1)
		if err == nil || err.Error() != "identifier or sequence number mismatch: got id=12345 seq=1, expected id=12345 seq=2" {
			t.Fatalf("expected sequence number mismatch error, got %q", err)
		}

		// Test with wrong type
		reply[20] = 1 // Set to an invalid ICMP type
		err = validateICMPEchoReply(reply, identifier, sequence)
		if err == nil || err.Error() != "unexpected ICMP reply type: 1" {
			t.Fatalf("expected ICMP type mismatch error, got %q", err)
		}

		// Test with short reply
		shortReply := reply[:26] // Make it shorter than expected
		err = validateICMPEchoReply(shortReply, identifier, sequence)
		if err == nil || err.Error() != "reply too short, not a valid ICMP echo reply" {
			t.Fatalf("expected short reply error, got %q", err)
		}
	})
}

func TestCalculateChecksum(t *testing.T) {
	t.Parallel()

	t.Run("Even number of bytes", func(t *testing.T) {
		t.Parallel()

		data := []byte{
			8, 0, 0, 0, // ICMP type and code, with zeroed checksum
			0, 13, // Identifier
			0, 37, // Sequence number
		}
		expectedChecksum := uint16(0xf7cd) // This matches the checksum output you received

		checksum := calculateChecksum(data)
		if checksum != expectedChecksum {
			t.Errorf("expected checksum 0x%x, got 0x%x", expectedChecksum, checksum)
		}
	})

	t.Run("Odd number of bytes", func(t *testing.T) {
		t.Parallel()

		data := []byte{
			8, 0, 0, 0, // ICMP type and code, with zeroed checksum
			0, 13, // Identifier
			0, // Odd byte padding
		}
		expectedChecksum := uint16(0xf7f2) // Corrected expected checksum

		checksum := calculateChecksum(data)
		if checksum != expectedChecksum {
			t.Errorf("expected checksum 0x%x, got 0x%x", expectedChecksum, checksum)
		}
	})

	t.Run("Empty data", func(t *testing.T) {
		t.Parallel()

		data := []byte{}
		expectedChecksum := uint16(0xffff) // Checksum for empty data should be 0xffff

		checksum := calculateChecksum(data)
		if checksum != expectedChecksum {
			t.Errorf("expected checksum 0x%x, got 0x%x", expectedChecksum, checksum)
		}
	})

	t.Run("Single byte data", func(t *testing.T) {
		t.Parallel()

		data := []byte{
			8, // Single byte, checksum calculation should treat this as 8 shifted left by 8 bits (0x0800)
		}
		expectedChecksum := uint16(0xf7ff) // Precomputed expected checksum

		checksum := calculateChecksum(data)
		if checksum != expectedChecksum {
			t.Errorf("expected checksum 0x%x, got 0x%x", expectedChecksum, checksum)
		}
	})

	t.Run("Maximum short packet", func(t *testing.T) {
		t.Parallel()

		data := []byte{
			8, 0, 0, 0, 0, 13, 0, 37, 255, 255, // Longer packet
		}
		expectedChecksum := uint16(0xf7cd) // Corrected expected checksum

		checksum := calculateChecksum(data)
		if checksum != expectedChecksum {
			t.Errorf("expected checksum 0x%x, got 0x%x", expectedChecksum, checksum)
		}
	})

	t.Run("ICMP packet with checksum included", func(t *testing.T) {
		t.Parallel()

		data := []byte{
			8, 0, 199, 197, 48, 57, 0, 1, // ICMP type, code, and checksum included
		}
		expectedChecksum := uint16(0x0000) // When recalculated over full packet with checksum included, should be 0

		checksum := calculateChecksum(data)
		if checksum != expectedChecksum {
			t.Errorf("expected checksum 0x%x, got 0x%x", expectedChecksum, checksum)
		}
	})
}

// Mocking DialContext to simulate ICMP behavior
type mockDialer struct {
	dialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)
}

// Implement the DialContext method of the Dialer interface
func (m *mockDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return m.dialContextFunc(ctx, network, address)
}

type mockConn struct {
	reply        []byte
	lastSequence uint16
	lastID       uint16
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if len(m.reply) == 0 {
		// Simulate a valid ICMP Echo Reply with the correct sequence number and identifier
		mockReply := make([]byte, 28)                              // 20 bytes for IP header, 8 bytes for ICMP header
		mockReply[20] = 0                                          // ICMP Echo Reply type
		binary.BigEndian.PutUint16(mockReply[24:], m.lastID)       // Identifier from the last Write call
		binary.BigEndian.PutUint16(mockReply[26:], m.lastSequence) // Sequence number from the last Write call
		copy(b, mockReply)
		return len(mockReply), nil
	}
	copy(b, m.reply)
	return len(m.reply), nil
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	// Extract identifier and sequence number from the request
	m.lastID = binary.BigEndian.Uint16(b[4:6])
	m.lastSequence = binary.BigEndian.Uint16(b[6:8])
	return len(b), nil
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.IPAddr{}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.IPAddr{}
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// Mock connection that returns an error when SetReadDeadline is called
type mockConnWithDeadlineError struct {
	mockConn
}

func (m *mockConnWithDeadlineError) SetReadDeadline(t time.Time) error {
	return errors.New("mock deadline error")
}

// Mock connection that returns an error when Write is called
type mockConnWithWriteError struct {
	mockConn
}

func (m *mockConnWithWriteError) Write(b []byte) (n int, err error) {
	return 0, errors.New("mock write error")
}

// Mock connection that returns an error when Read is called
type mockConnWithReadError struct {
	mockConn
}

func (m *mockConnWithReadError) Read(b []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}

// Mock connection that simulates a delayed response, causing context cancellation
type mockConnWithDelay struct {
	mockConn
	delay time.Duration
}

func (m *mockConnWithDelay) Read(b []byte) (n int, err error) {
	time.Sleep(m.delay) // Simulate a delay that should trigger context cancellation
	return 0, nil       // Return zero bytes read with no error, the context cancellation should trigger the error
}
