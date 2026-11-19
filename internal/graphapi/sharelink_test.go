package graphapi_test

import (
	"context"
	"strings"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/fgax"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/sharelinkaccessevent"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/core/v2/pkg/objects/storage/proxy"
)

const (
	storedFileBucket      = "openlane-test-files"
	storedFileKey         = "org/object/report.pdf"
	shareLinkRecipient    = "recipient@example.com"
	shareLinkUseLimit     = int64(5)
	shareLinkDefaultUses  = int64(0)
	shareLinkSingleUse    = int64(1)
	shareLinkExpiryTooFar = 2 * proxy.MaxShareDuration
)

// createShareLinkURLMutation creates a share link selecting the computed url, which the generated client does not request
const createShareLinkURLMutation = `mutation CreateShareLink($input: CreateShareLinkInput!) {
  createShareLink(input: $input) {
    shareLink { id url requireSession }
  }
}`

// getShareLinkURLQuery reads the computed url of a share link, which the generated client does not request
const getShareLinkURLQuery = `query GetShareLinkByID($shareLinkId: ID!) {
  shareLink(id: $shareLinkId) { id url requireSession }
}`

// shareLinkURLFields are the share link fields selected by the url documents
type shareLinkURLFields struct {
	ID             string  `json:"id"`
	URL            *string `json:"url"`
	RequireSession *bool   `json:"requireSession"`
}

// createShareLinkURLResponse is the response of createShareLinkURLMutation
type createShareLinkURLResponse struct {
	CreateShareLink struct {
		ShareLink shareLinkURLFields `json:"shareLink"`
	} `json:"createShareLink"`
}

// getShareLinkURLResponse is the response of getShareLinkURLQuery
type getShareLinkURLResponse struct {
	ShareLink shareLinkURLFields `json:"shareLink"`
}

// seedStoredFile creates a file record on the provider in the user's organization, viewable by its members
func seedStoredFile(t *testing.T, user th.TestUserDetails, provider storagetypes.ProviderType) *generated.File {
	file, err := suite.Client.DB.File.Create().
		AddOrganizationIDs(user.OrganizationID).
		SetProvidedFileName("report.pdf").
		SetProvidedFileExtension("pdf").
		SetDetectedContentType("application/pdf").
		SetStorageProvider(string(provider)).
		SetStorageVolume(storedFileBucket).
		SetStoragePath(storedFileKey).
		Save(th.SetInternalContext(user.UserCtx, suite.Client.DB))
	assert.NilError(t, err)

	// a file record created directly carries no access tuples, upload writes them
	tuple := fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:       user.OrganizationID,
		SubjectType:     "organization",
		SubjectRelation: fgax.MemberRelation,
		ObjectID:        file.ID,
		ObjectType:      "file",
		Relation:        fgax.CanView,
	})

	_, err = suite.Client.DB.Authz.WriteTupleKeys(user.UserCtx, []fgax.TupleKey{tuple}, nil)
	assert.NilError(t, err)

	return file
}

// enableExternalSharing turns on external sharing for the user's organization
func enableExternalSharing(t *testing.T, user th.TestUserDetails) {
	err := suite.Client.DB.OrganizationSetting.Update().
		Where(organizationsetting.OrganizationID(user.OrganizationID)).
		SetExternalSharingEnabled(true).
		Exec(th.SetInternalContext(user.UserCtx, suite.Client.DB))
	assert.NilError(t, err)
}

// seedShareLink creates a share link for one R2 file in the user's organization
func seedShareLink(t *testing.T, user th.TestUserDetails) string {
	return seedShareLinkWithUses(t, user, nil)
}

// seedShareLinkWithUses creates a share link with the use limit for one R2 file in the user's organization
func seedShareLinkWithUses(t *testing.T, user th.TestUserDetails, maxUses *int64) string {
	enableExternalSharing(t, user)

	file := seedStoredFile(t, user, storagetypes.R2Provider)

	resp, err := suite.Client.API.CreateShareLink(user.UserCtx, testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, MaxUses: maxUses})
	assert.NilError(t, err)

	return resp.CreateShareLink.ShareLink.ID
}

// shareLinkManagerRole is the organization role that lets a member manage share links
const shareLinkManagerRole = "share_link_manager"

