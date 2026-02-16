package gala

import (
	"context"
	"testing"
	"time"

	"github.com/theopenlane/utils/testutils"
)

const (
	// defaultTestImage is the default docker image for test databases
	defaultTestImage = "docker://postgres:17-alpine"
	// defaultTestQueueName is the default queue name for gala tests
	defaultTestQueueName = "gala_test"
	// defaultTestWorkerCount is the default worker count for gala tests
	defaultTestWorkerCount = 5
	// defaultTestMaxRetries is the default max retries for gala tests
	defaultTestMaxRetries = 3
	// defaultExpiryMinutes is the default container expiry for test databases
	defaultExpiryMinutes = 5
)

// TestGalaFixture provides a fully initialized Gala instance for testing
type TestGalaFixture struct {
	// Gala is the initialized gala instance
	Gala *Gala
	// ConnectionURI is the database connection URI
	ConnectionURI string
	// cleanup functions to run on Close
	cleanupFuncs []func()
}

// TestGalaOption configures test gala creation
type TestGalaOption func(*testGalaConfig)

type testGalaConfig struct {
	image             string
	queueName         string
	workerCount       int
	maxRetries        int
	expiryMins        int
	startWorkers      bool
	fetchCooldown     time.Duration
	fetchPollInterval time.Duration
}

// WithTestImage sets the docker image for the test database
func WithTestImage(image string) TestGalaOption {
	return func(c *testGalaConfig) {
		c.image = image
	}
}

// WithTestQueueName sets the queue name for gala tests
func WithTestQueueName(name string) TestGalaOption {
	return func(c *testGalaConfig) {
		c.queueName = name
	}
}

// WithTestWorkerCount sets the worker count for gala tests
func WithTestWorkerCount(count int) TestGalaOption {
	return func(c *testGalaConfig) {
		c.workerCount = count
	}
}

// WithTestMaxRetries sets the max retries for gala tests
func WithTestMaxRetries(retries int) TestGalaOption {
	return func(c *testGalaConfig) {
		c.maxRetries = retries
	}
}

// WithTestExpiryMinutes sets the container expiry for test databases
func WithTestExpiryMinutes(mins int) TestGalaOption {
	return func(c *testGalaConfig) {
		c.expiryMins = mins
	}
}

// WithTestStartWorkers controls whether workers are started automatically
func WithTestStartWorkers(start bool) TestGalaOption {
	return func(c *testGalaConfig) {
		c.startWorkers = start
	}
}

// WithTestFetchCooldown sets the minimum time between job fetches (default 100ms)
func WithTestFetchCooldown(d time.Duration) TestGalaOption {
	return func(c *testGalaConfig) {
		c.fetchCooldown = d
	}
}

// WithTestFetchPollInterval sets the fallback polling interval (default 1s)
func WithTestFetchPollInterval(d time.Duration) TestGalaOption {
	return func(c *testGalaConfig) {
		c.fetchPollInterval = d
	}
}

// NewTestGala creates a fully initialized Gala instance backed by a real PostgreSQL database
// for integration testing. The database container is automatically cleaned up when the test completes.
func NewTestGala(t *testing.T, opts ...TestGalaOption) *TestGalaFixture {
	t.Helper()

	cfg := &testGalaConfig{
		image:        defaultTestImage,
		queueName:    defaultTestQueueName,
		workerCount:  defaultTestWorkerCount,
		maxRetries:   defaultTestMaxRetries,
		expiryMins:   defaultExpiryMinutes,
		startWorkers: true,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	ctx := context.Background()

	dbFixture := testutils.GetTestURI(
		testutils.WithImage(cfg.image),
		testutils.WithExpiryMinutes(cfg.expiryMins),
	)

	galaApp, err := NewGala(ctx, Config{
		Enabled:           true,
		ConnectionURI:     dbFixture.URI,
		QueueName:         cfg.queueName,
		WorkerCount:       cfg.workerCount,
		MaxRetries:        cfg.maxRetries,
		RunMigrations:     true,
		FetchCooldown:     cfg.fetchCooldown,
		FetchPollInterval: cfg.fetchPollInterval,
	})
	if err != nil {
		t.Fatalf("failed to create gala: %v", err)
	}

	fixture := &TestGalaFixture{
		Gala:          galaApp,
		ConnectionURI: dbFixture.URI,
	}

	fixture.cleanupFuncs = append(fixture.cleanupFuncs, func() {
		_ = galaApp.Close()
	})

	if cfg.startWorkers {
		if err := galaApp.StartWorkers(ctx); err != nil {
			t.Fatalf("failed to start gala workers: %v", err)
		}

		fixture.cleanupFuncs = append(fixture.cleanupFuncs, func() {
			_ = galaApp.StopWorkers(context.Background())
		})
	}

	t.Cleanup(fixture.Close)

	return fixture
}

// Close cleans up the test fixture
func (f *TestGalaFixture) Close() {
	for i := len(f.cleanupFuncs) - 1; i >= 0; i-- {
		f.cleanupFuncs[i]()
	}
}

// Registry returns the gala registry for listener registration
func (f *TestGalaFixture) Registry() *Registry {
	return f.Gala.Registry()
}

// newTestGalaInMemory creates a gala instance with a mock dispatcher for unit tests
// that don't require database backing. Use this for fast unit tests.
func newTestGalaInMemory(t *testing.T, dispatcher Dispatcher) *Gala {
	t.Helper()

	g := &Gala{}
	if err := g.initialize(dispatcher); err != nil {
		t.Fatalf("failed to build gala: %v", err)
	}

	return g
}
