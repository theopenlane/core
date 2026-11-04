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

	"github.com/theopenlane/iam/tokens"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/poll"
)

// errArchiveUnavailable stands in for redis being unreachable
var errArchiveUnavailable = errors.New("archive unavailable")

// startWatcher wires a token manager from dir and starts the watcher against it, the same
// order cmd/serve.go applies the options in
func startWatcher(t *testing.T, dir string, base tokens.Config) *ServerOptions {
	t.Helper()

	conf, err := tokenConfigFromKeyDir(base, dir)
	assert.NilError(t, err)

	tm, err := tokens.New(conf)
	assert.NilError(t, err)

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

	poll.WaitOn(t, func(poll.LogT) poll.Result {
		jwks, err := so.Config.Handler.TokenManager.Keys()
		if err == nil && jwks.Len() == 1 {
			return poll.Success()
		}

		return poll.Continue("waiting for 1 key")
	}, poll.WithTimeout(5*time.Second), poll.WithDelay(100*time.Millisecond))

	writeKey(t, dir, "second")

	poll.WaitOn(t, func(poll.LogT) poll.Result {
		jwks, err := so.Config.Handler.TokenManager.Keys()
		if err == nil && jwks.Len() == 2 {
			return poll.Success()
		}

		return poll.Continue("waiting for 2 keys")
	}, poll.WithTimeout(5*time.Second), poll.WithDelay(100*time.Millisecond))
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
	assert.Assert(t, is.Equal(oldKID, captured.CurrentKeyID()))

	issued, _, err := captured.CreateTokenPair(testClaims())
	assert.NilError(t, err)

	newKID := writeKeyWithCert(t, dir, "primary", now.Add(-2*time.Hour), 90*day)
	assert.Assert(t, oldKID != newKID)

	poll.WaitOn(t, func(poll.LogT) poll.Result {
		return poll.Compare(is.Equal(captured.CurrentKeyID(), newKID))
	}, poll.WithTimeout(10*time.Second), poll.WithDelay(100*time.Millisecond))

	assert.Assert(t, is.Equal(captured, so.Config.Handler.TokenManager))

	// the pointer the caller held signs with the new key and still verifies the old one
	fresh, _, err := captured.CreateTokenPair(testClaims())
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(newKID, tokenKID(t, fresh)))

	_, err = captured.Verify(issued)
	assert.NilError(t, err)
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

	assert.NilError(t, os.Remove(legacyPath))

	newKID := writeKeyWithCert(t, dir, "secondary", now.Add(-2*time.Hour), 120*day)

	poll.WaitOn(t, func(poll.LogT) poll.Result {
		return poll.Compare(is.Equal(captured.CurrentKeyID(), newKID))
	}, poll.WithTimeout(10*time.Second), poll.WithDelay(100*time.Millisecond))

	jwks, err := captured.Keys()
	assert.NilError(t, err)

	_, ok := jwks.LookupKeyID("historical-kid")
	assert.Assert(t, ok)
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
	assert.NilError(t, err)

	retiredKID, err := thumbprintKID(retiredPub)
	assert.NilError(t, err)

	inner := testArchive(t)
	assert.NilError(t, inner.Record(ctx, map[string]crypto.PublicKey{retiredKID: retiredPub}))

	archive := &gatedArchive{inner: inner}

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	assert.NilError(t, err)

	tm, err := tokens.New(conf)
	assert.NilError(t, err)

	so := newServerOptions()
	so.Config.Settings.Auth.Token = conf
	so.Config.Handler.TokenManager = tm

	// the initial reload fails while the archive is down
	WithKeyDirWatcher(dir, baseConfig(), archive).apply(so)

	jwks, err := tm.Keys()
	assert.NilError(t, err)

	_, ok := jwks.LookupKeyID(retiredKID)
	assert.Assert(t, !ok)

	archive.up.Store(true)

	poll.WaitOn(t, func(poll.LogT) poll.Result {
		jwks, err := tm.Keys()
		if err != nil {
			return poll.Continue("keys unavailable: %v", err)
		}

		if _, ok := jwks.LookupKeyID(retiredKID); ok {
			return poll.Success()
		}

		return poll.Continue("retired key not yet verifiable")
	}, poll.WithTimeout(30*time.Second), poll.WithDelay(250*time.Millisecond))
}
