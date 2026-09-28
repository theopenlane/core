package registry

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/samber/lo"

	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// versionSecondCredential is the credential type behind the extra slot added in the version test
type versionSecondCredential struct{}

// versionMetadata is the derived installation metadata type behind the surface tests
type versionMetadata struct {
	Tenant string `json:"tenant"`
}

// retiredUserInput is the user input layout an earlier definition version stored
type retiredUserInput struct {
	Zone string `json:"zone"`
}

var (
	// versionSecondCredentialRef is the extra credential slot added in the version test
	versionSecondCredentialRef = integrationtypes.CredentialRefOf[versionSecondCredential]()
	// retiredUserInputRef is the user input layout an earlier definition version stored
	retiredUserInputRef = integrationtypes.NewUserInputRef[retiredUserInput]("retiredUserInput")
	// surfaceUserInputRef is the current user input layout taking over the retired layout with a backfill
	surfaceUserInputRef = integrationtypes.NewUserInputRef[testUserInput]("testUserInput").
				Replacing(retiredUserInputRef, func(r retiredUserInput) testUserInput { return testUserInput{Region: r.Zone} }).
				Backfilled(func(context.Context, integrationtypes.InstallationRequest, *testUserInput) error { return nil })
)

// surfaceDefinition returns a definition exercising every surfaced kind, declared in reverse name order to prove sorting
func surfaceDefinition(id string) integrationtypes.Definition {
	def, clientRef := minimalDefinition(id)
	defRef := integrationtypes.NewDefinitionRef(id)

	def.UserInput = &integrationtypes.UserInputRegistration{
		Schema:   jsonx.SchemaFrom[testUserInput](),
		Replaces: surfaceUserInputRef.Replaces(),
		Convert:  surfaceUserInputRef.Convert,
		Backfill: surfaceUserInputRef.Backfill,
	}

	def.Connections = []integrationtypes.ConnectionRegistration{
		{
			CredentialRef: testCredentialRef.ID(),
			Integration: integrationtypes.NewInstallationRef(func(context.Context, integrationtypes.InstallationRequest) (versionMetadata, bool, error) {
				return versionMetadata{}, true, nil
			}).Registration(),
		},
	}

	def.Operations = []integrationtypes.OperationRegistration{
		{
			Name:         "sync.users",
			Replaces:     []string{"sync.people"},
			Topic:        defRef.OperationTopic("sync.users"),
			ClientRef:    clientRef.ID(),
			ConfigSchema: jsonx.SchemaFrom[testOperationConfig](),
			Handle:       newTestHandler(),
		},
		{
			Name:         "sync.groups",
			Topic:        defRef.OperationTopic("sync.groups"),
			ClientRef:    clientRef.ID(),
			ConfigSchema: jsonx.SchemaFrom[testOperationConfig](),
			Handle:       newTestHandler(),
		},
	}

	def.Webhooks = []integrationtypes.WebhookRegistration{
		{
			Name: "static",
			Event: func(integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
				return integrationtypes.WebhookReceivedEvent{}, nil
			},
		},
		{
			Name:     "events.v2",
			Replaces: []string{"events.v1"},
			Event: func(integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
				return integrationtypes.WebhookReceivedEvent{}, nil
			},
			Events: []integrationtypes.WebhookEventRegistration{
				{
					Name:   "member.removed",
					Topic:  defRef.WebhookEventTopic("member.removed"),
					Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil },
				},
				{
					Name:   "member.joined",
					Topic:  defRef.WebhookEventTopic("member.joined"),
					Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil },
				},
			},
		},
	}

	return def
}

func TestDefinitionSurface(t *testing.T) {
	t.Parallel()

	def := surfaceDefinition("surface-def")

	reg := New()
	if err := reg.Register(def); err != nil {
		t.Fatalf("register: %v", err)
	}

	registered, _ := reg.Definition(def.ID)
	surface := DefinitionSurface(registered)

	retired := retiredUserInputRef.Name()

	if surface.UserInput == nil || !surface.UserInput.Backfill || !slices.Equal(surface.UserInput.Replaces, []string{retired}) {
		t.Fatalf("UserInput = %+v, want backfilled schema replacing %s", surface.UserInput, retired)
	}

	if len(surface.UserInput.Schema) == 0 {
		t.Fatal("expected the user input schema to be surfaced")
	}

	if got := lo.Map(surface.Metadata, func(m SurfaceConnectionMetadata, _ int) string { return m.Connection }); !slices.Equal(got, []string{testCredentialRef.String()}) {
		t.Fatalf("Metadata connections = %v", got)
	}

	if len(surface.Metadata[0].Schema) == 0 {
		t.Fatal("expected the metadata schema to be surfaced")
	}

	if got := lo.Map(surface.Operations, func(o SurfaceNamed, _ int) string { return o.Name }); !slices.Equal(got, []string{"sync.groups", "sync.users"}) {
		t.Fatalf("Operations = %v, want sorted names", got)
	}

	if got := surface.Operations[1].Replaces; !slices.Equal(got, []string{"sync.people"}) {
		t.Fatalf("sync.users Replaces = %v", got)
	}

	if got := lo.Map(surface.Webhooks, func(w SurfaceWebhook, _ int) string { return w.Name }); !slices.Equal(got, []string{"events.v2", "static"}) {
		t.Fatalf("Webhooks = %v, want sorted names", got)
	}

	if got := surface.Webhooks[0].Replaces; !slices.Equal(got, []string{"events.v1"}) {
		t.Fatalf("events.v2 Replaces = %v", got)
	}

	if got := lo.Map(surface.Webhooks[0].Events, func(e SurfaceNamed, _ int) string { return e.Name }); !slices.Equal(got, []string{"member.joined", "member.removed"}) {
		t.Fatalf("events.v2 Events = %v, want sorted names", got)
	}

	if len(surface.Webhooks[1].Events) != 0 || surface.Webhooks[1].Replaces != nil {
		t.Fatalf("static webhook = %+v, want no events and no replacements", surface.Webhooks[1])
	}

	minimal, _ := minimalDefinition("minimal-def")

	encoded, err := json.Marshal(DefinitionSurface(minimal))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var keys map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &keys); err != nil {
		t.Fatalf("decode: %v", err)
	}

	for _, absent := range []string{"userInput", "metadata", "webhooks"} {
		if _, present := keys[absent]; present {
			t.Fatalf("expected %s to be omitted from a surface without it: %s", absent, encoded)
		}
	}
}

