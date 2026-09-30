package registry

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
)

// sectionConfig is the switchable operation config type located as a user input section in the finalize tests
type sectionConfig struct {
	// Disable switches the operation off for the installation
	Disable bool `json:"disable,omitempty"`
	// Limit bounds the number of records the operation reads
	Limit int `json:"limit,omitempty"`
}

// sectionUserInput is the user input layout carrying sectionConfig under the section key
type sectionUserInput struct {
	// Primary is a top-level knob outside any section
	Primary string `json:"primary,omitempty"`
	// Section is the operation config section
	Section sectionConfig `json:"section,omitempty"`
}

// doubledSectionUserInput is a user input layout referencing sectionConfig from two properties
type doubledSectionUserInput struct {
	// First references the section type
	First sectionConfig `json:"first,omitempty"`
	// Second references the same section type
	Second sectionConfig `json:"second,omitempty"`
}

// sectionDefinitionRef is the definition identity behind the finalize tests
var sectionDefinitionRef = integrationtypes.NewDefinitionRef("def_finalize")

// sectionOperation returns a reconciled operation registration over sectionConfig with the given name
func sectionOperation(name string) integrationtypes.OperationRegistration {
	return integrationtypes.NewOperationRef[sectionConfig](name).
		HandlesRequest(func(context.Context, integrationtypes.OperationRequest, sectionConfig) (json.RawMessage, error) {
			return nil, nil
		}).
		Registration(sectionDefinitionRef, integrationtypes.OperationRegistration{Policy: integrationtypes.ExecutionPolicy{Reconcile: true}})
}

// sectionDefinition returns a definition whose user input is the given layout and whose operations are given
func sectionDefinition(userInput *integrationtypes.UserInputRegistration, operations ...integrationtypes.OperationRegistration) integrationtypes.Definition {
	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: sectionDefinitionRef.ID()},
		UserInput:      userInput,
		Operations:     operations,
	}
}

// TestFinalizeCredentialFormSchema verifies form slots receive their stored schema as the form schema, auth-managed slots keep none, and hand-built literals gain a stored schema
func TestFinalizeCredentialFormSchema(t *testing.T) {
	t.Parallel()

	literalSchema := json.RawMessage(`{"type":"object"}`)
	literalSlot := integrationtypes.NewCredentialSlotID("literal")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_finalize_credentials"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRef.Registration(integrationtypes.CredentialRegistration{}),
			testAuthCredentialRef.Registration(integrationtypes.CredentialRegistration{}),
			{Ref: literalSlot, Schema: literalSchema},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			integrationtypes.NewConnectionRef(testAuthCredentialRef).Registration(integrationtypes.ConnectionRegistration{
				Auth: &integrationtypes.AuthRegistration{CredentialRef: testAuthCredentialRef.ID()},
			}),
		},
	}

	finalized, err := finalizeDefinition(def)
	if err != nil {
		t.Fatalf("finalizeDefinition() error = %v", err)
	}

	form, _ := finalized.CredentialRegistration(testCredentialRef.ID())
	if !sameJSON(form.Schema, testCredentialRef.Schema()) {
		t.Fatalf("form slot Schema = %s, want its stored schema", form.Schema)
	}

	authManaged, _ := finalized.CredentialRegistration(testAuthCredentialRef.ID())
	if len(authManaged.Schema) != 0 {
		t.Fatalf("auth-managed slot Schema = %s, want empty", authManaged.Schema)
	}

	if !sameJSON(authManaged.StoredSchema, testAuthCredentialRef.Schema()) {
		t.Fatalf("auth-managed slot StoredSchema = %s, want the reflected schema", authManaged.StoredSchema)
	}

	literal, _ := finalized.CredentialRegistration(literalSlot)
	if !sameJSON(literal.StoredSchema, literalSchema) || !sameJSON(literal.Schema, literalSchema) {
		t.Fatalf("literal slot = %+v, want stored and form schema from the literal", literal)
	}

	if len(def.CredentialRegistrations[0].Schema) != 0 {
		t.Fatal("expected finalize to leave the builder's registrations untouched")
	}
}

