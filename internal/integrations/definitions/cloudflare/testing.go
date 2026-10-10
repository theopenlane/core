//go:build test

package cloudflare

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"sync"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

const (
	// MockAccountID is the account the mock runtime client is scoped to
	MockAccountID = "mock-account"
	// mockAPIToken is the token the mock runtime client is built with
	mockAPIToken = "mock-token"
)

// MockCloudflareRuntime holds a mock Cloudflare API server for integration test suites
type MockCloudflareRuntime struct {
	// Server is the mock HTTP server backing the Cloudflare API
	Server *httptest.Server

	mu          sync.Mutex
	scanResults map[string]int
}

// NewMockCloudflareRuntime creates a mock Cloudflare server that answers URL Scanner result
// lookups with the status registered for each scan result id
func NewMockCloudflareRuntime() *MockCloudflareRuntime {
	m := &MockCloudflareRuntime{scanResults: map[string]int{}}

	m.Server = httptest.NewServer(http.HandlerFunc(m.serveScanResult))

	return m
}

// SetScanResult registers the status returned for a URL Scanner result id; http.StatusOK
// returns a successful scan result
func (m *MockCloudflareRuntime) SetScanResult(scanResultID string, status int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scanResults[scanResultID] = status
}

func (m *MockCloudflareRuntime) serveScanResult(w http.ResponseWriter, req *http.Request) {
	m.mu.Lock()
	status, ok := m.scanResults[path.Base(req.URL.Path)]
	m.mu.Unlock()

	if !ok {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if status != http.StatusOK {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "errors": []map[string]any{{"code": status, "message": http.StatusText(status)}}})

		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"task": map[string]any{"success": true, "errors": []any{}},
	})
}

// Close shuts down the mock HTTP server
func (m *MockCloudflareRuntime) Close() {
	m.Server.Close()
}

// mockCloudflareClient builds a CloudflareClient whose API is pointed at the mock server
func mockCloudflareClient(baseURL string) *CloudflareClient {
	return &CloudflareClient{
		Client: cf.NewClient(
			option.WithAPIToken(mockAPIToken),
			option.WithBaseURL(baseURL),
			option.WithMaxRetries(0),
		),
		Config: ClientConfig{
			AccountID: MockAccountID,
			APIToken:  mockAPIToken,
		},
	}
}

// Builder returns a Cloudflare definition builder backed by the mock server; every client
// build, including the runtime client, returns a CloudflareClient pointed at the mock
func (m *MockCloudflareRuntime) Builder() registry.Builder {
	baseURL := m.Server.URL + "/"

	return registry.Builder(func() (types.Definition, error) {
		def, err := Builder(&RuntimeConfig{APIToken: mockAPIToken, AccountID: MockAccountID})()
		if err != nil {
			return types.Definition{}, err
		}

		connections := def.ConnectionList()
		for i := range connections {
			for name := range connections[i].Clients {
				connections[i].Clients[name] = func(context.Context, types.ConnectionInput) (any, error) {
					return mockCloudflareClient(baseURL), nil
				}
			}
		}

		def.Connections = lo.Map(connections, func(c types.Connection, _ int) types.Connector {
			return c
		})

		if def.RuntimeIntegration != nil {
			def.RuntimeIntegration.Build = func(_ context.Context, _ json.RawMessage) (any, error) {
				return mockCloudflareClient(baseURL), nil
			}
		}

		return def, nil
	})
}

// EmitDomainScanPoll emits a domain scan poll envelope for tests driving the poll saga
func EmitDomainScanPoll(ctx context.Context, g *gala.Gala, envelope DomainScanPollEnvelope) error {
	_, err := g.EmitWithHeaders(ctx, domainScanPollTopic.Name, envelope, gala.Headers{})

	return err
}
