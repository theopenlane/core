//go:build test

package hooks_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/objects"
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/objects/storage/proxy"
)

const (
	shareLinkBucket    = "openlane-prod-files"
	shareLinkObjectKey = "org/object/report.pdf"
	shareLinkTTL       = 48 * time.Hour
	shareLinkBaseURL   = "https://files.example.com"
	shareLinkTolerance = time.Minute
)

// shareLinkFixture is a user with external sharing configured and one file in their organization
type shareLinkFixture struct {
	orgID       string
	internalCtx context.Context
	file        *generated.File
}

// seedShareLinkFile creates a file record on the provider scoped to the organization
func (s *HookTestSuite) seedShareLinkFile(ctx context.Context, orgID string, provider storagetypes.ProviderType) *generated.File {
	created, err := s.client.File.Create().
		AddOrganizationIDs(orgID).
		SetProvidedFileName("report.pdf").
		SetProvidedFileExtension("pdf").
		SetDetectedContentType("application/pdf").
		SetStorageProvider(string(provider)).
		SetStorageVolume(shareLinkBucket).
		SetStoragePath(shareLinkObjectKey).
		Save(ctx)
	assert.NilError(s.T(), err)

	return created
}

// seedShareLinkFixture seeds a user whose organization has the given external sharing setting and owns one R2 file
func (s *HookTestSuite) seedShareLinkFixture(sharingEnabled bool) shareLinkFixture {
	t := s.T()

	user := s.seedUser()
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	err := entitlements.CreateFeatureTuples(context.Background(), &s.client.Authz, orgID, []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})
	assert.NilError(t, err)

	userCtx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, orgID), s.client)
	internalCtx := auth.WithInternalOperationContext(userCtx)

	err = s.client.OrganizationSetting.Update().
		Where(organizationsetting.OrganizationID(orgID)).
		SetExternalSharingEnabled(sharingEnabled).
		Exec(internalCtx)
	assert.NilError(t, err)

	return shareLinkFixture{
		orgID:       orgID,
		internalCtx: internalCtx,
		file:        s.seedShareLinkFile(internalCtx, orgID, storagetypes.R2Provider),
	}
}

// assertShareTokenVerifies checks the share token against the token rebuilt from the stored row and its secret
func assertShareTokenVerifies(t *testing.T, link *generated.ShareLink) {
	assert.Assert(t, link.Token != "")

	orgID, signature, ok := proxy.SplitShareToken(link.Token)
	assert.Assert(t, ok)
	assert.Equal(t, link.OwnerID, orgID)

	token := tokens.ShareToken{ShareID: link.ID, OrgID: ulid.MustParse(orgID), SigningInfo: tokens.SigningInfo{ExpiresAt: link.ExpiresAt}}
	assert.NilError(t, token.Verify(signature, link.Secret))
	assert.ErrorIs(t, token.Verify(signature+"x", link.Secret), tokens.ErrTokenInvalid)
}

// openClientWithLinks opens a second client on the suite database whose object manager carries the given links config
func (s *HookTestSuite) openClientWithLinks(links storage.LinksConfig) *generated.Client {
	t := s.T()

	opts := append(slices.Clone(s.opts), generated.ObjectManager(objects.NewService(objects.Config{Links: links})))

	client, err := generated.Open(dialect.Postgres, s.tf.URI, opts...)
	assert.NilError(t, err)

	t.Cleanup(func() { assert.NilError(t, client.Close()) })

	return client
}

