package config

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// ProviderWithRefresh shows a config provider with automatic refresh; it contains fields and methods to manage the configuration,
// and refresh it periodically based on a specified interval
type ProviderWithRefresh struct {
	sync.RWMutex

	config *Config

	configProvider Provider

	refreshInterval time.Duration

	ticker *time.Ticker
	stop   chan bool
}

// NewProviderWithRefresh function is a constructor function that creates a new instance of ProviderWithRefresh
func NewProviderWithRefresh(cfgProvider Provider) (*ProviderWithRefresh, error) {
	cfg, err := cfgProvider.Get()
	if err != nil {
		return nil, err
	}

	cfgRefresh := &ProviderWithRefresh{
		config:          cfg,
		configProvider:  cfgProvider,
		refreshInterval: cfg.Settings.RefreshInterval,
	}
	cfgRefresh.initialize()

	return cfgRefresh, nil
}

// GetConfig retrieves the current echo server configuration; it acquires a read lock to ensure thread safety and returns the `config` field
func (s *ProviderWithRefresh) Get() (*Config, error) {
	s.RLock()
	defer s.RUnlock()

	return s.config, nil
}

// initialize the config provider with refresh
func (s *ProviderWithRefresh) initialize() {
	if s.refreshInterval != 0 {
		s.stop = make(chan bool)
		s.ticker = time.NewTicker(s.refreshInterval)

		go s.refreshConfig()
	}
}

func (s *ProviderWithRefresh) refreshConfig() {
	for {
		select {
		case <-s.stop:
			break
		case <-s.ticker.C:
		}

		newConfig, err := s.configProvider.Get()
		if err != nil {
			log.Error().Msg("failed to load new server configuration")
			continue
		}

		log.Info().Msg("loaded new server configuration")

		s.Lock()
		s.config = newConfig
		s.Unlock()
	}
}

// Close function is used to stop the automatic refresh of the configuration.
// It stops the ticker that triggers the refresh and closes the stop channel,
// which signals the goroutine to stop refreshing the configuration
func (s *ProviderWithRefresh) Close() {
	if s.ticker != nil {
		s.ticker.Stop()
	}

	if s.stop != nil {
		s.stop <- true
		close(s.stop)
	}
}
