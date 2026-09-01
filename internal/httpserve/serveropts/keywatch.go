package serveropts

import (
	"context"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/logx"
)

const (
	// keyReloadDebounce coalesces the burst of events a projected volume swap produces
	keyReloadDebounce = time.Second

	// keyReloadInterval re-evaluates key selection even when no event arrives, since
	// fsnotify on a projected volume watches a symlink swap rather than the key files
	keyReloadInterval = time.Hour

	// keyReloadRetry is the initial delay before retrying a failed reload, so a transient
	// archive outage does not leave the manager without retired keys until the next event
	keyReloadRetry = 5 * time.Second

	// keyReloadMaxRetry caps the backoff between retries of a persistently failing reload
	keyReloadMaxRetry = 5 * time.Minute
)

// WithKeyDirWatcher watches dir and applies key changes to the existing TokenManager.
// The manager is mutated rather than replaced so consumers holding the pointer, notably
// the ent client, keep signing and verifying with current keys
func WithKeyDirWatcher(dir string, base tokens.Config, archive keyArchive) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		configured, err := configuredVerificationKeys(base.Keys)
		if err != nil {
			log.Panic().Err(err).Msg("unable to load configured verification keys")
		}

		reload := func() bool {
			ctx := context.Background()

			keySet, healthy, err := keySetFromKeyDir(ctx, base, dir, configured, archive)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("dir", dir).Msg("unable to reload signing keys from key directory")

				return false
			}

			previous := s.Config.Handler.TokenManager.CurrentKeyID()

			if err := s.Config.Handler.TokenManager.SetKeySet(keySet); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("dir", dir).Msg("unable to apply signing keys to token manager")

				return false
			}

			if !healthy {
				logx.FromContext(ctx).Warn().Str("kid", keySet.KID).Msg("no signing key has enough certificate headroom; signing with the longest lived key available")
			}

			if keySet.KID != previous {
				logx.FromContext(ctx).Info().Str("kid", keySet.KID).Str("previous", previous).Msg("signing key changed")
			}

			return true
		}

		if _, err := os.Stat(dir); err != nil {
			log.Error().Str("dir", dir).Err(err).Msg("key directory not found")

			return
		}

		loaded := reload()

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Error().Err(err).Msg("failed to create fsnotify watcher")

			return
		}

		if err := watcher.Add(dir); err != nil {
			log.Error().Err(err).Msg("failed to watch key directory")

			return
		}

		go func() {
			defer watcher.Close()

			// one timer drives everything: debounce after events, backoff after a failed
			// reload, and the periodic re-evaluation after a successful one
			delay := keyReloadInterval
			if !loaded {
				delay = keyReloadRetry
			}

			timer := time.NewTimer(delay)
			defer timer.Stop()

			retry := keyReloadRetry

			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}

					if event.Has(fsnotify.Create | fsnotify.Write | fsnotify.Remove | fsnotify.Rename) {
						timer.Reset(keyReloadDebounce)
					}
				case <-timer.C:
					if reload() {
						retry = keyReloadRetry
						timer.Reset(keyReloadInterval)

						continue
					}

					timer.Reset(retry)
					retry = min(retry*2, keyReloadMaxRetry) //nolint:mnd
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}

					log.Error().Err(err).Msg("watcher error")
				}
			}
		}()
	})
}
