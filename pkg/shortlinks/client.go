package shortlinks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
)

const (
	// defaultEndpointURL is the hosted shortlink service API endpoint
	defaultEndpointURL = "https://admin.s.theopenlane.io/api/links"
	// defaultRequestTimeout is the default timeout for requests to the shortlink service
	defaultRequestTimeout = 10 * time.Second
	// headerAccessClientID is the header for the access client ID
	headerAccessClientID = "Cf-Access-Client-Id"
	// headerAccessClientSecret is the header for the access client secret
	headerAccessClientSecret = "Cf-Access-Client-Secret"
)

// Config holds the configuration for the shortlinks client
type Config struct {
	// Enabled indicates whether shortlinks functionality is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// ClientID is the Cloudflare Access client ID for shortlink API requests
	ClientID string `json:"clientid" koanf:"clientid" default:"d5d5babbdd4ba64d59ae543ac3b6a74d.access"`
	// ClientSecret is the Cloudflare Access client secret for shortlink API requests
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" default:"" sensitive:"true"`
	// EndpointURL is the shortlink service API endpoint
	EndpointURL string `json:"endpointurl" koanf:"endpointurl" default:"https://admin.s.theopenlane.io/api/links"`
}

// Client wraps the shortlink service credentials and provides methods for creating short URLs
type Client struct {
	clientID     string
	clientSecret string
	endpointURL  string
}

// Option is a functional option for configuring the Client
type Option func(*Client)

// WithEndpointURL sets a custom endpoint URL for the shortlink service
func WithEndpointURL(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.endpointURL = url
		}
	}
}

// NewClient creates a new shortlinks client with the provided credentials
func NewClient(clientID, clientSecret string, opts ...Option) (*Client, error) {
	if clientID == "" || clientSecret == "" {
		return nil, ErrMissingAuthenticationParams
	}

	c := &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		endpointURL:  defaultEndpointURL,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// NewClientFromConfig creates a new shortlinks client from the provided config
func NewClientFromConfig(cfg Config) (*Client, error) {
	var opts []Option

	if cfg.EndpointURL != "" {
		opts = append(opts, WithEndpointURL(cfg.EndpointURL))
	}

	return NewClient(cfg.ClientID, cfg.ClientSecret, opts...)
}

// createLinkRequest represents the payload for creating a shortlink
type createLinkRequest struct {
	// URL is the original URL to shorten (the target URL)
	URL string `json:"url"`
	// Slug is an optional custom slug for the shortlink
	Slug string `json:"slug,omitempty"`
}

// responseError represents an error response from the shortlinks API
type responseError struct {
	// StatusCode is that 302'ing up your biz
	StatusCode int
	// Status is the HTTP status text
	Status string
}

// Error formats a shortlinks response error
func (e *responseError) Error() string {
	if e == nil {
		return ""
	}

	return fmt.Sprintf("shortlinks: request failed (%s)", e.Status)
}

// Create issues a POST to the hosted shortlink API and returns the short URL
func (c *Client) Create(ctx context.Context, url, slug string) (string, error) {
	if url == "" {
		return "", ErrMissingURL
	}

	payload := createLinkRequest{
		URL:  url,
		Slug: strings.TrimSpace(slug),
	}

	opts := []httpsling.Option{
		httpsling.Post(c.endpointURL),
		httpsling.Body(payload),
		httpsling.ContentType(httpsling.ContentTypeJSON),
		httpsling.Accept(httpsling.ContentTypeJSON),
		httpsling.Header(headerAccessClientID, c.clientID),
		httpsling.Header(headerAccessClientSecret, c.clientSecret),
		httpsling.Client(httpclient.Timeout(defaultRequestTimeout)),
	}

	resp, err := httpsling.ReceiveWithContext(ctx, nil, opts...)
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", ErrEmptyResponse
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read shortlink response: %w", err)
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return "", &responseError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}

	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return "", ErrEmptyResponseBody
	}

	var response struct {
		ShortURL      string `json:"shortUrl"`
		ShortURLSnake string `json:"short_url"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("decode shortlink response: %w", err)
	}

	shortURL := strings.TrimSpace(response.ShortURL)
	if shortURL == "" {
		return "", ErrMissingShortURL
	}

	return shortURL, nil
}
