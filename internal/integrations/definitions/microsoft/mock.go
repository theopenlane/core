//go:build test

package microsoft

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

const (
	// MockTenantID is the tenant the mock provider issues tokens for by default
	MockTenantID = "00000000-0000-0000-0000-000000000001"
	// MockApplicationID is the application (client) ID of the mock app config
	MockApplicationID = "mock-application-id"
	// MockClientSecret is the client secret of the mock app config
	MockClientSecret = "mock-client-secret"
	// MockAccessToken is the bearer token the mock provider issues and requires on ARM and Graph requests
	MockAccessToken = "mock-access-token"
	// MockRedirectURL is the redirect URL of the mock app config
	MockRedirectURL = "https://mock.local/v1/integrations/auth/callback"
	// mockGraphPathPrefix is the path the mock serves Graph requests under
	mockGraphPathPrefix = "/graph/v1.0"
	// mockARMAudience is the token audience requested for ARM calls
	mockARMAudience = "https://management.azure.com"
)

// MockSubscription is one subscription returned by the mock ARM subscriptions listing
type MockSubscription struct {
	// ID is the subscription identifier
	ID string
	// State is the subscription lifecycle state
	State string
}

// MockMicrosoftProvider serves the Microsoft identity platform, ARM, and Graph endpoints for tests
type MockMicrosoftProvider struct {
	// Server is the mock TLS server backing every Microsoft endpoint
	Server *httptest.Server

	mux *http.ServeMux

	mu                   sync.Mutex
	tokenStatus          int
	tokenTenants         []string
	subscriptions        []MockSubscription
	subscriptionPageSize int
	subscriptionStatus   int
}

// NewMockMicrosoftProvider creates a mock Microsoft provider that issues MockAccessToken for any tenant
// and lists no subscriptions until SetSubscriptions is called
func NewMockMicrosoftProvider() *MockMicrosoftProvider {
	m := &MockMicrosoftProvider{
		mux:                http.NewServeMux(),
		tokenStatus:        http.StatusOK,
		subscriptionStatus: http.StatusOK,
	}

	m.mux.HandleFunc("GET /{tenant}/v2.0/.well-known/openid-configuration", m.serveOpenIDConfiguration)
	m.mux.HandleFunc("POST /{tenant}/oauth2/v2.0/token", m.serveToken)
	m.mux.Handle("GET /subscriptions", bearerRequired(m.serveSubscriptions))

	m.Server = httptest.NewTLSServer(m.mux)

	return m
}

// Close shuts down the mock server
func (m *MockMicrosoftProvider) Close() {
	m.Server.Close()
}

// AppConfig returns an app config whose token, ARM, and Graph calls are sent to the mock server
func (m *MockMicrosoftProvider) AppConfig() AppConfig {
	return AppConfig{
		ApplicationID: MockApplicationID,
		ClientSecret:  MockClientSecret,
		RedirectURL:   MockRedirectURL,
		endpoints: &endpointOverrides{
			cloud: cloud.Configuration{
				ActiveDirectoryAuthorityHost: m.Server.URL + "/",
				Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
					cloud.ResourceManager: {Endpoint: m.Server.URL, Audience: mockARMAudience},
				},
			},
			graphURL:   m.Server.URL + mockGraphPathPrefix,
			httpClient: m.Server.Client(),
		},
	}
}

// HandleGraph registers a Graph handler for a method and a path relative to the v1.0 root, e.g. ("GET", "/users")
func (m *MockMicrosoftProvider) HandleGraph(method, path string, handler http.HandlerFunc) {
	m.mux.Handle(method+" "+mockGraphPathPrefix+path, bearerRequired(handler))
}

// HandleARM registers an ARM handler for a method and an absolute ARM path, e.g. ("GET", "/subscriptions/{id}/providers/Microsoft.Security/assessments")
func (m *MockMicrosoftProvider) HandleARM(method, path string, handler http.HandlerFunc) {
	m.mux.Handle(method+" "+path, bearerRequired(handler))
}

