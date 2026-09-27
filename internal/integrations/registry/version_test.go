package registry

import (
	"slices"
	"testing"

	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
)

// versionSecondCredential is the credential type behind the extra slot added in the version test
type versionSecondCredential struct{}

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
	requiredField.CredentialRegistrations[0].Ref = integrationtypes.NewCredentialRef[testCredential]()

	addedSlot, _ := minimalDefinition("version-def")
	addedSlot.CredentialRegistrations = append(addedSlot.CredentialRegistrations, integrationtypes.CredentialRegistration{
		Ref: integrationtypes.NewCredentialRef[versionSecondCredential](),
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

	retired := integrationtypes.NewCredentialRef[versionSecondCredential]()

	replacing, _ := minimalDefinition("version-def")
	replacing.CredentialRegistrations[0].Ref = integrationtypes.Replacing(testCredentialRef, retired, nil)

	if versionOf(t, replacing) == versionOf(t, base) {
		t.Fatal("expected declaring a replacement to change the version")
	}

	surface := DefinitionSurface(replacing)

	if got, want := surface.Credentials[0].Replaces, []string{retired.ID().String()}; !slices.Equal(got, want) {
		t.Fatalf("Replaces = %v, want %v", got, want)
	}
}
