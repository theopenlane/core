package operations

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/openapi"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestFindMapping(t *testing.T) {
	t.Parallel()

	mappings := []types.MappingRegistration{
		{Schema: "asset", Variant: "", Spec: types.MappingOverride{MapExpr: "asset_expr"}},
		{Schema: "contact", Variant: "primary", Spec: types.MappingOverride{MapExpr: "contact_primary_expr"}},
		{Schema: "contact", Variant: "secondary", Spec: types.MappingOverride{MapExpr: "contact_secondary_expr"}},
	}

	tests := []struct {
		name      string
		schema    string
		variant   string
		wantFound bool
		wantExpr  string
	}{
		{
			name:      "exact match no variant",
			schema:    "asset",
			variant:   "",
			wantFound: true,
			wantExpr:  "asset_expr",
		},
		{
			name:      "exact match with variant",
			schema:    "contact",
			variant:   "primary",
			wantFound: true,
			wantExpr:  "contact_primary_expr",
		},
		{
			name:      "second variant match",
			schema:    "contact",
			variant:   "secondary",
			wantFound: true,
			wantExpr:  "contact_secondary_expr",
		},
		{
			name:      "unknown schema",
			schema:    "unknown",
			variant:   "",
			wantFound: false,
		},
		{
			name:      "wrong variant",
			schema:    "contact",
			variant:   "tertiary",
			wantFound: false,
		},
		{
			name:      "empty mappings",
			schema:    "asset",
			variant:   "",
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := mappings
			if tc.name == "empty mappings" {
				input = nil
			}

			got, found := findMapping(input, tc.schema, tc.variant)
			assert.Equal(t, found, tc.wantFound)

			if found {
				assert.Equal(t, got.MapExpr, tc.wantExpr)
			}
		})
	}
}

func TestContractIncludesSchema(t *testing.T) {
	t.Parallel()

	contracts := []types.IngestContract{
		{Schema: "asset"},
		{Schema: "contact"},
		{Schema: "finding"},
	}

	tests := []struct {
		name   string
		schema string
		want   bool
	}{
		{"present schema", "asset", true},
		{"another present schema", "finding", true},
		{"absent schema", "risk", false},
		{"empty string", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := contractIncludesSchema(contracts, tc.schema)
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestContractIncludesSchema_EmptyContracts(t *testing.T) {
	t.Parallel()

	assert.Equal(t, contractIncludesSchema(nil, "asset"), false)
}

func TestResolveInstallationFilterExpr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		operations    map[string]json.RawMessage
		operationName string
		wantExpr      string
		wantErr       bool
	}{
		{
			name:          "no stored operation config returns empty",
			operationName: "directory-sync",
			wantExpr:      "",
		},
		{
			name:          "empty stored document returns empty",
			operations:    map[string]json.RawMessage{"directory-sync": json.RawMessage(`{}`)},
			operationName: "directory-sync",
			wantExpr:      "",
		},
		{
			name:          "filterExpr read from the operation's stored document",
			operations:    map[string]json.RawMessage{"directory-sync": json.RawMessage(`{"filterExpr":"payload.is_external == false"}`)},
			operationName: "directory-sync",
			wantExpr:      "payload.is_external == false",
		},
		{
			name:          "invalid stored document",
			operations:    map[string]json.RawMessage{"directory-sync": json.RawMessage(`{not json`)},
			operationName: "directory-sync",
			wantErr:       true,
		},
		{
			name:          "another operation's document is not consulted",
			operations:    map[string]json.RawMessage{"directory-sync": json.RawMessage(`{"filterExpr":"payload.type == \"user\""}`)},
			operationName: "asset-sync",
			wantExpr:      "",
		},
		{
			name:          "unknown operationName returns empty",
			operations:    map[string]json.RawMessage{"other-op": json.RawMessage(`{"filterExpr":"resource == \"groups\""}`)},
			operationName: "unknown-op",
			wantExpr:      "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			installation := &ent.Integration{
				OperationConfig: openapi.IntegrationOperationConfig{Operations: tc.operations},
			}

			expr, err := resolveInstallationFilterExpr(installation, tc.operationName)
			if tc.wantErr {
				assert.Assert(t, err != nil, "expected error")
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, expr, tc.wantExpr)
		})
	}
}

