package serveropts

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"github.com/theopenlane/core/v2/config"
	serverconfig "github.com/theopenlane/core/v2/internal/httpserve/config"
	"github.com/theopenlane/iam/tokens"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

const day = 24 * time.Hour

// baseConfig mirrors the deployed token durations so signingMargin is ~7d in tests
func baseConfig() tokens.Config {
	return tokens.Config{
		Audience:                            "http://example.com",
		Issuer:                              "http://example.com",
		AccessDuration:                      time.Hour,
		RefreshDuration:                     2 * time.Hour,
		RefreshOverlap:                      -15 * time.Minute,
		AssessmentAccessDuration:            168 * time.Hour,
		TrustCenterNDARequestAccessDuration: 168 * time.Hour,
	}
}

// writeKey writes a bare ed25519 key PEM and returns its expected kid and path
func writeKey(t *testing.T, dir, name string) (string, string) {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	assert.NilError(t, err)

	path := filepath.Join(dir, name+pemExtension)

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	assert.NilError(t, err)

	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	assert.NilError(t, os.WriteFile(path, privatePEM, 0o600))

	kid, err := thumbprintKID(publicKey)
	assert.NilError(t, err)

	return kid, path
}

// writeKeyWithCert writes an RSA key PEM plus a self-signed cert and returns its kid
func writeKeyWithCert(t *testing.T, dir, name string, notBefore time.Time, lifetime time.Duration) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NilError(t, err)

	return writeSignerWithCert(t, dir, name, notBefore, lifetime, privateKey)
}

// writeSignerWithCert writes any signer's key PEM plus a self-signed cert and returns its kid
func writeSignerWithCert(t *testing.T, dir, name string, notBefore time.Time, lifetime time.Duration, privateKey crypto.Signer) string {
	t.Helper()

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	assert.NilError(t, err)

	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	assert.NilError(t, os.WriteFile(filepath.Join(dir, name+pemExtension), privatePEM, 0o600))

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		NotBefore:    notBefore,
		NotAfter:     notBefore.Add(lifetime),
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, privateKey.Public(), privateKey)
	assert.NilError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	assert.NilError(t, os.WriteFile(filepath.Join(dir, name+certExtension), certPEM, 0o600))

	kid, err := thumbprintKID(privateKey.Public())
	assert.NilError(t, err)

	return kid
}

func newServerOptions() *ServerOptions {
	return &ServerOptions{
		Config: serverconfig.Config{
			Settings: config.Config{
				Auth: config.Auth{
					Token: baseConfig(),
				},
			},
		},
	}
}

func tokenKID(t *testing.T, tks string) string {
	t.Helper()

	token, _, err := jwt.NewParser().ParseUnverified(tks, &jwt.RegisteredClaims{})
	assert.NilError(t, err)

	kid, ok := token.Header["kid"].(string)
	assert.Assert(t, ok)

	return kid
}

func testClaims() *tokens.Claims {
	return &tokens.Claims{Subject: "user"}
}

func TestKeyDirThumbprintKIDs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	kid, path := writeKey(t, dir, "anything")

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	assert.NilError(t, err)

	// the kid comes from the key material, never the filename
	assert.Assert(t, is.DeepEqual(map[string]string{kid: path}, conf.Keys))
	assert.Assert(t, is.Equal(kid, conf.KID))
}

func TestKeyDirEmpty(t *testing.T) {
	t.Parallel()

	_, err := tokenConfigFromKeyDir(baseConfig(), t.TempDir())
	assert.ErrorIs(t, err, ErrNoSigningKeys)
}

func TestKeyDirConfiguredKeysNeverSign(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	legacyDir := t.TempDir()

	_, legacyPath := writeKey(t, legacyDir, "legacy")
	certKID := writeKeyWithCert(t, dir, "primary", time.Now().Add(-2*time.Hour), 90*day)

	base := baseConfig()
	base.Keys = map[string]string{"historical-kid": legacyPath}

	conf, err := tokenConfigFromKeyDir(base, dir)
	assert.NilError(t, err)

	assert.Assert(t, is.Equal(legacyPath, conf.Keys["historical-kid"]))
	assert.Assert(t, is.Contains(conf.Keys, certKID))
	assert.Assert(t, is.Equal(certKID, conf.KID))
}

