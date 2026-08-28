//go:build test

package eventstest_test

import (
	"context"
	"strings"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

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
