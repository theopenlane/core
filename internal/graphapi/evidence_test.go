package graphapi_test

import (
	"context"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryEvidence(t *testing.T) {
	evidenceNoParent := (&th.EvidenceBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)

	(&th.ProgramMemberBuilder{Client: suite.Client, UserID: th.SharedViewOnlyUser.ID, ProgramID: program.ID}).MustNew(th.SharedAdminUser.UserCtx, t)

	// create an Evidence to be queried using adminUser
	// org owner (th.SharedTestUser1) should automatically have access to the Evidence
	evidence := (&th.EvidenceBuilder{Client: suite.Client, ProgramID: program.ID}).MustNew(th.SharedAdminUser.UserCtx, t)

	// create a control to be queried using adminUser that access is granted via the control
	control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	evidenceControl := (&th.EvidenceBuilder{Client: suite.Client, ControlID: control.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	internalPolicy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	evidenceInternalPolicy := (&th.EvidenceBuilder{Client: suite.Client, InternalPolicyID: internalPolicy.ID}).MustNew(th.SharedAdminUser.UserCtx, t)

	procedure := (&th.ProcedureBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	evidenceProcedure := (&th.EvidenceBuilder{Client: suite.Client, ProcedureID: procedure.ID}).MustNew(th.SharedAdminUser.UserCtx, t)

	anonymousContext := th.CreateAnonymousTrustCenterContext(ulids.New().String(), th.SharedTestUser1.OrganizationID)

	// add test cases for querying the Evidence
	testCases := []struct {
		name        string
		queryID     string
		client      *testclient.TestClient
		ctx         context.Context
		errorMsg    string
		policyID    string
		procedureID string
	}{
		{
			name:    "happy path, creator of the evidence no parent",
			queryID: evidenceNoParent.ID,
			client:  suite.Client.API,
			ctx:     th.SharedAdminUser.UserCtx,
		},
		{
			name:    "happy path, creator of the evidence with program parent",
			queryID: evidence.ID,
			client:  suite.Client.API,
			ctx:     th.SharedAdminUser.UserCtx,
		},
		{
			name:    "happy path, permissions via control",
			queryID: evidenceControl.ID,
			client:  suite.Client.API,
			ctx:     th.SharedAdminUser.UserCtx,
		},
		{
			name:    "happy path, org owner",
			queryID: evidence.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "read only user in organization, authorized via program",
			queryID: evidence.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:     "read only user in organization, no access given via parent",
			queryID:  evidence.ID,
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "read only user in organization, no access",
			queryID:  evidenceNoParent.ID,
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "read only user 2 in organization, no access",
			queryID:  evidenceNoParent.ID,
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:    "read only user in organization, has access via control",
			queryID: evidenceControl.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:     "read only user in organization, has access via internal policy",
			queryID:  evidenceInternalPolicy.ID,
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			policyID: internalPolicy.ID,
		},
		{
			name:        "read only user in organization, has access via procedure",
			queryID:     evidenceProcedure.ID,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			procedureID: procedure.ID,
		},
		{
			name:    "happy path using personal access token",
			queryID: evidence.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "evidence not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "evidence not found, using not authorized user",
			queryID:  evidence.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.Client.API,
			ctx:      anonymousContext,
			queryID:  evidence.ID,
			errorMsg: th.NotFoundErrorMsg,
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

			if tc.policyID != "" {
				assert.Check(t, is.Len(resp.Evidence.InternalPolicies.Edges, 1))
				assert.Check(t, is.Equal(tc.policyID, resp.Evidence.InternalPolicies.Edges[0].Node.ID))
			}

			if tc.procedureID != "" {
				assert.Check(t, is.Len(resp.Evidence.Procedures.Edges, 1))
				assert.Check(t, is.Equal(tc.procedureID, resp.Evidence.Procedures.Edges[0].Node.ID))
			}
		})
	}

	// delete created evidence
	(&th.Cleanup[*generated.EvidenceDeleteOne]{Client: suite.Client.DB.Evidence, IDs: []string{evidence.ID, evidenceControl.ID, evidenceInternalPolicy.ID, evidenceProcedure.ID, evidenceNoParent.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: control.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: internalPolicy.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProcedureDeleteOne]{Client: suite.Client.DB.Procedure, ID: procedure.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryEvidences(t *testing.T) {
	// create multiple objects by adminUser, org owner should have access to them as well
	e1 := (&th.EvidenceBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	e2 := (&th.EvidenceBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)

	userAnotherOrg := suite.UserBuilder(context.Background(), t)

	// add evidence for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	e3 := (&th.EvidenceBuilder{Client: suite.Client}).MustNew(userAnotherOrg.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.Client.API,
			ctx:             th.SharedAdminUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org, access not automatically granted",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 0,
		},
		{
			name:            "happy path, using api token, includes evidence scope",
			client:          suite.Client.APIWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat, which is for the org owner so access is granted",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no Evidences should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
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
	(&th.Cleanup[*generated.EvidenceDeleteOne]{Client: suite.Client.DB.Evidence, IDs: []string{e1.ID, e2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.EvidenceDeleteOne]{Client: suite.Client.DB.Evidence, ID: e3.ID}).MustDelete(userAnotherOrg.UserCtx, t)
}

func TestMutationCreateEvidence(t *testing.T) {
	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	task := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	pngFile := th.UploadFile(t, th.LogoFilePath)
	csvFile := th.UploadFile(t, "testdata/uploads/orgs.csv")
	pdfFile := th.UploadFile(t, th.PdfFilePath)
	txtFile := th.UploadFile(t, th.TxtFilePath)

	// create edges to be used in the test cases
	control1 := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	control2 := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	controlObjective1 := (&th.ControlObjectiveBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	controlObjective2 := (&th.ControlObjectiveBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	subcontrol1 := (&th.SubcontrolBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	subcontrol2 := (&th.SubcontrolBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	internalPolicy1 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	internalPolicy2 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	procedure1 := (&th.ProcedureBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	procedure2 := (&th.ProcedureBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)

	// create system owned control to test that it cannot be linked
	systemOwnedSubcontrol := (&th.SubcontrolBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
	systemOwnedControl := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
	systemOwnedInternalPolicy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
	systemOwnedProcedure := (&th.ProcedureBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	// create a task for view only user
	taskViewOnly := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedViewOnlyUser.UserCtx, t)

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
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "happy path, view only user should be able to associate evidence to a task they can edit",
			request: testclient.CreateEvidenceInput{
				Name:    "Test Evidence",
				TaskIDs: []string{taskViewOnly.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedViewOnlyUser.UserCtx,
		},
		{
			name: "happy path, view only user should be able to associate evidence to a task they can edit and control they can view",
			request: testclient.CreateEvidenceInput{
				Name:       "Test Evidence",
				TaskIDs:    []string{taskViewOnly.ID},
				ControlIDs: []string{control1.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedViewOnlyUser.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateEvidenceInput{
				Name:                "Test Evidence",
				Description:         lo.ToPtr("This is a test Evidence"),
				CollectionProcedure: lo.ToPtr("This is how we collected the Evidence"),
				Source:              lo.ToPtr("meows"),
				CreationDate:        lo.ToPtr(models.DateTime(time.Now().Add(-time.Hour))),
				RenewalDate:         lo.ToPtr(models.DateTime(time.Now().Add(365 * 24 * time.Hour))),
				IsAutomated:         lo.ToPtr(true),
				URL:                 lo.ToPtr("https://example.com/my-evidence.png"),
				ProgramIDs:          []string{program.ID},
				TaskIDs:             []string{task.ID},
				ControlIDs:          []string{control1.ID, control2.ID},
				ControlObjectiveIDs: []string{controlObjective1.ID, controlObjective2.ID},
				SubcontrolIDs:       []string{subcontrol1.ID, subcontrol2.ID},
				InternalPolicyIDs:   []string{internalPolicy1.ID, internalPolicy2.ID},
				ProcedureIDs:        []string{procedure1.ID, procedure2.ID},
			},
			files: []*graphql.Upload{
				pngFile,
				csvFile,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, admin user in org",
			request: testclient.CreateEvidenceInput{
				Name:                "Test Evidence",
				Description:         lo.ToPtr("This is a test Evidence"),
				CollectionProcedure: lo.ToPtr("This is how we collected the Evidence"),
				Source:              lo.ToPtr("meows"),
				CreationDate:        lo.ToPtr(models.DateTime(time.Now().Add(-time.Hour))),
				RenewalDate:         lo.ToPtr(models.DateTime(time.Now().Add(365 * 24 * time.Hour))),
				IsAutomated:         lo.ToPtr(true),
				URL:                 lo.ToPtr("https://example.com/my-evidence.png"),
				ControlIDs:          []string{control1.ID, control2.ID},                   // ensure the same controls can be added to multiple evidences
				ControlObjectiveIDs: []string{controlObjective1.ID, controlObjective2.ID}, // ensure the same control objectives can be added to multiple evidences
				SubcontrolIDs:       []string{subcontrol1.ID, subcontrol2.ID},             // ensure the same subcontrols can be added to multiple evidences
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "happy path, auditor can associate evidence to a program they can view",
			request: testclient.CreateEvidenceInput{
				Name:       "Test Evidence",
				ProgramIDs: []string{program.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedAuditorUser.UserCtx,
		},
		{
			name: "attempt to link system owned control",
			request: testclient.CreateEvidenceInput{
				Name:       "Test Evidence",
				TaskIDs:    []string{taskViewOnly.ID},
				ControlIDs: []string{systemOwnedControl.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "attempt to link system owned subcontrol",
			request: testclient.CreateEvidenceInput{
				Name:          "Test Evidence",
				TaskIDs:       []string{taskViewOnly.ID},
				SubcontrolIDs: []string{systemOwnedSubcontrol.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "attempt to link system owned internal policy",
			request: testclient.CreateEvidenceInput{
				Name:              "System owned Evidence",
				TaskIDs:           []string{taskViewOnly.ID},
				InternalPolicyIDs: []string{systemOwnedInternalPolicy.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "attempt to link system owned procedure",
			request: testclient.CreateEvidenceInput{
				Name:         "System owned Evidence",
				TaskIDs:      []string{taskViewOnly.ID},
				ProcedureIDs: []string{systemOwnedProcedure.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "attempt to link system owned control and org owned control",
			request: testclient.CreateEvidenceInput{
				Name:       "System owned Evidence",
				TaskIDs:    []string{taskViewOnly.ID},
				ControlIDs: []string{systemOwnedControl.ID, control1.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateEvidenceInput{
				Name:    "Test Evidence - TSK-123",
				TaskIDs: []string{task.ID},
				OwnerID: &th.SharedTestUser1.OrganizationID,
			},
			files: []*graphql.Upload{
				pdfFile,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateEvidenceInput{
				Name: "Test Evidence - TSK-123",
			},
			files: []*graphql.Upload{
				txtFile,
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions and no linked objects",
			request: testclient.CreateEvidenceInput{
				Name: "Test Evidence",
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "user not authorized, not enough permissions and edit access to linked task",
			request: testclient.CreateEvidenceInput{
				Name:    "Test Evidence",
				TaskIDs: []string{task.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "no access to linked control",
			request: testclient.CreateEvidenceInput{
				Name:        "Test Evidence",
				Description: lo.ToPtr("This is a test Evidence"),
				ControlIDs:  []string{control1.ID, control2.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "missing required field",
			request: testclient.CreateEvidenceInput{
				Description: lo.ToPtr("This is a test Evidence"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "invalid url",
			request: testclient.CreateEvidenceInput{
				Name: "TSK-11123F Evidence",
				URL:  lo.ToPtr("invalid"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "invalid or unparsable field",
		},
		{
			name: "creation date in the future",
			request: testclient.CreateEvidenceInput{
				Name:         "Test Evidence",
				CreationDate: lo.ToPtr(models.DateTime(time.Now().Add(time.Hour))),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "time cannot be in the future",
		},
		{
			name: "renewal date in the past",
			request: testclient.CreateEvidenceInput{
				Name:        "Test Evidence",
				RenewalDate: lo.ToPtr(models.DateTime(time.Now().Add(-time.Hour))),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "time cannot be in the past",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if len(tc.files) > 0 {
				th.ExpectUploadNillable(t, suite.Client.MockProvider, tc.files)
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
				diff := time.Time(resp.CreateEvidence.Evidence.CreationDate).Sub(time.Time(*tc.request.CreationDate))
				assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			} else {
				assert.Check(t, !resp.CreateEvidence.Evidence.CreationDate.IsZero())
				diff := time.Until(time.Time(resp.CreateEvidence.Evidence.CreationDate))
				assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			}

			if tc.request.RenewalDate != nil {
				assert.Check(t, !resp.CreateEvidence.Evidence.RenewalDate.IsZero())
				diff := time.Time(*resp.CreateEvidence.Evidence.RenewalDate).Sub(time.Time(*tc.request.RenewalDate))
				assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			} else {
				assert.Check(t, !resp.CreateEvidence.Evidence.RenewalDate.IsZero())
				diff := time.Time(*resp.CreateEvidence.Evidence.RenewalDate).Sub(time.Now().Add(365 * 24 * time.Hour)) // check that it is 1 year from now
				assert.Check(t, diff >= -2*time.Minute && diff <= 2*time.Minute, "time difference is not within 2 minutes")
			}

			if tc.request.IsAutomated != nil {
				assert.Check(t, is.Equal(*tc.request.IsAutomated, *resp.CreateEvidence.Evidence.IsAutomated))
			} else {
				assert.Check(t, !*resp.CreateEvidence.Evidence.IsAutomated)
			}

			if tc.request.Status == nil {
				// we should always take the sent status; we just want to set missing artifact
				// if its created or updated and has not file or url and status isn't sent explicitly
				hasURL := tc.request.URL != nil && *tc.request.URL != ""
				hasFiles := len(tc.files) > 0
				if !hasURL && !hasFiles {
					assert.Check(t, is.Equal(*resp.CreateEvidence.Evidence.Status, enums.EvidenceStatusMissingArtifact))
				} else {
					assert.Check(t, is.Equal(*resp.CreateEvidence.Evidence.Status, enums.EvidenceStatusSubmitted))
				}
			} else {
				// explicit status should always be respected
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

			if tc.request.InternalPolicyIDs != nil {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.InternalPolicies.Edges, len(tc.request.InternalPolicyIDs)))
			} else {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.InternalPolicies.Edges, 0))
			}

			if tc.request.ProcedureIDs != nil {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.Procedures.Edges, len(tc.request.ProcedureIDs)))
			} else {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.Procedures.Edges, 0))
			}

			if len(tc.files) > 0 {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.Files.Edges, len(tc.files)))
			} else {
				assert.Check(t, is.Len(resp.CreateEvidence.Evidence.Files.Edges, 0))
			}

			// attempt to retrieve the created evidence by org owner, no matter who created it
			// the org owner should have access to it
			resp2, err := suite.Client.API.GetEvidenceByID(th.SharedTestUser1.UserCtx, resp.CreateEvidence.Evidence.ID)
			assert.NilError(t, err)
			assert.Assert(t, resp2 != nil)

			// delete the created evidence, update for the token user cases
			if tc.ctx == context.Background() {
				tc.ctx = th.SharedTestUser1.UserCtx
			}

			// delete the evidence
			(&th.Cleanup[*generated.EvidenceDeleteOne]{Client: suite.Client.DB.Evidence, ID: resp.CreateEvidence.Evidence.ID}).MustDelete(tc.ctx, t)
			// delete the files created for the evidence
			for _, file := range resp.CreateEvidence.Evidence.Files.Edges {
				(&th.Cleanup[*generated.FileDeleteOne]{Client: suite.Client.DB.File, IDs: []string{file.Node.ID}}).MustDelete(tc.ctx, t)
			}
		})
	}
	// delete created objects
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: []string{control1.ID, control2.ID, subcontrol1.ControlID, subcontrol2.ControlID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ControlObjectiveDeleteOne]{Client: suite.Client.DB.ControlObjective, IDs: []string{controlObjective1.ID, controlObjective2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, IDs: []string{subcontrol1.ID, subcontrol2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, IDs: []string{task.ID, taskViewOnly.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: []string{internalPolicy1.ID, internalPolicy2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProcedureDeleteOne]{Client: suite.Client.DB.Procedure, IDs: []string{procedure1.ID, procedure2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete system owned controls and subcontrols
	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, ID: systemOwnedSubcontrol.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: []string{systemOwnedControl.ID, systemOwnedSubcontrol.ControlID}}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: systemOwnedInternalPolicy.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
	(&th.Cleanup[*generated.ProcedureDeleteOne]{Client: suite.Client.DB.Procedure, ID: systemOwnedProcedure.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}

func TestMutationCreateBulkCSVEvidence(t *testing.T) {
	bulkFile := th.UploadFile(t, "testdata/uploads/evidence.csv")
	plainTagFile := th.UploadFile(t, "testdata/uploads/evidence_invalid.csv")

	evidences := []string{}
	testCases := []struct {
		name         string
		client       *testclient.TestClient
		fileInput    graphql.Upload
		ctx          context.Context
		expectedErr  string
		expectedLen  int
		expectedTags int
	}{
		{
			name:         "happy path, valid file with json array tags",
			client:       suite.Client.API,
			ctx:          th.SharedTestUser1.UserCtx,
			fileInput:    *bulkFile,
			expectedLen:  2,
			expectedTags: 3,
		},
		{
			name:         "happy path, plain string tag converted to array",
			client:       suite.Client.API,
			ctx:          th.SharedTestUser1.UserCtx,
			fileInput:    *plainTagFile,
			expectedLen:  1,
			expectedTags: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Bulk Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateBulkCSVEvidence(tc.ctx, tc.fileInput)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.CreateBulkCSVEvidence.Evidences, tc.expectedLen))

			for _, evidence := range resp.CreateBulkCSVEvidence.Evidences {
				assert.Check(t, is.Len(evidence.Tags, tc.expectedTags))
				assert.Check(t, evidence.Name != "")
				assert.Check(t, evidence.Description != nil)
				assert.Check(t, evidence.CollectionProcedure != nil)
			}

			for _, evidence := range resp.CreateBulkCSVEvidence.Evidences {
				evidences = append(evidences, evidence.ID)
			}
		})
	}

	(&th.Cleanup[*generated.EvidenceDeleteOne]{Client: suite.Client.DB.Evidence, IDs: evidences}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationUpdateEvidence(t *testing.T) {
	program := (&th.ProgramBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	evidence := (&th.EvidenceBuilder{Client: suite.Client, ProgramID: program.ID}).MustNew(th.SharedAdminUser.UserCtx, t)
	internalPolicy1 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	internalPolicy2 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	procedure1 := (&th.ProcedureBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	procedure2 := (&th.ProcedureBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)

	// add view only user to the program so that they have access to the evidence for testing update permissions
	pm := (&th.ProgramMemberBuilder{Client: suite.Client, UserID: th.SharedViewOnlyUser.ID, ProgramID: program.ID}).MustNew(th.SharedAdminUser.UserCtx, t)

	pdfFile := th.UploadFile(t, th.PdfFilePath)

	testCases := []struct {
		name        string
		request     testclient.UpdateEvidenceInput
		files       []*graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
		policyID    string
		procedureID string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateEvidenceInput{
				CollectionProcedure: lo.ToPtr("This is how we collected the updated Evidence"),
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
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
				pdfFile,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "member allowed to add comment",
			request: testclient.UpdateEvidenceInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "This is a comment",
				},
			},
			client: suite.Client.API,
			ctx:    th.SharedViewOnlyUser.UserCtx,
		},
		{
			name: "auditor allowed to updated",
			request: testclient.UpdateEvidenceInput{
				Status: &enums.EvidenceStatusAuditorApproved,
			},
			client: suite.Client.API,
			ctx:    th.SharedAuditorUser.UserCtx,
		},
		{
			name: "add internal policy and procedure",
			request: testclient.UpdateEvidenceInput{
				AddInternalPolicyIDs: []string{internalPolicy1.ID},
				AddProcedureIDs:      []string{procedure1.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedAdminUser.UserCtx,
			policyID:    internalPolicy1.ID,
			procedureID: procedure1.ID,
		},
		{
			name: "replace internal policy and procedure",
			request: testclient.UpdateEvidenceInput{
				AddInternalPolicyIDs:    []string{internalPolicy2.ID},
				RemoveInternalPolicyIDs: []string{internalPolicy1.ID},
				AddProcedureIDs:         []string{procedure2.ID},
				RemoveProcedureIDs:      []string{procedure1.ID},
			},
			client:      suite.Client.API,
			ctx:         th.SharedAdminUser.UserCtx,
			policyID:    internalPolicy2.ID,
			procedureID: procedure2.ID,
		},
		{
			name: "clear internal policies and procedures",
			request: testclient.UpdateEvidenceInput{
				ClearInternalPolicies: lo.ToPtr(true),
				ClearProcedures:       lo.ToPtr(true),
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "update not allowed, no permissions to update but can view due to program membership",
			request: testclient.UpdateEvidenceInput{
				Description: lo.ToPtr("This is an updated description of evidence"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateEvidenceInput{
				Source: lo.ToPtr("woofs"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "update not allowed, creation date is in the future",
			request: testclient.UpdateEvidenceInput{
				CreationDate: lo.ToPtr(models.DateTime(time.Now().Add(time.Minute))),
			},
			client:      suite.Client.API,
			ctx:         th.SharedAdminUser.UserCtx,
			expectedErr: "time cannot be in the future",
		},
		{
			name: "update not allowed, renewal date is in the past",
			request: testclient.UpdateEvidenceInput{
				RenewalDate: lo.ToPtr(models.DateTime(time.Now().Add(-time.Hour))),
			},
			client:      suite.Client.API,
			ctx:         th.SharedAdminUser.UserCtx,
			expectedErr: "time cannot be in the past",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if len(tc.files) > 0 {
				th.ExpectUploadNillable(t, suite.Client.MockProvider, tc.files)
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

			if tc.policyID != "" {
				assert.Check(t, is.Len(resp.UpdateEvidence.Evidence.InternalPolicies.Edges, 1))
				assert.Check(t, is.Equal(tc.policyID, resp.UpdateEvidence.Evidence.InternalPolicies.Edges[0].Node.ID))
			} else if tc.request.ClearInternalPolicies != nil && *tc.request.ClearInternalPolicies {
				assert.Check(t, is.Len(resp.UpdateEvidence.Evidence.InternalPolicies.Edges, 0))
			}

			if tc.procedureID != "" {
				assert.Check(t, is.Len(resp.UpdateEvidence.Evidence.Procedures.Edges, 1))
				assert.Check(t, is.Equal(tc.procedureID, resp.UpdateEvidence.Evidence.Procedures.Edges[0].Node.ID))
			} else if tc.request.ClearProcedures != nil && *tc.request.ClearProcedures {
				assert.Check(t, is.Len(resp.UpdateEvidence.Evidence.Procedures.Edges, 0))
			}
		})
	}

	// delete created evidence
	(&th.Cleanup[*generated.EvidenceDeleteOne]{Client: suite.Client.DB.Evidence, ID: evidence.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete created program and membership
	(&th.Cleanup[*generated.ProgramMembershipDeleteOne]{Client: suite.Client.DB.ProgramMembership, ID: pm.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProgramDeleteOne]{Client: suite.Client.DB.Program, ID: program.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: []string{internalPolicy1.ID, internalPolicy2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ProcedureDeleteOne]{Client: suite.Client.DB.Procedure, IDs: []string{procedure1.ID, procedure2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteEvidence(t *testing.T) {
	// create objects to be deleted
	evidence1 := (&th.EvidenceBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	evidence2 := (&th.EvidenceBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)

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
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: evidence1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedAdminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  evidence1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: evidence2.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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

func TestEvidenceMissingArtifactStatus(t *testing.T) {
	pngFile := th.UploadFile(t, th.LogoFilePath)

	testCases := []struct {
		name           string
		createInput    testclient.CreateEvidenceInput
		createFiles    []*graphql.Upload
		updateInput    *testclient.UpdateEvidenceInput
		expectedStatus enums.EvidenceStatus
		description    string
	}{
		{
			name: "create evidence without files and without URL should set MISSING_ARTIFACT",
			createInput: testclient.CreateEvidenceInput{
				Name: "Evidence without artifacts",
			},
			expectedStatus: enums.EvidenceStatusMissingArtifact,
			description:    "Evidence created without files or URL and without explicit status should have MISSING_ARTIFACT status",
		},
		{
			name: "create evidence with files should not set MISSING_ARTIFACT",
			createInput: testclient.CreateEvidenceInput{
				Name: "Evidence with files",
			},
			createFiles: []*graphql.Upload{
				pngFile,
			},
			expectedStatus: enums.EvidenceStatusSubmitted,
			description:    "Evidence created with files should not have MISSING_ARTIFACT status",
		},
		{
			name: "create evidence with URL should not set MISSING_ARTIFACT",
			createInput: testclient.CreateEvidenceInput{
				Name: "Evidence with URL",
				URL:  lo.ToPtr("https://example.com/evidence.pdf"),
			},
			expectedStatus: enums.EvidenceStatusSubmitted,
			description:    "Evidence created with URL should not have MISSING_ARTIFACT status",
		},
		{
			name: "create evidence with both files and URL should not set MISSING_ARTIFACT",
			createInput: testclient.CreateEvidenceInput{
				Name: "Evidence with files and URL",
				URL:  lo.ToPtr("https://example.com/evidence.pdf"),
			},
			createFiles: []*graphql.Upload{
				pngFile,
			},
			expectedStatus: enums.EvidenceStatusSubmitted,
			description:    "Evidence created with both files and URL should not have MISSING_ARTIFACT status",
		},
		{
			name: "update evidence to clear URL when no files should set MISSING_ARTIFACT",
			createInput: testclient.CreateEvidenceInput{
				Name: "Evidence with URL only",
				URL:  lo.ToPtr("https://example.com/evidence.pdf"),
			},
			updateInput: &testclient.UpdateEvidenceInput{
				ClearURL: lo.ToPtr(true),
			},
			expectedStatus: enums.EvidenceStatusMissingArtifact,
			description:    "Evidence updated to clear URL when no files and without explicit status should have MISSING_ARTIFACT status",
		},
		{
			name: "update evidence to add URL should clear MISSING_ARTIFACT",
			createInput: testclient.CreateEvidenceInput{
				Name: "Evidence without artifacts",
			},
			updateInput: &testclient.UpdateEvidenceInput{
				URL: lo.ToPtr("https://example.com/evidence.pdf"),
			},
			expectedStatus: enums.EvidenceStatusSubmitted,
			description:    "Evidence updated to add URL should not have MISSING_ARTIFACT status",
		},
		{
			name: "create evidence without files and without URL but with explicit status should respect explicit status",
			createInput: testclient.CreateEvidenceInput{
				Name:   "Evidence without artifacts but explicit status",
				Status: lo.ToPtr(enums.EvidenceStatusSubmitted),
			},
			expectedStatus: enums.EvidenceStatusSubmitted,
			description:    "Explicit status should always be respected, even when evidence has no files or URL",
		},
		{
			name: "create evidence with files but explicit MISSING_ARTIFACT status should respect explicit status",
			createInput: testclient.CreateEvidenceInput{
				Name:   "Evidence with files but explicit MISSING_ARTIFACT",
				Status: lo.ToPtr(enums.EvidenceStatusMissingArtifact),
			},
			createFiles: []*graphql.Upload{
				pngFile,
			},
			expectedStatus: enums.EvidenceStatusMissingArtifact,
			description:    "Explicit status should always be respected, even when evidence has files",
		},
		{
			name: "update evidence without files and without URL but with explicit status should respect explicit status",
			createInput: testclient.CreateEvidenceInput{
				Name: "Evidence without artifacts",
			},
			updateInput: &testclient.UpdateEvidenceInput{
				Status: lo.ToPtr(enums.EvidenceStatusSubmitted),
			},
			expectedStatus: enums.EvidenceStatusSubmitted,
			description:    "Explicit status should always be respected on update, even when evidence has no files or URL",
		},
		{
			name: "update evidence to clear URL when no files but with explicit status should respect explicit status",
			createInput: testclient.CreateEvidenceInput{
				Name: "Evidence with URL only",
				URL:  lo.ToPtr("https://example.com/evidence.pdf"),
			},
			updateInput: &testclient.UpdateEvidenceInput{
				ClearURL: lo.ToPtr(true),
				Status:   lo.ToPtr(enums.EvidenceStatusSubmitted),
			},
			expectedStatus: enums.EvidenceStatusSubmitted,
			description:    "Explicit status should always be respected on update, even when clearing URL and no files remain",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.createFiles) > 0 {
				th.ExpectUploadNillable(t, suite.Client.MockProvider, tc.createFiles)
			}

			createResp, err := suite.Client.API.CreateEvidence(th.SharedAdminUser.UserCtx, tc.createInput, tc.createFiles)
			assert.NilError(t, err)
			assert.Assert(t, createResp != nil)

			evidenceID := createResp.CreateEvidence.Evidence.ID

			if tc.updateInput == nil {
				assert.Check(t, is.Equal(tc.expectedStatus, *createResp.CreateEvidence.Evidence.Status), tc.description)
			} else {
				updateResp, err := suite.Client.API.UpdateEvidence(th.SharedAdminUser.UserCtx, evidenceID, *tc.updateInput, nil)
				assert.NilError(t, err)
				assert.Assert(t, updateResp != nil)

				assert.Check(t, is.Equal(tc.expectedStatus, *updateResp.UpdateEvidence.Evidence.Status), tc.description)
			}

			evidenceResp, err := suite.Client.API.GetEvidenceByID(th.SharedAdminUser.UserCtx, evidenceID)
			if err == nil && evidenceResp != nil {
				for _, edge := range evidenceResp.Evidence.Files.Edges {
					(&th.Cleanup[*generated.FileDeleteOne]{Client: suite.Client.DB.File, ID: edge.Node.ID}).MustDelete(th.SharedAdminUser.UserCtx, t)
				}
			}
			(&th.Cleanup[*generated.EvidenceDeleteOne]{Client: suite.Client.DB.Evidence, ID: evidenceID}).MustDelete(th.SharedAdminUser.UserCtx, t)
		})
	}
}

func TestEvidence_NextReviewDate(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	initialCreationDate := models.DateTime(now.AddDate(0, 0, -14))
	updatedCreationDate := models.DateTime(now.AddDate(0, 0, -7))
	monthlyFreq := enums.FrequencyMonthly
	quarterlyFreq := enums.FrequencyQuarterly

	createResp, err := suite.Client.API.CreateEvidence(th.SharedAdminUser.UserCtx, testclient.CreateEvidenceInput{
		Name:            "Evidence review date",
		CreationDate:    lo.ToPtr(initialCreationDate),
		ReviewFrequency: lo.ToPtr(monthlyFreq),
	}, nil)
	assert.NilError(t, err)
	assert.Assert(t, createResp != nil)

	id := createResp.CreateEvidence.Evidence.ID

	bulkUpdateResp, err := suite.Client.API.CreateEvidence(th.SharedAdminUser.UserCtx, testclient.CreateEvidenceInput{
		Name:            "Bulk evidence review date",
		CreationDate:    lo.ToPtr(initialCreationDate),
		ReviewFrequency: lo.ToPtr(quarterlyFreq),
	}, nil)
	assert.NilError(t, err)
	assert.Assert(t, bulkUpdateResp != nil)

	bulkUpdateID := bulkUpdateResp.CreateEvidence.Evidence.ID

	resp, err := suite.Client.API.GetEvidenceByID(th.SharedAdminUser.UserCtx, id)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Assert(t, resp.Evidence.RenewalDate != nil)
	assert.Check(t, is.Equal(time.Time(initialCreationDate).AddDate(0, 1, 0), time.Time(*resp.Evidence.RenewalDate)))
	assert.Assert(t, resp.Evidence.ReviewFrequency != nil)
	assert.Check(t, is.Equal(monthlyFreq, *resp.Evidence.ReviewFrequency))

	_, err = suite.Client.API.UpdateEvidence(th.SharedAdminUser.UserCtx, id, testclient.UpdateEvidenceInput{
		CreationDate: lo.ToPtr(updatedCreationDate),
	}, nil)
	assert.NilError(t, err)

	resp, err = suite.Client.API.GetEvidenceByID(th.SharedAdminUser.UserCtx, id)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Assert(t, resp.Evidence.RenewalDate != nil)
	assert.Check(t, is.Equal(time.Time(updatedCreationDate).AddDate(0, 1, 0), time.Time(*resp.Evidence.RenewalDate)))

	_, err = suite.Client.API.UpdateEvidence(th.SharedAdminUser.UserCtx, id, testclient.UpdateEvidenceInput{
		ReviewFrequency: lo.ToPtr(quarterlyFreq),
	}, nil)
	assert.NilError(t, err)

	resp, err = suite.Client.API.GetEvidenceByID(th.SharedAdminUser.UserCtx, id)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Assert(t, resp.Evidence.RenewalDate != nil)
	assert.Check(t, is.Equal(time.Time(updatedCreationDate).AddDate(0, 3, 0), time.Time(*resp.Evidence.RenewalDate)))

	updateBulkResp, err := suite.Client.API.UpdateBulkEvidence(th.SharedAdminUser.UserCtx, []string{id, bulkUpdateID}, testclient.UpdateEvidenceInput{
		ReviewFrequency: lo.ToPtr(monthlyFreq),
	})
	assert.NilError(t, err)
	assert.Assert(t, updateBulkResp != nil)
	assert.Check(t, is.Len(updateBulkResp.UpdateBulkEvidence.UpdatedIDs, 2))

	expectedRenewalDates := map[string]time.Time{
		id:           time.Time(updatedCreationDate).AddDate(0, 1, 0),
		bulkUpdateID: time.Time(initialCreationDate).AddDate(0, 1, 0),
	}

	for evidenceID, expectedRenewalDate := range expectedRenewalDates {
		resp, err = suite.Client.API.GetEvidenceByID(th.SharedAdminUser.UserCtx, evidenceID)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Assert(t, resp.Evidence.RenewalDate != nil)
		assert.Check(t, is.Equal(expectedRenewalDate, time.Time(*resp.Evidence.RenewalDate)))
	}

	(&th.Cleanup[*generated.EvidenceDeleteOne]{Client: suite.Client.DB.Evidence, IDs: []string{id, bulkUpdateID}}).MustDelete(th.SharedAdminUser.UserCtx, t)
}