func TestVersionChangesWhenAnOperationIsRenamedAndIsStableAcrossOrdering(t *testing.T) {
	t.Parallel()

	base := surfaceDefinition("version-def")

	reordered := surfaceDefinition("version-def")
	slices.Reverse(reordered.Operations)
	slices.Reverse(reordered.Webhooks)
	slices.Reverse(reordered.Webhooks[0].Events)

	if versionOf(t, reordered) != versionOf(t, base) {
		t.Fatal("expected declaration order not to change the version")
	}

	renamed := surfaceDefinition("version-def")
	renamed.Operations[1].Name = "sync.teams"
	renamed.Operations[1].Replaces = []string{"sync.groups"}

	if versionOf(t, renamed) == versionOf(t, base) {
		t.Fatal("expected an operation rename to change the version")
	}

	replacing := surfaceDefinition("version-def")
	replacing.Operations[1].Replaces = []string{"sync.teams"}

	if versionOf(t, replacing) == versionOf(t, base) {
		t.Fatal("expected declaring an operation replacement to change the version")
	}

	surface := DefinitionSurface(replacing)

	if got := surface.Operations[0].Replaces; !slices.Equal(got, []string{"sync.teams"}) {
		t.Fatalf("sync.groups Replaces = %v", got)
	}
}

func versionOf(t *testing.T, def integrationtypes.Definition) string {
	t.Helper()

	reg := New()
	if err := reg.Register(def); err != nil {
		t.Fatalf("register: %v", err)
	}

	version := reg.Version(def.ID)
	if version == "" {
		t.Fatal("expected a version")
	}

	return version
}

func TestVersionIsStableAndChangesWithTheDefinition(t *testing.T) {
	t.Parallel()

	base, _ := minimalDefinition("version-def")
	same, _ := minimalDefinition("version-def")

	if versionOf(t, base) != versionOf(t, same) {
		t.Fatal("expected identical definitions to share a version")
	}

	type testCredential struct {
		Token string `json:"token" jsonschema:"required"`
	}

	requiredField, _ := minimalDefinition("version-def")
	requiredField.CredentialRegistrations[0].Schema = jsonx.SchemaFrom[testCredential]()

	addedSlot, _ := minimalDefinition("version-def")
	addedSlot.CredentialRegistrations = append(addedSlot.CredentialRegistrations, integrationtypes.CredentialRegistration{
		Ref:    versionSecondCredentialRef.ID(),
		Schema: versionSecondCredentialRef.Schema(),
	})

	connectionAdded, _ := minimalDefinition("version-def")
	connectionAdded.Connections = []integrationtypes.ConnectionRegistration{
		{CredentialRef: testCredentialRef.ID()},
	}

	for name, def := range map[string]integrationtypes.Definition{
		"same slot with a required field": requiredField,
		"added slot":                      addedSlot,
		"connection added":                connectionAdded,
	} {
		if versionOf(t, def) == versionOf(t, base) {
			t.Fatalf("%s: expected the version to change", name)
		}
	}

	description, _ := minimalDefinition("version-def")
	description.Description = "changed"

	if versionOf(t, description) != versionOf(t, base) {
		t.Fatal("description: expected the version to be unchanged")
	}

	if New().Version("missing") != "" {
		t.Fatal("expected an unregistered definition to have no version")
	}
}

func TestVersionChangesWhenASlotDeclaresAReplacement(t *testing.T) {
	t.Parallel()

	base, _ := minimalDefinition("version-def")

	slot := testCredentialRef.Replacing(versionSecondCredentialRef, nil)

	replacing, _ := minimalDefinition("version-def")
	replacing.CredentialRegistrations[0].Replaces = slot.Replaces()
	replacing.CredentialRegistrations[0].Convert = slot.Convert

	if versionOf(t, replacing) == versionOf(t, base) {
		t.Fatal("expected declaring a replacement to change the version")
	}

	surface := DefinitionSurface(replacing)

	if got, want := surface.Credentials[0].Replaces, []string{versionSecondCredentialRef.String()}; !slices.Equal(got, want) {
		t.Fatalf("Replaces = %v, want %v", got, want)
	}
}
