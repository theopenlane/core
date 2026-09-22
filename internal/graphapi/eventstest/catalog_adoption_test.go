//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/asset"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/cloudflare"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
	"github.com/theopenlane/core/v2/pkg/domainscan"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// orgEntitiesNamed loads the organization's entities whose name matches, case-insensitively
func orgEntitiesNamed(t *testing.T, ctx context.Context, orgID, name string) []*ent.Entity {
	t.Helper()

	rows, err := suite.Client.DB.Entity.Query().
		Where(entity.OwnerID(orgID), entity.NameEqualFold(name)).
		All(ctx)
	th.RequireNoError(t, err)

	return rows
}

// entityHasIntegration reports whether the entity carries the integration edge
func entityHasIntegration(t *testing.T, ctx context.Context, entityID, integrationID string) bool {
	t.Helper()

	linked, err := suite.Client.DB.Entity.Query().
		Where(entity.ID(entityID), entity.HasIntegrationsWith(integration.ID(integrationID))).
		Exist(ctx)
	th.RequireNoError(t, err)

	return linked
}

// harnessDefinition returns one definition registered on the harness runtime
func harnessDefinition(t *testing.T, id string) integrationtypes.Definition {
	t.Helper()

	def, ok := suite.IntegrationsRT.Registry().Definition(id)
	assert.Assert(t, ok, "definition %s must be registered on the harness runtime", id)

	return def
}

// ensureInstallation installs the definition for the organization and removes it when the subtest ends
func ensureInstallation(t *testing.T, ctx context.Context, user th.TestUserDetails, def integrationtypes.Definition) *ent.Integration {
	t.Helper()

	install, created, err := suite.IntegrationsRT.EnsureInstallation(ctx, user.OrganizationID, "", def)
	th.RequireNoError(t, err)
	assert.Check(t, created)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: install.ID}).MustDelete(user.UserCtx, t)
	})

	return install
}

