//go:build test

package graphapi_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	cloudflaredef "github.com/theopenlane/core/v2/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/core/v2/pkg/gala"
)

func TestDomainScanListeners(t *testing.T) {
	user := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	ctx := setContext(user.UserCtx, suite.client.db)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// workers on the dispatch runtime are never started, so dispatched cloudflare runs
	// stay queued for counting instead of executing against the fake credentials
	dispatchGala, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode:  gala.DispatchModeDurable,
		ConnectionURI: suite.tf.URI,
		// short base name: gala derives per-kind queues as <name>_<kind> and river caps
		// queue names at 64 chars
		QueueName:         fmt.Sprintf("sdispatch_%d", time.Now().UnixNano()),
		WorkerCount:       1,
		RunMigrations:     true,
		FetchCooldown:     time.Millisecond,
		FetchPollInterval: 10 * time.Millisecond,
	})
	assert.NilError(t, err)

	defer func() { _ = dispatchGala.Close() }()

	credStore, err := keystore.NewStore(suite.client.db)
	assert.NilError(t, err)

	rt, err := intruntime.New(intruntime.Config{
		DB:          suite.client.db,
		Gala:        dispatchGala,
		Keystore:    credStore,
		RedisClient: coreutils.NewRedisClient(),
		DefinitionBuilders: []registry.Builder{
			cloudflaredef.Builder(&cloudflaredef.RuntimeConfig{APIToken: "test-token", AccountID: "test-account"}),
		},
	})
	assert.NilError(t, err)

	restoreRuntime, err := gala.ReplaceValue(suite.galaRuntime, rt)
	assert.NilError(t, err)
	defer restoreRuntime()

	setup, err := graphapi.SetupListenerRuntime(suite.galaRuntime, hooks.DomainScanListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	fragment, err := integrationtypes.PropertiesFragment(map[string]string{
		"operation":    cloudflaredef.DomainScanRequestOp.Name(),
		"definitionId": cloudflaredef.DefinitionID.ID(),
		"runType":      enums.IntegrationRunTypeEvent.String(),
	})
	assert.NilError(t, err)

	countRuns := func(t *testing.T) int {
		t.Helper()

		count, err := dispatchGala.CountActiveJobsWithMetadata(context.Background(), fragment)
		assert.NilError(t, err)

		return count
	}

	baseline := countRuns(t)

	t.Run("pending system domain scan create dispatches one run", func(t *testing.T) {
		_, err := suite.client.db.Scan.Create().
			SetOwnerID(user.OrganizationID).
			SetTarget("created.dispatch.example.com").
			SetScanType(enums.ScanTypeDomain).
			SetStatus(enums.ScanStatusPending).
			SetPerformedBy(cloudflaredef.DomainScanPerformedBy).
			Save(ctx)
		assert.NilError(t, err)

		waitForCondition(t, func() bool { return countRuns(t) == baseline+1 }, "matching scan create should dispatch one domain scan run")
	})

	t.Run("scan create without the system marker dispatches nothing", func(t *testing.T) {
		_, err := suite.client.db.Scan.Create().
			SetOwnerID(user.OrganizationID).
			SetTarget("manual.dispatch.example.com").
			SetScanType(enums.ScanTypeDomain).
			SetStatus(enums.ScanStatusPending).
			SetPerformedBy("third-party-pentest").
			Save(ctx)
		assert.NilError(t, err)

		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(baseline+1, countRuns(t)))
	})

	t.Run("organization setting domains update dispatches one run per domain", func(t *testing.T) {
		setting, err := suite.client.db.OrganizationSetting.Query().
			Where(organizationsetting.OrganizationID(user.OrganizationID)).
			Only(allowCtx)
		assert.NilError(t, err)

		domains := []string{"one.dispatch.example.com", "two.dispatch.example.com"}

		assert.NilError(t, suite.client.db.OrganizationSetting.UpdateOneID(setting.ID).
			SetDomains(domains).
			Exec(allowCtx))

		waitForCondition(t, func() bool { return countRuns(t) == baseline+1+len(domains) }, "domains update should dispatch one run per current domain")
	})
}
