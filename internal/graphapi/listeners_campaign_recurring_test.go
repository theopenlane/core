//go:build test

package graphapi_test

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
)

func TestCampaignRecurringListener(t *testing.T) {
	user := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	ctx := setContext(user.UserCtx, suite.client.db)

	setup, err := graphapi.SetupListenerRuntime(suite.galaRuntime, hooks.CampaignRecurringListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	newCampaign := func(t *testing.T, name string, active bool) *generated.CampaignCreate {
		t.Helper()

		return suite.client.db.Campaign.Create().
			SetName(name).
			SetOwnerID(user.OrganizationID).
			SetStatus(enums.CampaignStatusActive).
			SetIsRecurring(true).
			SetRecurrenceFrequency(enums.FrequencyMonthly).
			SetRecurrenceInterval(1).
			SetIsActive(active)
	}

	nextRunAt := func(t *testing.T, id string) *models.DateTime {
		t.Helper()

		camp, err := suite.client.db.Campaign.Get(ctx, id)
		assert.NilError(t, err)

		return camp.NextRunAt
	}

	t.Run("activation schedules next run", func(t *testing.T) {
		camp, err := newCampaign(t, "campaign listener activation", false).Save(ctx)
		assert.NilError(t, err)
		assert.Check(t, camp.NextRunAt == nil)

		assert.NilError(t, suite.client.db.Campaign.UpdateOneID(camp.ID).SetIsActive(true).Exec(ctx))

		waitForCondition(t, func() bool { return nextRunAt(t, camp.ID) != nil }, "activation should set next_run_at")
		assert.Check(t, time.Time(*nextRunAt(t, camp.ID)).After(time.Now()))
	})

	t.Run("deactivation clears next run", func(t *testing.T) {
		camp, err := newCampaign(t, "campaign listener deactivation", true).
			SetNextRunAt(models.DateTime(time.Now().Add(time.Hour))).
			Save(ctx)
		assert.NilError(t, err)

		assert.NilError(t, suite.client.db.Campaign.UpdateOneID(camp.ID).SetIsActive(false).Exec(ctx))

		waitForCondition(t, func() bool { return nextRunAt(t, camp.ID) == nil }, "deactivation should clear next_run_at")
	})

	t.Run("recurrence shape change recomputes next run", func(t *testing.T) {
		seeded := time.Now().Add(30 * time.Minute)

		camp, err := newCampaign(t, "campaign listener shape change", true).
			SetNextRunAt(models.DateTime(seeded)).
			Save(ctx)
		assert.NilError(t, err)

		assert.NilError(t, suite.client.db.Campaign.UpdateOneID(camp.ID).SetRecurrenceInterval(2).Exec(ctx))

		waitForCondition(t, func() bool {
			next := nextRunAt(t, camp.ID)

			return next != nil && time.Time(*next).Sub(seeded) > time.Hour
		}, "recurrence interval change should recompute next_run_at")
	})

	t.Run("unrelated update leaves next run untouched", func(t *testing.T) {
		seeded := time.Now().Add(time.Hour)

		camp, err := newCampaign(t, "campaign listener unrelated update", true).
			SetNextRunAt(models.DateTime(seeded)).
			Save(ctx)
		assert.NilError(t, err)

		assert.NilError(t, suite.client.db.Campaign.UpdateOneID(camp.ID).SetDescription("no schedule fields touched").Exec(ctx))

		waitForGala(t, setup.Runtime)

		next := nextRunAt(t, camp.ID)
		assert.Assert(t, next != nil)

		drift := time.Time(*next).Sub(seeded)
		if drift < 0 {
			drift = -drift
		}

		assert.Check(t, drift < 2*time.Second)
	})
}
