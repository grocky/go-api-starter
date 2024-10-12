package log

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// contextKey is a private string type to prevent collisions in the context map.
type contextKey string

// loggerKey points to the value in the context where the logger is stored.
const loggerKey = contextKey("logger")

var (
	// defaultLogger is the default logger. It is initialized once per package
	// include upon calling DefaultLogger.
	defaultLogger     *Logger
	defaultLoggerOnce sync.Once
)

type Logger struct {
	l     *slog.Logger
	level slog.Level
	name  string
}

// New creates a new Logger with the given configuration.
func New(w io.Writer, level slog.Level) *Logger {
	return &Logger{
		l:     slog.New(newHandler(level, WithDestinationWriter(w))),
		level: level,
	}
}

func DefaultLogger() *Logger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = New(os.Stdout, slog.LevelInfo)
	})
	return defaultLogger
}

// Named adds a new path segment to the logger's name. Segments are joined by periods.
// By default, Loggers are unnamed.
func (l *Logger) Named(s string) *Logger {
	if s == "" {
		return l
	}
	c := l.clone()
	if l.name == "" {
		c.name = s
	} else {
		c.name = strings.Join([]string{c.name, s}, ".")
	}

	c.l = c.l.With("name", c.name)

	return c
}

func (l *Logger) Debug(msg string, args ...any) {
	l.l.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.l.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.l.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.l.Error(msg, args...)
}

func (l *Logger) With(args ...any) *Logger {
	c := l.clone()
	c.l = c.l.With(args)

	return c
}

func (l *Logger) CreateLogLogger() *log.Logger {
	return slog.NewLogLogger(l.l.Handler(), l.level)
}

func (l *Logger) clone() *Logger {
	clone := *l
	return &clone
}

func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return logger
	}
	return DefaultLogger()
}