// grantFileEdit makes the user an editor of the file
func grantFileEdit(t *testing.T, owner, user th.TestUserDetails, fileID string) {
	tuple := fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:   user.ID,
		SubjectType: "user",
		ObjectID:    fileID,
		ObjectType:  "file",
		Relation:    fgax.EditorRelation,
	})

	_, err := suite.Client.DB.Authz.WriteTupleKeys(owner.UserCtx, []fgax.TupleKey{tuple}, nil)
	assert.NilError(t, err)
}

func TestQueryShareLink(t *testing.T) {
	shareLinkID := seedShareLink(t, th.SharedTestUser1)
	manager := suite.OrgMemberWithFunctionalRoles(t, th.SharedTestUser1, shareLinkManagerRole)

	testCases := []struct {
		name     string
		queryID  string
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: shareLinkID,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, another member with edit access",
			queryID: shareLinkID,
			ctx:     th.SharedAdminUser.UserCtx,
		},
		{
			name:    "happy path, member with the share link manager role",
			queryID: shareLinkID,
			ctx:     manager.UserCtx,
		},
		{
			name:     "share link in another organization",
			queryID:  shareLinkID,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "not authorized, view only user",
			queryID:  shareLinkID,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.GetShareLinkByID(tc.ctx, tc.queryID)
			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)

			link := resp.ShareLink
			assert.Check(t, is.Equal(tc.queryID, link.ID))
			assert.Check(t, is.Equal(int64(1), link.Files.TotalCount))
			assert.Check(t, is.Equal(int64(0), link.ShareLinkAccessEvents.TotalCount))
		})
	}

	t.Run("List not authorized, view only user", func(t *testing.T) {
		_, err := suite.Client.API.GetAllShareLinks(th.SharedViewOnlyUser.UserCtx, nil, nil, nil, nil, nil)
		assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)

		resp, err := suite.Client.API.GetAllShareLinks(th.SharedAdminUser.UserCtx, nil, nil, nil, nil, nil)
		assert.NilError(t, err)
		assert.Check(t, resp.ShareLinks.TotalCount > 0)
	})

	(&th.Cleanup[*generated.ShareLinkDeleteOne]{Client: suite.Client.DB.ShareLink, ID: shareLinkID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestShareLinkURL(t *testing.T) {
	enableExternalSharing(t, th.SharedTestUser1)

	file := seedStoredFile(t, th.SharedTestUser1, storagetypes.R2Provider)

	concreteClient, ok := suite.Client.API.TestGraphClient.(*testclient.Client)
	assert.Assert(t, ok)

	testCases := []struct {
		name           string
		requireSession *bool
		expected       bool
	}{
		{name: "require session defaults to off"},
		{name: "require session round trips", requireSession: lo.ToPtr(true), expected: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vars := map[string]any{
				"input": testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, RequireSession: tc.requireSession},
			}

			var created createShareLinkURLResponse
			err := concreteClient.Client.Post(th.SharedTestUser1.UserCtx, "CreateShareLink", createShareLinkURLMutation, &created, vars)
			assert.NilError(t, err)

			link := created.CreateShareLink.ShareLink
			urlPrefix := coreutils.MockStorageLinksBaseURL + "/share/" + link.ID + "?token="

			assert.Check(t, strings.HasPrefix(lo.FromPtr(link.URL), urlPrefix), "createShareLink returns the share url")
			assert.Check(t, len(lo.FromPtr(link.URL)) > len(urlPrefix), "the share url carries the token")
			assert.Check(t, is.Equal(tc.expected, lo.FromPtr(link.RequireSession)))

			var fetched getShareLinkURLResponse
			err = concreteClient.Client.Post(th.SharedTestUser1.UserCtx, "GetShareLinkByID", getShareLinkURLQuery, &fetched, map[string]any{"shareLinkId": link.ID})
			assert.NilError(t, err)

			assert.Check(t, is.Equal(lo.FromPtr(link.URL), lo.FromPtr(fetched.ShareLink.URL)), "the queried url is the one returned at creation")
			assert.Check(t, is.Equal(tc.expected, lo.FromPtr(fetched.ShareLink.RequireSession)))

			// the url carries the share token, so another member who can read the link does not get it
			var other getShareLinkURLResponse
			err = concreteClient.Client.Post(th.SharedAdminUser.UserCtx, "GetShareLinkByID", getShareLinkURLQuery, &other, map[string]any{"shareLinkId": link.ID})
			assert.NilError(t, err)

			assert.Check(t, is.Equal(link.ID, other.ShareLink.ID))
			assert.Check(t, is.Equal("", lo.FromPtr(other.ShareLink.URL)), "only the creator is given the share url")

			(&th.Cleanup[*generated.ShareLinkDeleteOne]{Client: suite.Client.DB.ShareLink, ID: link.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}
}

func TestMutationCreateShareLink(t *testing.T) {
	enableExternalSharing(t, th.SharedTestUser1)

	file := seedStoredFile(t, th.SharedTestUser1, storagetypes.R2Provider)
	secondFile := seedStoredFile(t, th.SharedTestUser1, storagetypes.R2Provider)
	diskFile := seedStoredFile(t, th.SharedTestUser1, storagetypes.DiskProvider)
	otherOrgFile := seedStoredFile(t, th.SharedTestUser2, storagetypes.R2Provider)

	manager := suite.OrgMemberWithFunctionalRoles(t, th.SharedTestUser1, shareLinkManagerRole)
	managerFile := seedStoredFile(t, th.SharedTestUser1, storagetypes.R2Provider)
	grantFileEdit(t, th.SharedTestUser1, manager, managerFile.ID)

	// a member who can edit the file but does not hold the role
	grantFileEdit(t, th.SharedTestUser1, th.SharedViewOnlyUser, managerFile.ID)

	testCases := []struct {
		name                string
		request             testclient.CreateShareLinkInput
		ctx                 context.Context
		expectedLandingPage bool
		expectedMaxUses     int64
		expectedSession     bool
		expectedRecipients  []string
		expectedErr         string
	}{
		{
			name:            "happy path, single file is served directly",
			request:         testclient.CreateShareLinkInput{FileIDs: []string{file.ID}},
			ctx:             th.SharedTestUser1.UserCtx,
			expectedMaxUses: shareLinkDefaultUses,
		},
		{
			name:                "happy path, multiple files open the landing page",
			request:             testclient.CreateShareLinkInput{FileIDs: []string{file.ID, secondFile.ID}},
			ctx:                 th.SharedTestUser1.UserCtx,
			expectedLandingPage: true,
			expectedMaxUses:     shareLinkDefaultUses,
		},
		{
			name: "happy path, recipients and a use limit",
			request: testclient.CreateShareLinkInput{
				FileIDs:         []string{file.ID},
				RecipientEmails: []string{shareLinkRecipient},
				MaxUses:         lo.ToPtr(shareLinkUseLimit),
			},
			ctx:                 th.SharedTestUser1.UserCtx,
			expectedLandingPage: true,
			expectedMaxUses:     shareLinkUseLimit,
		},
		{
			name:                "happy path, landing page requested for a single file",
			request:             testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, LandingPage: lo.ToPtr(true)},
			ctx:                 th.SharedTestUser1.UserCtx,
			expectedLandingPage: true,
			expectedMaxUses:     shareLinkDefaultUses,
		},
		{
			name:                "happy path, a use limit opens the landing page",
			request:             testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, MaxUses: lo.ToPtr(shareLinkUseLimit)},
			ctx:                 th.SharedTestUser1.UserCtx,
			expectedLandingPage: true,
			expectedMaxUses:     shareLinkUseLimit,
		},
		{
			name:            "happy path, explicit unlimited uses are served directly",
			request:         testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, MaxUses: lo.ToPtr(shareLinkDefaultUses)},
			ctx:             th.SharedTestUser1.UserCtx,
			expectedMaxUses: shareLinkDefaultUses,
		},
		{
			name:                "happy path, recipient display name is dropped and the address lowercased",
			request:             testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, RecipientEmails: []string{"Bob <BOB@example.com>"}},
			ctx:                 th.SharedTestUser1.UserCtx,
			expectedLandingPage: true,
			expectedMaxUses:     shareLinkDefaultUses,
			expectedRecipients:  []string{"bob@example.com"},
		},
		{
			name:        "empty recipient",
			request:     testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, RecipientEmails: []string{""}},
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "mail: no address",
		},
		{
			name:        "blank recipient",
			request:     testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, RecipientEmails: []string{"  "}},
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "mail: no address",
		},
		{
			name:        "recipient that is not an email address",
			request:     testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, RecipientEmails: []string{shareLinkRecipient, "not-an-email"}},
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "mail: missing '@' or angle-addr",
		},
		{
			name:            "happy path, session required",
			request:         testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, RequireSession: lo.ToPtr(true)},
			ctx:             th.SharedTestUser1.UserCtx,
			expectedMaxUses: shareLinkDefaultUses,
			expectedSession: true,
		},
		{
			name:        "file is not stored on R2",
			request:     testclient.CreateShareLinkInput{FileIDs: []string{diskFile.ID}},
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: hooks.ErrInvalidInput.Error(),
		},
		{
			name:        "file belongs to another organization",
			request:     testclient.CreateShareLinkInput{FileIDs: []string{otherOrgFile.ID}},
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "expiry beyond the maximum duration",
			request:     testclient.CreateShareLinkInput{FileIDs: []string{file.ID}, ExpiresAt: lo.ToPtr(time.Now().Add(shareLinkExpiryTooFar))},
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: hooks.ErrShareLinkExpiryTooLong.Error(),
		},
		{
			name:        "external sharing is not enabled for the organization",
			request:     testclient.CreateShareLinkInput{FileIDs: []string{otherOrgFile.ID}},
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: hooks.ErrShareLinkExternalSharingDisabled.Error(),
		},
		{
			name:        "not authorized, view only user",
			request:     testclient.CreateShareLinkInput{FileIDs: []string{file.ID}},
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:            "happy path, share link manager who can edit the file",
			request:         testclient.CreateShareLinkInput{FileIDs: []string{managerFile.ID}},
			ctx:             manager.UserCtx,
			expectedMaxUses: shareLinkDefaultUses,
		},
		{
			name:        "not authorized, share link manager who cannot edit the file",
			request:     testclient.CreateShareLinkInput{FileIDs: []string{file.ID}},
			ctx:         manager.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "not authorized, file editor without the share link manager role",
			request:     testclient.CreateShareLinkInput{FileIDs: []string{managerFile.ID}},
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.CreateShareLink(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)

			link := resp.CreateShareLink.ShareLink
			assert.Check(t, is.Equal(tc.expectedLandingPage, lo.FromPtr(link.LandingPage)))
			assert.Check(t, is.Equal(tc.expectedMaxUses, lo.FromPtr(link.MaxUses)))
			assert.Check(t, is.Equal(tc.expectedSession, lo.FromPtr(link.RequireSession)))
			assert.Check(t, is.Equal(int64(0), link.UseCount))
			assert.Check(t, is.DeepEqual(lo.Ternary(tc.expectedRecipients != nil, tc.expectedRecipients, tc.request.RecipientEmails), link.RecipientEmails))
			assert.Check(t, is.Equal(int64(len(tc.request.FileIDs)), link.Files.TotalCount))
			assert.Check(t, lo.FromPtr(link.ExpiresAt).After(time.Now()))
			assert.Check(t, link.RevokedAt == nil)

			(&th.Cleanup[*generated.ShareLinkDeleteOne]{Client: suite.Client.DB.ShareLink, ID: link.ID}).MustDelete(tc.ctx, t)
		})
	}
}

