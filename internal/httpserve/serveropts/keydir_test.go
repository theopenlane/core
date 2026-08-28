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
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/v2/config"
	serverconfig "github.com/theopenlane/core/v2/internal/httpserve/config"
	"github.com/theopenlane/iam/tokens"
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
	require.NoError(t, err)

	path := filepath.Join(dir, name+pemExtension)

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	require.NoError(t, os.WriteFile(path, privatePEM, 0o600))

	kid, err := thumbprintKID(publicKey)
	require.NoError(t, err)

	return kid, path
}

// writeKeyWithCert writes an RSA key PEM plus a self-signed cert and returns its kid
func writeKeyWithCert(t *testing.T, dir, name string, notBefore time.Time, lifetime time.Duration) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return writeSignerWithCert(t, dir, name, notBefore, lifetime, privateKey)
}

// writeSignerWithCert writes any signer's key PEM plus a self-signed cert and returns its kid
func writeSignerWithCert(t *testing.T, dir, name string, notBefore time.Time, lifetime time.Duration, privateKey crypto.Signer) string {
	t.Helper()

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+pemExtension), privatePEM, 0o600))

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		NotBefore:    notBefore,
		NotAfter:     notBefore.Add(lifetime),
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, privateKey.Public(), privateKey)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+certExtension), certPEM, 0o600))

	kid, err := thumbprintKID(privateKey.Public())
	require.NoError(t, err)

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
	require.NoError(t, err)

	kid, ok := token.Header["kid"].(string)
	require.True(t, ok)

	return kid
}

func testClaims() *tokens.Claims {
	return &tokens.Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user"}}
}

func TestKeyDirThumbprintKIDs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	kid, path := writeKey(t, dir, "anything")

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	require.NoError(t, err)

	// the kid comes from the key material, never the filename
	require.Equal(t, map[string]string{kid: path}, conf.Keys)
	require.Equal(t, kid, conf.KID)
}

func TestKeyDirEmpty(t *testing.T) {
	t.Parallel()

	_, err := tokenConfigFromKeyDir(baseConfig(), t.TempDir())
	require.ErrorIs(t, err, ErrNoSigningKeys)
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
	require.NoError(t, err)

	require.Equal(t, legacyPath, conf.Keys["historical-kid"])
	require.Contains(t, conf.Keys, certKID)
	require.Equal(t, certKID, conf.KID)
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
	require.NoError(t, err)
	require.Equal(t, expiringKID, conf.KID)

	tm, err := tokens.New(conf)
	require.NoError(t, err)
	require.Equal(t, expiringKID, tm.CurrentKeyID())

	atks, _, err := tm.CreateTokenPair(testClaims())
	require.NoError(t, err)
	require.Equal(t, expiringKID, tokenKID(t, atks))
	require.NotEqual(t, legacyULID, tokenKID(t, atks))
}

func TestKeyDirConfiguredKIDReplacedByDiskSelection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	kid, _ := writeKey(t, dir, "ondisk")

	base := baseConfig()
	base.KID = "does-not-exist"

	conf, err := tokenConfigFromKeyDir(base, dir)
	require.NoError(t, err)
	require.Equal(t, kid, conf.KID)
}

func TestKeyDirSelectsMostHeadroom(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	primaryKID := writeKeyWithCert(t, dir, "primary", now.Add(-60*day), 90*day)
	secondaryKID := writeKeyWithCert(t, dir, "secondary", now.Add(-30*day), 90*day)

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	require.NoError(t, err)

	require.Contains(t, conf.Keys, primaryKID)
	require.Contains(t, conf.Keys, secondaryKID)
	require.Equal(t, secondaryKID, conf.KID)

	tm, err := tokens.New(conf)
	require.NoError(t, err)
	require.Equal(t, secondaryKID, tm.CurrentKeyID())
}

