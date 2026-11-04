//go:build test

package hooks_test

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/campaign"
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/iam/auth"
)

func TestCampaignRecurringListenerSchedulesNextRun(t *testing.T) {
	s := suite.setupTest(t)

	userCtx, internalCtx, orgID := s.setupCampaignOrg(t)

	camp, err := s.newRecurringCampaign(orgID, "recurring schedule listener").
		SetIsActive(true).
		Save(internalCtx)
	assert.NilError(t, err)
	assert.Check(t, camp.NextRunAt == nil)

	s.emitCampaignMutation(t, userCtx, camp.ID, campaign.FieldIsActive)

	updated, err := s.client.Campaign.Get(internalCtx, camp.ID)
	assert.NilError(t, err)
	assert.Assert(t, updated.NextRunAt != nil)
	assert.Check(t, time.Time(*updated.NextRunAt).After(time.Now()))
}

func TestCampaignRecurringListenerClearsNextRunOnDeactivation(t *testing.T) {
	s := suite.setupTest(t)

	userCtx, internalCtx, orgID := s.setupCampaignOrg(t)

	camp, err := s.newRecurringCampaign(orgID, "deactivated schedule listener").
		SetIsActive(false).
		SetNextRunAt(models.DateTime(time.Now().Add(time.Hour))).
		Save(internalCtx)
	assert.NilError(t, err)

	s.emitCampaignMutation(t, userCtx, camp.ID, campaign.FieldIsActive)

	updated, err := s.client.Campaign.Get(internalCtx, camp.ID)
	assert.NilError(t, err)
	assert.Check(t, updated.NextRunAt == nil)
}

func TestCampaignRecurringListenerSkipsDeletedCampaign(t *testing.T) {
	s := suite.setupTest(t)

	userCtx, internalCtx, orgID := s.setupCampaignOrg(t)

	camp, err := s.newRecurringCampaign(orgID, "deleted schedule listener").
		SetIsActive(true).
		Save(internalCtx)
	assert.NilError(t, err)

	err = s.client.Campaign.DeleteOneID(camp.ID).Exec(internalCtx)
	assert.NilError(t, err)

	s.emitCampaignMutation(t, userCtx, camp.ID, campaign.FieldIsActive)
}

func (s *HookTestSuite) setupCampaignOrg(t *testing.T) (userCtx, internalCtx context.Context, orgID string) {
	user := s.seedUser()
	orgID = user.Edges.OrgMemberships[0].OrganizationID

	err := entitlements.CreateFeatureTuples(context.Background(), &s.client.Authz, orgID,
		[]models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})
	assert.NilError(t, err)

	userCtx = generated.NewContext(auth.NewTestContextWithOrgID(user.ID, orgID), s.client)

	return userCtx, auth.WithInternalOperationContext(userCtx), orgID
}

func (s *HookTestSuite) newRecurringCampaign(orgID, name string) *generated.CampaignCreate {
	return s.client.Campaign.Create().
		SetName(name).
		SetOwnerID(orgID).
		SetIsRecurring(true).
		SetStatus(enums.CampaignStatusActive).
		SetRecurrenceFrequency(enums.FrequencyMonthly).
		SetRecurrenceInterval(1)
}

// emitCampaignMutation emits a campaign mutation event through the public gala surface
// and waits for the in-memory pool to drain so DB side effects are observable
func (s *HookTestSuite) emitCampaignMutation(_ *testing.T, ctx context.Context, campaignID string, changedFields ...string) {
	entityops.EmitMutation(ctx, []*gala.Gala{s.galaRuntime}, entityops.MutationPayload{
		MutationType: generated.TypeCampaign,
		Operation:    entityops.OpUpdateOne,
		EntityID:     campaignID,
		ChangeSet: entityops.ChangeSet{
			ChangedFields: changedFields,
		},
	})

	s.waitForEvents()
}
