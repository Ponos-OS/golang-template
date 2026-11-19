package main

import (
	"net/http"
	"time"

	"go.uber.org/fx"

	// Autoload .env file
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fx.New(
		fx.Provide(
			TelemetryProvider,
			LoggerProvider,
			DatabaseProvider,
			NewHTTPServer,
			NewHTTPRouter,
		),
		fx.Invoke(
			func(_ *http.Server) {},
		),
		fx.StopTimeout(1*time.Minute),
	).Run()
}
