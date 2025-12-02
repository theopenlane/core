package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryControlImplementation(t *testing.T) {
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

	// ensure view only user can access controlImplementation with associated controls
	control3 := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlImplementation4 := (&ControlImplementationBuilder{client: suite.client, ControlIDs: []string{control3.ID}}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the controlImplementation
	testCases := []struct {
		name                  string
		queryID               string
		client                *testclient.TestClient
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
			name:               "happy path, controlImplementation with associated controls and group viewer by org owner",
			queryID:            controlImplementation4.ID,
			client:             suite.client.api,
			ctx:                testUser1.UserCtx,
			shouldHaveControls: true,
		},
		{
			name:               "happy path, controlImplementation with associated controls and group viewer by view only user",
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
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.ControlImplementation.ID))

			// check if the controlImplementation has the expected fields
			assert.Check(t, len(resp.ControlImplementation.ID) != 0)

			// model tests set up defaults here
			assert.Check(t, resp.ControlImplementation.Details != nil)
			assert.Check(t, !resp.ControlImplementation.ImplementationDate.IsZero())

			if tc.shouldHaveControls {
				assert.Check(t, resp.ControlImplementation.Controls.Edges != nil)
			}

			if tc.shouldHaveSubcontrols {
				assert.Check(t, len(resp.ControlImplementation.Subcontrols.Edges) != 0)
				assert.Check(t, is.Len(resp.ControlImplementation.Subcontrols.Edges, 2))
				assert.Check(t, is.Equal(subcontrol1.ID, resp.ControlImplementation.Subcontrols.Edges[0].Node.ID))
				assert.Check(t, is.Equal(subcontrol2.ID, resp.ControlImplementation.Subcontrols.Edges[1].Node.ID))
			}
		})
	}

	(&Cleanup[*generated.ControlImplementationDeleteOne]{client: suite.client.db.ControlImplementation, IDs: []string{controlImplementation1.ID, controlImplementation3.ID, controlImplementation4.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlImplementationDeleteOne]{client: suite.client.db.ControlImplementation, IDs: []string{controlImplementation2.ID}}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: []string{subcontrol1.ID, subcontrol2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID}}).MustDelete(testUser1.UserCtx, t)

}

func TestQueryControlImplementations(t *testing.T) {
	// create a new user cause its a count test and we don't want to interfere with other tests
	testUser := suite.userBuilder(context.Background(), t)
	apiClient := suite.setupAPITokenClient(testUser.UserCtx, t)
	patClient := suite.setupPatClient(testUser, t)
	viewUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &viewUser, enums.RoleMember, testUser.OrganizationID)

	anotherUser := suite.userBuilder(context.Background(), t)

	// create multiple controlImplementations to be queried using testUser
	numCIs := 5
	ciIDs := []string{}
	for range numCIs {
		ci := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
		ciIDs = append(ciIDs, ci.ID)
	}

	// view only users should be able to see these because they are associated with a control
	numCIsWithAssociatedControls := 2
	controlIDs := []string{}
	for range numCIsWithAssociatedControls {
		control1 := (&ControlBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
		ci := (&ControlImplementationBuilder{client: suite.client, ControlIDs: []string{control1.ID}}).MustNew(testUser.UserCtx, t)
		ciIDs = append(ciIDs, ci.ID)

		controlIDs = append(controlIDs, control1.ID)
	}

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: numCIs + numCIsWithAssociatedControls,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewUser.UserCtx,
			expectedResults: numCIsWithAssociatedControls,
		},
		{
			name:            "happy path, using api token",
			client:          apiClient,
			ctx:             context.Background(),
			expectedResults: numCIsWithAssociatedControls, // only the ones with linked controls will be returned
		},
		{
			name:            "happy path, using pat",
			client:          patClient,
			ctx:             context.Background(),
			expectedResults: numCIs + numCIsWithAssociatedControls,
		},
		{
			name:            "another user, no controlImplementations should be returned",
			client:          suite.client.api,
			ctx:             anotherUser.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			// ensure we don't conflict with other tests
			where := &testclient.ControlImplementationWhereInput{
				IDIn: ciIDs,
			}

			resp, err := tc.client.GetControlImplementations(tc.ctx, where)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.ControlImplementations.Edges, tc.expectedResults))
		})
	}

	// cleanup
	(&Cleanup[*generated.ControlImplementationDeleteOne]{client: suite.client.db.ControlImplementation, IDs: ciIDs}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDs}).MustDelete(testUser.UserCtx, t)
}