// TestKeyDirNeverPromotesLegacyULIDKey covers the failure this selection logic exists to
// prevent: the token manager's own fallback prefers ULID kids, so leaving selection to it
// would promote a key kept purely for verification back into signing
func TestKeyDirNeverPromotesLegacyULIDKey(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	legacyDir := t.TempDir()
	now := time.Now()

	_, legacyPath := writeKey(t, legacyDir, "legacy")
	legacyULID := ulid.Make().String()

	base := baseConfig()
	base.Keys = map[string]string{legacyULID: legacyPath}

	// well past the point where any key has enough headroom, which is exactly when the
	// old code returned an empty kid and handed selection back to the token manager
	expiringKID := writeKeyWithCert(t, dir, "primary", now.Add(-89*day), 90*day)

	conf, err := tokenConfigFromKeyDir(base, dir)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(expiringKID, conf.KID))

	tm, err := tokens.New(conf)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(expiringKID, tm.CurrentKeyID()))

	atks, _, err := tm.CreateTokenPair(testClaims())
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(expiringKID, tokenKID(t, atks)))
	assert.Assert(t, legacyULID != tokenKID(t, atks))
}

func TestKeyDirConfiguredKIDReplacedByDiskSelection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	kid, _ := writeKey(t, dir, "ondisk")

	base := baseConfig()
	base.KID = "does-not-exist"

	conf, err := tokenConfigFromKeyDir(base, dir)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(kid, conf.KID))
}

func TestKeyDirSelectsMostHeadroom(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	primaryKID := writeKeyWithCert(t, dir, "primary", now.Add(-60*day), 90*day)
	secondaryKID := writeKeyWithCert(t, dir, "secondary", now.Add(-30*day), 90*day)

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	assert.NilError(t, err)

	assert.Assert(t, is.Contains(conf.Keys, primaryKID))
	assert.Assert(t, is.Contains(conf.Keys, secondaryKID))
	assert.Assert(t, is.Equal(secondaryKID, conf.KID))

	tm, err := tokens.New(conf)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(secondaryKID, tm.CurrentKeyID()))
}

func TestKeyDirSkipsCertExpiringInsideMargin(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	primaryKID := writeKeyWithCert(t, dir, "primary", now.Add(-50*day), 90*day)
	// expires in a day, inside the ~7d margin, so it cannot outlive tokens it would sign
	writeKeyWithCert(t, dir, "secondary", now.Add(-62*day), 63*day)

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(primaryKID, conf.KID))
}

func TestKeyDirSignsWithBestAvailableWhenNoneQualify(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	nearest := writeKeyWithCert(t, dir, "primary", now.Add(-89*day), 90*day)

	entries, err := scanKeyDir(dir)
	assert.NilError(t, err)

	kid, healthy := chooseSigningKID(entries, signingMargin(baseConfig()))
	assert.Assert(t, is.Equal(nearest, kid))
	assert.Assert(t, !healthy)
}

func TestKeyDirNewCertWaitsForPropagation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	// secondary has more headroom but is minutes old, so primary keeps signing until the
	// new key's JWKS entry has had time to reach validator caches
	primaryKID := writeKeyWithCert(t, dir, "primary", now.Add(-30*day), 90*day)
	writeKeyWithCert(t, dir, "secondary", now.Add(-time.Minute), 120*day)

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(primaryKID, conf.KID))
}

