package config_test

import (
	"testing"

	"github.com/incognick/base/internal/config"
)

func lookupFrom(env map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := env[k]

		return v, ok
	}
}

const validDB = "postgres://u:p@localhost:5432/db?sslmode=disable"

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load(
		lookupFrom(map[string]string{"DATABASE_URL": validDB}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Otel.Enabled || cfg.Otel.LogShipping != config.LogShippingStdout {
		t.Fatalf("unexpected defaults: %+v", cfg.Otel)
	}
}

func TestLoadValid(t *testing.T) {
	cfg, err := config.Load(lookupFrom(map[string]string{
		"DATABASE_URL":                validDB,
		"OTEL_ENABLED":                "true",
		"OTEL_LOG_SHIPPING":           "otlp",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4318",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Otel.Enabled || cfg.Otel.LogShipping != config.LogShippingOTLP {
		t.Fatalf("unexpected config: %+v", cfg.Otel)
	}
}

func TestLoadInvalid(t *testing.T) {
	tests := map[string]map[string]string{
		"missing database url": {},
		"bad database scheme":  {"DATABASE_URL": "mysql://localhost/db"},
		"bad log shipping": {
			"DATABASE_URL":      validDB,
			"OTEL_LOG_SHIPPING": "file",
		},
		"bad bool": {"DATABASE_URL": validDB, "OTEL_ENABLED": "1"},
		"bad endpoint": {
			"DATABASE_URL":                validDB,
			"OTEL_EXPORTER_OTLP_ENDPOINT": "localhost",
		},
	}
	for name, env := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := config.Load(lookupFrom(env)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestLoadReportsAllErrors(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"OTEL_ENABLED":      "maybe",
		"OTEL_LOG_SHIPPING": "file",
	}))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, name := range []string{"DATABASE_URL", "OTEL_ENABLED", "OTEL_LOG_SHIPPING"} {
		if !contains(err.Error(), name) {
			t.Errorf("error should mention %s: %v", name, err)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}

	return false
}
