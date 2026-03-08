package keystore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

func TestOperationManagerRunUsesStoredCredential(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	payload := models.CredentialSet{APIToken: "stored"}

	source := &credentialSourceStub{getPayload: payload}

	var captured types.OperationInput
	descriptor := types.OperationDescriptor{
		Provider: provider,
		Name:     types.OperationName("health"),
		Run: func(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
			captured = input
			return types.OperationResult{Status: types.OperationStatusOK, Summary: "ok"}, nil
		},
	}

	manager, err := NewOperationManager(source, []types.OperationDescriptor{descriptor})
	if err != nil {
		t.Fatalf("NewOperationManager() error = %v", err)
	}

	result, err := manager.Run(context.Background(), types.OperationRequest{
		OrgID:    "org-1",
		Provider: provider,
		Name:     descriptor.Name,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Status != types.OperationStatusOK {
		t.Fatalf("expected status ok, got %s", result.Status)
	}
	if captured.Credential.APIToken != payload.APIToken {
		t.Fatalf("expected stored credential, got %s", captured.Credential.APIToken)
	}
	if source.getCount != 1 || source.mintCount != 0 {
		t.Fatalf("expected one Get call, zero Mint calls; got %d/%d", source.getCount, source.mintCount)
	}
}

func TestOperationManagerRunForceRefresh(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	stored := models.CredentialSet{APIToken: "stored"}
	minted := models.CredentialSet{APIToken: "minted"}

	source := &credentialSourceStub{
		getPayload:  stored,
		mintPayload: minted,
	}

	var captured models.CredentialSet
	descriptor := types.OperationDescriptor{
		Provider: provider,
		Name:     types.OperationName("refresh"),
		Run: func(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
			captured = input.Credential
			return types.OperationResult{}, nil
		},
	}

	manager, err := NewOperationManager(source, []types.OperationDescriptor{descriptor})
	if err != nil {
		t.Fatalf("NewOperationManager() error = %v", err)
	}

	result, err := manager.Run(context.Background(), types.OperationRequest{
		OrgID:    "org-1",
		Provider: provider,
		Name:     descriptor.Name,
		Force:    true,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if source.mintCount != 1 || source.getCount != 0 {
		t.Fatalf("expected one Mint call and zero Get calls, got %d/%d", source.mintCount, source.getCount)
	}
	if captured.APIToken != minted.APIToken {
		t.Fatalf("expected minted credential, got %s", captured.APIToken)
	}
	if result.Status != types.OperationStatusUnknown {
		t.Fatalf("expected status defaulted to unknown, got %s", result.Status)
	}
}

func TestOperationManagerRunUsesIntegrationScopedCredential(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	source := &credentialSourceStub{
		getPayload:               models.CredentialSet{APIToken: "default"},
		getForIntegrationPayload: models.CredentialSet{APIToken: "scoped"},
	}

	var captured types.OperationInput
	descriptor := types.OperationDescriptor{
		Provider: provider,
		Name:     types.OperationName("notify"),
		Run: func(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
			captured = input
			return types.OperationResult{Status: types.OperationStatusOK}, nil
		},
	}

	manager, err := NewOperationManager(source, []types.OperationDescriptor{descriptor})
	if err != nil {
		t.Fatalf("NewOperationManager() error = %v", err)
	}

	_, err = manager.Run(context.Background(), types.OperationRequest{
		OrgID:         "org-1",
		IntegrationID: "int-1",
		Provider:      provider,
		Name:          descriptor.Name,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if source.getForIntegrationCount != 1 || source.getCount != 0 {
		t.Fatalf("expected integration scoped Get call, got scoped=%d plain=%d", source.getForIntegrationCount, source.getCount)
	}
	if source.lastGetIntegrationID != "int-1" {
		t.Fatalf("expected integration id int-1, got %s", source.lastGetIntegrationID)
	}
	if captured.Credential.APIToken != "scoped" {
		t.Fatalf("expected scoped credential payload, got %s", captured.Credential.APIToken)
	}
}

func TestOperationManagerRunForceRefreshUsesIntegrationScopedMint(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	source := &credentialSourceStub{
		mintPayload:               models.CredentialSet{APIToken: "default"},
		mintForIntegrationPayload: models.CredentialSet{APIToken: "scoped-minted"},
	}

	var captured models.CredentialSet
	descriptor := types.OperationDescriptor{
		Provider: provider,
		Name:     types.OperationName("refresh"),
		Run: func(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
			captured = input.Credential
			return types.OperationResult{Status: types.OperationStatusOK}, nil
		},
	}

	manager, err := NewOperationManager(source, []types.OperationDescriptor{descriptor})
	if err != nil {
		t.Fatalf("NewOperationManager() error = %v", err)
	}

	_, err = manager.Run(context.Background(), types.OperationRequest{
		OrgID:         "org-1",
		IntegrationID: "int-2",
		Provider:      provider,
		Name:          descriptor.Name,
		Force:         true,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if source.mintForIntegrationCount != 1 || source.mintCount != 0 {
		t.Fatalf("expected integration scoped Mint call, got scoped=%d plain=%d", source.mintForIntegrationCount, source.mintCount)
	}
	if source.lastMintIntegrationID != "int-2" {
		t.Fatalf("expected integration id int-2, got %s", source.lastMintIntegrationID)
	}
	if captured.APIToken != "scoped-minted" {
		t.Fatalf("expected scoped minted credential, got %s", captured.APIToken)
	}
}

func TestOperationManagerRunRequiresClientManager(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	source := &credentialSourceStub{
		getPayload: models.CredentialSet{APIToken: "token"},
	}

	descriptor := types.OperationDescriptor{
		Provider: provider,
		Name:     types.OperationName("uses-client"),
		Client:   types.ClientName("rest"),
		Run: func(context.Context, types.OperationInput) (types.OperationResult, error) {
			return types.OperationResult{Status: types.OperationStatusOK}, nil
		},
	}

	manager, err := NewOperationManager(source, []types.OperationDescriptor{descriptor})
	if err != nil {
		t.Fatalf("NewOperationManager() error = %v", err)
	}

	_, err = manager.Run(context.Background(), types.OperationRequest{
		OrgID:    "org-1",
		Provider: provider,
		Name:     descriptor.Name,
	})
	if !errors.Is(err, ErrOperationClientManagerRequired) {
		t.Fatalf("expected ErrOperationClientManagerRequired, got %v", err)
	}
}

func TestOperationManagerRunResolvesClientAndConfig(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	source := &credentialSourceStub{
		getPayload:  models.CredentialSet{APIToken: "stored"},
		mintPayload: models.CredentialSet{APIToken: "minted"},
	}

	var builderRegion any
	clientDescriptor := types.ClientDescriptor{
		Provider: provider,
		Name:     types.ClientName("rest"),
		Build: func(_ context.Context, payload models.CredentialSet, config json.RawMessage) (types.ClientInstance, error) {
			if payload.APIToken == "" {
				t.Fatalf("expected credential payload")
			}
			decoded := map[string]any{}
			if len(config) > 0 {
				_ = json.Unmarshal(config, &decoded)
			}
			builderRegion = decoded["region"]
			return types.NewClientInstance(&pooledClient{id: payload.APIToken}), nil
		},
	}

	clientManager, err := NewClientPoolManager(source, []types.ClientDescriptor{clientDescriptor})
	if err != nil {
		t.Fatalf("NewClientPoolManager() error = %v", err)
	}

	var captured types.OperationInput
	opDescriptor := types.OperationDescriptor{
		Provider: provider,
		Name:     types.OperationName("with-client"),
		Client:   clientDescriptor.Name,
		Run: func(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
			captured = input
			return types.OperationResult{Status: types.OperationStatusOK}, nil
		},
	}

	manager, err := NewOperationManager(source, []types.OperationDescriptor{opDescriptor},
		WithOperationClients(clientManager))
	if err != nil {
		t.Fatalf("NewOperationManager() error = %v", err)
	}

	reqConfig := map[string]any{"region": "us-west-2"}
	var reqConfigDoc json.RawMessage
	if err := jsonx.RoundTrip(reqConfig, &reqConfigDoc); err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}

	result, err := manager.Run(context.Background(), types.OperationRequest{
		OrgID:       "org-1",
		Provider:    provider,
		Name:        opDescriptor.Name,
		Config:      reqConfigDoc,
		ClientForce: true,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	client, ok := types.ClientInstanceAs[*pooledClient](captured.Client)
	if !ok {
		t.Fatalf("expected pooled client type, got %T", captured.Client)
	}
	if client.id != source.mintPayload.APIToken {
		t.Fatalf("expected client built from minted credential, got %s", client.id)
	}

	if source.mintCount == 0 {
		t.Fatalf("expected client force flag to mint credentials")
	}

	capturedConfig, err := jsonx.ToMap(captured.Config)
	if err != nil {
		t.Fatalf("expected decodable operation config, got %v", err)
	}
	if capturedConfig["region"] != "us-west-2" {
		t.Fatalf("expected operation config clone, got %v", capturedConfig["region"])
	}
	if reqConfig["region"] != "us-west-2" {
		t.Fatalf("expected request config to remain unchanged, got %v", reqConfig["region"])
	}
	if builderRegion != "us-west-2" {
		t.Fatalf("expected client builder to receive cloned config, got %v", builderRegion)
	}
	if result.Status != types.OperationStatusOK {
		t.Fatalf("expected status ok, got %s", result.Status)
	}
}

func TestOperationManagerRunWithCredentialUsesProvidedCredential(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	stored := models.CredentialSet{APIToken: "stored"}
	provided := models.CredentialSet{APIToken: "provided"}

	source := &credentialSourceStub{getPayload: stored, mintPayload: stored}

	var captured types.OperationInput
	descriptor := types.OperationDescriptor{
		Provider: provider,
		Name:     types.OperationName("health"),
		Run: func(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
			captured = input
			return types.OperationResult{Status: types.OperationStatusOK}, nil
		},
	}

	manager, err := NewOperationManager(source, []types.OperationDescriptor{descriptor})
	if err != nil {
		t.Fatalf("NewOperationManager() error = %v", err)
	}

	result, err := manager.RunWithCredential(context.Background(), types.OperationRequest{
		OrgID:    "org-1",
		Provider: provider,
		Name:     descriptor.Name,
	}, provided)
	if err != nil {
		t.Fatalf("RunWithCredential() error = %v", err)
	}

	if result.Status != types.OperationStatusOK {
		t.Fatalf("expected status ok, got %s", result.Status)
	}
	if captured.Credential.APIToken != provided.APIToken {
		t.Fatalf("expected provided credential, got %s", captured.Credential.APIToken)
	}
	if source.getCount != 0 || source.mintCount != 0 {
		t.Fatalf("expected no credential store calls; got get=%d mint=%d", source.getCount, source.mintCount)
	}
}

func TestOperationManagerRunWithCredentialValidatesRequest(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	source := &credentialSourceStub{}
	payload := models.CredentialSet{}

	manager, err := NewOperationManager(source, nil)
	if err != nil {
		t.Fatalf("NewOperationManager() error = %v", err)
	}

	tests := []struct {
		name    string
		req     types.OperationRequest
		wantErr error
	}{
		{
			name:    "missing org id",
			req:     types.OperationRequest{Provider: provider, Name: "health"},
			wantErr: ErrOrgIDRequired,
		},
		{
			name:    "missing provider",
			req:     types.OperationRequest{OrgID: "org-1", Name: "health"},
			wantErr: ErrProviderRequired,
		},
		{
			name:    "missing operation name",
			req:     types.OperationRequest{OrgID: "org-1", Provider: provider},
			wantErr: ErrOperationNameRequired,
		},
		{
			name:    "operation not registered",
			req:     types.OperationRequest{OrgID: "org-1", Provider: provider, Name: "unknown"},
			wantErr: ErrOperationNotRegistered,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := manager.RunWithCredential(context.Background(), tc.req, payload)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}