func TestHookShareLinkCreateMintsToken(t *testing.T) {
	s := suite.setupTest(t)

	fx := s.seedShareLinkFixture(true)
	expiresAt := time.Now().Add(shareLinkTTL).Truncate(time.Second)

	link, err := s.client.ShareLink.Create().
		SetOwnerID(fx.orgID).
		SetName("audit evidence").
		SetExpiresAt(expiresAt).
		SetMaxUses(3).
		SetRecipientEmails([]string{"  Bob@Example.COM ", "carol@example.com", "bob@example.com"}).
		AddFileIDs(fx.file.ID).
		Save(fx.internalCtx)
	assert.NilError(t, err)

	assert.Check(t, is.DeepEqual([]string{"bob@example.com", "carol@example.com"}, link.RecipientEmails))
	assert.Check(t, is.Equal(3, link.MaxUses))
	assert.Check(t, is.Equal(0, link.UseCount))
	assert.Check(t, is.Equal(fx.orgID, link.OwnerID))
	assert.Check(t, link.LandingPage, "recipients force the landing page")
	assert.Check(t, link.ExpiresAt.Sub(expiresAt).Abs() < shareLinkTolerance)
	assert.Check(t, is.Equal(proxy.ComposeShareURL(shareLinkBaseURL, link.ID, link.Token), link.URL), "the created link carries the share URL")

	assertShareTokenVerifies(t, link)

	// the token must keep verifying against the row as it is read back, not only as it was returned
	stored, err := s.client.ShareLink.Get(fx.internalCtx, link.ID)
	assert.NilError(t, err)
	assert.Check(t, stored.ExpiresAt.Equal(link.ExpiresAt))

	assertShareTokenVerifies(t, stored)
}

func TestHookShareLinkCreateDefaults(t *testing.T) {
	s := suite.setupTest(t)

	fx := s.seedShareLinkFixture(true)
	before := time.Now()

	link, err := s.client.ShareLink.Create().
		SetOwnerID(fx.orgID).
		AddFileIDs(fx.file.ID).
		Save(fx.internalCtx)
	assert.NilError(t, err)

	assert.Check(t, is.Equal(proxy.UnlimitedShareUses, link.MaxUses))
	assert.Check(t, !link.LandingPage, "a default link has no cap and serves its file directly")
	assert.Check(t, is.Len(link.RecipientEmails, 0))
	assert.Check(t, link.ExpiresAt.Sub(before.Add(7*24*time.Hour)).Abs() < shareLinkTolerance)

	assert.Check(t, !link.RequireSession)

	assertShareTokenVerifies(t, link)
}

func TestHookShareLinkCreateLandingPage(t *testing.T) {
	s := suite.setupTest(t)

	fx := s.seedShareLinkFixture(true)
	second := s.seedShareLinkFile(fx.internalCtx, fx.orgID, storagetypes.R2Provider)

	testCases := []struct {
		name         string
		fileIDs      []string
		requested    bool
		recipients   []string
		maxUses      int
		expectedPage bool
	}{
		{name: "single file without recipients leaves the landing page off", fileIDs: []string{fx.file.ID}},
		{name: "explicit landing page is kept", fileIDs: []string{fx.file.ID}, requested: true, expectedPage: true},
		{name: "multiple files force the landing page", fileIDs: []string{fx.file.ID, second.ID}, expectedPage: true},
		{name: "recipients force the landing page", fileIDs: []string{fx.file.ID}, recipients: []string{"bob@example.com"}, expectedPage: true},
		{name: "use cap forces the landing page", fileIDs: []string{fx.file.ID}, maxUses: 1, expectedPage: true},
		{name: "explicit unlimited uses leave the landing page off", fileIDs: []string{fx.file.ID}, maxUses: proxy.UnlimitedShareUses},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			create := s.client.ShareLink.Create().
				SetOwnerID(fx.orgID).
				AddFileIDs(tc.fileIDs...).
				SetMaxUses(tc.maxUses).
				SetLandingPage(tc.requested)

			if len(tc.recipients) > 0 {
				create = create.SetRecipientEmails(tc.recipients)
			}

			link, err := create.Save(fx.internalCtx)
			assert.NilError(t, err)

			assert.Check(t, is.Equal(tc.expectedPage, link.LandingPage))

			assertShareTokenVerifies(t, link)

			files, err := link.QueryFiles().IDs(fx.internalCtx)
			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(slices.Sorted(slices.Values(tc.fileIDs)), slices.Sorted(slices.Values(files))))
		})
	}
}

