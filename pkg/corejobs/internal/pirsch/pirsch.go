package pirsch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

var (
	// ErrAuthFailed is returned when authentication fails
	ErrAuthFailed = errors.New("authentication failed")
	// ErrRetryableStatus is returned when a retryable status code is received
	ErrRetryableStatus = errors.New("retryable status code received")
	// ErrListDomainsFailed is returned when listing domains fails
	ErrListDomainsFailed = errors.New("list domains failed")
	// ErrGetDomainFailed is returned when getting a domain fails
	ErrGetDomainFailed = errors.New("get domain failed")
	// ErrCreateDomainFailed is returned when creating a domain fails
	ErrCreateDomainFailed = errors.New("create domain failed")
	// ErrDeleteDomainFailed is returned when deleting a domain fails
	ErrDeleteDomainFailed = errors.New("delete domain failed")
)

// Client interface for interacting with the Pirsch API
type Client interface {
	ListDomains(ctx context.Context, search string) ([]Domain, error)
	GetDomain(ctx context.Context, domainID string) (*Domain, error)
	CreateDomain(ctx context.Context, req CreateDomainRequest) (*Domain, error)
	DeleteDomain(ctx context.Context, domainID string) error
}

const (
	baseURL = "https://api.pirsch.io/api/v1"

	// Default retry and timeout configuration
	defaultMaxRetries      = 3
	defaultMaxDelaySeconds = 30
	defaultTimeoutSeconds  = 10
	exponentialBackoffBase = 2
)

// BackoffStrategy defines the backoff calculation method
type BackoffStrategy string

const (
	BackoffStrategyExponential BackoffStrategy = "exponential"
	BackoffStrategyLinear      BackoffStrategy = "linear"
	BackoffStrategyFixed       BackoffStrategy = "fixed"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffStrategy BackoffStrategy
	RetryableStatus []int // HTTP status codes that should trigger a retry
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      defaultMaxRetries,
		InitialDelay:    1 * time.Second,
		MaxDelay:        defaultMaxDelaySeconds * time.Second,
		BackoffStrategy: BackoffStrategyExponential,
		RetryableStatus: []int{429, 500, 502, 503, 504},
	}
}

// client handles API requests to Pirsch
type client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	accessToken  string
	expiresAt    time.Time
	retryConfig  RetryConfig
}

// ClientOption is a function that configures a client
type ClientOption func(*client)

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *client) {
		c.httpClient.Timeout = timeout
	}
}

