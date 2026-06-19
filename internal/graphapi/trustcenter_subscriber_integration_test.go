//go:build test

package graphapi_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/newman/providers/mock"
)

// TestTrustCenterAnonymousSubscribe verifies an anonymous trust center visitor can subscribe, that
// the subscription is scoped to the trust center and its owning org from the JWT, and that a
// mismatched trust center in the input is rejected
func TestTrustCenterAnonymousSubscribe(t *testing.T) {
	tc := createFreshOrgWithTrustCenter(t)

	subscriberEmail := gofakeit.Email()
	anonCtx, _ := createAnonymousTrustCenterContextWithEmail(tc.trustCenter.ID, tc.trustCenter.OwnerID, subscriberEmail)

	resp, err := suite.client.api.CreateSubscriber(anonCtx, testclient.CreateSubscriberInput{
		Email:         subscriberEmail,
		TrustCenterID: &tc.trustCenter.ID,
	})
	assert.NilError(t, err)
	assert.Equal(t, strings.ToLower(subscriberEmail), resp.CreateSubscriber.Subscriber.Email)

	dbCtx := privacy.DecisionContext(setContext(tc.owner.UserCtx, suite.client.db), privacy.Allow)

	sub, err := suite.client.db.Subscriber.Get(dbCtx, resp.CreateSubscriber.Subscriber.ID)
	assert.NilError(t, err)
	assert.Equal(t, tc.trustCenter.ID, lo.FromPtr(sub.TrustCenterID))
	assert.Equal(t, tc.trustCenter.OwnerID, sub.OwnerID)

	t.Run("rejects mismatched trust center", func(t *testing.T) {
		other := createFreshOrgWithTrustCenter(t)

		_, err := suite.client.api.CreateSubscriber(anonCtx, testclient.CreateSubscriberInput{
			Email:         gofakeit.Email(),
			TrustCenterID: &other.trustCenter.ID,
		})
		assert.Assert(t, err != nil)
	})

	(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: resp.CreateSubscriber.Subscriber.ID}).MustDelete(tc.owner.UserCtx, t)
}

// TestTrustCenterSubscriberScopedPerTrustCenter verifies the same email can subscribe to different
// trust centers, producing distinct subscriptions scoped to each trust center
func TestTrustCenterSubscriberScopedPerTrustCenter(t *testing.T) {
	tc1 := createFreshOrgWithTrustCenter(t)
	tc2 := createFreshOrgWithTrustCenter(t)

	sharedEmail := gofakeit.Email()

	ctx1, _ := createAnonymousTrustCenterContextWithEmail(tc1.trustCenter.ID, tc1.trustCenter.OwnerID, sharedEmail)
	ctx2, _ := createAnonymousTrustCenterContextWithEmail(tc2.trustCenter.ID, tc2.trustCenter.OwnerID, sharedEmail)

	resp1, err := suite.client.api.CreateSubscriber(ctx1, testclient.CreateSubscriberInput{
		Email:         sharedEmail,
		TrustCenterID: &tc1.trustCenter.ID,
	})
	assert.NilError(t, err)

	resp2, err := suite.client.api.CreateSubscriber(ctx2, testclient.CreateSubscriberInput{
		Email:         sharedEmail,
		TrustCenterID: &tc2.trustCenter.ID,
	})
	assert.NilError(t, err)

	assert.Assert(t, resp1.CreateSubscriber.Subscriber.ID != resp2.CreateSubscriber.Subscriber.ID)

	dbCtx := privacy.DecisionContext(setContext(tc1.owner.UserCtx, suite.client.db), privacy.Allow)

	sub1, err := suite.client.db.Subscriber.Get(dbCtx, resp1.CreateSubscriber.Subscriber.ID)
	assert.NilError(t, err)
	assert.Equal(t, tc1.trustCenter.ID, lo.FromPtr(sub1.TrustCenterID))

	sub2, err := suite.client.db.Subscriber.Get(dbCtx, resp2.CreateSubscriber.Subscriber.ID)
	assert.NilError(t, err)
	assert.Equal(t, tc2.trustCenter.ID, lo.FromPtr(sub2.TrustCenterID))

	// each subscriber is owned by its own org, so clean up with each org's context
	(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: sub1.ID}).MustDelete(tc1.owner.UserCtx, t)
	(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: sub2.ID}).MustDelete(tc2.owner.UserCtx, t)
}

