package checker

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/containeroo/portpatrol/internal/testutils"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func TestNewICMPChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid IPv4 Configuration", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			return "2s" // valid timeout
		}

		checker, err := NewICMPChecker("TestIPv4", "icmp://google.com", 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if checker == nil {
			t.Fatal("expected valid checker, got nil")
		}

		expected := "TestIPv4"
		if expected != checker.String() {
			t.Fatalf("expected %q, got %q", expected, checker)
		}

		icmpChecker := checker.(*ICMPChecker)
		if icmpChecker.ReadTimeout != 2*time.Second {
			t.Errorf("expected timeout of 2s, got %v", icmpChecker.ReadTimeout)
		}
	})

	t.Run("Valid IPv6 Configuration", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			return "" // will fall back to default
		}

		checker, err := NewICMPChecker("TestIPv6", "icmp://0:0:0:0:0:0:0:0", 1*time.Second, mockEnv)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if checker == nil {
			t.Fatal("expected valid checker, got nil")
		}

		icmpChecker := checker.(*ICMPChecker)
		if icmpChecker.ReadTimeout != time.Second {
			t.Errorf("expected default timeout of 1s, got %v", icmpChecker.ReadTimeout)
		}
	})

	t.Run("Invalid IP Address", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			return ""
		}

		_, err := NewICMPChecker("TestInvalidIP", "icmp://0.260.0.0", 1*time.Second, mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to create ICMP protocol: invalid or unresolvable address: 0.260.0.0"
		if err.Error() != expected {
			t.Errorf("expected error %q, got %q", expected, err.Error())
		}
	})

	t.Run("Invalid Read Timeout", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			return "invalid"
		}

		_, err := NewICMPChecker("TestInvalidTimeout", "icmp://localhost", 1*time.Second, mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: invalid", envICMPReadTimeout)
		if err.Error() != expected {
			t.Errorf("expected error %q, got %q", expected, err.Error())
		}
	})

	t.Run("Negative Read Timeout", func(t *testing.T) {
		t.Parallel()

		mockEnv := func(key string) string {
			return "-1s"
		}

		_, err := NewICMPChecker("TestNegativeTimeout", "icmp://localhost", 1*time.Second, mockEnv)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := fmt.Sprintf("invalid %s value: -1s", envICMPReadTimeout)
		if err.Error() != expected {
			t.Errorf("expected error %q, got %q", expected, err.Error())
		}
	})
}

