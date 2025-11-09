package pirsch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

type PirschClient interface {
	ListDomains(search string) ([]Domain, error)
	GetDomain(domainID string) (*Domain, error)
	CreateDomain(req CreateDomainRequest) (*Domain, error)
	DeleteDomain(domainID string) error
}

const (
	baseURL = "https://api.pirsch.io/api/v1"
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
		MaxRetries:      3,
		InitialDelay:    1 * time.Second,
		MaxDelay:        30 * time.Second,
		BackoffStrategy: BackoffStrategyExponential,
		RetryableStatus: []int{429, 500, 502, 503, 504},
	}
}

// Client handles API requests to Pirsch
type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	accessToken  string
	expiresAt    time.Time
	retryConfig  RetryConfig
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithRetryConfig sets the retry configuration
func WithRetryConfig(config RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = config
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new Pirsch API client with optional configuration
func NewClient(clientID, clientSecret string, opts ...ClientOption) *Client {
	c := &Client{
		httpClient:   &http.Client{Timeout: 10 * time.Second},
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
func (c *Client) authenticate() error {
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

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/token", baseURL),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
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
func (c *Client) calculateBackoff(attempt int) time.Duration {
	var delay time.Duration

	switch c.retryConfig.BackoffStrategy {
	case BackoffStrategyExponential:
		delay = c.retryConfig.InitialDelay * time.Duration(math.Pow(2, float64(attempt)))
	case BackoffStrategyLinear:
		delay = c.retryConfig.InitialDelay * time.Duration(attempt+1)
	case BackoffStrategyFixed:
		delay = c.retryConfig.InitialDelay
	default:
		delay = c.retryConfig.InitialDelay * time.Duration(math.Pow(2, float64(attempt)))
	}

	if delay > c.retryConfig.MaxDelay {
		delay = c.retryConfig.MaxDelay
	}

	return delay
}

// isRetryable checks if an HTTP status code should trigger a retry
func (c *Client) isRetryable(statusCode int) bool {
	for _, code := range c.retryConfig.RetryableStatus {
		if code == statusCode {
			return true
		}
	}
	return false
}

// doRequest performs an authenticated HTTP request with retry logic
func (c *Client) doRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	if err := c.authenticate(); err != nil {
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

		req, err := http.NewRequest(method, fmt.Sprintf("%s%s", baseURL, endpoint), reqBody)
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
			lastErr = fmt.Errorf("retryable status %d (attempt %d/%d): %s",
				resp.StatusCode, attempt+1, c.retryConfig.MaxRetries+1, string(body))

			// Re-authenticate if we got a 401
			if resp.StatusCode == http.StatusUnauthorized {
				c.expiresAt = time.Time{} // Force re-authentication
				if err := c.authenticate(); err != nil {
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
func (c *Client) ListDomains(search string) ([]Domain, error) {
	endpoint := "/domain"
	if search != "" {
		endpoint = fmt.Sprintf("%s?search=%s", endpoint, search)
	}

	resp, err := c.doRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list domains failed with status %d: %s", resp.StatusCode, string(body))
	}

	var domains []Domain
	if err := json.NewDecoder(resp.Body).Decode(&domains); err != nil {
		return nil, fmt.Errorf("failed to decode domains: %w", err)
	}

	return domains, nil
}

// GetDomain retrieves a specific domain by ID
func (c *Client) GetDomain(domainID string) (*Domain, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("/domain?id=%s", domainID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get domain failed with status %d: %s", resp.StatusCode, string(body))
	}

	var domain Domain
	if err := json.NewDecoder(resp.Body).Decode(&domain); err != nil {
		return nil, fmt.Errorf("failed to decode domain: %w", err)
	}

	return &domain, nil
}

// CreateDomain creates a new domain
func (c *Client) CreateDomain(req CreateDomainRequest) (*Domain, error) {
	resp, err := c.doRequest(http.MethodPost, "/domain", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create domain failed with status %d: %s", resp.StatusCode, string(body))
	}

	var domain Domain
	if err := json.NewDecoder(resp.Body).Decode(&domain); err != nil {
		return nil, fmt.Errorf("failed to decode domain: %w", err)
	}

	return &domain, nil
}

// DeleteDomain deletes a domain by ID
func (c *Client) DeleteDomain(domainID string) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("/domain?id=%s", domainID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete domain failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
