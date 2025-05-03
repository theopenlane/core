package testutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"
)

// NewRedisClient creates a new redis client for testing using miniredis
func NewRedisClient() *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:             mr.Addr(),
		DisableIndentity: true, // # spellcheck:off
	})

	return client
}

// CreateSessionManager creates a new session manager for testing
func CreateSessionManager() sessions.Store[map[string]any] {
	hashKey := randomString(32)  //nolint:mnd
	blockKey := randomString(32) //nolint:mnd

	sm := sessions.NewCookieStore[map[string]any](sessions.DebugCookieConfig,
		hashKey, blockKey,
	)

	return sm
}

// randomString generates a random string of n bytes
func randomString(n int) []byte {
	id := make([]byte, n)

	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		panic(err) // This shouldn't happen
	}

	return id
}

// CreateTokenManager creates a new token manager for testing
func CreateTokenManager(refreshOverlap time.Duration) (*tokens.TokenManager, error) {
	conf := tokens.Config{
		Audience:        "http://localhost:17608",
		Issuer:          "http://localhost:17608",
		AccessDuration:  1 * time.Hour, //nolint:mnd
		RefreshDuration: 2 * time.Hour, //nolint:mnd
		RefreshOverlap:  refreshOverlap,
		JWKSEndpoint:    "http://localhost:17608/.well-known/jwks.json",
	}

	// generate keys if needed
	conf, err := generateKeys(conf)
	if err != nil {
		log.Fatalf("Error generating keys: %v", err)
	}

	return tokens.New(conf)
}

func generateKeys(conf tokens.Config) (tokens.Config, error) {
	privFileName := "../../../private_key.pem"

	// generate a new private key if one doesn't exist
	if _, err := os.Stat(privFileName); err != nil {
		// Generate a new RSA private key with 2048 bits
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048) //nolint:mnd
		if err != nil {
			log.Fatalf("Error generating RSA private key")
		}

		// Encode the private key to the PEM format
		privateKeyPEM := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}

		privateKeyFile, err := os.Create(privFileName)
		if err != nil {
			log.Fatalf("Error creating private key file")
		}

		if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
			log.Fatalf("unable to encode pem on startup")
		}

		privateKeyFile.Close()
	}

	keys := map[string]string{}

	// check if kid was passed in
	kidPriv := conf.KID

	// if we didn't get a kid in the settings, assign one
	if kidPriv == "" {
		kidPriv = ulids.New().String()
	}

	keys[kidPriv] = fmt.Sprintf("%v", privFileName)

	conf.Keys = keys

	return conf, nil
}
