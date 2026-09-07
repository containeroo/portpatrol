package checker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/containeroo/never/internal/testutils"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// TestNewICMPCheckerValidIPv4 tests creating an ICMPChecker with a valid IPv4 address.
func TestNewICMPCheckerValidIPv4(t *testing.T) {
	t.Parallel()

	protocolConfig := DefaultICMPConfig()
	protocolConfig.ReadTimeout = 2 * time.Second
	protocolConfig.WriteTimeout = 2 * time.Second
	checker, err := NewICMPChecker("ValidIPv4", testutils.LocalhostIPv4, protocolConfig)

	require.NoError(t, err)
	assert.Equal(t, checker.Name(), "ValidIPv4")
	assert.Equal(t, checker.Address(), testutils.LocalhostIPv4)
}

// TestNewICMPCheckerInvalidAddress tests creating an ICMPChecker with an invalid address.
func TestNewICMPCheckerInvalidAddress(t *testing.T) {
	t.Parallel()

	_, err := NewICMPChecker("UnresolvedAddress", "not-yet-ready.invalid", DefaultICMPConfig())
	require.NoError(t, err)
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
			return icmpv4Network
		},
		ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
			return &testutils.MockPacketConn{}, nil
		},
	}

	checker := &ICMPChecker{
		name:        "SuccessChecker",
		address:     testutils.LocalhostIPv4,
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	require.NoError(t, err)
}

// TestICMPCheckerCheckResolveError tests ICMP checking with an address resolution failure.
func TestICMPCheckerCheckResolveError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		NetworkFunc: func() string {
			return icmpv4Network
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
	require.Error(t, err)

	// Use contains since the error message may vary on different platforms
	assert.Contains(t, err.Error(), "failed to resolve IP address 'invalid-host':")

	// Unwrap the underlying cause
	var dnsErr *net.DNSError
	if assert.ErrorAs(t, err, &dnsErr) {
		assert.True(t, dnsErr.IsNotFound || dnsErr.IsTemporary, "expected NXDOMAIN or temporary DNS failure, got: %+v", dnsErr)
	}
}

// TestICMPCheckerCheckWriteError tests ICMP checking with a failure to write to the connection.
func TestICMPCheckerCheckWriteError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			return []byte{}, nil
		},
		NetworkFunc: func() string {
			return icmpv4Network
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
		address:     testutils.LocalhostIPv4,
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	require.Error(t, err)
	assert.EqualError(t, err, "failed to send ICMP request: mock write error")
}

// TestICMPCheckerCheckListenPacketError verifies the expected behavior.
func TestICMPCheckerCheckListenPacketError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			return []byte{}, nil
		},
		NetworkFunc: func() string {
			return icmpv4Network
		},
		ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
			return nil, fmt.Errorf("mock listen packet error")
		},
	}

	checker := &ICMPChecker{
		name:         "ListenPacketErrorChecker",
		address:      testutils.LocalhostIPv4,
		protocol:     mockProtocol,
		readTimeout:  2 * time.Second,
		writeTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	require.Error(t, err)
	assert.EqualError(t, err, "failed to listen for ICMP packets: mock listen packet error")
}

// TestICMPCheckerCheckMakeRequestError verifies the expected behavior.
func TestICMPCheckerCheckMakeRequestError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			return []byte{}, errors.New("mock make request error")
		},
		ValidateReplyFunc: func(reply []byte, id, seq uint16) error {
			return fmt.Errorf("mock validation error")
		},
		NetworkFunc: func() string {
			return icmpv4Network
		},
		ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
			return &testutils.MockPacketConn{}, nil
		},
	}

	checker := &ICMPChecker{
		name:         "WriteDeadlineErrorChecker",
		address:      testutils.LocalhostIPv4,
		protocol:     mockProtocol,
		readTimeout:  2 * time.Second,
		writeTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	require.Error(t, err)
	assert.EqualError(t, err, "failed to create ICMP request: mock make request error")
}

// TestICMPCheckerCheckWriteDeadlineError verifies the expected behavior.
func TestICMPCheckerCheckWriteDeadlineError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			return []byte{}, nil
		},
		NetworkFunc: func() string {
			return icmpv4Network
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
		name:         "WriteDeadlineErrorChecker",
		address:      testutils.LocalhostIPv4,
		protocol:     mockProtocol,
		readTimeout:  2 * time.Second,
		writeTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	require.Error(t, err)
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
			return icmpv4Network
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
		address:     testutils.LocalhostIPv4,
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	require.Error(t, err)
	assert.EqualError(t, err, "failed to read ICMP reply: mock read error")
}

// TestICMPCheckerSetWriteDeadlineError verifies the expected behavior.
func TestICMPCheckerSetWriteDeadlineError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			return []byte{}, nil
		},
		ValidateReplyFunc: func(reply []byte, id, seq uint16) error {
			return nil
		},
		NetworkFunc: func() string {
			return icmpv4Network
		},
		ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
			return &testutils.MockPacketConn{
				SetWriteDeadlineFunc: func(t time.Time) error {
					return fmt.Errorf("mock write deadline error")
				},
			}, nil
		},
	}

	checker := &ICMPChecker{
		name:        "SetWriteDeadlineErrorChecker",
		address:     testutils.LocalhostIPv4,
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	require.Error(t, err)
	assert.EqualError(t, err, "failed to set write deadline: mock write deadline error")
}

