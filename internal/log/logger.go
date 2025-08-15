// Package log provides application logging helpers built on top of slog.
package log

import (
	"log/slog"
	"os"
)

// New constructs a new slog.Logger that writes text logs to stdout at the given level.
func New(level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