// TestCatalogAdoptionFromImportedVendors verifies adoption, refresh, and isolation of org copies
func TestCatalogAdoptionFromImportedVendors(t *testing.T) {
	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.CatalogListeners())
	assert.NilError(t, err)

	t.Cleanup(setup.Teardown)

	systemCtx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)
	systemVendorType := systemVendorEntityType(t)

	suffix := ulids.New().String()

	rows, _ := importFixtureVendors(t, systemCtx, fixturePrefixDefault, fixtureAlpha, fixtureBeta)
	catalog := rows[fixtureAlpha.ExternalID]

	org1Ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)
	org2Ctx := th.SetContext(th.SharedTestUser2.UserCtx, suite.Client.DB)
	adminOrgCtx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)

	org1VendorTypeID := orgVendorEntityTypeID(t, org1Ctx, th.SharedTestUser1.OrganizationID)

	resp, err := suite.Client.API.AdoptEntity(th.SharedTestUser1.UserCtx, catalog.ID)
	th.RequireNoError(t, err)

	adopted := resp.AdoptEntity.Entity

	t.Cleanup(func() {
		deleteEntities(t, th.SharedTestUser1.UserCtx, adopted.ID)
	})

	assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, lo.FromPtr(adopted.OwnerID)))
	assert.Assert(t, adopted.CatalogEntity != nil)
	assert.Check(t, is.Equal(catalog.ID, adopted.CatalogEntity.ID))
	assert.Check(t, is.Equal(catalog.ID, lo.FromPtr(adopted.CatalogEntityID)))
	assert.Assert(t, adopted.EntityType != nil)
	assert.Check(t, is.Equal(vendorEntityTypeName, adopted.EntityType.Name))
	assert.Check(t, is.Equal(org1VendorTypeID, adopted.EntityType.ID), "the copy must use the organization's own vendor entity type")
	assert.Check(t, systemVendorType.ID != adopted.EntityType.ID, "the copy must not point at the system-owned vendor entity type")
	assert.Check(t, is.Equal(fixtureAlpha.Name, lo.FromPtr(adopted.Name)))
	assert.Check(t, is.Equal(fixtureAlpha.DisplayName, lo.FromPtr(adopted.DisplayName)))
	assert.Check(t, is.Equal(fixtureAlpha.Description, lo.FromPtr(adopted.Description)))
	assert.Check(t, is.DeepEqual([]string{fixtureAlpha.Domain}, adopted.Domains))
	assert.Check(t, is.Equal(fixtureAlpha.LogoRemoteURL(), lo.FromPtr(adopted.LogoRemoteURL)))

	copy1, err := suite.Client.DB.Entity.Get(org1Ctx, adopted.ID)
	th.RequireNoError(t, err)
	assert.Check(t, !copy1.SystemOwned, "an adopted copy is organization-owned")
	assert.Check(t, copy1.ApprovedForUse, "an adopted vendor is approved for use")
	assert.Check(t, is.Equal(catalog.ID, copy1.CatalogEntityID))

	again, err := suite.Client.API.AdoptEntity(th.SharedTestUser1.UserCtx, catalog.ID)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(adopted.ID, again.AdoptEntity.Entity.ID), "adopting again must return the same copy")

	other, err := suite.Client.API.AdoptEntity(th.SharedTestUser2.UserCtx, catalog.ID)
	th.RequireNoError(t, err)

	otherID := other.AdoptEntity.Entity.ID

	t.Cleanup(func() {
		deleteEntities(t, th.SharedTestUser2.UserCtx, otherID)
	})

	assert.Check(t, adopted.ID != otherID, "a second organization gets its own copy")
	assert.Check(t, is.Equal(th.SharedTestUser2.OrganizationID, lo.FromPtr(other.AdoptEntity.Entity.OwnerID)))
	assert.Check(t, is.Equal(catalog.ID, other.AdoptEntity.Entity.CatalogEntity.ID))

	adminResp, err := suite.Client.API.AdoptEntity(th.SharedSystemAdminUser.UserCtx, catalog.ID)
	th.RequireNoError(t, err)

	adminCopyID := adminResp.AdoptEntity.Entity.ID

	t.Cleanup(func() {
		deleteEntities(t, th.SharedSystemAdminUser.UserCtx, adminCopyID)
	})

	adminCopy, err := suite.Client.DB.Entity.Get(adminOrgCtx, adminCopyID)
	th.RequireNoError(t, err)
	assert.Check(t, !adminCopy.SystemOwned, "an explicit owner makes the admin's copy organization-owned")
	assert.Check(t, is.Equal(th.SharedSystemAdminUser.OrganizationID, adminCopy.OwnerID))
	assert.Check(t, is.Equal(catalog.ID, adminCopy.CatalogEntityID))

	t.Run("catalogue changes refresh every adopted copy", func(t *testing.T) {
		result := runSystemImport(t, systemCtx, fixturePrefixUpdated)
		assert.Check(t, is.Equal(0, result.Failed))
		assert.Check(t, is.Equal(2, result.Changed), "both rewritten catalogue rows must count as changed")

		refreshed := func(ctx context.Context, id string) bool {
			row, err := suite.Client.DB.Entity.Get(ctx, id)

			return err == nil && row.Description == fixtureAlphaUpdated.Description && row.DisplayName == fixtureAlphaUpdated.DisplayName
		}

		waitForCondition(t, func() bool {
			return refreshed(org1Ctx, adopted.ID) && refreshed(org2Ctx, otherID)
		}, "adopted copies should receive the refreshed catalogue fields")

		waitForGala(t, setup.Runtime)

		copy1, err := suite.Client.DB.Entity.Get(org1Ctx, adopted.ID)
		th.RequireNoError(t, err)
		assert.Check(t, is.Equal(org1VendorTypeID, copy1.EntityTypeID), "the refresh must not touch the copy's entity type")
		assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, copy1.OwnerID))
		assert.Check(t, copy1.ApprovedForUse)
		assert.Check(t, !copy1.SystemOwned)
		assert.Check(t, is.Equal(catalog.ID, copy1.CatalogEntityID))
	})

	t.Run("an organization edit never fans out", func(t *testing.T) {
		th.RequireNoError(t, suite.Client.DB.Entity.UpdateOneID(adopted.ID).SetDisplayName("Org One "+suffix).Exec(org1Ctx))

		waitForGala(t, setup.Runtime)

		systemRow, err := suite.Client.DB.Entity.Get(systemCtx, catalog.ID)
		th.RequireNoError(t, err)
		assert.Check(t, is.Equal(fixtureAlphaUpdated.DisplayName, systemRow.DisplayName), "the catalogue row must not change from an organization edit")

		copy2, err := suite.Client.DB.Entity.Get(org2Ctx, otherID)
		th.RequireNoError(t, err)
		assert.Check(t, is.Equal(fixtureAlphaUpdated.DisplayName, copy2.DisplayName), "another organization's copy must not change")
	})
}

