package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/incognick/base/internal/logging"
)

// requestLogger logs one line per request.
// It must run inside otelhttp so the record carries the trace and span IDs.
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			logger.InfoContext(
				r.Context(),
				"request completed",
				slog.String(
					logging.KeyRequestID,
					middleware.GetReqID(
						r.Context(),
					), // What Chi has, different from trace ID
				),
				slog.String(logging.KeyMethod, r.Method),
				slog.String(logging.KeyPath, r.URL.Path),
				slog.Int(logging.KeyStatus, status),
				slog.Int(logging.KeyBytes, ww.BytesWritten()),
				slog.Duration(logging.KeyDuration, time.Since(start)),
			)
		})
	}
}
