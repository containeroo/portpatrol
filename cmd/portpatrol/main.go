package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/containeroo/portpatrol/internal/config"
	"github.com/containeroo/portpatrol/internal/factory"
	"github.com/containeroo/portpatrol/internal/logging"
	"github.com/containeroo/portpatrol/internal/wait"
	"golang.org/x/sync/errgroup"
)

const version = "0.5.0"

// run is the main function of the application.
func run(ctx context.Context, args []string, output io.Writer) error {
	// Create a new context that listens for interrupt signals
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Parse command-line flags
	f, err := config.ParseFlags(args, version)
	if err != nil {
		var helpErr *config.HelpRequested
		if errors.As(err, &helpErr) {
			fmt.Fprint(output, helpErr.Message)
			return nil
		}
		return fmt.Errorf("configuration error: %w", err)
	}

	// Initialize target checkers
	checkers, err := factory.BuildCheckers(f.DynFlags, f.DefaultCheckInterval)
	if err != nil {
		return fmt.Errorf("failed to initialize target checkers: %w", err)
	}

	if len(checkers) == 0 {
		return errors.New("configuration error: no checkers configured")
	}

	logger := logging.SetupLogger(version, output)

	// Run checkers concurrently
	eg, ctx := errgroup.WithContext(ctx)
	for _, chk := range checkers {
		checker := chk // Capture loop variable
		eg.Go(func() error {
			err := wait.WaitUntilReady(ctx, checker.Interval, checker.Checker, logger)
			if err != nil {
				return fmt.Errorf("checker '%s' failed: %w", checker.Checker.GetName(), err)
			}
			return nil
		})
	}

	// Wait for all checkers to finish or return error
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func main() {
	// Create a root context
	ctx := context.Background()

	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