func TestKeyDirSkipsCertExpiringInsideMargin(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	primaryKID := writeKeyWithCert(t, dir, "primary", now.Add(-50*day), 90*day)
	// expires in a day, inside the ~7d margin, so it cannot outlive tokens it would sign
	writeKeyWithCert(t, dir, "secondary", now.Add(-62*day), 63*day)

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	require.NoError(t, err)
	require.Equal(t, primaryKID, conf.KID)
}

func TestKeyDirSignsWithBestAvailableWhenNoneQualify(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	nearest := writeKeyWithCert(t, dir, "primary", now.Add(-89*day), 90*day)

	entries, err := scanKeyDir(dir)
	require.NoError(t, err)

	kid, healthy := chooseSigningKID(entries, signingMargin(baseConfig()))
	require.Equal(t, nearest, kid)
	require.False(t, healthy)
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
	require.NoError(t, err)
	require.Equal(t, primaryKID, conf.KID)
}

func TestKeyDirMixedKeyTypes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Now()

	_, edPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	edKID := writeSignerWithCert(t, dir, "primary", now.Add(-24*time.Hour), 90*day, edPriv)

	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	require.NoError(t, err)

	tm, err := tokens.New(conf)
	require.NoError(t, err)

	edToken, _, err := tm.CreateTokenPair(testClaims())
	require.NoError(t, err)
	require.Equal(t, edKID, tokenKID(t, edToken))

	rsaKID := writeKeyWithCert(t, dir, "secondary", now.Add(-2*time.Hour), 120*day)

	conf, err = tokenConfigFromKeyDir(baseConfig(), dir)
	require.NoError(t, err)

	mixed, err := tokens.New(conf)
	require.NoError(t, err)

	rsaToken, _, err := mixed.CreateTokenPair(testClaims())
	require.NoError(t, err)
	require.Equal(t, rsaKID, tokenKID(t, rsaToken))

	for _, tks := range []string{edToken, rsaToken} {
		verified, err := mixed.Verify(tks)
		require.NoError(t, err)
		require.Equal(t, "user", verified.Subject)
	}

	jwks, err := mixed.Keys()
	require.NoError(t, err)
	require.Equal(t, 2, jwks.Len())
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
	require.NoError(t, err)

	keySet, healthy, err := keySetFromKeyDir(context.Background(), base, dir, configured, noopKeyArchive{})
	require.NoError(t, err)
	require.True(t, healthy)

	require.Equal(t, primaryKID, keySet.KID)
	require.Contains(t, keySet.Signing, primaryKID)
	require.NotContains(t, keySet.Signing, legacyKID)
	require.Contains(t, keySet.Verification, legacyKID)
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
	require.NoError(t, err)

	issued, _, err := oldTM.CreateTokenPair(testClaims())
	require.NoError(t, err)
	require.Equal(t, legacyULID, tokenKID(t, issued))

	// the new cert-backed pair takes over signing; the legacy file stays in the dir
	primaryKID := writeKeyWithCert(t, dir, "primary", now.Add(-2*time.Hour), 90*day)

	keySet, healthy, err := keySetFromKeyDir(context.Background(), baseConfig(), dir, nil, noopKeyArchive{})
	require.NoError(t, err)
	require.True(t, healthy)
	require.Equal(t, primaryKID, keySet.KID)
	require.Contains(t, keySet.Verification, legacyULID)

	tm, err := tokens.NewWithKey(keySet.Signing[keySet.KID], baseConfig())
	require.NoError(t, err)
	require.NoError(t, tm.SetKeySet(keySet))

	claims, err := tm.Verify(issued)
	require.NoError(t, err)
	require.Equal(t, "user", claims.Subject)

	// the startup lane carries the legacy kid too, so tokens verify before the watcher runs
	conf, err := tokenConfigFromKeyDir(baseConfig(), dir)
	require.NoError(t, err)
	require.Equal(t, primaryKID, conf.KID)
	require.Equal(t, legacyPath, conf.Keys[legacyULID])

	boot, err := tokens.New(conf)
	require.NoError(t, err)

	_, err = boot.Verify(issued)
	require.NoError(t, err)
}
