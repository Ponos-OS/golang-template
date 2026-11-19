// Package telemetry sets up OpenTelemetry tracing and log shipping.
//
// Only two switches are ours (OTEL_ENABLED and OTEL_LOG_SHIPPING).
// Everything else (endpoint, sampler, batching, resource attributes, ...)
// is read by the OTel SDK from the standard OTEL_* environment variables.
// See the "OpenTelemetry" section of the README.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

const (
	serviceName     = "base"
	logShippingOTLP = "otlp"
)

// Telemetry owns the OpenTelemetry providers. The zero value is the disabled state.
type Telemetry struct {
	traces     *sdktrace.TracerProvider
	logs       *sdklog.LoggerProvider
	logHandler slog.Handler
}

// New starts tracing and, when OTEL_LOG_SHIPPING=otlp, log shipping.
// With OTEL_ENABLED unset or false it does nothing and returns a disabled Telemetry.
func New(ctx context.Context) (*Telemetry, error) {
	enabled, err := envBool("OTEL_ENABLED")
	if err != nil {
		return nil, err
	}
	if !enabled {
		return &Telemetry{}, nil
	}

	// Later options win, so OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES override our default name.
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("build resource: %w", err)
	}

	t := &Telemetry{}
	if err = t.startTraces(ctx, res); err != nil {
		return nil, err
	}
	if os.Getenv("OTEL_LOG_SHIPPING") == logShippingOTLP {
		if err = t.startLogs(ctx, res); err != nil {
			return nil, errors.Join(err, t.Shutdown(ctx))
		}
	}

	return t, nil
}

func (t *Telemetry) startTraces(
	ctx context.Context,
	res *resource.Resource,
) error {
	// Endpoint comes from OTEL_EXPORTER_OTLP_ENDPOINT.
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return fmt.Errorf("create trace exporter: %w", err)
	}

	// Sampler comes from OTEL_TRACES_SAMPLER(_ARG), batching from OTEL_BSP_*.
	t.traces = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(t.traces)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

func (t *Telemetry) startLogs(
	ctx context.Context,
	res *resource.Resource,
) error {
	exporter, err := otlploghttp.New(ctx)
	if err != nil {
		return fmt.Errorf("create log exporter: %w", err)
	}

	// Batching comes from OTEL_BLRP_*.
	t.logs = sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
	t.logHandler = otelslog.NewHandler(
		serviceName,
		otelslog.WithLoggerProvider(t.logs),
	)

	return nil
}

// LogHandler returns the handler that ships records to the collector, or nil when logs are not shipped.
func (t *Telemetry) LogHandler() slog.Handler {
	return t.logHandler
}

// Shutdown flushes what is queued and stops the exporters.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error
	if t.traces != nil {
		errs = append(errs, t.traces.Shutdown(ctx))
	}
	if t.logs != nil {
		errs = append(errs, t.logs.Shutdown(ctx))
	}

	return errors.Join(errs...)
}

func envBool(key string) (bool, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return false, nil
	}

	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s: %w", key, err)
	}

	return v, nil
}
