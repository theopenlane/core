package types

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestCredentialBindingsResolve(t *testing.T) {
	t.Parallel()

	refA := NewCredentialSlotID("oauth")
	refB := NewCredentialSlotID("api_key")

	bindings := CredentialBindings{
		{Ref: refA, Credential: CredentialSet{Data: json.RawMessage(`{"token":"abc"}`)}},
		{Ref: refB, Credential: CredentialSet{Data: json.RawMessage(`{"key":"xyz"}`)}},
	}

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		lookup := NewCredentialSlotID("oauth")
		cred, ok := bindings.Resolve(lookup)
		if !ok {
			t.Fatal("expected binding to be found")
		}

		if string(cred.Data) != `{"token":"abc"}` {
			t.Fatalf("unexpected credential data: %s", cred.Data)
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		lookup := NewCredentialSlotID("nonexistent")
		_, ok := bindings.Resolve(lookup)
		if ok {
			t.Fatal("expected binding not to be found")
		}
	})
}

func TestDefinitionCredentialRegistration(t *testing.T) {
	t.Parallel()

	refA := NewCredentialSlotID("oauth")
	refB := NewCredentialSlotID("api_key")

	def := Definition{
		DefinitionSpec: DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []CredentialRegistration{
			{Ref: refA, Name: "OAuth"},
			{Ref: refB, Name: "API Key"},
		},
	}

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		lookup := NewCredentialSlotID("api_key")
		reg, err := def.CredentialRegistration(lookup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if reg.Name != "API Key" {
			t.Fatalf("got name %q, want %q", reg.Name, "API Key")
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		lookup := NewCredentialSlotID("missing")
		_, err := def.CredentialRegistration(lookup)
		if !errors.Is(err, ErrCredentialRefNotFound) {
			t.Fatalf("got error %v, want ErrCredentialRefNotFound", err)
		}
	})
}

func TestDefinitionConnectionRegistration(t *testing.T) {
	t.Parallel()

	refA := NewCredentialSlotID("oauth")
	refB := NewCredentialSlotID("api_key")

	def := Definition{
		DefinitionSpec: DefinitionSpec{ID: "test-def"},
		Connections: []ConnectionRegistration{
			{CredentialRef: refA, Name: "OAuth Flow"},
			{CredentialRef: refB, Name: "API Key Flow"},
		},
	}

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		lookup := NewCredentialSlotID("oauth")
		reg, err := def.ConnectionRegistration(lookup)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if reg.Name != "OAuth Flow" {
			t.Fatalf("got name %q, want %q", reg.Name, "OAuth Flow")
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		lookup := NewCredentialSlotID("missing")
		_, err := def.ConnectionRegistration(lookup)
		if !errors.Is(err, ErrConnectionRefNotFound) {
			t.Fatalf("got error %v, want ErrConnectionRefNotFound", err)
		}
	})
}

