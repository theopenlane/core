package entdb

import (
	"context"
	"database/sql"

	"ariga.io/entcache"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	migratedb "github.com/theopenlane/core/db"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"

	_ "github.com/jackc/pgx/v5/stdlib" // add pgx driver
)

type client struct {
	// config is the entdb configuration
	config *entx.Config
	// pc is the primary ent client
	pc *ent.Client
	// sc is the secondary ent client
	sc *ent.Client
}

// New returns a ent client with a primary and secondary, if configured, write database
func New(ctx context.Context, c entx.Config, jobOpts []riverqueue.Option, opts ...ent.Option) (*ent.Client, error) {
	client := &client{
		config: &c,
	}

	dbOpts := []entx.DBOption{}

	if c.MultiWrite {
		dbOpts = append(dbOpts, entx.WithSecondaryDB())
	}

	entConfig, err := entx.NewDBConfig(c, dbOpts...)
	if err != nil {
		return nil, err
	}

	// Decorates the sql.Driver with entcache.Driver on the primaryDB
	drvPrimary := entcache.NewDriver(
		entConfig.GetPrimaryDB(),
		entcache.TTL(c.CacheTTL), // set the TTL on the cache
		entcache.ContextLevel(),
	)

	client.pc = client.createEntDBClient(entConfig.GetPrimaryDB())

	if c.RunMigrations {
		if err := client.runMigrations(ctx); err != nil {
			log.Error().Err(err).Msg("failed running migrations")

			return nil, err
		}
	}

	var cOpts []ent.Option

	if c.MultiWrite {
		// Decorates the sql.Driver with entcache.Driver on the primaryDB
		drvSecondary := entcache.NewDriver(
			entConfig.GetSecondaryDB(),
			entcache.TTL(c.CacheTTL), // set the TTL on the cache
			entcache.ContextLevel(),
		)

		client.sc = client.createEntDBClient(entConfig.GetSecondaryDB())

		if c.RunMigrations {
			if err := client.runMigrations(ctx); err != nil {
				log.Error().Err(err).Msg("failed running migrations")

				return nil, err
			}
		}

		// Create Multiwrite driver
		cOpts = []ent.Option{ent.Driver(&entx.MultiWriteDriver{Wp: drvPrimary, Ws: drvSecondary})}
	} else {
		cOpts = []ent.Option{ent.Driver(drvPrimary)}
	}

	cOpts = append(cOpts, opts...)

	// add job client to the config
	cOpts = append(cOpts, ent.Job(ctx, jobOpts...))

	if c.Debug {
		cOpts = append(cOpts,
			ent.Log(log.Print),
			ent.Debug(),
			ent.Driver(drvPrimary),
		)
	}

	db := ent.NewClient(cOpts...)

	db.Config = entConfig

	// add authz hooks
	db.WithAuthz()

	// add job client to the client
	db.WithJobClient()

	if c.EnableHistory {
		// add history hooks
		db.WithHistory()
	}

	db.Intercept(interceptors.QueryLogger())

	// Register the global hooks
	pool := hooks.InitEventPool(db)
	hooks.RegisterGlobalHooks(db, pool)
	hooks.RegisterListeners(pool)

	return db, nil
}

// runMigrations runs the migrations based on the configured migration provider on startup
func (c *client) runMigrations(ctx context.Context) error {
	switch c.config.MigrationProvider {
	case "goose":
		return c.runGooseMigrations()
	default: // atlas
		return c.runAtlasMigrations(ctx)
	}
}

// runGooseMigrations runs the goose migrations
func (c *client) runGooseMigrations() error {
	driver, err := entx.CheckEntDialect(c.config.DriverName)
	if err != nil {
		return err
	}

	drv, err := sql.Open(c.config.DriverName, c.config.PrimaryDBSource)
	if err != nil {
		return err
	}
	defer drv.Close()

	migrations := migratedb.GooseMigrationsPG

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect(driver); err != nil {
		return err
	}

	migrationsDir := "migrations-goose-postgres"

	if err := goose.Up(drv, migrationsDir); err != nil {
		return err
	}

	return nil
}

// runAtlasMigrations runs the atlas auto-migrations
// this do not use the generated versioned migrations files from ent
func (c *client) runAtlasMigrations(ctx context.Context) error {
	// Run the automatic migration tool to create all schema resources.
	// entcache.Driver will skip the caching layer when running the schema migration
	if err := c.pc.Schema.Create(entcache.Skip(ctx)); err != nil {
		log.Error().Err(err).Msg("failed creating schema resources")

		return err
	}

	return nil
}

// createEntDBClient creates a new ent client with configured options
func (c *client) createEntDBClient(db *entsql.Driver) *ent.Client {
	cOpts := []ent.Option{ent.Driver(db)}

	if c.config.Debug {
		cOpts = append(cOpts,
			ent.Log(log.Print),
			ent.Debug(),
		)
	}

	return ent.NewClient(cOpts...)
}
