//revive:disable:var-naming
package common

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/rs/zerolog/log"
)

const (
	devDatabaseHost = "postgres://postgres:password@0.0.0.0:5432/jobs?sslmode=disable"
)

func NewInsertOnlyRiverClient() *river.Client[pgx.Tx] {
	dbPool, err := pgxpool.New(context.Background(), devDatabaseHost)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating job queue database connection")
	}

	client, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("error creating river client")
	}

	return client
}