func TestDefinitionProviderState(t *testing.T) {
	t.Parallel()

	def := Definition{
		DefinitionSpec: DefinitionSpec{ID: "github_app"},
	}

	t.Run("nil providers map", func(t *testing.T) {
		t.Parallel()

		state := IntegrationProviderState{Providers: nil}
		got, err := def.ProviderState(state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.CredentialRef.String() != "" {
			t.Fatalf("expected zero DefinitionProviderState, got credential ref %q", got.CredentialRef.String())
		}
	})

	t.Run("missing key", func(t *testing.T) {
		t.Parallel()

		state := IntegrationProviderState{
			Providers: map[string]json.RawMessage{
				"other_provider": json.RawMessage(`{"credentialRef":"oauth"}`),
			},
		}

		got, err := def.ProviderState(state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.CredentialRef.String() != "" {
			t.Fatalf("expected zero DefinitionProviderState, got credential ref %q", got.CredentialRef.String())
		}
	})

	t.Run("valid state", func(t *testing.T) {
		t.Parallel()

		state := IntegrationProviderState{
			Providers: map[string]json.RawMessage{
				"github_app": json.RawMessage(`{"credentialRef":"oauth"}`),
			},
		}

		got, err := def.ProviderState(state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.CredentialRef.String() != "oauth" {
			t.Fatalf("got credential ref %q, want %q", got.CredentialRef.String(), "oauth")
		}
	})
}

func TestDefinitionWithProviderState(t *testing.T) {
	t.Parallel()

	def := Definition{
		DefinitionSpec: DefinitionSpec{ID: "github_app"},
	}

	next := DefinitionProviderState{
		CredentialRef: NewCredentialSlotID("oauth"),
	}

	t.Run("empty state", func(t *testing.T) {
		t.Parallel()

		got, err := def.WithProviderState(IntegrationProviderState{}, next)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Providers == nil {
			t.Fatal("expected non-nil providers map")
		}

		if _, ok := got.Providers["github_app"]; !ok {
			t.Fatal("expected github_app key in providers")
		}
	})

	t.Run("updates existing", func(t *testing.T) {
		t.Parallel()

		initial := IntegrationProviderState{
			Providers: map[string]json.RawMessage{
				"github_app": json.RawMessage(`{"credentialRef":"old"}`),
			},
		}

		got, err := def.WithProviderState(initial, next)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var parsed DefinitionProviderState
		if err := json.Unmarshal(got.Providers["github_app"], &parsed); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		if parsed.CredentialRef.String() != "oauth" {
			t.Fatalf("got credential ref %q, want %q", parsed.CredentialRef.String(), "oauth")
		}
	})

	t.Run("preserves other providers", func(t *testing.T) {
		t.Parallel()

		initial := IntegrationProviderState{
			Providers: map[string]json.RawMessage{
				"other_provider": json.RawMessage(`{"credentialRef":"api_key"}`),
			},
		}

		got, err := def.WithProviderState(initial, next)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := got.Providers["other_provider"]; !ok {
			t.Fatal("expected other_provider key to be preserved")
		}

		if _, ok := got.Providers["github_app"]; !ok {
			t.Fatal("expected github_app key to be added")
		}
	})
}

func TestScopeVarsCELVars(t *testing.T) {
	t.Parallel()

	expectedKeys := []string{
		ScopeVariablePayload,
		ScopeVariableResource,
		ScopeVariableDefinition,
		ScopeVariableOperation,
		ScopeVariableConfig,
		ScopeVariableInstallationConfig,
		ScopeVariableOrgID,
		ScopeVariableInstallationID,
	}

	t.Run("all keys present", func(t *testing.T) {
		t.Parallel()

		vars := ScopeVars{}
		cel := vars.CELVars()

		for _, key := range expectedKeys {
			if _, ok := cel[key]; !ok {
				t.Fatalf("missing key %q in CEL vars", key)
			}
		}

		if len(cel) != len(expectedKeys) {
			t.Fatalf("got %d keys, want %d", len(cel), len(expectedKeys))
		}
	})

	t.Run("empty values", func(t *testing.T) {
		t.Parallel()

		vars := ScopeVars{}
		cel := vars.CELVars()

		if cel[ScopeVariableResource] != "" {
			t.Fatalf("expected empty resource, got %v", cel[ScopeVariableResource])
		}

		if cel[ScopeVariableOrgID] != "" {
			t.Fatalf("expected empty org_id, got %v", cel[ScopeVariableOrgID])
		}
	})

	t.Run("populated values", func(t *testing.T) {
		t.Parallel()

		vars := ScopeVars{
			Payload:            json.RawMessage(`{"action":"opened"}`),
			Resource:           "repo:123",
			Definition:         "github_app",
			Operation:          "pull_request",
			Config:             json.RawMessage(`{"enabled":true}`),
			InstallationConfig: json.RawMessage(`{"org":"acme"}`),
			OrgID:              "org_abc",
			InstallationID:     "inst_xyz",
		}

		cel := vars.CELVars()

		if cel[ScopeVariableResource] != "repo:123" {
			t.Fatalf("got resource %v, want %q", cel[ScopeVariableResource], "repo:123")
		}

		if cel[ScopeVariableDefinition] != "github_app" {
			t.Fatalf("got definition %v, want %q", cel[ScopeVariableDefinition], "github_app")
		}

		if cel[ScopeVariableOperation] != "pull_request" {
			t.Fatalf("got operation %v, want %q", cel[ScopeVariableOperation], "pull_request")
		}

		if cel[ScopeVariableOrgID] != "org_abc" {
			t.Fatalf("got org_id %v, want %q", cel[ScopeVariableOrgID], "org_abc")
		}

		if cel[ScopeVariableInstallationID] != "inst_xyz" {
			t.Fatalf("got installation_id %v, want %q", cel[ScopeVariableInstallationID], "inst_xyz")
		}

		// payload should decode to a map
		payloadMap, ok := cel[ScopeVariablePayload].(map[string]any)
		if !ok {
			t.Fatalf("expected payload to be map[string]any, got %T", cel[ScopeVariablePayload])
		}

		if payloadMap["action"] != "opened" {
			t.Fatalf("got payload action %v, want %q", payloadMap["action"], "opened")
		}
	})
}

func TestCredentialRefJSONRoundtrip(t *testing.T) {
	t.Parallel()

	original := NewCredentialSlotID("oauth_token")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded CredentialSlotID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.String() != original.String() {
		t.Fatalf("got %q, want %q", decoded.String(), original.String())
	}
}

func TestCredentialRefEmptyUnmarshal(t *testing.T) {
	t.Parallel()

	var ref CredentialSlotID
	if err := json.Unmarshal([]byte(`""`), &ref); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if ref.String() != "" {
		t.Fatalf("expected empty string, got %q", ref.String())
	}

	if ref != (CredentialSlotID{}) {
		t.Fatal("expected zero CredentialSlotID")
	}
}

func TestWebhookRefNameBasic(t *testing.T) {
	t.Parallel()

	ref := NewWebhookRef("push.events")
	if ref.Name() != "push.events" {
		t.Fatalf("got %q, want %q", ref.Name(), "push.events")
	}
}

func TestWebhookEventRefName(t *testing.T) {
	t.Parallel()

	ref := NewWebhookEventRef[struct{}]("pull_request.opened")
	if ref.Name() != "pull_request.opened" {
		t.Fatalf("got %q, want %q", ref.Name(), "pull_request.opened")
	}
}

func TestWebhookEventTopic(t *testing.T) {
	t.Parallel()

	got := WebhookEventTopic("def_001", "pull_request.opened")
	want := "integration.def_001.webhook.pull_request.opened"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestInstallationRefResolve(t *testing.T) {
	t.Parallel()

	type meta struct {
		AccountID string `json:"accountId"`
	}

	t.Run("returns data", func(t *testing.T) {
		t.Parallel()

		ref := NewInstallationRef(func(context.Context, InstallationRequest) (meta, bool, error) {
			return meta{AccountID: "acct_123"}, true, nil
		})

		got, ok, err := ref.Resolve(context.Background(), InstallationRequest{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected ok to be true")
		}

		if got.Attributes == nil {
			t.Fatal("expected non-nil attributes")
		}

		var parsed meta
		if err := json.Unmarshal(got.Attributes, &parsed); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		if parsed.AccountID != "acct_123" {
			t.Fatalf("got %q, want %q", parsed.AccountID, "acct_123")
		}
	})

	t.Run("returns false", func(t *testing.T) {
		t.Parallel()

		ref := NewInstallationRef(func(context.Context, InstallationRequest) (meta, bool, error) {
			return meta{}, false, nil
		})

		_, ok, err := ref.Resolve(context.Background(), InstallationRequest{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatal("expected ok to be false")
		}
	})

	t.Run("returns error", func(t *testing.T) {
		t.Parallel()

		errBoom := errors.New("boom")

		ref := NewInstallationRef(func(context.Context, InstallationRequest) (meta, bool, error) {
			return meta{}, false, errBoom
		})

		_, _, err := ref.Resolve(context.Background(), InstallationRequest{})
		if !errors.Is(err, errBoom) {
			t.Fatalf("got error %v, want %v", err, errBoom)
		}
	})
}

func TestInstallationRefRegistration(t *testing.T) {
	t.Parallel()

	ref := NewInstallationRef(func(context.Context, InstallationRequest) (struct{}, bool, error) {
		return struct{}{}, true, nil
	})

	reg := ref.Registration()
	if reg == nil {
		t.Fatal("expected non-nil registration")
	}

	if reg.Resolve == nil {
		t.Fatal("expected non-nil Resolve function")
	}
}
