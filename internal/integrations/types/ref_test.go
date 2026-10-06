package types //nolint:revive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/theopenlane/core/v2/pkg/gala"
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
	OperationSettings
	Region string `json:"region" jsonschema:"required"`
}

// refTestRetiredInput is the shape an earlier definition version stored the region under
type refTestRetiredInput struct {
	Zone string `json:"zone"`
}

// refTestUpgradeCredential maps a payload stored under the retired credential slot onto the current layout
func refTestUpgradeCredential(_ context.Context, _ InstallationRequest, from string, stored json.RawMessage) (refTestCredential, error) {
	switch from {
	case "refTestRetiredCredential":
		old, err := jsonx.Decode[refTestRetiredCredential](stored)
		if err != nil {
			return refTestCredential{}, err
		}

		return refTestCredential{Token: old.AccessToken}, nil
	default:
		return jsonx.Decode[refTestCredential](stored)
	}
}

// refTestUpgradeInput maps a document stored under the retired user input layout onto the current one
func refTestUpgradeInput(_ context.Context, _ InstallationRequest, from string, stored json.RawMessage) (refTestInput, error) {
	switch from {
	case "refTestRetiredInput":
		old, err := jsonx.Decode[refTestRetiredInput](stored)
		if err != nil {
			return refTestInput{}, err
		}

		return refTestInput{Region: old.Zone}, nil
	default:
		return jsonx.Decode[refTestInput](stored)
	}
}

// refTestConfig is an operation config type the operation ref tests reflect
type refTestConfig struct {
	OperationSettings
	// Limit bounds the number of records the operation reads
	Limit int `json:"limit"`
}

// refTestRetiredConfig is the shape an earlier definition version stored the limit under
type refTestRetiredConfig struct {
	OperationSettings
	// Max is the retired name of the limit
	Max int `json:"max"`
}

// refTestZetaConfig is an empty retired config named to sort last
type refTestZetaConfig struct {
	OperationSettings
}

// refTestAlphaConfig is an empty retired config named to sort first
type refTestAlphaConfig struct {
	OperationSettings
}

// refTestUpgradeConfig maps a document stored under the retired operation onto the current config layout
func refTestUpgradeConfig(_ context.Context, _ InstallationRequest, from string, stored json.RawMessage) (refTestConfig, error) {
	switch from {
	case "refTestRetiredConfig":
		old, err := jsonx.Decode[refTestRetiredConfig](stored)
		if err != nil {
			return refTestConfig{}, err
		}

		return refTestConfig{OperationSettings: old.OperationSettings, Limit: old.Max}, nil
	default:
		return jsonx.Decode[refTestConfig](stored)
	}
}

