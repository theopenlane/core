//go:build test

package graphapi_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/newman/providers/mock"
)

// TestTrustCenterAnonymousSubscribe verifies an anonymous trust center visitor can subscribe, that
// the subscription is scoped to the trust center and its owning org from the JWT, and that a
// mismatched trust center in the input is rejected
func TestTrustCenterAnonymousSubscribe(t *testing.T) {
	tc := th.CreateFreshOrgWithTrustCenter(t)

	subscriberEmail := gofakeit.Email()
	anonCtx, _ := th.CreateAnonymousTrustCenterContextWithEmail(tc.TrustCenter.ID, tc.TrustCenter.OwnerID, subscriberEmail)

	resp, err := suite.Client.API.CreateSubscriber(anonCtx, testclient.CreateSubscriberInput{
		Email:         subscriberEmail,
		TrustCenterID: &tc.TrustCenter.ID,
	})
	assert.NilError(t, err)
	assert.Equal(t, strings.ToLower(subscriberEmail), resp.CreateSubscriber.Subscriber.Email)

	dbCtx := privacy.DecisionContext(th.SetContext(tc.Owner.UserCtx, suite.Client.DB), privacy.Allow)

	sub, err := suite.Client.DB.Subscriber.Get(dbCtx, resp.CreateSubscriber.Subscriber.ID)
	assert.NilError(t, err)
	assert.Equal(t, tc.TrustCenter.ID, lo.FromPtr(sub.TrustCenterID))
	assert.Equal(t, tc.TrustCenter.OwnerID, sub.OwnerID)

	t.Run("rejects mismatched trust center", func(t *testing.T) {
		other := th.CreateFreshOrgWithTrustCenter(t)

		_, err := suite.Client.API.CreateSubscriber(anonCtx, testclient.CreateSubscriberInput{
			Email:         gofakeit.Email(),
			TrustCenterID: &other.TrustCenter.ID,
		})
		assert.Assert(t, err != nil)
	})

	(&th.Cleanup[*generated.SubscriberDeleteOne]{Client: suite.Client.DB.Subscriber, ID: resp.CreateSubscriber.Subscriber.ID}).MustDelete(tc.Owner.UserCtx, t)
}

