package types //nolint:revive

import (
	"encoding/json"
	"errors"
	"testing"
)

// refTestCredential is the credential type the credential ref tests reflect
type refTestCredential struct {
	Token string `json:"token" jsonschema:"required"`
}

// refTestRetiredCredential is the shape an earlier definition version stored the token under
type refTestRetiredCredential struct {
	AccessToken string `json:"accessToken"`
}

func TestReplacingConvert(t *testing.T) {
	t.Parallel()

	retired := NewCredentialRef[refTestRetiredCredential]()

	tests := []struct {
		name    string
		slot    CredentialRef[refTestCredential]
		from    CredentialSlotID
		payload string
		want    string
		wantErr error
	}{
		{
			name:    "nil convert decodes matching fields",
			slot:    Replacing(NewCredentialRef[refTestCredential](), retired, nil),
			from:    retired.ID(),
			payload: `{"token":"t","accessToken":"ignored"}`,
			want:    `{"token":"t"}`,
		},
		{
			name: "convert maps fields",
			slot: Replacing(NewCredentialRef[refTestCredential](), retired, func(r refTestRetiredCredential) refTestCredential {
				return refTestCredential{Token: r.AccessToken}
			}),
			from:    retired.ID(),
			payload: `{"accessToken":"t"}`,
			want:    `{"token":"t"}`,
		},
		{
			name:    "unknown from is not replaced",
			slot:    Replacing(NewCredentialRef[refTestCredential](), retired, nil),
			from:    NewCredentialSlotID("unknown"),
			payload: `{"token":"t"}`,
			wantErr: ErrCredentialNotReplaced,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.slot.Convert(tc.from, json.RawMessage(tc.payload))

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			if string(got) != tc.want {
				t.Fatalf("Convert() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestNewCredentialRefDerivesIdentityAndSchemaFromType(t *testing.T) {
	t.Parallel()

	ref := NewCredentialRef[refTestCredential]()

	if ref.ID() != NewCredentialSlotID("refTestCredential") {
		t.Fatalf("expected the slot to be named after the type, got %q", ref.String())
	}

	var schema struct {
		Ref  string `json:"$ref"`
		Defs map[string]struct {
			Required []string `json:"required"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(ref.Schema(), &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	if schema.Ref != "#/$defs/refTestCredential" || len(schema.Defs["refTestCredential"].Required) != 1 {
		t.Fatalf("expected the reflected schema of the type, got %s", ref.Schema())
	}

	var slot CredentialSlot = ref
	if slot.ID() != ref.ID() {
		t.Fatal("expected the typed ref to satisfy CredentialSlot")
	}
}

func TestCredentialRefSchemaIsCloned(t *testing.T) {
	t.Parallel()

	ref := NewCredentialRef[refTestCredential]()

	first := ref.Schema()
	first[0] = 'x'

	if ref.Schema()[0] != '{' {
		t.Fatal("expected Schema to return a copy the caller cannot mutate")
	}
}

func TestCredentialRefMarshalsAsSlotName(t *testing.T) {
	t.Parallel()

	encoded, err := json.Marshal(NewCredentialRef[refTestCredential]())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if string(encoded) != `"refTestCredential"` {
		t.Fatalf("expected the slot name on the wire, got %s", encoded)
	}
}

func TestDefinitionRefID(t *testing.T) {
	t.Parallel()

	ref := NewDefinitionRef("def_01TEST0000000000000000001")
	if ref.ID() != "def_01TEST0000000000000000001" {
		t.Fatalf("DefinitionRef.ID() = %q", ref.ID())
	}
}

func TestClientRefIDsAreDistinct(t *testing.T) {
	t.Parallel()

	first := NewClientRef[string]()
	second := NewClientRef[string]()

	if !first.ID().Valid() {
		t.Fatal("first client ref is not valid")
	}

	if !second.ID().Valid() {
		t.Fatal("second client ref is not valid")
	}

	if first.ID() == second.ID() {
		t.Fatal("client refs should be distinct")
	}

	if first.ID().String() == second.ID().String() {
		t.Fatal("client ref strings should be distinct")
	}
}

func TestOperationRefName(t *testing.T) {
	t.Parallel()

	ref := NewOperationRef[struct{}]("health.default")
	if ref.Name() != "health.default" {
		t.Fatalf("OperationRef.Name() = %q", ref.Name())
	}
}

func TestOperationTopic(t *testing.T) {
	t.Parallel()

	ref := NewDefinitionRef("def_001")
	got := ref.OperationTopic("health.default")
	want := "integration.run.def_001.health.default"
	if string(got) != want {
		t.Fatalf("OperationTopic() = %q, want %q", got, want)
	}
}

func TestWebhookRefName(t *testing.T) {
	t.Parallel()

	ref := NewWebhookRef("installation.events")
	if ref.Name() != "installation.events" {
		t.Fatalf("WebhookRef.Name() = %q", ref.Name())
	}
}

func TestClientRefCastSuccess(t *testing.T) {
	t.Parallel()

	ref := NewClientRef[string]()
	got, err := ref.Cast("hello")
	if err != nil {
		t.Fatalf("Cast() error = %v", err)
	}

	if got != "hello" {
		t.Fatalf("Cast() = %q, want %q", got, "hello")
	}
}

func TestClientRefCastFailure(t *testing.T) {
	t.Parallel()

	ref := NewClientRef[string]()
	_, err := ref.Cast(42)
	if err == nil {
		t.Fatal("Cast() expected error, got nil")
	}
}

func TestOperationRefUnmarshalConfig(t *testing.T) {
	t.Parallel()

	type cfg struct {
		Name string `json:"name"`
	}

	ref := NewOperationRef[cfg]("test.op")
	got, err := ref.UnmarshalConfig(json.RawMessage(`{"name":"foo"}`))
	if err != nil {
		t.Fatalf("UnmarshalConfig() error = %v", err)
	}

	if got.Name != "foo" {
		t.Fatalf("UnmarshalConfig() name = %q, want %q", got.Name, "foo")
	}
}

func TestOperationRefUnmarshalConfigNil(t *testing.T) {
	t.Parallel()

	ref := NewOperationRef[struct{}]("test.noop")
	_, err := ref.UnmarshalConfig(nil)
	if err != nil {
		t.Fatalf("UnmarshalConfig(nil) error = %v", err)
	}
}

func TestWebhookEventRefUnmarshalPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Action string `json:"action"`
	}

	ref := NewWebhookEventRef[payload]("created")
	got, err := ref.UnmarshalPayload(json.RawMessage(`{"action":"opened"}`))
	if err != nil {
		t.Fatalf("UnmarshalPayload() error = %v", err)
	}

	if got.Action != "opened" {
		t.Fatalf("UnmarshalPayload() action = %q, want %q", got.Action, "opened")
	}
}
