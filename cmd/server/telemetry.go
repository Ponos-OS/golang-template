package main

import (
	"context"
	"log/slog"
	"os"

	"go.uber.org/fx"

	"github.com/incognick/base/internal/config"
	"github.com/incognick/base/internal/logging"
	"github.com/incognick/base/internal/telemetry"
)

// TelemetryProvider starts OpenTelemetry and flushes it on shutdown.
// The logger and the HTTP server depend on it, so fx stops them first and their shutdown logs are still flushed.
func TelemetryProvider(
	lc fx.Lifecycle,
	cfg *config.Config,
) (*telemetry.Telemetry, error) {
	t, err := telemetry.New(context.Background(), cfg.Otel)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{OnStop: t.Shutdown})

	return t, nil
}

// LoggerProvider builds the logger and makes it the slog default.
func LoggerProvider(t *telemetry.Telemetry) *slog.Logger {
	logger := logging.New(os.Stdout, t.LogHandler())
	slog.SetDefault(logger)

	return logger
}