// TestFinalizeDerivesSectionResolverAndSwitch verifies an operation whose config type is a user input section resolves that section and is switched off by its disable flag
func TestFinalizeDerivesSectionResolverAndSwitch(t *testing.T) {
	t.Parallel()

	def := sectionDefinition(integrationtypes.NewUserInputRef[sectionUserInput]("sectionUserInput").Registration(), sectionOperation("sync"))

	finalized, err := finalizeDefinition(def)
	if err != nil {
		t.Fatalf("finalizeDefinition() error = %v", err)
	}

	operation := finalized.Operations[0]
	if operation.ConfigResolver == nil || operation.Disabled == nil {
		t.Fatal("expected the section resolver and switch to be derived")
	}

	if def.Operations[0].ConfigResolver != nil {
		t.Fatal("expected finalize to leave the builder's operations untouched")
	}

	disabled := json.RawMessage(`{"primary":"p","section":{"disable":true,"limit":3}}`)

	if got := operation.ConfigResolver(disabled); !sameJSON(got, json.RawMessage(`{"disable":true,"limit":3}`)) {
		t.Fatalf("ConfigResolver() = %s, want the section", got)
	}

	if !operation.DisabledFor(disabled) {
		t.Fatal("expected the section's disable flag to switch the operation off")
	}

	enabled := json.RawMessage(`{"section":{"limit":1}}`)
	if operation.DisabledFor(enabled) {
		t.Fatal("expected a section without the disable flag to leave the operation on")
	}

	absent := json.RawMessage(`{"primary":"p"}`)
	if got := operation.ConfigResolver(absent); got != nil {
		t.Fatalf("ConfigResolver() = %s, want nil for an absent section", got)
	}

	if operation.DisabledFor(absent) {
		t.Fatal("expected an absent section to leave the operation on")
	}
}

// TestFinalizeKeepsAuthoredResolverAndSwitch verifies authored resolvers and switches are not replaced by derived ones
func TestFinalizeKeepsAuthoredResolverAndSwitch(t *testing.T) {
	t.Parallel()

	operation := sectionOperation("sync")
	operation.ConfigResolver = func(json.RawMessage) json.RawMessage { return json.RawMessage(`{"limit":9}`) }
	operation.Disabled = func(json.RawMessage) bool { return true }

	finalized, err := finalizeDefinition(sectionDefinition(integrationtypes.NewUserInputRef[sectionUserInput]("sectionUserInput").Registration(), operation))
	if err != nil {
		t.Fatalf("finalizeDefinition() error = %v", err)
	}

	if got := finalized.Operations[0].ConfigResolver(json.RawMessage(`{"section":{"limit":1}}`)); !sameJSON(got, json.RawMessage(`{"limit":9}`)) {
		t.Fatalf("ConfigResolver() = %s, want the authored resolver", got)
	}

	if !finalized.Operations[0].DisabledFor(json.RawMessage(`{}`)) {
		t.Fatal("expected the authored switch to be kept")
	}
}

// TestFinalizeSectionErrors verifies ambiguous, mismatched, and missing sections are rejected
func TestFinalizeSectionErrors(t *testing.T) {
	t.Parallel()

	mismatched := &integrationtypes.UserInputRegistration{Schema: json.RawMessage(`{
		"$ref": "#/$defs/input",
		"$defs": {
			"input": {"type": "object", "properties": {"section": {"$ref": "#/$defs/sectionConfig"}}},
			"sectionConfig": {"type": "object", "properties": {"other": {"type": "string"}}}
		}
	}`)}

	flat := &integrationtypes.UserInputRegistration{Schema: json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer"}}}`)}

	cases := []struct {
		name    string
		def     integrationtypes.Definition
		wantErr error
	}{
		{
			name:    "two properties reference one config type",
			def:     sectionDefinition(integrationtypes.NewUserInputRef[doubledSectionUserInput]("doubledSectionUserInput").Registration(), sectionOperation("sync")),
			wantErr: ErrConfigSectionAmbiguous,
		},
		{
			name:    "two operations resolve to one section",
			def:     sectionDefinition(integrationtypes.NewUserInputRef[sectionUserInput]("sectionUserInput").Registration(), sectionOperation("sync.a"), sectionOperation("sync.b")),
			wantErr: ErrConfigSectionAmbiguous,
		},
		{
			name:    "section declares a different schema for the config type",
			def:     sectionDefinition(mismatched, sectionOperation("sync")),
			wantErr: ErrConfigSectionMismatch,
		},
		{
			name:    "reconciled configurable operation without a section",
			def:     sectionDefinition(flat, sectionOperation("sync")),
			wantErr: ErrConfigSectionRequired,
		},
		{
			name:    "reconciled configurable operation without user input",
			def:     sectionDefinition(nil, sectionOperation("sync")),
			wantErr: ErrConfigSectionRequired,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := finalizeDefinition(tc.def); !errors.Is(err, tc.wantErr) {
				t.Fatalf("finalizeDefinition() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// TestFinalizeSectionRequiredExemptions verifies inline operations and operations with an authored resolver need no section
func TestFinalizeSectionRequiredExemptions(t *testing.T) {
	t.Parallel()

	inline := sectionOperation("inline")
	inline.Policy = integrationtypes.ExecutionPolicy{Inline: true}

	authored := sectionOperation("authored")
	authored.ConfigResolver = func(userInput json.RawMessage) json.RawMessage { return userInput }

	if _, err := finalizeDefinition(sectionDefinition(nil, inline, authored)); err != nil {
		t.Fatalf("finalizeDefinition() error = %v", err)
	}
}
