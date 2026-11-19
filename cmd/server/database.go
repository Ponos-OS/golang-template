package main

import (
	"context"
	"errors"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

// ErrDatabaseURLNotSet is returned when DATABASE_URL is missing.
var ErrDatabaseURLNotSet = errors.New(
	"DATABASE_URL environment variable not set",
)

func DatabaseProvider(lc fx.Lifecycle) (*sqlx.DB, error) {
	url, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, ErrDatabaseURLNotSet
	}

	var err error
	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}
