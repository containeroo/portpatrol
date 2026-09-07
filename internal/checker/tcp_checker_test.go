package checker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/containeroo/never/internal/testutils"
)

// TestNewTCPChecker_Valid verifies the expected behavior.
func TestNewTCPChecker_Valid(t *testing.T) {
	t.Parallel()

	listener := testutils.ListenLocalTCP(t)
	defer listener.Close() // nolint:errcheck

	protocolConfig := DefaultTCPConfig()
	protocolConfig.Timeout = 1 * time.Second
	checker, err := NewTCPChecker("example", listener.Addr().String(), protocolConfig)
	require.NoError(t, err)

	assert.Equal(t, checker.Name(), "example")
	assert.Equal(t, checker.Address(), listener.Addr().String())
	assert.Equal(t, checker.Type(), TCP.String())
}

// TestTCPChecker_ValidConnection verifies the expected behavior.
func TestTCPChecker_ValidConnection(t *testing.T) {
	t.Parallel()

	listener := testutils.ListenLocalTCP(t)
	defer listener.Close() // nolint:errcheck

	protocolConfig := DefaultTCPConfig()
	protocolConfig.Timeout = 1 * time.Second
	checker, err := NewTCPChecker("example", listener.Addr().String(), protocolConfig)
	require.NoError(t, err)

	ctx := context.Background()
	err = checker.Check(ctx)
	require.NoError(t, err)
}

// TestTCPChecker_FailedConnection verifies the expected behavior.
func TestTCPChecker_FailedConnection(t *testing.T) {
	t.Parallel()

	protocolConfig := DefaultTCPConfig()
	protocolConfig.Timeout = 1 * time.Second
	checker, err := NewTCPChecker("example", testutils.LocalTCPAddr(t), protocolConfig)
	require.NoError(t, err)

	ctx := context.Background()
	err = checker.Check(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "connect: connection refused")
}

// TestTCPChecker_Timeout verifies the expected behavior.
func TestTCPChecker_Timeout(t *testing.T) {
	t.Parallel()

	protocolConfig := DefaultTCPConfig()
	protocolConfig.Timeout = 1 * time.Second
	checker, err := NewTCPChecker("example", testutils.LocalTCPAddr(t), protocolConfig)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)

	err = checker.Check(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
