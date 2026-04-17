package graphapi_test

import (
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/internal/graphapi"
)

func TestIdentityResolution(t *testing.T) {
	ctx := setContext(testUser1.UserCtx, suite.client.db)

	irSetup, err := graphapi.SetupIdentityResolution(ctx, suite.client.db, suite.tf.URI)
	assert.NilError(t, err)
	defer irSetup.Teardown()

	t.Run("single create with primary source", func(t *testing.T) {
		email := "single-create@testresolution.io"

		da := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			DisplayName:    "Single Create User",
			GivenName:      lo.ToPtr("Single"),
			FamilyName:     lo.ToPtr("Create"),
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			JobTitle:       lo.ToPtr("Engineer"),
			Department:     lo.ToPtr("Platform"),
			PhoneNumber:    lo.ToPtr("800-867-5309"),
			EmailAliases:   []string{"single-create@mail.testresolution.io"},
			OwnerID:        testUser1.OrganizationID,
			ExternalID:     "1234871001",
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
		assert.NilError(t, err)
		assert.Assert(t, linked.IdentityHolderID != nil)

		holder, err := suite.client.db.IdentityHolder.Get(ctx, *linked.IdentityHolderID)
		assert.NilError(t, err)

		assert.Check(t, is.Equal(email, holder.Email))
		assert.Check(t, is.Equal("Single Create User", holder.FullName))
		assert.Check(t, is.Equal(enums.UserStatusActive, holder.Status))
		assert.Check(t, holder.IsActive)
		assert.Check(t, is.Equal("Engineer", holder.Title))
		assert.Check(t, is.Equal("Platform", holder.Department))
		assert.Check(t, is.Equal("800-867-5309", holder.PhoneNumber))
		assert.Check(t, is.DeepEqual([]string{"single-create@mail.testresolution.io"}, holder.EmailAliases))
		assert.Check(t, is.Equal("1234871001", holder.ExternalUserID))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holder.ID}).MustDelete(ctx, t)
	})

	t.Run("cross directory email match", func(t *testing.T) {
		email := "cross-dir@testresolution.io"

		da1 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			DisplayName:    "Cross Dir User",
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			JobTitle:       lo.ToPtr("SRE"),
			Department:     lo.ToPtr("Infra"),
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked1, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da1.ID)
		assert.NilError(t, err)
		assert.Assert(t, linked1.IdentityHolderID != nil)

		holderID := *linked1.IdentityHolderID

		da2 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			DisplayName:    "Cross Dir User (GitHub)",
			DirectoryName:  lo.ToPtr("github"),
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
			PhoneNumber:    lo.ToPtr("800-867-5309"),
			EmailAliases:   []string{"single-create@mail.testresolution.io"},
			ExternalID:     "1234871001",
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked2, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da2.ID)
		assert.NilError(t, err)
		assert.Assert(t, linked2.IdentityHolderID != nil)
		assert.Check(t, is.Equal(holderID, *linked2.IdentityHolderID))

		holder, err := suite.client.db.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)

		// There should be one record from the secondary directory
		assert.Check(t, is.Len(holder.EmailAliases, 1))

		// ensure phone number, external id are not set from the non-primary record
		assert.Check(t, holder.PhoneNumber == "")
		assert.Check(t, holder.ExternalUserID == da1.ExternalID)

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, IDs: []string{da1.ID, da2.ID}}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holderID}).MustDelete(ctx, t)
	})

	t.Run("cross directory name match different emails", func(t *testing.T) {
		primaryEmail := "primary-alias@testresolution.io"
		secondaryEmail := "secondary-alias@testresolution.io"

		da1 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &primaryEmail,
			GivenName:      lo.ToPtr("Alias"),
			FamilyName:     lo.ToPtr("Tester"),
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked1, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da1.ID)
		assert.NilError(t, err)

		holderID := *linked1.IdentityHolderID

		da2 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &secondaryEmail,
			GivenName:      lo.ToPtr("Alias"),
			FamilyName:     lo.ToPtr("Tester"),
			DirectoryName:  lo.ToPtr("github"),
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked2, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da2.ID)
		assert.NilError(t, err)
		assert.Assert(t, linked2.IdentityHolderID != nil)
		assert.Check(t, is.Equal(holderID, *linked2.IdentityHolderID))

		holder, err := suite.client.db.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)
		assert.Check(t, is.Contains(holder.EmailAliases, secondaryEmail))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, IDs: []string{da1.ID, da2.ID}}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holderID}).MustDelete(ctx, t)
	})

	t.Run("primary source enrichment overrides non-primary", func(t *testing.T) {
		email := "primary-enrichment@testresolution.io"

		da1 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			DisplayName:    "Enrichment User",
			DirectoryName:  lo.ToPtr("slack"),
			PrimarySource:  false,
			Status:         enums.DirectoryAccountStatusActive,
			JobTitle:       lo.ToPtr("Slack Title"),
			Department:     lo.ToPtr("Slack Dept"),
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked1, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da1.ID)
		assert.NilError(t, err)

		holderID := *linked1.IdentityHolderID

		holder, err := suite.client.db.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.UserStatusUnknown, holder.Status))

		da2 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			DisplayName:    "Enrichment User (Primary)",
			GivenName:      lo.ToPtr("Enrichment"),
			FamilyName:     lo.ToPtr("User"),
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusSuspended,
			JobTitle:       lo.ToPtr("Staff Engineer"),
			Department:     lo.ToPtr("Platform"),
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked2, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da2.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(holderID, *linked2.IdentityHolderID))

		holder, err = suite.client.db.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)

		assert.Check(t, is.Equal(enums.UserStatusSuspended, holder.Status))
		assert.Check(t, !holder.IsActive)
		assert.Check(t, is.Equal("Staff Engineer", holder.Title))
		assert.Check(t, is.Equal("Platform", holder.Department))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, IDs: []string{da1.ID, da2.ID}}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holderID}).MustDelete(ctx, t)
	})

	t.Run("status mapping", func(t *testing.T) {
		testCases := []struct {
			name           string
			dirStatus      enums.DirectoryAccountStatus
			expectedStatus enums.UserStatus
			expectedActive bool
		}{
			{"active maps to active", enums.DirectoryAccountStatusActive, enums.UserStatusActive, true},
			{"inactive maps to inactive", enums.DirectoryAccountStatusInactive, enums.UserStatusInactive, false},
			{"suspended maps to suspended", enums.DirectoryAccountStatusSuspended, enums.UserStatusSuspended, false},
			{"deleted maps to deactivated", enums.DirectoryAccountStatusDeleted, enums.UserStatusDeactivated, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				email := "status-" + tc.dirStatus.String() + "@testresolution.io"

				da := (&DirectoryAccountBuilder{
					client:         suite.client,
					CanonicalEmail: &email,
					DisplayName:    "Status " + tc.dirStatus.String(),
					DirectoryName:  lo.ToPtr("googleworkspace"),
					PrimarySource:  true,
					Status:         tc.dirStatus,
					OwnerID:        testUser1.OrganizationID,
				}).MustNew(ctx, t)

				irSetup.Runtime.WaitIdle()

				linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
				assert.NilError(t, err)
				assert.Assert(t, linked.IdentityHolderID != nil)

				holder, err := suite.client.db.IdentityHolder.Get(ctx, *linked.IdentityHolderID)
				assert.NilError(t, err)

				assert.Check(t, is.Equal(tc.expectedStatus, holder.Status))
				assert.Check(t, is.Equal(tc.expectedActive, holder.IsActive))

				(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
				(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holder.ID}).MustDelete(ctx, t)
			})
		}
	})

	t.Run("no email no holder", func(t *testing.T) {
		da := (&DirectoryAccountBuilder{
			client:        suite.client,
			DisplayName:   "No Email User",
			DirectoryName: lo.ToPtr("github"),
			Status:        enums.DirectoryAccountStatusActive,
			OwnerID:       testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		account, err := suite.client.db.DirectoryAccount.Get(ctx, da.ID)
		assert.NilError(t, err)
		assert.Check(t, account.IdentityHolderID == nil || *account.IdentityHolderID == "")

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
	})

	t.Run("delete directory account syncs aliases", func(t *testing.T) {
		primaryEmail := "delete-sync-primary@testresolution.io"
		secondaryEmail := "delete-sync-secondary@testresolution.io"

		da1 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &primaryEmail,
			GivenName:      lo.ToPtr("Delete"),
			FamilyName:     lo.ToPtr("Sync"),
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked1, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da1.ID)
		assert.NilError(t, err)

		holderID := *linked1.IdentityHolderID

		da2 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &secondaryEmail,
			GivenName:      lo.ToPtr("Delete"),
			FamilyName:     lo.ToPtr("Sync"),
			DirectoryName:  lo.ToPtr("github"),
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		_, err = graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da2.ID)
		assert.NilError(t, err)

		// wait for the handler to fully complete (syncEmailAliases runs after identity_holder_id is set)
		irSetup.Runtime.WaitIdle()

		holder, err := suite.client.db.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)
		assert.Check(t, is.Contains(holder.EmailAliases, secondaryEmail))

		// hard delete the secondary directory account — triggers alias sync via hook
		err = suite.client.db.DirectoryAccount.DeleteOneID(da2.ID).Exec(ctx)
		assert.NilError(t, err)

		holder, err = suite.client.db.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)

		for _, alias := range holder.EmailAliases {
			assert.Check(t, alias != secondaryEmail, "secondary email should have been removed from aliases")
		}

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da1.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holderID}).MustDelete(ctx, t)
	})

	t.Run("soft delete holder clears directory account links", func(t *testing.T) {
		email1 := "holder-delete-link1@testresolution.io"
		email2 := "holder-delete-link2@testresolution.io"

		da1 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email1,
			GivenName:      lo.ToPtr("HolderDel"),
			FamilyName:     lo.ToPtr("Test"),
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked1, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da1.ID)
		assert.NilError(t, err)

		holderID := *linked1.IdentityHolderID

		da2 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email2,
			GivenName:      lo.ToPtr("HolderDel"),
			FamilyName:     lo.ToPtr("Test"),
			DirectoryName:  lo.ToPtr("github"),
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked2, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da2.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(holderID, *linked2.IdentityHolderID))

		// soft-delete the identity holder
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holderID}).MustDelete(ctx, t)

		acct1, err := suite.client.db.DirectoryAccount.Get(ctx, da1.ID)
		assert.NilError(t, err)
		assert.Check(t, acct1.IdentityHolderID == nil || *acct1.IdentityHolderID == "", "da1 identity_holder_id should be cleared after holder soft-delete")

		acct2, err := suite.client.db.DirectoryAccount.Get(ctx, da2.ID)
		assert.NilError(t, err)
		assert.Check(t, acct2.IdentityHolderID == nil || *acct2.IdentityHolderID == "", "da2 identity_holder_id should be cleared after holder soft-delete")

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, IDs: []string{da1.ID, da2.ID}}).MustDelete(ctx, t)
	})

	t.Run("update re-enriches from primary source", func(t *testing.T) {
		email := "update-enrich@testresolution.io"

		da := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			DisplayName:    "Update Enrich",
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			JobTitle:       lo.ToPtr("Engineer"),
			Department:     lo.ToPtr("Platform"),
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
		assert.NilError(t, err)

		holderID := *linked.IdentityHolderID

		err = suite.client.db.DirectoryAccount.UpdateOneID(da.ID).
			SetJobTitle("Staff Engineer").
			SetDepartment("SRE").
			SetStatus(enums.DirectoryAccountStatusSuspended).
			Exec(ctx)
		assert.NilError(t, err)

		irSetup.Runtime.WaitIdle()

		holder, err := suite.client.db.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)

		assert.Check(t, is.Equal(enums.UserStatusSuspended, holder.Status))
		assert.Check(t, !holder.IsActive)
		assert.Check(t, is.Equal("Staff Engineer", holder.Title))
		assert.Check(t, is.Equal("SRE", holder.Department))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holderID}).MustDelete(ctx, t)
	})

	t.Run("update unlinked account triggers resolution", func(t *testing.T) {
		da := (&DirectoryAccountBuilder{
			client:        suite.client,
			DisplayName:   "Unlinked Update",
			DirectoryName: lo.ToPtr("github"),
			Status:        enums.DirectoryAccountStatusActive,
			OwnerID:       testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		account, err := suite.client.db.DirectoryAccount.Get(ctx, da.ID)
		assert.NilError(t, err)
		assert.Check(t, account.IdentityHolderID == nil || *account.IdentityHolderID == "")

		email := "unlinked-update@testresolution.io"

		err = suite.client.db.DirectoryAccount.UpdateOneID(da.ID).
			SetCanonicalEmail(email).
			Exec(ctx)
		assert.NilError(t, err)

		irSetup.Runtime.WaitIdle()

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
		assert.NilError(t, err)
		assert.Assert(t, linked.IdentityHolderID != nil)

		holderID := *linked.IdentityHolderID

		holder, err := suite.client.db.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(email, holder.Email))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holderID}).MustDelete(ctx, t)
	})

	t.Run("existing holder email match step 1", func(t *testing.T) {
		email := "preexisting-holder@testresolution.io"

		holder := (&IdentityHolderBuilder{
			client: suite.client,
			Email:  email,
		}).MustNew(ctx, t)

		da := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			DisplayName:    "Preexisting Match",
			DirectoryName:  lo.ToPtr("googleworkspace"),
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(holder.ID, *linked.IdentityHolderID))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holder.ID}).MustDelete(ctx, t)
	})

	t.Run("given name only", func(t *testing.T) {
		email := "given-only@testresolution.io"

		da := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			GivenName:      lo.ToPtr("Madonna"),
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
		assert.NilError(t, err)

		holder, err := suite.client.db.IdentityHolder.Get(ctx, *linked.IdentityHolderID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("Madonna", holder.FullName))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holder.ID}).MustDelete(ctx, t)
	})

	t.Run("family name only", func(t *testing.T) {
		email := "family-only@testresolution.io"

		da := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			FamilyName:     lo.ToPtr("McLovin"),
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
		assert.NilError(t, err)

		holder, err := suite.client.db.IdentityHolder.Get(ctx, *linked.IdentityHolderID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("McLovin", holder.FullName))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holder.ID}).MustDelete(ctx, t)
	})

	t.Run("display name takes precedence", func(t *testing.T) {
		email := "displayname-prio@testresolution.io"

		da := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			DisplayName:    "Display Wins",
			GivenName:      lo.ToPtr("Given"),
			FamilyName:     lo.ToPtr("Family"),
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
		assert.NilError(t, err)

		holder, err := suite.client.db.IdentityHolder.Get(ctx, *linked.IdentityHolderID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("Display Wins", holder.FullName))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holder.ID}).MustDelete(ctx, t)
	})

	t.Run("multiple directory sources three dirs", func(t *testing.T) {
		primaryEmail := "multi-source@testresolution.io"
		slackEmail := "multi-source-slack@testresolution.io"
		githubEmail := "multi-source-github@testresolution.io"

		da1 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &primaryEmail,
			DisplayName:    "Multi Source User",
			GivenName:      lo.ToPtr("Multi"),
			FamilyName:     lo.ToPtr("Source"),
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			JobTitle:       lo.ToPtr("Principal Engineer"),
			Department:     lo.ToPtr("Engineering"),
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked1, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da1.ID)
		assert.NilError(t, err)

		holderID := *linked1.IdentityHolderID

		da2 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &slackEmail,
			GivenName:      lo.ToPtr("Multi"),
			FamilyName:     lo.ToPtr("Source"),
			DirectoryName:  lo.ToPtr("slack"),
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked2, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da2.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(holderID, *linked2.IdentityHolderID))

		da3 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &githubEmail,
			GivenName:      lo.ToPtr("Multi"),
			FamilyName:     lo.ToPtr("Source"),
			DirectoryName:  lo.ToPtr("github"),
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked3, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da3.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(holderID, *linked3.IdentityHolderID))

		accounts, err := suite.client.db.DirectoryAccount.Query().
			Where(directoryaccount.IdentityHolderID(holderID)).
			All(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(accounts, 3))

		holder, err := suite.client.db.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)
		assert.Check(t, is.Contains(holder.EmailAliases, slackEmail))
		assert.Check(t, is.Contains(holder.EmailAliases, githubEmail))

		for _, alias := range holder.EmailAliases {
			assert.Check(t, alias != primaryEmail, "primary email should not appear in aliases")
		}

		assert.Check(t, is.Equal("Principal Engineer", holder.Title))
		assert.Check(t, is.Equal("Engineering", holder.Department))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, IDs: []string{da1.ID, da2.ID, da3.ID}}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holderID}).MustDelete(ctx, t)
	})

	t.Run("ambiguous name match creates new holder", func(t *testing.T) {
		email1 := "ambig-name-1@testresolution.io"
		email2 := "ambig-name-2@testresolution.io"
		email3 := "ambig-name-3@testresolution.io"

		da1 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email1,
			GivenName:      lo.ToPtr("John"),
			FamilyName:     lo.ToPtr("Smith"),
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked1, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da1.ID)
		assert.NilError(t, err)

		holderID1 := *linked1.IdentityHolderID

		da2 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email2,
			GivenName:      lo.ToPtr("John"),
			FamilyName:     lo.ToPtr("Smith"),
			DirectoryName:  lo.ToPtr("slack"),
			PrimarySource:  false,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked2, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da2.ID)
		assert.NilError(t, err)

		holderID2 := *linked2.IdentityHolderID
		assert.Check(t, is.Equal(holderID1, holderID2))

		// manually re-link da2 to a different holder to simulate ambiguity
		holder2 := (&IdentityHolderBuilder{
			client: suite.client,
			Email:  email2,
		}).MustNew(ctx, t)

		err = suite.client.db.DirectoryAccount.UpdateOneID(da2.ID).SetIdentityHolderID(holder2.ID).Exec(ctx)
		assert.NilError(t, err)

		// third account with same name — two different holders exist, step 3 refuses to guess
		da3 := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email3,
			GivenName:      lo.ToPtr("John"),
			FamilyName:     lo.ToPtr("Smith"),
			DirectoryName:  lo.ToPtr("github"),
			PrimarySource:  false,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked3, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da3.ID)
		assert.NilError(t, err)

		holderID3 := *linked3.IdentityHolderID
		assert.Check(t, holderID3 != holderID1, "ambiguous name match should create a new holder")
		assert.Check(t, holderID3 != holder2.ID, "ambiguous name match should not pick either existing")

		holder3, err := suite.client.db.IdentityHolder.Get(ctx, holderID3)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(email3, holder3.Email))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, IDs: []string{da1.ID, da2.ID, da3.ID}}).MustDelete(ctx, t)

		for _, id := range []string{holderID1, holder2.ID, holderID3} {
			exists, _ := suite.client.db.IdentityHolder.Query().Where(identityholder.IDEQ(id)).Exist(ctx)
			if exists {
				(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: id}).MustDelete(ctx, t)
			}
		}
	})

	t.Run("non-primary source unknown status no metadata", func(t *testing.T) {
		email := "non-primary-status@testresolution.io"

		da := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			DisplayName:    "NonPrimary User",
			DirectoryName:  lo.ToPtr("slack"),
			PrimarySource:  false,
			Status:         enums.DirectoryAccountStatusActive,
			JobTitle:       lo.ToPtr("Should Not Propagate"),
			Department:     lo.ToPtr("Should Not Propagate"),
			OwnerID:        testUser1.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
		assert.NilError(t, err)

		holder, err := suite.client.db.IdentityHolder.Get(ctx, *linked.IdentityHolderID)
		assert.NilError(t, err)

		assert.Check(t, is.Equal(enums.UserStatusUnknown, holder.Status))
		assert.Check(t, is.Equal("", holder.Title))
		assert.Check(t, is.Equal("", holder.Department))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holder.ID}).MustDelete(ctx, t)
	})
}