func TestEnvelopeIncludedByFilters(t *testing.T) {
	t.Parallel()

	envelope := types.MappingEnvelope{
		Variant:  "user",
		Resource: "users",
		Action:   "create",
		Payload:  json.RawMessage(`{"name":"alice"}`),
	}

	tests := []struct {
		name                   string
		installationFilterExpr string
		mappingFilterExpr      string
		wantMatch              bool
		wantErr                bool
	}{
		{
			name:      "both empty passes",
			wantMatch: true,
		},
		{
			name:                   "installation filter matches",
			installationFilterExpr: `resource == "users"`,
			wantMatch:              true,
		},
		{
			name:                   "installation filter rejects",
			installationFilterExpr: `resource == "groups"`,
			wantMatch:              false,
		},
		{
			name:              "mapping filter matches",
			mappingFilterExpr: `action == "create"`,
			wantMatch:         true,
		},
		{
			name:              "mapping filter rejects",
			mappingFilterExpr: `action == "delete"`,
			wantMatch:         false,
		},
		{
			name:                   "both filters match",
			installationFilterExpr: `resource == "users"`,
			mappingFilterExpr:      `action == "create"`,
			wantMatch:              true,
		},
		{
			name:                   "installation matches mapping rejects",
			installationFilterExpr: `resource == "users"`,
			mappingFilterExpr:      `action == "delete"`,
			wantMatch:              false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			matched, err := envelopeIncludedByFilters(context.Background(), tc.installationFilterExpr, tc.mappingFilterExpr, envelope, types.MappingInstallation{})
			if tc.wantErr {
				assert.Assert(t, err != nil, "expected error")
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, matched, tc.wantMatch)
		})
	}
}

func TestMapIngestRecord(t *testing.T) {
	t.Parallel()

	definition := types.Definition{
		Mappings: []types.MappingRegistration{
			{
				Schema:  "asset",
				Variant: "",
				Spec: types.MappingOverride{
					MapExpr: `{"name": payload.name}`,
				},
			},
			{
				Schema:  "asset",
				Variant: "filtered",
				Spec: types.MappingOverride{
					FilterExpr: `resource == "excluded"`,
					MapExpr:    `{"name": payload.name}`,
				},
			},
		},
	}

	tests := []struct {
		name        string
		schema      string
		envelope    types.MappingEnvelope
		wantInclude bool
		wantErr     error
	}{
		{
			name:   "successful mapping",
			schema: "asset",
			envelope: types.MappingEnvelope{
				Variant: "",
				Payload: json.RawMessage(`{"name":"server-01"}`),
			},
			wantInclude: true,
		},
		{
			name:   "filtered out by mapping filter",
			schema: "asset",
			envelope: types.MappingEnvelope{
				Variant:  "filtered",
				Resource: "included",
				Payload:  json.RawMessage(`{"name":"server-01"}`),
			},
			wantInclude: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mapping, found := findMapping(definition.Mappings, tc.schema, tc.envelope.Variant)
			assert.Assert(t, found)

			record, include, err := mapIngestRecord(context.Background(), mapping, tc.schema, tc.envelope, "", types.MappingInstallation{})
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, include, tc.wantInclude)
			if include {
				assert.Equal(t, record.Schema, tc.schema)
			}
		})
	}
}

func TestWrapIngestPersistError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantNil bool
		wantErr error
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:    "validation error wraps as mapped document invalid",
			err:     &ent.ValidationError{Name: "field"},
			wantErr: ErrIngestMappedDocumentInvalid,
		},
		{
			name:    "not singular error wraps as upsert conflict",
			err:     &ent.NotSingularError{},
			wantErr: ErrIngestUpsertConflict,
		},
		{
			name:    "constraint error wraps as upsert conflict",
			err:     &ent.ConstraintError{},
			wantErr: ErrIngestUpsertConflict,
		},
		{
			name:    "generic error wraps as persist failed",
			err:     errors.New("database timeout"),
			wantErr: ErrIngestPersistFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := wrapIngestPersistError(tc.err)
			if tc.wantNil {
				assert.Assert(t, got == nil)
				return
			}
			assert.ErrorIs(t, got, tc.wantErr)
		})
	}
}

