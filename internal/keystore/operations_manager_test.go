package keystore

import (
	"context"
	"errors"
	"testing"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/models"
)

func TestOperationManagerRunUsesStoredCredential(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	payload := types.CredentialPayload{
		Provider: provider,
		Kind:     types.CredentialKindAPIKey,
		Data:     models.CredentialSet{APIToken: "stored"},
	}

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
	if captured.Credential.Data.APIToken != payload.Data.APIToken {
		t.Fatalf("expected stored credential, got %s", captured.Credential.Data.APIToken)
	}
	if source.getCount != 1 || source.mintCount != 0 {
		t.Fatalf("expected one Get call, zero Mint calls; got %d/%d", source.getCount, source.mintCount)
	}
}

func TestOperationManagerRunForceRefresh(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	stored := types.CredentialPayload{
		Provider: provider,
		Kind:     types.CredentialKindAPIKey,
		Data:     models.CredentialSet{APIToken: "stored"},
	}
	minted := types.CredentialPayload{
		Provider: provider,
		Kind:     types.CredentialKindAPIKey,
		Data:     models.CredentialSet{APIToken: "minted"},
	}

	source := &credentialSourceStub{
		getPayload:  stored,
		mintPayload: minted,
	}

	var captured types.CredentialPayload
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
	if captured.Data.APIToken != minted.Data.APIToken {
		t.Fatalf("expected minted credential, got %s", captured.Data.APIToken)
	}
	if result.Status != types.OperationStatusUnknown {
		t.Fatalf("expected status defaulted to unknown, got %s", result.Status)
	}
}

func TestOperationManagerRunRequiresClientManager(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	source := &credentialSourceStub{
		getPayload: types.CredentialPayload{
			Provider: provider,
			Kind:     types.CredentialKindAPIKey,
			Data:     models.CredentialSet{APIToken: "token"},
		},
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
		getPayload: types.CredentialPayload{
			Provider: provider,
			Kind:     types.CredentialKindAPIKey,
			Data:     models.CredentialSet{APIToken: "stored"},
		},
		mintPayload: types.CredentialPayload{
			Provider: provider,
			Kind:     types.CredentialKindAPIKey,
			Data:     models.CredentialSet{APIToken: "minted"},
		},
	}

	var builderRegion any
	clientDescriptor := types.ClientDescriptor{
		Provider: provider,
		Name:     types.ClientName("rest"),
		Build: func(_ context.Context, payload types.CredentialPayload, config map[string]any) (any, error) {
			if payload.Data.APIToken == "" {
				t.Fatalf("expected credential payload")
			}
			builderRegion = config["region"]
			config["region"] = "builder-mutated"
			return &pooledClient{id: payload.Data.APIToken}, nil
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
	result, err := manager.Run(context.Background(), types.OperationRequest{
		OrgID:       "org-1",
		Provider:    provider,
		Name:        opDescriptor.Name,
		Config:      reqConfig,
		ClientForce: true,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	client, ok := captured.Client.(*pooledClient)
	if !ok {
		t.Fatalf("expected pooled client type, got %T", captured.Client)
	}
	if client.id != source.mintPayload.Data.APIToken {
		t.Fatalf("expected client built from minted credential, got %s", client.id)
	}

	if source.mintCount == 0 {
		t.Fatalf("expected client force flag to mint credentials")
	}

	if captured.Config["region"] != "us-west-2" {
		t.Fatalf("expected operation config clone, got %v", captured.Config["region"])
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
