//go:build test

package eventstest_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
)

func TestIdentityResolutionUpdateCascade(t *testing.T) {
	idUser := suite.UserBuilder(context.Background(), t)
	ctx := th.SetContext(idUser.UserCtx, suite.Client.DB)

	irSetup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.IdentityResolutionListeners())
	assert.NilError(t, err)
	defer irSetup.Teardown()

	t.Run("canonical email change keeps the linked holder", func(t *testing.T) {
		originalEmail := "keeplink-original@updatelistener.io"
		targetEmail := "keeplink-target@updatelistener.io"

		targetHolder := (&th.IdentityHolderBuilder{
			Client: suite.Client,
			Email:  targetEmail,
		}).MustNew(idUser.UserCtx, t)

		da := (&th.DirectoryAccountBuilder{
			Client:         suite.Client,
			CanonicalEmail: &originalEmail,
			DisplayName:    "Keep Link User",
			DirectoryName:  lo.ToPtr("googleworkspace"),
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        idUser.OrganizationID,
		}).MustNew(ctx, t)

		waitForGala(t, irSetup.Runtime)

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.Client.DB, da.ID)
		assert.NilError(t, err)
		assert.Assert(t, linked.IdentityHolderID != nil)

		originalHolderID := *linked.IdentityHolderID
		assert.Check(t, originalHolderID != targetHolder.ID)

		err = suite.Client.DB.DirectoryAccount.UpdateOneID(da.ID).
			SetCanonicalEmail(targetEmail).
			Exec(ctx)
		assert.NilError(t, err)

		waitForGala(t, irSetup.Runtime)

		account, err := suite.Client.DB.DirectoryAccount.Get(ctx, da.ID)
		assert.NilError(t, err)
		assert.Assert(t, account.IdentityHolderID != nil)
		assert.Check(t, is.Equal(originalHolderID, *account.IdentityHolderID), "linked accounts only re-enrich, never relink")

		(&th.Cleanup[*generated.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&th.Cleanup[*generated.IdentityHolderDeleteOne]{Client: suite.Client.DB.IdentityHolder, IDs: []string{originalHolderID, targetHolder.ID}}).MustDelete(ctx, t)
	})

	t.Run("linked account still enriches when the cascade would resolve nothing", func(t *testing.T) {
		email := "stillenrich@updatelistener.io"

		da := (&th.DirectoryAccountBuilder{
			Client:         suite.Client,
			CanonicalEmail: &email,
			DisplayName:    "Still Enrich User",
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        idUser.OrganizationID,
		}).MustNew(ctx, t)

		waitForGala(t, irSetup.Runtime)

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.Client.DB, da.ID)
		assert.NilError(t, err)
		assert.Assert(t, linked.IdentityHolderID != nil)

		holderID := *linked.IdentityHolderID

		holder, err := suite.Client.DB.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.UserStatusActive, holder.Status))

		err = suite.Client.DB.DirectoryAccount.UpdateOneID(da.ID).
			ClearCanonicalEmail().
			ClearEmailAliases().
			SetStatus(enums.DirectoryAccountStatusSuspended).
			Exec(ctx)
		assert.NilError(t, err)

		suspended, err := listenerPoll(func() (*generated.IdentityHolder, error) {
			return suite.Client.DB.IdentityHolder.Get(ctx, holderID)
		}, func(h *generated.IdentityHolder) bool {
			return h.Status == enums.UserStatusSuspended
		})
		assert.NilError(t, err)
		assert.Check(t, !suspended.IsActive)

		account, err := suite.Client.DB.DirectoryAccount.Get(ctx, da.ID)
		assert.NilError(t, err)
		assert.Assert(t, account.IdentityHolderID != nil)
		assert.Check(t, is.Equal(holderID, *account.IdentityHolderID))

		(&th.Cleanup[*generated.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&th.Cleanup[*generated.IdentityHolderDeleteOne]{Client: suite.Client.DB.IdentityHolder, ID: holderID}).MustDelete(ctx, t)
	})

	t.Run("unlinked account links to an existing holder on update", func(t *testing.T) {
		email := "late-link@updatelistener.io"

		holder := (&th.IdentityHolderBuilder{
			Client: suite.Client,
			Email:  email,
		}).MustNew(idUser.UserCtx, t)

		da := (&th.DirectoryAccountBuilder{
			Client:        suite.Client,
			DisplayName:   "Late Link User",
			DirectoryName: lo.ToPtr("github"),
			Status:        enums.DirectoryAccountStatusActive,
			OwnerID:       idUser.OrganizationID,
		}).MustNew(ctx, t)

		waitForGala(t, irSetup.Runtime)

		account, err := suite.Client.DB.DirectoryAccount.Get(ctx, da.ID)
		assert.NilError(t, err)
		assert.Check(t, account.IdentityHolderID == nil || *account.IdentityHolderID == "")

		err = suite.Client.DB.DirectoryAccount.UpdateOneID(da.ID).
			SetCanonicalEmail(email).
			Exec(ctx)
		assert.NilError(t, err)

		linked, err := listenerPoll(func() (*generated.DirectoryAccount, error) {
			return suite.Client.DB.DirectoryAccount.Get(ctx, da.ID)
		}, func(account *generated.DirectoryAccount) bool {
			return account.IdentityHolderID != nil && *account.IdentityHolderID == holder.ID
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(holder.ID, *linked.IdentityHolderID))

		(&th.Cleanup[*generated.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&th.Cleanup[*generated.IdentityHolderDeleteOne]{Client: suite.Client.DB.IdentityHolder, ID: holder.ID}).MustDelete(ctx, t)
	})
}
