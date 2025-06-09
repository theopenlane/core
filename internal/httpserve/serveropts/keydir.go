package serveropts

import (
	"os"
	"path/filepath"
	"strings"

	"maps"

	"github.com/rs/zerolog/log"
)

// WithKeyDir loads all PEM files in dir into the token key map. The filename (without extension)
// is used as the kid. This allows keys to be provided via mounted secrets
func WithKeyDir(dir string) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			log.Panic().Err(err).Msg("unable to read key directory")
		}

		keys := make(map[string]string)

		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".pem") {
				continue
			}

			kid := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
			keys[kid] = filepath.Join(dir, e.Name())
		}

		// If no kid found, check config for kid
		if len(keys) == 0 {
			configKid := s.Config.Settings.Auth.Token.KID
			if configKid == "" {
				log.Panic().Msg("no key id (kid) found in key directory and auth.token.kid is not set")
			}
		}

		// merge with existing keys if any
		maps.Copy(keys, s.Config.Settings.Auth.Token.Keys)

		s.Config.Settings.Auth.Token.Keys = keys
	})
}
