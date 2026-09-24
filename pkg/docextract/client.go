package docextract

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth"
	"cloud.google.com/go/auth/credentials"
	"google.golang.org/genai"
)

// DefaultModel is the Gemini model used when none is configured
const DefaultModel = "gemini-3.1-pro-preview"

// cloudPlatformScope is the oauth scope Vertex AI requests need
const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

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
	// CredentialsJSON is a service account key for the Vertex AI backend; application default credentials are used when empty
	CredentialsJSON string
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

// WithCredentialsJSON sets a service account key for the Vertex AI backend
func WithCredentialsJSON(credentialsJSON string) func(*Config) {
	return func(c *Config) {
		c.CredentialsJSON = credentialsJSON
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

	if config.Backend == genai.BackendVertexAI {
		creds, err := LoadCredentials(config.CredentialsJSON)
		if err != nil {
			return nil, err
		}

		clientConfig.Credentials = creds
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, err
	}

	return &Client{Client: *client, backend: config.Backend, model: config.Model, systemInstruction: config.SystemInstruction}, nil
}

// LoadCredentials builds Google Cloud credentials from a service account key, or returns nil so
// callers fall back to application default credentials when no key is configured
func LoadCredentials(credentialsJSON string) (*auth.Credentials, error) {
	if credentialsJSON == "" {
		return nil, nil
	}

	creds, err := credentials.NewCredentialsFromJSON(credentials.ServiceAccount, []byte(credentialsJSON), &credentials.DetectOptions{Scopes: []string{cloudPlatformScope}})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCredentialsInvalid, err)
	}

	return creds, nil
}
