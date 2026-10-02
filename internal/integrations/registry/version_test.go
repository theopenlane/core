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
	retiredUserInputRef = integrationtypes.UserInputRefOf[retiredUserInput]()
	// surfaceUserInputRef is the current layout upgrading documents stored under the retired one
	surfaceUserInputRef = integrationtypes.UserInputRefOf[testUserInput]().
				Upgraded(func(_ context.Context, _ integrationtypes.InstallationRequest, from string, stored json.RawMessage) (testUserInput, error) {
			switch from {
			case retiredUserInputRef.Name():
				old, err := jsonx.Decode[retiredUserInput](stored)
				if err != nil {
					return testUserInput{}, err
				}

				return testUserInput{Region: old.Zone}, nil
			default:
				return jsonx.Decode[testUserInput](stored)
			}
		})
)

// surfaceDefinition returns a definition exercising every surfaced kind in reverse name order
func surfaceDefinition(id string) integrationtypes.Definition {
	def, clientRef := minimalDefinition(id)
	defRef := integrationtypes.NewDefinitionRef(id)

	def.UserInput = surfaceUserInputRef.Registration()

	def.Connections = []integrationtypes.ConnectionRegistration{
		{CredentialRef: testCredentialRef.ID()},
	}

	def.HealthCheck = &integrationtypes.HealthCheckRegistration{ClientRef: clientRef.ID(), Handle: newTestHandler()}

	def.Installation = integrationtypes.NewInstallationRef(func(context.Context, integrationtypes.InstallationRequest) (versionMetadata, bool, error) {
		return versionMetadata{}, true, nil
	}).Registration()

	def.Operations = []integrationtypes.OperationRegistration{
		{
			Name:         "sync.users",
			Replaces:     []string{"sync.people"},
			Topic:        defRef.OperationTopic("sync.users"),
			ClientRef:    clientRef.ID(),
			ConfigSchema: jsonx.SchemaFrom[testOperationConfig](),
			Input:        &integrationtypes.InputRegistration{Name: "sync.users"},
			Handle:       newTestHandler(),
		},
		{
			Name:         "sync.groups",
			Topic:        defRef.OperationTopic("sync.groups"),
			ClientRef:    clientRef.ID(),
			ConfigSchema: jsonx.SchemaFrom[testOperationConfig](),
			Input: &integrationtypes.InputRegistration{
				Name: "sync.groups",
				Upgrade: func(_ context.Context, _ integrationtypes.InstallationRequest, _ string, stored json.RawMessage) (json.RawMessage, error) {
					return stored, nil
				},
			},
			Handle: newTestHandler(),
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

	if surface.UserInput == nil || !surface.UserInput.Upgrade {
		t.Fatalf("UserInput = %+v, want a schema declaring an upgrade", surface.UserInput)
	}

	if len(surface.UserInput.Schema) == 0 {
		t.Fatal("expected the user input schema to be surfaced")
	}

	if surface.Installation == nil || len(surface.Installation.Schema) == 0 {
		t.Fatal("expected the installation metadata schema to be surfaced")
	}

	if got := lo.Map(surface.Operations, func(o SurfaceOperation, _ int) string { return o.Name }); !slices.Equal(got, []string{"sync.groups", "sync.users"}) {
		t.Fatalf("Operations = %v, want sorted names", got)
	}

	if got := surface.Operations[1].Replaces; !slices.Equal(got, []string{"sync.people"}) {
		t.Fatalf("sync.users Replaces = %v", got)
	}

	if len(surface.Operations[0].Schema) == 0 || len(surface.Operations[1].Schema) == 0 {
		t.Fatal("expected every operation's stored input schema to be surfaced")
	}

	if !surface.Operations[0].Upgrade || surface.Operations[1].Upgrade {
		t.Fatalf("expected only sync.groups to surface an upgrade, got %+v", surface.Operations)
	}

	root, _, err := jsonx.SchemaRoot(surface.Operations[1].Schema)
	if err != nil {
		t.Fatalf("SchemaRoot() error = %v", err)
	}

	if _, ok := root.Properties.Get("disable"); !ok {
		t.Fatal("expected the surfaced operation schema to carry the uniform settings")
	}

	if got := lo.Map(surface.Webhooks, func(w SurfaceWebhook, _ int) string { return w.Name }); !slices.Equal(got, []string{"events.v2", "static"}) {
		t.Fatalf("Webhooks = %v, want sorted names", got)
	}

	if got := surface.Webhooks[0].Replaces; !slices.Equal(got, []string{"events.v1"}) {
		t.Fatalf("events.v2 Replaces = %v", got)
	}

	if got := surface.Webhooks[0].Events; !slices.Equal(got, []string{"member.joined", "member.removed"}) {
		t.Fatalf("events.v2 Events = %v, want sorted names", got)
	}

	if len(surface.Webhooks[1].Events) != 0 || surface.Webhooks[1].Replaces != nil {
		t.Fatalf("static webhook = %+v, want no events and no replacements", surface.Webhooks[1])
	}

	unfinalized, _ := minimalDefinition("minimal-def")

	minimal, err := finalizeDefinition(unfinalized)
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}

	encoded, err := json.Marshal(DefinitionSurface(minimal))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var keys map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &keys); err != nil {
		t.Fatalf("decode: %v", err)
	}

	for _, absent := range []string{"userInput", "installation", "webhooks"} {
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
	requiredField.CredentialRegistrations[0].StoredSchema = jsonx.SchemaFrom[testCredential]()

	addedSlot, _ := minimalDefinition("version-def")
	addedSlot.CredentialRegistrations = append(addedSlot.CredentialRegistrations, versionSecondCredentialRef.Registration(integrationtypes.CredentialRegistration{}))

	connectionAdded, _ := minimalDefinition("version-def")
	connectionAdded.Connections = []integrationtypes.ConnectionRegistration{
		{CredentialRef: testCredentialRef.ID()},
	}
	connectionAdded.HealthCheck = newTestHealthCheck()

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

	slot := testCredentialRef.Replacing(versionSecondCredentialRef)

	replacing, _ := minimalDefinition("version-def")
	replacing.CredentialRegistrations[0].Replaces = slot.Replaces()

	if versionOf(t, replacing) == versionOf(t, base) {
		t.Fatal("expected declaring a replacement to change the version")
	}

	surface := DefinitionSurface(replacing)

	if got, want := surface.Credentials[0].Replaces, []string{versionSecondCredentialRef.String()}; !slices.Equal(got, want) {
		t.Fatalf("Replaces = %v, want %v", got, want)
	}
}

// TestVersionChangesWhenAReplacementIsRemoved verifies removing a replacement moves the version
func TestVersionChangesWhenAReplacementIsRemoved(t *testing.T) {
	t.Parallel()

	base := surfaceDefinition("version-def")

	userInput := surfaceDefinition("version-def")
	userInput.UserInput.Upgrade = nil

	operation := surfaceDefinition("version-def")
	operation.Operations[0].Replaces = nil

	webhook := surfaceDefinition("version-def")
	webhook.Webhooks[1].Replaces = nil

	for name, def := range map[string]integrationtypes.Definition{
		"user input upgrade removed":    userInput,
		"operation replacement removed": operation,
		"webhook replacement removed":   webhook,
	} {
		if versionOf(t, def) == versionOf(t, base) {
			t.Fatalf("%s: expected the version to change", name)
		}
	}
}

// TestVersionChangesWithSurfacedNameAndSchemaFields verifies surfaced fields move the version
func TestVersionChangesWithSurfacedNameAndSchemaFields(t *testing.T) {
	t.Parallel()

	base := surfaceDefinition("version-def")

	webhookName := surfaceDefinition("version-def")
	webhookName.Webhooks[0].Name = "static-v2"

	webhookReplaces := surfaceDefinition("version-def")
	webhookReplaces.Webhooks[0].Replaces = []string{"static-old"}

	webhookEventName := surfaceDefinition("version-def")
	webhookEventName.Webhooks[1].Events[0].Name = "member.left"

	userInputSchema := surfaceDefinition("version-def")
	userInputSchema.UserInput.Schema = json.RawMessage(`{"type":"object","required":["different"]}`)

	installationSchema := surfaceDefinition("version-def")
	installationSchema.Installation.Schema = json.RawMessage(`{"type":"object","properties":{"zone":{"type":"string"}}}`)

	operationConfigSchema := surfaceDefinition("version-def")
	operationConfigSchema.Operations[0].ConfigSchema = json.RawMessage(`{"type":"object","required":["limit"]}`)

	for name, def := range map[string]integrationtypes.Definition{
		"webhook name":              webhookName,
		"webhook replaces declared": webhookReplaces,
		"webhook event name":        webhookEventName,
		"user input schema alone":   userInputSchema,
		"installation schema":       installationSchema,
		"operation config schema":   operationConfigSchema,
	} {
		if versionOf(t, def) == versionOf(t, base) {
			t.Fatalf("%s: expected the version to change", name)
		}
	}
}

// TestVersionUnchangedForDescriptionMetaAndHandlers verifies non-surfaced fields don't change it
func TestVersionUnchangedForDescriptionMetaAndHandlers(t *testing.T) {
	t.Parallel()

	base := surfaceDefinition("version-def")

	changed := surfaceDefinition("version-def")
	changed.Description = "a new description"
	changed.Connections[0].Meta = map[string]integrationtypes.MetaInfo{"note": {Value: "hello"}}
	changed.Operations[0].Handle = newTestHandler()
	changed.Webhooks[1].Events[0].Handle = func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil }

	if versionOf(t, changed) != versionOf(t, base) {
		t.Fatal("expected description, meta, and handler changes to leave the version unchanged")
	}
}

// TestVersionChangesWhenAuthManagedCredentialSchemaChanges verifies its schema moves the version
func TestVersionChangesWhenAuthManagedCredentialSchemaChanges(t *testing.T) {
	t.Parallel()

	build := func(storedSchema json.RawMessage) integrationtypes.Definition {
		def, _ := minimalDefinition("version-def")
		def.CredentialRegistrations = append(def.CredentialRegistrations, integrationtypes.CredentialRegistration{Ref: testAuthCredentialRef.ID(), StoredSchema: storedSchema})
		def.Connections = []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  testAuthCredentialRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testAuthCredentialRef.ID()},
				Auth:           &integrationtypes.AuthRegistration{CredentialRef: testAuthCredentialRef.ID()},
			},
		}
		def.HealthCheck = newTestHealthCheck()

		return def
	}

	base := build(testAuthCredentialRef.Schema())
	changed := build(json.RawMessage(`{"type":"object","required":["token"]}`))

	if versionOf(t, changed) == versionOf(t, base) {
		t.Fatal("expected an auth-managed credential's Auth schema to change the version")
	}
}