// testDefinition builds a minimal definition with the given mappings registered in a registry
func testDefinition(t *testing.T, mappings []types.MappingRegistration) (*registry.Registry, types.Definition) {
	t.Helper()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:          "test-def",
			DisplayName: "Test",
			Active:      true,
		},
		Operations: []types.OperationRegistration{
			{
				Name:  "sync",
				Topic: "test.sync",
				Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
					return nil, nil
				},
			},
		},
		Mappings: mappings,
	}

	reg := registry.New()
	if err := reg.Register(def); err != nil {
		t.Fatalf("failed to register test definition: %v", err)
	}

	return reg, def
}

// testInstallationMetadata carries the instance id every ingest installation must have
var testInstallationMetadata = openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-test"}}

// applySingleEnvelopes runs each envelope through applyPayloadSets as a one-record batch
func applySingleEnvelopes(t *testing.T, ic IngestContext, operationName, schema string, envelopes ...types.MappingEnvelope) int {
	t.Helper()

	contracts := []types.IngestContract{{Schema: schema}}
	handled := 0

	for _, envelope := range envelopes {
		payloadSets := []types.IngestPayloadSet{{Schema: schema, Envelopes: []types.MappingEnvelope{envelope}}}

		_, err := applyPayloadSets(context.Background(), ic, ingestBatch{OperationName: operationName, Contracts: contracts, PayloadSets: payloadSets}, func(context.Context, mappedIngestRecord) (ingestOutcome, error) {
			handled++
			return ingestOutcome{}, nil
		})
		assert.NilError(t, err)
	}

	return handled
}

func TestProcessPayloadSets_DefinitionNotFound(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "nonexistent"},
	}

	_, err := applyPayloadSets(context.Background(), ic, ingestBatch{}, func(context.Context, mappedIngestRecord) (ingestOutcome, error) {
		return ingestOutcome{}, nil
	})

	assert.ErrorIs(t, err, ErrIngestDefinitionNotFound)
}

func TestProcessPayloadSets_InstanceIDRequired(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def"},
	}

	_, err := applyPayloadSets(context.Background(), ic, ingestBatch{}, func(context.Context, mappedIngestRecord) (ingestOutcome, error) {
		return ingestOutcome{}, nil
	})

	assert.ErrorIs(t, err, ErrIngestInstanceIDRequired)
}

func TestProcessPayloadSets_SchemaNotDeclared(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, []types.MappingRegistration{
		{Schema: entityops.SchemaAsset.Name, Variant: "", Spec: types.MappingOverride{MapExpr: `payload`}},
	})

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def", InstallationMetadata: testInstallationMetadata},
	}

	contracts := []types.IngestContract{{Schema: entityops.SchemaContact.Name}}
	payloadSets := []types.IngestPayloadSet{
		{Schema: entityops.SchemaAsset.Name, Envelopes: []types.MappingEnvelope{{Payload: json.RawMessage(`{}`)}}},
	}

	_, err := applyPayloadSets(context.Background(), ic, ingestBatch{Contracts: contracts, PayloadSets: payloadSets}, func(context.Context, mappedIngestRecord) (ingestOutcome, error) {
		return ingestOutcome{}, nil
	})

	assert.ErrorIs(t, err, ErrIngestSchemaNotDeclared)
}

func TestProcessPayloadSets_SchemaNotFound(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def", InstallationMetadata: testInstallationMetadata},
	}

	contracts := []types.IngestContract{{Schema: "totally_bogus_schema"}}
	payloadSets := []types.IngestPayloadSet{
		{Schema: "totally_bogus_schema", Envelopes: []types.MappingEnvelope{{Payload: json.RawMessage(`{}`)}}},
	}

	_, err := applyPayloadSets(context.Background(), ic, ingestBatch{Contracts: contracts, PayloadSets: payloadSets}, func(context.Context, mappedIngestRecord) (ingestOutcome, error) {
		return ingestOutcome{}, nil
	})

	assert.ErrorIs(t, err, ErrIngestSchemaNotFound)
}

