//go:build ignore

package main

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	// supported ent database drivers
	_ "github.com/lib/pq" // postgres driver

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"

	atlas "ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"github.com/theopenlane/core/internal/ent/generated/migrate"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/utils/testutils"
)

func main() {
	ctx := context.Background()

	// Create a local migration directory able to understand Atlas migration file format for replay.
	atlasDir, err := atlas.NewLocalDir("migrations")
	if err != nil {
		log.Fatal().Err(err).Msg("failed creating atlas migration directory")
	}

	gooseDirPG, err := sqltool.NewGooseDir("migrations-goose-postgres")
	if err != nil {
		log.Fatal().Err(err).Msg("failed creating goose migration directory")
	}

	// Migrate diff options.
	baseOpts := []schema.MigrateOption{
		schema.WithMigrationMode(schema.ModeReplay), // provide migration mode
		schema.WithDropColumn(true),
		schema.WithDropIndex(true),
	}

	postgresOpts := append(baseOpts, schema.WithDialect(dialect.Postgres))

	if len(os.Args) != 2 {
		log.Fatal().Msg("migration name is required. Use: 'go run -mod=mod create_migrations.go <name>'")
	}

	pgDBURI, ok := os.LookupEnv("ATLAS_POSTGRES_DB_URI")
	if !ok {
		log.Fatal().Msg("failed to load the ATLAS_POSTGRES_DB_URI env var")
	}

	maxConnections := 10

	tf, err := testutils.GetPostgresDockerTest(pgDBURI, 5*time.Minute, maxConnections)
	if err != nil {
		log.Fatal().Err(err).Msg("failed creating postgres test container")
	}

	defer testutils.TeardownFixture(tf)

	log.Info().Msgf("postgres test container started on %s", tf.URI)

	db, err := sql.Open("postgres", tf.URI)
	if err != nil {
		log.Fatal().Err(err).Msg("failed opening postgres connection")
	}

	// Generate migrations using Atlas support for postgres (note the Ent dialect option passed above).
	atlasOpts := append(baseOpts,
		schema.WithDialect(dialect.Postgres),
		schema.WithDir(atlasDir),
		schema.WithFormatter(atlas.DefaultFormatter),
	)

	if err := entdb.EnablePostgresExtensions(db); err != nil {
		log.Fatal().Err(err).Msg("failed enabling citext extension")
	}

	if err := migrate.NamedDiff(ctx, tf.URI, os.Args[1], atlasOpts...); err != nil {
		log.Fatal().Err(err).Msg("failed generating atlas migration file")
	}

	// Generate migrations using Goose support for postgres
	gooseOptsPG := append(postgresOpts, schema.WithDir(gooseDirPG))

	if err := entdb.EnablePostgresExtensions(db); err != nil {
		log.Fatal().Err(err).Msg("failed enabling citext extension")
	}

	if err = migrate.NamedDiff(ctx, tf.URI, os.Args[1], gooseOptsPG...); err != nil {
		log.Fatal().Err(err).Msg("failed generating goose migration file for postgres")
	}
}
