/*
Package logger provides structured logging for goride services built on log/slog.

Call Setup once at the beginning of main, then use L() (or the slog default
logger) everywhere else. Every record automatically includes:

  - "service":  the configured service name
  - "trace_id"/"span_id": OpenTelemetry trace context, when a span is
    present in the request context

Configuration via environment variables:
  - LOG_LEVEL:  debug | info | warn | error (default: info)
  - LOG_FORMAT: json | text (default: json)
*/
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/trunglq04/goride/shared/env"
	"go.opentelemetry.io/otel/trace"
)

var (
	setupOnce sync.Once
)

// Setup initializes the process-wide default slog logger. It is safe to call
// multiple times — only the first call takes effect.
func Setup(serviceName string) {
	setupOnce.Do(func() {
		level := parseLevel(env.GetString("LOG_LEVEL", "INFO"))
		format := strings.ToLower(strings.TrimSpace(env.GetString("LOG_FORMAT", "json")))

		opts := &slog.HandlerOptions{Level: level}
		var handler slog.Handler
		if format == "text" {
			handler = slog.NewTextHandler(os.Stdout, opts)
		} else {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		}

		l := slog.New(&traceHandler{Handler: handler}).With(slog.String("service", serviceName))
		slog.SetDefault(l)
	})
}

// L returns the default logger.
func L() *slog.Logger {
	return slog.Default()
}

// Fatal logs at error level and terminates the process with exit code 1.
// Use only during startup, before it is safe to keep running.
func Fatal(msg string, args ...any) {
	slog.Default().Error(msg, args...)
	os.Exit(1)
}

// FatalContext is like Fatal but includes the trace context from ctx.
func FatalContext(ctx context.Context, msg string, args ...any) {
	slog.Default().ErrorContext(ctx, msg, args...)
	os.Exit(1)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// traceHandler enriches every record with the OpenTelemetry trace IDs found
// in the context, so logs can be correlated with distributed traces.
type traceHandler struct {
	slog.Handler
}

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() && sc.HasTraceID() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithGroup(name)}
}
