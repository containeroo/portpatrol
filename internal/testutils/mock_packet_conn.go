package testutils

import (
	"net"
	"time"
)

// MockProtocol is a mock implementation of the Protocol interface for testing.
type MockProtocol struct {
	MakeRequestFunc   func(identifier, sequence uint16) ([]byte, error)
	ValidateReplyFunc func(reply []byte, identifier, sequence uint16) error
	NetworkFunc       func() string
	ListenPacketFunc  func(network, address string) (net.PacketConn, error)
	SetDeadlineFunc   func(t time.Time) error
}

func (m *MockProtocol) MakeRequest(identifier, sequence uint16) ([]byte, error) {
	return m.MakeRequestFunc(identifier, sequence)
}

func (m *MockProtocol) ValidateReply(reply []byte, identifier, sequence uint16) error {
	return m.ValidateReplyFunc(reply, identifier, sequence)
}

func (m *MockProtocol) Network() string {
	return m.NetworkFunc()
}

func (m *MockProtocol) ListenPacket(network, address string) (net.PacketConn, error) {
	return m.ListenPacketFunc(network, address)
}

func (m *MockProtocol) SetDeadline(t time.Time) error {
	return m.SetDeadlineFunc(t)
}

// MockPacketConn is a mock implementation of net.PacketConn for testing purposes.
type MockPacketConn struct {
	SetDeadlineFunc      func(t time.Time) error
	SetReadDeadlineFunc  func(t time.Time) error
	SetWriteDeadlineFunc func(t time.Time) error
	WriteToFunc          func(b []byte, addr net.Addr) (int, error)
	ReadFromFunc         func(b []byte) (int, net.Addr, error)
	CloseFunc            func() error
	LocalAddrFunc        func() net.Addr
	RemoteAddrFunc       func() net.Addr
}

func (m *MockPacketConn) SetDeadline(t time.Time) error {
	if m.SetDeadlineFunc != nil {
		return m.SetDeadlineFunc(t)
	}
	return nil
}

func (m *MockPacketConn) SetReadDeadline(t time.Time) error {
	if m.SetReadDeadlineFunc != nil {
		return m.SetReadDeadlineFunc(t)
	}
	return nil
}

func (m *MockPacketConn) SetWriteDeadline(t time.Time) error {
	if m.SetWriteDeadlineFunc != nil {
		return m.SetWriteDeadlineFunc(t)
	}
	return nil
}

func (m *MockPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if m.WriteToFunc != nil {
		return m.WriteToFunc(b, addr)
	}
	return len(b), nil
}

func (m *MockPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if m.ReadFromFunc != nil {
		return m.ReadFromFunc(b)
	}
	return 0, nil, nil
}

func (m *MockPacketConn) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockPacketConn) LocalAddr() net.Addr {
	if m.LocalAddrFunc != nil {
		return m.LocalAddrFunc()
	}
	return &net.IPAddr{}
}

func (m *MockPacketConn) RemoteAddr() net.Addr {
	if m.RemoteAddrFunc != nil {
		return m.RemoteAddrFunc()
	}
	return &net.IPAddr{}
}
