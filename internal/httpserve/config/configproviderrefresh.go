package config

import (
	"sync"
	"time"
)

// ConfigProviderWithRefresh shows a config provider with automatic refresh; it contains fields and methods to manage the configuration,
// and refresh it periodically based on a specified interval
type ConfigProviderWithRefresh struct {
	sync.RWMutex

	config *Config

	configProvider ConfigProvider

	refreshInterval time.Duration

	ticker *time.Ticker
	stop   chan bool
}

// NewConfigProviderWithRefresh function is a constructor function that creates a new instance of ConfigProviderWithRefresh
func NewConfigProviderWithRefresh(cfgProvider ConfigProvider) (*ConfigProviderWithRefresh, error) {
	cfg, err := cfgProvider.GetConfig()
	if err != nil {
		return nil, err
	}

	cfgRefresh := &ConfigProviderWithRefresh{
		config:          cfg,
		configProvider:  cfgProvider,
		refreshInterval: cfg.Settings.RefreshInterval,
	}
	cfgRefresh.initialize()

	return cfgRefresh, nil
}

// GetConfig retrieves the current echo server configuration; it acquires a read lock to ensure thread safety and returns the `config` field
func (s *ConfigProviderWithRefresh) GetConfig() (*Config, error) {
	s.RLock()
	defer s.RUnlock()

	return s.config, nil
}

// initialize the config provider with refresh
func (s *ConfigProviderWithRefresh) initialize() {
	if s.refreshInterval != 0 {
		s.stop = make(chan bool)
		s.ticker = time.NewTicker(s.refreshInterval)

		go s.refreshConfig()
	}
}

func (s *ConfigProviderWithRefresh) refreshConfig() {
	for {
		select {
		case <-s.stop:
			break
		case <-s.ticker.C:
		}

		newConfig, err := s.configProvider.GetConfig()
		if err != nil {
			s.config.Logger.Error("failed to load new server configuration")
			continue
		}

		s.config.Logger.Info("loaded new server configuration")

		s.Lock()
		s.config = newConfig
		s.Unlock()
	}
}

// Close function is used to stop the automatic refresh of the configuration.
// It stops the ticker that triggers the refresh and closes the stop channel,
// which signals the goroutine to stop refreshing the configuration
func (s *ConfigProviderWithRefresh) Close() {
	if s.ticker != nil {
		s.ticker.Stop()
	}

	if s.stop != nil {
		s.stop <- true
		close(s.stop)
	}
}
