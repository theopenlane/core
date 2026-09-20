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
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// MockHTTPDefinitionID is the stable id for the mock HTTP provider that exercises the full
// connect/health/resolve/ingest flow against a real httptest server
var MockHTTPDefinitionID = types.NewDefinitionRef("def_01K0MOCKHTTP00000000000001")

// MockHTTPSyncOperation is the mock provider's directory sync operation, used to drive ingest through
// the installed definition
const MockHTTPSyncOperation = "mock.sync"

var (
	// mockHTTPSchema and MockHTTPCredential are the credential slot the mock provider authenticates with
	mockHTTPSchema, MockHTTPCredential = providerkit.CredentialSchema[mockHTTPCred]()
	// mockHTTPInstallation resolves installation metadata from the mock provider's instance endpoint
	mockHTTPInstallation = types.NewInstallationRef(resolveMockHTTPMetadata)
	// mockHTTPClient is the operation client that carries the stored credential into the ingest handler
	mockHTTPClient = types.NewClientRef[*mockHTTPClientInstance]()
)

// mockHTTPCred carries the bearer token and base URL of the mock provider the installation connects to
type mockHTTPCred struct {
	// Token is the bearer token the mock provider validates
	Token string `json:"token"`
	// BaseURL is the root URL of the mock provider, an httptest server in tests
	BaseURL string `json:"baseUrl"`
}

// MockHTTPCredentialSet builds the mock provider credential payload carrying the token and base URL
func MockHTTPCredentialSet(token, baseURL string) types.CredentialSet {
	raw, err := json.Marshal(mockHTTPCred{Token: token, BaseURL: baseURL})
	if err != nil {
		panic(err)
	}

	return types.CredentialSet{Data: raw}
}

// mockHTTPClientInstance is the mock provider operation client; the ingest handler reads its
// connection from the request credentials, so the client itself carries no state and exists only to
// bind the stored credential onto the operation request through the runtime's client resolution
type mockHTTPClientInstance struct{}

// buildMockHTTPClient validates the installation's stored credential resolves and returns the stateless
// mock provider client
func buildMockHTTPClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
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

// InstallationIdentity implements types.InstallationIdentifiable
func (m MockHTTPInstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{ExternalID: m.InstanceID}
}

