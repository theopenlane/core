package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
)

func cleanupOrganizationDataWithContext(ctx context.Context, t *testing.T) {
	t.Helper()
	caller, _ := auth.CallerFromContext(ctx)
	if caller == nil && caller.OrganizationID == "" {
		failNow(t)
	}
	_, err := suite.client.api.DeleteOrganization(ctx, caller.OrganizationID)
	requireNoError(t, err)
}

// cleanupTrustCenterData for the sharedTestUser1 context
func cleanupTrustCenterData(t *testing.T) {
	t.Helper()
	cleanupTrustCenterDataWithContext(sharedTestUser1.UserCtx, t)
}

// cleanupTrustCenterData removes all trust centers and watermark configs for the context of the passed in user context.
// This ensures the Only() query in hooks works correctly when tests expect a single watermark config.
func cleanupTrustCenterDataWithContext(ctx context.Context, t *testing.T) {
	t.Helper()
	ctx = setContext(ctx, suite.client.db)

	wcs, err := suite.client.db.TrustCenterWatermarkConfig.Query().All(ctx)
	requireNoError(t, err)
	for _, wc := range wcs {
		err := suite.client.db.TrustCenterWatermarkConfig.DeleteOneID(wc.ID).Exec(ctx)
		requireNoError(t, err)
	}

	tcs, err := suite.client.db.TrustCenter.Query().All(ctx)
	requireNoError(t, err)
	for _, tc := range tcs {
		err := suite.client.db.TrustCenter.DeleteOneID(tc.ID).Exec(ctx)
		requireNoError(t, err)
	}
}

type trustCenterOrg struct {
	organizationID string
	trustCenter    *generated.TrustCenter
	ndaTemplateID  *string
	ndaFileID      *string
	supportCtx     context.Context
	*testOrgUsers
}

type trustCenterOption func(ctx context.Context, t *testing.T, c *trustCenterConfig)
type trustCenterConfig struct {
	trustCenterID    *string
	customDomainID   *string
	ndaTemplateID    *string
	ndaFileID        *string
	seedAllUserTypes bool
	seedAPIClients   bool
	seedSupportUser  bool
}

// withSupportUser creates an org-scoped support session (auth.NewOrgSupportCaller) for the org,
// available as trustCenterOrg.supportCtx
func withSupportUser() trustCenterOption {
	return func(ctx context.Context, t *testing.T, c *trustCenterConfig) {
		c.seedSupportUser = true
	}
}

// withAllUserTypes creates the owner, super admin, admin (with api and pat clients), member, and auditor users
func withAllUserTypes() trustCenterOption {
	return func(ctx context.Context, t *testing.T, c *trustCenterConfig) {
		c.seedAllUserTypes = true
	}
}

// withAPIClients adds the admin pat and api token clients, this isn't needed when withAllUserTypes is used because that will always create the api clients
func withAPIClients() trustCenterOption {
	return func(ctx context.Context, t *testing.T, c *trustCenterConfig) {
		c.seedAPIClients = true
	}
}

// withCustomDomain adds the custom domain for the trust center
func withCustomDomain() trustCenterOption {
	return func(ctx context.Context, t *testing.T, c *trustCenterConfig) {
		if ctx == nil || c.customDomainID != nil {
			return
		}

		cd := (&CustomDomainBuilder{client: suite.client}).MustNew(ctx, t)
		c.customDomainID = &cd.ID
	}
}

// withNDATemplate adds an nda template for the trust center
func withNDATemplate() trustCenterOption {
	return func(ctx context.Context, t *testing.T, c *trustCenterConfig) {
		if ctx == nil || c.trustCenterID == nil || c.ndaTemplateID != nil {
			return
		}

		ndaFile := (&FileBuilder{
			client:  suite.client,
			Name:    "hello.pdf",
			MD5Hash: getMD5Hash(t, pdfFilePath),
		}).MustNew(ctx, t)

		tmpl := (&TemplateBuilder{
			client:        suite.client,
			Kind:          enums.TemplateKindTrustCenterNda,
			TrustCenterID: *c.trustCenterID,
			FileIDs:       []string{ndaFile.ID},
		}).MustNew(ctx, t)

		c.ndaTemplateID = &tmpl.ID
		c.ndaFileID = &ndaFile.ID
	}
}

