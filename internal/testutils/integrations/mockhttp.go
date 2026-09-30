//go:build test

package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// MockHTTPDefinitionID is the id of the mock HTTP provider definition
var MockHTTPDefinitionID = types.NewDefinitionRef("def_01K0MOCKHTTP00000000000001")

// MockHTTPSyncOperation is the mock provider's directory sync operation name
const MockHTTPSyncOperation = "mock.sync"

var (
	// MockHTTPCredential is the credential slot the mock provider authenticates with
	MockHTTPCredential = types.CredentialRefOf[mockHTTPCred]()
	// mockHTTPInstallation is the installation metadata resolver
	mockHTTPInstallation = types.NewInstallationRef(resolveMockHTTPMetadata)
	// mockHTTPClient is the client ref for the mock provider
	mockHTTPClient = types.ClientRefOf[*mockHTTPClientInstance]().Using(MockHTTPCredential)
	// mockHTTPConnection is the mock provider connection mode
	mockHTTPConnection = types.NewConnectionRef(MockHTTPCredential)
	// mockHTTPSyncOp is the mock provider's directory sync operation
	mockHTTPSyncOp = types.NewOperationRef[mockHTTPSync](MockHTTPSyncOperation).Ingests(mockHTTPClient, mockHTTPIngest)
)

// mockHTTPSync is the config for the mock provider's directory sync operation
type mockHTTPSync struct{}

// mockHTTPCred is the mock provider's bearer token and base URL
type mockHTTPCred struct {
	// Token is the bearer token the mock provider validates
	Token string `json:"token"`
	// BaseURL is the root URL of the mock provider
	BaseURL string `json:"baseUrl"`
}

// MockHTTPCredentialSet returns the mock provider credential payload
func MockHTTPCredentialSet(token, baseURL string) types.CredentialSet {
	return credentialSet(mockHTTPCred{Token: token, BaseURL: baseURL})
}

// mockHTTPClientInstance is the stateless mock provider operation client
type mockHTTPClientInstance struct{}

// buildMockHTTPClient returns the mock provider client if the credential resolves
func buildMockHTTPClient(_ context.Context, req types.ClientBuildRequest) (*mockHTTPClientInstance, error) {
	if _, ok, err := MockHTTPCredential.Resolve(req.Credentials); err != nil || !ok {
		return nil, ErrMockHTTPUnhealthy
	}

	return &mockHTTPClientInstance{}, nil
}

// MockHTTPInstallationMetadata is the identity the mock provider reports for an installation
type MockHTTPInstallationMetadata struct {
	// InstanceID is the external instance id the mock provider returns
	InstanceID string `json:"instanceId,omitempty"`
}

// InstallationIdentity returns the installation identity
func (m MockHTTPInstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{ExternalID: m.InstanceID}
}

// resolveMockHTTPMetadata returns the mock provider's external instance id
func resolveMockHTTPMetadata(ctx context.Context, req types.InstallationRequest) (MockHTTPInstallationMetadata, bool, error) {
	body, err := mockHTTPGet(ctx, req.Credentials, "/instance")
	if err != nil {
		return MockHTTPInstallationMetadata{}, false, err
	}

	var metadata MockHTTPInstallationMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return MockHTTPInstallationMetadata{}, false, ErrMockHTTPDecode
	}

	if metadata.InstanceID == "" {
		return MockHTTPInstallationMetadata{}, false, nil
	}

	return metadata, true, nil
}

// mockHTTPHealthCheck returns the mock provider's health check result
func mockHTTPHealthCheck(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
	if _, err := mockHTTPGet(ctx, req.Credentials, "/health"); err != nil {
		return nil, err
	}

	return json.RawMessage(`{"ok":true}`), nil
}

// mockHTTPGet returns the response body of a GET to path on the mock provider
func mockHTTPGet(ctx context.Context, credentials types.CredentialBindings, path string) ([]byte, error) {
	cred, ok, err := MockHTTPCredential.Resolve(credentials)
	if err != nil || !ok {
		return nil, ErrMockHTTPUnhealthy
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, cred.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+cred.Token)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrMockHTTPUnhealthy, response.StatusCode)
	}

	return io.ReadAll(response.Body)
}

// mockHTTPDirectory is the directory sync payload the mock provider serves
type mockHTTPDirectory struct {
	// Accounts are the raw directory account records the provider reports
	Accounts []json.RawMessage `json:"accounts"`
	// Groups are the raw directory group records the provider reports
	Groups []json.RawMessage `json:"groups"`
	// Memberships are the raw directory membership records the provider reports
	Memberships []json.RawMessage `json:"memberships"`
}

// mockHTTPIngest returns the mock provider's directory as ingest payload sets
func mockHTTPIngest(ctx context.Context, req types.OperationRequest, _ *mockHTTPClientInstance, _ mockHTTPSync) ([]types.IngestPayloadSet, error) {
	body, err := mockHTTPGet(ctx, req.Credentials, "/directory")
	if err != nil {
		return nil, err
	}

	var directory mockHTTPDirectory
	if err := json.Unmarshal(body, &directory); err != nil {
		return nil, ErrMockHTTPDecode
	}

	return []types.IngestPayloadSet{
		{Schema: entityops.SchemaDirectoryAccount.Name, Envelopes: mockHTTPEnvelopes(directory.Accounts), SnapshotComplete: true},
		{Schema: entityops.SchemaDirectoryGroup.Name, Envelopes: mockHTTPEnvelopes(directory.Groups), SnapshotComplete: true},
		{Schema: entityops.SchemaDirectoryMembership.Name, Envelopes: mockHTTPEnvelopes(directory.Memberships), SnapshotComplete: true},
	}, nil
}

