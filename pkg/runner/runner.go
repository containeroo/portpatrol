package runner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/toast/pkg/checker"
	"github.com/containeroo/toast/pkg/config"
)

// RunLoop continuously attempts to connect to the specified target until it becomes available or the context is canceled.
func RunLoop(ctx context.Context, cfg config.Config, checker checker.Checker, logger *slog.Logger) error {
	logger.Info(fmt.Sprintf("Waiting for %s to become ready...", checker))

	for {
		err := checker.Check(ctx)
		if err == nil {
			logger.Info(fmt.Sprintf("%s is ready ✓", checker))
			return nil
		}

		logger.Warn(fmt.Sprintf("%s is not ready ✗", checker), slog.String("error", err.Error()))

		select {
		case <-time.After(cfg.Interval):
			// Continue to the next connection attempt after the interval
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				return nil // Treat context cancellation as expected behavior
			}
			return ctx.Err()
		}
	}
}
