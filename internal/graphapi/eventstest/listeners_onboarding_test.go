//go:build test

package eventstest_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/notification"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func TestOnboardingProgramListener(t *testing.T) {
	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.OnboardingProgramListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	user := suite.UserBuilder(context.Background(), t)
	ctx := th.SetInternalContext(user.UserCtx, suite.Client.DB)
	internalCtx := auth.WithInternalOperationContext(ctx)

	tx, err := suite.Client.DB.Tx(ctx)
	assert.NilError(t, err)
	defer tx.Rollback()
	txCtx := generated.NewContext(generated.NewTxContext(ctx, tx), tx.Client())

	onboarding, err := tx.Client().Onboarding.Create().
		SetInput(generated.CreateOnboardingInput{
			CompanyName: "Program Co one",
			Compliance: map[string]any{
				"frameworks":    []string{"other", "other"},
				"auditor_name":  "New Auditor",
				"auditor_email": "auditor@example.com",
			},
		}).Save(txCtx)
	assert.NilError(t, err)

	ok, err := tx.Client().Program.Query().
		Where(program.OwnerID(onboarding.OrganizationID)).
		Exist(auth.WithInternalOperationContext(txCtx))
	assert.NilError(t, err)
	assert.Assert(t, !ok)
	assert.NilError(t, tx.Commit())
	waitForGala(t, setup.Runtime)

	created, err := suite.Client.DB.Program.Query().
		Where(program.OwnerID(onboarding.OrganizationID)).
		Only(internalCtx)
	assert.NilError(t, err)
	assert.Equal(t, created.FrameworkName, "Other")
	assert.Equal(t, created.Auditor, "New Auditor")
	assert.Equal(t, created.AuditorEmail, "auditor@example.com")

}

func TestOnboardingDemoDataSeedListener(t *testing.T) {
	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.OnboardingProgramListeners())
	assert.NilError(t, err)
	t.Cleanup(func() { setup.Teardown() })

	testCases := []struct {
		name      string
		seedDemo  *bool
		wantCount int
	}{
		{name: "chose demo data", seedDemo: new(true), wantCount: 1},
		{name: "declined demo data", seedDemo: new(false), wantCount: 0},
		{name: "unanswered", wantCount: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := suite.UserBuilder(context.Background(), t)
			ctx := th.SetInternalContext(user.UserCtx, suite.Client.DB)

			onboarding, err := suite.Client.DB.Onboarding.Create().
				SetInput(generated.CreateOnboardingInput{
					CompanyName:  "Demo Seed Co " + ulids.New().String(),
					SeedDemoData: tc.seedDemo,
					DemoDataset:  new("minimal"),
				}).Save(ctx)
			assert.NilError(t, err)

			waitForGala(t, setup.Runtime)

			count, err := suite.Client.DB.Notification.Query().
				Where(notification.OwnerIDEQ(onboarding.OrganizationID), notification.TitleEQ("Demo data import completed")).
				Count(auth.WithInternalOperationContext(ctx))
			assert.NilError(t, err)
			assert.Check(t, is.Equal(tc.wantCount, count))
		})
	}
}
