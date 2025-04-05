package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryControlImplementation() {
	t := suite.T()

	// create an controlImplementation1 to be queried using testUser1
	controlImplementation1 := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create another with associated controls in another org
	control1 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlImplementation2 := (&ControlImplementationBuilder{client: suite.client, ControlIDs: []string{control1.ID}}).MustNew(testUser2.UserCtx, t)

	// create a controlImplementation with controls and subcontrols
	control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subcontrol1 := (&SubcontrolBuilder{client: suite.client, ControlID: control2.ID}).MustNew(testUser1.UserCtx, t)
	subcontrol2 := (&SubcontrolBuilder{client: suite.client, ControlID: control2.ID}).MustNew(testUser1.UserCtx, t)
	controlImplementation3 := (&ControlImplementationBuilder{client: suite.client, ControlIDs: []string{control2.ID}, SubcontrolIDs: []string{subcontrol1.ID, subcontrol2.ID}}).MustNew(testUser1.UserCtx, t)

	// give viewOnlyUser access to the parent control via a group
	// this will give the user access to the controlImplementation as well
	groupViewer := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, GroupID: groupViewer.ID, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	control3 := (&ControlBuilder{client: suite.client, ControlViewerGroupID: groupViewer.ID}).MustNew(testUser1.UserCtx, t)
	controlImplementation4 := (&ControlImplementationBuilder{client: suite.client, ControlIDs: []string{control3.ID}}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the controlImplementation
	testCases := []struct {
		name                  string
		queryID               string
		client                *openlaneclient.OpenlaneClient
		ctx                   context.Context
		shouldHaveControls    bool
		shouldHaveSubcontrols bool
		errorMsg              string
	}{
		{
			name:    "happy path",
			queryID: controlImplementation1.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:               "happy path, controlImplementation with associated controls",
			queryID:            controlImplementation2.ID,
			client:             suite.client.api,
			ctx:                testUser2.UserCtx,
			shouldHaveControls: true,
		},
		{
			name:     "controlImplementation with associated controls, but no access",
			queryID:  controlImplementation2.ID,
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:                  "happy path, controlImplementation with associated controls and subcontrols",
			queryID:               controlImplementation3.ID,
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
			shouldHaveControls:    true,
			shouldHaveSubcontrols: true,
		},
		{
			name:     "controlImplementation with associated controls, but no access",
			queryID:  controlImplementation3.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "happy path, read only user, no access to control or controlImplementation",
			queryID:  controlImplementation1.ID,
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:               "happy path, controlImplementation with associated controls and group viewer",
			queryID:            controlImplementation4.ID,
			client:             suite.client.api,
			ctx:                testUser1.UserCtx,
			shouldHaveControls: true,
		},
		{
			name:               "happy path, controlImplementation with associated controls and group viewer",
			queryID:            controlImplementation4.ID,
			client:             suite.client.api,
			ctx:                viewOnlyUser.UserCtx,
			shouldHaveControls: true,
		},
		{
			name:    "happy path using personal access token",
			queryID: controlImplementation1.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "controlImplementation not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "controlImplementation not found, using not authorized user",
			queryID:  controlImplementation1.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetControlImplementationByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.ControlImplementation)

			assert.Equal(t, tc.queryID, resp.ControlImplementation.ID)

			// check if the controlImplementation has the expected fields
			assert.NotEmpty(t, resp.ControlImplementation.ID)

			// model tests set up defaults here
			assert.NotEmpty(t, resp.ControlImplementation.Details)
			assert.NotEmpty(t, resp.ControlImplementation.ImplementationDate)

			if tc.shouldHaveControls {
				assert.NotEmpty(t, resp.ControlImplementation.Controls)
			}

			if tc.shouldHaveSubcontrols {
				assert.NotEmpty(t, resp.ControlImplementation.Subcontrols)
				assert.Len(t, resp.ControlImplementation.Subcontrols.Edges, 2)
				assert.Equal(t, subcontrol1.ID, resp.ControlImplementation.Subcontrols.Edges[0].Node.ID)
				assert.Equal(t, subcontrol2.ID, resp.ControlImplementation.Subcontrols.Edges[1].Node.ID)
			}
		})
	}
}

