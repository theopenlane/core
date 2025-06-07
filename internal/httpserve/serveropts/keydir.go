package serveropts

import (
	"os"
	"path/filepath"
	"strings"

	"maps"

	"github.com/rs/zerolog/log"
)

// WithKeyDir loads all PEM files in dir into the token key map. The filename (without extension)
// is used as the kid. This allows keys to be provided via mounted secrets.
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

		// merge with existing keys if any
		maps.Copy(keys, s.Config.Settings.Auth.Token.Keys)

		s.Config.Settings.Auth.Token.Keys = keys
	})
}

// WithKeyDirFromEnv loads keys from the directory specified by the
// CORE_AUTH_TOKEN_KEYDIR environment variable, if set.
func WithKeyDirFromEnv() ServerOption {
	dir := os.Getenv("CORE_AUTH_TOKEN_KEYDIR")
	if dir == "" {
		return newApplyFunc(func(*ServerOptions) {})
	}

	return WithKeyDir(dir)
}

// WithKeyDirOption allows the key directory to be set via server config.
func WithKeyDirOption() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Auth.Token.KeyDir != "" {
			WithKeyDir(s.Config.Settings.Auth.Token.KeyDir).apply(s)
		}
	})
}
