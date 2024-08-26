package runner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
)

// LoopUntilReady continuously attempts to connect to the specified target until it becomes available or the context is canceled.
func LoopUntilReady(ctx context.Context, interval time.Duration, checker checker.Checker, logger *slog.Logger) error {
	logger.Info(fmt.Sprintf("Waiting for %s to become ready...", checker))

	for {
		err := checker.Check(ctx)
		if err == nil {
			logger.Info(fmt.Sprintf("%s is ready ✓", checker))
			return nil // Successfully connected to the target
		}

		logger.Warn(fmt.Sprintf("%s is not ready ✗", checker), slog.String("error", err.Error()))

		select {
		case <-time.After(interval):
			// Continue to the next connection attempt after the interval
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				return nil // Treat context cancellation as expected behavior
			}
			return ctx.Err()
		}
	}
}
