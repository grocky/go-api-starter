package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

const timeFormat = "[15:04:05.000]"

type Replacer func([]string, slog.Attr) slog.Attr

func suppressDefaults(next Replacer) Replacer {
	return func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.TimeKey, slog.LevelKey, slog.MessageKey:
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}

type handler struct {
	h                slog.Handler
	b                *bytes.Buffer
	m                *sync.Mutex
	writer           io.Writer
	outputEmptyAttrs bool
}

func newHandler(level slog.Level, options ...Option) *handler {
	opts := &slog.HandlerOptions{
		AddSource: false, // when slog.Logger is wrapped, source is always in the wrapper
		Level:     level,
	}

	buf := &bytes.Buffer{}
	handler := &handler{
		b: buf,
		h: slog.NewJSONHandler(buf, &slog.HandlerOptions{
			AddSource:   opts.AddSource,
			Level:       opts.Level,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		m: &sync.Mutex{},
	}

	for _, opt := range options {
		opt(handler)
	}

	return handler
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{
		h:                h.h.WithAttrs(attrs),
		b:                h.b,
		m:                h.m,
		writer:           h.writer,
		outputEmptyAttrs: h.outputEmptyAttrs,
	}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{
		h:                h.h.WithGroup(name),
		b:                h.b,
		m:                h.m,
		writer:           h.writer,
		outputEmptyAttrs: h.outputEmptyAttrs,
	}
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	var colorCode code

	switch r.Level {
	case slog.LevelDebug:
		colorCode = darkGray
	case slog.LevelInfo:
		colorCode = green
	case slog.LevelWarn:
		colorCode = lightYellow
	case slog.LevelError:
		colorCode = lightRed
	}

	level := colorize(colorCode, r.Level.String()+":")
	timestamp := colorize(lightGray, r.Time.Format(timeFormat))
	msg := colorize(white, r.Message)

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	var attrsAsBytes []byte

	if h.outputEmptyAttrs || len(attrs) > 0 {
		attrsAsBytes, err = json.MarshalIndent(attrs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal attributes: %w", err)
		}
	}

	attrsColorized := colorize(darkGray, string(attrsAsBytes))

	out := strings.Builder{}
	if len(timestamp) > 0 {
		out.WriteString(timestamp)
		out.WriteString(" ")
	}
	if len(level) > 0 {
		out.WriteString(level)
		out.WriteString(" ")
	}
	if len(msg) > 0 {
		out.WriteString(msg)
		out.WriteString(" ")
	}
	if len(attrsColorized) > 0 {
		out.WriteString(attrsColorized)
	}

	if _, err := io.WriteString(h.writer, out.String()+"\n"); err != nil {
		return err
	}

	return nil
}

func (h *handler) computeAttrs(ctx context.Context, r slog.Record) (map[string]any,
	error) {
	h.m.Lock()
	defer func() {
		h.b.Reset()
		h.m.Unlock()
	}()

	if err := h.h.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner http's Handle: %w", err)
	}

	var attrs map[string]any
	if err := json.Unmarshal(h.b.Bytes(), &attrs); err != nil {
		return nil, fmt.Errorf("error when unmarshaling attrs: %w", err)
	}

	// only provide source in debug and error levels
	if r.Level > slog.LevelDebug && r.Level < slog.LevelError {
		delete(attrs, slog.SourceKey)
	}

	return attrs, nil
}

type Option func(h *handler)

func WithDestinationWriter(writer io.Writer) Option {
	return func(h *handler) {
		h.writer = writer
	}
}
