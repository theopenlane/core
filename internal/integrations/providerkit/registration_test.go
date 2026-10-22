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