func TestICMPChecker(t *testing.T) {
	t.Parallel()

	t.Run("Successful ICMP Check", func(t *testing.T) {
		t.Parallel()

		expectedIdentifier := uint16(1234)
		expectedSequence := uint16(1)

		mockPacketConn := &testutils.MockPacketConn{
			WriteToFunc: func(b []byte, addr net.Addr) (int, error) {
				return len(b), nil
			},
			ReadFromFunc: func(b []byte) (int, net.Addr, error) {
				msg := icmp.Message{
					Type: ipv4.ICMPTypeEchoReply,
					Code: 0,
					Body: &icmp.Echo{
						ID:   int(expectedIdentifier),
						Seq:  int(expectedSequence),
						Data: []byte("HELLO-R-U-THERE"),
					},
				}
				msgBytes, _ := msg.Marshal(nil)
				copy(b, msgBytes)
				return len(msgBytes), &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}, nil
			},
			SetReadDeadlineFunc: func(t time.Time) error {
				return nil
			},
			CloseFunc: func() error {
				return nil
			},
			LocalAddrFunc: func() net.Addr {
				return &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}
			},
		}

		mockProtocol := &testutils.MockProtocol{
			MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
				body := &icmp.Echo{
					ID:   int(expectedIdentifier),
					Seq:  int(expectedSequence),
					Data: []byte("HELLO-R-U-THERE"),
				}
				msg := icmp.Message{
					Type: ipv4.ICMPTypeEcho,
					Code: 0,
					Body: body,
				}
				return msg.Marshal(nil)
			},
			ValidateReplyFunc: func(reply []byte, id, seq uint16) error {
				parsedMsg, err := icmp.ParseMessage(1, reply)
				if err != nil {
					return err
				}
				body, ok := parsedMsg.Body.(*icmp.Echo)
				if !ok || body.ID != int(expectedIdentifier) || body.Seq != int(expectedSequence) {
					return fmt.Errorf("identifier or sequence mismatch")
				}
				return nil
			},
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(network, address string) (net.PacketConn, error) {
				return mockPacketConn, nil
			},
		}

		checker := &ICMPChecker{
			Name:        "TestChecker",
			Address:     "127.0.0.1",
			Protocol:    mockProtocol,
			ReadTimeout: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Error Listening for ICMP Packets", func(t *testing.T) {
		t.Parallel()

		mockProtocol := &testutils.MockProtocol{
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(network, address string) (net.PacketConn, error) {
				return nil, fmt.Errorf("mock listen packet error")
			},
		}

		checker := &ICMPChecker{
			Name:        "TestCheckerListenError",
			Address:     "127.0.0.1",
			Protocol:    mockProtocol,
			ReadTimeout: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil || err.Error() != "failed to listen for ICMP packets: mock listen packet error" {
			t.Errorf("expected listen packet error, got %v", err)
		}
	})

	t.Run("Error Setting Read Deadline", func(t *testing.T) {
		t.Parallel()

		mockPacketConn := &testutils.MockPacketConn{
			SetReadDeadlineFunc: func(t time.Time) error {
				return fmt.Errorf("mock set read deadline error")
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockProtocol := &testutils.MockProtocol{
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(network, address string) (net.PacketConn, error) {
				return mockPacketConn, nil
			},
		}

		checker := &ICMPChecker{
			Name:        "TestCheckerDeadlineError",
			Address:     "127.0.0.1",
			Protocol:    mockProtocol,
			ReadTimeout: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil || err.Error() != "failed to set read deadline: mock set read deadline error" {
			t.Errorf("expected set read deadline error, got %v", err)
		}
	})

	t.Run("Error Creating ICMP Request", func(t *testing.T) {
		t.Parallel()

		mockPacketConn := &testutils.MockPacketConn{
			CloseFunc: func() error {
				return nil
			},
			SetReadDeadlineFunc: func(t time.Time) error {
				return nil
			},
		}

		mockProtocol := &testutils.MockProtocol{
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(network, address string) (net.PacketConn, error) {
				return mockPacketConn, nil
			},
			MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
				return nil, fmt.Errorf("mock make request error")
			},
		}

		checker := &ICMPChecker{
			Name:        "TestCheckerRequestError",
			Address:     "127.0.0.1",
			Protocol:    mockProtocol,
			ReadTimeout: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil || err.Error() != "failed to create ICMP request: mock make request error" {
			t.Errorf("expected make request error, got %v", err)
		}
	})

	t.Run("Error Resolving IP Address", func(t *testing.T) {
		t.Parallel()

		// You don't need to mock a PacketConn here because the error occurs before it is used.
		mockProtocol := &testutils.MockProtocol{
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(network, address string) (net.PacketConn, error) {
				// This won't be called due to the address resolution error
				return nil, nil
			},
			MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
				return []byte{}, nil
			},
		}

		checker := &ICMPChecker{
			Name:        "TestCheckerResolveError",
			Address:     "invalid-address",
			Protocol:    mockProtocol,
			ReadTimeout: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to resolve IP address: lookup invalid-address: no such host"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err)
		}
	})

	t.Run("Error Sending ICMP Request", func(t *testing.T) {
		t.Parallel()

		mockPacketConn := &testutils.MockPacketConn{
			WriteToFunc: func(b []byte, addr net.Addr) (int, error) {
				return 0, fmt.Errorf("mock write to error")
			},
			SetReadDeadlineFunc: func(t time.Time) error {
				// Ensure this function is properly mocked to avoid nil pointer dereference
				return nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockProtocol := &testutils.MockProtocol{
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(network, address string) (net.PacketConn, error) {
				return mockPacketConn, nil
			},
			MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
				return []byte{}, nil
			},
		}

		checker := &ICMPChecker{
			Name:        "TestCheckerWriteError",
			Address:     "127.0.0.1",
			Protocol:    mockProtocol,
			ReadTimeout: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil || err.Error() != "failed to send ICMP request to 127.0.0.1: mock write to error" {
			t.Errorf("expected write to error, got %v", err)
		}
	})

	t.Run("Error Reading ICMP Reply", func(t *testing.T) {
		t.Parallel()

		mockPacketConn := &testutils.MockPacketConn{
			WriteToFunc: func(b []byte, addr net.Addr) (int, error) {
				return len(b), nil
			},
			ReadFromFunc: func(b []byte) (int, net.Addr, error) {
				return 0, nil, fmt.Errorf("mock read from error")
			},
			SetReadDeadlineFunc: func(t time.Time) error {
				// Ensure this function is properly mocked.
				return nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockProtocol := &testutils.MockProtocol{
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(network, address string) (net.PacketConn, error) {
				return mockPacketConn, nil
			},
			MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
				return []byte{}, nil
			},
		}

		checker := &ICMPChecker{
			Name:        "TestCheckerReadError",
			Address:     "127.0.0.1",
			Protocol:    mockProtocol,
			ReadTimeout: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil || err.Error() != "failed to read ICMP reply from 127.0.0.1: mock read from error" {
			t.Errorf("expected read from error, got %v", err)
		}
	})

	t.Run("Error Validating ICMP Reply", func(t *testing.T) {
		t.Parallel()

		mockPacketConn := &testutils.MockPacketConn{
			WriteToFunc: func(b []byte, addr net.Addr) (int, error) {
				return len(b), nil
			},
			ReadFromFunc: func(b []byte) (int, net.Addr, error) {
				msg := icmp.Message{
					Type: ipv4.ICMPTypeEchoReply,
					Code: 0,
					Body: &icmp.Echo{
						ID:   int(1234), // incorrect ID to force validation error
						Seq:  int(1),
						Data: []byte("HELLO-R-U-THERE"),
					},
				}
				msgBytes, _ := msg.Marshal(nil)
				copy(b, msgBytes)
				return len(msgBytes), &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}, nil
			},
			SetReadDeadlineFunc: func(t time.Time) error {
				return nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockProtocol := &testutils.MockProtocol{
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(network, address string) (net.PacketConn, error) {
				return mockPacketConn, nil
			},
			MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
				return []byte{}, nil
			},
			ValidateReplyFunc: func(reply []byte, id, seq uint16) error {
				return fmt.Errorf("mock validation error")
			},
		}

		checker := &ICMPChecker{
			Name:        "TestCheckerValidationError",
			Address:     "127.0.0.1",
			Protocol:    mockProtocol,
			ReadTimeout: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil || err.Error() != "mock validation error" {
			t.Errorf("expected validation error, got %v", err)
		}
	})
}
