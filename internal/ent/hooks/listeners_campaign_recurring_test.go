//go:build test

package hooks_test

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/iam/auth"
)

func (suite *HookTestSuite) TestCampaignRecurringListenerSchedulesNextRun() {
	t := suite.T()

	userCtx, allowCtx, orgID := suite.setupCampaignOrg(t)

	camp, err := suite.newRecurringCampaign(orgID, "recurring schedule listener").
		SetIsActive(true).
		Save(allowCtx)
	assert.NilError(t, err)
	assert.Check(t, camp.NextRunAt == nil)

	suite.emitCampaignMutation(t, userCtx, camp.ID, campaign.FieldIsActive)

	updated, err := suite.client.Campaign.Get(allowCtx, camp.ID)
	assert.NilError(t, err)
	assert.Assert(t, updated.NextRunAt != nil)
	assert.Check(t, time.Time(*updated.NextRunAt).After(time.Now()))
}

func (suite *HookTestSuite) TestCampaignRecurringListenerClearsNextRunOnDeactivation() {
	t := suite.T()

	userCtx, allowCtx, orgID := suite.setupCampaignOrg(t)

	camp, err := suite.newRecurringCampaign(orgID, "deactivated schedule listener").
		SetIsActive(false).
		SetNextRunAt(models.DateTime(time.Now().Add(time.Hour))).
		Save(allowCtx)
	assert.NilError(t, err)

	suite.emitCampaignMutation(t, userCtx, camp.ID, campaign.FieldIsActive)

	updated, err := suite.client.Campaign.Get(allowCtx, camp.ID)
	assert.NilError(t, err)
	assert.Check(t, updated.NextRunAt == nil)
}

func (suite *HookTestSuite) TestCampaignRecurringListenerSkipsDeletedCampaign() {
	t := suite.T()

	userCtx, allowCtx, orgID := suite.setupCampaignOrg(t)

	camp, err := suite.newRecurringCampaign(orgID, "deleted schedule listener").
		SetIsActive(true).
		Save(allowCtx)
	assert.NilError(t, err)

	err = suite.client.Campaign.DeleteOneID(camp.ID).Exec(allowCtx)
	assert.NilError(t, err)

	suite.emitCampaignMutation(t, userCtx, camp.ID, campaign.FieldIsActive)
}

func (suite *HookTestSuite) setupCampaignOrg(t *testing.T) (userCtx, allowCtx context.Context, orgID string) {
	user := suite.seedUser()
	orgID = user.Edges.OrgMemberships[0].OrganizationID

	err := entitlements.CreateFeatureTuples(context.Background(), &suite.client.Authz, orgID,
		[]models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})
	assert.NilError(t, err)

	userCtx = generated.NewContext(auth.NewTestContextWithOrgID(user.ID, orgID), suite.client)

	return userCtx, privacy.DecisionContext(userCtx, privacy.Allow), orgID
}

func (suite *HookTestSuite) newRecurringCampaign(orgID, name string) *generated.CampaignCreate {
	return suite.client.Campaign.Create().
		SetName(name).
		SetOwnerID(orgID).
		SetIsRecurring(true).
		SetStatus(enums.CampaignStatusActive).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetRecurrenceInterval(1)
}

// emitCampaignMutation emits a campaign mutation event through the public gala surface
// and waits for the in-memory pool to drain so DB side effects are observable
func (suite *HookTestSuite) emitCampaignMutation(t *testing.T, ctx context.Context, campaignID string, changedFields ...string) {
	entityops.EmitMutation(ctx, []*gala.Gala{suite.galaRuntime}, entityops.MutationPayload{
		MutationType: generated.TypeCampaign,
		Operation:    entityops.OpUpdateOne,
		EntityID:     campaignID,
		ChangeSet: entityops.ChangeSet{
			ChangedFields: changedFields,
		},
	})

	suite.galaRuntime.WaitIdle()
}
