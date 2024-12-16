package checker

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/containeroo/portpatrol/internal/testutils"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// TestNewICMPCheckerValidIPv4 tests creating an ICMPChecker with a valid IPv4 address.
func TestNewICMPCheckerValidIPv4(t *testing.T) {
	t.Parallel()

	w := WithICMPReadTimeout(2 * time.Second)
	checker, err := newICMPChecker("ValidIPv4", "127.0.0.1", w)

	assert.NoError(t, err)
	assert.Equal(t, checker.GetName(), "ValidIPv4")
	assert.Equal(t, checker.GetAddress(), "127.0.0.1")
}

// TestNewICMPCheckerInvalidAddress tests creating an ICMPChecker with an invalid address.
func TestNewICMPCheckerInvalidAddress(t *testing.T) {
	t.Parallel()

	_, err := newICMPChecker("InvalidAddress", "invalid-address")
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "failed to create ICMP protocol: invalid or unresolvable address: invalid-address")
}

// TestICMPCheckerCheckSuccess tests successful ICMP checking.
func TestICMPCheckerCheckSuccess(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			msg := icmp.Message{
				Type: ipv4.ICMPTypeEcho,
				Code: 0,
				Body: &icmp.Echo{
					ID:   int(id),
					Seq:  int(seq),
					Data: []byte("HELLO-R-U-THERE"),
				},
			}
			return msg.Marshal(nil)
		},
		ValidateReplyFunc: func(reply []byte, id, seq uint16) error {
			return nil
		},
		NetworkFunc: func() string {
			return "ip4:icmp"
		},
		ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
			return &testutils.MockPacketConn{}, nil
		},
	}

	checker := &ICMPChecker{
		name:        "SuccessChecker",
		address:     "127.0.0.1",
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	assert.NoError(t, err)
}

// TestICMPCheckerCheckResolveError tests ICMP checking with an address resolution failure.
func TestICMPCheckerCheckResolveError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		NetworkFunc: func() string {
			return "ip4:icmp"
		},
	}

	checker := &ICMPChecker{
		name:        "ResolveErrorChecker",
		address:     "invalid-host",
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to resolve IP address 'invalid-host': lookup invalid-host: no such host")
}

// TestICMPCheckerCheckWriteError tests ICMP checking with a failure to write to the connection.
func TestICMPCheckerCheckWriteError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			return []byte{}, nil
		},
		NetworkFunc: func() string {
			return "ip4:icmp"
		},
		ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
			return &testutils.MockPacketConn{
				WriteToFunc: func(b []byte, addr net.Addr) (int, error) {
					return 0, fmt.Errorf("mock write error")
				},
			}, nil
		},
	}

	checker := &ICMPChecker{
		name:        "WriteErrorChecker",
		address:     "127.0.0.1",
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to send ICMP request: mock write error")
}

// TestICMPCheckerCheckReadError tests ICMP checking with a failure to read from the connection.
func TestICMPCheckerCheckReadError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			return []byte{}, nil
		},
		NetworkFunc: func() string {
			return "ip4:icmp"
		},
		ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
			return &testutils.MockPacketConn{
				ReadFromFunc: func(b []byte) (int, net.Addr, error) {
					return 0, nil, fmt.Errorf("mock read error")
				},
			}, nil
		},
	}

	checker := &ICMPChecker{
		name:        "ReadErrorChecker",
		address:     "127.0.0.1",
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to read ICMP reply: mock read error")
}

// TestICMPCheckerCheckValidationError tests ICMP checking with a failure to validate the ICMP reply.
func TestICMPCheckerCheckValidationError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			return []byte{}, nil
		},
		ValidateReplyFunc: func(reply []byte, id, seq uint16) error {
			return fmt.Errorf("mock validation error")
		},
		NetworkFunc: func() string {
			return "ip4:icmp"
		},
		ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
			return &testutils.MockPacketConn{}, nil
		},
	}

	checker := &ICMPChecker{
		name:        "ValidationErrorChecker",
		address:     "127.0.0.1",
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to validate ICMP reply: mock validation error")
}
