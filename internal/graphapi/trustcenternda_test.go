package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
)

func TestMutationSendTrustCenterNDAEmail(t *testing.T) {
	// Create test trust centers
	trustCenter1 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// Create anonymous trust center context helper
	createAnonymousTrustCenterContext := func(trustCenterID, organizationID string) context.Context {
		anonUserID := fmt.Sprintf("anon_%s", ulids.New().String())

		anonUser := &auth.AnonymousTrustCenterUser{
			SubjectID:          anonUserID,
			SubjectName:        "Anonymous User",
			OrganizationID:     organizationID,
			AuthenticationType: auth.JWTAuthentication,
			TrustCenterID:      trustCenterID,
		}

		ctx := context.Background()
		return auth.WithAnonymousTrustCenterUser(ctx, anonUser)
	}

	testCases := []struct {
		name        string
		input       testclient.SendTrustCenterNDAInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path - anonymous user sends email for their trust center",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         gofakeit.Email(),
			},
			client: suite.client.api,
			ctx:    createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID),
		},
		{
			name: "happy path - system admin can send email for any trust center",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         gofakeit.Email(),
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "not authorized - anonymous user tries to send for different trust center",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter2.ID, // Different trust center
				Email:         gofakeit.Email(),
			},
			client:      suite.client.api,
			ctx:         createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID), // Anonymous user for trustCenter1
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - regular user cannot send NDA email",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         gofakeit.Email(),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - view only user cannot send NDA email",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         gofakeit.Email(),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "trust center not found",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: "non-existent-id",
				Email:         gofakeit.Email(),
			},
			client:      suite.client.api,
			ctx:         systemAdminUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "invalid email format",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         "invalid-email",
			},
			client: suite.client.api,
			ctx:    createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID),
		},
		{
			name: "empty email",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         "",
			},
			client: suite.client.api,
			ctx:    createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID),
			// Note: Empty email validation might be handled at GraphQL schema level
			expectedErr: "email is required",
		},
	}

	for _, tc := range testCases {
		t.Run("Send "+tc.name, func(t *testing.T) {
			resp, err := tc.client.SendTrustCenterNDAEmail(tc.ctx, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Verify the response indicates success
			assert.Check(t, resp.SendTrustCenterNDAEmail.Success)
		})
	}

	// Clean up trust centers
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}
