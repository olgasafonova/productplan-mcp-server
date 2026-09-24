// Package logging configures the server's structured JSON logger.
//
// It is log/slog with one handler setup: JSON to stderr (stdout carries the
// MCP protocol), with the built-in keys renamed to the ones this server has
// always emitted: "ts" (RFC 3339, UTC, nanoseconds), "level" (lower case)
// and "msg". Callers log through *slog.Logger and build attributes with
// slog directly; the helpers below exist only for values that need a
// conversion (durations as milliseconds, errors as their message).
package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// New returns a JSON logger writing to stderr at level and above.
func New(level slog.Leveler) *slog.Logger {
	return NewWithWriter(os.Stderr, level)
}

// NewWithWriter returns a JSON logger writing to w at level and above.
func NewWithWriter(w io.Writer, level slog.Leveler) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: renameBuiltins,
	}))
}

// Nop returns a logger that discards everything.
func Nop() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

// renameBuiltins maps slog's top-level time and level attributes onto the
// server's field names and formats. Attributes inside groups pass through.
func renameBuiltins(groups []string, a slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return a
	}
	switch a.Key {
	case slog.TimeKey:
		return slog.String("ts", a.Value.Time().UTC().Format(time.RFC3339Nano))
	case slog.LevelKey:
		return slog.String(slog.LevelKey, strings.ToLower(a.Value.String()))
	}
	return a
}

// Duration logs d as whole milliseconds under "dur_ms".
func Duration(d time.Duration) slog.Attr {
	return slog.Int64("dur_ms", d.Milliseconds())
}

// Error logs err's message under "error".
func Error(err error) slog.Attr {
	return slog.String("error", err.Error())
}
