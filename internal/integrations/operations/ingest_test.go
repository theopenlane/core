package operations

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

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
			if found != tc.wantFound {
				t.Fatalf("findMapping found=%v, want %v", found, tc.wantFound)
			}

			if found && got.MapExpr != tc.wantExpr {
				t.Fatalf("findMapping MapExpr=%q, want %q", got.MapExpr, tc.wantExpr)
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
			if got != tc.want {
				t.Fatalf("contractIncludesSchema(%q)=%v, want %v", tc.schema, got, tc.want)
			}
		})
	}
}

func TestContractIncludesSchema_EmptyContracts(t *testing.T) {
	t.Parallel()

	if contractIncludesSchema(nil, "asset") {
		t.Fatal("expected false for nil contracts")
	}
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
			if got != tc.want {
				t.Fatalf("needsDirectorySyncRun()=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestResolveInstallationFilterExpr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   json.RawMessage
		wantExpr string
		wantErr  bool
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
			name:     "config with filter expression",
			config:   json.RawMessage(`{"filterExpr":"resource == \"users\""}`),
			wantExpr: `resource == "users"`,
		},
		{
			name:    "invalid JSON config",
			config:  json.RawMessage(`{not json`),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			installation := &ent.Integration{
				Config: openapi.IntegrationConfig{ClientConfig: tc.config},
			}

			expr, err := resolveInstallationFilterExpr(installation)
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr != tc.wantExpr {
				t.Fatalf("expr=%q, want %q", expr, tc.wantExpr)
			}
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
		name                  string
		installationFilterExpr string
		mappingFilterExpr      string
		wantMatch             bool
		wantErr               bool
	}{
		{
			name:                  "both empty passes",
			installationFilterExpr: "",
			mappingFilterExpr:      "",
			wantMatch:             true,
		},
		{
			name:                  "installation filter matches",
			installationFilterExpr: `resource == "users"`,
			mappingFilterExpr:      "",
			wantMatch:             true,
		},
		{
			name:                  "installation filter rejects",
			installationFilterExpr: `resource == "groups"`,
			mappingFilterExpr:      "",
			wantMatch:             false,
		},
		{
			name:                  "mapping filter matches",
			installationFilterExpr: "",
			mappingFilterExpr:      `action == "create"`,
			wantMatch:             true,
		},
		{
			name:                  "mapping filter rejects",
			installationFilterExpr: "",
			mappingFilterExpr:      `action == "delete"`,
			wantMatch:             false,
		},
		{
			name:                  "both filters match",
			installationFilterExpr: `resource == "users"`,
			mappingFilterExpr:      `action == "create"`,
			wantMatch:             true,
		},
		{
			name:                  "installation matches mapping rejects",
			installationFilterExpr: `resource == "users"`,
			mappingFilterExpr:      `action == "delete"`,
			wantMatch:             false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			matched, err := envelopeIncludedByFilters(context.Background(), tc.installationFilterExpr, tc.mappingFilterExpr, envelope)
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if matched != tc.wantMatch {
				t.Fatalf("matched=%v, want %v", matched, tc.wantMatch)
			}
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
			switch {
			case tc.wantErr != nil && !errors.Is(err, tc.wantErr):
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			case tc.wantErr != nil:
				return
			case err != nil:
				t.Fatalf("unexpected error: %v", err)
			}

			if include != tc.wantInclude {
				t.Fatalf("include=%v, want %v", include, tc.wantInclude)
			}

			if include && record.Schema != tc.schema {
				t.Fatalf("record.Schema=%q, want %q", record.Schema, tc.schema)
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
			if tc.wantNil && got != nil {
				t.Fatalf("expected nil, got %v", got)
			}
			if tc.wantErr != nil && !errors.Is(got, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, got)
			}
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
		Registry:     reg,
		Installation: &ent.Integration{DefinitionID: "nonexistent"},
	}

	err := processPayloadSets(context.Background(), ic, nil, nil, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return nil
	})

	if !errors.Is(err, ErrIngestDefinitionNotFound) {
		t.Fatalf("expected ErrIngestDefinitionNotFound, got %v", err)
	}
}

func TestProcessPayloadSets_SchemaNotDeclared(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, []types.MappingRegistration{
		{Schema: integrationgenerated.IntegrationMappingSchemaAsset, Variant: "", Spec: types.MappingOverride{MapExpr: `payload`}},
	})

	ic := IngestContext{
		Registry:     reg,
		Installation: &ent.Integration{DefinitionID: "test-def"},
	}

	// contracts say "contact", but payload set says "asset"
	contracts := []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaContact}}
	payloadSets := []types.IngestPayloadSet{
		{Schema: integrationgenerated.IntegrationMappingSchemaAsset, Envelopes: []types.MappingEnvelope{{Payload: json.RawMessage(`{}`)}}},
	}

	err := processPayloadSets(context.Background(), ic, contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return nil
	})

	if !errors.Is(err, ErrIngestSchemaNotDeclared) {
		t.Fatalf("expected ErrIngestSchemaNotDeclared, got %v", err)
	}
}

