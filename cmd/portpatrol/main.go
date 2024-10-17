package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/containeroo/portpatrol/internal/flags"
	"github.com/containeroo/portpatrol/internal/logger"
	"github.com/containeroo/portpatrol/internal/wait"
	"golang.org/x/sync/errgroup"
)

const version = "0.5.0"

// run is the main function of the application.
func run(ctx context.Context, args []string, output io.Writer) error {
	// Create a new context that listens for interrupt signals
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	f, err := flags.ParseCommandLineFlags(args, version)
	if err != nil {
		switch e := err.(type) {
		case *flags.HelpRequested:
			fmt.Fprintf(output, e.Message)
			return nil
		case *flags.VersionRequested:
			fmt.Fprintf(output, e.Version)
			return nil
		default:
			return fmt.Errorf("configuration error: %s", err)
		}
	}
	f.Version = version

	checkers, err := flags.InitializeTargetCheckers(f.Targets, f.DefaultCheckInterval)
	if err != nil {
		return fmt.Errorf("initalize target checkers error: %w", err)
	}

	logger := logger.SetupLogger(f, output)

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

	return eg.Wait()
}

func main() {
	// Create a root context with no cancellation or deadline.
	ctx := context.Background()

	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
