// Package logging provides structured JSON logging for the MCP server.
// All logs go to stderr since stdout is reserved for MCP protocol messages.
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents log severity levels.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLevel converts a string to a Level.
func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value any
}

// F creates a Field with the given key and value.
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// Common field constructors for type safety and consistency.
func RequestID(id string) Field      { return F("req_id", id) }
func Operation(op string) Field      { return F("op", op) }
func Duration(d time.Duration) Field { return F("dur_ms", d.Milliseconds()) }
func Status(s string) Field          { return F("status", s) }
func Error(err error) Field          { return F("error", err.Error()) }
func Tool(name string) Field         { return F("tool", name) }
func Endpoint(e string) Field        { return F("endpoint", e) }
func StatusCode(code int) Field      { return F("status_code", code) }
func Count(n int) Field              { return F("count", n) }

// Logger is the interface for structured logging.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	WithFields(fields ...Field) Logger
	WithRequestID(id string) Logger
}

// JSONLogger implements Logger with JSON output.
type JSONLogger struct {
	out    io.Writer
	level  Level
	fields []Field
	mu     sync.Mutex
}

// New creates a new JSONLogger writing to stderr.
func New(level Level) *JSONLogger {
	return &JSONLogger{
		out:   os.Stderr,
		level: level,
	}
}

// NewWithWriter creates a JSONLogger with a custom writer (for testing).
func NewWithWriter(w io.Writer, level Level) *JSONLogger {
	return &JSONLogger{
		out:   w,
		level: level,
	}
}

// Nop returns a no-op logger that discards all output.
func Nop() Logger {
	return &nopLogger{}
}

type nopLogger struct{}

func (n *nopLogger) Debug(msg string, fields ...Field) {}
func (n *nopLogger) Info(msg string, fields ...Field)  {}
func (n *nopLogger) Warn(msg string, fields ...Field)  {}
func (n *nopLogger) Error(msg string, fields ...Field) {}
func (n *nopLogger) WithFields(fields ...Field) Logger { return n }
func (n *nopLogger) WithRequestID(id string) Logger    { return n }

func (l *JSONLogger) log(level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Build the log entry as a map for flexible field ordering
	entry := make(map[string]any, len(l.fields)+len(fields)+3)
	entry["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	entry["level"] = level.String()
	entry["msg"] = msg

	// Add base fields
	for _, f := range l.fields {
		entry[f.Key] = f.Value
	}

	// Add call-specific fields (can override base fields)
	for _, f := range fields {
		entry[f.Key] = f.Value
	}

	// Encode and write
	data, err := json.Marshal(entry)
	if err != nil {
		_, _ = fmt.Fprintf(l.out, `{"ts":"%s","level":"error","msg":"failed to marshal log entry","error":"%s"}`+"\n",
			time.Now().UTC().Format(time.RFC3339Nano), err.Error())
		return
	}

	_, _ = l.out.Write(data)
	_, _ = l.out.Write([]byte{'\n'})
}

func (l *JSONLogger) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, fields...)
}

func (l *JSONLogger) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, fields...)
}

func (l *JSONLogger) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, fields...)
}

func (l *JSONLogger) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
}

// WithFields returns a new logger with additional base fields.
func (l *JSONLogger) WithFields(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &JSONLogger{
		out:    l.out,
		level:  l.level,
		fields: newFields,
	}
}

// WithRequestID returns a new logger with the request ID field.
func (l *JSONLogger) WithRequestID(id string) Logger {
	return l.WithFields(RequestID(id))
}

// SetLevel changes the minimum log level.
func (l *JSONLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level.
func (l *JSONLogger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}
