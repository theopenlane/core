package docextract

import (
	"context"

	"google.golang.org/genai"
)

// DefaultModel is the Gemini model used when none is configured
const DefaultModel = "gemini-3.1-pro-preview"

// Client wraps the genai client with the model and system instruction used for extraction
type Client struct {
	genai.Client

	backend           genai.Backend
	model             string
	systemInstruction string
}

// Config holds the configuration options for the Client
type Config struct {
	// APIKey is the API key for authenticating with the GenAI service
	APIKey string
	// Backend is the GenAI backend to use, defaulting to the Gemini API
	Backend genai.Backend
	// Project is the Google Cloud project, required for the Vertex AI backend
	Project string
	// Location is the Google Cloud region, required for the Vertex AI backend
	Location string
	// Model is the Gemini model used for content generation
	Model string
	// SystemInstruction is the system prompt applied to every extraction request
	SystemInstruction string
}

// WithAPIKey sets the API key for the client configuration
func WithAPIKey(apiKey string) func(*Config) {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

// WithBackend sets the backend for the client configuration
func WithBackend(backend genai.Backend) func(*Config) {
	return func(c *Config) {
		c.Backend = backend
	}
}

// WithProject sets the Google Cloud project for the Vertex AI backend
func WithProject(project string) func(*Config) {
	return func(c *Config) {
		c.Project = project
	}
}

// WithLocation sets the Google Cloud region for the Vertex AI backend
func WithLocation(location string) func(*Config) {
	return func(c *Config) {
		c.Location = location
	}
}

// WithModel sets the model for the client configuration
func WithModel(model string) func(*Config) {
	return func(c *Config) {
		if model != "" {
			c.Model = model
		}
	}
}

// WithSystemInstruction sets the system prompt applied to every extraction request
func WithSystemInstruction(instruction string) func(*Config) {
	return func(c *Config) {
		c.SystemInstruction = instruction
	}
}

// NewClient creates a new Client with the provided configuration options
func NewClient(ctx context.Context, opts ...func(*Config)) (*Client, error) {
	config := &Config{
		Backend: genai.BackendGeminiAPI,
		Model:   DefaultModel,
	}
	for _, opt := range opts {
		opt(config)
	}

	clientConfig := &genai.ClientConfig{
		APIKey:   config.APIKey,
		Backend:  config.Backend,
		Project:  config.Project,
		Location: config.Location,
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, err
	}

	return &Client{Client: *client, backend: config.Backend, model: config.Model, systemInstruction: config.SystemInstruction}, nil
}
