package serveropts

import (
	"context"
	"crypto"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/theopenlane/iam/tokens"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func testArchive(t *testing.T) *redisKeyArchive {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})

	t.Cleanup(func() {
		assert.NilError(t, client.Close())
	})

	return newRedisKeyArchive(client, time.Hour)
}

func TestKeyArchiveRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	archive := testArchive(t)
	ctx := context.Background()

	kid, _ := writeKey(t, dir, "key")

	entries, err := scanKeyDir(dir)
	assert.NilError(t, err)

	assert.NilError(t, archive.Record(ctx, map[string]crypto.PublicKey{kid: entries[0].signer.Public()}))

	loaded, err := archive.Load(ctx)
	assert.NilError(t, err)
	assert.Assert(t, is.Contains(loaded, kid))
	assert.Assert(t, is.DeepEqual(entries[0].signer.Public(), loaded[kid]))
}

func TestKeyArchiveKeepsRotatedKeyVerifiable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	archive := testArchive(t)
	ctx := context.Background()
	now := time.Now()

	oldKID := writeKeyWithCert(t, dir, "primary", now.Add(-30*day), 90*day)

	keySet, _, err := keySetFromKeyDir(ctx, baseConfig(), dir, nil, archive)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(oldKID, keySet.KID))

	tm, err := tokens.NewWithKey(keySet.Signing[oldKID], baseConfig())
	assert.NilError(t, err)
	assert.NilError(t, tm.SetKeySet(keySet))

	issued, _, err := tm.CreateTokenPair(testClaims())
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(oldKID, tokenKID(t, issued)))

	// cert-manager replaces the key material at the same path, so the old private key is
	// gone from disk entirely
	newKID := writeKeyWithCert(t, dir, "primary", now.Add(-2*time.Hour), 90*day)
	assert.Assert(t, oldKID != newKID)

	rotated, _, err := keySetFromKeyDir(ctx, baseConfig(), dir, nil, archive)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(newKID, rotated.KID))

	_, signsWithOld := rotated.Signing[oldKID]
	assert.Assert(t, !signsWithOld)
	assert.Assert(t, is.Contains(rotated.Verification, oldKID))

	assert.NilError(t, tm.SetKeySet(rotated))

	claims, err := tm.Verify(issued)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal("user", claims.Subject))
}

// failingArchive stands in for redis being unreachable
type failingArchive struct{}

func (failingArchive) Record(_ context.Context, _ map[string]crypto.PublicKey) error {
	return errors.New("archive unavailable") //nolint:err113
}

func (failingArchive) Load(_ context.Context) (map[string]crypto.PublicKey, error) {
	return nil, errors.New("archive unavailable") //nolint:err113
}

// TestKeySetFromKeyDirFailsWhenArchiveUnavailable pins the deliberate tradeoff: applying a
// key set without the retired keys would stop verifying tokens still in flight, so a
// broken archive aborts the reload instead
func TestKeySetFromKeyDirFailsWhenArchiveUnavailable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeKeyWithCert(t, dir, "primary", time.Now().Add(-30*day), 90*day)

	_, _, err := keySetFromKeyDir(context.Background(), baseConfig(), dir, nil, failingArchive{})
	assert.Assert(t, err != nil)
}
