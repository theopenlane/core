package serveropts

import (
	"context"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/catalog"
)

// WithModuleCatalog loads the module catalog from the configured path
func WithModuleCatalog(path string) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		cat, err := catalog.LoadCatalog(path)
		if err != nil {
			log.Panic().Err(err).Str("path", path).Msg("failed to load module catalog")
		}
		if s.Config.Handler.Entitlements != nil {
			if err := cat.EnsurePrices(context.Background(), s.Config.Handler.Entitlements, "usd"); err != nil {
				log.Panic().Err(err).Msg("invalid catalog pricing")
			}
		}
		s.Config.Handler.Catalog = cat
	})
}

// WithModuleCatalogWatcher reloads the catalog when the file changes
func WithModuleCatalogWatcher(path string) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Error().Err(err).Msg("failed to create catalog watcher")
			return
		}
		if err := watcher.Add(path); err != nil {
			log.Error().Err(err).Msg("failed to watch catalog file")
			return
		}
		go func() {
			defer watcher.Close()
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Has(fsnotify.Create | fsnotify.Write | fsnotify.Rename) {
						cat, err := catalog.LoadCatalog(path)
						if err != nil {
							log.Error().Err(err).Msg("failed to reload module catalog")
							continue
						}
						if s.Config.Handler.Entitlements != nil {
							if err := cat.EnsurePrices(context.Background(), s.Config.Handler.Entitlements, "usd"); err != nil {
								log.Error().Err(err).Msg("invalid catalog pricing")
								continue
							}
						}
						s.Config.Handler.Catalog = cat
						log.Info().Str("path", path).Msg("reloaded module catalog")
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Error().Err(err).Msg("catalog watcher error")
				}
			}
		}()
	})
}
