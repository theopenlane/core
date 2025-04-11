package entdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"ariga.io/entcache"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/utils/testutils"

	migratedb "github.com/theopenlane/core/db"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"

	_ "github.com/jackc/pgx/v5/stdlib" // add pgx driver
)

// postgresExtensions is a list of postgres extensions to enable
var postgresExtensions = []string{
	"citext",
}

const (
	// defaultDBTestImage is the default docker image to use for testing
	defaultDBTestImage = "docker://postgres:17-alpine"
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

	// setup driver(s)
	var drvPrimary dialect.Driver

	var drvSecondary dialect.Driver

	drvPrimary = entConfig.GetPrimaryDB()
	client.pc = client.createEntDBClient(entConfig.GetPrimaryDB())

	// run migrations on primary driver
	if c.RunMigrations {
		if err := client.runMigrations(ctx); err != nil {
			log.Error().Err(err).Msg("failed running migrations")

			return nil, err
		}
	}

	var cOpts []ent.Option

	// if multi-write is enabled, create a secondary client
	if c.MultiWrite {
		drvSecondary = entConfig.GetSecondaryDB()
		client.sc = client.createEntDBClient(entConfig.GetSecondaryDB())

		// run  migrations on secondary driver
		if c.RunMigrations {
			if err := client.runMigrations(ctx); err != nil {
				log.Error().Err(err).Msg("failed running migrations")

				return nil, err
			}
		}
	}

	// if cache TTL is set, wrap the driver with the cache driver
	// as of (2025-02-18) entcache needs to be enabled after migrations are run
	// if using atlas migrations due to an incompatibility in versions
	// even with entcache.Skip(ctx) set on atlas migrations
	if c.CacheTTL > 0 {
		drvPrimary = entcacheDriver(drvPrimary, c.CacheTTL)

		if drvSecondary != nil {
			drvSecondary = entcacheDriver(drvSecondary, c.CacheTTL)
		}
	}

	// add the option to the client for the drivers
	if drvSecondary != nil {
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

	// add event emission for mutations
	eventer := hooks.NewEventerPool(db)
	hooks.RegisterGlobalHooks(db, eventer)

	if err := hooks.RegisterListeners(eventer); err != nil {
		return nil, err
	}

	db.Use(hooks.MetricsHook())

	return db, nil
}

func entcacheDriver(driver dialect.Driver, cacheTTL time.Duration) *entcache.Driver {
	return entcache.NewDriver(
		driver,
		entcache.TTL(cacheTTL), // set the TTL on the cache
		entcache.ContextLevel(),
	)
}

// runMigrations runs the migrations based on the configured migration provider on startup
func (c *client) runMigrations(ctx context.Context) error {
	switch c.config.MigrationProvider {
	case "goose":
		if err := c.runGooseMigrations(); err != nil {
			return err
		}

		return c.seedData()
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

	// enable extensions
	if err := EnablePostgresExtensions(drv); err != nil {
		log.Error().Err(err).Msg("failed enabling citext extension")

		return err
	}

	if err := goose.Up(drv, migrationsDir); err != nil {
		log.Error().Err(err).Msg("failed running goose migrations")

		return err
	}

	return nil
}

// runAtlasMigrations runs the atlas auto-migrations
// this do not use the generated versioned migrations files from ent
func (c *client) runAtlasMigrations(ctx context.Context) error {
	// Run the automatic migration tool to create all schema resources.
	if err := c.pc.Schema.Create(ctx,
		EnablePostgresOption(c.pc.DB())); err != nil {
		log.Error().Err(err).Msg("failed creating schema resources")

		return err
	}

	return nil
}

// EnablePostgresExtensions enables the postgres extensions
// needed when running migrations
func EnablePostgresExtensions(db *sql.DB) error {
	for _, ext := range postgresExtensions {
		if _, err := db.Exec(fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS %s WITH SCHEMA public;`, ext)); err != nil {
			return fmt.Errorf("could not enable %s extension: %w", ext, err)
		}
	}

	return nil
}

// EnablePostgresOption returns a schema.MigrateOption
// that will enable the Postgres extension if needed for running atlas migrations
func EnablePostgresOption(db *sql.DB) schema.MigrateOption {
	return schema.WithHooks(func(next schema.Creator) schema.Creator {
		return schema.CreateFunc(func(ctx context.Context, tables ...*schema.Table) error {
			if err := EnablePostgresExtensions(db); err != nil {
				return err
			}

			return next.Create(ctx, tables...)
		})
	})
}

// seedData runs the data seed using goose
func (c *client) seedData() error {
	driver, err := entx.CheckEntDialect(c.config.DriverName)
	if err != nil {
		return err
	}

	drv, err := sql.Open(c.config.DriverName, c.config.PrimaryDBSource)
	if err != nil {
		return err
	}
	defer drv.Close()

	seeds := migratedb.SeedMigrationsPG

	goose.SetBaseFS(seeds)

	if err := goose.SetDialect(driver); err != nil {
		return err
	}

	migrationsDir := "seed"

	if err := goose.Up(drv, migrationsDir, goose.WithNoVersioning()); err != nil {
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

// NewTestFixture creates a test container for testing purposes
func NewTestFixture() *testutils.TestFixture {
	// Grab the DB environment variable or use the default
	testDBURI := os.Getenv("TEST_DB_URL")
	testDBContainerExpiry := os.Getenv("TEST_DB_CONTAINER_EXPIRY")

	// If the DB URI is not set, use the default docker image
	if testDBURI == "" {
		testDBURI = defaultDBTestImage
	}

	if testDBContainerExpiry == "" {
		testDBContainerExpiry = "5" // default expiry of 5 minutes
	}

	expiry, err := strconv.Atoi(testDBContainerExpiry)
	if err != nil {
		panic(fmt.Sprintf("failed to convert TEST_DB_CONTAINER_EXPIRY to int: %v", err))
	}

	return testutils.GetTestURI(testutils.WithImage(testDBURI),
		testutils.WithExpiryMinutes(expiry),
		testutils.WithMaxConn(200)) // nolint:mnd
}

// NewTestClient creates a entdb client that can be used for TEST purposes ONLY
func NewTestClient(ctx context.Context, ctr *testutils.TestFixture, jobOpts []riverqueue.Option, entOpts []ent.Option) (*ent.Client, error) {
	dbconf := entx.Config{
		Debug:           true,
		DriverName:      ctr.Dialect,
		PrimaryDBSource: ctr.URI,
		EnableHistory:   true,            // enable history so the code path is checked during unit tests
		CacheTTL:        0 * time.Second, // do not cache results in tests
	}

	// Create the db client
	var db *ent.Client

	// Retry the connection to the database to ensure it is up and running
	var err error

	// run migrations for tests
	jobOpts = append(jobOpts, riverqueue.WithRunMigrations(true))

	// If a test container is used, retry the connection to the database to ensure it is up and running
	if ctr.Pool != nil {
		err = ctr.Pool.Retry(func() error {
			log.Info().Msg("connecting to database...")

			db, err = New(ctx, dbconf, jobOpts, entOpts...)
			if err != nil {
				log.Info().Err(err).Msg("retrying connection to database...")
			}

			return err
		})
	} else {
		db, err = New(ctx, dbconf, jobOpts, entOpts...)
	}

	if err != nil {
		return nil, err
	}

	client := &client{
		config: &dbconf,
		pc:     db,
	}

	if err := client.runAtlasMigrations(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