func TestProcessPayloadSets_MappingNotFound(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def", InstallationMetadata: testInstallationMetadata},
	}

	contracts := []types.IngestContract{{Schema: entityops.SchemaAsset.Name}}
	payloadSets := []types.IngestPayloadSet{
		{
			Schema: entityops.SchemaAsset.Name,
			Envelopes: []types.MappingEnvelope{
				{Variant: "", Payload: json.RawMessage(`{"name":"test"}`)},
			},
		},
	}

	_, err := applyPayloadSets(context.Background(), ic, ingestBatch{Contracts: contracts, PayloadSets: payloadSets}, func(context.Context, mappedIngestRecord) (ingestOutcome, error) {
		return ingestOutcome{}, nil
	})

	assert.NilError(t, err, "unmappable records are skipped, not fatal")
}

func TestProcessPayloadSets_InvalidInstallationFilterConfig(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry: reg,
		Integration: &ent.Integration{
			DefinitionID:         "test-def",
			InstallationMetadata: testInstallationMetadata,
			OperationConfig:      openapi.IntegrationOperationConfig{Operations: map[string]json.RawMessage{"": json.RawMessage(`{invalid`)}},
		},
	}

	_, err := applyPayloadSets(context.Background(), ic, ingestBatch{}, func(context.Context, mappedIngestRecord) (ingestOutcome, error) {
		return ingestOutcome{}, nil
	})

	assert.ErrorIs(t, err, ErrIngestInstallationFilterConfigInvalid)
}

func TestProcessPayloadSets_SuccessfulMapping(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, []types.MappingRegistration{
		{
			Schema:  entityops.SchemaAsset.Name,
			Variant: "",
			Spec:    types.MappingOverride{MapExpr: `{"sourceIdentifier": payload.id}`},
		},
	})

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def", InstallationMetadata: testInstallationMetadata},
	}

	contracts := []types.IngestContract{{Schema: entityops.SchemaAsset.Name}}
	payloadSets := []types.IngestPayloadSet{
		{
			Schema: entityops.SchemaAsset.Name,
			Envelopes: []types.MappingEnvelope{
				{Variant: "", Payload: json.RawMessage(`{"id":"asset-001"}`)},
			},
		},
	}

	var handled []mappedIngestRecord

	_, err := applyPayloadSets(context.Background(), ic, ingestBatch{Contracts: contracts, PayloadSets: payloadSets}, func(_ context.Context, record mappedIngestRecord) (ingestOutcome, error) {
		handled = append(handled, record)
		return ingestOutcome{}, nil
	})

	assert.NilError(t, err)
	assert.Equal(t, len(handled), 1)
	assert.Equal(t, handled[0].Schema, entityops.SchemaAsset.Name)
}

func TestProcessPayloadSets_EmptyPayloadSets(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def", InstallationMetadata: testInstallationMetadata},
	}

	_, err := applyPayloadSets(context.Background(), ic, ingestBatch{}, func(context.Context, mappedIngestRecord) (ingestOutcome, error) {
		t.Fatal("handler should not be called for empty payload sets")
		return ingestOutcome{}, nil
	})

	assert.NilError(t, err)
}

func TestProcessPayloadSets_FilteredEnvelopes(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, []types.MappingRegistration{
		{
			Schema:  entityops.SchemaAsset.Name,
			Variant: "",
			Spec: types.MappingOverride{
				FilterExpr: `resource == "wanted"`,
				MapExpr:    `payload`,
			},
		},
	})

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def", InstallationMetadata: testInstallationMetadata},
	}

	handled := applySingleEnvelopes(t, ic, "", entityops.SchemaAsset.Name,
		types.MappingEnvelope{Resource: "unwanted", Payload: json.RawMessage(`{"name":"skip"}`)},
		types.MappingEnvelope{Resource: "wanted", Payload: json.RawMessage(`{"name":"keep"}`)},
	)

	assert.Equal(t, handled, 1)
}

