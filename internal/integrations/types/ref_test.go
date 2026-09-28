package types //nolint:revive

import (
	"context"
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

// refTestInput is the user input type the user input ref tests reflect
type refTestInput struct {
	Region string `json:"region" jsonschema:"required"`
}

// refTestRetiredInput is the shape an earlier definition version stored the region under
type refTestRetiredInput struct {
	Zone string `json:"zone"`
}

func TestCredentialRefReplacingConvert(t *testing.T) {
	t.Parallel()

	retired := NewCredentialRef[refTestRetiredCredential]("refTestRetiredCredential")

	tests := []struct {
		name    string
		ref     CredentialRef[refTestCredential]
		from    CredentialSlotID
		payload string
		want    string
		wantErr error
	}{
		{
			name:    "nil convert decodes the retired payload directly",
			ref:     NewCredentialRef[refTestCredential]("refTestCredential").Replacing(retired, nil),
			from:    retired.ID(),
			payload: `{"accessToken":"t"}`,
			want:    `{"token":""}`,
		},
		{
			name:    "payload outside the retired layout is rejected",
			ref:     NewCredentialRef[refTestCredential]("refTestCredential").Replacing(retired, nil),
			from:    retired.ID(),
			payload: `{"token":"t"}`,
			wantErr: ErrLayoutMismatch,
		},
		{
			name: "convert maps fields",
			ref: NewCredentialRef[refTestCredential]("refTestCredential").Replacing(retired, func(r refTestRetiredCredential) refTestCredential {
				return refTestCredential{Token: r.AccessToken}
			}),
			from:    retired.ID(),
			payload: `{"accessToken":"t"}`,
			want:    `{"token":"t"}`,
		},
		{
			name:    "undeclared slot is not replaced",
			ref:     NewCredentialRef[refTestCredential]("refTestCredential").Replacing(retired, nil),
			from:    NewCredentialSlotID("unknown"),
			payload: `{"token":"t"}`,
			wantErr: ErrNotReplaced,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.ref.Convert(tc.from, json.RawMessage(tc.payload))

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

func TestCredentialRefReplacingDoesNotAliasTheSource(t *testing.T) {
	t.Parallel()

	base := NewCredentialRef[refTestCredential]("refTestCredential").Replacing(NewCredentialRef[refTestRetiredCredential]("refTestRetiredCredential"), nil)
	extended := base.Replacing(NewCredentialRef[struct{}]("other"), nil)

	if got := len(base.Replaces()); got != 1 {
		t.Fatalf("expected the source ref to keep one replacement, got %d", got)
	}

	if got := extended.Replaces(); len(got) != 2 || got[0] != NewCredentialSlotID("other") || got[1] != NewCredentialSlotID("refTestRetiredCredential") {
		t.Fatalf("expected sorted replacements on the extended ref, got %v", got)
	}
}

func TestCredentialRefBackfillWithoutDeclarationReturnsPayloadUnchanged(t *testing.T) {
	t.Parallel()

	plain := NewCredentialRef[refTestCredential]("refTestCredential")

	unchanged, err := plain.Backfill(context.Background(), InstallationRequest{}, json.RawMessage(`{"token":""}`))
	if err != nil || string(unchanged) != `{"token":""}` {
		t.Fatalf("expected the payload unchanged, got %s, %v", unchanged, err)
	}
}

func TestCredentialRefBackfilledReceivesTheInstallationRequest(t *testing.T) {
	t.Parallel()

	var seen InstallationRequest

	ref := NewCredentialRef[refTestCredential]("refTestCredential").Backfilled(func(_ context.Context, req InstallationRequest, c *refTestCredential) error {
		seen = req
		c.Token = "derived"

		return nil
	})

	got, err := ref.Backfill(context.Background(), InstallationRequest{Input: json.RawMessage(`{"marker":true}`)}, json.RawMessage(`{"token":""}`))
	if err != nil {
		t.Fatalf("Backfill() error = %v", err)
	}

	if string(got) != `{"token":"derived"}` {
		t.Fatalf("Backfill() = %s", got)
	}

	if string(seen.Input) != `{"marker":true}` {
		t.Fatalf("expected the backfill to receive the explicit request, got %s", seen.Input)
	}
}

func TestNewCredentialRefKeepsNameAsSlotID(t *testing.T) {
	t.Parallel()

	ref := NewCredentialRef[refTestCredential]("refTestCredential")

	if ref.ID() != NewCredentialSlotID("refTestCredential") {
		t.Fatalf("expected the slot to carry the given name, got %q", ref.String())
	}

	encoded, err := json.Marshal(ref.ID())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if string(encoded) != `"refTestCredential"` {
		t.Fatalf("expected the slot name on the wire, got %s", encoded)
	}
}

func TestUserInputRefName(t *testing.T) {
	t.Parallel()

	ref := NewUserInputRef[refTestInput]("installation.input")

	if ref.Name() != "installation.input" {
		t.Fatalf("Name() = %q", ref.Name())
	}
}

func TestUserInputRefReplacingConvert(t *testing.T) {
	t.Parallel()

	retired := NewUserInputRef[refTestRetiredInput]("refTestRetiredInput")

	tests := []struct {
		name    string
		ref     UserInputRef[refTestInput]
		payload string
		want    string
		wantErr error
	}{
		{
			name:    "nil convert decodes the retired payload directly",
			ref:     NewUserInputRef[refTestInput]("refTestInput").Replacing(retired, nil),
			payload: `{"zone":"eu"}`,
			want:    `{"region":""}`,
		},
		{
			name: "convert maps fields",
			ref: NewUserInputRef[refTestInput]("refTestInput").Replacing(retired, func(r refTestRetiredInput) refTestInput {
				return refTestInput{Region: r.Zone}
			}),
			payload: `{"zone":"eu"}`,
			want:    `{"region":"eu"}`,
		},
		{
			name:    "payload matching no retired layout is rejected",
			ref:     NewUserInputRef[refTestInput]("refTestInput").Replacing(retired, nil),
			payload: `{"region":"eu"}`,
			wantErr: ErrLayoutMismatch,
		},
		{
			name:    "layout without replacements rejects every payload",
			ref:     NewUserInputRef[refTestInput]("refTestInput"),
			payload: `{"zone":"eu"}`,
			wantErr: ErrLayoutMismatch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.ref.Convert(json.RawMessage(tc.payload))

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

func TestUserInputRefReplacingDoesNotAliasTheSource(t *testing.T) {
	t.Parallel()

	base := NewUserInputRef[refTestInput]("refTestInput").Replacing(NewUserInputRef[refTestRetiredInput]("refTestRetiredInput"), nil)
	extended := base.Replacing(NewUserInputRef[struct{}]("other"), nil)

	if got := len(base.Replaces()); got != 1 {
		t.Fatalf("expected the source ref to keep one replacement, got %d", got)
	}

	if got := extended.Replaces(); len(got) != 2 || got[0] != "other" || got[1] != "refTestRetiredInput" {
		t.Fatalf("expected sorted replacements on the extended ref, got %v", got)
	}
}

func TestUserInputRefBackfilledReceivesTheInstallationRequest(t *testing.T) {
	t.Parallel()

	var seen InstallationRequest

	ref := NewUserInputRef[refTestInput]("refTestInput").Backfilled(func(_ context.Context, req InstallationRequest, in *refTestInput) error {
		seen = req
		in.Region = "derived"

		return nil
	})

	got, err := ref.Backfill(context.Background(), InstallationRequest{Input: json.RawMessage(`{"marker":true}`)}, json.RawMessage(`{"region":""}`))
	if err != nil {
		t.Fatalf("Backfill() error = %v", err)
	}

	if string(got) != `{"region":"derived"}` {
		t.Fatalf("Backfill() = %s", got)
	}

	if string(seen.Input) != `{"marker":true}` {
		t.Fatalf("expected the backfill to receive the explicit request, got %s", seen.Input)
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

	ref := NewOperationRef[refTestCredential]("refTestCredential")
	if ref.Name() != "refTestCredential" {
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

	ref := NewOperationRef[cfg]("cfg")
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

	ref := NewOperationRef[struct{}]("empty")
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
	if ref.Name() != "created" {
		t.Fatalf("WebhookEventRef.Name() = %q", ref.Name())
	}

	got, err := ref.UnmarshalPayload(json.RawMessage(`{"action":"opened"}`))
	if err != nil {
		t.Fatalf("UnmarshalPayload() error = %v", err)
	}

	if got.Action != "opened" {
		t.Fatalf("UnmarshalPayload() action = %q, want %q", got.Action, "opened")
	}
}