// TestICMPCheckerSetReadDeadlineError verifies the expected behavior.
func TestICMPCheckerSetReadDeadlineError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			return []byte{}, nil
		},
		ValidateReplyFunc: func(reply []byte, id, seq uint16) error {
			return nil
		},
		NetworkFunc: func() string {
			return icmpv4Network
		},
		ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
			return &testutils.MockPacketConn{
				SetReadDeadlineFunc: func(t time.Time) error {
					return fmt.Errorf("mock write deadline error")
				},
			}, nil
		},
	}

	checker := &ICMPChecker{
		name:        "SetReadDeadlineErrorChecker",
		address:     testutils.LocalhostIPv4,
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	require.Error(t, err)
	assert.EqualError(t, err, "failed to set read deadline: mock write deadline error")
}

// TestICMPCheckerValidateReplyError verifies the expected behavior.
func TestICMPCheckerValidateReplyError(t *testing.T) {
	t.Parallel()

	mockProtocol := &testutils.MockProtocol{
		MakeRequestFunc: func(id, seq uint16) ([]byte, error) {
			return []byte{}, nil
		},
		ValidateReplyFunc: func(reply []byte, id, seq uint16) error {
			return fmt.Errorf("unrelated reply")
		},
		NetworkFunc: func() string {
			return icmpv4Network
		},
		ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
			reads := 0
			return &testutils.MockPacketConn{ReadFromFunc: func(b []byte) (int, net.Addr, error) {
				reads++
				if reads > 1 {
					return 0, nil, errors.New("read deadline reached")
				}
				return 0, &net.IPAddr{IP: net.ParseIP("127.0.0.1")}, nil
			}}, nil
		},
	}

	checker := &ICMPChecker{
		name:        "ValidateReplyErrorChecker",
		address:     testutils.LocalhostIPv4,
		protocol:    mockProtocol,
		readTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := checker.Check(ctx)
	require.Error(t, err)
	assert.EqualError(t, err, "failed to read ICMP reply: read deadline reached")
}

func TestICMPDNSIsRetried(t *testing.T) {
	c, err := NewICMPChecker("dns", "eventually-ready.invalid", DefaultICMPConfig())
	require.NoError(t, err)
	calls := 0
	c.lookupIP = func(ctx context.Context, network, host string) ([]net.IP, error) {
		calls++
		if calls == 1 {
			return nil, &net.DNSError{Name: host, IsNotFound: true}
		}
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}
	c.protocol = &testutils.MockProtocol{ListenPacketFunc: func(context.Context, string, string) (net.PacketConn, error) { return &testutils.MockPacketConn{}, nil }}
	require.Error(t, c.Check(context.Background()))
	require.NoError(t, c.Check(context.Background()))
	require.Equal(t, 2, calls)
}

func TestICMPDNSHonorsDeadline(t *testing.T) {
	protocolConfig := DefaultICMPConfig()
	protocolConfig.ReadTimeout = 10 * time.Millisecond
	protocolConfig.WriteTimeout = 10 * time.Millisecond
	c, err := NewICMPChecker("dns", "slow.invalid", protocolConfig)
	require.NoError(t, err)
	c.lookupIP = func(ctx context.Context, _, _ string) ([]net.IP, error) { <-ctx.Done(); return nil, ctx.Err() }
	require.ErrorIs(t, c.Check(context.Background()), context.DeadlineExceeded)
}

func TestICMPReadHonorsCancellation(t *testing.T) {
	c, err := NewICMPChecker("cancel", "127.0.0.1", DefaultICMPConfig())
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	closed := make(chan struct{})
	var once sync.Once
	c.protocol = &testutils.MockProtocol{ListenPacketFunc: func(context.Context, string, string) (net.PacketConn, error) {
		return &testutils.MockPacketConn{
			ReadFromFunc: func([]byte) (int, net.Addr, error) { cancel(); <-closed; return 0, nil, net.ErrClosed },
			CloseFunc:    func() error { once.Do(func() { close(closed) }); return nil },
		}, nil
	}}
	require.ErrorIs(t, c.Check(ctx), context.Canceled)
}

func TestICMPIgnoresUnrelatedPackets(t *testing.T) {
	c, err := NewICMPChecker("matching", "127.0.0.1", DefaultICMPConfig())
	require.NoError(t, err)
	reads, validations := 0, 0
	c.protocol = &testutils.MockProtocol{
		ListenPacketFunc: func(context.Context, string, string) (net.PacketConn, error) {
			return &testutils.MockPacketConn{
				ReadFromFunc: func([]byte) (int, net.Addr, error) {
					reads++
					if reads > 3 {
						return 0, nil, errors.New("no matching reply")
					}
					ip := "127.0.0.1"
					if reads == 1 {
						ip = "127.0.0.2"
					}
					return 0, &net.IPAddr{IP: net.ParseIP(ip)}, nil
				},
			}, nil
		},
		ValidateReplyFunc: func([]byte, uint16, uint16) error {
			validations++
			if validations == 1 {
				return errors.New("other sequence")
			}
			return nil
		},
	}
	require.NoError(t, c.Check(context.Background()))
	require.Equal(t, 3, reads)
}
