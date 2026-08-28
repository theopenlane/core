package eventstest_test

import (
	"context"
	"strings"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestMutationCreateTrustCenterNDARequest(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())
	trustCenterNoApproval := tcOrg.TrustCenter

	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate(), th.WithAllUserTypes())
	trustCenterWithApproval := tcOrg2.TrustCenter

	_, err := suite.Client.API.UpdateTrustCenter(tcOrg2.Admin.UserCtx, trustCenterWithApproval.ID, testclient.UpdateTrustCenterInput{
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
	approvalEmail := "Pending Approval"
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
			client:          suite.Client.API,
			ctx:             tcOrg.Owner.UserCtx,
			expectedStatus:  enums.TrustCenterNDARequestStatusRequested,
			expectEmailSent: ndaEmail,
		},
		{
			name:            "happy path - resend request with no approval required, status should be REQUESTED",
			input:           noApprovalRequiredRequest,
			client:          suite.Client.API,
			ctx:             tcOrg.Owner.UserCtx,
			expectedStatus:  enums.TrustCenterNDARequestStatusRequested,
			expectEmailSent: ndaEmail,
			setStatus:       &enums.TrustCenterNDARequestStatusSigned,
		},
		{
			name:                   "happy path - approval required, status should be NEEDS_APPROVAL, set to approved",
			input:                  emailApprovedRequest,
			client:                 suite.Client.API,
			ctx:                    tcOrg2.Owner.UserCtx,
			expectedStatus:         enums.TrustCenterNDARequestStatusNeedsApproval,
			expectEmailSent:        approvalEmail, // approvers notified the request is pending approval
			setStatus:              &enums.TrustCenterNDARequestStatusApproved,
			expectedSecondaryEmail: ndaEmail, // should get nda email
		},
		{
			name:            "happy path - sign after approval",
			input:           emailApprovedRequest,
			client:          suite.Client.API,
			ctx:             tcOrg2.Admin.UserCtx,
			expectedStatus:  enums.TrustCenterNDARequestStatusApproved,
			expectEmailSent: ndaEmail,
			setStatus:       &enums.TrustCenterNDARequestStatusSigned,
		},
		{
			name:            "happy path - re-request after approval",
			input:           emailApprovedRequest,
			client:          suite.Client.API,
			ctx:             tcOrg2.SuperAdmin.UserCtx,
			expectedStatus:  enums.TrustCenterNDARequestStatusSigned,
			expectEmailSent: authEmail, // no email because already signed and not updating in the request here

		},
		{
			name:                   "happy path - approval required, status should be NEEDS_APPROVAL, set to declined for next test",
			input:                  emailDeclinedRequest,
			client:                 suite.Client.API,
			ctx:                    tcOrg2.Owner.UserCtx,
			expectedStatus:         enums.TrustCenterNDARequestStatusNeedsApproval,
			expectEmailSent:        approvalEmail, // approvers notified the request is pending approval
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
			client:          suite.Client.API,
			ctx:             tcOrg.Owner.UserCtx,
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
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "user cannot create in another org's trust center",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenterWithApproval.ID,
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "invalid email",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         "invalid-email",
				TrustCenterID: &trustCenterNoApproval.ID,
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
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
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: "first_name",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// Clear any existing jobs and emails
			err := suite.Client.DB.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			mockEmailSender().Reset()

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

			// Verify the email was or was not sent based on expectation
			waitForEvents()

			if tc.expectEmailSent != "" {
				msgs := mockEmailSender().Messages()
				assert.Assert(t, len(msgs) == 1, "expected 1 email, got %d", len(msgs))

				found := strings.Contains(msgs[0].Subject, tc.expectEmailSent) ||
					strings.Contains(msgs[0].HTML, tc.expectEmailSent) ||
					strings.Contains(msgs[0].Text, tc.expectEmailSent)
				assert.Assert(t, found, "expected email containing '%s' to be sent", tc.expectEmailSent)
			} else {
				msgs := mockEmailSender().Messages()
				assert.Assert(t, len(msgs) == 0, "expected no emails, got %d", len(msgs))
			}

			if tc.setStatus != nil || tc.setEmptyStatus {
				// Clear any existing jobs and emails
				err = suite.Client.DB.Job.TruncateRiverTables(tc.ctx)
				assert.NilError(t, err)

				mockEmailSender().Reset()

				resp, err := suite.Client.API.UpdateTrustCenterNDARequest(tc.ctx, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
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

				waitForEvents()

				if tc.expectedSecondaryEmail != "" {
					msgs := mockEmailSender().Messages()
					assert.Assert(t, len(msgs) == 1, "expected 1 email, got %d", len(msgs))

					found := strings.Contains(msgs[0].Subject, tc.expectedSecondaryEmail) ||
						strings.Contains(msgs[0].HTML, tc.expectedSecondaryEmail) ||
						strings.Contains(msgs[0].Text, tc.expectedSecondaryEmail)
					assert.Assert(t, found, "expected email containing '%s' to be sent", tc.expectedSecondaryEmail)
				} else {
					msgs := mockEmailSender().Messages()
					assert.Assert(t, len(msgs) == 0, "expected no emails, got %d", len(msgs))
				}
			}

			if tc.input.TrustCenterID == &trustCenterNoApproval.ID {
				reqCleanupOrg1 = append(reqCleanupOrg1, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
			} else if tc.input.TrustCenterID == &trustCenterWithApproval.ID {
				reqCleanupOrg2 = append(reqCleanupOrg2, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
			}
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestMutationUpdateTrustCenterNDARequest(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())
	trustCenter := tcOrg.TrustCenter

	_, err := suite.Client.API.UpdateTrustCenter(tcOrg.Owner.UserCtx, trustCenter.ID, testclient.UpdateTrustCenterInput{
		UpdateTrustCenterSetting: &testclient.UpdateTrustCenterSettingInput{
			NdaApprovalRequired: lo.ToPtr(true),
		},
	})
	assert.NilError(t, err)

	ndaRequest, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
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
			client:          suite.Client.API,
			ctx:             tcOrg.Owner.UserCtx,
			expectEmailSent: false,
		},
		{
			name: "happy path - update status to approved",
			input: testclient.UpdateTrustCenterNDARequestInput{
				Status: lo.ToPtr(enums.TrustCenterNDARequestStatusApproved),
			},
			client:          suite.Client.API,
			ctx:             tcOrg.Admin.UserCtx,
			expectEmailSent: true,
		},
		{
			name: "view only user cannot update",
			input: testclient.UpdateTrustCenterNDARequestInput{
				FirstName: lo.ToPtr("ShouldNotUpdate"),
			},
			client:          suite.Client.API,
			ctx:             tcOrg.Member.UserCtx,
			expectedErr:     th.NotAuthorizedErrorMsg,
			expectEmailSent: false,
		},
		{
			name: "different org cannot update",
			input: testclient.UpdateTrustCenterNDARequestInput{
				FirstName: lo.ToPtr("ShouldNotUpdate"),
			},
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedErr:     th.NotFoundErrorMsg,
			expectEmailSent: false,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			// Clear any existing jobs and emails
			err := suite.Client.DB.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			mockEmailSender().Reset()

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

			// Verify the email was or was not sent based on expectation
			waitForEvents()

			if tc.expectEmailSent {
				msgs := mockEmailSender().Messages()
				assert.Assert(t, len(msgs) == 1, "expected 1 email, got %d", len(msgs))
			} else {
				msgs := mockEmailSender().Messages()
				assert.Assert(t, len(msgs) == 0, "expected no emails, got %d", len(msgs))
			}
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationTrustCenterNDARequestApprovalEmailsUseConfiguredGroup(t *testing.T) {
	trustcenterOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate(), th.WithAllUserTypes())
	trustCenter := trustcenterOrg.TrustCenter

	group := (&th.GroupBuilder{Client: suite.Client}).MustNew(trustcenterOrg.Owner.UserCtx, t)
	(&th.GroupMemberBuilder{Client: suite.Client, GroupID: group.ID, UserID: trustcenterOrg.Member.ID}).MustNew(trustcenterOrg.Owner.UserCtx, t)

	_, err := suite.Client.API.UpdateTrustCenter(trustcenterOrg.Owner.UserCtx, trustCenter.ID, testclient.UpdateTrustCenterInput{
		UpdateTrustCenterSetting: &testclient.UpdateTrustCenterSettingInput{
			NdaApprovalRequired: lo.ToPtr(true),
			NdaApproverGroupID:  &group.ID,
		},
	})
	assert.NilError(t, err)

	err = suite.Client.DB.Job.TruncateRiverTables(trustcenterOrg.Owner.UserCtx)
	assert.NilError(t, err)
	mockEmailSender().Reset()

	req, err := suite.Client.API.CreateTrustCenterNDARequest(trustcenterOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)
	assert.Equal(t, enums.TrustCenterNDARequestStatusNeedsApproval, *req.CreateTrustCenterNDARequest.TrustCenterNDARequest.Status)

	waitForEvents()

	msgs := mockEmailSender().Messages()
	assert.Assert(t, len(msgs) == 1, "expected 1 email, got multiple ( %d )", len(msgs))
	assert.Assert(t, lo.Contains(msgs[0].To, trustcenterOrg.Member.UserInfo.Email), "expected approval email to go to configured group member")
	assert.Assert(t, !lo.Contains(msgs[0].To, trustcenterOrg.Owner.UserInfo.Email), "expected owner not to receive approval email when group is configured")
	assert.Assert(t, !lo.Contains(msgs[0].To, trustcenterOrg.Admin.UserInfo.Email), "expected admin not to receive approval email when group is configured")
	assert.Assert(t, !lo.Contains(msgs[0].To, trustcenterOrg.SuperAdmin.UserInfo.Email), "expected super admin not to receive approval email when group is configured")

	th.CleanupOrganizationDataWithContext(trustcenterOrg.Owner.UserCtx, t)
}

func TestMutationTrustCenterNDARequestApprovalEmailsFallBackToApproverRoles(t *testing.T) {
	trustcenterOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate(), th.WithAllUserTypes())
	trustCenter := trustcenterOrg.TrustCenter

	_, err := suite.Client.API.UpdateTrustCenter(trustcenterOrg.Owner.UserCtx, trustCenter.ID, testclient.UpdateTrustCenterInput{
		UpdateTrustCenterSetting: &testclient.UpdateTrustCenterSettingInput{
			NdaApprovalRequired: lo.ToPtr(true),
		},
	})
	assert.NilError(t, err)

	err = suite.Client.DB.Job.TruncateRiverTables(trustcenterOrg.Owner.UserCtx)
	assert.NilError(t, err)
	mockEmailSender().Reset()

	req, err := suite.Client.API.CreateTrustCenterNDARequest(trustcenterOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)
	assert.Equal(t, enums.TrustCenterNDARequestStatusNeedsApproval, *req.CreateTrustCenterNDARequest.TrustCenterNDARequest.Status)
	waitForEvents()

	msgs := mockEmailSender().Messages()
	assert.Assert(t, len(msgs) == 1, "expected 1 email, got multiple ( %d )", len(msgs))
	assert.Assert(t, lo.Contains(msgs[0].To, trustcenterOrg.Owner.UserInfo.Email), "expected approval email to go to owner when no group is configured")
	assert.Assert(t, lo.Contains(msgs[0].To, trustcenterOrg.Admin.UserInfo.Email), "expected approval email to go to admin when no group is configured")
	assert.Assert(t, lo.Contains(msgs[0].To, trustcenterOrg.SuperAdmin.UserInfo.Email), "expected approval email to go to super admin when no group is configured")
	assert.Assert(t, !lo.Contains(msgs[0].To, trustcenterOrg.Member.UserInfo.Email), "expected member not to receive approval email when no group is configured")

	th.CleanupOrganizationDataWithContext(trustcenterOrg.Owner.UserCtx, t)
}

func TestMutationCreateTrustCenterNDARequestAsAnonymousUser(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())
	trustCenter := tcOrg.TrustCenter
	pdfHash := th.GetMD5Hash(t, th.PdfFilePath)

	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())
	otherTrustCenter := tcOrg2.TrustCenter

	anonEmail := gofakeit.Email()
	anonCtx, anonUser := th.CreateAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, anonEmail)
	wrongTrustCenterAnonCtx := th.CreateAnonymousTrustCenterContext(otherTrustCenter.ID, otherTrustCenter.OwnerID)

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
			client:          suite.Client.API,
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
			client:          suite.Client.API,
			ctx:             anonCtx,
			expectedErr:     th.NotAuthorizedErrorMsg,
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
			client:          suite.Client.API,
			ctx:             wrongTrustCenterAnonCtx,
			expectedErr:     th.NotAuthorizedErrorMsg,
			expectEmailSent: false,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// Clear any existing jobs and emails
			err := suite.Client.DB.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			mockEmailSender().Reset()

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

			// Verify the email was or was not sent based on expectation
			waitForEvents()

			if tc.expectEmailSent {
				msgs := mockEmailSender().Messages()
				assert.Assert(t, len(msgs) == 1, "expected 1 email, got %d", len(msgs))
			} else {
				msgs := mockEmailSender().Messages()
				assert.Assert(t, len(msgs) == 0, "expected no emails, got %d", len(msgs))
			}

			if tc.testResponse {
				th.ExpectAttestedUpload(t, suite.Client.MockProvider)

				// now sign the nda to ensure status is set correctly
				_, err = suite.Client.API.SubmitTrustCenterNDAResponse(anonCtx, testclient.SubmitTrustCenterNDAResponseInput{
					TemplateID: *tcOrg.NDATemplateID,
					Response: map[string]any{
						"signatory_info": map[string]any{
							"email": anonUser.SubjectEmail,
						},
						"acknowledgment": true,
						"signature_metadata": map[string]any{
							"ip_address": "192.168.1.100",
							"timestamp":  "2025-09-22T19:37:59.988Z",
							"pdf_hash":   pdfHash,
							"user_id":    anonUser.SubjectID,
						},
						"pdf_file_id":     *tcOrg.NDAFileID,
						"trust_center_id": trustCenter.ID,
					},
				})
				assert.NilError(t, err)

				// Fetch the updated request as org owner to verify status — anon TC users cannot read NDA requests
				updatedReq, err := suite.Client.API.GetTrustCenterNDARequestByID(tcOrg.Owner.UserCtx, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
				assert.NilError(t, err)
				assert.Equal(t, enums.TrustCenterNDARequestStatusSigned, *updatedReq.TrustCenterNDARequest.Status)
			}
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestMutationRequestNewTrustCenterToken(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())
	trustCenter := tcOrg.TrustCenter

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

	ndaRequestSigned, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, ndaSigned)
	assert.NilError(t, err)

	_, err = suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, ndaRequested)
	assert.NilError(t, err)

	ndaRequestNeedsApproval, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, ndaNeedsApproval)
	assert.NilError(t, err)

	_, err = suite.Client.API.UpdateTrustCenterNDARequest(tcOrg.Owner.UserCtx, ndaRequestSigned.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
		Status: lo.ToPtr(enums.TrustCenterNDARequestStatusSigned),
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.UpdateTrustCenterNDARequest(tcOrg.Owner.UserCtx, ndaRequestNeedsApproval.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
		Status: lo.ToPtr(enums.TrustCenterNDARequestStatusNeedsApproval),
	})
	assert.NilError(t, err)

	anonCtxSigned, _ := th.CreateAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, ndaSigned.Email)
	anonCtxRequested, _ := th.CreateAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, ndaRequested.Email)
	anonCtxNeedsApproval, _ := th.CreateAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, ndaNeedsApproval.Email)
	anonCtxRandom, _ := th.CreateAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, gofakeit.Email())

	ndaEmail := "Trust Center NDA Request"
	authEmail := "Access"
	approvalEmail := "Pending Approval"
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
			client:          suite.Client.API,
			ctx:             anonCtxSigned,
			expectEmailSent: authEmail,
		},
		{
			name:            "happy path - not signed, resends nda email",
			email:           ndaRequested.Email,
			client:          suite.Client.API,
			ctx:             anonCtxRequested,
			expectEmailSent: ndaEmail,
		},
		{
			name:            "needs approval, approver notification email sent, requester not emailed because not approved yet",
			email:           ndaNeedsApproval.Email,
			client:          suite.Client.API,
			ctx:             anonCtxNeedsApproval,
			expectEmailSent: approvalEmail,
		},
		{
			name:   "no nda request, no-op",
			email:  gofakeit.Email(),
			client: suite.Client.API,
			ctx:    anonCtxRandom,
		},
		{
			name:        "not anonymous context, error",
			email:       gofakeit.Email(),
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// Clear any existing jobs and emails
			err := suite.Client.DB.Job.TruncateRiverTables(tc.ctx)
			assert.NilError(t, err)

			mockEmailSender().Reset()

			resp, err := tc.client.RequestNewTrustCenterToken(tc.ctx, tc.email)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, resp.RequestNewTrustCenterToken.Success == true)

			// Verify the email was or was not sent based on expectation
			waitForEvents()

			if tc.expectEmailSent != "" {
				msgs := mockEmailSender().Messages()
				assert.Assert(t, len(msgs) == 1, "expected 1 email, got %d", len(msgs))

				found := strings.Contains(msgs[0].Subject, tc.expectEmailSent) ||
					strings.Contains(msgs[0].HTML, tc.expectEmailSent) ||
					strings.Contains(msgs[0].Text, tc.expectEmailSent)
				assert.Assert(t, found, "expected email containing '%s' to be sent", tc.expectEmailSent)
			} else {
				msgs := mockEmailSender().Messages()
				assert.Assert(t, len(msgs) == 0, "expected no emails, got %d", len(msgs))
			}
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}
