package main

import (
	"net/http"
	"time"

	"go.uber.org/fx"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fx.New(
		fx.Provide(
			ConfigProvider,
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
