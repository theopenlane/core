package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
)

func TestTemplateAnonymousTrustCenterAccess(t *testing.T) {
	// two orgs each with a trust center and NDA template
	tcOrg1 := createFreshOrgWithTrustCenter(t, withNDATemplate())
	tcOrg2 := createFreshOrgWithTrustCenter(t, withNDATemplate())

	trustCenter1 := tcOrg1.trustCenter
	trustCenter2 := tcOrg2.trustCenter

	// a regular questionnaire template in org1 with no trust_center_id
	regularTemplate := (&TemplateBuilder{
		client: suite.client,
		Kind:   enums.TemplateKindQuestionnaire,
	}).MustNew(tcOrg1.owner.UserCtx, t)

	// a questionnaire template linked to tc1 — accessible to tc1 anon users but not cross-org
	questionnaireTemplate := (&TemplateBuilder{
		client:        suite.client,
		Kind:          enums.TemplateKindQuestionnaire,
		TrustCenterID: trustCenter1.ID,
	}).MustNew(tcOrg1.owner.UserCtx, t)

	anonCtxOrg1 := createAnonymousTrustCenterContext(trustCenter1.ID, trustCenter1.OwnerID)
	anonCtxOrg2 := createAnonymousTrustCenterContext(trustCenter2.ID, trustCenter2.OwnerID)

	testCases := []struct {
		name        string
		ctx         context.Context
		queryID     string
		expectedErr string
	}{
		{
			name:    "anon TC user can access NDA template for their trust center",
			ctx:     anonCtxOrg1,
			queryID: *tcOrg1.ndaTemplateID,
		},
		{
			name:        "anon TC user cannot access regular org template (no trust_center_id)",
			ctx:         anonCtxOrg1,
			queryID:     regularTemplate.ID,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "anon TC user from org1 cannot access org2 NDA template",
			ctx:         anonCtxOrg1,
			queryID:     *tcOrg2.ndaTemplateID,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:    "anon TC user can access questionnaire template linked to their trust center",
			ctx:     anonCtxOrg1,
			queryID: questionnaireTemplate.ID,
		},
		{
			name:        "anon TC user from org2 cannot access org1 questionnaire template",
			ctx:         anonCtxOrg2,
			queryID:     questionnaireTemplate.ID,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:    "org owner can access their NDA template",
			ctx:     tcOrg1.owner.UserCtx,
			queryID: *tcOrg1.ndaTemplateID,
		},
		{
			name:    "org owner can access regular org template",
			ctx:     tcOrg1.owner.UserCtx,
			queryID: regularTemplate.ID,
		},
		{
			name:        "org2 owner cannot access org1 NDA template",
			ctx:         tcOrg2.owner.UserCtx,
			queryID:     *tcOrg1.ndaTemplateID,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.client.api.GetTemplateByID(tc.ctx, tc.queryID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tc.queryID, resp.Template.ID)
		})
	}

	t.Run("anon TC user list templates only returns their trust center templates", func(t *testing.T) {
		resp, err := suite.client.api.GetAllTemplates(anonCtxOrg1)
		assert.NilError(t, err)
		// should see both NDA template and questionnaire template linked to tc1
		ids := make([]string, 0, len(resp.Templates.Edges))
		for _, edge := range resp.Templates.Edges {
			ids = append(ids, edge.Node.ID)
		}
		assert.Check(t, lo.Contains(ids, *tcOrg1.ndaTemplateID), "anon TC user should see NDA template")
		assert.Check(t, lo.Contains(ids, questionnaireTemplate.ID), "anon TC user should see tc-linked questionnaire template")
		assert.Check(t, !lo.Contains(ids, regularTemplate.ID), "anon TC user should not see regular template without trust_center_id")
		assert.Check(t, !lo.Contains(ids, *tcOrg2.ndaTemplateID), "anon TC user should not see org2 templates")
	})

	t.Run("anon TC user from org2 cannot see org1 templates in list", func(t *testing.T) {
		resp, err := suite.client.api.GetAllTemplates(anonCtxOrg2)
		assert.NilError(t, err)
		for _, edge := range resp.Templates.Edges {
			assert.Check(t, edge.Node.ID != *tcOrg1.ndaTemplateID, "org2 anon user should not see org1 NDA template")
			assert.Check(t, edge.Node.ID != regularTemplate.ID, "org2 anon user should not see org1 regular template")
			assert.Check(t, edge.Node.ID != questionnaireTemplate.ID, "org2 anon user should not see org1 questionnaire template")
		}
	})

	t.Run("org owner list templates sees all their templates", func(t *testing.T) {
		resp, err := suite.client.api.GetAllTemplates(tcOrg1.owner.UserCtx)
		assert.NilError(t, err)

		ids := make([]string, 0, len(resp.Templates.Edges))
		for _, edge := range resp.Templates.Edges {
			ids = append(ids, edge.Node.ID)
		}

		assert.Check(t, lo.Contains(ids, *tcOrg1.ndaTemplateID), "org owner should see NDA template")
		assert.Check(t, lo.Contains(ids, regularTemplate.ID), "org owner should see regular template")
		assert.Check(t, lo.Contains(ids, questionnaireTemplate.ID), "org owner should see questionnaire template")
		assert.Check(t, !lo.Contains(ids, *tcOrg2.ndaTemplateID), "org owner should not see other org NDA template")
	})

	cleanupOrganizationDataWithContext(tcOrg1.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestSubscriberAnonymousTrustCenterAccess(t *testing.T) {
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	subscriberEmail := gofakeit.Email()
	anonCtx, _ := createAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, subscriberEmail)

	// confirm create path works for anon TC users
	createResp, err := suite.client.api.CreateSubscriber(anonCtx, testclient.CreateSubscriberInput{
		Email:         subscriberEmail,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)
	assert.Equal(t, subscriberEmail, createResp.CreateSubscriber.Subscriber.Email)

	// a regular org subscriber for comparison
	orgSubscriber := (&SubscriberBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)

	t.Run("anon TC user cannot list subscribers", func(t *testing.T) {
		_, err := suite.client.api.GetAllSubscribers(anonCtx, nil, nil, nil, nil, nil)
		assert.Assert(t, err != nil, "anon TC user should be denied from listing subscribers")
	})

	t.Run("anon TC user cannot get subscriber by email", func(t *testing.T) {
		_, err := suite.client.api.GetSubscriberByEmail(anonCtx, subscriberEmail)
		assert.Assert(t, err != nil, "anon TC user should not be able to read subscriber data")
	})

	t.Run("anon TC user cannot get org subscriber by email", func(t *testing.T) {
		_, err := suite.client.api.GetSubscriberByEmail(anonCtx, orgSubscriber.Email)
		assert.Assert(t, err != nil, "anon TC user should not be able to read org subscriber data")
	})

	t.Run("org owner can list subscribers", func(t *testing.T) {
		resp, err := suite.client.api.GetAllSubscribers(tcOrg.owner.UserCtx, nil, nil, nil, nil, nil)
		assert.NilError(t, err)
		assert.Check(t, len(resp.Subscribers.Edges) >= 1)
	})

	// org cleanup handles all subscriber deletion
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestTrustCenterNDARequestAnonymousDataIsolation(t *testing.T) {
	tcOrg1 := createFreshOrgWithTrustCenter(t, withNDATemplate())
	tcOrg2 := createFreshOrgWithTrustCenter(t, withNDATemplate())

	trustCenter1 := tcOrg1.trustCenter
	trustCenter2 := tcOrg2.trustCenter

	anonEmail1 := gofakeit.Email()
	anonEmail2 := gofakeit.Email()
	anonCtxOrg1, _ := createAnonymousTrustCenterContextWithEmail(trustCenter1.ID, trustCenter1.OwnerID, anonEmail1)
	anonCtxOrg2, _ := createAnonymousTrustCenterContextWithEmail(trustCenter2.ID, trustCenter2.OwnerID, anonEmail2)

	req1, err := suite.client.api.CreateTrustCenterNDARequest(anonCtxOrg1, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         anonEmail1,
		TrustCenterID: &trustCenter1.ID,
	})
	assert.NilError(t, err)

	req2, err := suite.client.api.CreateTrustCenterNDARequest(anonCtxOrg2, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         anonEmail2,
		TrustCenterID: &trustCenter2.ID,
	})
	assert.NilError(t, err)

	ndaID1 := req1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID
	ndaID2 := req2.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID

	t.Run("anon TC user from org1 cannot retrieve own NDA request by ID", func(t *testing.T) {
		_, err := suite.client.api.GetTrustCenterNDARequestByID(anonCtxOrg1, ndaID1)
		assert.Assert(t, err != nil, "anon TC user should not be able to read NDA requests")
	})

	t.Run("anon TC user from org1 cannot retrieve org2 NDA request by ID", func(t *testing.T) {
		_, err := suite.client.api.GetTrustCenterNDARequestByID(anonCtxOrg1, ndaID2)
		assert.Assert(t, err != nil, "anon TC user should not be able to read other org NDA requests")
	})

	t.Run("anon TC user cannot list NDA requests", func(t *testing.T) {
		_, err := suite.client.api.GetAllTrustCenterNDARequests(anonCtxOrg1, nil, nil, nil, nil, nil)
		assert.Assert(t, err != nil, "anon TC user should be denied from listing NDA requests")
	})

	t.Run("anon TC user from org2 cannot list NDA requests", func(t *testing.T) {
		_, err := suite.client.api.GetAllTrustCenterNDARequests(anonCtxOrg2, nil, nil, nil, nil, nil)
		assert.Assert(t, err != nil, "anon TC user should be denied from listing NDA requests")
	})

	t.Run("org1 owner can retrieve org1 NDA request", func(t *testing.T) {
		resp, err := suite.client.api.GetTrustCenterNDARequestByID(tcOrg1.owner.UserCtx, ndaID1)
		assert.NilError(t, err)
		assert.Equal(t, ndaID1, resp.TrustCenterNDARequest.ID)
	})

	t.Run("org2 owner cannot retrieve org1 NDA request", func(t *testing.T) {
		_, err := suite.client.api.GetTrustCenterNDARequestByID(tcOrg2.owner.UserCtx, ndaID1)
		assert.ErrorContains(t, err, notFoundErrorMsg)
	})

	t.Run("org1 owner list sees only their NDA requests", func(t *testing.T) {
		resp, err := suite.client.api.GetAllTrustCenterNDARequests(tcOrg1.owner.UserCtx, nil, nil, nil, nil, nil)
		assert.NilError(t, err)

		ids := make([]string, 0, len(resp.TrustCenterNdaRequests.Edges))
		for _, edge := range resp.TrustCenterNdaRequests.Edges {
			ids = append(ids, edge.Node.ID)
		}

		assert.Check(t, lo.Contains(ids, ndaID1), "org1 owner should see org1 NDA request")
		assert.Check(t, !lo.Contains(ids, ndaID2), "org1 owner should not see org2 NDA request")
	})

	cleanupOrganizationDataWithContext(tcOrg1.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestControlAnonymousTrustCenterAccess(t *testing.T) {
	tcOrg1 := createFreshOrgWithTrustCenter(t)
	tcOrg2 := createFreshOrgWithTrustCenter(t)

	org1DBCtx := setContext(tcOrg1.owner.UserCtx, suite.client.db)
	org2DBCtx := setContext(tcOrg2.owner.UserCtx, suite.client.db)

	// public TC control for org1
	tcControl1, err := suite.client.db.Control.Create().
		SetRefCode("OTS-TC-" + ulids.New().String()).
		SetTitle("Public TC Control Org1").
		SetSource(enums.ControlSourceUserDefined).
		SetIsTrustCenterControl(true).
		SetOwnerID(tcOrg1.organizationID).
		Save(org1DBCtx)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateControl(tcOrg1.owner.UserCtx, tcControl1.ID, testclient.UpdateControlInput{
		TrustCenterVisibility: lo.ToPtr(enums.TrustCenterControlVisibilityPubliclyVisible),
	})
	assert.NilError(t, err)

	// private (not publicly visible) TC control for org1 — anon users must not see this
	tcControlPrivate, err := suite.client.db.Control.Create().
		SetRefCode("OTS-TC-" + ulids.New().String()).
		SetTitle("Private TC Control Org1").
		SetSource(enums.ControlSourceUserDefined).
		SetIsTrustCenterControl(true).
		SetOwnerID(tcOrg1.organizationID).
		Save(org1DBCtx)
	assert.NilError(t, err)

	// public TC control for org2 — org1 anon users must not see this
	tcControl2, err := suite.client.db.Control.Create().
		SetRefCode("OTS-TC-" + ulids.New().String()).
		SetTitle("Public TC Control Org2").
		SetSource(enums.ControlSourceUserDefined).
		SetIsTrustCenterControl(true).
		SetOwnerID(tcOrg2.organizationID).
		Save(org2DBCtx)
	assert.NilError(t, err)

	_, err = suite.client.api.UpdateControl(tcOrg2.owner.UserCtx, tcControl2.ID, testclient.UpdateControlInput{
		TrustCenterVisibility: lo.ToPtr(enums.TrustCenterControlVisibilityPubliclyVisible),
	})
	assert.NilError(t, err)

	// regular non-TC compliance control for org1 — must never be visible to anon TC users
	regularControl, err := suite.client.db.Control.Create().
		SetRefCode("REG-" + ulids.New().String()).
		SetTitle("Regular Compliance Control Org1").
		SetSource(enums.ControlSourceUserDefined).
		SetOwnerID(tcOrg1.organizationID).
		Save(org1DBCtx)
	assert.NilError(t, err)

	anonCtxOrg1 := createAnonymousTrustCenterContext(tcOrg1.trustCenter.ID, tcOrg1.organizationID)
	anonCtxOrg2 := createAnonymousTrustCenterContext(tcOrg2.trustCenter.ID, tcOrg2.organizationID)

	// GetTrustCenterControlByID: minimal query (no org-member edges) — anon TC users should succeed
	tcControlTestCases := []struct {
		name        string
		ctx         context.Context
		queryID     string
		expectedErr string
	}{
		{
			name:    "anon TC user (org1) can see publicly visible TC control via TC query",
			ctx:     anonCtxOrg1,
			queryID: tcControl1.ID,
		},
		{
			name:        "anon TC user (org1) cannot see private TC control via TC query",
			ctx:         anonCtxOrg1,
			queryID:     tcControlPrivate.ID,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "anon TC user (org1) cannot see org2 TC control via TC query",
			ctx:         anonCtxOrg1,
			queryID:     tcControl2.ID,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "anon TC user (org1) cannot see non-TC compliance control via TC query",
			ctx:         anonCtxOrg1,
			queryID:     regularControl.ID,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "anon TC user (org2) cannot see org1 TC control via TC query",
			ctx:         anonCtxOrg2,
			queryID:     tcControl1.ID,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:    "anon TC user (org2) can see their own publicly visible TC control via TC query",
			ctx:     anonCtxOrg2,
			queryID: tcControl2.ID,
		},
		{
			name:    "org1 owner can see all their controls",
			ctx:     tcOrg1.owner.UserCtx,
			queryID: regularControl.ID,
		},
	}

	for _, tc := range tcControlTestCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.client.api.GetTrustCenterControlByID(tc.ctx, tc.queryID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tc.queryID, resp.Control.ID)
		})
	}

	// GetControlByID: full query (loads standard, subcontrols edges) — anon TC users should get not found
	// because resolving org-member-only edges fails for anon callers
	t.Run("anon TC user (org1) cannot use full GetControlByID query", func(t *testing.T) {
		_, err := suite.client.api.GetControlByID(anonCtxOrg1, tcControl1.ID)
		assert.ErrorContains(t, err, notFoundErrorMsg)
	})

	t.Run("anon TC user (org1) list controls only includes their org public TC controls", func(t *testing.T) {
		resp, err := suite.client.api.GetTrustCenterControls(anonCtxOrg1)
		assert.NilError(t, err)

		ids := make([]string, 0, len(resp.Controls.Edges))
		for _, edge := range resp.Controls.Edges {
			ids = append(ids, edge.Node.ID)
		}

		assert.Check(t, lo.Contains(ids, tcControl1.ID), "anon TC user should see publicly visible TC control")
		assert.Check(t, !lo.Contains(ids, tcControlPrivate.ID), "anon TC user should not see private TC control")
		assert.Check(t, !lo.Contains(ids, tcControl2.ID), "anon TC user should not see org2 TC controls")
		assert.Check(t, !lo.Contains(ids, regularControl.ID), "anon TC user should not see non-TC compliance controls")
	})

	t.Run("anon TC user (org2) list controls only includes org2 public TC controls", func(t *testing.T) {
		resp, err := suite.client.api.GetTrustCenterControls(anonCtxOrg2)
		assert.NilError(t, err)

		ids := make([]string, 0, len(resp.Controls.Edges))
		for _, edge := range resp.Controls.Edges {
			ids = append(ids, edge.Node.ID)
		}

		assert.Check(t, lo.Contains(ids, tcControl2.ID), "org2 anon TC user should see org2 publicly visible TC control")
		assert.Check(t, !lo.Contains(ids, tcControl1.ID), "org2 anon TC user should not see org1 TC controls")
		assert.Check(t, !lo.Contains(ids, regularControl.ID), "org2 anon TC user should not see org1 non-TC controls")
	})

	cleanupOrganizationDataWithContext(tcOrg1.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}
