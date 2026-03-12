package keystore

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/theopenlane/core/internal/integrations/types"
)

type pooledClient struct {
	id string
}

func TestClientPoolManagerGetReusesClients(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	payload := types.CredentialSet{OAuthAccessToken: "ak-1"}

	source := &credentialSourceStub{getPayload: payload}

	var buildCount atomic.Int32
	descriptor := types.ClientDescriptor{
		Provider: provider,
		Name:     types.ClientName("rest"),
		Build: func(_ context.Context, cred types.CredentialSet, config json.RawMessage) (types.ClientInstance, error) {
			buildCount.Add(1)
			if cred.OAuthAccessToken == "" {
				t.Fatalf("expected credential payload in builder")
			}
			return types.NewClientInstance(&pooledClient{id: cred.OAuthAccessToken}), nil
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

	reqConfig, _ := json.Marshal(map[string]any{"region": "us-west-2"})
	if _, err := manager.Get(ctx, "org-2", provider, descriptor.Name, WithClientConfig(reqConfig)); err != nil {
		t.Fatalf("manager.Get() with config error = %v", err)
	}

	if buildCount.Load() != 2 {
		t.Fatalf("expected builder to run for new org, ran %d times", buildCount.Load())
	}

	if _, err := manager.Get(ctx, "org-1", provider, types.ClientName("missing")); !errors.Is(err, ErrClientNotRegistered) {
		t.Fatalf("expected ErrClientNotRegistered, got %v", err)
	}
}

func TestClientPoolManagerRegisterDescriptorValidation(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	payload := types.CredentialSet{OAuthAccessToken: "token"}

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
				Build: func(context.Context, types.CredentialSet, json.RawMessage) (types.ClientInstance, error) {
					return types.EmptyClientInstance(), nil
				},
			},
			wantErr: ErrProviderRequired,
		},
		{
			name: "missing name",
			descriptor: types.ClientDescriptor{
				Provider: provider,
				Build: func(context.Context, types.CredentialSet, json.RawMessage) (types.ClientInstance, error) {
					return types.EmptyClientInstance(), nil
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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := manager.RegisterDescriptor(tc.descriptor); !errors.Is(err, tc.wantErr) {
				t.Fatalf("RegisterDescriptor() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestClientPoolManagerBuildFromPayload(t *testing.T) {
	t.Parallel()

	provider := types.ProviderType("acme")
	payload := types.CredentialSet{OAuthAccessToken: "direct-token"}

	var captured types.CredentialSet
	var capturedConfig json.RawMessage
	descriptor := types.ClientDescriptor{
		Provider: provider,
		Name:     types.ClientName("rest"),
		Build: func(_ context.Context, cred types.CredentialSet, config json.RawMessage) (types.ClientInstance, error) {
			captured = cred
			capturedConfig = config
			return types.NewClientInstance(&pooledClient{id: cred.OAuthAccessToken}), nil
		},
	}

	manager, err := NewClientPoolManager(&credentialSourceStub{getPayload: payload}, []types.ClientDescriptor{descriptor})
	if err != nil {
		t.Fatalf("NewClientPoolManager() error = %v", err)
	}

	reqConfig, _ := json.Marshal(map[string]any{"region": "eu-west-1"})
	result, err := manager.BuildFromPayload(context.Background(), provider, descriptor.Name, payload, reqConfig)
	if err != nil {
		t.Fatalf("BuildFromPayload() error = %v", err)
	}

	client, ok := types.ClientInstanceAs[*pooledClient](result)
	if !ok {
		t.Fatalf("expected *pooledClient, got %T", result)
	}
	if client.id != payload.OAuthAccessToken {
		t.Fatalf("expected client built from provided payload, got %s", client.id)
	}
	if captured.OAuthAccessToken != payload.OAuthAccessToken {
		t.Fatalf("expected builder to receive provided payload, got %s", captured.OAuthAccessToken)
	}

	var decodedConfig map[string]any
	if err := json.Unmarshal(capturedConfig, &decodedConfig); err != nil {
		t.Fatalf("expected decodable config, got %v", err)
	}
	if decodedConfig["region"] != "eu-west-1" {
		t.Fatalf("expected config to be passed to builder, got %v", decodedConfig["region"])
	}

	// Verify original config is not mutated by builder (it's a separate copy)
	var origDecoded map[string]any
	if err := json.Unmarshal(reqConfig, &origDecoded); err != nil {
		t.Fatalf("expected original config to remain decodable, got %v", err)
	}
	if origDecoded["region"] != "eu-west-1" {
		t.Fatalf("expected caller config to remain unchanged, got %v", origDecoded["region"])
	}

	_, err = manager.BuildFromPayload(context.Background(), provider, types.ClientName("missing"), payload, nil)
	if !errors.Is(err, ErrClientNotRegistered) {
		t.Fatalf("expected ErrClientNotRegistered for unknown client, got %v", err)
	}
}
