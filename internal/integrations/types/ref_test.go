package types //nolint:revive

import (
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

// refTestInput is the user input type the user input ref tests reflect
type refTestInput struct {
	OperationSettings
	Region string `json:"region" jsonschema:"required"`
}

// refTestRetiredInput is the shape an earlier definition version stored the region under
type refTestRetiredInput struct {
	Zone string `json:"zone"`
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

	got, err := reg.Upgrade(context.Background(), InstallationRequest{UserInput: json.RawMessage(`{"marker":true}`)}, "refTestInput", json.RawMessage(`{"region":""}`))
	if err != nil {
		t.Fatalf("Upgrade() error = %v", err)
	}

	if string(got) != `{"region":"derived"}` {
		t.Fatalf("Upgrade() = %s", got)
	}

	if string(seen.UserInput) != `{"marker":true}` {
		t.Fatalf("expected the upgrade to receive the explicit request, got %s", seen.UserInput)
	}
}

func TestDefinitionRefID(t *testing.T) {
	t.Parallel()

	ref := NewDefinitionRef("def_01TEST0000000000000000001")
	if ref.ID() != "def_01TEST0000000000000000001" {
		t.Fatalf("DefinitionRef.ID() = %q", ref.ID())
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

	reg := OperationPayloadOf[refTestCredential]().
		Upgraded(func(_ context.Context, _ InstallationRequest, _ string, stored json.RawMessage) (refTestCredential, error) {
			return jsonx.Decode[refTestCredential](stored)
		}).
		Registration()

	if reg.Name != "refTestCredential" || reg.Stored {
		t.Fatalf("expected a payload operation named from its schema with no stored input, got %+v", reg)
	}

	if jsonx.SchemaID(reg.Input.Schema) != "refTestCredential" {
		t.Fatalf("expected the payload schema as the input schema, got %s", reg.Input.Schema)
	}

	stored := OperationRefOf[refTestConfig]().Registration()

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

func TestOperationRefHandles(t *testing.T) {
	t.Parallel()

	ref := NewOperationRef[refTestConfig]("sync").Handles(func(_ context.Context, _ OperationRequest, c string, cfg refTestConfig) (json.RawMessage, error) {
		return json.Marshal(struct {
			Client string `json:"client"`
			Limit  int    `json:"limit"`
		}{Client: c, Limit: cfg.Limit})
	})

	reg := ref.Registration()

	if reg.Handle == nil || reg.IngestHandle != nil {
		t.Fatalf("expected only Handle projected, got %+v", reg)
	}

	if reg.ClientRef != "string" {
		t.Fatalf("ClientRef = %q, want %q", reg.ClientRef, "string")
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

	ref := NewOperationRef[refTestConfig]("sync").Ingests(func(_ context.Context, _ OperationRequest, c string, cfg refTestConfig) ([]IngestPayloadSet, error) {
		return []IngestPayloadSet{{Schema: c, Envelopes: make([]MappingEnvelope, cfg.Limit)}}, nil
	})

	reg := ref.Registration()

	if reg.IngestHandle == nil || reg.Handle != nil {
		t.Fatalf("expected only IngestHandle projected, got %+v", reg)
	}

	if reg.ClientRef != "string" {
		t.Fatalf("ClientRef = %q, want %q", reg.ClientRef, "string")
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

	ref := NewOperationRef[refTestConfig]("sweep").HandlesRequest(func(_ context.Context, _ OperationRequest, cfg refTestConfig) (json.RawMessage, error) {
		return json.Marshal(cfg.Limit)
	})

	reg := ref.Registration()

	if reg.Handle == nil || reg.ClientRef != "" {
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

	base := OperationRefOf[refTestConfig]()
	zeta := OperationRefOf[refTestZetaConfig]()
	alpha := OperationRefOf[refTestAlphaConfig]()
	replaced := base.Replacing(zeta).Replacing(alpha).Replacing(zeta)

	reg := replaced.Registration()

	if got := reg.Replaces; !slices.Equal(got, []string{"refTestAlphaConfig", "refTestZetaConfig"}) {
		t.Fatalf("Replaces = %v, want sorted unique names", got)
	}

	if reg.Input.Upgrade != nil {
		t.Fatal("expected no upgrade projected from identity replacements alone")
	}

	source := base.Registration()

	if source.Replaces != nil {
		t.Fatalf("expected the source ref to stay without replacements, got %+v", source.Replaces)
	}
}

func TestOperationRefUpgraded(t *testing.T) {
	t.Parallel()

	retired := OperationRefOf[refTestRetiredConfig]()
	reg := OperationRefOf[refTestConfig]().Replacing(retired).Upgraded(refTestUpgradeConfig).Registration()

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

	plain := OperationRefOf[refTestConfig]().Registration()

	if plain.Input.Upgrade != nil {
		t.Fatal("expected no upgrade projected on a plain operation")
	}
}

func TestOperationRefChainProjectsBehavior(t *testing.T) {
	t.Parallel()

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

	reg := ref.Registration()

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

	plain := OperationRefOf[refTestConfig]().Registration()

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

	t.Run("without a client", func(t *testing.T) {
		t.Parallel()

		ref := NewOperationRef[refTestInput]("refTestInput").Description("sync")
		reg := ref.Registration()

		if reg.Name != "refTestInput" {
			t.Fatalf("Name = %q", reg.Name)
		}

		if !reg.Stored || reg.Input.Name != "refTestInput" || string(reg.Input.Schema) != string(ref.Schema()) {
			t.Fatalf("Input = %+v, want the reflected layout", reg.Input)
		}

		if reg.Input.Upgrade != nil {
			t.Fatalf("expected no upgrade on a plain operation, got %+v", reg.Input)
		}

		if reg.ClientRef != "" {
			t.Fatal("expected no client ref")
		}

		if reg.Description != "sync" {
			t.Fatalf("expected the ref description projected, got %+v", reg)
		}
	})

	t.Run("with a client", func(t *testing.T) {
		t.Parallel()

		reg := NewOperationRef[refTestInput]("refTestInput").Handles(func(context.Context, OperationRequest, string, refTestInput) (json.RawMessage, error) {
			return nil, nil
		}).Registration()

		if reg.ClientRef != "string" {
			t.Fatalf("ClientRef = %q, want %q", reg.ClientRef, "string")
		}
	})

	t.Run("health check declared on the ref is projected", func(t *testing.T) {
		t.Parallel()

		reg := NewOperationRef[refTestInput]("refTestInput").HealthCheck(func(context.Context, OperationRequest, string) error {
			return nil
		}).Registration()

		if reg.HealthCheck == nil || reg.ClientRef != "string" {
			t.Fatalf("expected HealthCheck and its client projected, got %+v", reg)
		}

		if got, err := reg.HealthCheck(context.Background(), OperationRequest{Client: "live"}); err != nil || got != nil {
			t.Fatalf("HealthCheck() = %s, %v", got, err)
		}

		if _, err := reg.HealthCheck(context.Background(), OperationRequest{Client: 1}); !errors.Is(err, ErrClientCastFailed) {
			t.Fatalf("expected %v, got %v", ErrClientCastFailed, err)
		}
	})
}

func TestOperationRefHandlesDoesNotAliasTheReceiver(t *testing.T) {
	t.Parallel()

	base := NewOperationRef[refTestInput]("refTestInput")
	bound := base.Handles(func(context.Context, OperationRequest, string, refTestInput) (json.RawMessage, error) {
		return nil, nil
	})

	if reg := base.Registration(); reg.ClientRef != "" || reg.Handle != nil {
		t.Fatal("expected the source ref to stay unbound")
	}

	if reg := bound.Registration(); reg.ClientRef == "" || reg.Handle == nil {
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

	reg := NewWebhookEventRef[refTestInput]("created").Registration(WebhookEventRegistration{Ingest: []IngestContract{{Schema: "User"}}})

	if reg.Name != "created" {
		t.Fatalf("Name = %q", reg.Name)
	}

	if len(reg.Ingest) != 1 || reg.Ingest[0].Schema != "User" {
		t.Fatalf("expected base fields preserved, got %+v", reg)
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

func TestOperationRefValidatedIgnoresSettingsKeys(t *testing.T) {
	t.Parallel()

	reg := OperationRefOf[refTestConfig]().Validated(func(_ context.Context, _ InstallationRequest, config *refTestConfig) error {
		if config.Limit > 10 {
			return errRefTestRegion
		}

		return nil
	}).Registration()

	if err := reg.Input.Validate(context.Background(), InstallationRequest{}, json.RawMessage(`{"disable":true,"filterExpr":"x","limit":3}`)); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if err := reg.Input.Validate(context.Background(), InstallationRequest{}, json.RawMessage(`{"limit":11}`)); !errors.Is(err, errRefTestRegion) {
		t.Fatalf("Validate() error = %v, want %v", err, errRefTestRegion)
	}

	if OperationRefOf[refTestConfig]().Registration().Input.Validate != nil {
		t.Fatal("expected no validation projected on a plain operation")
	}
}

// refTestClientA is one client type the client conflict test binds
type refTestClientA struct{}

// refTestClientB is the other client type the client conflict test binds
type refTestClientB struct{}

// TestOperationRefClientConflict verifies handlers binding different client types keep the first client and record the second as the conflict
func TestOperationRefClientConflict(t *testing.T) {
	t.Parallel()

	consistent := OperationRefOf[refTestConfig]().
		HealthCheck(func(context.Context, OperationRequest, *refTestClientA) error { return nil }).
		Ingests(func(context.Context, OperationRequest, *refTestClientA, refTestConfig) ([]IngestPayloadSet, error) {
			return nil, nil
		}).
		Registration()

	if consistent.ClientRef != clientName[*refTestClientA]() || consistent.ClientConflict != "" {
		t.Fatalf("expected one client bound with no conflict, got ref %q conflict %q", consistent.ClientRef, consistent.ClientConflict)
	}

	mixed := OperationRefOf[refTestConfig]().
		HealthCheck(func(context.Context, OperationRequest, *refTestClientA) error { return nil }).
		Ingests(func(context.Context, OperationRequest, *refTestClientB, refTestConfig) ([]IngestPayloadSet, error) {
			return nil, nil
		}).
		Registration()

	if mixed.ClientRef != clientName[*refTestClientA]() || mixed.ClientConflict != clientName[*refTestClientB]() {
		t.Fatalf("expected the first client kept and the second recorded as the conflict, got ref %q conflict %q", mixed.ClientRef, mixed.ClientConflict)
	}
}
