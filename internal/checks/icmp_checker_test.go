package checks

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
			return "2s"
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
			ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
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
			ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
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
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to listen for ICMP packets: mock listen packet error"
		if err.Error() != expected {
			t.Errorf("expected listen packet error, got %v", err)
		}
	})

	t.Run("Error Setting Write Deadline", func(t *testing.T) {
		t.Parallel()

		mockPacketConn := &testutils.MockPacketConn{
			SetWriteDeadlineFunc: func(t time.Time) error {
				return fmt.Errorf("mock set write deadline error")
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockProtocol := &testutils.MockProtocol{
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
				return mockPacketConn, nil
			},
		}

		checker := &ICMPChecker{
			Name:        "TestCheckerWriteDeadlineError",
			Address:     "127.0.0.1",
			Protocol:    mockProtocol,
			ReadTimeout: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to set write deadline: mock set write deadline error"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
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
			ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
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
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to set read deadline: mock set read deadline error"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
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
			ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
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
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "mock make request error"
		if err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	})

	t.Run("Error Resolving IP Address", func(t *testing.T) {
		t.Parallel()

		// You don't need to mock a PacketConn here because the error occurs before it is used.
		mockProtocol := &testutils.MockProtocol{
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
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
			ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
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
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to send ICMP request to 127.0.0.1: mock write to error"
		if err.Error() != expected {
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
			ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
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
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to read ICMP reply from 127.0.0.1: mock read from error"
		if err.Error() != expected {
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
			ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
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
		if err == nil {
			t.Error("expected an error, got none")
		}

		expected := "mock validation error"
		if err.Error() != expected {
			t.Errorf("expected validation error, got %v", err)
		}
	})
}

func TestMakeICMPRequest(t *testing.T) {
	t.Parallel()

	c := &ICMPChecker{
		Protocol: &testutils.MockProtocol{
			MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
				body := &icmp.Echo{
					ID:   int(id),
					Seq:  int(seq),
					Data: []byte("HELLO-R-U-THERE"),
				}
				msg := icmp.Message{
					Type: ipv4.ICMPTypeEcho,
					Code: 0,
					Body: body,
				}
				msgBytes, err := msg.Marshal(nil)
				if err != nil {
					return nil, err
				}

				t.Logf("Generated ICMP Request: %v", msgBytes)

				return msgBytes, nil
			},
		},
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		msg, err := c.Protocol.MakeRequest(1234, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msg) == 0 {
			t.Fatalf("expected non-empty message, got empty")
		}

		t.Logf("Generated message in test: %v", msg)
	})
}

func TestWriteICMPRequest(t *testing.T) {
	t.Parallel()

	c := &ICMPChecker{
		Protocol: &testutils.MockProtocol{},
	}

	mockConn := &testutils.MockPacketConn{
		WriteToFunc: func(b []byte, addr net.Addr) (int, error) {
			return len(b), nil
		},
		SetWriteDeadlineFunc: func(t time.Time) error {
			return nil
		},
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := c.writeICMPRequest(ctx, mockConn, []byte{0x01, 0x02, 0x03}, &net.IPAddr{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("Context Canceled", func(t *testing.T) {
		t.Parallel()

		mockConn := &testutils.MockPacketConn{
			ReadFromFunc: func(b []byte) (int, net.Addr, error) {
				// Simulate a slow response
				time.Sleep(3 * time.Second)
				copy(b, []byte("valid"))
				return 5, &net.IPAddr{}, nil
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		cancel()

		err := c.writeICMPRequest(ctx, mockConn, []byte{0x01, 0x02, 0x03}, &net.IPAddr{})
		if err == nil {
			t.Fatalf("expected context canceled error, got nil")
		}
	})

	t.Run("Write Deadline Error", func(t *testing.T) {
		t.Parallel()

		mockConn.SetWriteDeadlineFunc = func(t time.Time) error {
			return fmt.Errorf("mock set write deadline error")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := c.writeICMPRequest(ctx, mockConn, []byte{0x01, 0x02, 0x03}, &net.IPAddr{})
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to set write deadline: mock set write deadline error"
		if err.Error() != expected {
			t.Fatalf("expected write deadline error, got %v", err)
		}
	})
}

func TestReadICMPReply(t *testing.T) {
	t.Parallel()

	c := &ICMPChecker{
		Protocol: &testutils.MockProtocol{},
	}

	mockConn := &testutils.MockPacketConn{
		ReadFromFunc: func(b []byte) (int, net.Addr, error) {
			copy(b, []byte("valid"))
			return 5, &net.IPAddr{}, nil // Return the exact number of bytes written.
		},
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		reply, err := c.readICMPReply(ctx, mockConn)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if string(reply) != "valid" {
			t.Fatalf("expected 'valid', got %v", string(reply))
		}
	})

	t.Run("Context Canceled", func(t *testing.T) {
		t.Parallel()

		mockConn := &testutils.MockPacketConn{
			ReadFromFunc: func(b []byte) (int, net.Addr, error) {
				// Simulate a slow response
				time.Sleep(3 * time.Second)
				copy(b, []byte("valid"))
				return 5, &net.IPAddr{}, nil
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := c.readICMPReply(ctx, mockConn)
		if err == nil {
			t.Fatalf("expected context canceled error, got nil")
		}
	})

	t.Run("Read Deadline Error", func(t *testing.T) {
		t.Parallel()

		mockConn.SetReadDeadlineFunc = func(t time.Time) error {
			return fmt.Errorf("mock set read deadline error")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := c.readICMPReply(ctx, mockConn)
		if err == nil {
			t.Fatal("expected an error, got none")
		}

		expected := "failed to set read deadline: mock set read deadline error"
		if err.Error() != expected {
			t.Fatalf("expected write deadline error, got %v", err)
		}
	})
}

func TestValidateICMPReply(t *testing.T) {
	t.Parallel()

	c := &ICMPChecker{
		Protocol: &testutils.MockProtocol{
			ValidateReplyFunc: func(reply []byte, identifier, sequence uint16) error {
				if string(reply) == "valid" {
					return nil
				}
				return fmt.Errorf("identifier or sequence mismatch")
			},
		},
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		err := c.validateICMPReply(ctx, []byte("valid"), 1234, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("Validation Failure", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		err := c.validateICMPReply(ctx, []byte("invalid"), 1234, 1)
		if err == nil {
			t.Fatalf("expected validation error, got nil")
		}

		expectedErr := "identifier or sequence mismatch"
		if err.Error() != expectedErr {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("Context Canceled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := c.validateICMPReply(ctx, []byte("valid"), 1234, 1)
		if err == nil {
			t.Fatalf("expected context canceled error, got nil")
		}
	})
}
