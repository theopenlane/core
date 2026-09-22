//go:build test

package eventstest_test

import (
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/objectstore"
)

// TestObjectStoreSystemImportCreatesVendors verifies the import creates system-owned vendors with provenance
func TestObjectStoreSystemImportCreatesVendors(t *testing.T) {
	systemCtx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)
	vendorType := systemVendorEntityType(t)

	rows, result := importFixtureVendors(t, systemCtx, fixturePrefixDefault, fixtureAlpha, fixtureBeta)
	assert.Check(t, is.Equal(1, result.Filtered), "the customer record must be filtered out by the vendor variant")

	for _, vendor := range []fixtureVendor{fixtureAlpha, fixtureBeta} {
		row := rows[vendor.ExternalID]

		assert.Check(t, row.SystemOwned, "the runtime import must create system-owned rows")
		assert.Check(t, is.Equal("", row.OwnerID), "a system-owned row carries no owner")
		assert.Check(t, is.Equal(vendor.ExternalID, row.ExternalID))
		assert.Check(t, is.Equal(vendor.ExternalID, lo.FromPtr(row.SystemInternalID)))
		assert.Check(t, is.Equal(vendorType.ID, row.EntityTypeID), "the vendor variant links the system-owned vendor entity type")
		assert.Check(t, is.Equal(objectstore.DefinitionID.ID(), row.SourceDefinitionID))
		assert.Check(t, is.Equal(vendor.Name, row.Name))
		assert.Check(t, is.Equal(vendor.DisplayName, row.DisplayName))
		assert.Check(t, is.Equal(vendor.Description, row.Description))
		assert.Check(t, is.DeepEqual([]string{vendor.Domain}, row.Domains))
		assert.Check(t, is.DeepEqual(fixtureTags, row.Tags))
		assert.Check(t, is.Equal(vendor.LogoRemoteURL(), lo.FromPtr(row.LogoRemoteURL)))
	}

	filtered, err := suite.Client.DB.Entity.Query().
		Where(entity.ExternalID(fixtureCustomer.ExternalID)).
		Count(systemCtx)
	th.RequireNoError(t, err)
	assert.Check(t, is.Equal(0, filtered), "a record with another entity type must not be created")
}

// TestObjectStoreSystemImportUpsertsExisting verifies the import upserts by external id
func TestObjectStoreSystemImportUpsertsExisting(t *testing.T) {
	systemCtx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)
	systemVendorEntityType(t)

	deleteSystemVendors(t, fixtureAlpha, fixtureBeta)

	t.Cleanup(func() {
		deleteSystemVendors(t, fixtureAlpha, fixtureBeta)
	})

	seeded, err := suite.Client.DB.Entity.Create().
		SetName(fixtureAlpha.Name).
		SetExternalID(fixtureAlpha.ExternalID).
		SetDescription("seeded description").
		Save(systemCtx)
	th.RequireNoError(t, err)
	assert.Assert(t, seeded.SystemOwned, "a system admin create without an owner must be system-owned")

	rowsByExternalID := func(t *testing.T) []*ent.Entity {
		t.Helper()

		rows, err := suite.Client.DB.Entity.Query().Where(entity.ExternalID(fixtureAlpha.ExternalID)).All(systemCtx)
		th.RequireNoError(t, err)

		return rows
	}

	t.Run("import updates the seeded system row", func(t *testing.T) {
		result := runSystemImport(t, systemCtx, fixturePrefixDefault)
		assert.Check(t, is.Equal(0, result.Failed))
		assert.Check(t, is.Equal(2, result.Changed), "the seeded row's takeover and the other vendor's create both count as changed")

		rows := rowsByExternalID(t)
		assert.Assert(t, is.Len(rows, 1), "the import must converge on the seeded row instead of creating a second")
		assert.Check(t, is.Equal(seeded.ID, rows[0].ID))
		assert.Check(t, is.Equal(fixtureAlpha.Description, rows[0].Description))
		assert.Check(t, is.Equal(fixtureAlpha.DisplayName, rows[0].DisplayName))
		assert.Check(t, rows[0].SystemOwned, "the row must stay system-owned after the import")
		assert.Check(t, is.Equal("", rows[0].OwnerID))
	})

	t.Run("importing the same prefix again leaves the rows unchanged", func(t *testing.T) {
		result := runSystemImport(t, systemCtx, fixturePrefixDefault)
		assert.Check(t, is.Equal(0, result.Failed))
		assert.Check(t, is.Equal(0, result.Changed), "an unchanged record must not count as changed")

		rows := rowsByExternalID(t)
		assert.Assert(t, is.Len(rows, 1))
		assert.Check(t, is.Equal(fixtureAlpha.Description, rows[0].Description))
	})

	t.Run("the rewritten prefix updates the rows in place", func(t *testing.T) {
		result := runSystemImport(t, systemCtx, fixturePrefixUpdated)
		assert.Check(t, is.Equal(0, result.Failed))
		assert.Check(t, is.Equal(2, result.Changed), "both rewritten records count as changed")

		rows := rowsByExternalID(t)
		assert.Assert(t, is.Len(rows, 1))
		assert.Check(t, is.Equal(seeded.ID, rows[0].ID))
		assert.Check(t, is.Equal(fixtureAlphaUpdated.DisplayName, rows[0].DisplayName))
		assert.Check(t, is.Equal(fixtureAlphaUpdated.Description, rows[0].Description))
		assert.Check(t, rows[0].SystemOwned)

		beta := systemVendorByExternalID(t, systemCtx, fixtureBeta.ExternalID)
		assert.Check(t, is.Equal(fixtureBetaUpdated.DisplayName, beta.DisplayName))
		assert.Check(t, is.Equal(fixtureBetaUpdated.Description, beta.Description))
	})
}
