package serveropts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/utils/ulids"
)

// WithSecretManagerKeys loads RSA keys from Google Secret Manager.
// The secret payload should contain either a PEM encoded private key or
// a JSON object mapping kid to PEM strings. The keys are written to
// temporary files and merged with any keys already defined.
func WithSecretManagerKeys(secretName string) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if secretName == "" {
			return
		}

		ctx := context.Background()

		client, err := secretmanager.NewClient(ctx)
		if err != nil {
			log.Panic().Err(err).Msg("failed to create secretmanager client")
		}

		defer client.Close()

		req := &secretmanagerpb.AccessSecretVersionRequest{
			Name: fmt.Sprintf("%s/versions/latest", secretName),
		}

		result, err := client.AccessSecretVersion(ctx, req)
		if err != nil {
			log.Panic().Err(err).Msg("failed to access secret version")
		}

		data := result.Payload.GetData()

		keyMap := map[string]string{}

		if err := json.Unmarshal(data, &keyMap); err != nil || len(keyMap) == 0 {
			kid := s.Config.Settings.Auth.Token.KID
			if kid == "" {
				kid = ulids.New().String()
			}

			keyMap = map[string]string{kid: string(data)}
		}

		if s.Config.Settings.Auth.Token.Keys == nil {
			s.Config.Settings.Auth.Token.Keys = map[string]string{}
		}

		for kid, pemStr := range keyMap {
			if _, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pemStr)); err != nil {
				log.Panic().Err(err).Msg("invalid PEM in secret payload")
			}

			path := filepath.Join(os.TempDir(), fmt.Sprintf("%s.pem", kid))

			if err := os.WriteFile(path, []byte(pemStr), 0o600); err != nil { //nolint:mnd
				log.Panic().Err(err).Msg("failed to write key from secret manager")
			}

			s.Config.Settings.Auth.Token.Keys[kid] = path
		}
	})
}
