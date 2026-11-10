package providerkit

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/asset"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestDirectoryIngestContracts(t *testing.T) {
	t.Parallel()

	got := lo.Map(DirectoryIngestContracts(), func(contract types.IngestContract, _ int) string { return contract.Schema })
	want := []string{entityops.SchemaDirectoryAccount.Name, entityops.SchemaDirectoryGroup.Name, entityops.SchemaDirectoryMembership.Name}

	if !slices.Equal(got, want) {
		t.Fatalf("expected contracts %v, got %v", want, got)
	}
}

func TestDirectoryMappings(t *testing.T) {
	t.Parallel()

	mappings := DirectoryMappings("account", "group", "membership")
	if len(mappings) != 3 {
		t.Fatalf("expected 3 mappings, got %d", len(mappings))
	}

	for i, want := range []struct {
		schema  string
		mapExpr string
		links   int
	}{
		{schema: entityops.SchemaDirectoryAccount.Name, mapExpr: "account"},
		{schema: entityops.SchemaDirectoryGroup.Name, mapExpr: "group"},
		{schema: entityops.SchemaDirectoryMembership.Name, mapExpr: "membership", links: 2},
	} {
		got := mappings[i]
		if got.Schema != want.schema || got.Spec.MapExpr != want.mapExpr || got.Spec.FilterExpr != matchAllFilterExpr || len(got.Spec.Links) != want.links {
			t.Fatalf("mapping %d: unexpected registration %+v", i, got)
		}
	}
}

func TestFindingMapping(t *testing.T) {
	t.Parallel()

	got := FindingMapping("finding")
	if got.Schema != entityops.SchemaFinding.Name || got.Spec.MapExpr != "finding" || got.Spec.FilterExpr != matchAllFilterExpr {
		t.Fatalf("unexpected registration %+v", got)
	}

	want := []types.LinkRule{
		{
			TargetSchema: entityops.SchemaControl.Name,
			TargetField:  control.FieldRefCode,
			SourceField:  entityops.FindingFields.Category.InputKey,
			SourceList:   entityops.FindingFields.Categories.InputKey,
		},
		{
			TargetSchema: entityops.SchemaAsset.Name,
			TargetField:  asset.FieldName,
			SourceField:  entityops.FindingFields.ResourceName.InputKey,
			SourceList:   entityops.FindingFields.Targets.InputKey,
		},
		{
			TargetSchema: entityops.SchemaIdentityHolder.Name,
			Edge:         finding.EdgeInternalOwnerIdentityHolder,
			TargetField:  identityholder.FieldEmail,
			SourceField:  entityops.FindingFields.InternalOwner.InputKey,
		},
	}

	if !slices.Equal(got.Spec.Links, want) {
		t.Fatalf("expected links %+v, got %+v", want, got.Spec.Links)
	}
}

func TestCheckResultMapping(t *testing.T) {
	t.Parallel()

	got := CheckResultMapping("check")
	if got.Schema != entityops.SchemaCheckResult.Name || got.Spec.MapExpr != "check" || got.Spec.FilterExpr != matchAllFilterExpr {
		t.Fatalf("unexpected registration %+v", got)
	}

	want := []types.LinkRule{
		{
			TargetSchema: entityops.SchemaControl.Name,
			TargetField:  control.FieldRefCode,
			SourceList:   entityops.CheckResultFields.Tags.InputKey,
		},
	}

	if !slices.Equal(got.Spec.Links, want) {
		t.Fatalf("expected links %+v, got %+v", want, got.Spec.Links)
	}
}

func TestVulnerabilityMapping(t *testing.T) {
	t.Parallel()

	got := VulnerabilityMapping("vulnerability")
	if got.Schema != entityops.SchemaVulnerability.Name || got.Spec.MapExpr != "vulnerability" || got.Spec.FilterExpr != matchAllFilterExpr {
		t.Fatalf("unexpected registration %+v", got)
	}

	want := []types.LinkRule{
		{
			TargetSchema: entityops.SchemaIdentityHolder.Name,
			Edge:         vulnerability.EdgeInternalOwnerIdentityHolder,
			TargetField:  identityholder.FieldEmail,
			SourceField:  entityops.VulnerabilityFields.InternalOwner.InputKey,
		},
	}

	if !slices.Equal(got.Spec.Links, want) {
		t.Fatalf("expected links %+v, got %+v", want, got.Spec.Links)
	}
}

