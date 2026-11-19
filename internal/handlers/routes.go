package handlers

import (
	"log/slog"
	"net/http"

	"github.com/incognick/base/internal/logging"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func RouteAPI(r chi.Router, _ *sqlx.DB) {
	r.Route("/hello", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("Hello, World!")); err != nil {
				slog.DebugContext(
					r.Context(),
					"hello write failed",
					slog.Any(logging.KeyError, err),
				)
			}
		})
	})
}
