package operations

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
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

func TestNeedsDirectorySyncRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		contracts []types.IngestContract
		want      bool
	}{
		{
			name:      "directory account triggers sync run",
			contracts: []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount}},
			want:      true,
		},
		{
			name:      "directory group triggers sync run",
			contracts: []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaDirectoryGroup}},
			want:      true,
		},
		{
			name:      "directory membership triggers sync run",
			contracts: []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaDirectoryMembership}},
			want:      true,
		},
		{
			name:      "asset does not trigger sync run",
			contracts: []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaAsset}},
			want:      false,
		},
		{
			name:      "mixed with directory schema triggers",
			contracts: []types.IngestContract{{Schema: "asset"}, {Schema: integrationgenerated.IntegrationMappingSchemaDirectoryGroup}},
			want:      true,
		},
		{
			name:      "nil contracts",
			contracts: nil,
			want:      false,
		},
		{
			name:      "empty contracts",
			contracts: []types.IngestContract{},
			want:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := needsDirectorySyncRun(tc.contracts)
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestResolveInstallationFilterExpr(t *testing.T) {
	t.Parallel()

	// configResolverFor returns a ConfigResolver that extracts the named top-level key as raw JSON
	configResolverFor := func(key string) func(json.RawMessage) json.RawMessage {
		return func(userInput json.RawMessage) json.RawMessage {
			var top map[string]json.RawMessage
			if err := json.Unmarshal(userInput, &top); err != nil {
				return nil
			}
			return top[key]
		}
	}

	tests := []struct {
		name          string
		config        json.RawMessage
		definition    types.Definition
		operationName string
		wantExpr      string
		wantErr       bool
	}{
		{
			name:     "nil config returns empty",
			config:   nil,
			wantExpr: "",
		},
		{
			name:     "empty config returns empty",
			config:   json.RawMessage(`{}`),
			wantExpr: "",
		},
		{
			name:     "flat top-level filterExpr",
			config:   json.RawMessage(`{"filterExpr":"resource == \"users\""}`),
			wantExpr: `resource == "users"`,
		},
		{
			name:    "invalid JSON config",
			config:  json.RawMessage(`{not json`),
			wantErr: true,
		},
		{
			name:   "nested filterExpr resolved via ConfigResolver",
			config: json.RawMessage(`{"directorySync":{"filterExpr":"payload.is_external == false"}}`),
			definition: types.Definition{
				Operations: []types.OperationRegistration{
					{
						Name:           "directory-sync",
						ConfigResolver: configResolverFor("directorySync"),
					},
				},
			},
			operationName: "directory-sync",
			wantExpr:      "payload.is_external == false",
		},
		{
			// when ConfigResolver extracts a section that has no filterExpr, the flat top-level key is used as fallback
			name:   "ConfigResolver section has no filterExpr falls back to flat",
			config: json.RawMessage(`{"directorySync":{},"filterExpr":"resource == \"users\""}`),
			definition: types.Definition{
				Operations: []types.OperationRegistration{
					{
						Name:           "directory-sync",
						ConfigResolver: configResolverFor("directorySync"),
					},
				},
			},
			operationName: "directory-sync",
			wantExpr:      `resource == "users"`,
		},
		{
			// when ConfigResolver returns nil (e.g. key not present), fall back to flat
			name:   "ConfigResolver returns nil falls back to flat",
			config: json.RawMessage(`{"filterExpr":"resource == \"users\""}`),
			definition: types.Definition{
				Operations: []types.OperationRegistration{
					{
						Name:           "directory-sync",
						ConfigResolver: func(json.RawMessage) json.RawMessage { return nil },
					},
				},
			},
			operationName: "directory-sync",
			wantExpr:      `resource == "users"`,
		},
		{
			// when the operationName doesn't match any registered operation, fall back to flat
			name:   "unknown operationName falls back to flat",
			config: json.RawMessage(`{"filterExpr":"resource == \"groups\""}`),
			definition: types.Definition{
				Operations: []types.OperationRegistration{
					{Name: "other-op"},
				},
			},
			operationName: "unknown-op",
			wantExpr:      `resource == "groups"`,
		},
		{
			// when the matching operation has no ConfigResolver, fall back to flat
			name:   "operation without ConfigResolver falls back to flat",
			config: json.RawMessage(`{"filterExpr":"resource == \"assets\""}`),
			definition: types.Definition{
				Operations: []types.OperationRegistration{
					{Name: "asset-sync"},
				},
			},
			operationName: "asset-sync",
			wantExpr:      `resource == "assets"`,
		},
		{
			// ConfigResolver for one operation does not affect a different operation
			name:   "ConfigResolver not used when operationName does not match",
			config: json.RawMessage(`{"directorySync":{"filterExpr":"payload.type == \"user\""},"filterExpr":"resource == \"assets\""}`),
			definition: types.Definition{
				Operations: []types.OperationRegistration{
					{
						Name:           "directory-sync",
						ConfigResolver: configResolverFor("directorySync"),
					},
					{Name: "asset-sync"},
				},
			},
			operationName: "asset-sync",
			wantExpr:      `resource == "assets"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			installation := &ent.Integration{
				Config: openapi.IntegrationConfig{ClientConfig: tc.config},
			}

			expr, err := resolveInstallationFilterExpr(installation, tc.definition, tc.operationName)
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

			matched, err := envelopeIncludedByFilters(context.Background(), tc.installationFilterExpr, tc.mappingFilterExpr, envelope)
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
		{
			name:   "no mapping found",
			schema: "unknown_schema",
			envelope: types.MappingEnvelope{
				Variant: "",
				Payload: json.RawMessage(`{}`),
			},
			wantErr: ErrIngestMappingNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			record, include, err := mapIngestRecord(context.Background(), definition, tc.schema, tc.envelope, "")
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

func TestProcessPayloadSets_DefinitionNotFound(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "nonexistent"},
	}

	err := applyPayloadSets(context.Background(), ic, "", nil, nil, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return nil
	})

	assert.ErrorIs(t, err, ErrIngestDefinitionNotFound)
}

func TestProcessPayloadSets_SchemaNotDeclared(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, []types.MappingRegistration{
		{Schema: integrationgenerated.IntegrationMappingSchemaAsset, Variant: "", Spec: types.MappingOverride{MapExpr: `payload`}},
	})

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def"},
	}

	// contracts say "contact", but payload set says "asset"
	contracts := []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaContact}}
	payloadSets := []types.IngestPayloadSet{
		{Schema: integrationgenerated.IntegrationMappingSchemaAsset, Envelopes: []types.MappingEnvelope{{Payload: json.RawMessage(`{}`)}}},
	}

	err := applyPayloadSets(context.Background(), ic, "", contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return nil
	})

	assert.ErrorIs(t, err, ErrIngestSchemaNotDeclared)
}

func TestProcessPayloadSets_SchemaNotFound(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def"},
	}

	contracts := []types.IngestContract{{Schema: "totally_bogus_schema"}}
	payloadSets := []types.IngestPayloadSet{
		{Schema: "totally_bogus_schema", Envelopes: []types.MappingEnvelope{{Payload: json.RawMessage(`{}`)}}},
	}

	err := applyPayloadSets(context.Background(), ic, "", contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return nil
	})

	assert.ErrorIs(t, err, ErrIngestSchemaNotFound)
}

func TestProcessPayloadSets_MappingNotFound(t *testing.T) {
	t.Parallel()

	// register definition with no mappings for asset
	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def"},
	}

	contracts := []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaAsset}}
	payloadSets := []types.IngestPayloadSet{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: []types.MappingEnvelope{
				{Variant: "", Payload: json.RawMessage(`{"name":"test"}`)},
			},
		},
	}

	err := applyPayloadSets(context.Background(), ic, "", contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return nil
	})

	assert.ErrorIs(t, err, ErrIngestMappingNotFound)
}

func TestProcessPayloadSets_InvalidInstallationFilterConfig(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry: reg,
		Integration: &ent.Integration{
			DefinitionID: "test-def",
			Config:       openapi.IntegrationConfig{ClientConfig: json.RawMessage(`{invalid`)},
		},
	}

	err := applyPayloadSets(context.Background(), ic, "", nil, nil, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return nil
	})

	assert.ErrorIs(t, err, ErrIngestInstallationFilterConfigInvalid)
}

func TestProcessPayloadSets_SuccessfulMapping(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, []types.MappingRegistration{
		{
			Schema:  integrationgenerated.IntegrationMappingSchemaAsset,
			Variant: "",
			Spec:    types.MappingOverride{MapExpr: `{"sourceIdentifier": payload.id}`},
		},
	})

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def"},
	}

	contracts := []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaAsset}}
	payloadSets := []types.IngestPayloadSet{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: []types.MappingEnvelope{
				{Variant: "", Payload: json.RawMessage(`{"id":"asset-001"}`)},
			},
		},
	}

	var handled []mappedIngestRecord

	err := applyPayloadSets(context.Background(), ic, "", contracts, payloadSets, IngestOptions{}, func(_ context.Context, record mappedIngestRecord) error {
		handled = append(handled, record)
		return nil
	})

	assert.NilError(t, err)
	assert.Equal(t, len(handled), 1)
	assert.Equal(t, handled[0].Schema, integrationgenerated.IntegrationMappingSchemaAsset)
}

func TestProcessPayloadSets_EmptyPayloadSets(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def"},
	}

	err := applyPayloadSets(context.Background(), ic, "", nil, nil, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		t.Fatal("handler should not be called for empty payload sets")
		return nil
	})

	assert.NilError(t, err)
}

func TestProcessPayloadSets_FilteredEnvelopes(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, []types.MappingRegistration{
		{
			Schema:  integrationgenerated.IntegrationMappingSchemaAsset,
			Variant: "",
			Spec: types.MappingOverride{
				FilterExpr: `resource == "wanted"`,
				MapExpr:    `payload`,
			},
		},
	})

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def"},
	}

	contracts := []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaAsset}}
	payloadSets := []types.IngestPayloadSet{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: []types.MappingEnvelope{
				{Resource: "unwanted", Payload: json.RawMessage(`{"name":"skip"}`)},
				{Resource: "wanted", Payload: json.RawMessage(`{"name":"keep"}`)},
			},
		},
	}

	var handled int

	err := applyPayloadSets(context.Background(), ic, "", contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		handled++
		return nil
	})

	assert.NilError(t, err)
	assert.Equal(t, handled, 1)
}

func TestProcessPayloadSets_HandleError(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, []types.MappingRegistration{
		{
			Schema:  integrationgenerated.IntegrationMappingSchemaAsset,
			Variant: "",
			Spec:    types.MappingOverride{MapExpr: `payload`},
		},
	})

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def"},
	}

	contracts := []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaAsset}}
	payloadSets := []types.IngestPayloadSet{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: []types.MappingEnvelope{
				{Payload: json.RawMessage(`{"name":"test"}`)},
			},
		},
	}

	handleErr := errors.New("persist failed")

	err := applyPayloadSets(context.Background(), ic, "", contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return handleErr
	})

	assert.ErrorIs(t, err, handleErr)
}

// TestProcessPayloadSets_NestedInstallationFilter verifies that a filterExpr stored inside a
// per-operation config section (e.g. repoSync.filterExpr) is applied when the operation's
// ConfigResolver is wired — this is the scenario that was previously broken for Slack and others
func TestProcessPayloadSets_NestedInstallationFilter(t *testing.T) {
	t.Parallel()

	type repoSyncCfg struct {
		FilterExpr string `json:"filterExpr,omitempty"`
	}
	type userInput struct {
		RepoSync repoSyncCfg `json:"repoSync,omitempty"`
	}

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
				ConfigResolver: func(raw json.RawMessage) json.RawMessage {
					var u userInput
					if err := json.Unmarshal(raw, &u); err != nil {
						return nil
					}
					out, _ := json.Marshal(u.RepoSync)
					return out
				},
			},
		},
		Mappings: []types.MappingRegistration{
			{
				Schema:  integrationgenerated.IntegrationMappingSchemaAsset,
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
			DefinitionID: "test-def",
			Config: openapi.IntegrationConfig{
				ClientConfig: json.RawMessage(`{"repoSync":{"filterExpr":"payload.is_private == true"}}`),
			},
		},
	}

	contracts := []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaAsset}}
	payloadSets := []types.IngestPayloadSet{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: []types.MappingEnvelope{
				{Payload: json.RawMessage(`{"name":"private-repo","is_private":true}`)},
				{Payload: json.RawMessage(`{"name":"public-repo","is_private":false}`)},
				{Payload: json.RawMessage(`{"name":"another-private","is_private":true}`)},
			},
		},
	}

	var handled int
	err := applyPayloadSets(context.Background(), ic, "repo-sync", contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		handled++
		return nil
	})

	assert.NilError(t, err)
	assert.Equal(t, handled, 2) // private repos pass; public-repo is filtered
}

// TestProcessPayloadSets_NestedFilterDoesNotLeakAcrossOperations verifies that a nested filterExpr
// stored under operation A's config section is not applied when processing under operation B
func TestProcessPayloadSets_NestedFilterDoesNotLeakAcrossOperations(t *testing.T) {
	t.Parallel()

	type findingSyncCfg struct {
		FilterExpr string `json:"filterExpr,omitempty"`
	}
	type assetSyncCfg struct {
		FilterExpr string `json:"filterExpr,omitempty"`
	}
	type userInput struct {
		FindingSync findingSyncCfg `json:"findingSync,omitempty"`
		DirectorySync assetSyncCfg `json:"directorySync,omitempty"`
	}

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:     "test-def",
			Active: true,
		},
		Operations: []types.OperationRegistration{
			{
				Name: "finding-sync",
				Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
					return nil, nil
				},
				ConfigResolver: func(raw json.RawMessage) json.RawMessage {
					var u userInput
					if err := json.Unmarshal(raw, &u); err != nil {
						return nil
					}
					out, _ := json.Marshal(u.FindingSync)
					return out
				},
			},
			{
				Name: "asset-sync",
				Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
					return nil, nil
				},
				ConfigResolver: func(raw json.RawMessage) json.RawMessage {
					var u userInput
					if err := json.Unmarshal(raw, &u); err != nil {
						return nil
					}
					out, _ := json.Marshal(u.DirectorySync)
					return out
				},
			},
		},
		Mappings: []types.MappingRegistration{
			{
				Schema:  integrationgenerated.IntegrationMappingSchemaAsset,
				Variant: "",
				Spec:    types.MappingOverride{MapExpr: `payload`},
			},
		},
	}

	reg := registry.New()
	assert.NilError(t, reg.Register(def))

	// findingSync has a restrictive filter; asset-sync has none
	ic := IngestContext{
		Registry: reg,
		Integration: &ent.Integration{
			DefinitionID: "test-def",
			Config: openapi.IntegrationConfig{
				ClientConfig: json.RawMessage(`{"findingSync":{"filterExpr":"payload.severity == \"CRITICAL\""},"directorySync":{}}`),
			},
		},
	}

	contracts := []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaAsset}}
	payloadSets := []types.IngestPayloadSet{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: []types.MappingEnvelope{
				{Payload: json.RawMessage(`{"name":"asset-001"}`)},
				{Payload: json.RawMessage(`{"name":"asset-002"}`)},
			},
		},
	}

	var handled int
	// running as asset-sync: the findingSync filterExpr must NOT apply
	err := applyPayloadSets(context.Background(), ic, "asset-sync", contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		handled++
		return nil
	})

	assert.NilError(t, err)
	assert.Equal(t, handled, 2) // both pass — asset-sync has no filter
}

func TestEmitPayloadSets_NilRuntime_NonDirectorySync(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, []types.MappingRegistration{
		{
			Schema:  integrationgenerated.IntegrationMappingSchemaAsset,
			Variant: "",
			Spec:    types.MappingOverride{MapExpr: `payload`},
		},
	})

	ic := IngestContext{
		Registry:    reg,
		Integration: &ent.Integration{DefinitionID: "test-def"},
		Runtime:     nil,
	}

	contracts := []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaAsset}}
	payloadSets := []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: []types.MappingEnvelope{{Payload: json.RawMessage(`{}`)}},
		},
	}

	err := EmitPayloadSets(context.Background(), ic, "sync", contracts, payloadSets, IngestOptions{})
	assert.ErrorIs(t, err, ErrGalaRequired)
}