// TestProcessPayloadSets_NestedInstallationFilter verifies nested filterExpr resolution
func TestProcessPayloadSets_NestedInstallationFilter(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:     "test-def",
			Active: true,
		},
		Operations: []types.OperationRegistration{
			{
				Name: "repo-sync",
				Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
					return nil, nil
				},
			},
		},
		Mappings: []types.MappingRegistration{
			{
				Schema:  entityops.SchemaAsset.Name,
				Variant: "",
				Spec:    types.MappingOverride{MapExpr: `payload`},
			},
		},
	}

	reg := registry.New()
	assert.NilError(t, reg.Register(def))

	ic := IngestContext{
		Registry: reg,
		Integration: &ent.Integration{
			DefinitionID:         "test-def",
			InstallationMetadata: testInstallationMetadata,
			OperationConfig: openapi.IntegrationOperationConfig{
				Operations: map[string]json.RawMessage{"repo-sync": json.RawMessage(`{"filterExpr":"payload.is_private == true"}`)},
			},
		},
	}

	handled := applySingleEnvelopes(t, ic, "repo-sync", entityops.SchemaAsset.Name,
		types.MappingEnvelope{Payload: json.RawMessage(`{"name":"private-repo","is_private":true}`)},
		types.MappingEnvelope{Payload: json.RawMessage(`{"name":"public-repo","is_private":false}`)},
		types.MappingEnvelope{Payload: json.RawMessage(`{"name":"another-private","is_private":true}`)},
	)

	assert.Equal(t, handled, 2)
}

// TestProcessPayloadSets_NestedFilterDoesNotLeakAcrossOperations verifies no cross-operation leak
func TestProcessPayloadSets_NestedFilterDoesNotLeakAcrossOperations(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:     "test-def",
			Active: true,
		},
		Operations: []types.OperationRegistration{
			{
				Name:  "finding-sync",
				Topic: types.NewDefinitionRef("test-def").OperationTopic("finding-sync"),
				Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
					return nil, nil
				},
			},
			{
				Name:  "asset-sync",
				Topic: types.NewDefinitionRef("test-def").OperationTopic("asset-sync"),
				Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
					return nil, nil
				},
			},
		},
		Mappings: []types.MappingRegistration{
			{
				Schema:  entityops.SchemaAsset.Name,
				Variant: "",
				Spec:    types.MappingOverride{MapExpr: `payload`},
			},
		},
	}

	reg := registry.New()
	assert.NilError(t, reg.Register(def))

	ic := IngestContext{
		Registry: reg,
		Integration: &ent.Integration{
			DefinitionID:         "test-def",
			InstallationMetadata: testInstallationMetadata,
			OperationConfig: openapi.IntegrationOperationConfig{
				Operations: map[string]json.RawMessage{
					"finding-sync": json.RawMessage(`{"filterExpr":"payload.severity == \"CRITICAL\""}`),
					"asset-sync":   json.RawMessage(`{}`),
				},
			},
		},
	}

	handled := applySingleEnvelopes(t, ic, "asset-sync", entityops.SchemaAsset.Name,
		types.MappingEnvelope{Payload: json.RawMessage(`{"name":"asset-001"}`)},
		types.MappingEnvelope{Payload: json.RawMessage(`{"name":"asset-002"}`)},
	)

	assert.Equal(t, handled, 2)
}

func TestStampProvenanceOverridesMappedValues(t *testing.T) {
	t.Parallel()

	installation := &ent.Integration{
		ID:                   "int_owner",
		OwnerID:              "org_1",
		DefinitionID:         "def_dir",
		InstallationMetadata: testInstallationMetadata,
	}

	payload := json.RawMessage(`{"external_id":"acct-1","managed_by":"int_other","source_instance_id":"tenant-other","owner_id":"org_other","source_definition_id":"def_other"}`)

	stamped := entityops.StampProvenance(payload, entityops.SchemaDirectoryAccount, installation, installation.DefinitionID, "run_1")

	assert.Equal(t, "int_owner", entityops.FieldValue(stamped, entityops.FieldManagedBy), "a mapping must not pick the managing installation")
	assert.Equal(t, "tenant-test", entityops.FieldValue(stamped, entityops.FieldSourceInstanceID), "a mapping must not pick the source instance")
	assert.Equal(t, "def_dir", entityops.FieldValue(stamped, entityops.FieldSourceDefinitionID), "a mapping must not pick the source definition")
	assert.Equal(t, "run_1", entityops.FieldValue(stamped, entityops.FieldIntegrationRunID))
	assert.Equal(t, "acct-1", entityops.FieldValue(stamped, "external_id"), "mapped non-provenance fields pass through")
}
