package dbtest

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"io"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/emailtemplates"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/enttest"
	utilstest "github.com/theopenlane/utils/testutils"
)

// fgaModelPath returns the absolute path to the OpenFGA model used in tests.
func fgaModelPath() string {
	_, filename, _, _ := runtime.Caller(0)
	base := filepath.Dir(filename)
	return filepath.Join(base, "..", "..", "..", "fga", "model", "model.fga")
}

// NewPostgresClient creates a new Postgres ent client backed by a testcontainer.
func NewPostgresClient(t *testing.T) *generated.Client {
	t.Helper()

	tf := utilstest.GetTestURI(utilstest.WithImage("docker://postgres:17-alpine"))
	t.Cleanup(func() { utilstest.TeardownFixture(tf) })

	db, err := sql.Open("postgres", tf.URI)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	fgaTF := fgatest.NewFGATestcontainer(context.Background(), fgatest.WithModelFile(fgaModelPath()))
	t.Cleanup(func() { _ = fgaTF.TeardownFixture() })

	fgaClient, err := fgaTF.NewFgaClient(context.Background())
	require.NoError(t, err)

	tm, err := createTokenManager(15 * time.Minute)
	require.NoError(t, err)
	sm := createSessionManager()
	rc := newRedisClient()

	sessionConfig := sessions.NewSessionConfig(
		sm,
		sessions.WithPersistence(rc),
	)
	sessionConfig.CookieConfig = sessions.DebugOnlyCookieConfig

	opts := []generated.Option{
		generated.Authz(*fgaClient),
		generated.TokenManager(tm),
		generated.SessionConfig(&sessionConfig),
		generated.Emailer(&emailtemplates.Config{}),
	}

	// enable required extensions before running migrations
	enableExtOption := schema.WithHooks(func(next schema.Creator) schema.Creator {
		return schema.CreateFunc(func(ctx context.Context, tables ...*schema.Table) error {
			if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;`); err != nil {
				return err
			}

			return next.Create(ctx, tables...)
		})
	})

	client := enttest.Open(t, dialect.Postgres, tf.URI,
		enttest.WithMigrateOptions(enableExtOption),
		enttest.WithOptions(opts...))

	client.WithAuthz()

	return client
}

func newRedisClient() *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:             mr.Addr(),
		DisableIndentity: true, // # spellcheck:off
	})

	return client
}

func createSessionManager() sessions.Store[map[string]any] {
	hashKey := randomString(32)
	blockKey := randomString(32)

	sm := sessions.NewCookieStore[map[string]any](sessions.DebugCookieConfig,
		hashKey, blockKey,
	)

	return sm
}

func randomString(n int) []byte {
	id := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		panic(err)
	}
	return id
}

func createTokenManager(refreshOverlap time.Duration) (*tokens.TokenManager, error) {
	if refreshOverlap >= 0 {
		refreshOverlap = -15 * time.Minute
	}

	conf := tokens.Config{
		Audience:        "http://localhost:17608",
		Issuer:          "http://localhost:17608",
		AccessDuration:  1 * time.Hour,
		RefreshDuration: 2 * time.Hour,
		RefreshOverlap:  refreshOverlap,
	}

	if -refreshOverlap > conf.AccessDuration {
		refreshOverlap = -(conf.AccessDuration - time.Second)
		conf.RefreshOverlap = refreshOverlap
	}

	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return tokens.NewWithKey(key, conf)
}
