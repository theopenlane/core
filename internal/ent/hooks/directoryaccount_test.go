//go:build test

package hooks_test

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/iam/auth"
)

// TestDirectoryAccountHookLinksExistingHolderByEmail verifies primary email matching
func (suite *HookTestSuite) TestDirectoryAccountHookLinksExistingHolderByEmail() {
	t := suite.T()
	user := suite.seedUser()
	ctx := suite.directoryAccountHookContext(user)
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	holder, err := suite.client.IdentityHolder.Create().
		SetOwnerID(orgID).
		SetFullName("Existing Holder").
		SetEmail("user@example.com").
		Save(ctx)
	require.NoError(t, err)

	account, err := suite.client.DirectoryAccount.Create().
		SetOwnerID(orgID).
		SetExternalID("existing-holder-email").
		SetCanonicalEmail("USER@example.com").
		SetDisplayName("Directory User").
		Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, account.IdentityHolderID)
	require.Equal(t, holder.ID, *account.IdentityHolderID)

	stored, err := suite.client.DirectoryAccount.Get(ctx, account.ID)
	require.NoError(t, err)
	require.NotNil(t, stored.IdentityHolderID)
	require.Equal(t, holder.ID, *stored.IdentityHolderID)
}

// TestDirectoryAccountHookLinksExistingHolderByAlternateEmail verifies alternate email matching
func (suite *HookTestSuite) TestDirectoryAccountHookLinksExistingHolderByAlternateEmail() {
	t := suite.T()
	user := suite.seedUser()
	ctx := suite.directoryAccountHookContext(user)
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	holder, err := suite.client.IdentityHolder.Create().
		SetOwnerID(orgID).
		SetFullName("Existing Holder").
		SetEmail("primary@example.com").
		SetAlternateEmail("user@example.com").
		Save(ctx)
	require.NoError(t, err)

	account, err := suite.client.DirectoryAccount.Create().
		SetOwnerID(orgID).
		SetExternalID("alternate-holder-email").
		SetCanonicalEmail("user@example.com").
		SetDisplayName("Directory User").
		Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, account.IdentityHolderID)
	require.Equal(t, holder.ID, *account.IdentityHolderID)

	count, err := suite.client.IdentityHolder.Query().Where(identityholder.OwnerID(orgID)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

// TestDirectoryAccountHookReusesHolderLinkedFromOtherDirectoryAccount verifies reuse through linked directory evidence
func (suite *HookTestSuite) TestDirectoryAccountHookReusesHolderLinkedFromOtherDirectoryAccount() {
	t := suite.T()
	user := suite.seedUser()
	ctx := suite.directoryAccountHookContext(user)
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	holder, err := suite.client.IdentityHolder.Create().
		SetOwnerID(orgID).
		SetFullName("Canonical Holder").
		SetEmail("canonical@example.com").
		Save(ctx)
	require.NoError(t, err)

	_, err = suite.client.DirectoryAccount.Create().
		SetOwnerID(orgID).
		SetExternalID("manually-linked-account").
		SetCanonicalEmail("alias@example.com").
		SetDisplayName("Alias User").
		SetIdentityHolderID(holder.ID).
		Save(ctx)
	require.NoError(t, err)

	account, err := suite.client.DirectoryAccount.Create().
		SetOwnerID(orgID).
		SetExternalID("second-alias-account").
		SetCanonicalEmail("ALIAS@example.com").
		SetDisplayName("Alias User Two").
		Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, account.IdentityHolderID)
	require.Equal(t, holder.ID, *account.IdentityHolderID)
}

