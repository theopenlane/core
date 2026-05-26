package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestMutationCreateTrustCenterFAQ(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	tcOrg2 := createFreshOrgWithTrustCenter(t)

	note1 := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	note2 := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	note3 := (&NoteBuilder{client: suite.client}).MustNew(tcOrg2.owner.UserCtx, t)

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
			ctx:    tcOrg.admin.UserCtx,
		},
		{
			name: "not authorized - view only user cannot create",
			request: testclient.CreateTrustCenterFAQInput{
				NoteID:        note2.ID,
				TrustCenterID: &trustCenter.ID,
			},
			client:      suite.client.api,
			ctx:         tcOrg.member.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - different org user cannot create",
			request: testclient.CreateTrustCenterFAQInput{
				NoteID:        note3.ID,
				TrustCenterID: &trustCenter.ID,
			},
			client:      suite.client.api,
			ctx:         tcOrg2.owner.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "trust center not found",
			request: testclient.CreateTrustCenterFAQInput{
				NoteID:        note2.ID,
				TrustCenterID: lo.ToPtr("non-existent-trust-center-id"),
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "note not found, invalid ulid",
			request: testclient.CreateTrustCenterFAQInput{
				NoteID:        "non-existent-note-id",
				TrustCenterID: &trustCenter.ID,
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: invalidInputErrorMsg,
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
		})
	}

	// clean up
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestMutationCreateTrustCenterFAQAsAnonymousUser(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	note := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)

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
			organizationID: tcOrg.organizationID,
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
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestQueryTrustCenterFAQByID(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t, withAllUserTypes())
	trustCenter := tcOrg.trustCenter

	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.trustCenter

	createResp, err := suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
		TrustCenterID: &trustCenter.ID,
		ReferenceLink: lo.ToPtr("https://example.com/faq"),
		DisplayOrder:  lo.ToPtr(int64(1)),
		CreateNote: &testclient.CreateNoteInput{
			Text: "faq note for trust center",
		},
	})
	assert.NilError(t, err)
	tcFAQ := createResp.CreateTrustCenterFaq.TrustCenterFaq

	_, err = suite.client.api.CreateTrustCenterFaq(tcOrg2.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
		TrustCenterID: &trustCenter2.ID,
		CreateNote: &testclient.CreateNoteInput{
			Text: "faq note for trust center 2",
		},
	})
	assert.NilError(t, err)

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
			ctx:     tcOrg.admin.UserCtx,
		},
		{
			name:    "happy path - view only user can get trust center faq",
			queryID: tcFAQ.ID,
			client:  suite.client.api,
			ctx:     tcOrg.member.UserCtx,
		},
		{
			name:    "happy path - anon user",
			queryID: tcFAQ.ID,
			client:  suite.client.api,
			ctx:     createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.organizationID),
		},
		{
			name:     "not found - different org user cannot access trust center faq",
			queryID:  tcFAQ.ID,
			client:   suite.client.api,
			ctx:      tcOrg2.owner.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found - different anonymous user cannot access trust center faq",
			queryID:  tcFAQ.ID,
			client:   suite.client.api,
			ctx:      createAnonymousTrustCenterContext(trustCenter2.ID, tcOrg2.organizationID),
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found - non-existent ID",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      tcOrg.superAdmin.UserCtx,
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
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestMutationUpdateTrustCenterFAQ(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.trustCenter

	noteOtherOrg := (&NoteBuilder{client: suite.client}).MustNew(tcOrg2.owner.UserCtx, t)

	_, err := suite.client.api.CreateTrustCenterFaq(tcOrg2.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        noteOtherOrg.ID,
		TrustCenterID: &trustCenter2.ID,
	})
	assert.NilError(t, err)

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
				note := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
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
			ctx:    tcOrg.admin.UserCtx,
		},
		{
			name: "happy path - clear reference link",
			setupFunc: func() string {
				note := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
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
			ctx:    tcOrg.owner.UserCtx,
		},
		{
			name: "not authorized - view only user cannot update",
			setupFunc: func() string {
				note := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
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
			ctx:         tcOrg.member.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - anon user cannot update",
			setupFunc: func() string {
				note := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
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
			ctx:         createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.organizationID),
			expectedErr: "could not identify authenticated user",
		},
		{
			name: "not authorized - different org user cannot update",
			setupFunc: func() string {
				note := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
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
			ctx:         tcOrg2.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:      "not found - non-existent ID",
			setupFunc: func() string { return "non-existent-id" },
			request: testclient.UpdateTrustCenterFAQInput{
				ReferenceLink: lo.ToPtr("https://example.com/updated"),
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			id := tc.setupFunc()

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
				assert.Check(t, resp.UpdateTrustCenterFaq.TrustCenterFaq.ReferenceLink == nil ||
					*resp.UpdateTrustCenterFaq.TrustCenterFaq.ReferenceLink == "")
			}
		})
	}

	// clean up
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestMutationDeleteTrustCenterFAQ(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	note1 := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	note2 := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)

	createResp1, err := suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note1.ID,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)
	tcFAQ1 := createResp1.CreateTrustCenterFaq.TrustCenterFaq

	createResp2, err := suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note2.ID,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)
	tcFAQ2 := createResp2.CreateTrustCenterFaq.TrustCenterFaq

	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.trustCenter
	note3 := (&NoteBuilder{client: suite.client}).MustNew(tcOrg2.owner.UserCtx, t)

	createResp3, err := suite.client.api.CreateTrustCenterFaq(tcOrg2.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
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
			ctx:    tcOrg.admin.UserCtx,
		},
		{
			name:        "not authorized - view only user cannot delete",
			id:          tcFAQ2.ID,
			client:      suite.client.api,
			ctx:         tcOrg.member.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not authorized - different org user cannot delete",
			id:          tcFAQ3.ID,
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not found - non-existent ID",
			id:          "non-existent-id",
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
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
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestQueryTrustCenterFAQs(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.trustCenter

	note1 := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	note2 := (&NoteBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)

	_, err := suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note1.ID,
		TrustCenterID: &trustCenter.ID,
		ReferenceLink: lo.ToPtr("https://example.com/faq1"),
	})
	assert.NilError(t, err)

	_, err = suite.client.api.CreateTrustCenterFaq(tcOrg.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note2.ID,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	note3 := (&NoteBuilder{client: suite.client}).MustNew(tcOrg2.owner.UserCtx, t)

	_, err = suite.client.api.CreateTrustCenterFaq(tcOrg2.owner.UserCtx, testclient.CreateTrustCenterFAQInput{
		NoteID:        note3.ID,
		TrustCenterID: &trustCenter2.ID,
	})
	assert.NilError(t, err)

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
			ctx:             tcOrg.owner.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "get all trust center faqs for another user",
			client:          suite.client.api,
			ctx:             tcOrg2.owner.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "view only user can see trust center faqs",
			client:          suite.client.api,
			ctx:             tcOrg.member.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "anonymous user can see trust center faqs",
			client:          suite.client.api,
			ctx:             createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.organizationID),
			expectedResults: 2,
		},
		{
			name:   "filter by trust center ID",
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
			where: &testclient.TrustCenterFAQWhereInput{
				TrustCenterID: &trustCenter.ID,
			},
			expectedResults: 2,
		},
		{
			name:   "filter by reference link",
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
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
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}
