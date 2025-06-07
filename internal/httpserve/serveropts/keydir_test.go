package serveropts

import (
	"crypto/rand"
	"crypto/rsa"
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
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	path := filepath.Join(dir, kid+".pem")
	data := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

func newServerOptions() *ServerOptions {
	return &ServerOptions{
		Config: serverconfig.Config{
			Settings: config.Config{
				Auth: config.Auth{
					Token: tokens.Config{
						Audience: "http://example.com",
						Issuer:   "http://example.com",
						Keys:     map[string]string{},
					},
				},
			},
		},
	}
}

func TestWithKeyDirLoadsKeys(t *testing.T) {
	dir := t.TempDir()
	kid := ulid.Make().String()
	writeKey(t, dir, kid)

	so := newServerOptions()
	WithKeyDir(dir).apply(so)

	require.Contains(t, so.Config.Settings.Auth.Token.Keys, kid)
	require.Equal(t, filepath.Join(dir, kid+".pem"), so.Config.Settings.Auth.Token.Keys[kid])
}

func TestWithKeyDirWatcherDetectsNewKey(t *testing.T) {
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
