package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryDNSVerificationByID(t *testing.T) {
	dnsVerification := (&th.DNSVerificationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
		},
		{
			name:                 "happy path, view only user",
			expectedCloudflareID: dnsVerification.CloudflareHostnameID,
			queryID:              dnsVerification.ID,
			client:               suite.Client.API,
			ctx:                  th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:                 "happy path, sysadmin user",
			expectedCloudflareID: dnsVerification.CloudflareHostnameID,
			queryID:              dnsVerification.ID,
			client:               suite.Client.API,
			ctx:                  th.SharedSystemAdminUser.UserCtx,
		},
		{
			name:     "verification not found",
			queryID:  "non-existent-id",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:                 "not authorized to query org",
			expectedCloudflareID: dnsVerification.CloudflareHostnameID,
			queryID:              dnsVerification.ID,
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser2.UserCtx,
			errorMsg:             th.NotFoundErrorMsg,
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
	(&th.Cleanup[*generated.DNSVerificationDeleteOne]{Client: suite.Client.DB.DNSVerification, ID: dnsVerification.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryDNSVerifications(t *testing.T) {
	dnsVerification1 := (&th.DNSVerificationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	dnsVerification2 := (&th.DNSVerificationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	dnsVerification3 := (&th.DNSVerificationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

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
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "return all, ro user",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:   "return all, sysadmin user",
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
			where: &testclient.DNSVerificationWhereInput{
				OwnerID: lo.ToPtr(th.SharedTestUser1.OrganizationID),
			},
			expectedResults: 2,
		},
		{
			name:   "query by cloudflare hostname ID",
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
			where: &testclient.DNSVerificationWhereInput{
				CloudflareHostnameID: &dnsVerification1.CloudflareHostnameID,
			},
			expectedResults: 1,
		},
		{
			name:   "query by cloudflare hostname ID, not found",
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
			where: &testclient.DNSVerificationWhereInput{
				CloudflareHostnameID: &nonExistentCloudflareID,
			},
			expectedResults: 0,
		},
		{
			name:   "query by DNS TXT record",
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
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
				assert.Check(t, is.Equal(*verification.Node.OwnerID, th.SharedTestUser1.OrganizationID))
			}
		})
	}

	(&th.Cleanup[*generated.DNSVerificationDeleteOne]{Client: suite.Client.DB.DNSVerification, IDs: []string{dnsVerification1.ID, dnsVerification2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.DNSVerificationDeleteOne]{Client: suite.Client.DB.DNSVerification, ID: dnsVerification3.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
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
				OwnerID:                    lo.ToPtr(th.SharedTestUser1.OrganizationID),
			},
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
		},
		{
			name: "not authorized",
			request: testclient.CreateDNSVerificationInput{
				CloudflareHostnameID:       "test-cloudflare-id-unauthorized",
				DNSTxtRecord:               "_openlane-challenge.unauthorized.example.com",
				DNSTxtValue:                "test-dns-value-unauthorized",
				AcmeChallengePath:          lo.ToPtr("acmepaththing"),
				ExpectedAcmeChallengeValue: lo.ToPtr("test-ssl-value"),
				OwnerID:                    lo.ToPtr(th.SharedTestUser1.OrganizationID),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "missing cloudflare hostname ID",
			request: testclient.CreateDNSVerificationInput{
				DNSTxtRecord:               "_openlane-challenge.missing.example.com",
				DNSTxtValue:                "test-dns-value-missing",
				AcmeChallengePath:          lo.ToPtr("acmepaththing"),
				ExpectedAcmeChallengeValue: lo.ToPtr("test-ssl-value"),
				OwnerID:                    lo.ToPtr(th.SharedTestUser1.OrganizationID),
			},
			client:      suite.Client.API,
			ctx:         th.SharedSystemAdminUser.UserCtx,
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
			(&th.Cleanup[*generated.DNSVerificationDeleteOne]{Client: suite.Client.DB.DNSVerification, ID: resp.CreateDNSVerification.DNSVerification.ID}).MustDelete(tc.ctx, t)
		})
	}
}

func TestMutationDeleteDNSVerification(t *testing.T) {
	dnsVerification := (&th.DNSVerificationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	dnsVerification2 := (&th.DNSVerificationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	dnsVerification3 := (&th.DNSVerificationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
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
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
		},
		{
			name:        "unauthorized",
			id:          dnsVerification3.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "verification not found",
			id:          nonExistentID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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
			assert.ErrorContains(t, err, th.NotFoundErrorMsg)
		})
	}
	(&th.Cleanup[*generated.DNSVerificationDeleteOne]{Client: suite.Client.DB.DNSVerification, IDs: []string{dnsVerification2.ID, dnsVerification3.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestUpdateDNSVerification(t *testing.T) {
	dnsVerification := (&th.DNSVerificationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client:  suite.Client.API,
			ctx:     th.SharedSystemAdminUser.UserCtx,
			updateInput: testclient.UpdateDNSVerificationInput{
				AcmeChallengeStatus:         lo.ToPtr(enums.SSLVerificationStatusActive),
				DNSVerificationStatus:       lo.ToPtr(enums.DNSVerificationStatusActive),
				AcmeChallengeStatusReason:   lo.ToPtr("all good!"),
				DNSVerificationStatusReason: lo.ToPtr("all good for the domain!"),
				OwnerID:                     lo.ToPtr(th.SharedTestUser1.OrganizationID),
			},
		},
		{
			name:    "not allowed",
			queryID: dnsVerification.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			updateInput: testclient.UpdateDNSVerificationInput{
				Tags: []string{"unauthorized"},
			},
			errorMsg: th.NotFoundErrorMsg,
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
	(&th.Cleanup[*generated.DNSVerificationDeleteOne]{Client: suite.Client.DB.DNSVerification, ID: dnsVerification.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestGetAllDNSVerifications(t *testing.T) {
	// Create test DNS verifications with different users
	dnsVerification1 := (&th.DNSVerificationBuilder{
		Client: suite.Client,
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	dnsVerification2 := (&th.DNSVerificationBuilder{
		Client: suite.Client,
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	dnsVerification3 := (&th.DNSVerificationBuilder{
		Client: suite.Client,
	}).MustNew(th.SharedTestUser2.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		expectedErr     string
	}{
		{
			name:            "happy path - regular user sees only their verifications",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 2, // Should see only verifications owned by th.SharedTestUser1
		},
		{
			name:            "happy path - admin user sees all verifications",
			client:          suite.Client.API,
			ctx:             th.SharedAdminUser.UserCtx,
			expectedResults: 2, // Should see all owned by testUser
		},
		{
			name:            "happy path - view only user",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 2, // Should see only verifications from their organization
		},
		{
			name:            "happy path - different user sees only their verifications",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
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
			if tc.ctx == th.SharedTestUser1.UserCtx || tc.ctx == th.SharedViewOnlyUser.UserCtx {
				for _, edge := range resp.DNSVerifications.Edges {
					assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, *edge.Node.OwnerID))
				}
			} else if tc.ctx == th.SharedTestUser2.UserCtx {
				for _, edge := range resp.DNSVerifications.Edges {
					assert.Check(t, is.Equal(th.SharedTestUser2.OrganizationID, *edge.Node.OwnerID))
				}
			}
		})
	}

	// Clean up created verifications
	(&th.Cleanup[*generated.DNSVerificationDeleteOne]{Client: suite.Client.DB.DNSVerification, IDs: []string{dnsVerification1.ID, dnsVerification2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.DNSVerificationDeleteOne]{Client: suite.Client.DB.DNSVerification, ID: dnsVerification3.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
}
