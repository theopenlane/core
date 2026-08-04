package serveropts

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/tokens"
)

// errArchiveUnavailable stands in for redis being unreachable
var errArchiveUnavailable = errors.New("archive unavailable")

// startWatcher wires a token manager from dir and starts the watcher against it, the same
// order cmd/serve.go applies the options in
func startWatcher(t *testing.T, dir string, base tokens.Config) *ServerOptions {
	t.Helper()

	conf, err := tokenConfigFromKeyDir(base, dir)
	require.NoError(t, err)

	tm, err := tokens.New(conf)
	require.NoError(t, err)

	so := newServerOptions()
	so.Config.Settings.Auth.Token = conf
	so.Config.Handler.TokenManager = tm

	WithKeyDirWatcher(dir, base, testArchive(t)).apply(so)

	return so
}

func TestKeyDirWatcherDetectsNewKey(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeKey(t, dir, "first")

	so := startWatcher(t, dir, baseConfig())

	require.Eventually(t, func() bool {
		jwks, err := so.Config.Handler.TokenManager.Keys()
		return err == nil && jwks.Len() == 1
	}, 5*time.Second, 100*time.Millisecond)

	writeKey(t, dir, "second")

	require.Eventually(t, func() bool {
		jwks, err := so.Config.Handler.TokenManager.Keys()
		return err == nil && jwks.Len() == 2
	}, 5*time.Second, 100*time.Millisecond)
}

// TestKeyDirWatcherKeepsTokenManagerPointer covers the bug that made rotations invisible to
// the ent client and middleware, which capture the manager pointer once at startup
func TestKeyDirWatcherKeepsTokenManagerPointer(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	oldKID := writeKeyWithCert(t, dir, "primary", now.Add(-30*day), 90*day)

	so := startWatcher(t, dir, baseConfig())

	captured := so.Config.Handler.TokenManager
	require.Equal(t, oldKID, captured.CurrentKeyID())

	issued, _, err := captured.CreateTokenPair(testClaims())
	require.NoError(t, err)

	newKID := writeKeyWithCert(t, dir, "primary", now.Add(-2*time.Hour), 90*day)
	require.NotEqual(t, oldKID, newKID)

	require.Eventually(t, func() bool {
		return captured.CurrentKeyID() == newKID
	}, 10*time.Second, 100*time.Millisecond)

	require.Same(t, captured, so.Config.Handler.TokenManager)

	// the pointer the caller held signs with the new key and still verifies the old one
	fresh, _, err := captured.CreateTokenPair(testClaims())
	require.NoError(t, err)
	require.Equal(t, newKID, tokenKID(t, fresh))

	_, err = captured.Verify(issued)
	require.NoError(t, err)
}

// TestKeyDirWatcherSurvivesConfiguredKeyRemoval pins the fix for wedged rotation:
// configured keys resolve once at startup, so deleting the file afterwards can neither
// abort reloads nor drop the historical kid from verification
func TestKeyDirWatcherSurvivesConfiguredKeyRemoval(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	legacyDir := t.TempDir()
	now := time.Now()

	_, legacyPath := writeKey(t, legacyDir, "legacy")
	writeKeyWithCert(t, dir, "primary", now.Add(-30*day), 90*day)

	base := baseConfig()
	base.Keys = map[string]string{"historical-kid": legacyPath}

	so := startWatcher(t, dir, base)
	captured := so.Config.Handler.TokenManager

	require.NoError(t, os.Remove(legacyPath))

	newKID := writeKeyWithCert(t, dir, "secondary", now.Add(-2*time.Hour), 120*day)

	require.Eventually(t, func() bool {
		return captured.CurrentKeyID() == newKID
	}, 10*time.Second, 100*time.Millisecond)

	jwks, err := captured.Keys()
	require.NoError(t, err)

	_, ok := jwks.LookupKeyID("historical-kid")
	require.True(t, ok)
}

// gatedArchive fails every call until released, standing in for redis coming back after
// an outage
type gatedArchive struct {
	inner keyArchive
	up    atomic.Bool
}

func (g *gatedArchive) Record(ctx context.Context, keys map[string]crypto.PublicKey) error {
	if !g.up.Load() {
		return errArchiveUnavailable
	}

	return g.inner.Record(ctx, keys)
}

func (g *gatedArchive) Load(ctx context.Context) (map[string]crypto.PublicKey, error) {
	if !g.up.Load() {
		return nil, errArchiveUnavailable
	}

	return g.inner.Load(ctx)
}

// TestKeyDirWatcherRetriesFailedReload pins the retry backoff: with no fs events at all,
// a reload that failed against an unavailable archive is retried and applied well before
// the hourly re-evaluation
func TestKeyDirWatcherRetriesFailedReload(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ctx := context.Background()

	writeKey(t, dir, "primary")

	// a retired key only the archive knows about; it appears in the manager only after a
	// successful reload
	retiredPub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	retiredKID, err := thumbprintKID(retiredPub)
	require.NoError(t, err)

	inner := testArchive(t)
	require.NoError(t, inner.Record(ctx, map[string]crypto.PublicKey{retiredKID: retiredPub}))

	archive := &gatedArchive{inner: inner}

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	require.NoError(t, err)

	tm, err := tokens.New(conf)
	require.NoError(t, err)

	so := newServerOptions()
	so.Config.Settings.Auth.Token = conf
	so.Config.Handler.TokenManager = tm

	// the initial reload fails while the archive is down
	WithKeyDirWatcher(dir, baseConfig(), archive).apply(so)

	jwks, err := tm.Keys()
	require.NoError(t, err)

	_, ok := jwks.LookupKeyID(retiredKID)
	require.False(t, ok)

	archive.up.Store(true)

	require.Eventually(t, func() bool {
		jwks, err := tm.Keys()
		if err != nil {
			return false
		}

		_, ok := jwks.LookupKeyID(retiredKID)

		return ok
	}, 30*time.Second, 250*time.Millisecond)
}