func TestProcessPayloadSets_SchemaNotFound(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry:     reg,
		Installation: &ent.Integration{DefinitionID: "test-def"},
	}

	contracts := []types.IngestContract{{Schema: "totally_bogus_schema"}}
	payloadSets := []types.IngestPayloadSet{
		{Schema: "totally_bogus_schema", Envelopes: []types.MappingEnvelope{{Payload: json.RawMessage(`{}`)}}},
	}

	err := processPayloadSets(context.Background(), ic, contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return nil
	})

	if !errors.Is(err, ErrIngestSchemaNotFound) {
		t.Fatalf("expected ErrIngestSchemaNotFound, got %v", err)
	}
}

func TestProcessPayloadSets_MappingNotFound(t *testing.T) {
	t.Parallel()

	// register definition with no mappings for asset
	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry:     reg,
		Installation: &ent.Integration{DefinitionID: "test-def"},
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

	err := processPayloadSets(context.Background(), ic, contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return nil
	})

	if !errors.Is(err, ErrIngestMappingNotFound) {
		t.Fatalf("expected ErrIngestMappingNotFound, got %v", err)
	}
}

func TestProcessPayloadSets_InvalidInstallationFilterConfig(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry: reg,
		Installation: &ent.Integration{
			DefinitionID: "test-def",
			Config:       openapi.IntegrationConfig{ClientConfig: json.RawMessage(`{invalid`)},
		},
	}

	err := processPayloadSets(context.Background(), ic, nil, nil, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return nil
	})

	if !errors.Is(err, ErrIngestInstallationFilterConfigInvalid) {
		t.Fatalf("expected ErrIngestInstallationFilterConfigInvalid, got %v", err)
	}
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
		Registry:     reg,
		Installation: &ent.Integration{DefinitionID: "test-def"},
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

	err := processPayloadSets(context.Background(), ic, contracts, payloadSets, IngestOptions{}, func(_ context.Context, record mappedIngestRecord) error {
		handled = append(handled, record)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(handled) != 1 {
		t.Fatalf("expected 1 handled record, got %d", len(handled))
	}
	if handled[0].Schema != integrationgenerated.IntegrationMappingSchemaAsset {
		t.Fatalf("schema=%q, want %q", handled[0].Schema, integrationgenerated.IntegrationMappingSchemaAsset)
	}
}

func TestProcessPayloadSets_EmptyPayloadSets(t *testing.T) {
	t.Parallel()

	reg, _ := testDefinition(t, nil)

	ic := IngestContext{
		Registry:     reg,
		Installation: &ent.Integration{DefinitionID: "test-def"},
	}

	err := processPayloadSets(context.Background(), ic, nil, nil, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		t.Fatal("handler should not be called for empty payload sets")
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
		Registry:     reg,
		Installation: &ent.Integration{DefinitionID: "test-def"},
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

	err := processPayloadSets(context.Background(), ic, contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		handled++
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if handled != 1 {
		t.Fatalf("expected 1 handled record (1 filtered), got %d", handled)
	}
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
		Registry:     reg,
		Installation: &ent.Integration{DefinitionID: "test-def"},
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

	err := processPayloadSets(context.Background(), ic, contracts, payloadSets, IngestOptions{}, func(context.Context, mappedIngestRecord) error {
		return handleErr
	})

	if !errors.Is(err, handleErr) {
		t.Fatalf("expected handleErr, got %v", err)
	}
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
		Registry:     reg,
		Installation: &ent.Integration{DefinitionID: "test-def"},
		Runtime:      nil,
	}

	contracts := []types.IngestContract{{Schema: integrationgenerated.IntegrationMappingSchemaAsset}}
	payloadSets := []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: []types.MappingEnvelope{{Payload: json.RawMessage(`{}`)}},
		},
	}

	err := EmitPayloadSets(context.Background(), ic, "sync", contracts, payloadSets, IngestOptions{})
	if !errors.Is(err, ErrGalaRequired) {
		t.Fatalf("expected ErrGalaRequired, got %v", err)
	}
}