// SetTokenStatus sets the status the token endpoint answers with; anything other than http.StatusOK returns an AAD error
func (m *MockMicrosoftProvider) SetTokenStatus(status int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokenStatus = status
}

// TokenTenants returns the tenant of every token request the mock has received, in order
func (m *MockMicrosoftProvider) TokenTenants() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]string(nil), m.tokenTenants...)
}

// SetSubscriptions sets the subscriptions returned by the ARM subscriptions listing
func (m *MockMicrosoftProvider) SetSubscriptions(subscriptions ...MockSubscription) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscriptions = subscriptions
}

// SetSubscriptionPageSize splits the subscriptions listing into pages of size n linked by nextLink; zero disables paging
func (m *MockMicrosoftProvider) SetSubscriptionPageSize(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscriptionPageSize = n
}

// SetSubscriptionStatus sets the status the subscriptions listing answers with
func (m *MockMicrosoftProvider) SetSubscriptionStatus(status int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscriptionStatus = status
}

// WriteJSON writes v as a JSON response with the given status, for use in registered handlers
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
}

// bearerRequired rejects requests that do not carry MockAccessToken before calling the handler
type bearerRequired http.HandlerFunc

func (h bearerRequired) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Authorization") != "Bearer "+MockAccessToken {
		WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": map[string]any{"code": "InvalidAuthenticationToken", "message": "missing or invalid bearer token"}})
		return
	}

	h(w, req)
}

func (m *MockMicrosoftProvider) serveOpenIDConfiguration(w http.ResponseWriter, req *http.Request) {
	tenantURL := m.Server.URL + "/" + req.PathValue("tenant")

	WriteJSON(w, http.StatusOK, map[string]any{
		"authorization_endpoint": tenantURL + "/oauth2/v2.0/authorize",
		"token_endpoint":         tenantURL + "/oauth2/v2.0/token",
		"issuer":                 tenantURL + "/v2.0",
	})
}

func (m *MockMicrosoftProvider) serveToken(w http.ResponseWriter, req *http.Request) {
	m.mu.Lock()
	m.tokenTenants = append(m.tokenTenants, req.PathValue("tenant"))
	status := m.tokenStatus
	m.mu.Unlock()

	if status != http.StatusOK {
		WriteJSON(w, status, map[string]any{"error": "invalid_client", "error_description": "mock token failure"})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"token_type":     "Bearer",
		"expires_in":     3600,
		"ext_expires_in": 3600,
		"access_token":   MockAccessToken,
	})
}

func (m *MockMicrosoftProvider) serveSubscriptions(w http.ResponseWriter, req *http.Request) {
	m.mu.Lock()
	status := m.subscriptionStatus
	subscriptions := m.subscriptions
	pageSize := m.subscriptionPageSize
	m.mu.Unlock()

	if status != http.StatusOK {
		WriteJSON(w, status, map[string]any{"error": map[string]any{"code": "MockFailure", "message": http.StatusText(status)}})
		return
	}

	start, _ := strconv.Atoi(req.URL.Query().Get("$skiptoken"))
	end := len(subscriptions)

	if pageSize > 0 && start+pageSize < end {
		end = start + pageSize
	}

	page := subscriptionListPage{Value: []subscription{}}
	for _, sub := range subscriptions[start:end] {
		page.Value = append(page.Value, subscription{SubscriptionID: sub.ID, State: sub.State})
	}

	if end < len(subscriptions) {
		page.NextLink = m.Server.URL + subscriptionsListPath + "&$skiptoken=" + strconv.Itoa(end)
	}

	WriteJSON(w, http.StatusOK, page)
}

// MockTokenCredential is a static azcore.TokenCredential for tests that do not need the token endpoint
type MockTokenCredential struct {
	// Err is returned from GetToken when set
	Err error
}

// GetToken returns MockAccessToken, or Err when set
func (c MockTokenCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if c.Err != nil {
		return azcore.AccessToken{}, c.Err
	}

	return azcore.AccessToken{Token: MockAccessToken, ExpiresOn: time.Now().Add(time.Hour)}, nil
}