func TestKeyDirMixedKeyTypes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	_, edPriv, err := ed25519.GenerateKey(rand.Reader)
	assert.NilError(t, err)

	edKID := writeSignerWithCert(t, dir, "primary", now.Add(-24*time.Hour), 90*day, edPriv)

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	assert.NilError(t, err)

	tm, err := tokens.New(conf)
	assert.NilError(t, err)

	edToken, _, err := tm.CreateTokenPair(testClaims())
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(edKID, tokenKID(t, edToken)))

	rsaKID := writeKeyWithCert(t, dir, "secondary", now.Add(-2*time.Hour), 120*day)

	conf, err = tokenConfigFromKeyDir(baseConfig(), dir)
	assert.NilError(t, err)

	mixed, err := tokens.New(conf)
	assert.NilError(t, err)

	rsaToken, _, err := mixed.CreateTokenPair(testClaims())
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(rsaKID, tokenKID(t, rsaToken)))

	for _, tks := range []string{edToken, rsaToken} {
		verified, err := mixed.Verify(tks)
		assert.NilError(t, err)
		assert.Assert(t, is.Equal("user", verified.Subject))
	}

	jwks, err := mixed.Keys()
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(2, jwks.Len()))
}

func TestKeySetFromKeyDirConfiguredKeysVerifyOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	legacyDir := t.TempDir()

	legacyKID, legacyPath := writeKey(t, legacyDir, "legacy")
	primaryKID := writeKeyWithCert(t, dir, "primary", time.Now().Add(-2*time.Hour), 90*day)

	base := baseConfig()
	base.Keys = map[string]string{legacyKID: legacyPath}

	configured, err := configuredVerificationKeys(base.Keys)
	assert.NilError(t, err)

	keySet, healthy, err := keySetFromKeyDir(context.Background(), base, dir, configured, noopKeyArchive{})
	assert.NilError(t, err)
	assert.Assert(t, healthy)

	assert.Assert(t, is.Equal(primaryKID, keySet.KID))
	assert.Assert(t, is.Contains(keySet.Signing, primaryKID))
	_, signsWithLegacy := keySet.Signing[legacyKID]
	assert.Assert(t, !signsWithLegacy)
	assert.Assert(t, is.Contains(keySet.Verification, legacyKID))
}

// TestKeyDirLegacyULIDFilenameKeptForVerification pins the upgrade path: tokens signed
// under the old scheme carry the PEM's ULID filename as their kid, so the same file must
// keep verifying under that kid after kids become thumbprint derived
func TestKeyDirLegacyULIDFilenameKeptForVerification(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	legacyULID := ulid.Make().String()
	_, legacyPath := writeKey(t, dir, legacyULID)

	// sign the way the old deployment did: the filename is the kid
	old := baseConfig()
	old.Keys = map[string]string{legacyULID: legacyPath}

	oldTM, err := tokens.New(old)
	assert.NilError(t, err)

	issued, _, err := oldTM.CreateTokenPair(testClaims())
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(legacyULID, tokenKID(t, issued)))

	// the new cert-backed pair takes over signing; the legacy file stays in the dir
	primaryKID := writeKeyWithCert(t, dir, "primary", now.Add(-2*time.Hour), 90*day)

	keySet, healthy, err := keySetFromKeyDir(context.Background(), baseConfig(), dir, nil, noopKeyArchive{})
	assert.NilError(t, err)
	assert.Assert(t, healthy)
	assert.Assert(t, is.Equal(primaryKID, keySet.KID))
	assert.Assert(t, is.Contains(keySet.Verification, legacyULID))

	tm, err := tokens.NewWithKey(keySet.Signing[keySet.KID], baseConfig())
	assert.NilError(t, err)
	assert.NilError(t, tm.SetKeySet(keySet))

	claims, err := tm.Verify(issued)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal("user", claims.Subject))

	// the startup lane carries the legacy kid too, so tokens verify before the watcher runs
	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(primaryKID, conf.KID))
	assert.Assert(t, is.Equal(legacyPath, conf.Keys[legacyULID]))

	boot, err := tokens.New(conf)
	assert.NilError(t, err)

	_, err = boot.Verify(issued)
	assert.NilError(t, err)
}
