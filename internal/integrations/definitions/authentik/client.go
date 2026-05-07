package authentik

import (
	"context"
	"net/http"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// authentikRequestTimeout is the per-request timeout in seconds
	authentikRequestTimeout = 30 * time.Second
	// authentikRateLimitMaxRetries is the maximum number of retries on rate limit
	authentikRateLimitMaxRetries = 3
	// authentikRetryBaseDelay is the base delay between retries
	authentikRetryBaseDelay = 2 * time.Second

	// authentikUsersEndpoint is the path for the Authentik users list endpoint
	authentikUsersEndpoint = "/api/v3/core/users/"
	// authentikGroupsEndpoint is the path for the Authentik groups list endpoint
	authentikGroupsEndpoint = "/api/v3/core/groups/"
	// authentikMeEndpoint is the path for the Authentik current-user endpoint
	authentikMeEndpoint = "/api/v3/core/users/me/"
	// authentikDomainsEndpoint is the path for the Authentik domains endpoint
	authentikDomainsEndpoint = "/api/v3/tenants/domains/"
)

// UserResponse represents a single Authentik user from the API
type UserResponse struct {
	PK       int    `json:"pk"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
	Type     string `json:"type"`
}

// GroupResponse represents a single Authentik group from the API
type GroupResponse struct {
	PK          string         `json:"pk"`
	Name        string         `json:"name"`
	IsSuperuser bool           `json:"is_superuser"`
	Parent      string         `json:"parent,omitempty"`
	MembersObj  []MemberObject `json:"members_obj"`
}

// MemberObject represents a group member embedded in a group response
type MemberObject struct {
	PK       int    `json:"pk"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// PaginatedResponse is the generic paginated wrapper Authentik uses for list endpoints
type PaginatedResponse[T any] struct {
	Pagination PaginationMeta `json:"pagination"`
	Results    []T            `json:"results"`
}

// PaginationMeta holds the pagination metadata from Authentik list responses
type PaginationMeta struct {
	Next       int `json:"next"`
	Previous   int `json:"previous"`
	Count      int `json:"count"`
	Current    int `json:"current"`
	TotalPages int `json:"total_pages"`
}

// DomainResponse represents a single domain entry from Authentik
type DomainResponse struct {
	ID        int    `json:"id"`
	Domain    string `json:"domain"`
	IsPrimary bool   `json:"is_primary"`
	Tenant    string `json:"tenant"`
}

// Client is the Authentik API client for one installation
type Client struct {
	// BaseURL is the base URL of the Authentik instance
	BaseURL string
	// Token is the Authentik API token
	Token string
	// HTTPClient is the underlying HTTP client
	HTTPClient *http.Client
}

// MeResponse represents the current authenticated user response from Authentik
type MeResponse struct {
	User MeUser `json:"user"`
}

// MeUser holds the current user details from the /api/v3/core/users/me/ endpoint
type MeUser struct {
	PK       int    `json:"pk"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	UID      string `json:"uid"`
}

// Build constructs the Authentik API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if cred.Token == "" {
		return nil, ErrAPITokenMissing
	}

	if cred.BaseURL == "" {
		return nil, ErrBaseURLMissing
	}

	return &Client{
		BaseURL:    cred.BaseURL,
		Token:      cred.Token,
		HTTPClient: &http.Client{Timeout: authentikRequestTimeout},
	}, nil
}

// resolveCredential extracts the CredentialSchema from the provided credential bindings
func resolveCredential(bindings types.CredentialBindings) (CredentialSchema, error) {
	cred, ok, err := authentikCredential.Resolve(bindings)
	if err != nil {
		return CredentialSchema{}, ErrCredentialDecode
	}

	if !ok {
		return CredentialSchema{}, ErrCredentialDecode
	}

	return cred, nil
}

// do executes an HTTP request with retry logic for rate limiting and server errors
func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	var (
		resp *http.Response
		err  error
	)

	for attempt := range authentikRateLimitMaxRetries {
		// clone the request for each attempt since Body can only be read once
		reqWithCtx := req.WithContext(ctx)

		resp, err = c.HTTPClient.Do(reqWithCtx)
		if err != nil {
			// network error, not worth retrying
			return nil, err
		}

		switch {
		case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
			// 2xx success, return immediately
			return resp, nil

		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError:
			// 429 or 5xx, wait and retry
			resp.Body.Close()

			if attempt < authentikRateLimitMaxRetries-1 {
				wait := authentikRetryBaseDelay * time.Duration(attempt+1)
				select {
				case <-time.After(wait):
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}

		default:
			// 4xx (not 429), return immediately, no retry
			return resp, nil
		}
	}

	return nil, ErrRequestFailed
}
