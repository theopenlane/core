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
	// MockHTTPConnection is the connection the mock provider authenticates with
	MockHTTPConnection = types.ConnectionOf[mockHTTPCred]()
	// mockHTTPInstallation is the installation metadata layout of the mock provider
	mockHTTPInstallation = types.InstallationOf[MockHTTPInstallationMetadata]()
	// mockHTTPSyncOp is the mock provider's directory sync operation
	mockHTTPSyncOp = types.NewOperationRef[mockHTTPSync](MockHTTPSyncOperation).
			Ingests(mockHTTPIngest).
			Policy(types.ExecutionPolicy{Snapshot: true}).
			Ingest(
			types.IngestContract{Schema: entityops.SchemaDirectoryAccount.Name},
			types.IngestContract{Schema: entityops.SchemaDirectoryGroup.Name},
			types.IngestContract{Schema: entityops.SchemaDirectoryMembership.Name},
		)
)

// mockHTTPSync is the config for the mock provider's directory sync operation
type mockHTTPSync struct {
	types.OperationSettings
}

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

// mockHTTPClientInstance is the mock provider operation client
type mockHTTPClientInstance struct {
	// cred is the bearer token and base URL the client requests with
	cred mockHTTPCred
}

// buildMockHTTPClient returns the mock provider client carrying the decoded credential
func buildMockHTTPClient(_ context.Context, req types.ConnectionRequest[mockHTTPCred]) (*mockHTTPClientInstance, error) {
	return &mockHTTPClientInstance{cred: req.Credential}, nil
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

// verifyMockHTTP returns the mock provider's external instance id
func verifyMockHTTP(ctx context.Context, req types.ConnectionRequest[mockHTTPCred], _ *mockHTTPClientInstance) (MockHTTPInstallationMetadata, error) {
	body, err := mockHTTPGet(ctx, req.Credential, "/instance")
	if err != nil {
		return MockHTTPInstallationMetadata{}, err
	}

	var metadata MockHTTPInstallationMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return MockHTTPInstallationMetadata{}, ErrMockHTTPDecode
	}

	return metadata, nil
}

// mockHTTPGet returns the response body of a GET to path on the mock provider
func mockHTTPGet(ctx context.Context, cred mockHTTPCred, path string) ([]byte, error) {
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
func mockHTTPIngest(ctx context.Context, _ types.OperationRequest, client *mockHTTPClientInstance, _ mockHTTPSync) ([]types.IngestPayloadSet, error) {
	body, err := mockHTTPGet(ctx, client.cred, "/directory")
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
			Installation: mockHTTPInstallation.Registration(),
			Connections: []types.Connector{
				MockHTTPConnection.
					Name("Mock HTTP").
					Description("Connect to the mock HTTP provider and resolve its instance id.").
					Provides(buildMockHTTPClient).
					Verified(verifyMockHTTP).
					Disconnects("Remove the persisted mock provider credential and disconnect this installation.", nil),
			},
			Operations: []types.OperationRegistration{
				mockHTTPSyncOp.Description("Directory sync ingest for the mock provider").Registration(),
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
