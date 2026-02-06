package graphapi_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
)

func TestMutationCreateTrustCenterNDARequest(t *testing.T) {
	cleanupTrustCenterData(t)
	trustCenterNoApproval := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	ndaTemplate1 := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: trustCenterNoApproval.ID,
	}).MustNew(testUser1.UserCtx, t)

	trustCenterWithApproval := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	ndaTemplate2 := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: trustCenterWithApproval.ID,
	}).MustNew(testUser2.UserCtx, t)

	_, err := suite.client.api.UpdateTrustCenter(testUser2.UserCtx, trustCenterWithApproval.ID, testclient.UpdateTrustCenterInput{
		UpdateTrustCenterSetting: &testclient.UpdateTrustCenterSettingInput{
			NdaApprovalRequired: lo.ToPtr(true),
		},
	})
	assert.NilError(t, err)

	noApprovalRequiredRequest := testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenterNoApproval.ID,
	}

	emailApprovedRequest := testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenterWithApproval.ID,
	}

	emailDeclinedRequest := testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenterWithApproval.ID,
	}

	reqCleanupOrg1 := []string{}
	reqCleanupOrg2 := []string{}
	ndaEmail := "Trust Center NDA Request"
	authEmail := "Access"
	testCases := []struct {
		name                   string
		input                  testclient.CreateTrustCenterNDARequestInput
		client                 *testclient.TestClient
		ctx                    context.Context
		expectedErr            string
		expectedStatus         enums.TrustCenterNDARequestStatus
		expectEmailSent        string
		setStatus              *enums.TrustCenterNDARequestStatus
		setEmptyStatus         bool
		expectedSecondaryEmail string
	}{
		{
			name:            "happy path - no approval required, status should be REQUESTED",
			input:           noApprovalRequiredRequest,
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedStatus:  enums.TrustCenterNDARequestStatusRequested,
			expectEmailSent: ndaEmail,
		},
		{
			name:                   "happy path - resend request with no approval required, status should be REQUESTED",
			input:                  noApprovalRequiredRequest,
			client:                 suite.client.api,
			ctx:                    testUser1.UserCtx,
			expectedStatus:         enums.TrustCenterNDARequestStatusRequested,
			expectEmailSent:        ndaEmail,
			setStatus:              &enums.TrustCenterNDARequestStatusSigned,
			expectedSecondaryEmail: authEmail,
		},
		{
			name:                   "happy path - approval required, status should be NEEDS_APPROVAL, set to approved",
			input:                  emailApprovedRequest,
			client:                 suite.client.api,
			ctx:                    testUser2.UserCtx,
			expectedStatus:         enums.TrustCenterNDARequestStatusNeedsApproval,
			expectEmailSent:        "", // no email because not approved yet
			setStatus:              &enums.TrustCenterNDARequestStatusApproved,
			expectedSecondaryEmail: ndaEmail, // should get nda email
		},
		{
			name:                   "happy path - sign after approval",
			input:                  emailApprovedRequest,
			client:                 suite.client.api,
			ctx:                    testUser2.UserCtx,
			expectedStatus:         enums.TrustCenterNDARequestStatusApproved,
			expectEmailSent:        ndaEmail,
			setStatus:              &enums.TrustCenterNDARequestStatusSigned,
			expectedSecondaryEmail: authEmail,
		},
		{
			name:            "happy path - re-request after approval",
			input:           emailApprovedRequest,
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedStatus:  enums.TrustCenterNDARequestStatusSigned,
			expectEmailSent: authEmail, // no email because already signed and not updating in the request here

		},
		{
			name:                   "happy path - approval required, status should be NEEDS_APPROVAL, set to declined for next test",
			input:                  emailDeclinedRequest,
			client:                 suite.client.api,
			ctx:                    testUser2.UserCtx,
			expectedStatus:         enums.TrustCenterNDARequestStatusNeedsApproval,
			expectEmailSent:        "", // no email because not approved yet
			setStatus:              &enums.TrustCenterNDARequestStatusDeclined,
			expectedSecondaryEmail: "",
		},
		{
			name: "happy path - with company name and reason",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				CompanyName:   lo.ToPtr(gofakeit.Company()),
				Reason:        lo.ToPtr("Need access to security documentation"),
				TrustCenterID: &trustCenterNoApproval.ID,
			},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedStatus:  enums.TrustCenterNDARequestStatusRequested,
			expectEmailSent: ndaEmail,
		},
		{
			name: "view only user cannot create",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenterNoApproval.ID,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user cannot create in another org's trust center",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenterWithApproval.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "invalid email",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         "invalid-email",
				TrustCenterID: &trustCenterNoApproval.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "validator failed",
		},
		{
			name: "missing first name",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     "",
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenterNoApproval.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "first_name",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.client.db.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			resp, err := tc.client.CreateTrustCenterNDARequest(tc.ctx, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Equal(t, tc.input.FirstName, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.FirstName)
			assert.Equal(t, tc.input.LastName, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.LastName)
			assert.Equal(t, tc.input.Email, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Email)
			assert.Equal(t, tc.expectedStatus, *resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Status)

			if tc.input.CompanyName != nil {
				assert.Equal(t, *tc.input.CompanyName, *resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.CompanyName)
			}

			if tc.input.Reason != nil {
				assert.Equal(t, *tc.input.Reason, *resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Reason)
			}

			// Verify the job was or was not created based on expectation
			if tc.expectEmailSent != "" {
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
					[]rivertest.ExpectedJob{
						{
							Args: jobs.EmailArgs{},
						},
					})
				assert.Assert(t, jobs != nil)
				assert.Assert(t, is.Len(jobs, 1))

				found := false
				for _, v := range jobs {
					if strings.Contains(string(v.EncodedArgs), tc.expectEmailSent) {
						found = true
						break
					}
				}

				assert.Assert(t, found, "expected email containing '%s' to be sent", tc.expectEmailSent)
			} else {
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()), &jobs.EmailArgs{}, nil)
			}

			if tc.setStatus != nil || tc.setEmptyStatus {
				// Clear any existing jobs
				err = suite.client.db.Job.TruncateRiverTables(tc.ctx)
				assert.NilError(t, err)

				resp, err := suite.client.api.UpdateTrustCenterNDARequest(tc.ctx, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
					Status: tc.setStatus,
				})
				assert.NilError(t, err)
				assert.Assert(t, resp != nil)

				if tc.setStatus != nil {
					assert.Equal(t, *tc.setStatus, *resp.UpdateTrustCenterNDARequest.TrustCenterNDARequest.Status)
				} else {
					assert.Equal(t, tc.expectedStatus, *resp.UpdateTrustCenterNDARequest.TrustCenterNDARequest.Status)
				}

				if tc.setStatus == &enums.TrustCenterNDARequestStatusSigned {
					assert.Check(t, resp.UpdateTrustCenterNDARequest.TrustCenterNDARequest.SignedAt != nil, "signed_at should be set when status is signed")
				}

				if tc.expectedSecondaryEmail != "" {
					jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
						[]rivertest.ExpectedJob{
							{
								Args: jobs.EmailArgs{},
							},
						})
					assert.Assert(t, jobs != nil)
					assert.Assert(t, is.Len(jobs, 1))

					found := false

					for _, v := range jobs {
						if strings.Contains(string(v.EncodedArgs), tc.expectedSecondaryEmail) {
							found = true
							break
						}
					}
					assert.Assert(t, found, "expected email containing '%s' to be sent", tc.expectedSecondaryEmail)
				} else {
					rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()), &jobs.EmailArgs{}, nil)
				}
			}

			if tc.input.TrustCenterID == &trustCenterNoApproval.ID {
				reqCleanupOrg1 = append(reqCleanupOrg1, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
			} else if tc.input.TrustCenterID == &trustCenterWithApproval.ID {
				reqCleanupOrg2 = append(reqCleanupOrg2, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
			}
		})
	}

	reqCleanupOrg1 = lo.Uniq(reqCleanupOrg1)
	reqCleanupOrg2 = lo.Uniq(reqCleanupOrg2)

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, IDs: reqCleanupOrg1}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, IDs: reqCleanupOrg2}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: ndaTemplate1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: ndaTemplate2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenterNoApproval.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenterWithApproval.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestQueryTrustCenterNDARequest(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)

	viewOnlyUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &viewOnlyUser, enums.RoleMember, testUser.OrganizationID)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	ndaTemplate := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser.UserCtx, t)

	ndaRequest, err := suite.client.api.CreateTrustCenterNDARequest(testUser.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
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
			name:    "happy path",
			queryID: ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:  suite.client.api,
			ctx:     testUser.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:     "not found, different org",
			queryID:  ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found, invalid id",
			queryID:  ulids.New().String(),
			client:   suite.client.api,
			ctx:      testUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterNDARequestByID(tc.ctx, tc.queryID)
			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.queryID, resp.TrustCenterNDARequest.ID)
		})
	}

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: ndaTemplate.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)

}