// TestTrustCenterSubscriberGate verifies the trust center allow_subscribers flag gates subscriber
// creation: when disabled the create is rejected, and when re-enabled it succeeds
func TestTrustCenterSubscriberGate(t *testing.T) {
	tc := th.CreateFreshOrgWithTrustCenter(t)

	dbCtx := privacy.DecisionContext(th.SetContext(tc.Owner.UserCtx, suite.Client.DB), privacy.Allow)

	tcLoaded, err := suite.Client.DB.TrustCenter.Query().
		Where(trustcenter.IDEQ(tc.TrustCenter.ID)).
		WithSetting().
		Only(dbCtx)
	assert.NilError(t, err)

	setting := tcLoaded.Edges.Setting
	assert.Assert(t, setting != nil)

	subscriberEmail := gofakeit.Email()
	anonCtx, _ := th.CreateAnonymousTrustCenterContextWithEmail(tc.TrustCenter.ID, tc.TrustCenter.OwnerID, subscriberEmail)

	t.Run("blocked when disabled", func(t *testing.T) {
		assert.NilError(t, suite.Client.DB.TrustCenterSetting.UpdateOneID(setting.ID).SetAllowSubscribers(false).Exec(dbCtx))

		_, err := suite.Client.API.CreateSubscriber(anonCtx, testclient.CreateSubscriberInput{
			Email:         subscriberEmail,
			TrustCenterID: &tc.TrustCenter.ID,
		})
		assert.Assert(t, err != nil)
	})

	t.Run("allowed when enabled", func(t *testing.T) {
		assert.NilError(t, suite.Client.DB.TrustCenterSetting.UpdateOneID(setting.ID).SetAllowSubscribers(true).Exec(dbCtx))

		resp, err := suite.Client.API.CreateSubscriber(anonCtx, testclient.CreateSubscriberInput{
			Email:         subscriberEmail,
			TrustCenterID: &tc.TrustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, strings.ToLower(subscriberEmail), resp.CreateSubscriber.Subscriber.Email)

		(&th.Cleanup[*generated.SubscriberDeleteOne]{Client: suite.Client.DB.Subscriber, ID: resp.CreateSubscriber.Subscriber.ID}).MustDelete(tc.Owner.UserCtx, t)
	})
}

// TestTrustCenterSubscriberScopedPerTrustCenter verifies the same email can subscribe to different
// trust centers, producing distinct subscriptions scoped to each trust center
func TestTrustCenterSubscriberScopedPerTrustCenter(t *testing.T) {
	tc1 := th.CreateFreshOrgWithTrustCenter(t)
	tc2 := th.CreateFreshOrgWithTrustCenter(t)

	sharedEmail := gofakeit.Email()

	ctx1, _ := th.CreateAnonymousTrustCenterContextWithEmail(tc1.TrustCenter.ID, tc1.TrustCenter.OwnerID, sharedEmail)
	ctx2, _ := th.CreateAnonymousTrustCenterContextWithEmail(tc2.TrustCenter.ID, tc2.TrustCenter.OwnerID, sharedEmail)

	resp1, err := suite.Client.API.CreateSubscriber(ctx1, testclient.CreateSubscriberInput{
		Email:         sharedEmail,
		TrustCenterID: &tc1.TrustCenter.ID,
	})
	assert.NilError(t, err)

	resp2, err := suite.Client.API.CreateSubscriber(ctx2, testclient.CreateSubscriberInput{
		Email:         sharedEmail,
		TrustCenterID: &tc2.TrustCenter.ID,
	})
	assert.NilError(t, err)

	assert.Assert(t, resp1.CreateSubscriber.Subscriber.ID != resp2.CreateSubscriber.Subscriber.ID)

	dbCtx := privacy.DecisionContext(th.SetContext(tc1.Owner.UserCtx, suite.Client.DB), privacy.Allow)

	sub1, err := suite.Client.DB.Subscriber.Get(dbCtx, resp1.CreateSubscriber.Subscriber.ID)
	assert.NilError(t, err)
	assert.Equal(t, tc1.TrustCenter.ID, lo.FromPtr(sub1.TrustCenterID))

	sub2, err := suite.Client.DB.Subscriber.Get(dbCtx, resp2.CreateSubscriber.Subscriber.ID)
	assert.NilError(t, err)
	assert.Equal(t, tc2.TrustCenter.ID, lo.FromPtr(sub2.TrustCenterID))

	// each subscriber is owned by its own org, so clean up with each org's context
	(&th.Cleanup[*generated.SubscriberDeleteOne]{Client: suite.Client.DB.Subscriber, ID: sub1.ID}).MustDelete(tc1.Owner.UserCtx, t)
	(&th.Cleanup[*generated.SubscriberDeleteOne]{Client: suite.Client.DB.Subscriber, ID: sub2.ID}).MustDelete(tc2.Owner.UserCtx, t)
}

// TestTrustCenterCampaignDispatchBranding verifies a trust center update campaign renders one email
// per target branded from the trust center setting, with the per-recipient unsubscribe link resolved
func TestTrustCenterCampaignDispatchBranding(t *testing.T) {
	tc := th.CreateFreshOrgWithTrustCenter(t)

	dbCtx := privacy.DecisionContext(th.SetContext(tc.Owner.UserCtx, suite.Client.DB), privacy.Allow)

	// ensure the trust center has a branded setting linked via the setting edge
	tcLoaded, err := suite.Client.DB.TrustCenter.Query().Where(trustcenter.IDEQ(tc.TrustCenter.ID)).WithSetting().Only(dbCtx)
	assert.NilError(t, err)

	setting := tcLoaded.Edges.Setting
	if setting == nil {
		setting, err = suite.Client.DB.TrustCenterSetting.Create().SetTrustCenterID(tc.TrustCenter.ID).Save(dbCtx)
		assert.NilError(t, err)
		assert.NilError(t, suite.Client.DB.TrustCenter.UpdateOneID(tc.TrustCenter.ID).SetSettingID(setting.ID).Exec(dbCtx))
	}

	assert.NilError(t, suite.Client.DB.TrustCenterSetting.UpdateOneID(setting.ID).
		SetCompanyName("SecureCorp").
		SetPrimaryColor("#0f3d3a").
		SetAccentColor("#3fc2b4").
		SetLogoRemoteURL("https://securecorp.example.com/logo.png").
		Exec(dbCtx))

	// the campaign metadata carries the post data, as the automated triggers supply it
	campaignObj, err := suite.Client.DB.Campaign.Create().
		SetName("June Update").
		SetOwnerID(tc.OrganizationID).
		SetCampaignType(enums.CampaignTypeTrustCenterUpdate).
		SetTrustCenterID(tc.TrustCenter.ID).
		SetMetadata(map[string]any{
			"postTitle":      "June update",
			"postText":       "We updated our subprocessors.",
			"unsubscribeURL": "https://securecorp.example.com/unsubscribe?token={{ .unsubscribeToken }}",
		}).
		SetRecurrenceFrequency(enums.FrequencyNone).
		Save(dbCtx)
	assert.NilError(t, err)

	targetA, err := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetOwnerID(tc.OrganizationID).
		SetEmail("ada@example.com").
		SetFullName("Ada Lovelace").
		SetMetadata(map[string]any{email.MetadataUnsubscribeTokenKey: "tok_ada"}).
		Save(dbCtx)
	assert.NilError(t, err)

	targetGrace, err := suite.Client.DB.CampaignTarget.Create().
		SetCampaignID(campaignObj.ID).
		SetOwnerID(tc.OrganizationID).
		SetEmail("grace@example.com").
		SetFullName("Grace Hopper").
		SetMetadata(map[string]any{email.MetadataUnsubscribeTokenKey: "tok_grace"}).
		Save(dbCtx)
	assert.NilError(t, err)

	defer func() {
		(&th.Cleanup[*generated.CampaignTargetDeleteOne]{Client: suite.Client.DB.CampaignTarget, IDs: []string{targetA.ID, targetGrace.ID}}).MustDelete(tc.Owner.UserCtx, t)
		(&th.Cleanup[*generated.CampaignDeleteOne]{Client: suite.Client.DB.Campaign, ID: campaignObj.ID}).MustDelete(tc.Owner.UserCtx, t)
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
		DB:     suite.Client.DB,
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

	// each subscriber receives a message
	assert.Assert(t, strings.Contains(combinedTo, "ada@example.com"))
	assert.Assert(t, strings.Contains(combinedTo, "grace@example.com"))

	// branding sourced from trust center setting
	assert.Assert(t, strings.Contains(combinedHTML, "SecureCorp"))
	assert.Assert(t, strings.Contains(combinedHTML, "https://securecorp.example.com/logo.png"))

	// per-recipient unsubscribe link
	assert.Assert(t, strings.Contains(combinedHTML, "https://securecorp.example.com/unsubscribe?token=tok_ada"))
	assert.Assert(t, strings.Contains(combinedHTML, "https://securecorp.example.com/unsubscribe?token=tok_grace"))
}

// TestTrustCenterPostNotificationEmail verifies that publishing a trust center post flagged for
// subscriber notification, once stable past the grace window, dispatches a branded update email to the
// trust center's active subscribers rendering the post content, branding, and an unsubscribe link
func TestTrustCenterPostNotificationEmail(t *testing.T) {
	tc := th.CreateFreshOrgWithTrustCenter(t)

	dbCtx := privacy.DecisionContext(th.SetContext(tc.Owner.UserCtx, suite.Client.DB), privacy.Allow)

	// brand the live trust center setting so the email pulls trust center branding
	tcLoaded, err := suite.Client.DB.TrustCenter.Query().Where(trustcenter.IDEQ(tc.TrustCenter.ID)).WithSetting().Only(dbCtx)
	assert.NilError(t, err)

	setting := tcLoaded.Edges.Setting
	assert.Assert(t, setting != nil)

	assert.NilError(t, suite.Client.DB.TrustCenterSetting.UpdateOneID(setting.ID).
		SetCompanyName("SecureCorp").
		SetLogoRemoteURL("https://securecorp.example.com/logo.png").
		Exec(dbCtx))

	// an active, verified subscriber to the trust center
	sub, err := suite.Client.DB.Subscriber.Create().
		SetOwnerID(tc.TrustCenter.OwnerID).
		SetTrustCenterID(tc.TrustCenter.ID).
		SetEmail("ada@example.com").
		SetActive(true).
		SetVerifiedEmail(true).
		Save(dbCtx)
	assert.NilError(t, err)

	// a published post flagged for notification, back-dated so it is stable past the grace window
	stale := time.Now().Add(-2 * time.Hour)
	assert.NilError(t, suite.Client.DB.Note.Create().
		SetOwnerID(tc.TrustCenter.OwnerID).
		SetTrustCenterID(tc.TrustCenter.ID).
		SetTitle("June trust center update").
		SetText("We added a new subprocessor and refreshed our security documentation.").
		SetNotifySubscribers(true).
		SetUpdatedAt(stale).
		Exec(dbCtx))

	// let the subscriber create hook's confirmation email settle, then clear it so only the post
	// notification remains
	suite.WaitForEvents()
	suite.MockEmailSender().Reset()

	_, err = suite.IntegrationsRT.HandleReconcile(context.Background(), operations.ReconcileEnvelope{
		OperationContext: types.NewOperationContext("", email.TrustCenterNotificationOp.Name(), types.IntegrationSource{
			DefinitionID: email.DefinitionID.ID(),
			RunType:      enums.IntegrationRunTypeScheduled,
			Runtime:      true,
		}),
	})
	assert.NilError(t, err)

	suite.WaitForEvents()

	messages := suite.MockEmailSender().Messages()
	assert.Assert(t, len(messages) >= 1)

	var allHTML, allTo []string
	for _, msg := range messages {
		allHTML = append(allHTML, msg.HTML)
		allTo = append(allTo, msg.To...)
	}

	combinedHTML := strings.Join(allHTML, "\n")
	combinedTo := strings.Join(allTo, " ")

	// subscriber receives the post notification
	assert.Assert(t, strings.Contains(combinedTo, "ada@example.com"))

	// post content and trust center branding render
	assert.Assert(t, strings.Contains(combinedHTML, "June trust center update"))
	assert.Assert(t, strings.Contains(combinedHTML, "We added a new subprocessor"))
	assert.Assert(t, strings.Contains(combinedHTML, "SecureCorp"))

	// per-recipient unsubscribe link
	assert.Assert(t, strings.Contains(combinedHTML, "/unsubscribe?token="))
	assert.Assert(t, strings.Contains(combinedHTML, sub.Token))
}

// TestTrustCenterSubprocessorNotificationEmail verifies that a subprocessor change on a trust center
// that opted in, once stable past the grace window, sends the controlled subprocessor system email to
// the trust center's active subscribers, rendering the changed vendor and a per-recipient unsubscribe link
func TestTrustCenterSubprocessorNotificationEmail(t *testing.T) {
	tc := th.CreateFreshOrgWithTrustCenter(t)

	dbCtx := privacy.DecisionContext(th.SetContext(tc.Owner.UserCtx, suite.Client.DB), privacy.Allow)

	tcLoaded, err := suite.Client.DB.TrustCenter.Query().Where(trustcenter.IDEQ(tc.TrustCenter.ID)).WithSetting().Only(dbCtx)
	assert.NilError(t, err)

	setting := tcLoaded.Edges.Setting
	assert.Assert(t, setting != nil)

	// opt the trust center into subprocessor notifications and brand it; the watermark is backdated
	// past the change created below, since enabling the flag stamps it to now otherwise
	assert.NilError(t, suite.Client.DB.TrustCenterSetting.UpdateOneID(setting.ID).
		SetNotifySubscribersOnSubprocessorChange(true).
		SetSubprocessorsNotifiedAt(time.Now().Add(-3*time.Hour)).
		SetCompanyName("SecureCorp").
		SetLogoRemoteURL("https://securecorp.example.com/logo.png").
		Exec(dbCtx))

	// an active, verified subscriber to the trust center
	sub, err := suite.Client.DB.Subscriber.Create().
		SetOwnerID(tc.TrustCenter.OwnerID).
		SetTrustCenterID(tc.TrustCenter.ID).
		SetEmail("ada@example.com").
		SetActive(true).
		SetVerifiedEmail(true).
		Save(dbCtx)
	assert.NilError(t, err)

	vendor, err := suite.Client.DB.Subprocessor.Create().
		SetOwnerID(tc.TrustCenter.OwnerID).
		SetName("Amazon Web Services").
		SetLogoRemoteURL("https://securecorp.example.com/logos/aws.png").
		Save(dbCtx)
	assert.NilError(t, err)

	// create the change already stable past the grace window (set on create, since the audit mixin
	// resets updated_at to now on any update)
	stale := time.Now().Add(-2 * time.Hour)
	assert.NilError(t, suite.Client.DB.TrustCenterSubprocessor.Create().
		SetTrustCenterID(tc.TrustCenter.ID).
		SetSubprocessorID(vendor.ID).
		SetCountries([]string{"US", "DE"}).
		SetUpdatedAt(stale).
		Exec(dbCtx))

	suite.WaitForEvents()
	suite.MockEmailSender().Reset()

	_, err = suite.IntegrationsRT.HandleReconcile(context.Background(), operations.ReconcileEnvelope{
		OperationContext: types.NewOperationContext("", email.TrustCenterNotificationOp.Name(), types.IntegrationSource{
			DefinitionID: email.DefinitionID.ID(),
			RunType:      enums.IntegrationRunTypeScheduled,
			Runtime:      true,
		}),
	})
	assert.NilError(t, err)

	suite.WaitForEvents()

	messages := suite.MockEmailSender().Messages()
	assert.Assert(t, len(messages) >= 1)

	var allHTML, allTo []string
	for _, msg := range messages {
		allHTML = append(allHTML, msg.HTML)
		allTo = append(allTo, msg.To...)
	}

	combinedHTML := strings.Join(allHTML, "\n")
	combinedTo := strings.Join(allTo, " ")

	// subscriber receives the subprocessor notification
	assert.Assert(t, strings.Contains(combinedTo, "ada@example.com"))

	// changed vendor and trust center branding render
	assert.Assert(t, strings.Contains(combinedHTML, "Amazon Web Services"))
	assert.Assert(t, strings.Contains(combinedHTML, "SecureCorp"))

	// per-recipient unsubscribe link
	assert.Assert(t, strings.Contains(combinedHTML, "/unsubscribe?token="))
	assert.Assert(t, strings.Contains(combinedHTML, sub.Token))
}
