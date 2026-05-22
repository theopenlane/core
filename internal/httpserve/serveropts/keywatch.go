package serveropts

import (
	"os"
	"path/filepath"
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

			if err := applyKeyDir(s, dir); err != nil {
				log.Error().Str("dir", dir).Err(err).Msg("unable to reload key directory")

				return
			}

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

		for _, watchDir := range keyWatchDirs(s, dir) {
			if watchDir == dir {
				continue
			}

			if _, err := os.Stat(watchDir); err != nil {
				log.Warn().Str("dir", watchDir).Err(err).Msg("key watch directory not found")
				continue
			}

			if err := watcher.Add(watchDir); err != nil {
				log.Warn().Str("dir", watchDir).Err(err).Msg("failed to watch key directory")
			}
		}

		go func() {
			defer watcher.Close()

			debounce := time.NewTimer(time.Second)
			if !debounce.Stop() {
				<-debounce.C
			}

			poll := time.NewTicker(time.Minute)
			defer poll.Stop()

			scheduleReload := func() {
				if !debounce.Stop() {
					select {
					case <-debounce.C:
					default:
					}
				}

				debounce.Reset(time.Second)
			}

			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}

					if event.Has(fsnotify.Create | fsnotify.Write | fsnotify.Remove | fsnotify.Rename) {
						scheduleReload()
					}
				case <-debounce.C:
					reload()
				case <-poll.C:
					reload()
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

// keyWatchDirs returns key directories that should be watched for mounted key changes
func keyWatchDirs(s *ServerOptions, dir string) []string {
	dirs := map[string]struct{}{dir: {}}

	for _, slot := range s.Config.Settings.Keywatcher.CertSlots {
		if slot.Name == "" {
			continue
		}

		dirs[filepath.Join(dir, slot.Name)] = struct{}{}
	}

	out := make([]string, 0, len(dirs))
	for watchDir := range dirs {
		out = append(out, watchDir)
	}

	return out
}