// TestTrustCenterCampaignDispatchBranding verifies a trust center update campaign renders one email
// per target branded from the trust center setting, with the per-recipient unsubscribe link resolved
func TestTrustCenterCampaignDispatchBranding(t *testing.T) {
	tc := createFreshOrgWithTrustCenter(t)

	dbCtx := privacy.DecisionContext(setContext(tc.owner.UserCtx, suite.client.db), privacy.Allow)

	// ensure the trust center has a branded setting linked via the setting edge
	tcLoaded := suite.client.db.TrustCenter.Query().Where(trustcenter.IDEQ(tc.trustCenter.ID)).WithSetting().OnlyX(dbCtx)

	setting := tcLoaded.Edges.Setting
	if setting == nil {
		setting = suite.client.db.TrustCenterSetting.Create().SetTrustCenterID(tc.trustCenter.ID).SaveX(dbCtx)
		suite.client.db.TrustCenter.UpdateOneID(tc.trustCenter.ID).SetSettingID(setting.ID).SaveX(dbCtx)
	}

	suite.client.db.TrustCenterSetting.UpdateOneID(setting.ID).
		SetCompanyName("SecureCorp").
		SetPrimaryColor("#0f3d3a").
		SetAccentColor("#3fc2b4").
		SetLogoRemoteURL("https://securecorp.example.com/logo.png").
		SaveX(dbCtx)

	emailTemplate := suite.client.db.EmailTemplate.Create().
		SetName("Trust Center Update Template").
		SetKey(email.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		SetTrustCenterID(tc.trustCenter.ID).
		SetDefaults(map[string]any{
			"subject":        "{{ .companyName }} trust center update",
			"title":          "Hi {{ .firstName }}",
			"intros":         []any{"We updated our subprocessors."},
			"unsubscribeURL": "https://securecorp.example.com/unsubscribe?token={{ .unsubscribeToken }}",
		}).
		SaveX(dbCtx)

	campaignObj := suite.client.db.Campaign.Create().
		SetName("June Subprocessor Update").
		SetOwnerID(tc.organizationID).
		SetCampaignType(enums.CampaignTypeTrustCenterUpdate).
		SetTrustCenterID(tc.trustCenter.ID).
		SetEmailTemplateID(emailTemplate.ID).
		SetRecurrenceFrequency(enums.FrequencyNone).
		SaveX(dbCtx)

	targetA := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetOwnerID(tc.organizationID).
		SetEmail("ada@example.com").
		SetFullName("Ada Lovelace").
		SetMetadata(map[string]any{email.MetadataUnsubscribeTokenKey: "tok_ada"}).
		SaveX(dbCtx)

	targetGrace := suite.client.db.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetOwnerID(tc.organizationID).
		SetEmail("grace@example.com").
		SetFullName("Grace Hopper").
		SetMetadata(map[string]any{email.MetadataUnsubscribeTokenKey: "tok_grace"}).
		SaveX(dbCtx)

	defer func() {
		(&Cleanup[*generated.CampaignTargetDeleteOne]{client: suite.client.db.CampaignTarget, IDs: []string{targetA.ID, targetGrace.ID}}).MustDelete(tc.owner.UserCtx, t)
		(&Cleanup[*generated.CampaignDeleteOne]{client: suite.client.db.Campaign, ID: campaignObj.ID}).MustDelete(tc.owner.UserCtx, t)
		(&Cleanup[*generated.EmailTemplateDeleteOne]{client: suite.client.db.EmailTemplate, ID: emailTemplate.ID}).MustDelete(tc.owner.UserCtx, t)
	}()

	mockSender, err := mock.New("")
	assert.NilError(t, err)

	emailClient := &email.Client{
		Sender: mockSender,
		Config: *email.MockRuntimeConfig(),
	}

	cfg := email.SendBrandedCampaignRequest{CampaignDispatchInput: email.CampaignDispatchInput{CampaignID: campaignObj.ID}}

	configBytes, err := json.Marshal(cfg)
	assert.NilError(t, err)

	req := types.OperationRequest{
		Client: emailClient,
		DB:     suite.client.db,
		Config: configBytes,
	}

	_, err = email.SendBrandedCampaign{}.Run(dbCtx, req, emailClient, cfg)
	assert.NilError(t, err)

	messages := mockSender.Messages()
	assert.Assert(t, is.Len(messages, 2))

	var allHTML, allTo []string
	for _, msg := range messages {
		allHTML = append(allHTML, msg.HTML)
		allTo = append(allTo, msg.To...)
	}

	combinedHTML := strings.Join(allHTML, "\n")
	combinedTo := strings.Join(allTo, " ")

	t.Run("each subscriber receives a message", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combinedTo, "ada@example.com"))
		assert.Assert(t, strings.Contains(combinedTo, "grace@example.com"))
	})

	t.Run("branding sourced from trust center setting", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combinedHTML, "SecureCorp"))
		assert.Assert(t, strings.Contains(combinedHTML, "https://securecorp.example.com/logo.png"))
	})

	t.Run("per-recipient unsubscribe link", func(t *testing.T) {
		assert.Assert(t, strings.Contains(combinedHTML, "https://securecorp.example.com/unsubscribe?token=tok_ada"))
		assert.Assert(t, strings.Contains(combinedHTML, "https://securecorp.example.com/unsubscribe?token=tok_grace"))
	})
}
