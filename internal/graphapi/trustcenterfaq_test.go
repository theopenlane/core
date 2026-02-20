package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestMutationCreateTrustCenterFAQ(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)

	viewOnlyUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &viewOnlyUser, enums.RoleMember, testUser.OrganizationID)

	adminUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &adminUser, enums.RoleAdmin, testUser.OrganizationID)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	note1 := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	note2 := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	note3 := (&NoteBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.CreateTrustCenterFAQInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path - admin can create trust center faq",
			request: testclient.CreateTrustCenterFAQInput{
				NoteID:        note1.ID,
				TrustCenterID: &trustCenter.ID,
				ReferenceLink: lo.ToPtr("https://example.com/faq"),
				DisplayOrder:  lo.ToPtr(int64(1)),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "not authorized - view only user cannot create",
			request: testclient.CreateTrustCenterFAQInput{
				NoteID:        note2.ID,
				TrustCenterID: &trustCenter.ID,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - different org user cannot create",
			request: testclient.CreateTrustCenterFAQInput{
				NoteID:        note3.ID,
				TrustCenterID: &trustCenter.ID,
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "trust center not found",
			request: testclient.CreateTrustCenterFAQInput{
				NoteID:        note2.ID,
				TrustCenterID: lo.ToPtr("non-existent-trust-center-id"),
			},
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "note not found",
			request: testclient.CreateTrustCenterFAQInput{
				NoteID:        "non-existent-note-id",
				TrustCenterID: &trustCenter.ID,
			},
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTrustCenterFaq(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.CreateTrustCenterFaq.TrustCenterFaq.ID != "")
			if tc.request.TrustCenterID != nil {
				assert.Check(t, is.Equal(*tc.request.TrustCenterID, *resp.CreateTrustCenterFaq.TrustCenterFaq.TrustCenterID))
			}
			if tc.request.ReferenceLink != nil {
				assert.Check(t, is.Equal(*tc.request.ReferenceLink, *resp.CreateTrustCenterFaq.TrustCenterFaq.ReferenceLink))
			}
			if tc.request.DisplayOrder != nil {
				assert.Check(t, is.Equal(*tc.request.DisplayOrder, *resp.CreateTrustCenterFaq.TrustCenterFaq.DisplayOrder))
			}

			(&Cleanup[*generated.TrustCenterFAQDeleteOne]{client: suite.client.db.TrustCenterFAQ, ID: resp.CreateTrustCenterFaq.TrustCenterFaq.ID}).MustDelete(tc.ctx, t)
		})
	}

	// clean up
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateTrustCenterFAQAsAnonymousUser(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	note := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	testCases := []struct {
		name           string
		request        testclient.CreateTrustCenterFAQInput
		trustCenterID  string
		organizationID string
		client         *testclient.TestClient
		expectedErr    string
	}{
		{
			name: "anonymous user cannot create trust center faq",
			request: testclient.CreateTrustCenterFAQInput{
				NoteID:        note.ID,
				TrustCenterID: &trustCenter.ID,
			},
			trustCenterID:  trustCenter.ID,
			organizationID: testUser.OrganizationID,
			client:         suite.client.api,
			expectedErr:    "could not identify authenticated user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			anonCtx := createAnonymousTrustCenterContext(tc.trustCenterID, tc.organizationID)

			resp, err := tc.client.CreateTrustCenterFaq(anonCtx, tc.request)

			assert.ErrorContains(t, err, tc.expectedErr)
			assert.Check(t, resp.CreateTrustCenterFaq.TrustCenterFaq.ID == "")
		})
	}

	// clean up
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
}

