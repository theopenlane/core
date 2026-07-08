package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
)

func TestDocumentDataAnonymousTrustCenterAccess(t *testing.T) {
	tcOrg := createFreshOrgWithTrustCenter(t, withNDATemplate())
	trustCenter := tcOrg.trustCenter

	pdfHash := getMD5Hash(t, pdfFilePath)

	newAnonCtx := func(email string) (context.Context, string) {
		anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())
		caller := auth.NewTrustCenterCaller(trustCenter.OwnerID, anonUserID, "Anon User", email)
		return newAnonTrustCenterCtxFromCaller(caller, trustCenter.ID), anonUserID
	}

	submitNDA := func(ctx context.Context, anonUserID, email string) string {
		_, err := suite.client.api.CreateTrustCenterNDARequest(ctx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     "Anon",
			LastName:      "User",
			CompanyName:   lo.ToPtr("Test Company"),
			Email:         email,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)

		expectAttestedUpload(t, suite.client.mockProvider)

		resp, err := suite.client.api.SubmitTrustCenterNDAResponse(ctx, testclient.SubmitTrustCenterNDAResponseInput{
			TemplateID: *tcOrg.ndaTemplateID,
			Response: map[string]any{
				"signatory_info": map[string]any{
					"email": email,
				},
				"acknowledgment": true,
				"signature_metadata": map[string]any{
					"ip_address": "192.168.1.100",
					"timestamp":  "2025-09-22T19:37:59.988Z",
					"pdf_hash":   pdfHash,
					"user_id":    anonUserID,
				},
				"pdf_file_id":     *tcOrg.ndaFileID,
				"trust_center_id": trustCenter.ID,
			},
		})
		assert.NilError(t, err)

		return resp.SubmitTrustCenterNDAResponse.DocumentData.ID
	}

	anonCtx1, anonUserID1 := newAnonCtx("anon1@example.com")
	anonCtx2, anonUserID2 := newAnonCtx("anon2@example.com")

	docDataID1 := submitNDA(anonCtx1, anonUserID1, "anon1@example.com")
	docDataID2 := submitNDA(anonCtx2, anonUserID2, "anon2@example.com")

	testCases := []struct {
		name        string
		ctx         context.Context
		queryID     string
		expectedErr string
	}{
		{
			name:        "anon user cannot read their own document data",
			ctx:         anonCtx1,
			queryID:     docDataID1,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "anon user cannot read another user's document data",
			ctx:         anonCtx1,
			queryID:     docDataID2,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:    "org owner can read document data",
			ctx:     tcOrg.owner.UserCtx,
			queryID: docDataID1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.client.api.GetDocumentDataByID(tc.ctx, tc.queryID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tc.queryID, resp.DocumentData.ID)
		})
	}

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}