// TestCatalogCreateVendorOnInstallation verifies each createVendor branch on installation
func TestCatalogCreateVendorOnInstallation(t *testing.T) {
	systemCtx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)
	systemVendorEntityType(t)

	rows, _ := importFixtureVendors(t, systemCtx, fixturePrefixOpenlane, fixtureOpenlane)
	catalog := rows[fixtureOpenlane.ExternalID]

	testDef := harnessDefinition(t, testint.DefinitionID.ID())
	mockDef := harnessDefinition(t, testint.MockHTTPDefinitionID.ID())
	assert.Assert(t, is.Equal(fixtureOpenlane.Name, testDef.Family), "the fixture vendor must be named after the harness definition family")
	assert.Assert(t, is.Equal(fixtureOpenlane.Name, mockDef.Family), "the fixture vendor must be named after the mock definition family")

	t.Run("existing org row is linked without adoption", func(t *testing.T) {
		user := suite.UserBuilder(context.Background(), t)
		ctx := th.SetContext(user.UserCtx, suite.Client.DB)
		vendorTypeID := orgVendorEntityTypeID(t, ctx, user.OrganizationID)

		existing := (&th.EntityBuilder{Client: suite.Client, Name: fixtureOpenlane.Name, TypeID: vendorTypeID}).MustNew(user.UserCtx, t)

		t.Cleanup(func() {
			deleteEntities(t, user.UserCtx, existing.ID)
		})

		install := ensureInstallation(t, ctx, user, testDef)

		rows := orgEntitiesNamed(t, ctx, user.OrganizationID, fixtureOpenlane.Name)
		assert.Assert(t, is.Len(rows, 1), "no new row may be created when the org already has the vendor")
		assert.Check(t, is.Equal(existing.ID, rows[0].ID))
		assert.Check(t, is.Equal("", rows[0].CatalogEntityID), "an existing org row is linked, not adopted")
		assert.Check(t, entityHasIntegration(t, ctx, existing.ID, install.ID), "the existing row must carry the integration edge")
	})

	t.Run("catalogue vendor is adopted for the installing org", func(t *testing.T) {
		user := suite.UserBuilder(context.Background(), t)
		ctx := th.SetContext(user.UserCtx, suite.Client.DB)
		vendorTypeID := orgVendorEntityTypeID(t, ctx, user.OrganizationID)

		install := ensureInstallation(t, ctx, user, mockDef)

		rows := orgEntitiesNamed(t, ctx, user.OrganizationID, fixtureOpenlane.Name)
		assert.Assert(t, is.Len(rows, 1), "exactly one adopted row must exist")

		t.Cleanup(func() {
			deleteEntities(t, user.UserCtx, rows[0].ID)
		})

		assert.Check(t, is.Equal(catalog.ID, rows[0].CatalogEntityID), "the org row must point at the catalogue row")
		assert.Check(t, is.Equal(user.OrganizationID, rows[0].OwnerID))
		assert.Check(t, !rows[0].SystemOwned)
		assert.Check(t, is.Equal(vendorTypeID, rows[0].EntityTypeID), "the adopted row uses the org's own vendor entity type")
		assert.Check(t, rows[0].ApprovedForUse)
		assert.Check(t, is.Equal(fixtureOpenlane.DisplayName, rows[0].DisplayName))
		assert.Check(t, entityHasIntegration(t, ctx, rows[0].ID, install.ID), "the adopted row must carry the integration edge")
	})

	t.Run("family without a catalogue row creates a plain vendor", func(t *testing.T) {
		deleteSystemVendors(t, fixtureOpenlane)

		user := suite.UserBuilder(context.Background(), t)
		ctx := th.SetContext(user.UserCtx, suite.Client.DB)
		vendorTypeID := orgVendorEntityTypeID(t, ctx, user.OrganizationID)

		install := ensureInstallation(t, ctx, user, testDef)

		rows := orgEntitiesNamed(t, ctx, user.OrganizationID, fixtureOpenlane.Name)
		assert.Assert(t, is.Len(rows, 1), "exactly one plain vendor must be created")

		t.Cleanup(func() {
			deleteEntities(t, user.UserCtx, rows[0].ID)
		})

		assert.Check(t, is.Equal("", rows[0].CatalogEntityID), "a family with no catalogue row must not be adopted")
		assert.Check(t, is.Equal(user.OrganizationID, rows[0].OwnerID))
		assert.Check(t, !rows[0].SystemOwned)
		assert.Check(t, is.Equal(vendorTypeID, rows[0].EntityTypeID))
		assert.Check(t, is.Contains(rows[0].Tags, "integration"))
		assert.Check(t, entityHasIntegration(t, ctx, rows[0].ID, install.ID))
	})
}

