package serveropts

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
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
