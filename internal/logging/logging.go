// Package logging builds the service logger on top of log/slog.
package logging

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// Log attribute keys shared across the service.
const (
	KeyTraceID   = "trace_id"
	KeySpanID    = "span_id"
	KeyRequestID = "request_id"
	KeyMethod    = "method"
	KeyPath      = "path"
	KeyStatus    = "status"
	KeyBytes     = "bytes"
	KeyDuration  = "duration"
	KeyError     = "error"
)

// New creates a logger that writes human-readable lines to out (e.g. os.Stdout).
// Records also go to each non-nil handler in also, e.g. the OTLP log shipper.
// The lines written to out carry trace_id and span_id when the context has a span,
// so pass the request context: logger.InfoContext(ctx, ...).
func New(out io.Writer, also ...slog.Handler) *slog.Logger {
	handlers := []slog.Handler{traceHandler{slog.NewTextHandler(out, nil)}}
	for _, h := range also {
		if h != nil {
			handlers = append(handlers, h)
		}
	}

	if len(handlers) == 1 {
		return slog.New(handlers[0])
	}

	return slog.New(slog.NewMultiHandler(handlers...))
}

// traceHandler adds the trace context of the record's context.
type traceHandler struct{ slog.Handler }

func (h traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String(KeyTraceID, sc.TraceID().String()),
			slog.String(KeySpanID, sc.SpanID().String()),
		)
	}

	return h.Handler.Handle(ctx, r)
}

func (h traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceHandler{h.Handler.WithAttrs(attrs)}
}

func (h traceHandler) WithGroup(name string) slog.Handler {
	return traceHandler{h.Handler.WithGroup(name)}
}
