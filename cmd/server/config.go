package main

import (
	"os"

	"github.com/incognick/base/internal/config"
)

// ConfigProvider loads and validates the environment. It is the first provider every other one depends on, so a bad environment aborts startup before anything is connected or started.
//
// godotenv/autoload populates the process environment in package init, which runs before main, so os.LookupEnv already sees .env values here.
// Real environment variables win over .env, as godotenv never overrides them.
func ConfigProvider() (*config.Config, error) {
	return config.Load(os.LookupEnv)
}
