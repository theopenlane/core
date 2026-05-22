package serveropts

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/config"
	serverconfig "github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/iam/tokens"
)

func writeKey(t *testing.T, dir, kid string) string {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	path := filepath.Join(dir, kid+".pem")

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	publicDER, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)

	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	data := append(privatePEM, publicPEM...)

	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

// writeCertSlot writes a cert-manager style tls.key and tls.crt pair for a slot
func writeCertSlot(t *testing.T, dir, slot string, notAfter time.Time) (keyPath, certPath, kid string) {
	t.Helper()

	slotDir := filepath.Join(dir, slot)
	require.NoError(t, os.MkdirAll(slotDir, 0o700))

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	keyPath = filepath.Join(slotDir, "tls.key")
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})
	require.NoError(t, os.WriteFile(keyPath, privatePEM, 0o600))

	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	require.NoError(t, err)

	cert := &x509.Certificate{
		SerialNumber:          serial,
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, cert, cert, publicKey, privateKey)
	require.NoError(t, err)

	certPath = filepath.Join(slotDir, "tls.crt")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	require.NoError(t, os.WriteFile(certPath, certPEM, 0o600))

	kid, err = keyThumbprintID(keyPath)
	require.NoError(t, err)

	return keyPath, certPath, kid
}

func newServerOptions() *ServerOptions {
	return &ServerOptions{
		Config: serverconfig.Config{
			Settings: config.Config{
				Auth: config.Auth{
					Token: tokens.Config{
						Audience:        "http://example.com",
						Issuer:          "http://example.com",
						AccessDuration:  time.Hour,
						RefreshDuration: 2 * time.Hour,
						RefreshOverlap:  -15 * time.Minute,
						Keys:            map[string]string{},
					},
				},
			},
		},
	}
}

// configureCertSlots applies cert slot settings needed by keydir tests
func configureCertSlots(so *ServerOptions, slots ...config.KeyWatcherCertSlot) {
	so.Config.Settings.Keywatcher.MaxSigningTTL = 168 * time.Hour
	so.Config.Settings.Keywatcher.SafetyWindow = 48 * time.Hour
	so.Config.Settings.Keywatcher.CertSlots = slots
}

func TestWithKeyDirLoadsKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	kid := ulid.Make().String()
	writeKey(t, dir, kid)

	so := newServerOptions()
	WithKeyDir(dir).apply(so)

	require.Contains(t, so.Config.Settings.Auth.Token.Keys, kid)
	require.Equal(t, filepath.Join(dir, kid+".pem"), so.Config.Settings.Auth.Token.Keys[kid])
}

func TestWithKeyDirWatcherDetectsNewKey(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	kid1 := ulid.Make().String()
	kid2 := ulid.Make().String()
	writeKey(t, dir, kid1)

	so := newServerOptions()

	WithKeyDirWatcher(dir).apply(so)

	require.Eventually(t, func() bool {
		jwks, err := so.Config.Handler.TokenManager.Keys()
		return err == nil && jwks.Len() == 1
	}, 5*time.Second, 100*time.Millisecond)

	// add second key
	writeKey(t, dir, kid2)
	require.Eventually(t, func() bool {
		jwks, err := so.Config.Handler.TokenManager.Keys()
		return err == nil && jwks.Len() == 2
	}, 5*time.Second, 100*time.Millisecond)
}

func TestWithKeyDirNoKID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Write a key file with a random name (simulate no KID in config)
	keyFile := writeKey(t, dir, "nokid")

	so := newServerOptions()
	// Remove any KID from config
	so.Config.Settings.Auth.Token.KID = ""
	// Remove any keys from config
	so.Config.Settings.Auth.Token.Keys = map[string]string{}

	WithKeyDir(dir).apply(so)

	// Should have exactly one key, with the filename (without .pem) as the KID
	require.Len(t, so.Config.Settings.Auth.Token.Keys, 1)
	for loadedKID, path := range so.Config.Settings.Auth.Token.Keys {
		require.Equal(t, keyFile, path)
		require.Equal(t, "nokid", loadedKID)
	}
}

// TestWithKeyDirLoadsCertSlotWithThumbprintKID verifies cert slots use public key thumbprints as kids
func TestWithKeyDirLoadsCertSlotWithThumbprintKID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	legacyKID := ulid.Make().String()
	writeKey(t, dir, legacyKID)

	slotKeyPath, _, slotKID := writeCertSlot(t, dir, "current", time.Now().Add(90*24*time.Hour))

	so := newServerOptions()
	configureCertSlots(so, config.KeyWatcherCertSlot{
		Name:        "current",
		RenewBefore: 30 * 24 * time.Hour,
	})

	WithKeyDir(dir).apply(so)

	require.Contains(t, so.Config.Settings.Auth.Token.Keys, legacyKID)
	require.Contains(t, so.Config.Settings.Auth.Token.Keys, slotKID)
	require.Equal(t, slotKeyPath, so.Config.Settings.Auth.Token.Keys[slotKID])
	require.Equal(t, slotKID, so.Config.Settings.Auth.Token.KID)
}

// TestWithKeyDirSwitchesFromUnsafeCertSlot verifies signing moves away from a slot near renewal
func TestWithKeyDirSwitchesFromUnsafeCertSlot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, _, unsafeKID := writeCertSlot(t, dir, "current", time.Now().Add((30*24+7*24+48)*time.Hour-time.Hour))
	_, _, safeKID := writeCertSlot(t, dir, "standby", time.Now().Add(90*24*time.Hour))

	so := newServerOptions()
	so.Config.Settings.Auth.Token.KID = unsafeKID
	configureCertSlots(so,
		config.KeyWatcherCertSlot{
			Name:        "current",
			RenewBefore: 30 * 24 * time.Hour,
		},
		config.KeyWatcherCertSlot{
			Name:        "standby",
			RenewBefore: 15 * 24 * time.Hour,
		},
	)

	WithKeyDir(dir).apply(so)

	require.Equal(t, safeKID, so.Config.Settings.Auth.Token.KID)
}

// TestWithKeyDirReloadDropsRotatedCertSlotKID verifies in-place key rotation does not preserve stale kids
func TestWithKeyDirReloadDropsRotatedCertSlotKID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, _, oldKID := writeCertSlot(t, dir, "current", time.Now().Add(90*24*time.Hour))

	so := newServerOptions()
	configureCertSlots(so, config.KeyWatcherCertSlot{
		Name:        "current",
		RenewBefore: 30 * 24 * time.Hour,
	})

	WithKeyDir(dir).apply(so)
	require.Contains(t, so.Config.Settings.Auth.Token.Keys, oldKID)
	require.Equal(t, oldKID, so.Config.Settings.Auth.Token.KID)

	_, _, newKID := writeCertSlot(t, dir, "current", time.Now().Add(90*24*time.Hour))
	require.NotEqual(t, oldKID, newKID)

	WithKeyDir(dir).apply(so)

	require.NotContains(t, so.Config.Settings.Auth.Token.Keys, oldKID)
	require.Contains(t, so.Config.Settings.Auth.Token.Keys, newKID)
	require.Equal(t, newKID, so.Config.Settings.Auth.Token.KID)
}