func TestMutationUpdateShareLink(t *testing.T) {
	shareLinkID := seedShareLinkWithUses(t, th.SharedTestUser1, lo.ToPtr(shareLinkSingleUse))
	directLinkID := seedShareLink(t, th.SharedTestUser1)
	revokedAt := time.Now()

	testCases := []struct {
		name        string
		linkID      string
		request     testclient.UpdateShareLinkInput
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path, raise the use limit",
			request: testclient.UpdateShareLinkInput{MaxUses: lo.ToPtr(shareLinkUseLimit)},
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, revoke",
			request: testclient.UpdateShareLinkInput{RevokedAt: &revokedAt},
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:        "a revocation cannot be undone",
			request:     testclient.UpdateShareLinkInput{ClearRevokedAt: lo.ToPtr(true)},
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: hooks.ErrShareLinkRevocationPermanent.Error(),
		},
		{
			name:        "a use limit on a link that serves its file directly",
			linkID:      directLinkID,
			request:     testclient.UpdateShareLinkInput{MaxUses: lo.ToPtr(shareLinkUseLimit)},
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: hooks.ErrShareLinkDirectUseCap.Error(),
		},
		{
			name:        "share link in another organization",
			request:     testclient.UpdateShareLinkInput{RevokedAt: &revokedAt},
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not authorized, view only user",
			request:     testclient.UpdateShareLinkInput{RevokedAt: &revokedAt},
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.UpdateShareLink(tc.ctx, lo.CoalesceOrEmpty(tc.linkID, shareLinkID), tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)

			link := resp.UpdateShareLink.ShareLink
			assert.Check(t, is.Equal(shareLinkUseLimit, lo.FromPtr(link.MaxUses)))
			assert.Check(t, is.Equal(tc.request.RevokedAt != nil, link.RevokedAt != nil))
		})
	}

	(&th.Cleanup[*generated.ShareLinkDeleteOne]{Client: suite.Client.DB.ShareLink, IDs: []string{shareLinkID, directLinkID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryShareLinkAccessEvents(t *testing.T) {
	shareLinkID := seedShareLink(t, th.SharedTestUser1)
	internalCtx := th.SetInternalContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	link, err := suite.Client.DB.ShareLink.Get(internalCtx, shareLinkID)
	assert.NilError(t, err)

	fileID, err := link.QueryFiles().OnlyID(internalCtx)
	assert.NilError(t, err)

	// access events are only written by the consume endpoint as an internal caller
	err = suite.Client.DB.ShareLinkAccessEvent.Create().
		SetShareLinkID(shareLinkID).
		SetOwnerID(th.SharedTestUser1.OrganizationID).
		SetFileID(fileID).
		SetIPAddress("203.0.113.10").
		SetUserAgent("test-agent").
		Exec(internalCtx)
	assert.NilError(t, err)

	testCases := []struct {
		name          string
		ctx           context.Context
		expectedCount int
		errorMsg      string
	}{
		{name: "creator", ctx: th.SharedTestUser1.UserCtx, expectedCount: 1},
		{name: "another member with edit access", ctx: th.SharedAdminUser.UserCtx, expectedCount: 1},
		{name: "member with the share link manager role", ctx: suite.OrgMemberWithFunctionalRoles(t, th.SharedTestUser1, shareLinkManagerRole).UserCtx, expectedCount: 1},
		{name: "another organization", ctx: th.SharedTestUser2.UserCtx},
		{name: "view only user", ctx: th.SharedViewOnlyUser.UserCtx, errorMsg: th.NotAuthorizedErrorMsg},
	}

	for _, tc := range testCases {
		t.Run("Query "+tc.name, func(t *testing.T) {
			events, err := suite.Client.DB.ShareLinkAccessEvent.Query().
				Where(sharelinkaccessevent.ShareLinkID(shareLinkID)).
				All(tc.ctx)
			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Len(events, 0))

				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Len(events, tc.expectedCount))
		})
	}

	t.Run("Get through the share link", func(t *testing.T) {
		resp, err := suite.Client.API.GetShareLinkByID(th.SharedTestUser1.UserCtx, shareLinkID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(int64(1), resp.ShareLink.ShareLinkAccessEvents.TotalCount))

		_, err = suite.Client.API.GetShareLinkByID(th.SharedViewOnlyUser.UserCtx, shareLinkID)
		assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)
	})

	(&th.Cleanup[*generated.ShareLinkDeleteOne]{Client: suite.Client.DB.ShareLink, ID: shareLinkID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteShareLink(t *testing.T) {
	shareLinkID := seedShareLink(t, th.SharedTestUser1)

	testCases := []struct {
		name        string
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "share link in another organization",
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not authorized, view only user",
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "happy path",
			ctx:  th.SharedTestUser1.UserCtx,
		},
		{
			name:        "already deleted",
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.DeleteShareLink(tc.ctx, shareLinkID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(shareLinkID, resp.DeleteShareLink.DeletedID))
		})
	}
}
