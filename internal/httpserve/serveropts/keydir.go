package serveropts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"maps"

	"github.com/rs/zerolog/log"
)

var errNoKID = errors.New("no key id (kid) found in key directory and auth.token.kid is not set")

// loadKeyDir loads all PEM files in dir into the token key map, using each filename (without
// extension) as the kid, so keys can be provided via mounted secrets. It returns an error rather
// than panicking: the startup path fails fast on it, the key watcher goroutine just logs it.
func loadKeyDir(s *ServerOptions, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("unable to read key directory: %w", err)
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
	if len(keys) == 0 && s.Config.Settings.Auth.Token.KID == "" {
		return errNoKID
	}

	// merge with existing keys if any
	maps.Copy(keys, s.Config.Settings.Auth.Token.Keys)

	s.Config.Settings.Auth.Token.Keys = keys

	return nil
}

// WithKeyDir loads the keys in dir at startup, panicking on failure since a ServerOption cannot
// return an error and a bad key directory must not boot silently.
func WithKeyDir(dir string) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if err := loadKeyDir(s, dir); err != nil {
			log.Panic().Err(err).Msg("unable to load key directory")
		}
	})
}
