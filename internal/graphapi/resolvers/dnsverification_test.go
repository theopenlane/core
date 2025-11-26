package resolvers_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryDNSVerificationByID(t *testing.T) {
	dnsVerification := (&DNSVerificationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name                 string
		expectedCloudflareID string
		queryID              string
		client               *testclient.TestClient
		ctx                  context.Context
		errorMsg             string
	}{
		{
			name:                 "happy path",
			expectedCloudflareID: dnsVerification.CloudflareHostnameID,
			queryID:              dnsVerification.ID,
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
		},
		{
			name:                 "happy path, view only user",
			expectedCloudflareID: dnsVerification.CloudflareHostnameID,
			queryID:              dnsVerification.ID,
			client:               suite.client.api,
			ctx:                  viewOnlyUser.UserCtx,
		},
		{
			name:                 "happy path, sysadmin user",
			expectedCloudflareID: dnsVerification.CloudflareHostnameID,
			queryID:              dnsVerification.ID,
			client:               suite.client.api,
			ctx:                  systemAdminUser.UserCtx,
		},
		{
			name:     "verification not found",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:                 "not authorized to query org",
			expectedCloudflareID: dnsVerification.CloudflareHostnameID,
			queryID:              dnsVerification.ID,
			client:               suite.client.api,
			ctx:                  testUser2.UserCtx,
			errorMsg:             notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetDNSVerificationByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.DNSVerification.ID))
			assert.Check(t, is.Equal(tc.expectedCloudflareID, resp.DNSVerification.CloudflareHostnameID))
		})
	}
	(&Cleanup[*generated.DNSVerificationDeleteOne]{client: suite.client.db.DNSVerification, ID: dnsVerification.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryDNSVerifications(t *testing.T) {
	dnsVerification1 := (&DNSVerificationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	dnsVerification2 := (&DNSVerificationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	dnsVerification3 := (&DNSVerificationBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	nonExistentCloudflareID := "nonexistent-cloudflare-id"

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		where           *testclient.DNSVerificationWhereInput
	}{
		{
			name:            "return all",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "return all, ro user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:   "return all, sysadmin user",
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
			where: &testclient.DNSVerificationWhereInput{
				OwnerID: lo.ToPtr(testUser1.OrganizationID),
			},
			expectedResults: 2,
		},
		{
			name:   "query by cloudflare hostname ID",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.DNSVerificationWhereInput{
				CloudflareHostnameID: &dnsVerification1.CloudflareHostnameID,
			},
			expectedResults: 1,
		},
		{
			name:   "query by cloudflare hostname ID, not found",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.DNSVerificationWhereInput{
				CloudflareHostnameID: &nonExistentCloudflareID,
			},
			expectedResults: 0,
		},
		{
			name:   "query by DNS TXT record",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.DNSVerificationWhereInput{
				DNSTxtRecord: &dnsVerification2.DNSTxtRecord,
			},
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetDNSVerifications(tc.ctx, nil, nil, tc.where)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.expectedResults, resp.DNSVerifications.TotalCount))

			for _, verification := range resp.DNSVerifications.Edges {
				assert.Check(t, is.Equal(*verification.Node.OwnerID, testUser1.OrganizationID))
			}
		})
	}

	(&Cleanup[*generated.DNSVerificationDeleteOne]{client: suite.client.db.DNSVerification, IDs: []string{dnsVerification1.ID, dnsVerification2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.DNSVerificationDeleteOne]{client: suite.client.db.DNSVerification, ID: dnsVerification3.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateDNSVerification(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateDNSVerificationInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path",
			request: testclient.CreateDNSVerificationInput{
				CloudflareHostnameID:       "test-cloudflare-id",
				DNSTxtRecord:               "_openlane-challenge.example.com",
				DNSTxtValue:                "test-dns-value",
				AcmeChallengePath:          lo.ToPtr("acmepaththing"),
				ExpectedAcmeChallengeValue: lo.ToPtr("test-ssl-value"),
				OwnerID:                    lo.ToPtr(testUser1.OrganizationID),
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "not authorized",
			request: testclient.CreateDNSVerificationInput{
				CloudflareHostnameID:       "test-cloudflare-id-unauthorized",
				DNSTxtRecord:               "_openlane-challenge.unauthorized.example.com",
				DNSTxtValue:                "test-dns-value-unauthorized",
				AcmeChallengePath:          lo.ToPtr("acmepaththing"),
				ExpectedAcmeChallengeValue: lo.ToPtr("test-ssl-value"),
				OwnerID:                    lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "missing cloudflare hostname ID",
			request: testclient.CreateDNSVerificationInput{
				DNSTxtRecord:               "_openlane-challenge.missing.example.com",
				DNSTxtValue:                "test-dns-value-missing",
				AcmeChallengePath:          lo.ToPtr("acmepaththing"),
				ExpectedAcmeChallengeValue: lo.ToPtr("test-ssl-value"),
				OwnerID:                    lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         systemAdminUser.UserCtx,
			expectedErr: "cloudflare_hostname_id",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateDNSVerification(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.CloudflareHostnameID, resp.CreateDNSVerification.DNSVerification.CloudflareHostnameID))
			assert.Check(t, is.Equal(tc.request.DNSTxtRecord, resp.CreateDNSVerification.DNSVerification.DNSTxtRecord))
			assert.Check(t, is.Equal(tc.request.DNSTxtValue, resp.CreateDNSVerification.DNSVerification.DNSTxtValue))

			// Clean up
			(&Cleanup[*generated.DNSVerificationDeleteOne]{client: suite.client.db.DNSVerification, ID: resp.CreateDNSVerification.DNSVerification.ID}).MustDelete(tc.ctx, t)
		})
	}
}

func TestMutationDeleteDNSVerification(t *testing.T) {
	dnsVerification := (&DNSVerificationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	dnsVerification2 := (&DNSVerificationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	dnsVerification3 := (&DNSVerificationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	nonExistentID := "non-existent-id"

	testCases := []struct {
		name        string
		id          string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:   "delete verification",
			id:     dnsVerification.ID,
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name:        "unauthorized",
			id:          dnsVerification3.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "verification not found",
			id:          nonExistentID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteDNSVerification(tc.ctx, tc.id)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.id, resp.DeleteDNSVerification.DeletedID))

			// Verify the verification is deleted
			_, err = tc.client.GetDNSVerificationByID(tc.ctx, tc.id)
			assert.ErrorContains(t, err, notFoundErrorMsg)
		})
	}
	(&Cleanup[*generated.DNSVerificationDeleteOne]{client: suite.client.db.DNSVerification, IDs: []string{dnsVerification2.ID, dnsVerification3.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestUpdateDNSVerification(t *testing.T) {
	dnsVerification := (&DNSVerificationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		queryID     string
		client      *testclient.TestClient
		ctx         context.Context
		errorMsg    string
		updateInput testclient.UpdateDNSVerificationInput
	}{
		{
			name:    "happy path",
			queryID: dnsVerification.ID,
			client:  suite.client.api,
			ctx:     systemAdminUser.UserCtx,
			updateInput: testclient.UpdateDNSVerificationInput{
				AcmeChallengeStatus:         lo.ToPtr(enums.SSLVerificationStatusActive),
				DNSVerificationStatus:       lo.ToPtr(enums.DNSVerificationStatusActive),
				AcmeChallengeStatusReason:   lo.ToPtr("all good!"),
				DNSVerificationStatusReason: lo.ToPtr("all good for the domain!"),
				OwnerID:                     lo.ToPtr(testUser1.OrganizationID),
			},
		},
		{
			name:    "not allowed",
			queryID: dnsVerification.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			updateInput: testclient.UpdateDNSVerificationInput{
				Tags: []string{"unauthorized"},
			},
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateDNSVerification(tc.ctx, tc.queryID, tc.updateInput)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
		})
	}
	(&Cleanup[*generated.DNSVerificationDeleteOne]{client: suite.client.db.DNSVerification, ID: dnsVerification.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestGetAllDNSVerifications(t *testing.T) {
	// Create test DNS verifications with different users
	dnsVerification1 := (&DNSVerificationBuilder{
		client: suite.client,
	}).MustNew(testUser1.UserCtx, t)

	dnsVerification2 := (&DNSVerificationBuilder{
		client: suite.client,
	}).MustNew(testUser1.UserCtx, t)

	dnsVerification3 := (&DNSVerificationBuilder{
		client: suite.client,
	}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		expectedErr     string
	}{
		{
			name:            "happy path - regular user sees only their verifications",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 2, // Should see only verifications owned by testUser1
		},
		{
			name:            "happy path - admin user sees all verifications",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: 2, // Should see all owned by testUser
		},
		{
			name:            "happy path - view only user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2, // Should see only verifications from their organization
		},
		{
			name:            "happy path - different user sees only their verifications",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 1, // Should see only verifications owned by testUser2
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllDNSVerifications(tc.ctx)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.DNSVerifications.Edges != nil)

			// Verify the number of results
			assert.Check(t, is.Len(resp.DNSVerifications.Edges, int(tc.expectedResults)))
			assert.Check(t, is.Equal(tc.expectedResults, resp.DNSVerifications.TotalCount))

			// Verify pagination info
			assert.Check(t, resp.DNSVerifications.PageInfo.StartCursor != nil)

			// If we have results, verify the structure of the first result
			if tc.expectedResults > 0 {
				firstNode := resp.DNSVerifications.Edges[0].Node
				assert.Check(t, len(firstNode.ID) != 0)
				assert.Check(t, len(firstNode.CloudflareHostnameID) != 0)
				assert.Check(t, len(firstNode.DNSTxtRecord) != 0)
				assert.Check(t, firstNode.OwnerID != nil)
				assert.Check(t, firstNode.CreatedAt != nil)
			}

			// Verify that users only see verifications from their organization
			if tc.ctx == testUser1.UserCtx || tc.ctx == viewOnlyUser.UserCtx {
				for _, edge := range resp.DNSVerifications.Edges {
					assert.Check(t, is.Equal(testUser1.OrganizationID, *edge.Node.OwnerID))
				}
			} else if tc.ctx == testUser2.UserCtx {
				for _, edge := range resp.DNSVerifications.Edges {
					assert.Check(t, is.Equal(testUser2.OrganizationID, *edge.Node.OwnerID))
				}
			}
		})
	}

	// Clean up created verifications
	(&Cleanup[*generated.DNSVerificationDeleteOne]{client: suite.client.db.DNSVerification, IDs: []string{dnsVerification1.ID, dnsVerification2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.DNSVerificationDeleteOne]{client: suite.client.db.DNSVerification, ID: dnsVerification3.ID}).MustDelete(testUser2.UserCtx, t)
}