func (suite *GraphTestSuite) TestQuerycontrolImplementations() {
	t := suite.T()

	// create multiple controlImplementations to be queried using testUser1
	numCIs := 5
	for range numCIs {
		(&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	}

	// give viewOnlyUser access to the parent control via a group
	// this will give the user access to the controlImplementation as well
	groupViewer := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, GroupID: groupViewer.ID, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	numCIsWithGroupViewer := 2
	for range numCIsWithGroupViewer {
		control1 := (&ControlBuilder{client: suite.client, ControlViewerGroupID: groupViewer.ID}).MustNew(testUser1.UserCtx, t)
		(&ControlImplementationBuilder{client: suite.client, ControlIDs: []string{control1.ID}}).MustNew(testUser1.UserCtx, t)
	}

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: numCIs + numCIsWithGroupViewer,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: numCIsWithGroupViewer,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 0, // api token should not return any results because they have no access to the controls linked
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: numCIs + numCIsWithGroupViewer,
		},
		{
			name:            "another user, no controlImplementations should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllControlImplementations(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.ControlImplementations.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateControlImplementation() {
	t := suite.T()

	yesterday := time.Now().Add(-time.Hour * 24)

	groupEditor := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add viewOnlyUser to the group with editor permissions
	(&GroupMemberBuilder{client: suite.client, GroupID: groupEditor.ID, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	numControls := 2
	controlIDs := []string{}
	for range numControls {
		// create controls where the groupEditor has editor permissions
		control := (&ControlBuilder{client: suite.client, ControlEditorGroupID: groupEditor.ID}).MustNew(testUser1.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	// create one control without group permissions
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	allControlIDs := append(controlIDs, control.ID)

	numSubcontrols := 2
	subcontrolIDs := []string{}
	for range numSubcontrols {
		subcontrol := (&SubcontrolBuilder{client: suite.client, ControlID: controlIDs[0]}).MustNew(testUser1.UserCtx, t)
		subcontrolIDs = append(subcontrolIDs, subcontrol.ID)
	}

	testCases := []struct {
		name        string
		request     openlaneclient.CreateControlImplementationInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path, minimal input",
			request: openlaneclient.CreateControlImplementationInput{
				// there are no required fields in the model, you can create a blank implementation
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateControlImplementationInput{
				Details:            lo.ToPtr(gofakeit.Paragraph(3, 5, 30, "<br />")),
				Status:             &enums.DocumentNeedsApproval,
				ImplementationDate: &yesterday,
				Verified:           lo.ToPtr(true),
				VerificationDate:   &yesterday,
				Tags:               []string{"control-imp", "ref-code"},
				ControlIDs:         controlIDs,
				SubcontrolIDs:      subcontrolIDs,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateControlImplementationInput{
				Details: lo.ToPtr(gofakeit.Paragraph(3, 5, 30, "<br />")),
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateControlImplementationInput{
				Details: lo.ToPtr(gofakeit.Paragraph(3, 5, 30, "<br />")),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateControlImplementationInput{
				Details: lo.ToPtr(gofakeit.Paragraph(3, 5, 30, "<br />")),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user authorized because they have editor permissions to all the parent control",
			request: openlaneclient.CreateControlImplementationInput{
				Details:    lo.ToPtr(gofakeit.Paragraph(3, 5, 30, "<br />")),
				ControlIDs: controlIDs,
			},
			client: suite.client.api,
			ctx:    viewOnlyUser.UserCtx,
		},
		{
			name: "user not authorized, not enough permissions to one of the parent controls",
			request: openlaneclient.CreateControlImplementationInput{
				Details:    lo.ToPtr(gofakeit.Paragraph(3, 5, 30, "<br />")),
				ControlIDs: allControlIDs,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "no access to linked control",
			request: openlaneclient.CreateControlImplementationInput{
				Details:    lo.ToPtr(gofakeit.Paragraph(3, 5, 30, "<br />")),
				ControlIDs: controlIDs,
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateControlImplementation(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check default fields
			assert.NotEmpty(t, resp.CreateControlImplementation.ControlImplementation.ID)

			if tc.request.Details != nil {
				assert.Equal(t, *tc.request.Details, *resp.CreateControlImplementation.ControlImplementation.Details)
			} else {
				assert.Empty(t, resp.CreateControlImplementation.ControlImplementation.Details)
			}

			// check if the implementation is close, time is hard in CI so don't check exact
			if tc.request.ImplementationDate != nil {
				assert.WithinDuration(t, yesterday, *resp.CreateControlImplementation.ControlImplementation.ImplementationDate, time.Hour)
			} else {
				assert.Nil(t, resp.CreateControlImplementation.ControlImplementation.ImplementationDate)
			}

			if tc.request.Tags != nil {
				assert.Equal(t, tc.request.Tags, resp.CreateControlImplementation.ControlImplementation.Tags)
			} else {
				assert.Empty(t, resp.CreateControlImplementation.ControlImplementation.Tags)
			}

			if tc.request.Verified != nil {
				assert.Equal(t, *tc.request.Verified, *resp.CreateControlImplementation.ControlImplementation.Verified)
			} else {
				assert.False(t, *resp.CreateControlImplementation.ControlImplementation.Verified)
			}

			if tc.request.VerificationDate != nil {
				assert.WithinDuration(t, yesterday, *resp.CreateControlImplementation.ControlImplementation.VerificationDate, time.Hour)
			} else {
				assert.Nil(t, resp.CreateControlImplementation.ControlImplementation.VerificationDate)
			}

			if tc.request.Status != nil {
				assert.Equal(t, tc.request.Status, resp.CreateControlImplementation.ControlImplementation.Status)
			} else {
				// default value is DocumentDraft in the model
				assert.Equal(t, &enums.DocumentDraft, resp.CreateControlImplementation.ControlImplementation.Status)
			}

			if tc.request.ControlIDs != nil {
				assert.Len(t, resp.CreateControlImplementation.ControlImplementation.Controls.Edges, numControls)
				assert.Equal(t, int64(numControls), resp.CreateControlImplementation.ControlImplementation.Controls.TotalCount)
			} else {
				assert.Empty(t, resp.CreateControlImplementation.ControlImplementation.Controls.Edges)
			}

			if tc.request.SubcontrolIDs != nil {
				assert.Len(t, resp.CreateControlImplementation.ControlImplementation.Subcontrols.Edges, numControls)
				assert.Equal(t, int64(numControls), resp.CreateControlImplementation.ControlImplementation.Subcontrols.TotalCount)
			} else {
				assert.Empty(t, resp.CreateControlImplementation.ControlImplementation.Subcontrols.Edges)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateControlImplementation() {
	t := suite.T()

	controlImplementation1 := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlImplementation2 := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	yesterday := time.Now().Add(-time.Hour * 24)

	groupEditor := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add viewOnlyUser to the group with editor permissions
	(&GroupMemberBuilder{client: suite.client, GroupID: groupEditor.ID, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	numControls := 2
	controlIDs := []string{}
	for range numControls {
		// create controls where the groupEditor has editor permissions
		control := (&ControlBuilder{client: suite.client, ControlEditorGroupID: groupEditor.ID}).MustNew(testUser1.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	// create one control without group permissions
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	numSubcontrols := 3
	subcontrolIDs := []string{}
	for range numSubcontrols {
		subcontrol := (&SubcontrolBuilder{client: suite.client, ControlID: controlIDs[0]}).MustNew(testUser1.UserCtx, t)
		subcontrolIDs = append(subcontrolIDs, subcontrol.ID)
	}

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateControlImplementationInput
		id          string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: openlaneclient.UpdateControlImplementationInput{
				Details: lo.ToPtr(gofakeit.Paragraph(3, 5, 30, "<br />")),
			},
			id:     controlImplementation1.ID,
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: openlaneclient.UpdateControlImplementationInput{
				AddControlIDs:      controlIDs,
				AddSubcontrolIDs:   subcontrolIDs,
				Details:            lo.ToPtr(gofakeit.Paragraph(3, 5, 30, "<br />")),
				Status:             &enums.DocumentNeedsApproval,
				ImplementationDate: &yesterday,
				Verified:           lo.ToPtr(true),
				Tags:               []string{"control-imp", "ref-code"},
			},
			id:     controlImplementation1.ID,
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update verification date",
			request: openlaneclient.UpdateControlImplementationInput{
				VerificationDate: &yesterday,
			},
			id:     controlImplementation1.ID,
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "allowed because previous request added control IDs the user has access to ",
			request: openlaneclient.UpdateControlImplementationInput{
				Status: &enums.DocumentPublished,
			},
			id:     controlImplementation1.ID,
			client: suite.client.api,
			ctx:    viewOnlyUser.UserCtx,
		},
		{
			name: "happy path remove control IDs",
			request: openlaneclient.UpdateControlImplementationInput{
				RemoveControlIDs: controlIDs,
			},
			id:     controlImplementation1.ID,
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "no longer allowed because previous request removed control IDs the user has access to ",
			request: openlaneclient.UpdateControlImplementationInput{
				Status: &enums.DocumentPublished,
			},
			id:     controlImplementation1.ID,
			client: suite.client.api,
			ctx:    viewOnlyUser.UserCtx,
		},
		{
			name: "update not allowed, not enough permissions, not found",
			request: openlaneclient.UpdateControlImplementationInput{
				Status: &enums.DocumentPublished,
			},
			id:          controlImplementation2.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "happy path, add control IDs",
			request: openlaneclient.UpdateControlImplementationInput{
				Status: &enums.DocumentPublished,
				AddControlIDs: []string{
					control.ID,
				},
			},
			id:     controlImplementation2.ID,
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update still not allowed, not enough permissions, not permissions to control either",
			request: openlaneclient.UpdateControlImplementationInput{
				Status: &enums.DocumentPublished,
			},
			id:          controlImplementation2.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: openlaneclient.UpdateControlImplementationInput{
				Status: &enums.DocumentPublished,
			},
			id:          controlImplementation1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateControlImplementation(tc.ctx, tc.id, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.Details != nil {
				assert.Equal(t, *tc.request.Details, *resp.UpdateControlImplementation.ControlImplementation.Details)
			}

			if tc.request.Status != nil {
				assert.Equal(t, tc.request.Status, resp.UpdateControlImplementation.ControlImplementation.Status)
			}

			if tc.request.Verified != nil {
				assert.Equal(t, tc.request.Verified, resp.UpdateControlImplementation.ControlImplementation.Verified)
			}

			if tc.request.VerificationDate != nil {
				assert.WithinDuration(t, yesterday, *resp.UpdateControlImplementation.ControlImplementation.VerificationDate, time.Hour)
			} else if tc.request.Verified != nil && *tc.request.Verified {
				// default value is time.Now() in the model if verified is true
				assert.WithinDuration(t, time.Now(), *resp.UpdateControlImplementation.ControlImplementation.VerificationDate, time.Hour)
			}

			if tc.request.ImplementationDate != nil {
				assert.WithinDuration(t, yesterday, *resp.UpdateControlImplementation.ControlImplementation.ImplementationDate, time.Hour)
			}

			if tc.request.Tags != nil {
				assert.Equal(t, tc.request.Tags, resp.UpdateControlImplementation.ControlImplementation.Tags)
			}

			if tc.request.AddControlIDs != nil {
				assert.Len(t, resp.UpdateControlImplementation.ControlImplementation.Controls.Edges, len(tc.request.AddControlIDs))
				assert.Equal(t, int64(len(tc.request.AddControlIDs)), resp.UpdateControlImplementation.ControlImplementation.Controls.TotalCount)
			}

			if tc.request.AddSubcontrolIDs != nil {
				assert.Len(t, resp.UpdateControlImplementation.ControlImplementation.Subcontrols.Edges, numSubcontrols)
				assert.Equal(t, int64(numSubcontrols), resp.UpdateControlImplementation.ControlImplementation.Subcontrols.TotalCount)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeletecontrolImplementation() {
	t := suite.T()

	// create controlImplementations to be deleted
	controlImplementation1 := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlImplementation2 := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  controlImplementation1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: controlImplementation1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  controlImplementation1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: controlImplementation2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteControlImplementation(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteControlImplementation.DeletedID)
		})
	}
}