// TestDirectoryAccountHookLeavesAmbiguousAlternateEmailUnlinked verifies ambiguous matches stay unlinked
func (suite *HookTestSuite) TestDirectoryAccountHookLeavesAmbiguousAlternateEmailUnlinked() {
	t := suite.T()
	user := suite.seedUser()
	ctx := suite.directoryAccountHookContext(user)
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	_, err := suite.client.IdentityHolder.Create().
		SetOwnerID(orgID).
		SetFullName("First Holder").
		SetEmail("first@example.com").
		SetAlternateEmail("shared@example.com").
		Save(ctx)
	require.NoError(t, err)

	_, err = suite.client.IdentityHolder.Create().
		SetOwnerID(orgID).
		SetFullName("Second Holder").
		SetEmail("second@example.com").
		SetAlternateEmail("shared@example.com").
		Save(ctx)
	require.NoError(t, err)

	account, err := suite.client.DirectoryAccount.Create().
		SetOwnerID(orgID).
		SetExternalID("ambiguous-alternate-email").
		SetCanonicalEmail("shared@example.com").
		SetDisplayName("Shared User").
		Save(ctx)
	require.NoError(t, err)
	require.Nil(t, account.IdentityHolderID)

	count, err := suite.client.IdentityHolder.Query().Where(identityholder.OwnerID(orgID)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

// TestDirectoryAccountHookDoesNotCreateHolderForServiceAccounts verifies non-human accounts are ignored
func (suite *HookTestSuite) TestDirectoryAccountHookDoesNotCreateHolderForServiceAccounts() {
	t := suite.T()
	user := suite.seedUser()
	ctx := suite.directoryAccountHookContext(user)
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	account, err := suite.client.DirectoryAccount.Create().
		SetOwnerID(orgID).
		SetExternalID("service-account").
		SetCanonicalEmail("svc@example.com").
		SetDisplayName("Service Account").
		SetAccountType(enums.DirectoryAccountTypeService).
		Save(ctx)
	require.NoError(t, err)
	require.Nil(t, account.IdentityHolderID)

	count, err := suite.client.IdentityHolder.Query().Where(identityholder.OwnerID(orgID)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

// TestDirectoryAccountHookEnrichesPlaceholderHolderFields verifies conservative holder enrichment
func (suite *HookTestSuite) TestDirectoryAccountHookEnrichesPlaceholderHolderFields() {
	t := suite.T()
	user := suite.seedUser()
	ctx := suite.directoryAccountHookContext(user)
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	holder, err := suite.client.IdentityHolder.Create().
		SetOwnerID(orgID).
		SetFullName("user@example.com").
		SetEmail("user@example.com").
		Save(ctx)
	require.NoError(t, err)

	jobTitle := "Staff Engineer"
	department := "Engineering"

	_, err = suite.client.DirectoryAccount.Create().
		SetOwnerID(orgID).
		SetExternalID("placeholder-enrichment").
		SetCanonicalEmail("user@example.com").
		SetDisplayName("Directory User").
		SetJobTitle(jobTitle).
		SetDepartment(department).
		Save(ctx)
	require.NoError(t, err)

	holder, err = suite.client.IdentityHolder.Get(ctx, holder.ID)
	require.NoError(t, err)
	require.Equal(t, "Directory User", holder.FullName)
	require.Equal(t, jobTitle, holder.Title)
	require.Equal(t, department, holder.Department)
}

// TestDirectoryAccountHookAggregatesHolderLifecycleAcrossLinkedAccounts verifies aggregate lifecycle behavior
func (suite *HookTestSuite) TestDirectoryAccountHookAggregatesHolderLifecycleAcrossLinkedAccounts() {
	t := suite.T()
	user := suite.seedUser()
	ctx := suite.directoryAccountHookContext(user)
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	holder, err := suite.client.IdentityHolder.Create().
		SetOwnerID(orgID).
		SetFullName("Directory User").
		SetEmail("user@example.com").
		Save(ctx)
	require.NoError(t, err)

	_, err = suite.client.DirectoryAccount.Create().
		SetOwnerID(orgID).
		SetExternalID("lifecycle-active-1").
		SetCanonicalEmail("user@example.com").
		SetDisplayName("Directory User").
		SetIdentityHolderID(holder.ID).
		SetStatus(enums.DirectoryAccountStatusActive).
		Save(ctx)
	require.NoError(t, err)

	holder, err = suite.client.IdentityHolder.Get(ctx, holder.ID)
	require.NoError(t, err)
	require.Equal(t, enums.UserStatusActive, holder.Status)
	require.True(t, holder.IsActive)

	_, err = suite.client.DirectoryAccount.Create().
		SetOwnerID(orgID).
		SetExternalID("lifecycle-active-2").
		SetCanonicalEmail("user@example.com").
		SetDisplayName("Directory User").
		SetIdentityHolderID(holder.ID).
		SetStatus(enums.DirectoryAccountStatusActive).
		Save(ctx)
	require.NoError(t, err)

	holder, err = suite.client.IdentityHolder.Get(ctx, holder.ID)
	require.NoError(t, err)
	require.Equal(t, enums.UserStatusActive, holder.Status)
	require.True(t, holder.IsActive)

	secondAccount, err := suite.client.DirectoryAccount.Query().
		Where(directoryaccount.ExternalID("lifecycle-active-2")).
		Only(ctx)
	require.NoError(t, err)

	err = suite.client.DirectoryAccount.UpdateOneID(secondAccount.ID).
		SetStatus(enums.DirectoryAccountStatusSuspended).
		Exec(ctx)
	require.NoError(t, err)

	holder, err = suite.client.IdentityHolder.Get(ctx, holder.ID)
	require.NoError(t, err)
	require.Equal(t, enums.UserStatusUnknown, holder.Status)
	require.False(t, holder.IsActive)
}

// TestDirectoryAccountHookLinksAccountWhenEmailArrivesOnUpdate verifies delayed email linkage on update
func (suite *HookTestSuite) TestDirectoryAccountHookLinksAccountWhenEmailArrivesOnUpdate() {
	t := suite.T()
	user := suite.seedUser()
	ctx := suite.directoryAccountHookContext(user)
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	account, err := suite.client.DirectoryAccount.Create().
		SetOwnerID(orgID).
		SetExternalID("email-arrives-on-update").
		SetDisplayName("Directory User").
		Save(ctx)
	require.NoError(t, err)
	require.Nil(t, account.IdentityHolderID)

	err = suite.client.DirectoryAccount.UpdateOneID(account.ID).
		SetCanonicalEmail("user@example.com").
		SetDisplayName("Directory User").
		Exec(ctx)
	require.NoError(t, err)

	account, err = suite.client.DirectoryAccount.Query().
		Where(directoryaccount.ID(account.ID)).
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, account.IdentityHolderID)

	count, err := suite.client.IdentityHolder.Query().Where(identityholder.OwnerID(orgID)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

// directoryAccountHookContext builds a privacy-allowed context for directory account hook tests
func (suite *HookTestSuite) directoryAccountHookContext(user *generated.User) context.Context {
	ctx := auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].OrganizationID)
	ctx = generated.NewContext(ctx, suite.client)
	return privacy.DecisionContext(ctx, privacy.Allow)
}
