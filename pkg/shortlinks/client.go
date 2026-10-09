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
	// LinkTTL is how long a created link stays resolvable when the caller sets no expiration
	LinkTTL time.Duration `json:"linkttl" koanf:"linkttl" default:"2160h"`
}

// Client wraps the shortlink service credentials and provides methods for creating short URLs
type Client struct {
	clientID     string
	clientSecret string
	endpointURL  string
	linkTTL      time.Duration
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

// WithLinkTTL sets the default lifetime applied to links created without an explicit expiration
func WithLinkTTL(ttl time.Duration) Option {
	return func(c *Client) {
		if ttl > 0 {
			c.linkTTL = ttl
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

	opts = append(opts, WithLinkTTL(cfg.LinkTTL))

	return NewClient(cfg.ClientID, cfg.ClientSecret, opts...)
}

// CreateRequest describes the shortlink to create
type CreateRequest struct {
	// URL is the original URL to shorten (the target URL)
	URL string `json:"url"`
	// Slug is an optional custom slug for the shortlink
	Slug string `json:"slug,omitempty"`
	// Expiration is the unix timestamp after which the link stops resolving; zero applies the
	// client's LinkTTL so every link the service holds eventually expires
	Expiration int64 `json:"expiration,omitempty"`
	// Metadata is recorded on every click of the link
	Metadata Metadata `json:"metadata,omitzero"`
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
func (c *Client) Create(ctx context.Context, req CreateRequest) (string, error) {
	if req.URL == "" {
		return "", ErrMissingURL
	}

	req.Slug = strings.TrimSpace(req.Slug)
	if req.Expiration == 0 && c.linkTTL > 0 {
		req.Expiration = time.Now().Add(c.linkTTL).Unix()
	}

	opts := []httpsling.Option{
		httpsling.Post(c.endpointURL),
		httpsling.Body(req),
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