func TestCredentialRefUpgraded(t *testing.T) {
	t.Parallel()

	retired := NewCredentialRef[refTestRetiredCredential]("refTestRetiredCredential")
	reg := NewCredentialRef[refTestCredential]("refTestCredential").Replacing(retired).Upgraded(refTestUpgradeCredential).Registration(CredentialRegistration{})

	tests := []struct {
		name    string
		from    CredentialSlotID
		payload string
		want    string
	}{
		{
			name:    "payload stored under the retired slot is mapped",
			from:    retired.ID(),
			payload: `{"accessToken":"t"}`,
			want:    `{"token":"t"}`,
		},
		{
			name:    "payload stored under the current slot passes through",
			from:    reg.Ref,
			payload: `{"token":"t"}`,
			want:    `{"token":"t"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := reg.Stored.Upgrade(context.Background(), InstallationRequest{}, tc.from.String(), json.RawMessage(tc.payload))
			if err != nil {
				t.Fatalf("Upgrade() error = %v", err)
			}

			if string(got) != tc.want {
				t.Fatalf("Upgrade() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestCredentialRefReplacingDoesNotAliasTheSource(t *testing.T) {
	t.Parallel()

	base := NewCredentialRef[refTestCredential]("refTestCredential").Replacing(NewCredentialRef[refTestRetiredCredential]("refTestRetiredCredential"))
	extended := base.Replacing(NewCredentialRef[struct{}]("other"))

	if got := len(base.Replaces()); got != 1 {
		t.Fatalf("expected the source ref to keep one replacement, got %d", got)
	}

	if got := extended.Replaces(); len(got) != 2 || got[0] != NewCredentialSlotID("other") || got[1] != NewCredentialSlotID("refTestRetiredCredential") {
		t.Fatalf("expected sorted replacements on the extended ref, got %v", got)
	}
}

func TestCredentialRefWithoutUpgradeProjectsNone(t *testing.T) {
	t.Parallel()

	reg := NewCredentialRef[refTestCredential]("refTestCredential").Registration(CredentialRegistration{})

	if reg.Stored.Upgrade != nil {
		t.Fatal("expected no upgrade projected on a plain slot")
	}
}

func TestCredentialRefUpgradedReceivesTheInstallationRequest(t *testing.T) {
	t.Parallel()

	var seen InstallationRequest

	reg := NewCredentialRef[refTestCredential]("refTestCredential").Upgraded(func(_ context.Context, req InstallationRequest, _ string, _ json.RawMessage) (refTestCredential, error) {
		seen = req

		return refTestCredential{Token: "derived"}, nil
	}).Registration(CredentialRegistration{})

	got, err := reg.Stored.Upgrade(context.Background(), InstallationRequest{Input: json.RawMessage(`{"marker":true}`)}, "refTestCredential", json.RawMessage(`{"token":""}`))
	if err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}

	if string(got) != `{"token":"derived"}` {
		t.Fatalf("Upgrade() = %s", got)
	}

	if string(seen.Input) != `{"marker":true}` {
		t.Fatalf("expected the upgrade to receive the explicit request, got %s", seen.Input)
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

func TestUserInputRefOfDerivesNameFromSchema(t *testing.T) {
	t.Parallel()

	ref := UserInputRefOf[refTestInput]()

	if ref.Name() != "refTestInput" {
		t.Fatalf("Name() = %q", ref.Name())
	}

	if jsonx.SchemaID(ref.Registration().Schema) != "refTestInput" {
		t.Fatalf("expected the reflected schema retained, got %s", ref.Registration().Schema)
	}
}

func TestUserInputRefUpgraded(t *testing.T) {
	t.Parallel()

	retired := UserInputRefOf[refTestRetiredInput]()
	reg := UserInputRefOf[refTestInput]().Upgraded(refTestUpgradeInput).Registration()

	tests := []struct {
		name    string
		from    string
		payload string
		want    string
	}{
		{
			name:    "document stored under the retired layout is mapped",
			from:    retired.Name(),
			payload: `{"zone":"eu"}`,
			want:    `{"region":"eu"}`,
		},
		{
			name:    "document stored under the current layout passes through",
			from:    reg.Name,
			payload: `{"region":"eu"}`,
			want:    `{"region":"eu"}`,
		},
		{
			name:    "document stored without a layout passes through",
			from:    "",
			payload: `{"region":"eu"}`,
			want:    `{"region":"eu"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := reg.Upgrade(context.Background(), InstallationRequest{}, tc.from, json.RawMessage(tc.payload))
			if err != nil {
				t.Fatalf("Upgrade() error = %v", err)
			}

			if string(got) != tc.want {
				t.Fatalf("Upgrade() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestUserInputRefUpgradedReceivesTheInstallationRequest(t *testing.T) {
	t.Parallel()

	var seen InstallationRequest

	reg := UserInputRefOf[refTestInput]().Upgraded(func(_ context.Context, req InstallationRequest, _ string, _ json.RawMessage) (refTestInput, error) {
		seen = req

		return refTestInput{Region: "derived"}, nil
	}).Registration()

	got, err := reg.Upgrade(context.Background(), InstallationRequest{Input: json.RawMessage(`{"marker":true}`)}, "refTestInput", json.RawMessage(`{"region":""}`))
	if err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}

	if string(got) != `{"region":"derived"}` {
		t.Fatalf("Upgrade() = %s", got)
	}

	if string(seen.Input) != `{"marker":true}` {
		t.Fatalf("expected the upgrade to receive the explicit request, got %s", seen.Input)
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

	ref := NewOperationPayload[refTestCredential]("refTestCredential")
	if ref.Name() != "refTestCredential" {
		t.Fatalf("OperationRef.Name() = %q", ref.Name())
	}
}

func TestOperationPayloadRegistrationHasNoInput(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")

	reg := OperationPayloadOf[refTestCredential]().
		Upgraded(func(_ context.Context, _ InstallationRequest, _ string, stored json.RawMessage) (refTestCredential, error) {
			return jsonx.Decode[refTestCredential](stored)
		}).
		Registration(definition)

	if reg.Name != "refTestCredential" || reg.Stored {
		t.Fatalf("expected a payload operation named from its schema with no stored input, got %+v", reg)
	}

	if jsonx.SchemaID(reg.Input.Schema) != "refTestCredential" {
		t.Fatalf("expected the payload schema as the input schema, got %s", reg.Input.Schema)
	}

	stored := OperationRefOf[refTestConfig]().Registration(definition)

	if !stored.Stored || jsonx.SchemaID(stored.Input.Schema) != "refTestConfig" {
		t.Fatalf("expected a stored-input operation carrying its reflected schema, got %+v", stored)
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

	ref := NewOperationRef[refTestConfig]("sync").Handles(client, func(_ context.Context, _ OperationRequest, c string, cfg refTestConfig) (json.RawMessage, error) {
		return json.Marshal(struct {
			Client string `json:"client"`
			Limit  int    `json:"limit"`
		}{Client: c, Limit: cfg.Limit})
	})

	reg := ref.Registration(definition)

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

	ref := NewOperationRef[refTestConfig]("sync").Ingests(client, func(_ context.Context, _ OperationRequest, c string, cfg refTestConfig) ([]IngestPayloadSet, error) {
		return []IngestPayloadSet{{Schema: c, Envelopes: make([]MappingEnvelope, cfg.Limit)}}, nil
	})

	reg := ref.Registration(definition)

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

	ref := NewOperationRef[refTestConfig]("sweep").HandlesRequest(func(_ context.Context, _ OperationRequest, cfg refTestConfig) (json.RawMessage, error) {
		return json.Marshal(cfg.Limit)
	})

	reg := ref.Registration(definition)

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

func TestOperationRefReplacing(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")
	base := OperationRefOf[refTestConfig]()
	zeta := OperationRefOf[refTestZetaConfig]()
	alpha := OperationRefOf[refTestAlphaConfig]()
	replaced := base.Replacing(zeta).Replacing(alpha).Replacing(zeta)

	reg := replaced.Registration(definition)

	if got := reg.Replaces; !slices.Equal(got, []string{"refTestAlphaConfig", "refTestZetaConfig"}) {
		t.Fatalf("Replaces = %v, want sorted unique names", got)
	}

	if reg.Input.Upgrade != nil {
		t.Fatal("expected no upgrade projected from identity replacements alone")
	}

	source := base.Registration(definition)

	if source.Replaces != nil {
		t.Fatalf("expected the source ref to stay without replacements, got %+v", source.Replaces)
	}
}

func TestOperationRefUpgraded(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")
	retired := OperationRefOf[refTestRetiredConfig]()
	reg := OperationRefOf[refTestConfig]().Replacing(retired).Upgraded(refTestUpgradeConfig).Registration(definition)

	tests := []struct {
		name    string
		from    string
		payload string
		want    string
	}{
		{name: "settings survive the upgrade", from: retired.Name(), payload: `{"disable":true,"filterExpr":"x","max":3}`, want: `{"disable":true,"filterExpr":"x","limit":3}`},
		{name: "config without settings upgrades", from: retired.Name(), payload: `{"max":3}`, want: `{"limit":3}`},
		{name: "document stored under the current operation passes through with its settings", from: reg.Name, payload: `{"disable":true,"limit":3}`, want: `{"disable":true,"limit":3}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := reg.Input.Upgrade(context.Background(), InstallationRequest{}, tc.from, json.RawMessage(tc.payload))
			if err != nil {
				t.Fatalf("Upgrade() error = %v", err)
			}

			if string(got) != tc.want {
				t.Fatalf("Upgrade() = %s, want %s", got, tc.want)
			}
		})
	}

	plain := OperationRefOf[refTestConfig]().Registration(definition)

	if plain.Input.Upgrade != nil {
		t.Fatal("expected no upgrade projected on a plain operation")
	}
}

func TestOperationRefChainProjectsBehavior(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")
	schedule := &gala.Schedule{}

	ref := OperationRefOf[refTestConfig]().
		Description("sync").
		Policy(ExecutionPolicy{Reconcile: true, Snapshot: true}).
		Ingest(IngestContract{Schema: "User"}, IngestContract{Schema: "Group"}).
		Permissions("read:users", "read:groups").
		Schedule(schedule).
		SkipDefaultLookback().
		RateLimit(RateLimitPolicy{Limit: 2}).
		Internal().
		CustomerSelectable(false).
		RequiresPaymentMethod().
		DisabledForAll(true)

	reg := ref.Registration(definition)

	if reg.Description != "sync" {
		t.Fatalf("expected the chain description projected, got %+v", reg)
	}

	if !reg.Policy.Reconcile || !reg.Policy.Snapshot || reg.Policy.Inline {
		t.Fatalf("expected the chain policy projected, got %+v", reg.Policy)
	}

	if len(reg.Ingest) != 2 || reg.Ingest[0].Schema != "User" || reg.Ingest[1].Schema != "Group" {
		t.Fatalf("Ingest = %+v", reg.Ingest)
	}

	if !slices.Equal(reg.RequiredPermissions, []string{"read:users", "read:groups"}) {
		t.Fatalf("RequiredPermissions = %v", reg.RequiredPermissions)
	}

	if reg.Schedule != schedule || !reg.SkipDefaultLookback || reg.RateLimit == nil || reg.RateLimit.Limit != 2 {
		t.Fatalf("expected schedule, lookback, and rate limit projected, got %+v", reg)
	}

	if !reg.Internal || reg.CustomerSelectable == nil || *reg.CustomerSelectable || !reg.RequiresPaymentMethod || !reg.DisabledForAll {
		t.Fatalf("expected internal, selectable, payment, and disabled flags projected, got %+v", reg)
	}

	if reg.Replaces != nil {
		t.Fatalf("expected no replacements without Replacing, got %v", reg.Replaces)
	}

	plain := OperationRefOf[refTestConfig]().Registration(definition)

	if plain.Description != "" || plain.Policy != (ExecutionPolicy{}) || plain.Ingest != nil || plain.CustomerSelectable != nil || plain.Schedule != nil || plain.RateLimit != nil {
		t.Fatalf("expected an unconfigured ref to project zero behavior, got %+v", plain)
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

		if string(reg.Stored.Schema) != string(plain.Schema()) {
			t.Fatalf("expected Stored.Schema from the ref, got %s", reg.Stored.Schema)
		}

		if len(reg.Replaces) != 0 || reg.Stored.Upgrade != nil {
			t.Fatalf("expected no lifecycle on a plain slot, got %+v", reg)
		}
	})

	t.Run("lifecycle slot projects replaced slots and the upgrade", func(t *testing.T) {
		t.Parallel()

		ref := plain.Replacing(retired).Upgraded(refTestUpgradeCredential)

		reg := ref.Registration(base)

		if !slices.Equal(reg.Replaces, []CredentialSlotID{retired.ID()}) {
			t.Fatalf("Replaces = %v", reg.Replaces)
		}

		if reg.Stored.Upgrade == nil {
			t.Fatal("expected Upgrade projected")
		}

		upgraded, err := reg.Stored.Upgrade(context.Background(), InstallationRequest{}, retired.ID().String(), json.RawMessage(`{"accessToken":"t"}`))
		if err != nil || string(upgraded) != `{"token":"t"}` {
			t.Fatalf("Upgrade() = %s, %v", upgraded, err)
		}
	})

	t.Run("auth-managed slot keeps an empty schema", func(t *testing.T) {
		t.Parallel()

		reg := plain.Registration(CredentialRegistration{Name: "OAuth"})

		if len(reg.Schema) != 0 {
			t.Fatalf("expected Registration not to fill Schema, got %s", reg.Schema)
		}

		if string(reg.Stored.Schema) != string(plain.Schema()) {
			t.Fatalf("expected Stored.Schema from the ref, got %s", reg.Stored.Schema)
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

	t.Run("without a client", func(t *testing.T) {
		t.Parallel()

		ref := NewOperationRef[refTestInput]("refTestInput").Description("sync")
		reg := ref.Registration(definition)

		if reg.Name != "refTestInput" {
			t.Fatalf("Name = %q", reg.Name)
		}

		if string(reg.Topic) != "integration.run.def_001.refTestInput" {
			t.Fatalf("Topic = %q", reg.Topic)
		}

		if !reg.Stored || reg.Input.Name != "refTestInput" || string(reg.Input.Schema) != string(ref.Schema()) {
			t.Fatalf("Input = %+v, want the reflected layout", reg.Input)
		}

		if reg.Input.Upgrade != nil {
			t.Fatalf("expected no upgrade on a plain operation, got %+v", reg.Input)
		}

		if reg.ClientRef.Valid() {
			t.Fatal("expected no client ref")
		}

		if reg.Description != "sync" {
			t.Fatalf("expected the ref description projected, got %+v", reg)
		}
	})

	t.Run("with a client", func(t *testing.T) {
		t.Parallel()

		reg := NewOperationRef[refTestInput]("refTestInput").Handles(client, func(context.Context, OperationRequest, string, refTestInput) (json.RawMessage, error) {
			return nil, nil
		}).Registration(definition)

		if reg.ClientRef != client.ID() {
			t.Fatalf("ClientRef = %v, want %v", reg.ClientRef, client.ID())
		}
	})

	t.Run("health check declared on the ref is projected", func(t *testing.T) {
		t.Parallel()

		reg := NewOperationRef[refTestInput]("refTestInput").HealthCheck(CredentialHealthCheck(func(context.Context, OperationRequest) (json.RawMessage, error) {
			return json.RawMessage(`"ok"`), nil
		})).Registration(definition)

		if reg.HealthCheck == nil {
			t.Fatal("expected HealthCheck projected")
		}

		got, err := reg.HealthCheck(context.Background(), OperationRequest{})
		if err != nil || string(got) != `"ok"` {
			t.Fatalf("HealthCheck() = %s, %v", got, err)
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

	if reg := base.Registration(definition); reg.ClientRef.Valid() || reg.Handle != nil {
		t.Fatal("expected the source ref to stay unbound")
	}

	if reg := bound.Registration(definition); !reg.ClientRef.Valid() || reg.Handle == nil {
		t.Fatal("expected the bound ref to carry the client and handler")
	}
}

func TestWebhookRefRegistration(t *testing.T) {
	t.Parallel()

	reg := NewWebhookRef("installation.events").Replacing(NewWebhookRef("old")).Registration(WebhookRegistration{StaticRoute: "/hooks", Replaces: []string{"stale"}})

	if reg.Name != "installation.events" {
		t.Fatalf("Name = %q", reg.Name)
	}

	if reg.StaticRoute != "/hooks" || !slices.Equal(reg.Replaces, []string{"old"}) {
		t.Fatalf("expected the base route preserved and replacements taken from the ref only, got %+v", reg)
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

	t.Run("plain layout projects the name and schema", func(t *testing.T) {
		t.Parallel()

		reg := UserInputRefOf[refTestInput]().Registration()

		if reg.Name != "refTestInput" || jsonx.SchemaID(reg.Schema) != "refTestInput" {
			t.Fatalf("expected the reflected name and schema, got %+v", reg)
		}

		if reg.Upgrade != nil {
			t.Fatalf("expected no upgrade on a plain layout, got %+v", reg)
		}
	})

	t.Run("upgraded layout projects the upgrade", func(t *testing.T) {
		t.Parallel()

		retired := UserInputRefOf[refTestRetiredInput]()

		reg := UserInputRefOf[refTestInput]().Upgraded(refTestUpgradeInput).Registration()

		if reg.Upgrade == nil {
			t.Fatal("expected Upgrade projected")
		}

		upgraded, err := reg.Upgrade(context.Background(), InstallationRequest{}, retired.Name(), json.RawMessage(`{"zone":"eu"}`))
		if err != nil || string(upgraded) != `{"region":"eu"}` {
			t.Fatalf("Upgrade() = %s, %v", upgraded, err)
		}
	})
}

// errRefTestRegion is the semantic failure the Validated tests expect
var errRefTestRegion = errors.New("region is not reachable")

// refTestValidateRegion rejects any region other than eu
func refTestValidateRegion(_ context.Context, _ InstallationRequest, input *refTestInput) error {
	if input.Region != "eu" {
		return errRefTestRegion
	}

	return nil
}

func TestUserInputRefValidated(t *testing.T) {
	t.Parallel()

	reg := UserInputRefOf[refTestInput]().Validated(refTestValidateRegion).Registration()

	if err := reg.Validate(context.Background(), InstallationRequest{}, json.RawMessage(`{"region":"eu"}`)); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if err := reg.Validate(context.Background(), InstallationRequest{}, json.RawMessage(`{"region":"us"}`)); !errors.Is(err, errRefTestRegion) {
		t.Fatalf("Validate() error = %v, want %v", err, errRefTestRegion)
	}

	if UserInputRefOf[refTestInput]().Registration().Validate != nil {
		t.Fatal("expected no validation projected on a plain layout")
	}
}

func TestCredentialRefValidated(t *testing.T) {
	t.Parallel()

	reg := CredentialRefOf[refTestCredential]().Validated(func(_ context.Context, _ InstallationRequest, credential *refTestCredential) error {
		if credential.Token == "" {
			return errRefTestRegion
		}

		return nil
	}).Registration(CredentialRegistration{})

	if err := reg.Stored.Validate(context.Background(), InstallationRequest{}, json.RawMessage(`{"token":"t"}`)); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if err := reg.Stored.Validate(context.Background(), InstallationRequest{}, json.RawMessage(`{"token":""}`)); !errors.Is(err, errRefTestRegion) {
		t.Fatalf("Validate() error = %v, want %v", err, errRefTestRegion)
	}

	if CredentialRefOf[refTestCredential]().Registration(CredentialRegistration{}).Stored.Validate != nil {
		t.Fatal("expected no validation projected on a plain credential")
	}
}

func TestOperationRefValidatedIgnoresSettingsKeys(t *testing.T) {
	t.Parallel()

	definition := NewDefinitionRef("def_001")

	reg := OperationRefOf[refTestConfig]().Validated(func(_ context.Context, _ InstallationRequest, config *refTestConfig) error {
		if config.Limit > 10 {
			return errRefTestRegion
		}

		return nil
	}).Registration(definition)

	if err := reg.Input.Validate(context.Background(), InstallationRequest{}, json.RawMessage(`{"disable":true,"filterExpr":"x","limit":3}`)); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if err := reg.Input.Validate(context.Background(), InstallationRequest{}, json.RawMessage(`{"limit":11}`)); !errors.Is(err, errRefTestRegion) {
		t.Fatalf("Validate() error = %v, want %v", err, errRefTestRegion)
	}

	if OperationRefOf[refTestConfig]().Registration(definition).Input.Validate != nil {
		t.Fatal("expected no validation projected on a plain operation")
	}
}
