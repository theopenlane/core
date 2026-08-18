//go:build test

package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi"
)

func TestIdentityResolutionUpdateCascade(t *testing.T) {
	idUser := suite.userBuilder(context.Background(), t)
	ctx := setContext(idUser.UserCtx, suite.client.db)

	irSetup, err := graphapi.SetupListenerRuntime(ctx, suite.client.db, suite.tf.URI, hooks.IdentityResolutionListeners())
	assert.NilError(t, err)
	defer irSetup.Teardown()

	t.Run("canonical email change keeps the linked holder", func(t *testing.T) {
		originalEmail := "keeplink-original@updatelistener.io"
		targetEmail := "keeplink-target@updatelistener.io"

		targetHolder := (&IdentityHolderBuilder{
			client: suite.client,
			Email:  targetEmail,
		}).MustNew(idUser.UserCtx, t)

		da := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &originalEmail,
			DisplayName:    "Keep Link User",
			DirectoryName:  lo.ToPtr("googleworkspace"),
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        idUser.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
		assert.NilError(t, err)
		assert.Assert(t, linked.IdentityHolderID != nil)

		originalHolderID := *linked.IdentityHolderID
		assert.Check(t, originalHolderID != targetHolder.ID)

		err = suite.client.db.DirectoryAccount.UpdateOneID(da.ID).
			SetCanonicalEmail(targetEmail).
			Exec(ctx)
		assert.NilError(t, err)

		irSetup.Runtime.WaitIdle()

		account, err := suite.client.db.DirectoryAccount.Get(ctx, da.ID)
		assert.NilError(t, err)
		assert.Assert(t, account.IdentityHolderID != nil)
		assert.Check(t, is.Equal(originalHolderID, *account.IdentityHolderID), "linked accounts only re-enrich, never relink")

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, IDs: []string{originalHolderID, targetHolder.ID}}).MustDelete(ctx, t)
	})

	t.Run("linked account still enriches when the cascade would resolve nothing", func(t *testing.T) {
		email := "stillenrich@updatelistener.io"

		da := (&DirectoryAccountBuilder{
			client:         suite.client,
			CanonicalEmail: &email,
			DisplayName:    "Still Enrich User",
			DirectoryName:  lo.ToPtr("googleworkspace"),
			PrimarySource:  true,
			Status:         enums.DirectoryAccountStatusActive,
			OwnerID:        idUser.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		linked, err := graphapi.WaitForIdentityHolderLink(ctx, suite.client.db, da.ID)
		assert.NilError(t, err)
		assert.Assert(t, linked.IdentityHolderID != nil)

		holderID := *linked.IdentityHolderID

		holder, err := suite.client.db.IdentityHolder.Get(ctx, holderID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.UserStatusActive, holder.Status))

		err = suite.client.db.DirectoryAccount.UpdateOneID(da.ID).
			ClearCanonicalEmail().
			ClearEmailAliases().
			SetStatus(enums.DirectoryAccountStatusSuspended).
			Exec(ctx)
		assert.NilError(t, err)

		suspended, err := listenerPoll(func() (*generated.IdentityHolder, error) {
			return suite.client.db.IdentityHolder.Get(ctx, holderID)
		}, func(h *generated.IdentityHolder) bool {
			return h.Status == enums.UserStatusSuspended
		})
		assert.NilError(t, err)
		assert.Check(t, !suspended.IsActive)

		account, err := suite.client.db.DirectoryAccount.Get(ctx, da.ID)
		assert.NilError(t, err)
		assert.Assert(t, account.IdentityHolderID != nil)
		assert.Check(t, is.Equal(holderID, *account.IdentityHolderID))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holderID}).MustDelete(ctx, t)
	})

	t.Run("unlinked account links to an existing holder on update", func(t *testing.T) {
		email := "late-link@updatelistener.io"

		holder := (&IdentityHolderBuilder{
			client: suite.client,
			Email:  email,
		}).MustNew(idUser.UserCtx, t)

		da := (&DirectoryAccountBuilder{
			client:        suite.client,
			DisplayName:   "Late Link User",
			DirectoryName: lo.ToPtr("github"),
			Status:        enums.DirectoryAccountStatusActive,
			OwnerID:       idUser.OrganizationID,
		}).MustNew(ctx, t)

		irSetup.Runtime.WaitIdle()

		account, err := suite.client.db.DirectoryAccount.Get(ctx, da.ID)
		assert.NilError(t, err)
		assert.Check(t, account.IdentityHolderID == nil || *account.IdentityHolderID == "")

		err = suite.client.db.DirectoryAccount.UpdateOneID(da.ID).
			SetCanonicalEmail(email).
			Exec(ctx)
		assert.NilError(t, err)

		linked, err := listenerPoll(func() (*generated.DirectoryAccount, error) {
			return suite.client.db.DirectoryAccount.Get(ctx, da.ID)
		}, func(account *generated.DirectoryAccount) bool {
			return account.IdentityHolderID != nil && *account.IdentityHolderID == holder.ID
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(holder.ID, *linked.IdentityHolderID))

		(&Cleanup[*generated.DirectoryAccountDeleteOne]{client: suite.client.db.DirectoryAccount, ID: da.ID}).MustDelete(ctx, t)
		(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: holder.ID}).MustDelete(ctx, t)
	})
}
