package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/incognick/base/internal/logging"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/fx"

	"github.com/incognick/base/internal/handlers"
)

const (
	requestTimeout    = 5 * time.Minute
	idleTimeout       = 60 * time.Second
	corsMaxAgeSeconds = 300
)

func NewHTTPRouter(db *sqlx.DB, logger *slog.Logger) (chi.Router, error) {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(requestLogger(logger))
	r.Use(middleware.Timeout(requestTimeout))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"PATCH",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           corsMaxAgeSeconds,
		AllowOriginFunc: func(_ *http.Request, _ string) bool {
			return true
		},
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("ok")); err != nil {
			logger.DebugContext(
				r.Context(),
				"health write failed",
				slog.Any(logging.KeyError, err),
			)
		}
	})

	// Serve API endpoints from /api
	r.Route("/api/v1", func(r chi.Router) {
		handlers.RouteAPI(r, db)
	})

	return r, nil
}

func NewHTTPServer(
	lc fx.Lifecycle,
	r chi.Router,
	logger *slog.Logger,
) *http.Server {
	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      traced(r),
		ReadTimeout:  0,
		WriteTimeout: 0,
		IdleTimeout:  idleTimeout,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := httpServer.ListenAndServe(); err != nil &&
					!errors.Is(err, http.ErrServerClosed) {
					logger.ErrorContext(
						ctx,
						"Server error",
						slog.Any(logging.KeyError, err),
					)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.InfoContext(ctx, "Shutting down server")
			return httpServer.Shutdown(ctx)
		},
	})

	return httpServer
}

// traced wraps h so every request gets a span, and continues the caller's trace when it sent a traceparent header.
// It is a no-op until telemetry is enabled, because the global tracer provider is then the SDK's no-op one.
func traced(h http.Handler) http.Handler {
	return otelhttp.NewHandler(
		h,
		"http.server",
		otelhttp.WithFilter(
			func(r *http.Request) bool { return r.URL.Path != "/health" },
		),
	)
}
