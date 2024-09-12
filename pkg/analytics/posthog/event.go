package posthog

import (
	"time"

	"github.com/posthog/posthog-go"

	"github.com/theopenlane/core/pkg/analytics/machine"
)

// Config is the configuration for PostHog
type Config struct {
	// Enabled is a flag to enable or disable PostHog
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// APIKey is the PostHog API Key
	APIKey string `json:"apiKey" koanf:"apiKey"`
	// Host is the PostHog API Host
	Host string `json:"host" koanf:"host" default:"https://app.posthog.com"`
}

type PostHog struct {
	client     posthog.Client
	Identifier string
}

// Init returns a pointer to a PostHog object
func (c *Config) Init() *PostHog {
	if !c.Enabled || c.APIKey == "" || c.Host == "" || !machine.Available() {
		return nil
	}

	client, _ := posthog.NewWithConfig(c.APIKey, posthog.Config{
		Endpoint:  c.Host,
		BatchSize: 1,
	})

	if client != nil {
		return &PostHog{
			client:     client,
			Identifier: machine.ID(),
		}
	}

	return nil
}

// Event is used to send an event to PostHog
func (p *PostHog) Event(eventName string, properties posthog.Properties) {
	_ = p.client.Enqueue(posthog.Capture{
		DistinctId: p.Identifier,
		Event:      eventName,
		Timestamp:  time.Now(),
		Properties: properties,
	})
}

// Properties sets generic properties
func (p *PostHog) Properties(id, obj string, properties posthog.Properties) {
	_ = p.client.Enqueue(posthog.GroupIdentify{
		Type:       obj,
		Key:        id,
		Timestamp:  time.Now(),
		Properties: properties,
	})
}
