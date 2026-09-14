//go:build test

package eventstest_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

func TestOnboardingProgramListener(t *testing.T) {

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.OnboardingProgramListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	user := suite.UserBuilder(context.Background(), t)
	ctx := th.SetContext(user.UserCtx, suite.Client.DB)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	tx, err := suite.Client.DB.Tx(ctx)
	assert.NilError(t, err)
	defer tx.Rollback()
	txCtx := generated.NewContext(generated.NewTxContext(ctx, tx), tx.Client())

	onboarding, err := tx.Client().Onboarding.Create().
		SetInput(generated.CreateOnboardingInput{
			CompanyName: "Program Co one",
			Compliance: map[string]interface{}{
				"frameworks":    []string{"other", "other"},
				"auditor_name":  "New Auditor",
				"auditor_email": "auditor@example.com",
			},
		}).Save(txCtx)
	assert.NilError(t, err)

	ok, err := tx.Client().Program.Query().
		Where(program.OwnerID(onboarding.OrganizationID)).
		Exist(privacy.DecisionContext(txCtx, privacy.Allow))
	assert.NilError(t, err)
	assert.Assert(t, !ok)
	assert.NilError(t, tx.Commit())
	waitForGala(t, setup.Runtime)

	created, err := suite.Client.DB.Program.Query().
		Where(program.OwnerID(onboarding.OrganizationID)).
		Only(allowCtx)
	assert.NilError(t, err)
	assert.Equal(t, created.FrameworkName, "Other")
	assert.Equal(t, created.Auditor, "New Auditor")
	assert.Equal(t, created.AuditorEmail, "auditor@example.com")

}