func TestQueryTrustCenterFAQByID(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)
	viewOnlyUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &viewOnlyUser, enums.RoleMember, testUser.OrganizationID)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	note1 := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	createResp, err := suite.client.api.CreateTrustCenterFaq(testUser.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note1.ID,
		TrustCenterID: &trustCenter.ID,
		ReferenceLink: lo.ToPtr("https://example.com/faq"),
		DisplayOrder:  lo.ToPtr(int64(1)),
	})
	assert.NilError(t, err)
	tcFAQ := createResp.CreateTrustCenterFaq.TrustCenterFaq

	// create faq for a different org
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	note2 := (&NoteBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	createResp2, err := suite.client.api.CreateTrustCenterFaq(testUser2.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note2.ID,
		TrustCenterID: &trustCenter2.ID,
	})
	assert.NilError(t, err)
	tcFAQ2 := createResp2.CreateTrustCenterFaq.TrustCenterFaq

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path - get trust center faq",
			queryID: tcFAQ.ID,
			client:  suite.client.api,
			ctx:     testUser.UserCtx,
		},
		{
			name:    "happy path - view only user can get trust center faq",
			queryID: tcFAQ.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path - anon user",
			queryID: tcFAQ.ID,
			client:  suite.client.api,
			ctx:     createAnonymousTrustCenterContext(trustCenter.ID, testUser.OrganizationID),
		},
		{
			name:     "not found - different org user cannot access trust center faq",
			queryID:  tcFAQ.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found - different anonymous user cannot access trust center faq",
			queryID:  tcFAQ.ID,
			client:   suite.client.api,
			ctx:      createAnonymousTrustCenterContext(trustCenter2.ID, testUser2.OrganizationID),
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found - non-existent ID",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      testUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterFAQByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.queryID, resp.TrustCenterFaq.ID))
		})
	}

	// clean up
	(&Cleanup[*generated.TrustCenterFAQDeleteOne]{client: suite.client.db.TrustCenterFAQ, ID: tcFAQ.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterFAQDeleteOne]{client: suite.client.db.TrustCenterFAQ, ID: tcFAQ2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationUpdateTrustCenterFAQ(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)

	viewOnlyUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &viewOnlyUser, enums.RoleMember, testUser.OrganizationID)

	adminUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &adminUser, enums.RoleAdmin, testUser.OrganizationID)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	noteOtherOrg := (&NoteBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	createResp2, err := suite.client.api.CreateTrustCenterFaq(testUser2.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        noteOtherOrg.ID,
		TrustCenterID: &trustCenter2.ID,
	})
	assert.NilError(t, err)
	tcFAQ2 := createResp2.CreateTrustCenterFaq.TrustCenterFaq

	testCases := []struct {
		name        string
		setupFunc   func() string
		request     testclient.UpdateTrustCenterFAQInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path - update reference link and display order",
			setupFunc: func() string {
				note := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterFaq(testUser.UserCtx, testclient.CreateTrustCenterFAQInput{
					NoteID:        note.ID,
					TrustCenterID: &trustCenter.ID,
					ReferenceLink: lo.ToPtr("https://example.com/old"),
					DisplayOrder:  lo.ToPtr(int64(1)),
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterFaq.TrustCenterFaq.ID
			},
			request: testclient.UpdateTrustCenterFAQInput{
				ReferenceLink: lo.ToPtr("https://example.com/new"),
				DisplayOrder:  lo.ToPtr(int64(5)),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path - clear reference link",
			setupFunc: func() string {
				note := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterFaq(testUser.UserCtx, testclient.CreateTrustCenterFAQInput{
					NoteID:        note.ID,
					TrustCenterID: &trustCenter.ID,
					ReferenceLink: lo.ToPtr("https://example.com/to-clear"),
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterFaq.TrustCenterFaq.ID
			},
			request: testclient.UpdateTrustCenterFAQInput{
				ClearReferenceLink: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    testUser.UserCtx,
		},
		{
			name: "not authorized - view only user cannot update",
			setupFunc: func() string {
				note := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterFaq(testUser.UserCtx, testclient.CreateTrustCenterFAQInput{
					NoteID:        note.ID,
					TrustCenterID: &trustCenter.ID,
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterFaq.TrustCenterFaq.ID
			},
			request: testclient.UpdateTrustCenterFAQInput{
				ReferenceLink: lo.ToPtr("https://example.com/updated"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - anon user cannot update",
			setupFunc: func() string {
				note := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterFaq(testUser.UserCtx, testclient.CreateTrustCenterFAQInput{
					NoteID:        note.ID,
					TrustCenterID: &trustCenter.ID,
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterFaq.TrustCenterFaq.ID
			},
			request: testclient.UpdateTrustCenterFAQInput{
				ReferenceLink: lo.ToPtr("https://example.com/updated"),
			},
			client:      suite.client.api,
			ctx:         createAnonymousTrustCenterContext(trustCenter.ID, testUser.OrganizationID),
			expectedErr: "could not identify authenticated user",
		},
		{
			name: "not authorized - different org user cannot update",
			setupFunc: func() string {
				note := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterFaq(testUser.UserCtx, testclient.CreateTrustCenterFAQInput{
					NoteID:        note.ID,
					TrustCenterID: &trustCenter.ID,
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterFaq.TrustCenterFaq.ID
			},
			request: testclient.UpdateTrustCenterFAQInput{
				ReferenceLink: lo.ToPtr("https://example.com/updated"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:      "not found - non-existent ID",
			setupFunc: func() string { return "non-existent-id" },
			request: testclient.UpdateTrustCenterFAQInput{
				ReferenceLink: lo.ToPtr("https://example.com/updated"),
			},
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	var createdIDs []string

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			id := tc.setupFunc()
			if tc.expectedErr == "" {
				createdIDs = append(createdIDs, id)
			}

			resp, err := tc.client.UpdateTrustCenterFaq(tc.ctx, id, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(id, resp.UpdateTrustCenterFaq.TrustCenterFaq.ID))

			if tc.request.ReferenceLink != nil {
				assert.Check(t, is.Equal(*tc.request.ReferenceLink, *resp.UpdateTrustCenterFaq.TrustCenterFaq.ReferenceLink))
			}
			if tc.request.DisplayOrder != nil {
				assert.Check(t, is.Equal(*tc.request.DisplayOrder, *resp.UpdateTrustCenterFaq.TrustCenterFaq.DisplayOrder))
			}
			if tc.request.ClearReferenceLink != nil && *tc.request.ClearReferenceLink {
				assert.Check(t, resp.UpdateTrustCenterFaq.TrustCenterFaq.ReferenceLink == nil)
			}
		})
	}

	// clean up
	if len(createdIDs) > 0 {
		(&Cleanup[*generated.TrustCenterFAQDeleteOne]{client: suite.client.db.TrustCenterFAQ, IDs: createdIDs}).MustDelete(testUser.UserCtx, t)
	}
	(&Cleanup[*generated.TrustCenterFAQDeleteOne]{client: suite.client.db.TrustCenterFAQ, ID: tcFAQ2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationDeleteTrustCenterFAQ(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)

	viewOnlyUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &viewOnlyUser, enums.RoleMember, testUser.OrganizationID)

	adminUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &adminUser, enums.RoleAdmin, testUser.OrganizationID)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	note1 := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	note2 := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	createResp1, err := suite.client.api.CreateTrustCenterFaq(testUser.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note1.ID,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)
	tcFAQ1 := createResp1.CreateTrustCenterFaq.TrustCenterFaq

	createResp2, err := suite.client.api.CreateTrustCenterFaq(testUser.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note2.ID,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)
	tcFAQ2 := createResp2.CreateTrustCenterFaq.TrustCenterFaq

	testUserAnother := suite.userBuilder(context.Background(), t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUserAnother.UserCtx, t)
	note3 := (&NoteBuilder{client: suite.client}).MustNew(testUserAnother.UserCtx, t)

	createResp3, err := suite.client.api.CreateTrustCenterFaq(testUserAnother.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note3.ID,
		TrustCenterID: &trustCenter2.ID,
	})
	assert.NilError(t, err)
	tcFAQ3 := createResp3.CreateTrustCenterFaq.TrustCenterFaq

	testCases := []struct {
		name        string
		id          string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:   "happy path - delete trust center faq",
			id:     tcFAQ1.ID,
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:        "not authorized - view only user cannot delete",
			id:          tcFAQ2.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not authorized - different org user cannot delete",
			id:          tcFAQ3.ID,
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not found - non-existent ID",
			id:          "non-existent-id",
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustCenterFaq(tc.ctx, tc.id)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.id, resp.DeleteTrustCenterFaq.DeletedID))

			_, err = tc.client.GetTrustCenterFAQByID(tc.ctx, tc.id)
			assert.ErrorContains(t, err, notFoundErrorMsg)
		})
	}

	// clean up
	(&Cleanup[*generated.TrustCenterFAQDeleteOne]{client: suite.client.db.TrustCenterFAQ, ID: tcFAQ2.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterFAQDeleteOne]{client: suite.client.db.TrustCenterFAQ, ID: tcFAQ3.ID}).MustDelete(testUserAnother.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUserAnother.UserCtx, t)
}

func TestQueryTrustCenterFAQs(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)

	viewOnlyUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &viewOnlyUser, enums.RoleMember, testUser.OrganizationID)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	note1 := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	note2 := (&NoteBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	createResp1, err := suite.client.api.CreateTrustCenterFaq(testUser.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note1.ID,
		TrustCenterID: &trustCenter.ID,
		ReferenceLink: lo.ToPtr("https://example.com/faq1"),
	})
	assert.NilError(t, err)
	tcFAQ1 := createResp1.CreateTrustCenterFaq.TrustCenterFaq

	createResp2, err := suite.client.api.CreateTrustCenterFaq(testUser.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note2.ID,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)
	tcFAQ2 := createResp2.CreateTrustCenterFaq.TrustCenterFaq

	testUserAnother := suite.userBuilder(context.Background(), t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUserAnother.UserCtx, t)
	note3 := (&NoteBuilder{client: suite.client}).MustNew(testUserAnother.UserCtx, t)

	createResp3, err := suite.client.api.CreateTrustCenterFaq(testUserAnother.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note3.ID,
		TrustCenterID: &trustCenter2.ID,
	})
	assert.NilError(t, err)
	tcFAQ3 := createResp3.CreateTrustCenterFaq.TrustCenterFaq

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		where           *testclient.TrustCenterFAQWhereInput
	}{
		{
			name:            "get all trust center faqs for user1",
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "get all trust center faqs for another user",
			client:          suite.client.api,
			ctx:             testUserAnother.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "view only user can see trust center faqs",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "anonymous user can see trust center faqs",
			client:          suite.client.api,
			ctx:             createAnonymousTrustCenterContext(trustCenter.ID, testUser.OrganizationID),
			expectedResults: 2,
		},
		{
			name:   "filter by trust center ID",
			client: suite.client.api,
			ctx:    testUser.UserCtx,
			where: &testclient.TrustCenterFAQWhereInput{
				TrustCenterID: &trustCenter.ID,
			},
			expectedResults: 2,
		},
		{
			name:   "filter by reference link",
			client: suite.client.api,
			ctx:    testUser.UserCtx,
			where: &testclient.TrustCenterFAQWhereInput{
				ReferenceLinkNotNil: lo.ToPtr(true),
			},
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Query "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterFaqs(tc.ctx, nil, nil, nil, nil, nil, tc.where)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.expectedResults, resp.TrustCenterFAQs.TotalCount))

			if tc.expectedResults > 0 {
				assert.Check(t, is.Len(resp.TrustCenterFAQs.Edges, int(tc.expectedResults)))
				for _, edge := range resp.TrustCenterFAQs.Edges {
					assert.Check(t, edge.Node.ID != "")
				}
			}
		})
	}

	// clean up
	(&Cleanup[*generated.TrustCenterFAQDeleteOne]{client: suite.client.db.TrustCenterFAQ, IDs: []string{tcFAQ1.ID, tcFAQ2.ID}}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterFAQDeleteOne]{client: suite.client.db.TrustCenterFAQ, ID: tcFAQ3.ID}).MustDelete(testUserAnother.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUserAnother.UserCtx, t)
}