func TestHookShareLinkCreateRefusals(t *testing.T) {
	s := suite.setupTest(t)

	fx := s.seedShareLinkFixture(true)
	disabled := s.seedShareLinkFixture(false)
	foreign := s.seedShareLinkFixture(true)
	s3File := s.seedShareLinkFile(fx.internalCtx, fx.orgID, storagetypes.S3Provider)
	tooMany := lo.Times(proxy.MaxShareFiles+1, func(int) string {
		return s.seedShareLinkFile(fx.internalCtx, fx.orgID, storagetypes.R2Provider).ID
	})
	linksDisabled := s.openClientWithLinks(storage.LinksConfig{})

	testCases := []struct {
		name        string
		ctx         context.Context
		build       func() *generated.ShareLinkCreate
		expectedErr string
	}{
		{
			name: "external sharing disabled",
			ctx:  disabled.internalCtx,
			build: func() *generated.ShareLinkCreate {
				return s.client.ShareLink.Create().SetOwnerID(disabled.orgID).AddFileIDs(disabled.file.ID)
			},
			expectedErr: hooks.ErrShareLinkExternalSharingDisabled.Error(),
		},
		{
			name: "expiry beyond the maximum",
			ctx:  fx.internalCtx,
			build: func() *generated.ShareLinkCreate {
				return s.client.ShareLink.Create().SetOwnerID(fx.orgID).AddFileIDs(fx.file.ID).SetExpiresAt(time.Now().Add(proxy.MaxShareDuration + time.Hour))
			},
			expectedErr: hooks.ErrShareLinkExpiryTooLong.Error(),
		},
		{
			name: "expiry in the past",
			ctx:  fx.internalCtx,
			build: func() *generated.ShareLinkCreate {
				return s.client.ShareLink.Create().SetOwnerID(fx.orgID).AddFileIDs(fx.file.ID).SetExpiresAt(time.Now().Add(-time.Minute))
			},
			expectedErr: hooks.ErrShareLinkExpiryPast.Error(),
		},
		{
			name: "file the caller cannot read",
			ctx:  fx.internalCtx,
			build: func() *generated.ShareLinkCreate {
				return s.client.ShareLink.Create().SetOwnerID(fx.orgID).AddFileIDs(foreign.file.ID)
			},
			expectedErr: generated.ErrPermissionDenied.Error(),
		},
		{
			name: "file not stored on R2",
			ctx:  fx.internalCtx,
			build: func() *generated.ShareLinkCreate {
				return s.client.ShareLink.Create().SetOwnerID(fx.orgID).AddFileIDs(s3File.ID)
			},
			expectedErr: hooks.ErrInvalidInput.Error(),
		},
		{
			name: "more files than the maximum",
			ctx:  fx.internalCtx,
			build: func() *generated.ShareLinkCreate {
				return s.client.ShareLink.Create().SetOwnerID(fx.orgID).AddFileIDs(tooMany...)
			},
			expectedErr: hooks.ErrShareLinkTooManyFiles.Error(),
		},
		{
			name: "empty recipient",
			ctx:  fx.internalCtx,
			build: func() *generated.ShareLinkCreate {
				return s.client.ShareLink.Create().SetOwnerID(fx.orgID).AddFileIDs(fx.file.ID).SetRecipientEmails([]string{""})
			},
			expectedErr: "mail: no address",
		},
		{
			name: "blank recipient",
			ctx:  fx.internalCtx,
			build: func() *generated.ShareLinkCreate {
				return s.client.ShareLink.Create().SetOwnerID(fx.orgID).AddFileIDs(fx.file.ID).SetRecipientEmails([]string{"  "})
			},
			expectedErr: "mail: no address",
		},
		{
			name: "recipient that is not an email address among valid ones",
			ctx:  fx.internalCtx,
			build: func() *generated.ShareLinkCreate {
				return s.client.ShareLink.Create().SetOwnerID(fx.orgID).AddFileIDs(fx.file.ID).SetRecipientEmails([]string{"bob@example.com", "not-an-email"})
			},
			expectedErr: "mail: missing '@' or angle-addr",
		},
		{
			name: "storage links disabled",
			ctx:  fx.internalCtx,
			build: func() *generated.ShareLinkCreate {
				return linksDisabled.ShareLink.Create().SetOwnerID(fx.orgID).AddFileIDs(fx.file.ID)
			},
			expectedErr: hooks.ErrShareLinksDisabled.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.build().Save(tc.ctx)
			assert.ErrorContains(t, err, tc.expectedErr)
		})
	}
}

