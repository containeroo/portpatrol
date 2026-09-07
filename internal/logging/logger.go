package logging

import (
	"fmt"
	"io"
	"log/slog"
)

// LogFormat defines the supported log formats.
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// SetupLogger configures a structured logger.
func SetupLogger(logFormat LogFormat, output io.Writer, showPath ...bool) *slog.Logger {
	visible := len(showPath) > 0 && showPath[0]
	handlerOpts := &slog.HandlerOptions{ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Value.Kind() == slog.KindString {
			a.Value = slog.StringValue(RedactURLs(a.Value.String(), visible))
		}
		if a.Value.Kind() == slog.KindAny {
			if err, ok := a.Value.Any().(error); ok {
				a.Value = slog.StringValue(RedactURLs(fmt.Sprint(err), visible))
			}
		}
		return a
	}}

	var handler slog.Handler
	switch logFormat {
	case LogFormatJSON:
		handler = slog.NewJSONHandler(output, handlerOpts)
	case LogFormatText:
		handler = slog.NewTextHandler(output, handlerOpts)
	default:
		// Default to JSON if an invalid format is provided.
		handler = slog.NewJSONHandler(output, handlerOpts)
	}

	return slog.New(handler)
}
