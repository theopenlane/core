package types //nolint:revive

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/pkg/jsonx"
)

func TestOperationRegistrationDisabledFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		op    OperationRegistration
		input string
		want  bool
	}{
		{name: "stored input without the disable key", op: OperationRegistration{}, input: `{"limit":1}`, want: false},
		{name: "disabled for all ignores input", op: OperationRegistration{DisabledForAll: true}, input: `{}`, want: true},
		{name: "stored input disables", op: OperationRegistration{}, input: `{"disable":true,"limit":1}`, want: true},
		{name: "stored input explicitly enabled", op: OperationRegistration{}, input: `{"disable":false}`, want: false},
		{name: "nil input leaves the operation on", op: OperationRegistration{}, want: false},
		{name: "undecodable input leaves the operation on", op: OperationRegistration{}, input: `[]`, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var input json.RawMessage
			if tc.input != "" {
				input = json.RawMessage(tc.input)
			}

			if got := tc.op.DisabledFor(input); got != tc.want {
				t.Fatalf("DisabledFor() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOperationSettingsFrom(t *testing.T) {
	t.Parallel()

	settings, err := OperationSettingsFrom(json.RawMessage(`{"disable":true,"filterExpr":"payload.ok","limit":1}`))
	if err != nil {
		t.Fatalf("OperationSettingsFrom() error = %v", err)
	}

	if !settings.Disable || settings.FilterExpr != "payload.ok" {
		t.Fatalf("expected the uniform keys decoded beside config keys, got %+v", settings)
	}

	empty, err := OperationSettingsFrom(nil)
	if err != nil || empty != (OperationSettings{}) {
		t.Fatalf("expected empty input to decode to zero settings, got %+v, %v", empty, err)
	}

	if _, err := OperationSettingsFrom(json.RawMessage(`[]`)); err == nil {
		t.Fatal("expected a non-object document to fail decoding")
	}

	keys := lo.Map(jsonx.PropertyDescriptors[OperationSettings](), func(property jsonx.PropertyDescriptor, _ int) string {
		return property.Name
	})

	if !slices.Equal(keys, []string{"disable", "filterExpr"}) {
		t.Fatalf("settings schema keys = %v", keys)
	}

	if got := len(OperationSettingsSchema()); got == 0 {
		t.Fatal("expected the settings schema reflected")
	}
}

// TestDefinitionResolveOperation verifies exact names resolve directly and retired names resolve through their replacement
func TestDefinitionResolveOperation(t *testing.T) {
	t.Parallel()

	def := Definition{
		DefinitionSpec: DefinitionSpec{ID: "test-def"},
		Operations: []OperationRegistration{
			{Name: "sync.users"},
			{Name: "sync.groups", Replaces: []string{"sync.teams"}},
		},
	}

	t.Run("exact", func(t *testing.T) {
		t.Parallel()

		reg, replaced, ok := def.ResolveOperation("sync.users")
		if !ok || replaced || reg.Name != "sync.users" {
			t.Fatalf("expected sync.users to resolve directly, got %+v replaced=%v ok=%v", reg, replaced, ok)
		}
	})

	t.Run("retired", func(t *testing.T) {
		t.Parallel()

		reg, replaced, ok := def.ResolveOperation("sync.teams")
		if !ok || !replaced || reg.Name != "sync.groups" {
			t.Fatalf("expected sync.groups to replace sync.teams, got %+v replaced=%v ok=%v", reg, replaced, ok)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		if _, _, ok := def.ResolveOperation("sync.unknown"); ok {
			t.Fatal("expected an undeclared name not to resolve")
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

		if got.CredentialRef != "" {
			t.Fatalf("expected zero DefinitionProviderState, got credential ref %q", got.CredentialRef)
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

		if got.CredentialRef != "" {
			t.Fatalf("expected zero DefinitionProviderState, got credential ref %q", got.CredentialRef)
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

		if got.CredentialRef != "oauth" {
			t.Fatalf("got credential ref %q, want %q", got.CredentialRef, "oauth")
		}
	})
}

func TestDefinitionWithProviderState(t *testing.T) {
	t.Parallel()

	def := Definition{
		DefinitionSpec: DefinitionSpec{ID: "github_app"},
	}

	next := DefinitionProviderState{
		CredentialRef: "oauthCredential",
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

		if parsed.CredentialRef != "oauthCredential" {
			t.Fatalf("got credential ref %q, want %q", parsed.CredentialRef, "oauthCredential")
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

		payloadMap, ok := cel[ScopeVariablePayload].(map[string]any)
		if !ok {
			t.Fatalf("expected payload to be map[string]any, got %T", cel[ScopeVariablePayload])
		}

		if payloadMap["action"] != "opened" {
			t.Fatalf("got payload action %v, want %q", payloadMap["action"], "opened")
		}
	})
}

func TestWebhookRefNameBasic(t *testing.T) {
	t.Parallel()

	ref := NewWebhookRef("push.events")
	if ref.Name() != "push.events" {
		t.Fatalf("got %q, want %q", ref.Name(), "push.events")
	}
}

func TestWebhookEventTopic(t *testing.T) {
	t.Parallel()

	ref := NewDefinitionRef("def_001")
	got := ref.WebhookEventTopic("pull_request.opened")
	want := "integration.webhook.def_001.pull_request.opened"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
