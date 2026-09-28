package types //nolint:revive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/theopenlane/core/v2/pkg/jsonx"
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

func TestClientRefIdentity(t *testing.T) {
	t.Parallel()

	first := NewClientRef[string]("first")
	second := NewClientRef[string]("second")
	same := NewClientRef[string]("first")

	if !first.ID().Valid() || !second.ID().Valid() {
		t.Fatal("named client refs should be valid")
	}

	if first.ID() == second.ID() {
		t.Fatal("client refs with different names should be distinct")
	}

	if first.ID() != same.ID() {
		t.Fatal("client refs with the same name should be equal")
	}

	if got := ClientRefOf[*bytes.Buffer]().ID().String(); got != "bytes.Buffer" {
		t.Fatalf("ClientRefOf[*bytes.Buffer]() = %q, want %q", got, "bytes.Buffer")
	}

	if got := ClientRefOf[string]().ID().String(); got != "string" {
		t.Fatalf("ClientRefOf[string]() = %q, want %q", got, "string")
	}

	if (ClientID{}).Valid() {
		t.Fatal("zero ClientID should not be valid")
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

	ref := NewClientRef[string]("client")
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

	ref := NewClientRef[string]("client")
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

func TestCredentialRefOfDerivesSlotFromSchema(t *testing.T) {
	t.Parallel()

	ref := CredentialRefOf[refTestCredential]()

	if ref.ID() != NewCredentialSlotID("refTestCredential") {
		t.Fatalf("expected the slot named after the type, got %q", ref.String())
	}

	if jsonx.SchemaID(ref.Schema()) != "refTestCredential" {
		t.Fatalf("expected the reflected schema to be retained, got %s", ref.Schema())
	}
}

func TestCredentialRefSchemaReturnsACopy(t *testing.T) {
	t.Parallel()

	ref := NewCredentialRef[refTestCredential]("refTestCredential")

	first := ref.Schema()
	if len(first) == 0 {
		t.Fatal("expected NewCredentialRef to reflect the schema")
	}

	first[0] = 'x'

	if ref.Schema()[0] != '{' {
		t.Fatal("expected Schema() to return a copy")
	}
}

func TestCredentialRefRegistration(t *testing.T) {
	t.Parallel()

	retired := NewCredentialRef[refTestRetiredCredential]("refTestRetiredCredential")
	plain := NewCredentialRef[refTestCredential]("refTestCredential")
	base := CredentialRegistration{Name: "Token", Description: "desc", Schema: plain.Schema(), Recommended: true}

	t.Run("plain slot projects only the identity", func(t *testing.T) {
		t.Parallel()

		reg := plain.Registration(base)

		if reg.Ref != plain.ID() {
			t.Fatalf("Ref = %q, want %q", reg.Ref, plain.ID())
		}

		if reg.Name != "Token" || reg.Description != "desc" || !reg.Recommended {
			t.Fatalf("expected base fields preserved, got %+v", reg)
		}

		if string(reg.Schema) != string(plain.Schema()) {
			t.Fatalf("expected the authored schema preserved, got %s", reg.Schema)
		}

		if len(reg.Replaces) != 0 || reg.Convert != nil || reg.Backfill != nil {
			t.Fatalf("expected no lifecycle on a plain slot, got %+v", reg)
		}
	})

	t.Run("lifecycle slot projects convert and backfill", func(t *testing.T) {
		t.Parallel()

		ref := plain.Replacing(retired, func(r refTestRetiredCredential) refTestCredential {
			return refTestCredential{Token: r.AccessToken}
		}).Backfilled(func(_ context.Context, _ InstallationRequest, c *refTestCredential) error {
			c.Token = "derived"

			return nil
		})

		reg := ref.Registration(base)

		if !slices.Equal(reg.Replaces, []CredentialSlotID{retired.ID()}) {
			t.Fatalf("Replaces = %v", reg.Replaces)
		}

		if reg.Convert == nil || reg.Backfill == nil {
			t.Fatal("expected Convert and Backfill projected")
		}

		converted, err := reg.Convert(retired.ID(), json.RawMessage(`{"accessToken":"t"}`))
		if err != nil || string(converted) != `{"token":"t"}` {
			t.Fatalf("Convert() = %s, %v", converted, err)
		}

		filled, err := reg.Backfill(context.Background(), InstallationRequest{}, json.RawMessage(`{"token":""}`))
		if err != nil || string(filled) != `{"token":"derived"}` {
			t.Fatalf("Backfill() = %s, %v", filled, err)
		}
	})

	t.Run("auth-managed slot keeps an empty schema", func(t *testing.T) {
		t.Parallel()

		reg := plain.Registration(CredentialRegistration{Name: "OAuth"})

		if len(reg.Schema) != 0 {
			t.Fatalf("expected Registration not to fill Schema, got %s", reg.Schema)
		}
	})
}

func TestClientRefUsingDoesNotAliasTheReceiver(t *testing.T) {
	t.Parallel()

	first := NewCredentialRef[refTestCredential]("first")
	second := NewCredentialRef[refTestRetiredCredential]("second")
	third := NewCredentialRef[struct{}]("third")

	base := NewClientRef[string]("client").Using(first)
	withSecond := base.Using(second)
	withThird := base.Using(third)

	build := func(context.Context, ClientBuildRequest) (string, error) { return "", nil }

	if got := base.Registration(build, ClientRegistration{}).CredentialRefs; !slices.Equal(got, []CredentialSlotID{first.ID()}) {
		t.Fatalf("expected the source ref to keep one slot, got %v", got)
	}

	if got := withSecond.Registration(build, ClientRegistration{}).CredentialRefs; !slices.Equal(got, []CredentialSlotID{first.ID(), second.ID()}) {
		t.Fatalf("expected [first second], got %v", got)
	}

	if got := withThird.Registration(build, ClientRegistration{}).CredentialRefs; !slices.Equal(got, []CredentialSlotID{first.ID(), third.ID()}) {
		t.Fatalf("expected [first third], got %v", got)
	}

	if withSecond.ID() != base.ID() {
		t.Fatal("expected Using to keep the client identity")
	}
}

func TestClientRefRegistration(t *testing.T) {
	t.Parallel()

	cred := NewCredentialRef[refTestCredential]("refTestCredential")
	ref := NewClientRef[string]("client").Using(cred)

	build := func(_ context.Context, req ClientBuildRequest) (string, error) {
		return "built:" + string(req.Config), nil
	}

	reg := ref.Registration(build, ClientRegistration{Description: "client", ConfigSchema: json.RawMessage(`{}`)})

	if reg.Ref != ref.ID() {
		t.Fatalf("Ref = %v, want %v", reg.Ref, ref.ID())
	}

	if reg.Description != "client" || string(reg.ConfigSchema) != `{}` {
		t.Fatalf("expected base fields preserved, got %+v", reg)
	}

	if !slices.Equal(reg.CredentialRefs, []CredentialSlotID{cred.ID()}) {
		t.Fatalf("CredentialRefs = %v", reg.CredentialRefs)
	}

	got, err := reg.Build(context.Background(), ClientBuildRequest{Config: json.RawMessage(`"cfg"`)})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if got != `built:"cfg"` {
		t.Fatalf("Build() = %v", got)
	}

	reg.CredentialRefs[0] = NewCredentialSlotID("mutated")

	if again := ref.Registration(build, ClientRegistration{}); again.CredentialRefs[0] != cred.ID() {
		t.Fatal("expected Registration to copy the slots")
	}
}

func TestClientRefRegistrationPropagatesBuildErrors(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("boom")

	reg := NewClientRef[string]("client").Registration(func(context.Context, ClientBuildRequest) (string, error) {
		return "", errBoom
	}, ClientRegistration{})

	if _, err := reg.Build(context.Background(), ClientBuildRequest{}); !errors.Is(err, errBoom) {
		t.Fatalf("Build() error = %v, want %v", err, errBoom)
	}
}

func TestOperationRefOfDerivesNameFromSchema(t *testing.T) {
	t.Parallel()

	ref := OperationRefOf[refTestInput]()

	if ref.Name() != "refTestInput" {
		t.Fatalf("Name() = %q", ref.Name())
	}

	if jsonx.SchemaID(ref.Schema()) != "refTestInput" {
		t.Fatalf("expected the reflected schema retained, got %s", ref.Schema())
	}
}

func TestOperationRefRegistration(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")
	client := NewClientRef[string]("client")
	base := OperationRegistration{Description: "sync", Policy: ExecutionPolicy{Reconcile: true}, Replaces: []string{"old"}}

	t.Run("without a client", func(t *testing.T) {
		t.Parallel()

		ref := NewOperationRef[refTestInput]("refTestInput")
		reg := ref.Registration(definition, base)

		if reg.Name != "refTestInput" {
			t.Fatalf("Name = %q", reg.Name)
		}

		if string(reg.Topic) != "integration.run.def_001.refTestInput" {
			t.Fatalf("Topic = %q", reg.Topic)
		}

		if string(reg.ConfigSchema) != string(ref.Schema()) {
			t.Fatalf("ConfigSchema = %s", reg.ConfigSchema)
		}

		if reg.ClientRef.Valid() {
			t.Fatal("expected no client ref")
		}

		if reg.Description != "sync" || !reg.Policy.Reconcile || !slices.Equal(reg.Replaces, []string{"old"}) {
			t.Fatalf("expected base fields preserved, got %+v", reg)
		}
	})

	t.Run("with a client", func(t *testing.T) {
		t.Parallel()

		reg := NewOperationRef[refTestInput]("refTestInput").Using(client).Registration(definition, base)

		if reg.ClientRef != client.ID() {
			t.Fatalf("ClientRef = %v, want %v", reg.ClientRef, client.ID())
		}
	})

	t.Run("base client is kept when the ref declares none", func(t *testing.T) {
		t.Parallel()

		reg := NewOperationRef[refTestInput]("refTestInput").Registration(definition, OperationRegistration{ClientRef: client.ID()})

		if reg.ClientRef != client.ID() {
			t.Fatalf("ClientRef = %v, want %v", reg.ClientRef, client.ID())
		}
	})
}

func TestOperationRefUsingDoesNotAliasTheReceiver(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")
	base := NewOperationRef[refTestInput]("refTestInput")
	bound := base.Using(NewClientRef[string]("client"))

	if base.Registration(definition, OperationRegistration{}).ClientRef.Valid() {
		t.Fatal("expected the source ref to stay unbound")
	}

	if !bound.Registration(definition, OperationRegistration{}).ClientRef.Valid() {
		t.Fatal("expected the bound ref to carry the client")
	}
}

func TestOperationRefConfigFrom(t *testing.T) {
	t.Parallel()

	type userInput struct {
		Sync refTestInput `json:"sync"`
	}

	resolve := NewOperationRef[refTestInput]("refTestInput").ConfigFrom(func(u userInput) refTestInput { return u.Sync })

	if got := resolve(json.RawMessage(`{"sync":{"region":"eu"}}`)); string(got) != `{"region":"eu"}` {
		t.Fatalf("ConfigFrom() = %s", got)
	}

	if got := resolve(json.RawMessage(`not json`)); got != nil {
		t.Fatalf("expected nil on undecodable input, got %s", got)
	}
}

func TestWebhookRefRegistration(t *testing.T) {
	t.Parallel()

	reg := NewWebhookRef("installation.events").Registration(WebhookRegistration{StaticRoute: "/hooks", Replaces: []string{"old"}})

	if reg.Name != "installation.events" {
		t.Fatalf("Name = %q", reg.Name)
	}

	if reg.StaticRoute != "/hooks" || !slices.Equal(reg.Replaces, []string{"old"}) {
		t.Fatalf("expected base fields preserved, got %+v", reg)
	}
}

func TestWebhookEventRefOfDerivesNameFromSchema(t *testing.T) {
	t.Parallel()

	if ref := WebhookEventRefOf[refTestInput](); ref.Name() != "refTestInput" {
		t.Fatalf("Name() = %q", ref.Name())
	}
}

func TestWebhookEventRefRegistration(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")
	reg := NewWebhookEventRef[refTestInput]("created").Registration(definition, WebhookEventRegistration{Ingest: []IngestContract{{Schema: "User"}}})

	if reg.Name != "created" {
		t.Fatalf("Name = %q", reg.Name)
	}

	if string(reg.Topic) != "integration.webhook.def_001.created" {
		t.Fatalf("Topic = %q", reg.Topic)
	}

	if len(reg.Ingest) != 1 || reg.Ingest[0].Schema != "User" {
		t.Fatalf("expected base fields preserved, got %+v", reg)
	}
}

func TestConnectionRefRegistration(t *testing.T) {
	t.Parallel()

	cred := NewCredentialRef[refTestCredential]("refTestCredential")
	other := NewCredentialRef[refTestRetiredCredential]("other")
	client := NewClientRef[string]("client")
	ref := NewConnectionRef(cred).Enables(client)

	t.Run("derives the credential and clients", func(t *testing.T) {
		t.Parallel()

		authored := []CredentialSlotID{cred.ID(), other.ID()}
		reg := ref.Registration(ConnectionRegistration{Name: "Token", CredentialRefs: authored})

		if reg.CredentialRef != cred.ID() {
			t.Fatalf("CredentialRef = %q, want %q", reg.CredentialRef, cred.ID())
		}

		if !slices.Equal(reg.ClientRefs, []ClientID{client.ID()}) {
			t.Fatalf("ClientRefs = %v", reg.ClientRefs)
		}

		if !slices.Equal(reg.CredentialRefs, authored) {
			t.Fatalf("expected CredentialRefs untouched, got %v", reg.CredentialRefs)
		}

		if reg.Name != "Token" {
			t.Fatalf("expected base fields preserved, got %+v", reg)
		}
	})

	t.Run("leaves unset credential refs unset", func(t *testing.T) {
		t.Parallel()

		if reg := ref.Registration(ConnectionRegistration{}); reg.CredentialRefs != nil {
			t.Fatalf("expected nil CredentialRefs, got %v", reg.CredentialRefs)
		}
	})
}

func TestConnectionRefEnablesDoesNotAliasTheReceiver(t *testing.T) {
	t.Parallel()

	cred := NewCredentialRef[refTestCredential]("refTestCredential")
	first := NewClientRef[string]("first")
	second := NewClientRef[int]("second")
	third := NewClientRef[bool]("third")

	base := NewConnectionRef(cred).Enables(first)
	withSecond := base.Enables(second)
	withThird := base.Enables(third)

	if got := base.Registration(ConnectionRegistration{}).ClientRefs; !slices.Equal(got, []ClientID{first.ID()}) {
		t.Fatalf("expected the source ref to keep one client, got %v", got)
	}

	if got := withSecond.Registration(ConnectionRegistration{}).ClientRefs; !slices.Equal(got, []ClientID{first.ID(), second.ID()}) {
		t.Fatalf("expected [first second], got %v", got)
	}

	if got := withThird.Registration(ConnectionRegistration{}).ClientRefs; !slices.Equal(got, []ClientID{first.ID(), third.ID()}) {
		t.Fatalf("expected [first third], got %v", got)
	}
}
