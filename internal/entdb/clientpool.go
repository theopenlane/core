package entdb

import (
	"context"

	"ariga.io/entcache"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/theopenlane/entx"

	ent "github.com/theopenlane/core/internal/ent/generated"
)

// NewDBPool creates a new database pool with the given configuration
func NewDBPool(ctx context.Context, cfg *entx.Config) (context.Context, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.PrimaryDBSource)
	if err != nil {
		return nil, ErrFailedToParseConnectionString
	}

	// TODO build config from entx.Config

	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()
	poolConfig.MaxConns = 20 // nolint:gomnd

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, ErrFfailedToConnectToDatabase
	}

	poolDriver := NewPgxPoolDriver(pool)

	cacheDriver := entcache.NewDriver(
		poolDriver,
		entcache.ContextLevel(),
	)

	realClient := ent.NewClient(
		ent.Driver(cacheDriver),
	)

	if debugEnabled {
		realClient = realClient.Debug()
	}

	return context.WithValue(ctx, dbKey{}, &dbClient{Client: realClient}), nil
}