// TestCatalogQuestionnaireTransformAdopts verifies questionnaire entities adopt matching catalogue vendors
func TestCatalogQuestionnaireTransformAdopts(t *testing.T) {
	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.QuestionnaireTransformListeners())
	assert.NilError(t, err)

	t.Cleanup(setup.Teardown)

	systemCtx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)
	systemVendorEntityType(t)

	suffix := ulids.New().String()

	rows, _ := importFixtureVendors(t, systemCtx, fixturePrefixDefault, fixtureAlpha, fixtureBeta)
	catalog := rows[fixtureAlpha.ExternalID]

	user := suite.UserBuilder(context.Background(), t)
	orgID := user.OrganizationID
	allowCtx := privacy.DecisionContext(th.SetContext(user.UserCtx, suite.Client.DB), privacy.Allow)
	vendorTypeID := orgVendorEntityTypeID(t, allowCtx, orgID)

	template := (&th.TemplateBuilder{Client: suite.Client}).MustNew(user.UserCtx, t)
	assert.NilError(t, suite.Client.DB.Template.UpdateOneID(template.ID).SetTransformConfiguration(models.TemplateProjectionConfig{
		Enabled: true,
		Mappings: []models.TemplateProjectionFieldMapping{
			{From: "vendorName", To: "name"},
			{From: "vendorDomains", To: "domains"},
			{From: "vendorExternalID", To: "external_id"},
			{From: "vendorTypeID", To: "entity_type_id"},
		},
	}).Exec(allowCtx))

	assessment := (&th.AssessmentBuilder{Client: suite.Client, TemplateID: template.ID}).MustNew(user.UserCtx, t)

	var (
		responseIDs []string
		entityIDs   []string
	)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.AssessmentResponseDeleteOne]{Client: suite.Client.DB.AssessmentResponse, IDs: responseIDs}).MustDelete(user.UserCtx, t)
		deleteEntities(t, user.UserCtx, entityIDs...)
		(&th.Cleanup[*ent.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment.ID}).MustDelete(user.UserCtx, t)
		(&th.Cleanup[*ent.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: template.ID}).MustDelete(user.UserCtx, t)
	})

	completeResponse := func(t *testing.T, data map[string]any) *ent.AssessmentResponse {
		t.Helper()

		doc, err := suite.Client.DB.DocumentData.Create().
			SetOwnerID(orgID).
			SetTemplateID(template.ID).
			SetData(data).
			Save(allowCtx)
		assert.NilError(t, err)

		response := (&th.AssessmentResponseBuilder{Client: suite.Client, AssessmentID: assessment.ID, OwnerID: orgID}).MustNew(user.UserCtx, t)
		responseIDs = append(responseIDs, response.ID)

		assert.NilError(t, suite.Client.DB.AssessmentResponse.UpdateOneID(response.ID).
			SetDocumentDataID(doc.ID).
			SetStatus(enums.AssessmentResponseStatusCompleted).
			Exec(allowCtx))

		waitForCondition(t, func() bool {
			updated, err := suite.Client.DB.AssessmentResponse.Get(allowCtx, response.ID)

			return err == nil && updated.EntityID != ""
		}, "assessment response should link the transformed entity")

		waitForGala(t, setup.Runtime)

		updated, err := suite.Client.DB.AssessmentResponse.Get(allowCtx, response.ID)
		assert.NilError(t, err)

		return updated
	}

	externalID := "questionnaire-external-" + suffix

	var adoptedID string

	t.Run("matching response adopts the catalogue vendor", func(t *testing.T) {
		response := completeResponse(t, map[string]any{
			"vendorName":       fixtureAlpha.Name,
			"vendorDomains":    []string{fixtureAlpha.Domain},
			"vendorExternalID": externalID,
			"vendorTypeID":     vendorTypeID,
		})

		adoptedID = response.EntityID
		entityIDs = append(entityIDs, adoptedID)

		row, err := suite.Client.DB.Entity.Get(allowCtx, adoptedID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(catalog.ID, row.CatalogEntityID), "the transformed entity must be adopted from the catalogue")
		assert.Check(t, is.Equal(externalID, row.ExternalID), "the response's external id must land on the adopted row")
		assert.Check(t, is.Equal(vendorTypeID, row.EntityTypeID), "the adopted row uses the org's vendor entity type")
		assert.Check(t, is.Equal(orgID, row.OwnerID))
		assert.Check(t, !row.SystemOwned)
		assert.Check(t, row.ApprovedForUse)

		count, err := suite.Client.DB.Entity.Query().Where(entity.OwnerID(orgID), entity.CatalogEntityID(catalog.ID)).Count(allowCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, count))
	})

	t.Run("repeat response with the same external id updates the adopted row", func(t *testing.T) {
		response := completeResponse(t, map[string]any{
			"vendorName":       fixtureAlpha.Name,
			"vendorDomains":    []string{fixtureAlpha.Domain},
			"vendorExternalID": externalID,
			"vendorTypeID":     vendorTypeID,
		})

		assert.Check(t, is.Equal(adoptedID, response.EntityID), "the second response must link the same row")

		count, err := suite.Client.DB.Entity.Query().Where(entity.OwnerID(orgID), entity.ExternalID(externalID)).Count(allowCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, count), "no second row may be created for the same external id")
	})

	t.Run("unmatched response creates a plain org row", func(t *testing.T) {
		plainName := "plain-" + suffix

		response := completeResponse(t, map[string]any{
			"vendorName":       plainName,
			"vendorDomains":    []string{plainName + ".example.com"},
			"vendorExternalID": "questionnaire-plain-" + suffix,
			"vendorTypeID":     vendorTypeID,
		})

		entityIDs = append(entityIDs, response.EntityID)

		row, err := suite.Client.DB.Entity.Get(allowCtx, response.EntityID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("", row.CatalogEntityID), "an unmatched vendor must not be adopted")
		assert.Check(t, is.Equal(plainName, row.Name))
		assert.Check(t, is.Equal(orgID, row.OwnerID))
		assert.Check(t, is.Equal(vendorTypeID, row.EntityTypeID))
		assert.Check(t, !row.SystemOwned)
	})
}

