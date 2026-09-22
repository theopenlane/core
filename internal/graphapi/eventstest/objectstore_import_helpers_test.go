//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/entitytype"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/objectstore"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
)

// fixture prefixes under testharness/testdata/objectstore, the empty prefix selects the harness default
const (
	fixturePrefixDefault  = ""
	fixturePrefixUpdated  = "vendors-updated/"
	fixturePrefixOpenlane = "vendors-openlane/"
	fixturePrefixScan     = "vendors-scan/"
)

const vendorEntityTypeName = "vendor"

// fixtureVendor mirrors one committed catalogue record
type fixtureVendor struct {
	ExternalID  string
	Name        string
	DisplayName string
	Description string
	Domain      string
}

// LogoRemoteURL returns the favicon URL the fixture records carry
func (v fixtureVendor) LogoRemoteURL() string {
	return "https://www.google.com/s2/favicons?domain=" + v.Domain + "&sz=128"
}

var fixtureTags = []string{"technology", "fixture"}

var (
	fixtureAlpha = fixtureVendor{
		ExternalID:  "entity::fixture-alpha",
		Name:        "fixture-alpha-vendor",
		DisplayName: "Fixture Alpha",
		Description: "Fixture Alpha provides identity services for the objectstore import tests.",
		Domain:      "fixture-alpha.example.com",
	}
	fixtureBeta = fixtureVendor{
		ExternalID:  "entity::fixture-beta",
		Name:        "fixture-beta-vendor",
		DisplayName: "Fixture Beta",
		Description: "Fixture Beta provides billing services for the objectstore import tests.",
		Domain:      "fixture-beta.example.com",
	}
	fixtureAlphaUpdated = fixtureVendor{
		ExternalID:  fixtureAlpha.ExternalID,
		Name:        fixtureAlpha.Name,
		DisplayName: "Fixture Alpha Updated",
		Description: "Fixture Alpha now provides identity and access services.",
		Domain:      fixtureAlpha.Domain,
	}
	fixtureBetaUpdated = fixtureVendor{
		ExternalID:  fixtureBeta.ExternalID,
		Name:        fixtureBeta.Name,
		DisplayName: "Fixture Beta Updated",
		Description: "Fixture Beta now provides billing and invoicing services.",
		Domain:      fixtureBeta.Domain,
	}
	fixtureCustomer = fixtureVendor{
		ExternalID:  "entity::fixture-gamma-customer",
		Name:        "fixture-gamma-customer",
		DisplayName: "Fixture Gamma",
		Domain:      "fixture-gamma.example.com",
	}
	fixtureOpenlane = fixtureVendor{
		ExternalID:  "entity::fixture-openlane",
		Name:        "Openlane",
		DisplayName: "Openlane Fixture",
		Description: "Openlane catalogue fixture matching the harness integration definitions' family.",
		Domain:      "fixture-openlane.example.com",
	}
	fixtureScan = fixtureVendor{
		ExternalID:  "entity::fixture-scan",
		Name:        "fixture-scan-vendor",
		DisplayName: "Fixture Scan",
		Description: "Fixture Scan is the catalogue vendor the domain scan import attributes an asset to.",
		Domain:      "fixture-scan.example.com",
	}
)

// systemVendorEntityType finds or creates the system-owned vendor EntityType
func systemVendorEntityType(t *testing.T) *ent.EntityType {
	t.Helper()

	ctx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)

	existing, err := suite.Client.DB.EntityType.Query().
		Where(entitytype.NameEqualFold(vendorEntityTypeName), entitytype.SystemOwned(true)).
		First(ctx)
	if err == nil {
		return existing
	}

	assert.Assert(t, ent.IsNotFound(err), "system vendor entity type lookup failed: %v", err)

	created, err := suite.Client.DB.EntityType.Create().SetName(vendorEntityTypeName).Save(ctx)
	th.RequireNoError(t, err)
	assert.Assert(t, created.SystemOwned, "the vendor entity type created without an owner must be system-owned")

	return created
}

// runSystemImport runs the objectstore system import on the harness runtime for the prefix
func runSystemImport(t *testing.T, ctx context.Context, prefix string) operations.IngestResult {
	t.Helper()

	var config json.RawMessage

	if prefix != "" {
		raw, err := json.Marshal(objectstore.SystemImport{Prefix: prefix})
		th.RequireNoError(t, err)

		config = raw
	}

	raw, err := suite.IntegrationsRT.ExecuteRuntimeOperation(ctx, objectstore.DefinitionID.ID(), objectstore.SystemImportOp.Name(), config)
	th.RequireNoError(t, err)

	var result operations.IngestResult
	th.RequireNoError(t, json.Unmarshal(raw, &result))

	return result
}

// deleteSystemVendors removes the system-owned rows carrying the fixture external ids
func deleteSystemVendors(t *testing.T, vendors ...fixtureVendor) {
	t.Helper()

	ctx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)

	ids, err := suite.Client.DB.Entity.Query().
		Where(
			entity.ExternalIDIn(lo.Map(vendors, func(v fixtureVendor, _ int) string { return v.ExternalID })...),
			entity.SystemOwned(true),
		).
		IDs(ctx)
	th.RequireNoError(t, err)

	deleteEntities(t, th.SharedSystemAdminUser.UserCtx, ids...)
}

// importFixtureVendors imports the prefix from a clean state and returns the created rows by external id
func importFixtureVendors(t *testing.T, ctx context.Context, prefix string, vendors ...fixtureVendor) (map[string]*ent.Entity, operations.IngestResult) {
	t.Helper()

	deleteSystemVendors(t, vendors...)

	t.Cleanup(func() {
		deleteSystemVendors(t, vendors...)
	})

	result := runSystemImport(t, ctx, prefix)
	assert.Assert(t, is.Equal(0, result.Failed), "catalogue import must not fail: %+v", result.Failures)
	assert.Assert(t, is.Equal(len(vendors), result.Changed), "every fixture vendor must be created by the import")

	rows := make(map[string]*ent.Entity, len(vendors))
	for _, vendor := range vendors {
		rows[vendor.ExternalID] = systemVendorByExternalID(t, ctx, vendor.ExternalID)
	}

	return rows, result
}

// systemVendorByExternalID loads the one system-owned entity carrying externalID
func systemVendorByExternalID(t *testing.T, ctx context.Context, externalID string) *ent.Entity {
	t.Helper()

	row, err := suite.Client.DB.Entity.Query().
		Where(entity.ExternalID(externalID), entity.SystemOwned(true)).
		Only(ctx)
	th.RequireNoError(t, err)

	return row
}

// deleteEntities removes the entity rows under the caller context that owns them
func deleteEntities(t *testing.T, ctx context.Context, ids ...string) {
	t.Helper()

	if len(ids) == 0 {
		return
	}

	(&th.Cleanup[*ent.EntityDeleteOne]{Client: suite.Client.DB.Entity, IDs: ids}).MustDelete(ctx, t)
}

// orgVendorEntityTypeID returns the organization's own vendor EntityType id
func orgVendorEntityTypeID(t *testing.T, ctx context.Context, orgID string) string {
	t.Helper()

	id, err := suite.Client.DB.EntityType.Query().
		Where(entitytype.NameEqualFold(vendorEntityTypeName), entitytype.OwnerID(orgID)).
		OnlyID(ctx)
	th.RequireNoError(t, err)

	return id
}
