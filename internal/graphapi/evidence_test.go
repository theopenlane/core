package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestQueryEvidence(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	(&ProgramMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID, ProgramID: program.ID}).MustNew(adminUser.UserCtx, t)

	// create an Evidence to be queried using adminUser
	// org owner (testUser1) should automatically have access to the Evidence
	evidence := (&EvidenceBuilder{client: suite.client, ProgramID: program.ID}).MustNew(adminUser.UserCtx, t)

	// create a control to be queried using adminUser that access is granted via the control
	control := (&ControlBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	evidenceControl := (&EvidenceBuilder{client: suite.client, ControlID: control.ID}).MustNew(testUser1.UserCtx, t)
	anonymousContext := createAnonymousTrustCenterContext(ulids.New().String(), testUser1.OrganizationID)

	// add test cases for querying the Evidence
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path, creator of the evidence",
			queryID: evidence.ID,
			client:  suite.client.api,
			ctx:     adminUser.UserCtx,
		},
		{
			name:    "happy path, permissions via control",
			queryID: evidenceControl.ID,
			client:  suite.client.api,
			ctx:     adminUser.UserCtx,
		},
		{
			name:    "happy path, org owner",
			queryID: evidence.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "read only user in organization, authorized via program",
			queryID: evidence.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:     "read only user in organization, no access",
			queryID:  evidence.ID,
			client:   suite.client.api,
			ctx:      viewOnlyUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:    "read only user in organization, has access via control",
			queryID: evidenceControl.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: evidence.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "Evidence not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "Evidence not found, using not authorized user",
			queryID:  evidence.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			queryID:  evidence.ID,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetEvidenceByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Evidence.ID))

			assert.Check(t, len(resp.Evidence.Name) != 0)
			assert.Check(t, len(resp.Evidence.DisplayID) != 0)
			assert.Check(t, !resp.Evidence.CreatedAt.IsZero())
			assert.Check(t, !resp.Evidence.UpdatedAt.IsZero())
		})
	}

	// delete created evidence
	(&Cleanup[*generated.EvidenceDeleteOne]{client: suite.client.db.Evidence, IDs: []string{evidence.ID, evidenceControl.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryEvidences(t *testing.T) {
	// create multiple objects by adminUser, org owner should have access to them as well
	e1 := (&EvidenceBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	e2 := (&EvidenceBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	userAnotherOrg := suite.userBuilder(context.Background(), t)

	// add evidence for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	e3 := (&EvidenceBuilder{client: suite.client}).MustNew(userAnotherOrg.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org, access not automatically granted",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 0,
		},
		{
			name:            "happy path, using api token, access not automatically granted",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 0,
		},
		{
			name:            "happy path, using pat, which is for the org owner so access is granted",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no Evidences should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEvidences(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Evidences.Edges, tc.expectedResults))
		})
	}

	// delete created evidences
	(&Cleanup[*generated.EvidenceDeleteOne]{client: suite.client.db.Evidence, IDs: []string{e1.ID, e2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.EvidenceDeleteOne]{client: suite.client.db.Evidence, ID: e3.ID}).MustDelete(userAnotherOrg.UserCtx, t)
}

func TestMutationCreateEvidence(t *testing.T) {
	program := (&ProgramBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	pngFile, err := storage.NewUploadFile("testdata/uploads/logo.png")
	assert.NilError(t, err)

	csvFile, err := storage.NewUploadFile("testdata/uploads/orgs.csv")
	assert.NilError(t, err)

	pdfFile, err := storage.NewUploadFile("testdata/uploads/hello.pdf")
	assert.NilError(t, err)

	txtFile, err := storage.NewUploadFile("testdata/uploads/hello.txt")
	assert.NilError(t, err)

	// create edges to be used in the test cases
	control1 := (&ControlBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	controlObjective1 := (&ControlObjectiveBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	controlObjective2 := (&ControlObjectiveBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	subcontrol1 := (&SubcontrolBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	subcontrol2 := (&SubcontrolBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	// create system owned control to test that it cannot be linked
	systemOwnedSubcontrol := (&SubcontrolBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	systemOwnedControl := (&ControlBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	// create a task for view only user
	taskViewOnly := (&TaskBuilder{client: suite.client}).MustNew(viewOnlyUser.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.CreateEvidenceInput
		files       []*graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateEvidenceInput{
				Name: "Test Evidence",
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, view only user should be able to associate evidence to a task they can edit",
			request: testclient.CreateEvidenceInput{
				Name:    "Test Evidence",
				TaskIDs: []string{taskViewOnly.ID},
			},
			client: suite.client.api,
			ctx:    viewOnlyUser.UserCtx,
		},
		{
			name: "happy path, view only user should be able to associate evidence to a task they can edit and control they can view",
			request: testclient.CreateEvidenceInput{
				Name:       "Test Evidence",
				TaskIDs:    []string{taskViewOnly.ID},
				ControlIDs: []string{control1.ID},
			},
			client: suite.client.api,
			ctx:    viewOnlyUser.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateEvidenceInput{
				Name:                "Test Evidence",
				Description:         lo.ToPtr("This is a test Evidence"),
				CollectionProcedure: lo.ToPtr("This is how we collected the Evidence"),
				Source:              lo.ToPtr("meows"),
				CreationDate:        lo.ToPtr(time.Now().Add(-time.Hour)),
				RenewalDate:         lo.ToPtr(time.Now().Add(365 * 24 * time.Hour)),
				IsAutomated:         lo.ToPtr(true),
				URL:                 lo.ToPtr("https://example.com/my-evidence.png"),
				ProgramIDs:          []string{program.ID},
				TaskIDs:             []string{task.ID},
				ControlIDs:          []string{control1.ID, control2.ID},
				ControlObjectiveIDs: []string{controlObjective1.ID, controlObjective2.ID},
				SubcontrolIDs:       []string{subcontrol1.ID, subcontrol2.ID},
			},
			files: []*graphql.Upload{
				{
					File:        pngFile.RawFile,
					Filename:    pngFile.OriginalName,
					Size:        pngFile.Size,
					ContentType: pngFile.ContentType,
				},
				{
					File:        csvFile.RawFile,
					Filename:    csvFile.OriginalName,
					Size:        csvFile.Size,
					ContentType: csvFile.ContentType,
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, admin user in org",
			request: testclient.CreateEvidenceInput{
				Name:                "Test Evidence",
				Description:         lo.ToPtr("This is a test Evidence"),
				CollectionProcedure: lo.ToPtr("This is how we collected the Evidence"),
				Source:              lo.ToPtr("meows"),
				CreationDate:        lo.ToPtr(time.Now().Add(-time.Hour)),
				RenewalDate:         lo.ToPtr(time.Now().Add(365 * 24 * time.Hour)),
				IsAutomated:         lo.ToPtr(true),
				URL:                 lo.ToPtr("https://example.com/my-evidence.png"),
				ControlIDs:          []string{control1.ID, control2.ID},                   // ensure the same controls can be added to multiple evidences
				ControlObjectiveIDs: []string{controlObjective1.ID, controlObjective2.ID}, // ensure the same control objectives can be added to multiple evidences
				SubcontrolIDs:       []string{subcontrol1.ID, subcontrol2.ID},             // ensure the same subcontrols can be added to multiple evidences
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "attempt to link system owned control",
			request: testclient.CreateEvidenceInput{
				Name:       "Test Evidence",
				TaskIDs:    []string{taskViewOnly.ID},
				ControlIDs: []string{systemOwnedControl.ID},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "attempt to link system owned subcontrol",
			request: testclient.CreateEvidenceInput{
				Name:          "Test Evidence",
				TaskIDs:       []string{taskViewOnly.ID},
				SubcontrolIDs: []string{systemOwnedSubcontrol.ID},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "attempt to link system owned control and org owned control",
			request: testclient.CreateEvidenceInput{
				Name:       "Test Evidence",
				TaskIDs:    []string{taskViewOnly.ID},
				ControlIDs: []string{systemOwnedControl.ID, control1.ID},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateEvidenceInput{
				Name:    "Test Evidence - TSK-123",
				TaskIDs: []string{task.ID},
				OwnerID: &testUser1.OrganizationID,
			},
			files: []*graphql.Upload{
				{
					File:        pdfFile.RawFile,
					Filename:    pdfFile.OriginalName,
					Size:        pdfFile.Size,
					ContentType: pdfFile.ContentType,
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateEvidenceInput{
				Name: "Test Evidence - TSK-123",
			},
			files: []*graphql.Upload{
				{
					File:        txtFile.RawFile,
					Filename:    txtFile.OriginalName,
					Size:        txtFile.Size,
					ContentType: txtFile.ContentType,
				},
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions and no linked objects",
			request: testclient.CreateEvidenceInput{
				Name: "Test Evidence",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user not authorized, not enough permissions and edit access to linked task",
			request: testclient.CreateEvidenceInput{
				Name:    "Test Evidence",
				TaskIDs: []string{task.ID},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "no access to linked control",
			request: testclient.CreateEvidenceInput{
				Name:        "Test Evidence",
				Description: lo.ToPtr("This is a test Evidence"),
				ControlIDs:  []string{control1.ID, control2.ID},
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required field",
			request: testclient.CreateEvidenceInput{
				Description: lo.ToPtr("This is a test Evidence"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "invalid url",
			request: testclient.CreateEvidenceInput{
				Name: "TSK-11123F Evidence",
				URL:  lo.ToPtr("invalid"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "invalid or unparsable field",
		},
		{
			name: "creation date in the future",
			request: testclient.CreateEvidenceInput{
				Name:         "Test Evidence",
				CreationDate: lo.ToPtr(time.Now().Add(time.Hour)),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "time cannot be in the future",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if len(tc.files) > 0 {
				expectUploadNillable(t, suite.client.mockProvider, tc.files)
			}

			resp, err := tc.client.CreateEvidence(tc.ctx, tc.request, tc.files)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, len(resp.CreateEvidence.Evidence.ID) != 0)
			assert.Check(t, len(resp.CreateEvidence.Evidence.DisplayID) != 0)
			assert.Check(t, len(resp.CreateEvidence.Evidence.Name) != 0)

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateEvidence.Evidence.Description))
			} else {
				assert.Check(t, is.Equal(*resp.CreateEvidence.Evidence.Description, ""))
			}

			if tc.request.CollectionProcedure != nil {
				assert.Check(t, is.Equal(*tc.request.CollectionProcedure, *resp.CreateEvidence.Evidence.CollectionProcedure))
			} else {
				assert.Check(t, is.Equal(*resp.CreateEvidence.Evidence.CollectionProcedure, ""))
			}

			if tc.request.Source != nil {
				assert.Check(t, is.Equal(*tc.request.Source, *resp.CreateEvidence.Evidence.Source))
			} else {
				assert.Check(t, is.Equal(*resp.CreateEvidence.Evidence.Source, ""))
			}

			if tc.request.CreationDate != nil {
				assert.Check(t, !resp.CreateEvidence.Evidence.CreationDate.IsZero())
				diff := resp.CreateEvidence.Evidence.CreationDate.Sub(*tc.request.CreationDate)
				assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			} else {
				assert.Check(t, !resp.CreateEvidence.Evidence.CreationDate.IsZero())
				diff := time.Until(resp.CreateEvidence.Evidence.CreationDate)
				assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			}

			if tc.request.RenewalDate != nil {
				assert.Check(t, !resp.CreateEvidence.Evidence.RenewalDate.IsZero())
				diff := resp.CreateEvidence.Evidence.RenewalDate.Sub(*tc.request.RenewalDate)
				assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			} else {
				assert.Check(t, !resp.CreateEvidence.Evidence.RenewalDate.IsZero())
				diff := resp.CreateEvidence.Evidence.RenewalDate.Sub(time.Now().Add(365 * 24 * time.Hour)) // check that it is 1 year from now
				assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			}

			if tc.request.IsAutomated != nil {
				assert.Check(t, is.Equal(*tc.request.IsAutomated, *resp.CreateEvidence.Evidence.IsAutomated))
			} else {
				assert.Check(t, !*resp.CreateEvidence.Evidence.IsAutomated)
			}

			if tc.request.Status == nil {
				assert.Check(t, is.Equal(*resp.CreateEvidence.Evidence.Status, enums.EvidenceStatusSubmitted))
			} else {
				assert.Check(t, is.Equal(*resp.CreateEvidence.Evidence.Status, *tc.request.Status))
			}

			if tc.request.URL != nil {
				assert.Check(t, is.Equal(*tc.request.URL, *resp.CreateEvidence.Evidence.URL))
			} else {
				assert.Check(t, is.Equal(*resp.CreateEvidence.Evidence.URL, ""))
			}

			if tc.request.ProgramIDs != nil {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.Programs.Edges, len(tc.request.ProgramIDs)))
			} else {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.Programs.Edges, 0))
			}

			if tc.request.TaskIDs != nil {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.Tasks.Edges, len(tc.request.TaskIDs)))
			} else {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.Tasks.Edges, 0))
			}

			if len(tc.files) > 0 {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.Files.Edges, len(tc.files)))
			} else {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.Files.Edges, 0))
			}

			// attempt to retrieve the created evidence by org owner, no matter who created it
			// the org owner should have access to it
			resp2, err := suite.client.api.GetEvidenceByID(testUser1.UserCtx, resp.CreateEvidence.Evidence.ID)
			assert.NilError(t, err)
			assert.Assert(t, resp2 != nil)

			// delete the created evidence, update for the token user cases
			if tc.ctx == context.Background() {
				tc.ctx = testUser1.UserCtx
			}

			// delete the evidence
			(&Cleanup[*generated.EvidenceDeleteOne]{client: suite.client.db.Evidence, ID: resp.CreateEvidence.Evidence.ID}).MustDelete(tc.ctx, t)
			// delete the files created for the evidence
			for _, file := range resp.CreateEvidence.Evidence.Files.Edges {
				(&Cleanup[*generated.FileDeleteOne]{client: suite.client.db.File, IDs: []string{file.Node.ID}}).MustDelete(tc.ctx, t)
			}
		})
	}
	// delete created objects
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, subcontrol1.ControlID, subcontrol2.ControlID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlObjectiveDeleteOne]{client: suite.client.db.ControlObjective, IDs: []string{controlObjective1.ID, controlObjective2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: []string{subcontrol1.ID, subcontrol2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: []string{task.ID, taskViewOnly.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, ID: program.ID}).MustDelete(testUser1.UserCtx, t)
	// delete system owned controls and subcontrols
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, ID: systemOwnedSubcontrol.ID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{systemOwnedControl.ID, systemOwnedSubcontrol.ControlID}}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationUpdateEvidence(t *testing.T) {
	evidence := (&EvidenceBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	pdfFile, err := storage.NewUploadFile("testdata/uploads/hello.pdf")
	assert.NilError(t, err)

	testCases := []struct {
		name        string
		request     testclient.UpdateEvidenceInput
		files       []*graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateEvidenceInput{
				CollectionProcedure: lo.ToPtr("This is how we collected the updated Evidence"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, update multiple fields using PAT",
			request: testclient.UpdateEvidenceInput{
				Name:                lo.ToPtr("Updated Evidence"),
				Description:         lo.ToPtr("This is an updated Evidence"),
				CollectionProcedure: lo.ToPtr("This is how we collected the updated Evidence"),
				Source:              lo.ToPtr("meows"),
			},
			files: []*graphql.Upload{
				{
					File:        pdfFile.RawFile,
					Filename:    pdfFile.OriginalName,
					Size:        pdfFile.Size,
					ContentType: pdfFile.ContentType,
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateEvidenceInput{
				Description: lo.ToPtr("This is an updated description of evidence"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateEvidenceInput{
				Source: lo.ToPtr("woofs"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update not allowed, creation date is in the future",
			request: testclient.UpdateEvidenceInput{
				CreationDate: lo.ToPtr(time.Now().Add(time.Minute)),
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: "time cannot be in the future",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if len(tc.files) > 0 {
				expectUploadNillable(t, suite.client.mockProvider, tc.files)
			}

			resp, err := tc.client.UpdateEvidence(tc.ctx, evidence.ID, tc.request, tc.files)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// add checks for the updated fields if they were set in the request
			if tc.request.Name != nil {
				assert.Check(t, is.Equal(*tc.request.Name, resp.UpdateEvidence.Evidence.Name))
			}

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateEvidence.Evidence.Description))
			}

			if tc.request.CollectionProcedure != nil {
				assert.Check(t, is.Equal(*tc.request.CollectionProcedure, *resp.UpdateEvidence.Evidence.CollectionProcedure))
			}

			if tc.request.Source != nil {
				assert.Check(t, is.Equal(*tc.request.Source, *resp.UpdateEvidence.Evidence.Source))
			}

			if len(tc.files) > 0 {
				assert.Check(t, is.Len(resp.UpdateEvidence.Evidence.Files.Edges, len(tc.files)))
			}
		})
	}

	// delete created evidence
	(&Cleanup[*generated.EvidenceDeleteOne]{client: suite.client.db.Evidence, ID: evidence.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteEvidence(t *testing.T) {
	// create objects to be deleted
	evidence1 := (&EvidenceBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	evidence2 := (&EvidenceBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  evidence1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: evidence1.ID,
			client:     suite.client.api,
			ctx:        adminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  evidence1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: evidence2.ID,
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
			resp, err := tc.client.DeleteEvidence(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteEvidence.DeletedID))
		})
	}
}
