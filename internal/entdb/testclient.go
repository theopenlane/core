package entdb

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/theopenlane/utils/testutils"
)

const (
	// defaultDBTestImage is the default docker image to use for testing
	defaultDBTestImage = "docker://postgres:17-alpine"
)

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

	if err := db.Schema.Create(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