// WithRetryConfig sets the retry configuration
func WithRetryConfig(config RetryConfig) ClientOption {
	return func(c *client) {
		c.retryConfig = config
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new Pirsch API client with optional configuration
func NewClient(clientID, clientSecret string, opts ...ClientOption) Client {
	c := &client{
		httpClient:   &http.Client{Timeout: defaultTimeoutSeconds * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
		retryConfig:  DefaultRetryConfig(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// TokenResponse represents the authentication response
type TokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Domain represents a Pirsch domain
type Domain struct {
	ID                    string                 `json:"id"`
	DefTime               time.Time              `json:"def_time"`
	ModTime               time.Time              `json:"mod_time"`
	UserID                string                 `json:"user_id"`
	OrganizationID        *string                `json:"organization_id"`
	Hostname              string                 `json:"hostname"`
	Subdomain             string                 `json:"subdomain"`
	IdentificationCode    string                 `json:"identification_code"`
	Public                bool                   `json:"public"`
	GoogleUserID          *string                `json:"google_user_id"`
	GoogleUserEmail       *string                `json:"google_user_email"`
	GSCDomain             *string                `json:"gsc_domain"`
	NewOwner              *int64                 `json:"new_owner"`
	Timezone              *string                `json:"timezone"`
	GroupByTitle          bool                   `json:"group_by_title"`
	ActiveVisitorsSeconds *int64                 `json:"active_visitors_seconds"`
	DisableScripts        bool                   `json:"disable_scripts"`
	StatisticsStart       *time.Time             `json:"statistics_start"`
	ImportedStatistics    bool                   `json:"imported_statistics"`
	ThemeID               string                 `json:"theme_id"`
	Theme                 map[string]interface{} `json:"theme"`
	CustomDomain          *string                `json:"custom_domain"`
	DisplayName           *string                `json:"display_name"`
	UserRole              string                 `json:"user_role"`
	Settings              map[string]interface{} `json:"settings"`
	ThemeSettings         map[string]interface{} `json:"theme_settings"`
	Pinned                bool                   `json:"pinned"`
	SubscriptionActive    bool                   `json:"subscription_active"`
}

// CreateDomainRequest represents the request to create a domain
type CreateDomainRequest struct {
	Hostname                    string  `json:"hostname"`
	Subdomain                   string  `json:"subdomain"`
	Timezone                    string  `json:"timezone"`
	OrganizationID              *string `json:"organization_id,omitempty"`
	ThemeID                     *string `json:"theme_id,omitempty"`
	Public                      bool    `json:"public"`
	GroupByTitle                bool    `json:"group_by_title"`
	ActiveVisitorsSeconds       int     `json:"active_visitors_seconds"`
	DisableScripts              bool    `json:"disable_scripts"`
	DisplayName                 string  `json:"display_name,omitempty"`
	TrafficSpikeThreshold       int     `json:"traffic_spike_threshold"`
	TrafficWarningThresholdDays int     `json:"traffic_warning_threshold_days"`
}

// authenticate obtains an access token
func (c *client) authenticate(ctx context.Context) error {
	if time.Now().Before(c.expiresAt) {
		return nil // Token still valid
	}

	reqBody := map[string]string{
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/token", baseURL), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status %d: %s", ErrAuthFailed, resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.expiresAt = tokenResp.ExpiresAt

	return nil
}

// calculateBackoff calculates the delay for a given retry attempt
func (c *client) calculateBackoff(attempt int) time.Duration {
	var delay time.Duration

	switch c.retryConfig.BackoffStrategy {
	case BackoffStrategyExponential:
		delay = c.retryConfig.InitialDelay * time.Duration(math.Pow(exponentialBackoffBase, float64(attempt)))
	case BackoffStrategyLinear:
		delay = c.retryConfig.InitialDelay * time.Duration(attempt+1)
	case BackoffStrategyFixed:
		delay = c.retryConfig.InitialDelay
	default:
		delay = c.retryConfig.InitialDelay * time.Duration(math.Pow(exponentialBackoffBase, float64(attempt)))
	}

	if delay > c.retryConfig.MaxDelay {
		delay = c.retryConfig.MaxDelay
	}

	return delay
}

// isRetryable checks if an HTTP status code should trigger a retry
func (c *client) isRetryable(statusCode int) bool {
	for _, code := range c.retryConfig.RetryableStatus {
		if code == statusCode {
			return true
		}
	}
	return false
}

// doRequest performs an authenticated HTTP request with retry logic
func (c *client) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, err
	}

	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := c.calculateBackoff(attempt - 1)
			time.Sleep(delay)
		}

		var reqBody io.Reader
		if body != nil {
			data, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(data)
		}

		req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", baseURL, endpoint), reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed (attempt %d/%d): %w", attempt+1, c.retryConfig.MaxRetries+1, err)
			continue
		}

		// If status is retryable, close body and retry
		if c.isRetryable(resp.StatusCode) {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("%w: status %d (attempt %d/%d): %s",
				ErrRetryableStatus, resp.StatusCode, attempt+1, c.retryConfig.MaxRetries+1, string(body))

			// Re-authenticate if we got a 401
			if resp.StatusCode == http.StatusUnauthorized {
				c.expiresAt = time.Time{} // Force re-authentication
				if err := c.authenticate(ctx); err != nil {
					return nil, fmt.Errorf("re-authentication failed: %w", err)
				}
			}
			continue
		}

		// Success or non-retryable error
		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ListDomains lists all domains (for user clients)
func (c *client) ListDomains(ctx context.Context, search string) ([]Domain, error) {
	endpoint := "/domain"
	if search != "" {
		endpoint = fmt.Sprintf("%s?search=%s", endpoint, search)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrListDomainsFailed, resp.StatusCode, string(body))
	}

	var domains []Domain
	if err := json.NewDecoder(resp.Body).Decode(&domains); err != nil {
		return nil, fmt.Errorf("failed to decode domains: %w", err)
	}

	return domains, nil
}

// GetDomain retrieves a specific domain by ID
func (c *client) GetDomain(ctx context.Context, domainID string) (*Domain, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/domain?id=%s", domainID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrGetDomainFailed, resp.StatusCode, string(body))
	}

	var domain Domain
	if err := json.NewDecoder(resp.Body).Decode(&domain); err != nil {
		return nil, fmt.Errorf("failed to decode domain: %w", err)
	}

	return &domain, nil
}

// CreateDomain creates a new domain
func (c *client) CreateDomain(ctx context.Context, req CreateDomainRequest) (*Domain, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/domain", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrCreateDomainFailed, resp.StatusCode, string(body))
	}

	var domain Domain
	if err := json.NewDecoder(resp.Body).Decode(&domain); err != nil {
		return nil, fmt.Errorf("failed to decode domain: %w", err)
	}

	return &domain, nil
}

// DeleteDomain deletes a domain by ID
func (c *client) DeleteDomain(ctx context.Context, domainID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/domain?id=%s", domainID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status %d: %s", ErrDeleteDomainFailed, resp.StatusCode, string(body))
	}

	return nil
}