func TestMutationCreateControlImplementation(t *testing.T) {
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
		request     testclient.CreateControlImplementationInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path, minimal input",
			request: testclient.CreateControlImplementationInput{
				// there are no required fields in the model, you can create a blank implementation
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateControlImplementationInput{
				Details:            lo.ToPtr(gofakeit.Paragraph()),
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
			request: testclient.CreateControlImplementationInput{
				Details: lo.ToPtr(gofakeit.Paragraph()),
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateControlImplementationInput{
				Details: lo.ToPtr(gofakeit.Paragraph()),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateControlImplementationInput{
				Details: lo.ToPtr(gofakeit.Paragraph()),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user authorized because they have editor permissions to all the parent control",
			request: testclient.CreateControlImplementationInput{
				Details:    lo.ToPtr(gofakeit.Paragraph()),
				ControlIDs: controlIDs,
			},
			client: suite.client.api,
			ctx:    viewOnlyUser.UserCtx,
		},
		{
			name: "user not authorized, not enough permissions to one of the parent controls",
			request: testclient.CreateControlImplementationInput{
				Details:    lo.ToPtr(gofakeit.Paragraph()),
				ControlIDs: allControlIDs,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "no access to linked control",
			request: testclient.CreateControlImplementationInput{
				Details:    lo.ToPtr(gofakeit.Paragraph()),
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
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			// check default fields
			assert.Check(t, len(resp.CreateControlImplementation.ControlImplementation.ID) != 0)

			if tc.request.Details != nil {
				assert.Check(t, is.Equal(*tc.request.Details, *resp.CreateControlImplementation.ControlImplementation.Details))
			} else {
				assert.Check(t, is.Equal(*resp.CreateControlImplementation.ControlImplementation.Details, ""))
			}

			// check if the implementation is close, time is hard in CI so don't check exact
			if tc.request.ImplementationDate != nil {
				assert.Check(t, resp.CreateControlImplementation.ControlImplementation.ImplementationDate != nil)
				diff := resp.CreateControlImplementation.ControlImplementation.ImplementationDate.Sub(yesterday)
				assert.Check(t, diff >= -time.Hour && diff <= time.Hour, "time difference is not within 1 hour")
			} else {
				assert.Check(t, is.Nil(resp.CreateControlImplementation.ControlImplementation.ImplementationDate))
			}

			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.CreateControlImplementation.ControlImplementation.Tags))
			} else {
				assert.Check(t, is.Len(resp.CreateControlImplementation.ControlImplementation.Tags, 0))
			}

			if tc.request.Verified != nil {
				assert.Check(t, is.Equal(*tc.request.Verified, *resp.CreateControlImplementation.ControlImplementation.Verified))
			} else {
				assert.Check(t, !*resp.CreateControlImplementation.ControlImplementation.Verified)
			}

			if tc.request.VerificationDate != nil {
				assert.Check(t, resp.CreateControlImplementation.ControlImplementation.VerificationDate != nil)
				diff := resp.CreateControlImplementation.ControlImplementation.VerificationDate.Sub(yesterday)
				assert.Check(t, diff >= -time.Hour && diff <= time.Hour, "time difference is not within 1 hour")
			} else {
				assert.Check(t, is.Nil(resp.CreateControlImplementation.ControlImplementation.VerificationDate))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.DeepEqual(tc.request.Status, resp.CreateControlImplementation.ControlImplementation.Status))
			} else {
				// default value is DocumentDraft in the model
				assert.Check(t, is.DeepEqual(&enums.DocumentDraft, resp.CreateControlImplementation.ControlImplementation.Status))
			}

			if tc.request.ControlIDs != nil {
				assert.Check(t, is.Len(resp.CreateControlImplementation.ControlImplementation.Controls.Edges, numControls))
				assert.Check(t, is.Equal(int64(numControls), resp.CreateControlImplementation.ControlImplementation.Controls.TotalCount))
			} else {
				assert.Check(t, is.Len(resp.CreateControlImplementation.ControlImplementation.Controls.Edges, 0))
			}

			if tc.request.SubcontrolIDs != nil {
				assert.Check(t, is.Len(resp.CreateControlImplementation.ControlImplementation.Subcontrols.Edges, numControls))
				assert.Check(t, is.Equal(int64(numControls), resp.CreateControlImplementation.ControlImplementation.Subcontrols.TotalCount))
			} else {
				assert.Check(t, is.Len(resp.CreateControlImplementation.ControlImplementation.Subcontrols.Edges, 0))
			}

			(&Cleanup[*generated.ControlImplementationDeleteOne]{client: suite.client.db.ControlImplementation, ID: resp.CreateControlImplementation.ControlImplementation.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	// cleanup
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: allControlIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: subcontrolIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{groupEditor.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateControlImplementation(t *testing.T) {
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
		request     testclient.UpdateControlImplementationInput
		id          string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateControlImplementationInput{
				Details: lo.ToPtr(gofakeit.Paragraph()),
			},
			id:     controlImplementation1.ID,
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateControlImplementationInput{
				AddControlIDs:      controlIDs,
				AddSubcontrolIDs:   subcontrolIDs,
				Details:            lo.ToPtr(gofakeit.Paragraph()),
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
			request: testclient.UpdateControlImplementationInput{
				VerificationDate: &yesterday,
			},
			id:     controlImplementation1.ID,
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "allowed because previous request added control IDs the user has access to ",
			request: testclient.UpdateControlImplementationInput{
				Status: &enums.DocumentPublished,
			},
			id:     controlImplementation1.ID,
			client: suite.client.api,
			ctx:    viewOnlyUser.UserCtx,
		},
		{
			name: "happy path remove control IDs",
			request: testclient.UpdateControlImplementationInput{
				RemoveControlIDs: controlIDs,
			},
			id:     controlImplementation1.ID,
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "no longer allowed because previous request removed control IDs the user has access to ",
			request: testclient.UpdateControlImplementationInput{
				Status: &enums.DocumentPublished,
			},
			id:     controlImplementation1.ID,
			client: suite.client.api,
			ctx:    viewOnlyUser.UserCtx,
		},
		{
			name: "update not allowed, not enough permissions, not found",
			request: testclient.UpdateControlImplementationInput{
				Status: &enums.DocumentPublished,
			},
			id:          controlImplementation2.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "happy path, add control IDs",
			request: testclient.UpdateControlImplementationInput{
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
			name: "update still not allowed, not enough permissions, no edit permissions to control either",
			request: testclient.UpdateControlImplementationInput{
				Status: &enums.DocumentPublished,
			},
			id:          controlImplementation2.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateControlImplementationInput{
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
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			if tc.request.Details != nil {
				assert.Check(t, is.Equal(*tc.request.Details, *resp.UpdateControlImplementation.ControlImplementation.Details))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.DeepEqual(tc.request.Status, resp.UpdateControlImplementation.ControlImplementation.Status))
			}

			if tc.request.Verified != nil {
				assert.Check(t, is.DeepEqual(tc.request.Verified, resp.UpdateControlImplementation.ControlImplementation.Verified))
			}

			if tc.request.VerificationDate != nil {
				assert.Check(t, resp.UpdateControlImplementation.ControlImplementation.VerificationDate != nil)
				diff := resp.UpdateControlImplementation.ControlImplementation.VerificationDate.Sub(yesterday)
				assert.Check(t, diff >= -time.Hour && diff <= time.Hour, "time difference is not within 1 hour")
			} else if tc.request.Verified != nil && *tc.request.Verified {
				// default value is time.Now() in the model if verified is true
				assert.Check(t, resp.UpdateControlImplementation.ControlImplementation.VerificationDate != nil)
				diff := time.Until(*resp.UpdateControlImplementation.ControlImplementation.VerificationDate)
				assert.Check(t, diff >= -time.Hour && diff <= time.Hour, "time difference is not within 1 hour")
			}

			if tc.request.ImplementationDate != nil {
				assert.Check(t, resp.UpdateControlImplementation.ControlImplementation.ImplementationDate != nil)
				diff := resp.UpdateControlImplementation.ControlImplementation.ImplementationDate.Sub(yesterday)
				assert.Check(t, diff >= -time.Hour && diff <= time.Hour, "time difference is not within 1 hour")
			}

			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.UpdateControlImplementation.ControlImplementation.Tags))
			}

			if tc.request.AddControlIDs != nil {
				assert.Check(t, is.Len(resp.UpdateControlImplementation.ControlImplementation.Controls.Edges, len(tc.request.AddControlIDs)))
				assert.Check(t, is.Equal(int64(len(tc.request.AddControlIDs)), resp.UpdateControlImplementation.ControlImplementation.Controls.TotalCount))
			}

			if tc.request.AddSubcontrolIDs != nil {
				assert.Check(t, is.Len(resp.UpdateControlImplementation.ControlImplementation.Subcontrols.Edges, numSubcontrols))
				assert.Check(t, is.Equal(int64(numSubcontrols), resp.UpdateControlImplementation.ControlImplementation.Subcontrols.TotalCount))
			}
		})
	}

	(&Cleanup[*generated.ControlImplementationDeleteOne]{client: suite.client.db.ControlImplementation, IDs: []string{controlImplementation1.ID, controlImplementation2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: subcontrolIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{groupEditor.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteControlImplementation(t *testing.T) {
	// create controlImplementations to be deleted
	controlImplementation1 := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	controlImplementation2 := (&ControlImplementationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
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
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteControlImplementation.DeletedID))
		})
	}
}
