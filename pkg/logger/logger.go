package logger

import (
	"io"
	"log/slog"

	"github.com/containeroo/thor/pkg/config"
)

// SetupLogger configures the logger based on the configuration
func SetupLogger(cfg config.Config, output io.Writer) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{}

	if cfg.LogAdditionalFields {
		// Return a logger with the additional fields
		return slog.New(slog.NewTextHandler(output, handlerOpts)).With(
			slog.String("target_address", cfg.TargetAddress),
			slog.String("interval", cfg.Interval.String()),
			slog.String("dial_timeout", cfg.DialTimeout.String()),
			slog.String("checker_type", cfg.CheckType),
			slog.String("version", cfg.Version),
		)
	}

	// If logAdditionalFields is false, remove the error attribute from the handler
	handlerOpts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "error" {
			return slog.Attr{}
		}
		return a
	}

	// Return a logger without the additional fields and with the error attribute
	return slog.New(slog.NewTextHandler(output, handlerOpts))
}
