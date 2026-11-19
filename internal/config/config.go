// Package config loads and validates the service configuration from
// environment variables, so a bad deployment fails at startup instead of
// somewhere deep in a request.
//
// Only variables owned by this service are validated here. Standard OTEL_*
// variables (sampler, batching, ...) are read by the OTel SDK itself.
package config

import "errors"

// Values accepted by OTEL_LOG_SHIPPING.
const (
	LogShippingOTLP   = "otlp"
	LogShippingStdout = "stdout"
)

const defaultLogShipping = LogShippingStdout

// Config is the complete, validated service configuration.
type Config struct {
	// DatabaseURL is a postgres:// or postgresql:// connection string.
	DatabaseURL string
	Otel        Otel
}

// Otel holds OpenTelemetry settings.
type Otel struct {
	// Enabled turns tracing (and OTLP log shipping) on.
	Enabled bool
	// ExporterEndpoint is OTEL_EXPORTER_OTLP_ENDPOINT, validated here and
	// read by the SDK. Empty means unset (the SDK default applies).
	ExporterEndpoint string
	// LogShipping is otlp or stdout; only meaningful when Enabled.
	LogShipping string
}

// Load reads the configuration through lookup (pass os.LookupEnv).
// It returns every invalid variable at once.
//
// Call it after .env files are loaded (godotenv/autoload does that in
// package init, before main runs), because lookup sees the process
// environment only.
func Load(lookup func(key string) (string, bool)) (*Config, error) {
	l := &loader{lookup: lookup}

	cfg := &Config{
		DatabaseURL: l.requiredURL("DATABASE_URL", "postgres", "postgresql"),
		Otel: Otel{
			Enabled: l.boolean("OTEL_ENABLED", false),
			ExporterEndpoint: l.optionalURL(
				"OTEL_EXPORTER_OTLP_ENDPOINT",
				"http",
				"https",
			),
			LogShipping: l.oneOf(
				"OTEL_LOG_SHIPPING",
				defaultLogShipping,
				LogShippingOTLP,
				LogShippingStdout,
			),
		},
	}

	if err := errors.Join(l.errs...); err != nil {
		return nil, err
	}

	return cfg, nil
}
