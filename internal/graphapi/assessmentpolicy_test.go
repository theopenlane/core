package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

var policySignOffJSONConfig = map[string]any{
	"title": "Policy sign off",
	"pages": []map[string]any{
		{
			"name": "page1",
			"elements": []map[string]any{
				{
					"type": "pdfdocument",
					"name": "question1",
				},
				{
					"type":      "acknowledgement",
					"name":      "question2",
					"enableIf":  "{question1} = true",
					"statement": "I acknowledge that I’ve read and understand the document above.",
				},
			},
		},
	},
}

func createPolicySignOffAssessment(t *testing.T) *testclient.CreateAssessment_CreateAssessment_Assessment {
	t.Helper()

	resp, err := suite.Client.API.CreateAssessment(th.SharedTestUser1.UserCtx, testclient.CreateAssessmentInput{
		Name:       "Policy sign off " + gofakeit.UUID(),
		Jsonconfig: policySignOffJSONConfig,
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	assessment := resp.CreateAssessment.Assessment

	t.Cleanup(func() {
		(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	return &assessment
}

func TestMutationCreateAssessmentPolicy(t *testing.T) {
	assessment1 := createPolicySignOffAssessment(t)
	assessment2 := createPolicySignOffAssessment(t)

	policy1 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	policy2 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: []string{policy1.ID, policy2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	var createdIDs []string

	t.Cleanup(func() {
		(&th.Cleanup[*generated.AssessmentPolicyDeleteOne]{Client: suite.Client.DB.AssessmentPolicy, IDs: createdIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	testCases := []struct {
		name             string
		request          testclient.CreateAssessmentPolicyInput
		client           *testclient.TestClient
		ctx              context.Context
		expectedRevision string
		errorMsg         string
	}{
		{
			name: "happy path, revision defaults to current policy revision",
			request: testclient.CreateAssessmentPolicyInput{
				AssessmentID:     assessment1.ID,
				InternalPolicyID: policy1.ID,
			},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			expectedRevision: policy1.Revision,
		},
		{
			name: "happy path, second policy on the same assessment with explicit revision",
			request: testclient.CreateAssessmentPolicyInput{
				AssessmentID:     assessment1.ID,
				InternalPolicyID: policy2.ID,
				PolicyRevision:   lo.ToPtr("v1.2.3"),
			},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			expectedRevision: "v1.2.3",
		},
		{
			name: "happy path, same policy on a second assessment",
			request: testclient.CreateAssessmentPolicyInput{
				AssessmentID:     assessment2.ID,
				InternalPolicyID: policy1.ID,
			},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			expectedRevision: policy1.Revision,
		},
		{
			name: "duplicate policy on the same assessment",
			request: testclient.CreateAssessmentPolicyInput{
				AssessmentID:     assessment1.ID,
				InternalPolicyID: policy1.ID,
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: "already exists",
		},
		{
			name: "not authorized, view only user",
			request: testclient.CreateAssessmentPolicyInput{
				AssessmentID:     assessment2.ID,
				InternalPolicyID: policy2.ID,
			},
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
		},
		{
			name: "not authorized, different organization",
			request: testclient.CreateAssessmentPolicyInput{
				AssessmentID:     assessment2.ID,
				InternalPolicyID: policy2.ID,
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateAssessmentPolicy(tc.ctx, tc.request)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			created := resp.CreateAssessmentPolicy.AssessmentPolicy
			createdIDs = append(createdIDs, created.ID)

			assert.Check(t, is.Equal(tc.request.AssessmentID, created.AssessmentID))
			assert.Check(t, is.Equal(tc.request.InternalPolicyID, created.InternalPolicyID))
			assert.Check(t, created.PolicyRevision != nil)
			assert.Check(t, is.Equal(tc.expectedRevision, lo.FromPtr(created.PolicyRevision)))
		})
	}

	t.Run("policies linked across assessments", func(t *testing.T) {
		policiesOnAssessment, err := suite.Client.API.GetInternalPolicies(th.SharedTestUser1.UserCtx, nil, nil, nil, nil, nil, &testclient.InternalPolicyWhereInput{
			HasAssessmentsWith: []*testclient.AssessmentWhereInput{{ID: &assessment1.ID}},
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(int64(2), policiesOnAssessment.InternalPolicies.TotalCount))

		assessmentsForPolicy, err := suite.Client.API.GetAssessments(th.SharedTestUser1.UserCtx, nil, nil, &testclient.AssessmentWhereInput{
			HasInternalPoliciesWith: []*testclient.InternalPolicyWhereInput{{ID: &policy1.ID}},
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(int64(2), assessmentsForPolicy.Assessments.TotalCount))
	})
}

func TestMutationCreateBulkAssessmentPolicy(t *testing.T) {
	assessment := createPolicySignOffAssessment(t)

	policy1 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	policy2 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	policy3 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: []string{policy1.ID, policy2.ID, policy3.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	policies := map[string]string{
		policy1.ID: policy1.Revision,
		policy2.ID: policy2.Revision,
		policy3.ID: policy3.Revision,
	}

	resp, err := suite.Client.API.CreateBulkAssessmentPolicy(th.SharedTestUser1.UserCtx, []*testclient.CreateAssessmentPolicyInput{
		{AssessmentID: assessment.ID, InternalPolicyID: policy1.ID},
		{AssessmentID: assessment.ID, InternalPolicyID: policy2.ID},
		{AssessmentID: assessment.ID, InternalPolicyID: policy3.ID},
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	created := resp.CreateBulkAssessmentPolicy.AssessmentPolicies
	assert.Check(t, is.Len(created, len(policies)))

	var createdIDs []string

	for _, ap := range created {
		createdIDs = append(createdIDs, ap.ID)

		assert.Check(t, is.Equal(assessment.ID, ap.AssessmentID))
		assert.Check(t, is.Equal(policies[ap.InternalPolicyID], lo.FromPtr(ap.PolicyRevision)))
	}

	t.Cleanup(func() {
		(&th.Cleanup[*generated.AssessmentPolicyDeleteOne]{Client: suite.Client.DB.AssessmentPolicy, IDs: createdIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})
}

func TestMutationCreateAssessmentWithPolicies(t *testing.T) {
	policy1 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	policy2 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	otherOrgPolicy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: []string{policy1.ID, policy2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: otherOrgPolicy.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	})

	testCases := []struct {
		name              string
		policies          []*testclient.AssessmentPoliciesInput
		client            *testclient.TestClient
		ctx               context.Context
		expectedRevisions map[string]string
		errorMsg          string
	}{
		{
			name:              "happy path, no policies",
			client:            suite.Client.API,
			ctx:               th.SharedTestUser1.UserCtx,
			expectedRevisions: map[string]string{},
		},
		{
			name: "happy path, multiple policies with default and explicit revisions",
			policies: []*testclient.AssessmentPoliciesInput{
				{InternalPolicyID: policy1.ID},
				{InternalPolicyID: policy2.ID, PolicyRevision: lo.ToPtr("v2.0.0")},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
			expectedRevisions: map[string]string{
				policy1.ID: policy1.Revision,
				policy2.ID: "v2.0.0",
			},
		},
		{
			name: "happy path, same policy on another assessment",
			policies: []*testclient.AssessmentPoliciesInput{
				{InternalPolicyID: policy1.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
			expectedRevisions: map[string]string{
				policy1.ID: policy1.Revision,
			},
		},
		{
			name: "policy from another organization, assessment is not created",
			policies: []*testclient.AssessmentPoliciesInput{
				{InternalPolicyID: policy1.ID},
				{InternalPolicyID: otherOrgPolicy.ID},
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
		},
		{
			name: "not authorized, view only user",
			policies: []*testclient.AssessmentPoliciesInput{
				{InternalPolicyID: policy1.ID},
			},
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			name := "Policy sign off " + gofakeit.UUID()

			resp, err := tc.client.CreateAssessmentWithPolicies(tc.ctx, testclient.CreateAssessmentInput{
				Name:       name,
				Jsonconfig: policySignOffJSONConfig,
			}, tc.policies)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			created := resp.CreateAssessmentWithPolicies.Assessment

			var attestationIDs []string

			for _, edge := range created.PolicyAttestations.Edges {
				attestationIDs = append(attestationIDs, edge.Node.ID)
			}

			t.Cleanup(func() {
				(&th.Cleanup[*generated.AssessmentPolicyDeleteOne]{Client: suite.Client.DB.AssessmentPolicy, IDs: attestationIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
				(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: created.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
			})

			assert.Check(t, is.Equal(name, created.Name))
			assert.Check(t, is.Len(created.InternalPolicies.Edges, len(tc.expectedRevisions)))
			assert.Check(t, is.Len(created.PolicyAttestations.Edges, len(tc.expectedRevisions)))

			for _, edge := range created.PolicyAttestations.Edges {
				expected, ok := tc.expectedRevisions[edge.Node.InternalPolicyID]
				assert.Check(t, ok)
				assert.Check(t, is.Equal(expected, lo.FromPtr(edge.Node.PolicyRevision)))
			}
		})
	}
}
