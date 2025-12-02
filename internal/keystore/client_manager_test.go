package keystore

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/theopenlane/shared/integrations/types"
	"github.com/theopenlane/shared/models"
)

type pooledClient struct {
	id string
}

func TestClientPoolManagerGetReusesClients(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	payload := types.CredentialPayload{
		Provider: provider,
		Kind:     types.CredentialKindAPIKey,
		Data: models.CredentialSet{
			AccessKeyID: "ak-1",
		},
	}

	source := &credentialSourceStub{getPayload: payload}

	var buildCount atomic.Int32
	descriptor := types.ClientDescriptor{
		Provider: provider,
		Name:     types.ClientName("rest"),
		Build: func(_ context.Context, cred types.CredentialPayload, config map[string]any) (any, error) {
			buildCount.Add(1)
			if cred.Data.AccessKeyID == "" {
				t.Fatalf("expected credential payload in builder")
			}
			if config != nil {
				config["region"] = "builder"
			}
			return &pooledClient{id: cred.Data.AccessKeyID}, nil
		},
	}

	manager, err := NewClientPoolManager(source, []types.ClientDescriptor{descriptor})
	if err != nil {
		t.Fatalf("NewClientPoolManager() error = %v", err)
	}

	ctx := context.Background()

	first, err := manager.Get(ctx, "org-1", provider, descriptor.Name)
	if err != nil {
		t.Fatalf("manager.Get() error = %v", err)
	}
	second, err := manager.Get(ctx, "org-1", provider, descriptor.Name)
	if err != nil {
		t.Fatalf("manager.Get() second error = %v", err)
	}

	if buildCount.Load() != 1 {
		t.Fatalf("expected builder to run once, ran %d times", buildCount.Load())
	}

	if first != second {
		t.Fatalf("expected pooled client to be reused")
	}

	reqConfig := map[string]any{"region": "us-west-2"}
	if _, err := manager.Get(ctx, "org-2", provider, descriptor.Name, WithClientConfig(reqConfig)); err != nil {
		t.Fatalf("manager.Get() with config error = %v", err)
	}

	if buildCount.Load() != 2 {
		t.Fatalf("expected builder to run for new org, ran %d times", buildCount.Load())
	}

	if got := reqConfig["region"]; got != "us-west-2" {
		t.Fatalf("expected caller config to remain unchanged, got %v", got)
	}

	if _, err := manager.Get(ctx, "org-1", provider, types.ClientName("missing")); !errors.Is(err, ErrClientNotRegistered) {
		t.Fatalf("expected ErrClientNotRegistered, got %v", err)
	}
}

func TestClientPoolManagerRegisterDescriptorValidation(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	payload := types.CredentialPayload{
		Provider: provider,
		Kind:     types.CredentialKindAPIKey,
		Data: models.CredentialSet{
			APIToken: "token",
		},
	}

	manager, err := NewClientPoolManager(&credentialSourceStub{getPayload: payload}, nil)
	if err != nil {
		t.Fatalf("NewClientPoolManager() error = %v", err)
	}

	tests := []struct {
		name       string
		descriptor types.ClientDescriptor
		wantErr    error
	}{
		{
			name: "missing provider",
			descriptor: types.ClientDescriptor{
				Name: types.ClientName("rest"),
				Build: func(context.Context, types.CredentialPayload, map[string]any) (any, error) {
					return nil, nil
				},
			},
			wantErr: ErrProviderRequired,
		},
		{
			name: "missing name",
			descriptor: types.ClientDescriptor{
				Provider: provider,
				Build: func(context.Context, types.CredentialPayload, map[string]any) (any, error) {
					return nil, nil
				},
			},
			wantErr: ErrClientDescriptorInvalid,
		},
		{
			name: "missing builder",
			descriptor: types.ClientDescriptor{
				Provider: provider,
				Name:     types.ClientName("rest"),
			},
			wantErr: ErrClientBuilderRequired,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := manager.RegisterDescriptor(tc.descriptor); !errors.Is(err, tc.wantErr) {
				t.Fatalf("RegisterDescriptor() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