func TestQueryTrustCenterNDARequests(t *testing.T) {
	cleanupTrustCenterData(t)
	testUser := suite.userBuilder(context.Background(), t)
	om := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	viewOnlyUserCtx := auth.NewTestContextWithOrgID(om.UserID, testUser.OrganizationID)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	ndaTemplate := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser.UserCtx, t)

	ndaRequest1, err := suite.client.api.CreateTrustCenterNDARequest(testUser.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	ndaRequest2, err := suite.client.api.CreateTrustCenterNDARequest(testUser.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	testCases := []struct {
		name        string
		client      *testclient.TestClient
		ctx         context.Context
		expectCount int
	}{
		{
			name:        "happy path",
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectCount: 2,
		},
		{
			name:        "happy path, view only user",
			client:      suite.client.api,
			ctx:         viewOnlyUserCtx,
			expectCount: 2,
		},
		{
			name:        "different org, no results",
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllTrustCenterNDARequests(tc.ctx, nil, nil, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.expectCount, len(resp.TrustCenterNdaRequests.Edges))
		})
	}

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: ndaRequest2.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: ndaTemplate.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser.UserCtx, t)
}

func TestMutationUpdateTrustCenterNDARequest(t *testing.T) {
	cleanupTrustCenterData(t)
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	ndaTemplate := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	_, err := suite.client.api.UpdateTrustCenter(testUser1.UserCtx, trustCenter.ID, testclient.UpdateTrustCenterInput{
		UpdateTrustCenterSetting: &testclient.UpdateTrustCenterSettingInput{
			NdaApprovalRequired: lo.ToPtr(true),
		},
	})
	assert.NilError(t, err)

	ndaRequest, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)
	assert.Equal(t, enums.TrustCenterNDARequestStatusNeedsApproval, *ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.Status)

	testCases := []struct {
		name            string
		input           testclient.UpdateTrustCenterNDARequestInput
		client          *testclient.TestClient
		ctx             context.Context
		expectEmailSent bool
		expectedErr     string
	}{
		{
			name: "happy path - update first name",
			input: testclient.UpdateTrustCenterNDARequestInput{
				FirstName: lo.ToPtr("UpdatedFirstName"),
			},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectEmailSent: false,
		},
		{
			name: "happy path - update status to approved",
			input: testclient.UpdateTrustCenterNDARequestInput{
				Status: lo.ToPtr(enums.TrustCenterNDARequestStatusApproved),
			},
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectEmailSent: true,
		},
		{
			name: "view only user cannot update",
			input: testclient.UpdateTrustCenterNDARequestInput{
				FirstName: lo.ToPtr("ShouldNotUpdate"),
			},
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedErr:     notAuthorizedErrorMsg,
			expectEmailSent: false,
		},
		{
			name: "different org cannot update",
			input: testclient.UpdateTrustCenterNDARequestInput{
				FirstName: lo.ToPtr("ShouldNotUpdate"),
			},
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedErr:     notAuthorizedErrorMsg,
			expectEmailSent: false,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.client.db.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			resp, err := tc.client.UpdateTrustCenterNDARequest(tc.ctx, ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.input.FirstName != nil {
				assert.Equal(t, *tc.input.FirstName, resp.UpdateTrustCenterNDARequest.TrustCenterNDARequest.FirstName)
			}

			if tc.input.Status != nil {
				assert.Equal(t, *tc.input.Status, *resp.UpdateTrustCenterNDARequest.TrustCenterNDARequest.Status)

				if *tc.input.Status == enums.TrustCenterNDARequestStatusApproved {
					assert.Check(t, resp.UpdateTrustCenterNDARequest.TrustCenterNDARequest.ApprovedAt != nil, "approved_at should be set when status is approved")
				}
			}

			// Verify the job was or was not created based on expectation
			if tc.expectEmailSent {
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
					[]rivertest.ExpectedJob{
						{
							Args: jobs.EmailArgs{},
						},
					})
				assert.Assert(t, jobs != nil)
				assert.Assert(t, is.Len(jobs, 1))
			} else {
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()), &jobs.EmailArgs{}, nil)
			}
		})
	}

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: ndaTemplate.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateTrustCenterNDARequestAsAnonymousUser(t *testing.T) {
	cleanupTrustCenterData(t)
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	otherTrustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	ndaTemplate := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	otherNdaTemplate := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: otherTrustCenter.ID,
	}).MustNew(testUser2.UserCtx, t)

	anonEmail := gofakeit.Email()
	anonCtx, anonUser := createAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, anonEmail)
	wrongTrustCenterAnonCtx := createAnonymousTrustCenterContext(otherTrustCenter.ID, otherTrustCenter.OwnerID)

	companyName := gofakeit.Company()

	testCases := []struct {
		name            string
		input           testclient.CreateTrustCenterNDARequestInput
		client          *testclient.TestClient
		ctx             context.Context
		expectedErr     string
		expectedStatus  enums.TrustCenterNDARequestStatus
		expectEmailSent bool
		testResponse    bool
	}{
		{
			name: "happy path - anonymous user can create NDA request, email sent because approval not required, test signed response",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				CompanyName:   &companyName,
				Email:         anonEmail,
				TrustCenterID: &trustCenter.ID,
			},
			client:          suite.client.api,
			ctx:             anonCtx,
			expectedStatus:  enums.TrustCenterNDARequestStatusRequested,
			expectEmailSent: true,
			testResponse:    true,
		},
		{
			name: "anonymous user cannot create NDA request for different trust center",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				CompanyName:   &companyName,
				Email:         gofakeit.Email(),
				TrustCenterID: &otherTrustCenter.ID,
			},
			client:          suite.client.api,
			ctx:             anonCtx,
			expectedErr:     notAuthorizedErrorMsg,
			expectEmailSent: false,
		},
		{
			name: "anonymous user with wrong trust center context cannot create",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				CompanyName:   &companyName,
				TrustCenterID: &trustCenter.ID,
			},
			client:          suite.client.api,
			ctx:             wrongTrustCenterAnonCtx,
			expectedErr:     notAuthorizedErrorMsg,
			expectEmailSent: false,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.client.db.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			resp, err := tc.client.CreateTrustCenterNDARequest(tc.ctx, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Equal(t, tc.input.FirstName, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.FirstName)
			assert.Equal(t, tc.input.LastName, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.LastName)
			assert.Equal(t, tc.input.Email, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Email)
			assert.Equal(t, tc.expectedStatus, *resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Status)

			// Verify the job was or was not created based on expectation
			if tc.expectEmailSent {
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
					[]rivertest.ExpectedJob{
						{
							Args: jobs.EmailArgs{},
						},
					})
				assert.Assert(t, jobs != nil)
				assert.Assert(t, is.Len(jobs, 1))
			} else {
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()), &jobs.EmailArgs{}, nil)
			}

			if tc.testResponse {
				// now sign the nda to ensure status is set correctly
				_, err = suite.client.api.SubmitTrustCenterNDAResponse(anonCtx, testclient.SubmitTrustCenterNDAResponseInput{
					TemplateID: ndaTemplate.ID,
					Response: map[string]any{
						"signatory_info": map[string]any{
							"email": anonUser.SubjectEmail,
						},
						"acknowledgment": true,
						"signature_metadata": map[string]any{
							"ip_address": "192.168.1.100",
							"timestamp":  "2025-09-22T19:37:59.988Z",
							"pdf_hash":   "a1b2c3d4e5f6789012345678901234567890abcd",
							"user_id":    anonUser.SubjectID,
						},
						"pdf_file_id":     "some-pdf-file-id",
						"trust_center_id": trustCenter.ID,
					},
				})
				assert.NilError(t, err)

				// Fetch the updated request to verify status
				updatedReq, err := suite.client.api.GetTrustCenterNDARequestByID(tc.ctx, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
				assert.NilError(t, err)
				assert.Equal(t, enums.TrustCenterNDARequestStatusSigned, *updatedReq.TrustCenterNDARequest.Status)
			}

			(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: ndaTemplate.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: otherNdaTemplate.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: otherTrustCenter.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateTrustCenterNDARequestDuplicateEmail(t *testing.T) {
	cleanupTrustCenterData(t)
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	ndaTemplate := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	email := gofakeit.Email()

	originalRequest, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         email,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	t.Run("duplicate email returns existing request with REQUESTED status", func(t *testing.T) {
		resp, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         email,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	})

	t.Run("duplicate email returns existing request with NEEDS_APPROVAL status", func(t *testing.T) {
		_, err := suite.client.api.UpdateTrustCenterNDARequest(testUser1.UserCtx, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
			Status: lo.ToPtr(enums.TrustCenterNDARequestStatusNeedsApproval),
		})
		assert.NilError(t, err)

		resp, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         email,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	})

	t.Run("duplicate email returns existing request with APPROVED status", func(t *testing.T) {
		_, err := suite.client.api.UpdateTrustCenterNDARequest(testUser1.UserCtx, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
			Status: lo.ToPtr(enums.TrustCenterNDARequestStatusApproved),
		})
		assert.NilError(t, err)

		resp, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         email,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	})

	t.Run("different email creates new request", func(t *testing.T) {
		resp, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         gofakeit.Email(),
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Assert(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID != resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)

		(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	})

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: ndaTemplate.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteTrustCenterNDARequest(t *testing.T) {
	cleanupTrustCenterData(t)
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	ndaTemplate := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	ndaRequest1, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	ndaRequest2, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "view only user cannot delete",
			idToDelete:  ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "different org cannot delete",
			idToDelete:  ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path - delete nda request",
			idToDelete: ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:     suite.client.api,
			ctx:        adminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "invalid id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustCenterNDARequest(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.idToDelete, resp.DeleteTrustCenterNDARequest.DeletedID)
		})
	}

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: ndaRequest2.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: ndaTemplate.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationRequestNewTrustCenterToken(t *testing.T) {
	cleanupTrustCenterData(t)
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	ndaTemplate := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	ndaSigned := testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	}

	ndaRequested := testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	}

	ndaNeedsApproval := testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	}

	ndaRequestSigned, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, ndaSigned)
	assert.NilError(t, err)

	ndaRequestRequested, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, ndaRequested)
	assert.NilError(t, err)

	ndaRequestNeedsApproval, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, ndaNeedsApproval)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateTrustCenterNDARequest(testUser1.UserCtx, ndaRequestSigned.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
		Status: lo.ToPtr(enums.TrustCenterNDARequestStatusSigned),
	})
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateTrustCenterNDARequest(testUser1.UserCtx, ndaRequestNeedsApproval.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
		Status: lo.ToPtr(enums.TrustCenterNDARequestStatusNeedsApproval),
	})
	assert.NilError(t, err)

	anonCtxSigned, _ := createAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, ndaSigned.Email)
	anonCtxRequested, _ := createAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, ndaRequested.Email)
	anonCtxNeedsApproval, _ := createAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, ndaNeedsApproval.Email)
	anonCtxRandom, _ := createAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, gofakeit.Email())

	ndaEmail := "Trust Center NDA Request"
	authEmail := "Access"
	testCases := []struct {
		name            string
		email           string
		client          *testclient.TestClient
		ctx             context.Context
		expectedErr     string
		expectEmailSent string
	}{
		{
			name:            "happy path - already signed, request new token, email sent with NDA in it",
			email:           ndaSigned.Email,
			client:          suite.client.api,
			ctx:             anonCtxSigned,
			expectEmailSent: authEmail,
		},
		{
			name:            "happy path - not signed, resends nda email",
			email:           ndaRequested.Email,
			client:          suite.client.api,
			ctx:             anonCtxRequested,
			expectEmailSent: ndaEmail,
		},
		{
			name:   "needs approval, but no email sent still because not approved yet",
			email:  ndaNeedsApproval.Email,
			client: suite.client.api,
			ctx:    anonCtxNeedsApproval,
		},
		{
			name:   "no nda request, no-op",
			email:  gofakeit.Email(),
			client: suite.client.api,
			ctx:    anonCtxRandom,
		},
		{
			name:        "not anonymous context, error",
			email:       gofakeit.Email(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// Clear any existing jobs
			err := suite.client.db.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			resp, err := tc.client.RequestNewTrustCenterToken(tc.ctx, tc.email)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, resp.RequestNewTrustCenterToken.Success == true)

			// Verify the job was or was not created based on expectation
			if tc.expectEmailSent != "" {
				jobs := rivertest.RequireManyInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()),
					[]rivertest.ExpectedJob{
						{
							Args: jobs.EmailArgs{},
						},
					})
				assert.Assert(t, jobs != nil)
				assert.Assert(t, is.Len(jobs, 1))

				found := false
				for _, v := range jobs {
					if strings.Contains(string(v.EncodedArgs), tc.expectEmailSent) {
						found = true
						break
					}
				}

				assert.Assert(t, found, "expected email containing '%s' to be sent", tc.expectEmailSent)
			} else {
				rivertest.RequireNotInserted(tc.ctx, t, riverpgxv5.New(suite.client.db.Job.GetPool()), &jobs.EmailArgs{}, nil)
			}
		})
	}

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, IDs: []string{ndaRequestSigned.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, ndaRequestRequested.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: ndaTemplate.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationRevokeNDARequestsRemovesDocAccess(t *testing.T) {
	cleanupTrustCenterData(t)

	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	ndaTemplate := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindTrustCenterNda,
		TrustCenterID: trustCenter.ID,
	}).MustNew(testUser1.UserCtx, t)

	protectedDoc := (&TrustCenterDocBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
		Visibility:    enums.TrustCenterDocumentVisibilityProtected,
	}).MustNew(testUser1.UserCtx, t)

	// create two NDA requests
	ndaReq1, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	ndaReq2, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	ndaReqIDs := []string{
		ndaReq1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
		ndaReq2.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
	}

	// simulate signed NDA access by creating anonymous contexts with subject IDs
	// matching what generateTrustCenterJWT produces: AnonTrustCenterJWTPrefix + ndaRequestID
	anonCtxs := make([]context.Context, 0, len(ndaReqIDs))

	for _, id := range ndaReqIDs {
		subjectID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, id)

		anonCtx := auth.WithAnonymousTrustCenterUser(context.Background(), &auth.AnonymousTrustCenterUser{
			SubjectID:          subjectID,
			SubjectName:        "Anonymous User",
			OrganizationID:     trustCenter.OwnerID,
			AuthenticationType: auth.JWTAuthentication,
			TrustCenterID:      trustCenter.ID,
		})

		anonCtxs = append(anonCtxs, anonCtx)

		tuple := fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   subjectID,
			SubjectType: "user",
			ObjectID:    trustCenter.ID,
			ObjectType:  "trust_center",
			Relation:    "nda_signed",
		})

		_, err := suite.client.db.Authz.WriteTupleKeys(testUser1.UserCtx, []fgax.TupleKey{tuple}, nil)
		assert.NilError(t, err)
	}

	// verify both anon users CAN see protected doc file details
	for _, ctx := range anonCtxs {
		resp, err := suite.client.api.GetTrustCenterDocByID(ctx, protectedDoc.ID)
		assert.NilError(t, err)
		assert.Assert(t, resp.TrustCenterDoc.OriginalFile != nil, "anon user should see file details before revocation")
	}

	// now we are revoking the NDA requests
	revokeResp, err := suite.client.api.DeleteBulkTrustCenterNDARequest(testUser1.UserCtx, ndaReqIDs)
	assert.NilError(t, err)
	assert.Equal(t, len(ndaReqIDs), len(revokeResp.DeleteBulkTrustCenterNDARequest.DeletedIDs))

	// anon users should not be able to see the file after we revoked their access
	for _, anonCtx := range anonCtxs {
		resp, err := suite.client.api.GetTrustCenterDocByID(anonCtx, protectedDoc.ID)
		assert.NilError(t, err)
		assert.Assert(t, resp.TrustCenterDoc.OriginalFile == nil, "anon user should NOT see file details after revocation")
	}

	// verify the NDA requests are actually deleted
	for _, id := range ndaReqIDs {
		_, err = suite.client.api.GetTrustCenterNDARequestByID(testUser1.UserCtx, id)
		assert.ErrorContains(t, err, notFoundErrorMsg)
	}

	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: protectedDoc.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: ndaTemplate.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}
