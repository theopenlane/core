package registry

import (
	"errors"
	"testing"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated/control"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

func TestResolveLinkEdge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		schema  *entityops.Schema
		rule    integrationtypes.LinkRule
		want    string
		wantErr error
	}{
		{
			name:   "unambiguous target type resolves without edge",
			schema: entityops.SchemaFinding,
			rule:   integrationtypes.LinkRule{TargetSchema: entityops.SchemaControl.Name},
			want:   "controls",
		},
		{
			name:   "explicit edge resolves",
			schema: entityops.SchemaFinding,
			rule:   integrationtypes.LinkRule{TargetSchema: entityops.SchemaControl.Name, Edge: "controls"},
			want:   "controls",
		},
		{
			name:    "ambiguous target type requires edge",
			schema:  entityops.SchemaAsset,
			rule:    integrationtypes.LinkRule{TargetSchema: entityops.SchemaAsset.Name},
			wantErr: ErrLinkEdgeAmbiguous,
		},
		{
			name:    "unknown target type",
			schema:  entityops.SchemaFinding,
			rule:    integrationtypes.LinkRule{TargetSchema: "NotASchema"},
			wantErr: ErrLinkEdgeNotFound,
		},
		{
			name:    "unknown edge name",
			schema:  entityops.SchemaFinding,
			rule:    integrationtypes.LinkRule{TargetSchema: entityops.SchemaControl.Name, Edge: "not_an_edge"},
			wantErr: ErrLinkEdgeNotFound,
		},
		{
			name:    "explicit edge targeting a different type",
			schema:  entityops.SchemaFinding,
			rule:    integrationtypes.LinkRule{TargetSchema: entityops.SchemaRisk.Name, Edge: "controls"},
			wantErr: ErrLinkEdgeNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			edge, err := ResolveLinkEdge(tc.schema, tc.rule)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if edge.Name != tc.want {
				t.Fatalf("expected edge %q, got %q", tc.want, edge.Name)
			}
		})
	}
}

func TestValidateLinkRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rule    integrationtypes.LinkRule
		wantErr error
	}{
		{
			name: "valid field match with scalar and list sources",
			rule: integrationtypes.LinkRule{
				TargetSchema: entityops.SchemaControl.Name,
				TargetField:  control.FieldRefCode,
				SourceField:  entityops.InputKeyFindingCategory,
				SourceList:   entityops.InputKeyFindingCategories,
			},
		},
		{
			name: "valid field match with list source only",
			rule: integrationtypes.LinkRule{
				TargetSchema: entityops.SchemaControl.Name,
				TargetField:  control.FieldRefCode,
				SourceList:   entityops.InputKeyFindingCategories,
			},
		},
		{
			name: "valid expression",
			rule: integrationtypes.LinkRule{
				TargetSchema: entityops.SchemaControl.Name,
				Expression:   `target.ref_code == source.category`,
			},
		},
		{
			name: "field match and expression together",
			rule: integrationtypes.LinkRule{
				TargetSchema: entityops.SchemaControl.Name,
				TargetField:  control.FieldRefCode,
				SourceField:  entityops.InputKeyFindingCategory,
				Expression:   `true`,
			},
			wantErr: ErrLinkRuleInvalid,
		},
		{
			name: "neither field match nor expression",
			rule: integrationtypes.LinkRule{
				TargetSchema: entityops.SchemaControl.Name,
			},
			wantErr: ErrLinkRuleInvalid,
		},
		{
			name: "target field without source fields",
			rule: integrationtypes.LinkRule{
				TargetSchema: entityops.SchemaControl.Name,
				TargetField:  control.FieldRefCode,
			},
			wantErr: ErrLinkRuleInvalid,
		},
		{
			name: "target field is not a match key",
			rule: integrationtypes.LinkRule{
				TargetSchema: entityops.SchemaControl.Name,
				TargetField:  "not_a_field",
				SourceField:  entityops.InputKeyFindingCategory,
			},
			wantErr: ErrLinkTargetFieldInvalid,
		},
		{
			name: "source field is not a mapped input key",
			rule: integrationtypes.LinkRule{
				TargetSchema: entityops.SchemaControl.Name,
				TargetField:  control.FieldRefCode,
				SourceField:  "not_a_field",
			},
			wantErr: ErrLinkSourceFieldInvalid,
		},
		{
			name: "list input key in the scalar slot",
			rule: integrationtypes.LinkRule{
				TargetSchema: entityops.SchemaControl.Name,
				TargetField:  control.FieldRefCode,
				SourceField:  entityops.InputKeyFindingCategories,
			},
			wantErr: ErrLinkSourceFieldInvalid,
		},
		{
			name: "scalar input key in the list slot",
			rule: integrationtypes.LinkRule{
				TargetSchema: entityops.SchemaControl.Name,
				TargetField:  control.FieldRefCode,
				SourceList:   entityops.InputKeyFindingCategory,
			},
			wantErr: ErrLinkSourceFieldInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateLinkRules(entityops.SchemaFinding, []integrationtypes.LinkRule{tc.rule})
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRegisterValidatesMappingLinks(t *testing.T) {
	t.Parallel()

	def, _ := minimalDefinition("def_link_validation")
	def.Operations[0].Handle = nil
	def.Operations[0].IngestHandle = newTestIngestHandler()
	def.Operations[0].Ingest = []integrationtypes.IngestContract{{Schema: entityops.SchemaFinding.Name}}
	def.Mappings = []integrationtypes.MappingRegistration{
		{
			Schema: entityops.SchemaFinding.Name,
			Spec: integrationtypes.MappingOverride{
				MapExpr: "payload",
				Links: []integrationtypes.LinkRule{
					{TargetSchema: entityops.SchemaControl.Name, TargetField: "not_a_field", SourceField: entityops.InputKeyFindingCategory},
				},
			},
		},
	}

	err := New().Register(def)
	if !errors.Is(err, ErrLinkTargetFieldInvalid) {
		t.Fatalf("expected registration to fail with %v, got %v", ErrLinkTargetFieldInvalid, err)
	}
}

func TestRegisterPopulatesLinkTargets(t *testing.T) {
	t.Parallel()

	def, _ := minimalDefinition("def_link_targets")
	def.Operations[0].Handle = nil
	def.Operations[0].IngestHandle = newTestIngestHandler()
	def.Operations[0].Ingest = []integrationtypes.IngestContract{{Schema: entityops.SchemaFinding.Name}}
	def.Mappings = []integrationtypes.MappingRegistration{
		{Schema: entityops.SchemaFinding.Name, Spec: integrationtypes.MappingOverride{MapExpr: "payload"}},
	}

	reg := New()
	if err := reg.Register(def); err != nil {
		t.Fatalf("register: %v", err)
	}

	registered, ok := reg.Definition("def_link_targets")
	if !ok {
		t.Fatal("definition not found after registration")
	}

	targets := registered.Mappings[0].LinkTargets
	if len(targets) == 0 {
		t.Fatal("expected link targets to be populated from the entityops catalog")
	}

	controls, found := lo.Find(targets, func(ti integrationtypes.LinkTargetInfo) bool { return ti.Edge == "controls" })
	if !found {
		t.Fatalf("expected a controls edge entry, got %+v", lo.Map(targets, func(ti integrationtypes.LinkTargetInfo, _ int) string { return ti.Edge }))
	}

	if controls.TargetType != entityops.SchemaControl.Name {
		t.Fatalf("expected controls entry to target %s, got %s", entityops.SchemaControl.Name, controls.TargetType)
	}

	if !lo.ContainsBy(controls.TargetFields, func(f integrationtypes.LinkFieldInfo) bool { return f.Name == control.FieldRefCode }) {
		t.Fatal("expected target fields to include the ref_code match key")
	}

	if lo.ContainsBy(controls.TargetFields, func(f integrationtypes.LinkFieldInfo) bool { return f.Name == "mapped_categories" }) {
		t.Fatal("expected target fields to exclude non-match-key fields")
	}

	categories, found := lo.Find(controls.SourceFields, func(f integrationtypes.LinkFieldInfo) bool { return f.Name == entityops.InputKeyFindingCategories })
	if !found {
		t.Fatal("expected source fields to include the categories input key")
	}

	if categories.Type != "[]string" {
		t.Fatalf("expected categories to carry its list type for slot routing, got %q", categories.Type)
	}
}