// resolveMockHTTPMetadata fetches the external instance id from the mock provider using the stored credential
func resolveMockHTTPMetadata(ctx context.Context, req types.InstallationRequest) (MockHTTPInstallationMetadata, bool, error) {
	cred, ok, err := MockHTTPCredential.Resolve(req.Credentials)
	if err != nil || !ok {
		return MockHTTPInstallationMetadata{}, false, err
	}

	body, err := mockHTTPGet(ctx, cred.BaseURL+"/instance", cred.Token)
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

// mockHTTPHealthCheck validates the installation's credential against the mock provider's health endpoint
func mockHTTPHealthCheck(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
	cred, ok, err := MockHTTPCredential.Resolve(req.Credentials)
	if err != nil || !ok {
		return nil, ErrMockHTTPUnhealthy
	}

	if _, err := mockHTTPGet(ctx, cred.BaseURL+"/health", cred.Token); err != nil {
		return nil, err
	}

	return json.RawMessage(`{"ok":true}`), nil
}

// mockHTTPGet performs an authenticated GET against the mock provider, returning the body on a 200
// and an error otherwise
func mockHTTPGet(ctx context.Context, url, token string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+token)

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

// mockHTTPDirectory is the directory sync payload the mock provider serves and the ingest handler decodes
type mockHTTPDirectory struct {
	// Accounts are the raw directory account records the provider reports
	Accounts []json.RawMessage `json:"accounts"`
	// Groups are the raw directory group records the provider reports
	Groups []json.RawMessage `json:"groups"`
	// Memberships are the raw directory membership records the provider reports
	Memberships []json.RawMessage `json:"memberships"`
}

// mockHTTPIngest fetches the mock provider directory under the installation's credential and returns the
// account, group, and membership payload sets for the ingest pipeline to map, link, and persist
func mockHTTPIngest(ctx context.Context, req types.OperationRequest) ([]types.IngestPayloadSet, error) {
	cred, ok, err := MockHTTPCredential.Resolve(req.Credentials)
	if err != nil || !ok {
		return nil, ErrMockHTTPUnhealthy
	}

	body, err := mockHTTPGet(ctx, cred.BaseURL+"/directory", cred.Token)
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

// mockHTTPEnvelopes wraps each raw provider record as a mapping envelope for one payload set
func mockHTTPEnvelopes(records []json.RawMessage) []types.MappingEnvelope {
	return lo.Map(records, func(record json.RawMessage, _ int) types.MappingEnvelope {
		return types.MappingEnvelope{Payload: record}
	})
}

// MockHTTPBuilder returns the mock HTTP provider definition; a single credentialed connection with a
// health check and a metadata resolver and no operations, so an installation connects and comes up
// healthy without seeding reconcile loops
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
				{
					Ref:         MockHTTPCredential.ID(),
					Name:        "Mock HTTP Token",
					Description: "Bearer token and base URL the mock provider validates.",
					Schema:      mockHTTPSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  MockHTTPCredential.ID(),
					Name:           "Mock HTTP",
					Description:    "Connect to the mock HTTP provider and resolve its instance id.",
					CredentialRefs: []types.CredentialSlotID{MockHTTPCredential.ID()},
					HealthCheck:    &types.HealthCheckRegistration{Handle: mockHTTPHealthCheck},
					Integration:    mockHTTPInstallation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: MockHTTPCredential.ID(),
						Description:   "Remove the persisted mock provider credential and disconnect this installation.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            mockHTTPClient.ID(),
					CredentialRefs: []types.CredentialSlotID{MockHTTPCredential.ID()},
					Description:    "Mock provider client built from the stored token and base URL credential.",
					Build:          buildMockHTTPClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         MockHTTPSyncOperation,
					Description:  "Directory sync ingest for the mock provider",
					Topic:        MockHTTPDefinitionID.OperationTopic(MockHTTPSyncOperation),
					ClientRef:    mockHTTPClient.ID(),
					Policy:       types.ExecutionPolicy{Snapshot: true},
					IngestHandle: mockHTTPIngest,
					Ingest: []types.IngestContract{
						{Schema: entityops.SchemaDirectoryAccount.Name},
						{Schema: entityops.SchemaDirectoryGroup.Name},
						{Schema: entityops.SchemaDirectoryMembership.Name},
					},
				},
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

// MockHTTPServer is a concurrency-safe httptest-backed mock provider whose reported instance id and
// directory records can be mutated at runtime, so a test drives connect, health, reconnect, and ingest
// flows by changing what the provider reports rather than by writing installation state directly
type MockHTTPServer struct {
	// server is the running httptest server serving the mock provider endpoints
	server *httptest.Server
	// mu guards the mutable instance id and directory records against the server's concurrent handlers
	mu sync.Mutex
	// token is the bearer token the mock provider validates
	token string
	// instanceID is the external instance id the mock provider reports from its instance endpoint
	instanceID string
	// accounts are the directory account records the mock provider serves
	accounts []json.RawMessage
	// groups are the directory group records the mock provider serves
	groups []json.RawMessage
	// memberships are the directory membership records the mock provider serves
	memberships []json.RawMessage
}

// NewMockHTTPServer starts a mock provider httptest server that authenticates the bearer token and
// serves its health, instance, and directory endpoints from mutable state
func NewMockHTTPServer(token, instanceID string) *MockHTTPServer {
	server := &MockHTTPServer{token: token, instanceID: instanceID}
	server.server = httptest.NewServer(http.HandlerFunc(server.handle))

	return server
}

// handle routes one mock provider request, rejecting a bad bearer token and serving the health,
// instance, and directory endpoints from a snapshot taken under the mutex
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
		_ = json.NewEncoder(w).Encode(map[string]string{"instanceId": instanceID})
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

// SetInstanceID updates the instance id the mock provider reports from its instance endpoint
func (m *MockHTTPServer) SetInstanceID(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.instanceID = id
}

// SetDirectory replaces the account, group, and membership records the mock provider serves
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