// TestCatalogDomainScanImportAdopts verifies domain scan vendors adopt matching catalogue vendors
func TestCatalogDomainScanImportAdopts(t *testing.T) {
	systemCtx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)
	systemVendorEntityType(t)

	suffix := ulids.New().String()
	assetDomain := "app." + fixtureScan.Domain

	rows, _ := importFixtureVendors(t, systemCtx, fixturePrefixScan, fixtureScan)
	catalog := rows[fixtureScan.ExternalID]

	user := suite.UserBuilder(context.Background(), t)
	orgID := user.OrganizationID
	orgCtx := th.SetContext(user.UserCtx, suite.Client.DB)
	vendorTypeID := orgVendorEntityTypeID(t, orgCtx, orgID)

	metadata, err := jsonx.ToMap(domainscan.ScanReport{
		Assets: &domainscan.Assets{DNSRecords: []domainscan.DNSRecord{{Domain: assetDomain, Type: "A", Vendor: fixtureScan.Name}}},
	})
	th.RequireNoError(t, err)

	scanRow, err := suite.Client.DB.Scan.Create().
		SetTarget(fixtureScan.Domain).
		SetOwnerID(orgID).
		SetMetadata(metadata).
		Save(orgCtx)
	th.RequireNoError(t, err)

	var entityIDs []string

	t.Cleanup(func() {
		deleteEntities(t, user.UserCtx, entityIDs...)

		assets, err := suite.Client.DB.Asset.Query().Where(asset.OwnerID(orgID)).IDs(orgCtx)
		th.RequireNoError(t, err)

		if len(assets) > 0 {
			(&th.Cleanup[*ent.AssetDeleteOne]{Client: suite.Client.DB.Asset, IDs: assets}).MustDelete(user.UserCtx, t)
		}

		(&th.Cleanup[*ent.ScanDeleteOne]{Client: suite.Client.DB.Scan, ID: scanRow.ID}).MustDelete(user.UserCtx, t)
	})

	envelope := cloudflare.DomainScanImport{
		OrganizationID: orgID,
		ScanIDs:        []string{scanRow.ID},
		Vendors: []cloudflare.DomainScanImportVendor{
			{Ref: "vendor-1", Name: fixtureScan.Name, Domain: fixtureScan.Domain, Categories: []string{"saas"}},
		},
		Assets: []cloudflare.DomainScanImportAsset{
			{Ref: "asset-1", Name: "asset-" + suffix, Identifier: assetDomain, Website: "https://" + assetDomain},
		},
	}

	runImport := func(t *testing.T, envelope cloudflare.DomainScanImport) {
		t.Helper()

		raw, err := json.Marshal(envelope)
		th.RequireNoError(t, err)

		_, err = suite.IntegrationsRT.ExecuteRuntimeOperation(orgCtx, cloudflare.DefinitionID.ID(), cloudflare.DomainScanImportOp.Name(), raw)
		th.RequireNoError(t, err)
	}

	adoptedRows := func(t *testing.T) []*ent.Entity {
		t.Helper()

		rows, err := suite.Client.DB.Entity.Query().
			Where(entity.OwnerID(orgID), entity.CatalogEntityID(catalog.ID)).
			All(orgCtx)
		th.RequireNoError(t, err)

		return rows
	}

	t.Run("scanned vendor matching a catalogue domain is adopted and linked", func(t *testing.T) {
		runImport(t, envelope)

		rows := adoptedRows(t)
		assert.Assert(t, is.Len(rows, 1), "the scanned vendor must be adopted from the catalogue")

		entityIDs = append(entityIDs, rows[0].ID)

		assert.Check(t, is.Equal(orgID, rows[0].OwnerID))
		assert.Check(t, !rows[0].SystemOwned)
		assert.Check(t, is.Equal(vendorTypeID, rows[0].EntityTypeID))
		assert.Check(t, is.Contains(rows[0].Tags, "saas"))

		scanIDs, err := rows[0].QueryScans().Where(scan.ID(scanRow.ID)).IDs(orgCtx)
		th.RequireNoError(t, err)
		assert.Check(t, is.Len(scanIDs, 1), "the adopted row must link the envelope's scan")

		assetIDs, err := rows[0].QueryAssets().Where(asset.Name("asset-" + suffix)).IDs(orgCtx)
		th.RequireNoError(t, err)
		assert.Check(t, is.Len(assetIDs, 1), "the adopted row must link the asset the scan attributed to the vendor")
	})

	t.Run("re-running the envelope is idempotent", func(t *testing.T) {
		runImport(t, envelope)

		rows := adoptedRows(t)
		assert.Assert(t, is.Len(rows, 1), "a resubmission must not adopt a second copy")

		scanIDs, err := rows[0].QueryScans().IDs(orgCtx)
		th.RequireNoError(t, err)
		assert.Check(t, is.Len(scanIDs, 1), "the scan link must not be duplicated")

		assetIDs, err := rows[0].QueryAssets().IDs(orgCtx)
		th.RequireNoError(t, err)
		assert.Check(t, is.Len(assetIDs, 1), "the asset link must not be duplicated")
	})

	t.Run("a vendor named like an existing org row reuses it", func(t *testing.T) {
		existingName := "existing-" + suffix
		existing := (&th.EntityBuilder{Client: suite.Client, Name: existingName, TypeID: vendorTypeID}).MustNew(user.UserCtx, t)
		entityIDs = append(entityIDs, existing.ID)

		runImport(t, cloudflare.DomainScanImport{
			OrganizationID: orgID,
			ScanIDs:        []string{scanRow.ID},
			Vendors:        []cloudflare.DomainScanImportVendor{{Ref: "vendor-2", Name: existingName, Domain: fixtureScan.Domain}},
			Assets:         []cloudflare.DomainScanImportAsset{},
		})

		rows := orgEntitiesNamed(t, orgCtx, orgID, existingName)
		assert.Assert(t, is.Len(rows, 1), "the existing row must be reused, not duplicated")
		assert.Check(t, is.Equal(existing.ID, rows[0].ID))
		assert.Check(t, is.Equal("", rows[0].CatalogEntityID), "a reused org row is not adopted")
		assert.Check(t, is.Len(adoptedRows(t), 1), "no additional catalogue copy may be adopted")
	})
}
