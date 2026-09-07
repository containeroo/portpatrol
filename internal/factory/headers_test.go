package factory_test

import (
	"testing"
	"time"

	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCheckersDefaultsNameToID(t *testing.T) {
	t.Parallel()

	checkers, err := factory.BuildCheckers([]factory.TargetConfig{
		{
			ID:      "database",
			Address: testutils.LocalhostAddr("5432"),
			Config:  checker.DefaultTCPConfig(),
		},
	}, time.Second)

	require.NoError(t, err)
	require.Len(t, checkers, 1)
	assert.Equal(t, "database", checkers[0].Checker.Name())
}