func TestDirectoryPayloadSets(t *testing.T) {
	t.Parallel()

	envelopes := []types.MappingEnvelope{RawEnvelope("r", nil)}

	accounts := DirectoryAccountPayloadSets(envelopes)
	if len(accounts) != 1 || accounts[0].Schema != entityops.SchemaDirectoryAccount.Name || !accounts[0].SnapshotComplete || len(accounts[0].Envelopes) != 1 {
		t.Fatalf("unexpected account payload sets %+v", accounts)
	}

	groups := DirectoryGroupPayloadSets(envelopes, nil, false)
	if len(groups) != 2 || groups[0].Schema != entityops.SchemaDirectoryGroup.Name || groups[1].Schema != entityops.SchemaDirectoryMembership.Name {
		t.Fatalf("unexpected group payload sets %+v", groups)
	}

	if groups[0].SnapshotComplete || groups[1].SnapshotComplete {
		t.Fatal("expected incomplete group and membership payload sets")
	}
}

func TestOAuthToken(t *testing.T) {
	t.Parallel()

	expiry := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	withExpiry := OAuthToken("access", "refresh", &expiry)
	if withExpiry.AccessToken != "access" || withExpiry.RefreshToken != "refresh" || withExpiry.TokenType != bearerTokenType || !withExpiry.Expiry.Equal(expiry) {
		t.Fatalf("unexpected token %+v", withExpiry)
	}

	if withoutExpiry := OAuthToken("access", "", nil); !withoutExpiry.Expiry.IsZero() {
		t.Fatalf("expected zero expiry, got %v", withoutExpiry.Expiry)
	}
}

func TestEdgeKeyLinks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		schema      *entityops.Schema
		edge        string
		targetField string
		list        string
		field       string
	}{
		{name: "to-many edge reads external ids keyed by the edge", schema: entityops.SchemaProcedure, edge: "programs", targetField: externalIDField, list: "programs_external_ids"},
		{name: "unique edge reads an external id keyed by the edge", schema: entityops.SchemaCampaign, edge: "assessment", targetField: externalIDField, field: "assessment_external_id"},
		{name: "edges to the same target stay apart", schema: entityops.SchemaPlatform, edge: "out_of_scope_assets", targetField: externalIDField, list: "out_of_scope_assets_external_ids"},
		{name: "unique group edge reads a name", schema: entityops.SchemaControl, edge: "control_owner", targetField: nameField, field: "control_owner_name"},
		{name: "to-many group edge reads names", schema: entityops.SchemaControl, edge: "editors", targetField: nameField, list: "editors_names"},
		{name: "entity type edge reads a name", schema: entityops.SchemaEntity, edge: "entity_type", targetField: nameField, field: "entity_type_name"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rule, found := lo.Find(EdgeKeyLinks(tc.schema), func(rule types.LinkRule) bool {
				return rule.Edge == tc.edge && rule.TargetField == tc.targetField
			})
			if !found {
				t.Fatalf("expected a %s link through %s.%s", tc.targetField, tc.schema.Name, tc.edge)
			}

			if !rule.PayloadKeys || rule.SourceList != tc.list || rule.SourceField != tc.field {
				t.Fatalf("unexpected rule %+v", rule)
			}
		})
	}
}

func TestEdgeKeyLinksSkipCollidingKeys(t *testing.T) {
	t.Parallel()

	for _, schema := range entityops.AllSchemas() {
		for _, rule := range EdgeKeyLinks(schema) {
			key := lo.CoalesceOrEmpty(rule.SourceList, rule.SourceField)
			if inputKeyCollides(schema, key) {
				t.Fatalf("%s link key %s collides with an input key", schema.Name, key)
			}
		}
	}
}

func TestUpgradeFromSection(t *testing.T) {
	t.Parallel()

	upgrade := UpgradeFromSection[DirectorySync]("findingSync")

	tests := []struct {
		name   string
		stored string
		want   DirectorySync
	}{
		{name: "main's section under the key is decoded", stored: `{"region":"eu","findingSync":{"disable":true}}`, want: DirectorySync{OperationSettings: types.OperationSettings{Disable: true}}},
		{name: "a stored document without the key is decoded as is", stored: `{"disableGroupSync":true}`, want: DirectorySync{DisableGroupSync: true}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := upgrade(context.Background(), types.InstallationRequest{}, "", json.RawMessage(tc.stored))
			if err != nil {
				t.Fatalf("upgrade: %v", err)
			}

			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
