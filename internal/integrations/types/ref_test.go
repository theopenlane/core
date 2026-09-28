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

// refTestSwitchConfig is an operation config type carrying the embedded disable switch
type refTestSwitchConfig struct {
	Switch
	Limit int `json:"limit"`
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

func TestOperationRefHandles(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")
	client := NewClientRef[string]("client")

	ref := NewOperationRef[refTestSwitchConfig]("sync").Handles(client, func(_ context.Context, _ OperationRequest, c string, cfg refTestSwitchConfig) (json.RawMessage, error) {
		return json.Marshal(struct {
			Client string `json:"client"`
			Limit  int    `json:"limit"`
		}{Client: c, Limit: cfg.Limit})
	})

	reg := ref.Registration(definition, OperationRegistration{})

	if reg.Handle == nil || reg.IngestHandle != nil {
		t.Fatalf("expected only Handle projected, got %+v", reg)
	}

	if reg.ClientRef != client.ID() {
		t.Fatalf("ClientRef = %v, want %v", reg.ClientRef, client.ID())
	}

	tests := []struct {
		name    string
		request OperationRequest
		want    string
		wantErr error
	}{
		{name: "passes the typed client and config", request: OperationRequest{Client: "c", Config: json.RawMessage(`{"limit":3}`)}, want: `{"client":"c","limit":3}`},
		{name: "absent config is the zero config", request: OperationRequest{Client: "c"}, want: `{"client":"c","limit":0}`},
		{name: "client cast failure", request: OperationRequest{Client: 1, Config: json.RawMessage(`{"limit":3}`)}, wantErr: ErrClientCastFailed},
		{name: "config decode failure", request: OperationRequest{Client: "c", Config: json.RawMessage(`{"limit":"x"}`)}, wantErr: ErrOperationConfigInvalid},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := reg.Handle(context.Background(), tc.request)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Handle() error = %v", err)
			}

			if string(got) != tc.want {
				t.Fatalf("Handle() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestOperationRefIngests(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")
	client := NewClientRef[string]("client")

	ref := NewOperationRef[refTestSwitchConfig]("sync").Ingests(client, func(_ context.Context, _ OperationRequest, c string, cfg refTestSwitchConfig) ([]IngestPayloadSet, error) {
		return []IngestPayloadSet{{Schema: c, Envelopes: make([]MappingEnvelope, cfg.Limit)}}, nil
	})

	reg := ref.Registration(definition, OperationRegistration{})

	if reg.IngestHandle == nil || reg.Handle != nil {
		t.Fatalf("expected only IngestHandle projected, got %+v", reg)
	}

	if reg.ClientRef != client.ID() {
		t.Fatalf("ClientRef = %v, want %v", reg.ClientRef, client.ID())
	}

	t.Run("passes the typed client and config", func(t *testing.T) {
		t.Parallel()

		got, err := reg.IngestHandle(context.Background(), OperationRequest{Client: "User", Config: json.RawMessage(`{"limit":2}`)})
		if err != nil {
			t.Fatalf("IngestHandle() error = %v", err)
		}

		if len(got) != 1 || got[0].Schema != "User" || len(got[0].Envelopes) != 2 {
			t.Fatalf("IngestHandle() = %+v", got)
		}
	})

	t.Run("client cast failure", func(t *testing.T) {
		t.Parallel()

		if _, err := reg.IngestHandle(context.Background(), OperationRequest{Client: 1}); !errors.Is(err, ErrClientCastFailed) {
			t.Fatalf("expected %v, got %v", ErrClientCastFailed, err)
		}
	})

	t.Run("config decode failure", func(t *testing.T) {
		t.Parallel()

		if _, err := reg.IngestHandle(context.Background(), OperationRequest{Client: "User", Config: json.RawMessage(`not json`)}); !errors.Is(err, ErrOperationConfigInvalid) {
			t.Fatalf("expected %v, got %v", ErrOperationConfigInvalid, err)
		}
	})
}

func TestOperationRefHandlesRequest(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")

	ref := NewOperationRef[refTestSwitchConfig]("sweep").HandlesRequest(func(_ context.Context, _ OperationRequest, cfg refTestSwitchConfig) (json.RawMessage, error) {
		return json.Marshal(cfg.Limit)
	})

	reg := ref.Registration(definition, OperationRegistration{})

	if reg.Handle == nil || reg.ClientRef.Valid() {
		t.Fatalf("expected a client-less Handle projected, got %+v", reg)
	}

	got, err := reg.Handle(context.Background(), OperationRequest{Config: json.RawMessage(`{"limit":5}`)})
	if err != nil || string(got) != `5` {
		t.Fatalf("Handle() = %s, %v", got, err)
	}

	if _, err := reg.Handle(context.Background(), OperationRequest{Config: json.RawMessage(`[]`)}); !errors.Is(err, ErrOperationConfigInvalid) {
		t.Fatalf("expected %v, got %v", ErrOperationConfigInvalid, err)
	}
}

func TestOperationRefConfigDisabled(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")

	t.Run("switchable config projects the switch", func(t *testing.T) {
		t.Parallel()

		for _, ref := range []OperationRef[refTestSwitchConfig]{NewOperationRef[refTestSwitchConfig]("sync"), OperationRefOf[refTestSwitchConfig]()} {
			reg := ref.Registration(definition, OperationRegistration{})

			if reg.ConfigDisabled == nil {
				t.Fatalf("expected ConfigDisabled projected for %s", ref.Name())
			}

			if !reg.ConfigDisabled(json.RawMessage(`{"disable":true}`)) {
				t.Fatal("expected a set switch to disable the operation")
			}

			if reg.ConfigDisabled(json.RawMessage(`{"disable":false,"limit":1}`)) || reg.ConfigDisabled(nil) {
				t.Fatal("expected an unset or absent switch to leave the operation enabled")
			}

			if reg.ConfigDisabled(json.RawMessage(`not json`)) {
				t.Fatal("expected an undecodable section to leave the operation enabled")
			}

			if reg.Disabled != nil || reg.ConfigResolver != nil {
				t.Fatal("expected Registration not to touch Disabled or ConfigResolver")
			}
		}
	})

	t.Run("config without a switch projects nothing", func(t *testing.T) {
		t.Parallel()

		if reg := NewOperationRef[refTestInput]("plain").Registration(definition, OperationRegistration{}); reg.ConfigDisabled != nil {
			t.Fatal("expected no ConfigDisabled for a config without a switch")
		}
	})
}

func TestOperationRefReplacing(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")
	base := NewOperationRef[refTestInput]("current")
	replaced := base.Replacing(NewOperationRef[struct{}]("zeta")).Replacing(NewOperationRef[refTestRetiredInput]("alpha")).Replacing(NewOperationRef[struct{}]("zeta"))

	if got := replaced.Registration(definition, OperationRegistration{}).Replaces; !slices.Equal(got, []string{"alpha", "zeta"}) {
		t.Fatalf("Replaces = %v, want sorted unique [alpha zeta]", got)
	}

	if got := base.Registration(definition, OperationRegistration{}).Replaces; got != nil {
		t.Fatalf("expected the source ref to stay without replacements, got %v", got)
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

		if string(reg.StoredSchema) != string(plain.Schema()) {
			t.Fatalf("expected StoredSchema from the ref, got %s", reg.StoredSchema)
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

		if string(reg.StoredSchema) != string(plain.Schema()) {
			t.Fatalf("expected StoredSchema from the ref, got %s", reg.StoredSchema)
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

		reg := NewOperationRef[refTestInput]("refTestInput").Handles(client, func(context.Context, OperationRequest, string, refTestInput) (json.RawMessage, error) {
			return nil, nil
		}).Registration(definition, base)

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

func TestOperationRefHandlesDoesNotAliasTheReceiver(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")
	base := NewOperationRef[refTestInput]("refTestInput")
	bound := base.Handles(NewClientRef[string]("client"), func(context.Context, OperationRequest, string, refTestInput) (json.RawMessage, error) {
		return nil, nil
	})

	if reg := base.Registration(definition, OperationRegistration{}); reg.ClientRef.Valid() || reg.Handle != nil {
		t.Fatal("expected the source ref to stay unbound")
	}

	if reg := bound.Registration(definition, OperationRegistration{}); !reg.ClientRef.Valid() || reg.Handle == nil {
		t.Fatal("expected the bound ref to carry the client and handler")
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

func TestWebhookRefReplacing(t *testing.T) {
	t.Parallel()

	base := NewWebhookRef("installation.events")
	replaced := base.Replacing(NewWebhookRef("zeta.events")).Replacing(NewWebhookRef("alpha.events")).Replacing(NewWebhookRef("zeta.events"))

	if got := replaced.Registration(WebhookRegistration{}).Replaces; !slices.Equal(got, []string{"alpha.events", "zeta.events"}) {
		t.Fatalf("Replaces = %v, want sorted unique [alpha.events zeta.events]", got)
	}

	if got := base.Registration(WebhookRegistration{}).Replaces; got != nil {
		t.Fatalf("expected the source ref to stay without replacements, got %v", got)
	}

	if replaced.Name() != "installation.events" {
		t.Fatalf("expected Replacing to keep the contract name, got %q", replaced.Name())
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

		reg := ref.Registration(ConnectionRegistration{Name: "Token", CredentialRefs: []CredentialSlotID{other.ID()}})

		if reg.CredentialRef != cred.ID() {
			t.Fatalf("CredentialRef = %q, want %q", reg.CredentialRef, cred.ID())
		}

		if !slices.Equal(reg.ClientRefs, []ClientID{client.ID()}) {
			t.Fatalf("ClientRefs = %v", reg.ClientRefs)
		}

		if !slices.Equal(reg.CredentialRefs, []CredentialSlotID{cred.ID()}) {
			t.Fatalf("expected CredentialRefs derived from the selecting slot, got %v", reg.CredentialRefs)
		}

		if reg.Name != "Token" {
			t.Fatalf("expected base fields preserved, got %+v", reg)
		}
	})

	t.Run("projects the credential onto the disconnect flow without aliasing base", func(t *testing.T) {
		t.Parallel()

		disconnect := &DisconnectRegistration{Description: "teardown"}
		reg := ref.Registration(ConnectionRegistration{Disconnect: disconnect})

		if reg.Disconnect == nil || reg.Disconnect.CredentialRef != cred.ID() || reg.Disconnect.Description != "teardown" {
			t.Fatalf("Disconnect = %+v", reg.Disconnect)
		}

		if disconnect.CredentialRef != (CredentialSlotID{}) {
			t.Fatal("expected the authored disconnect registration left unmodified")
		}
	})

	t.Run("leaves an absent disconnect absent", func(t *testing.T) {
		t.Parallel()

		if reg := ref.Registration(ConnectionRegistration{}); reg.Disconnect != nil {
			t.Fatalf("expected nil Disconnect, got %+v", reg.Disconnect)
		}
	})
}

func TestClientRefHealthCheck(t *testing.T) {
	t.Parallel()

	client := NewClientRef[string]("client")
	check := client.HealthCheck(func(_ context.Context, _ OperationRequest, c string) (json.RawMessage, error) {
		return json.Marshal(c)
	})

	if check.ClientRef != client.ID() {
		t.Fatalf("ClientRef = %v, want %v", check.ClientRef, client.ID())
	}

	got, err := check.Handle(context.Background(), OperationRequest{Client: "live"})
	if err != nil || string(got) != `"live"` {
		t.Fatalf("Handle() = %s, %v", got, err)
	}

	if _, err := check.Handle(context.Background(), OperationRequest{Client: 1}); !errors.Is(err, ErrClientCastFailed) {
		t.Fatalf("expected %v, got %v", ErrClientCastFailed, err)
	}
}

func TestCredentialHealthCheck(t *testing.T) {
	t.Parallel()

	check := CredentialHealthCheck(func(_ context.Context, req OperationRequest) (json.RawMessage, error) {
		return json.Marshal(len(req.Credentials))
	})

	if check.ClientRef.Valid() {
		t.Fatal("expected a credential health check to carry no client")
	}

	got, err := check.Handle(context.Background(), OperationRequest{Credentials: CredentialBindings{{Ref: NewCredentialSlotID("slot")}}})
	if err != nil || string(got) != `1` {
		t.Fatalf("Handle() = %s, %v", got, err)
	}
}

func TestUserInputRefRegistration(t *testing.T) {
	t.Parallel()

	t.Run("plain layout projects only the schema", func(t *testing.T) {
		t.Parallel()

		reg := NewUserInputRef[refTestInput]("refTestInput").Registration()

		if jsonx.SchemaID(reg.Schema) != "refTestInput" {
			t.Fatalf("expected the reflected schema, got %s", reg.Schema)
		}

		if len(reg.Replaces) != 0 || reg.Convert != nil || reg.Backfill != nil {
			t.Fatalf("expected no lifecycle on a plain layout, got %+v", reg)
		}
	})

	t.Run("lifecycle layout projects convert and backfill", func(t *testing.T) {
		t.Parallel()

		reg := NewUserInputRef[refTestInput]("refTestInput").Replacing(NewUserInputRef[refTestRetiredInput]("refTestRetiredInput"), func(r refTestRetiredInput) refTestInput {
			return refTestInput{Region: r.Zone}
		}).Backfilled(func(_ context.Context, _ InstallationRequest, in *refTestInput) error {
			in.Region = "derived"

			return nil
		}).Registration()

		if !slices.Equal(reg.Replaces, []string{"refTestRetiredInput"}) {
			t.Fatalf("Replaces = %v", reg.Replaces)
		}

		converted, err := reg.Convert(json.RawMessage(`{"zone":"eu"}`))
		if err != nil || string(converted) != `{"region":"eu"}` {
			t.Fatalf("Convert() = %s, %v", converted, err)
		}

		filled, err := reg.Backfill(context.Background(), InstallationRequest{}, json.RawMessage(`{"region":""}`))
		if err != nil || string(filled) != `{"region":"derived"}` {
			t.Fatalf("Backfill() = %s, %v", filled, err)
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