func createFreshOrgWithTrustCenter(t *testing.T, opts ...trustCenterOption) *trustCenterOrg {
	t.Helper()
	config := trustCenterConfig{}

	// run setup options
	for _, opt := range opts {
		opt(nil, t, &config)
	}

	localUsers := &testOrgUsers{}
	if config.seedAllUserTypes {
		localUsers = suite.seedFreshOrgUsers(t)
	} else {
		users := suite.seedFreshMinimalOrgUsers(t, config.seedAPIClients)
		localUsers.owner = users.owner
		localUsers.admin = users.admin
		localUsers.member = users.member
		localUsers.adminApiClient = users.apiClient
		localUsers.adminPatClient = users.adminPatClient
	}

	ownerCtx := localUsers.owner.UserCtx

	// run pre-options post org creation
	for _, opt := range opts {
		opt(ownerCtx, t, &config)
	}

	customDomainID := ""
	if config.customDomainID != nil {
		customDomainID = *config.customDomainID
	}

	localTrustCenter := (&TrustCenterBuilder{client: suite.client, CustomDomainID: customDomainID}).MustNew(ownerCtx, t)

	config.trustCenterID = &localTrustCenter.ID

	// run post options
	for _, opt := range opts {
		opt(ownerCtx, t, &config)
	}

	var supportCtx context.Context
	if config.seedSupportUser {
		supportCtx = newSupportCtx(ownerCtx, localUsers.owner.OrganizationID)
	}

	return &trustCenterOrg{
		organizationID: localUsers.owner.OrganizationID,
		trustCenter:    localTrustCenter,
		ndaTemplateID:  config.ndaTemplateID,
		ndaFileID:      config.ndaFileID,
		supportCtx:     supportCtx,
		testOrgUsers:   localUsers,
	}
}

func cleanupWatermarkConfigsWithContext(ctx context.Context, t *testing.T) {
	t.Helper()
	wcs, err := suite.client.db.TrustCenterWatermarkConfig.Query().All(ctx)
	requireNoError(t, err)

	for _, wc := range wcs {
		err := suite.client.db.TrustCenterWatermarkConfig.DeleteOneID(wc.ID).Exec(ctx)
		requireNoError(t, err)
	}
}

// cleanupWatermarkConfigs removes all watermark configs for the test user's organization.
func cleanupWatermarkConfigs(t *testing.T) {
	t.Helper()

	ctx := privacy.DecisionContext(setContext(sharedTestUser1.UserCtx, suite.client.db), privacy.Allow)

	cleanupWatermarkConfigsWithContext(ctx, t)
}

// newAnonTrustCenterCtxFromCaller wraps an existing trust center caller in a properly initialized echo context
// so that ActiveTrustCenterIDKey and the caller both survive the HTTP pipeline in tests
func newAnonTrustCenterCtxFromCaller(caller *auth.Caller, trustCenterID string) context.Context {
	ec := echocontext.NewTestEchoContext()
	ctx := auth.WithCaller(ec.Request().Context(), caller)
	ctx = auth.ActiveTrustCenterIDKey.Set(ctx, trustCenterID)
	ctx = contextx.With(ctx, ec)
	ec.SetRequest(ec.Request().WithContext(ctx))
	return ctx
}

// createAnonymousTrustCenterContext creates a context for an anonymous trust center user
func createAnonymousTrustCenterContext(trustCenterID, organizationID string) context.Context {
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())
	caller := auth.NewTrustCenterCaller(organizationID, anonUserID, "Anonymous User", "")
	return newAnonTrustCenterCtxFromCaller(caller, trustCenterID)
}

// createAnonymousTrustCenterContextWithEmail creates a context for an anonymous trust center user with subject email
func createAnonymousTrustCenterContextWithEmail(trustCenterID, organizationID, email string) (context.Context, *auth.Caller) {
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())
	caller := auth.NewTrustCenterCaller(organizationID, anonUserID, "Anonymous User", email)
	return newAnonTrustCenterCtxFromCaller(caller, trustCenterID), caller
}
