package serveropts

import (
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/tokens"
)

// WithKeyDirWatcher watches dir for changes to PEM files and reloads the
// TokenManager when keys are added or removed.
func WithKeyDirWatcher(dir string) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		var mu sync.Mutex

		reload := func() {
			mu.Lock()
			defer mu.Unlock()

			WithKeyDir(dir).apply(s)

			tm, err := tokens.New(s.Config.Settings.Auth.Token)
			if err != nil {
				log.Error().Err(err).Msg("unable to reload token manager")

				return
			}

			keys, err := tm.Keys()
			if err != nil {
				log.Error().Err(err).Msg("unable to reload jwt keys")

				return
			}

			s.Config.Handler.JWTKeys = keys

			s.Config.Handler.TokenManager = tm
		}

		if _, err := os.Stat(dir); err != nil {
			log.Error().Str("dir", dir).Err(err).Msg("key directory not found")
			return
		}

		reload()

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

			debounce := time.NewTimer(time.Second)

			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}

					if event.Has(fsnotify.Create | fsnotify.Write | fsnotify.Remove | fsnotify.Rename) {
						if !debounce.Stop() {
							<-debounce.C
						}

						debounce.Reset(time.Second)
					}
				case <-debounce.C:
					reload()
					debounce.Reset(time.Second)
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
