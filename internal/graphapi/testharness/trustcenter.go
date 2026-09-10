//go:build test

package testharness

import (
	"context"
	"fmt"
	"testing"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/httpserve/authmanager"
)

type TrustCenterOrg struct {
	OrganizationID string
	TrustCenter    *generated.TrustCenter
	NDATemplateID  *string
	NDAFileID      *string
	SupportCtx     context.Context
	*TestOrgUsers
}

type TrustCenterOption func(ctx context.Context, t *testing.T, c *TrustCenterConfig)
type TrustCenterConfig struct {
	trustCenterID    *string
	customDomainID   *string
	ndaTemplateID    *string
	ndaFileID        *string
	seedAllUserTypes bool
	seedAPIClients   bool
	seedSupportUser  bool
}

// WithSupportUser creates an org-scoped support session (auth.NewOrgSupportCaller) for the org,
// available as TrustCenterOrg.supportCtx
func WithSupportUser() TrustCenterOption {
	return func(ctx context.Context, t *testing.T, c *TrustCenterConfig) {
		c.seedSupportUser = true
	}
}

// WithAllUserTypes creates the owner, super admin, admin (with api and pat clients), member, and auditor users
func WithAllUserTypes() TrustCenterOption {
	return func(ctx context.Context, t *testing.T, c *TrustCenterConfig) {
		c.seedAllUserTypes = true
	}
}

// WithAPIClients adds the admin pat and api token clients, this isn't needed when WithAllUserTypes is used because that will always create the api clients
func WithAPIClients() TrustCenterOption {
	return func(ctx context.Context, t *testing.T, c *TrustCenterConfig) {
		c.seedAPIClients = true
	}
}

// WithCustomDomain adds the custom domain for the trust center
func WithCustomDomain() TrustCenterOption {
	return func(ctx context.Context, t *testing.T, c *TrustCenterConfig) {
		if ctx == nil || c.customDomainID != nil {
			return
		}

		cd := (&CustomDomainBuilder{Client: Suite.Client}).MustNew(ctx, t)
		c.customDomainID = &cd.ID
	}
}

// WithNDATemplate adds an nda template for the trust center
func WithNDATemplate() TrustCenterOption {
	return func(ctx context.Context, t *testing.T, c *TrustCenterConfig) {
		if ctx == nil || c.trustCenterID == nil || c.ndaTemplateID != nil {
			return
		}

		ndaFile := (&FileBuilder{
			Client:  Suite.Client,
			Name:    "hello.pdf",
			MD5Hash: GetMD5Hash(t, PdfFilePath),
		}).MustNew(ctx, t)

		tmpl := (&TemplateBuilder{
			Client:        Suite.Client,
			Kind:          enums.TemplateKindTrustCenterNda,
			TrustCenterID: *c.trustCenterID,
			FileIDs:       []string{ndaFile.ID},
		}).MustNew(ctx, t)

		c.ndaTemplateID = &tmpl.ID
		c.ndaFileID = &ndaFile.ID
	}
}

func CreateFreshOrgWithTrustCenter(t *testing.T, opts ...TrustCenterOption) *TrustCenterOrg {
	t.Helper()
	config := TrustCenterConfig{}

	// run setup options
	for _, opt := range opts {
		opt(nil, t, &config)
	}

	localUsers := &TestOrgUsers{}
	if config.seedAllUserTypes {
		localUsers = Suite.SeedFreshOrgUsers(t)
	} else {
		users := Suite.SeedFreshMinimalOrgUsers(t, config.seedAPIClients)
		localUsers.Owner = users.Owner
		localUsers.Admin = users.Admin
		localUsers.Member = users.Member
		localUsers.AdminAPIClient = users.APIClient
		localUsers.AdminPatClient = users.AdminPatClient
	}

	ownerCtx := localUsers.Owner.UserCtx

	// run pre-options post org creation
	for _, opt := range opts {
		opt(ownerCtx, t, &config)
	}

	customDomainID := ""
	if config.customDomainID != nil {
		customDomainID = *config.customDomainID
	}

	localTrustCenter := (&TrustCenterBuilder{Client: Suite.Client, CustomDomainID: customDomainID}).MustNew(ownerCtx, t)

	config.trustCenterID = &localTrustCenter.ID

	// run post options
	for _, opt := range opts {
		opt(ownerCtx, t, &config)
	}

	var supportCtx context.Context
	if config.seedSupportUser {
		supportCtx = NewSupportCtx(ownerCtx, localUsers.Owner.OrganizationID)
	}

	return &TrustCenterOrg{
		OrganizationID: localUsers.Owner.OrganizationID,
		TrustCenter:    localTrustCenter,
		NDATemplateID:  config.ndaTemplateID,
		NDAFileID:      config.ndaFileID,
		SupportCtx:     supportCtx,
		TestOrgUsers:   localUsers,
	}
}

// NewAnonTrustCenterCtxFromCaller wraps an existing trust center caller in a properly initialized echo context
// so that ActiveTrustCenterIDKey and the caller both survive the HTTP pipeline in tests
func NewAnonTrustCenterCtxFromCaller(caller *auth.Caller, trustCenterID string) context.Context {
	ec := echocontext.NewTestEchoContext()
	ctx := auth.WithCaller(ec.Request().Context(), caller)
	ctx = auth.ActiveTrustCenterIDKey.Set(ctx, trustCenterID)
	ctx = contextx.With(ctx, ec)
	ec.SetRequest(ec.Request().WithContext(ctx))
	return ctx
}

// CreateAnonymousTrustCenterContext creates a context for an anonymous trust center user
func CreateAnonymousTrustCenterContext(trustCenterID, organizationID string) context.Context {
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())
	caller := auth.NewTrustCenterCaller(organizationID, anonUserID, "Anonymous User", "")
	return NewAnonTrustCenterCtxFromCaller(caller, trustCenterID)
}

// CreateAnonymousTrustCenterContextWithEmail creates a context for an anonymous trust center user with subject email
func CreateAnonymousTrustCenterContextWithEmail(trustCenterID, organizationID, email string) (context.Context, *auth.Caller) {
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())
	caller := auth.NewTrustCenterCaller(organizationID, anonUserID, "Anonymous User", email)
	return NewAnonTrustCenterCtxFromCaller(caller, trustCenterID), caller
}