// mockHTTPEnvelopes returns records wrapped as mapping envelopes
func mockHTTPEnvelopes(records []json.RawMessage) []types.MappingEnvelope {
	return lo.Map(records, func(record json.RawMessage, _ int) types.MappingEnvelope {
		return types.MappingEnvelope{Payload: record}
	})
}

// MockHTTPBuilder returns the mock HTTP provider definition
func MockHTTPBuilder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          MockHTTPDefinitionID.ID(),
				Family:      "Openlane",
				DisplayName: "Mock HTTP Provider",
				Description: "Test provider backed by an httptest server for connect/health/resolve flows.",
				Category:    "system",
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				MockHTTPCredential.Registration(types.CredentialRegistration{
					Name:        "Mock HTTP Token",
					Description: "Bearer token and base URL the mock provider validates.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				mockHTTPConnection.Registration(types.ConnectionRegistration{
					Name:        "Mock HTTP",
					Description: "Connect to the mock HTTP provider and resolve its instance id.",
					Disconnect: &types.DisconnectRegistration{
						Description: "Remove the persisted mock provider credential and disconnect this installation.",
					},
				}),
			},
			HealthCheck:  types.CredentialHealthCheck(mockHTTPHealthCheck),
			Installation: mockHTTPInstallation.Registration(),
			Clients: []types.ClientRegistration{
				mockHTTPClient.Registration(buildMockHTTPClient, types.ClientRegistration{
					Description: "Mock provider client built from the stored token and base URL credential.",
				}),
			},
			Operations: []types.OperationRegistration{
				mockHTTPSyncOp.Registration(MockHTTPDefinitionID, types.OperationRegistration{
					Description: "Directory sync ingest for the mock provider",
					Policy:      types.ExecutionPolicy{Snapshot: true},
					Ingest: []types.IngestContract{
						{Schema: entityops.SchemaDirectoryAccount.Name},
						{Schema: entityops.SchemaDirectoryGroup.Name},
						{Schema: entityops.SchemaDirectoryMembership.Name},
					},
				}),
			},
			Mappings: []types.MappingRegistration{
				{Schema: entityops.SchemaDirectoryAccount.Name, Spec: types.MappingOverride{MapExpr: "payload"}},
				{Schema: entityops.SchemaDirectoryGroup.Name, Spec: types.MappingOverride{MapExpr: "payload"}},
				{Schema: entityops.SchemaDirectoryMembership.Name, Spec: types.MappingOverride{
					MapExpr: "payload",
					Links: []types.LinkRule{
						{TargetSchema: entityops.SchemaDirectoryAccount.Name, TargetField: "external_id", SourceField: "directory_account_id"},
						{TargetSchema: entityops.SchemaDirectoryGroup.Name, TargetField: "external_id", SourceField: "directory_group_id"},
					},
				}},
			},
		}, nil
	})
}

// MockHTTPServer is a mutable httptest-backed mock provider server
type MockHTTPServer struct {
	// server is the running httptest server
	server *httptest.Server
	// mu guards the mutable fields
	mu sync.Mutex
	// token is the bearer token the mock provider validates
	token string
	// instanceID is the external instance id the mock provider reports
	instanceID string
	// accounts are the served directory account records
	accounts []json.RawMessage
	// groups are the served directory group records
	groups []json.RawMessage
	// memberships are the served directory membership records
	memberships []json.RawMessage
}

// NewMockHTTPServer returns a running mock provider httptest server
func NewMockHTTPServer(token, instanceID string) *MockHTTPServer {
	server := &MockHTTPServer{token: token, instanceID: instanceID}
	server.server = httptest.NewServer(http.HandlerFunc(server.handle))

	return server
}

// handle routes one mock provider request
func (m *MockHTTPServer) handle(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	token := m.token
	instanceID := m.instanceID
	directory := mockHTTPDirectory{Accounts: m.accounts, Groups: m.groups, Memberships: m.memberships}
	m.mu.Unlock()

	if r.Header.Get("Authorization") != "Bearer "+token {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	switch r.URL.Path {
	case "/health":
		w.WriteHeader(http.StatusOK)
	case "/instance":
		_ = json.NewEncoder(w).Encode(MockHTTPInstallationMetadata{InstanceID: instanceID})
	case "/directory":
		_ = json.NewEncoder(w).Encode(directory)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// URL returns the base URL of the running mock provider server
func (m *MockHTTPServer) URL() string {
	return m.server.URL
}

// SetInstanceID sets the instance id the mock provider reports
func (m *MockHTTPServer) SetInstanceID(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.instanceID = id
}

// SetDirectory replaces the mock provider's directory records
func (m *MockHTTPServer) SetDirectory(accounts, groups, memberships []json.RawMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accounts = accounts
	m.groups = groups
	m.memberships = memberships
}

// Close shuts down the running mock provider server
func (m *MockHTTPServer) Close() {
	m.server.Close()
}