func TestHookShareLinkCreateRecipients(t *testing.T) {
	s := suite.setupTest(t)

	fx := s.seedShareLinkFixture(true)

	testCases := []struct {
		name       string
		recipients []string
		expected   []string
	}{
		{name: "display name is dropped and the address lowercased", recipients: []string{"Bob <BOB@example.com>"}, expected: []string{"bob@example.com"}},
		{name: "the same address in different forms is stored once", recipients: []string{"Bob <BOB@example.com>", " bob@EXAMPLE.com", "carol@example.com"}, expected: []string{"bob@example.com", "carol@example.com"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			link, err := s.client.ShareLink.Create().
				SetOwnerID(fx.orgID).
				SetRecipientEmails(tc.recipients).
				AddFileIDs(fx.file.ID).
				Save(fx.internalCtx)
			assert.NilError(t, err)

			assert.Check(t, is.DeepEqual(tc.expected, link.RecipientEmails))
			assert.Check(t, link.LandingPage)

			stored, err := s.client.ShareLink.Get(fx.internalCtx, link.ID)
			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(tc.expected, stored.RecipientEmails))
		})
	}
}

func TestHookShareLinkUpdate(t *testing.T) {
	s := suite.setupTest(t)

	fx := s.seedShareLinkFixture(true)

	// seed creates a share link over the fixture file with the given cap
	seed := func(t *testing.T, maxUses int) *generated.ShareLink {
		link, err := s.client.ShareLink.Create().
			SetOwnerID(fx.orgID).
			SetMaxUses(maxUses).
			AddFileIDs(fx.file.ID).
			Save(fx.internalCtx)
		assert.NilError(t, err)

		return link
	}

	testCases := []struct {
		name            string
		maxUses         int
		revoked         bool
		update          func(*generated.ShareLinkUpdateOne) *generated.ShareLinkUpdateOne
		expectedErr     error
		expectedMaxUses int
		expectedRevoked bool
	}{
		{
			name:            "revoking is allowed",
			update:          func(u *generated.ShareLinkUpdateOne) *generated.ShareLinkUpdateOne { return u.SetRevokedAt(time.Now()) },
			expectedRevoked: true,
		},
		{
			name:            "a revocation cannot be cleared",
			revoked:         true,
			update:          func(u *generated.ShareLinkUpdateOne) *generated.ShareLinkUpdateOne { return u.ClearRevokedAt() },
			expectedErr:     hooks.ErrShareLinkRevocationPermanent,
			expectedRevoked: true,
		},
		{
			name:        "a cap on a direct link is refused",
			update:      func(u *generated.ShareLinkUpdateOne) *generated.ShareLinkUpdateOne { return u.SetMaxUses(3) },
			expectedErr: hooks.ErrShareLinkDirectUseCap,
		},
		{
			name: "unlimited uses on a direct link are allowed",
			update: func(u *generated.ShareLinkUpdateOne) *generated.ShareLinkUpdateOne {
				return u.SetMaxUses(proxy.UnlimitedShareUses)
			},
		},
		{
			name:            "a landing page link can change its cap",
			maxUses:         1,
			update:          func(u *generated.ShareLinkUpdateOne) *generated.ShareLinkUpdateOne { return u.SetMaxUses(3) },
			expectedMaxUses: 3,
		},
		{
			name:    "a landing page link can clear its cap",
			maxUses: 1,
			update:  func(u *generated.ShareLinkUpdateOne) *generated.ShareLinkUpdateOne { return u.ClearMaxUses() },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			link := seed(t, tc.maxUses)
			assert.Assert(t, is.Equal(tc.maxUses != proxy.UnlimitedShareUses, link.LandingPage))

			if tc.revoked {
				assert.NilError(t, s.client.ShareLink.UpdateOneID(link.ID).SetRevokedAt(time.Now()).Exec(fx.internalCtx))
			}

			err := tc.update(s.client.ShareLink.UpdateOneID(link.ID)).Exec(fx.internalCtx)

			switch {
			case tc.expectedErr != nil:
				assert.ErrorIs(t, err, tc.expectedErr)
			default:
				assert.NilError(t, err)
			}

			stored, err := s.client.ShareLink.Get(fx.internalCtx, link.ID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(lo.Ternary(tc.expectedErr != nil, tc.maxUses, tc.expectedMaxUses), stored.MaxUses))
			assert.Check(t, is.Equal(tc.expectedRevoked, stored.RevokedAt != nil))
		})
	}
}
