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
	"github.com/pressly/goose/v3"
	"github.com/theopenlane/entx"

	"go.uber.org/zap"

	migratedb "github.com/theopenlane/core/db"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/pkg/testutils"
)

const (
	// defaultDBTestImage is the default docker image to use for testing
	defaultDBTestImage = "docker://postgres:16-alpine"
)

type client struct {
	// config is the entdb configuration
	config *entx.Config
	// pc is the primary ent client
	pc *ent.Client
	// sc is the secondary ent client
	sc *ent.Client
	// logger holds the zap logger
	logger *zap.SugaredLogger
}

// NewMultiDriverDBClient returns a ent client with a primary and secondary, if configured, write database
func NewMultiDriverDBClient(ctx context.Context, c entx.Config, l *zap.SugaredLogger, opts []ent.Option) (*ent.Client, *entx.EntClientConfig, error) {
	client := &client{
		config: &c,
		logger: l,
	}

	dbOpts := []entx.DBOption{
		entx.WithLogger(l),
	}

	if c.MultiWrite {
		dbOpts = append(dbOpts, entx.WithSecondaryDB())
	}

	entConfig, err := entx.NewDBConfig(c, dbOpts...)
	if err != nil {
		return nil, nil, err
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
			client.logger.Errorf("failed running migrations", zap.Error(err))

			return nil, nil, err
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
				client.logger.Errorf("failed running migrations", zap.Error(err))

				return nil, nil, err
			}
		}

		// Create Multiwrite driver
		cOpts = []ent.Option{ent.Driver(&entx.MultiWriteDriver{Wp: drvPrimary, Ws: drvSecondary})}
	} else {
		cOpts = []ent.Option{ent.Driver(drvPrimary)}
	}

	cOpts = append(cOpts, opts...)

	if c.Debug {
		cOpts = append(cOpts,
			ent.Log(client.logger.Named("ent").Debugln),
			ent.Debug(),
			ent.Driver(drvPrimary),
		)
	}

	ec := ent.NewClient(cOpts...)

	// add authz hooks
	ec.WithAuthz()

	if c.EnableHistory {
		// add history hooks
		ec.WithHistory()
	}

	ec.Intercept(interceptors.QueryLogger(client.logger))

	return ec, entConfig, nil
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
	if driver == dialect.SQLite {
		migrations = migratedb.GooseMigrationsSQLite

		if _, err := drv.Exec("PRAGMA foreign_keys = off;", nil); err != nil {
			drv.Close()

			return fmt.Errorf("failed to disable foreign keys: %w", err)
		}
	}

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect(driver); err != nil {
		return err
	}

	migrationsDir := "migrations-goose-postgres"
	if driver == dialect.SQLite {
		migrationsDir = "migrations-goose-sqlite"
	}

	if err := goose.Up(drv, migrationsDir); err != nil {
		return err
	}

	if driver == dialect.SQLite {
		if _, err := drv.Exec("PRAGMA foreign_keys = on;", nil); err != nil {
			drv.Close()

			return fmt.Errorf("failed to enable foreign keys: %w", err)
		}
	}

	return nil
}

// runAtlasMigrations runs the atlas auto-migrations
// this do not use the generated versioned migrations files from ent
func (c *client) runAtlasMigrations(ctx context.Context) error {
	// Run the automatic migration tool to create all schema resources.
	// entcache.Driver will skip the caching layer when running the schema migration
	if err := c.pc.Schema.Create(entcache.Skip(ctx)); err != nil {
		c.logger.Errorf("failed creating schema resources", zap.Error(err))

		return err
	}

	return nil
}

// createEntDBClient creates a new ent client with configured options
func (c *client) createEntDBClient(db *entsql.Driver) *ent.Client {
	cOpts := []ent.Option{ent.Driver(db)}

	if c.config.Debug {
		cOpts = append(cOpts,
			ent.Log(c.logger.Named("ent").Debugln),
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

	return testutils.GetTestURI(testDBURI, expiry)
}

// NewTestClient creates a entdb client that can be used for TEST purposes ONLY
func NewTestClient(ctx context.Context, ctr *testutils.TestFixture, entOpts []ent.Option) (*ent.Client, error) {
	// setup logger
	logger := zap.NewNop().Sugar()

	dbconf := entx.Config{
		Debug:           true,
		DriverName:      ctr.Dialect,
		PrimaryDBSource: ctr.URI,
		EnableHistory:   true,            // enable history so the code path is checked during unit tests
		CacheTTL:        0 * time.Second, // do not cache results in tests
	}

	entOpts = append(entOpts, ent.Logger(*logger))

	var db *ent.Client

	// Retry the connection to the database to ensure it is up and running
	var err error

	// If a test container is used, retry the connection to the database to ensure it is up and running
	if ctr.Pool != nil {
		err = ctr.Pool.Retry(func() error {
			fmt.Println("connecting to database...")

			db, _, err = NewMultiDriverDBClient(ctx, dbconf, logger, entOpts)
			if err != nil {
				fmt.Printf("retrying connection to database: %v", err)
			}

			return err
		})
	} else {
		db, _, err = NewMultiDriverDBClient(ctx, dbconf, logger, entOpts)
	}

	if err != nil {
		return nil, err
	}

	if err := db.Schema.Create(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
