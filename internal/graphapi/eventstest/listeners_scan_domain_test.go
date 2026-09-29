//go:build test

package eventstest_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/notification"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/platform"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	cloudflaredef "github.com/theopenlane/core/v2/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/core/v2/pkg/gala"
)

func TestDomainScanListeners(t *testing.T) {
	org := suite.SeedOrgOwner(t)
	user := org.Owner
	ctx := org.Owner.UserCtx

	// workers on the dispatch runtime are never started, so dispatched cloudflare runs
	// stay queued for counting instead of executing against the fake credentials
	dispatchGala, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode:  gala.DispatchModeDurable,
		ConnectionURI: suite.TF.URI,
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

	credStore, err := keystore.NewStore(suite.Client.DB)
	assert.NilError(t, err)

	rt, err := intruntime.New(intruntime.Config{
		DB:          suite.Client.DB,
		Gala:        dispatchGala,
		Keystore:    credStore,
		RedisClient: coreutils.NewRedisClient(),
		DefinitionBuilders: []registry.Builder{
			cloudflaredef.Builder(&cloudflaredef.RuntimeConfig{APIToken: "test-token", AccountID: "test-account"}),
		},
	})
	assert.NilError(t, err)

	restoreRuntime, err := gala.ReplaceValue(suite.GalaRuntime, rt)
	assert.NilError(t, err)
	defer restoreRuntime()

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.DomainScanListeners())
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
		_, err := suite.Client.DB.Scan.Create().
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
		_, err := suite.Client.DB.Scan.Create().
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
		setting, err := suite.Client.DB.OrganizationSetting.Query().
			Where(organizationsetting.OrganizationID(user.OrganizationID)).
			Only(ctx)
		assert.NilError(t, err)

		domains := []string{"one.dispatch.example.com", "two.dispatch.example.com"}

		assert.NilError(t, suite.Client.DB.OrganizationSetting.UpdateOneID(setting.ID).
			SetDomains(domains).
			Exec(ctx))

		waitForCondition(t, func() bool { return countRuns(t) == baseline+1+len(domains) }, "domains update should dispatch one run per current domain")
	})
}

func TestImportDomainScanReviewRequiresCreatePermissions(t *testing.T) {
	org := suite.SeedFreshMinimalOrgUsers(t, false)
	owner := org.Owner
	riskManager := suite.OrgMemberWithFunctionalRoles(t, *owner, "risk_manager")

	ctx := owner.UserCtx

	domainScan, err := suite.Client.DB.Scan.Create().
		SetOwnerID(owner.OrganizationID).
		SetTarget("import-" + ulids.New().String() + ".example.com").
		SetScanType(enums.ScanTypeDomain).
		SetStatus(enums.ScanStatusCompleted).
		SetPerformedBy("third-party-pentest").
		Save(ctx)
	assert.NilError(t, err)

	_, err = suite.Client.API.GetScanByID(riskManager.UserCtx, domainScan.ID)
	assert.NilError(t, err)

	testCases := []struct {
		name          string
		ctx           context.Context
		withPlatform  bool
		expectCreated bool
	}{
		{
			name:          "happy path, owner import creates the reviewed vendor",
			ctx:           owner.UserCtx,
			expectCreated: true,
		},
		{
			name:          "happy path, owner import creates a platform linked to the scan",
			ctx:           owner.UserCtx,
			withPlatform:  true,
			expectCreated: true,
		},
		{
			name:          "risk manager who can view scans cannot create records through import",
			ctx:           riskManager.UserCtx,
			expectCreated: false,
		},
		{
			name:          "member cannot create objects",
			ctx:           org.Member.UserCtx,
			expectCreated: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vendorName := "import-vendor-" + ulids.New().String()
			platformName := "import-platform-" + ulids.New().String()

			input := domainScanReviewInput(domainScan.ID, vendorName)
			if tc.withPlatform {
				input.Platforms = []*testclient.ImportDomainScanReviewPlatformInput{{Ref: "platform-1", Name: platformName}}
			}

			_, err := suite.Client.API.ImportDomainScanReview(tc.ctx, input)
			assert.NilError(t, err)

			if !tc.expectCreated {
				waitForGala(t, suite.GalaRuntime)

				assert.Check(t, !domainScanVendorExists(ctx, t, owner.OrganizationID, vendorName))

				return
			}

			waitForCondition(t, func() bool {
				return domainScanVendorExists(ctx, t, owner.OrganizationID, vendorName)
			}, "import should create the vendor entity")

			if tc.withPlatform {
				assert.Check(t, domainScanPlatformLinked(ctx, t, platformName, domainScan.ID))
			}
		})
	}
}

func TestDomainScanPollFinalizesAndNotifiesGroup(t *testing.T) {
	org := suite.SeedOrgOwner(t)
	owner := org.Owner
	ctx := org.Owner.UserCtx

	readyScan := createProcessingDomainScan(ctx, t, owner.OrganizationID)
	brokenScan := createProcessingDomainScan(ctx, t, owner.OrganizationID)
	siblings := []string{readyScan.ID, brokenScan.ID}

	readyResultID := "ready-" + ulids.New().String()
	brokenResultID := "broken-" + ulids.New().String()

	suite.CloudflareMock.SetScanResult(readyResultID, http.StatusOK)
	suite.CloudflareMock.SetScanResult(brokenResultID, http.StatusInternalServerError)

	assert.NilError(t, cloudflaredef.EmitDomainScanPoll(owner.UserCtx, suite.GalaRuntime, cloudflaredef.DomainScanPollEnvelope{
		OrganizationID: owner.OrganizationID,
		ScanResultID:   readyResultID,
		InternalScanID: readyScan.ID,
		SiblingScanIDs: siblings,
	}))

	waitForCondition(t, func() bool {
		return domainScanStatus(ctx, t, readyScan.ID) == enums.ScanStatusCompleted
	}, "a successful poll should complete the scan")

	assert.Check(t, is.Len(domainScanGroupNotifications(ctx, t, owner.OrganizationID), 0))

	assert.NilError(t, cloudflaredef.EmitDomainScanPoll(owner.UserCtx, suite.GalaRuntime, cloudflaredef.DomainScanPollEnvelope{
		OrganizationID: owner.OrganizationID,
		ScanResultID:   brokenResultID,
		InternalScanID: brokenScan.ID,
		SiblingScanIDs: siblings,
	}))

	waitForCondition(t, func() bool {
		return len(domainScanGroupNotifications(ctx, t, owner.OrganizationID)) == 1
	}, "the last sibling to finish should send one group notification")

	assert.Check(t, is.Equal(enums.ScanStatusFailed, domainScanStatus(ctx, t, brokenScan.ID)))
	assert.Check(t, strings.Contains(domainScanGroupNotifications(ctx, t, owner.OrganizationID)[0].Body, "1 domain(s) failed"))
}

func domainScanReviewInput(scanID, vendorName string) testclient.ImportDomainScanReviewInput {
	return testclient.ImportDomainScanReviewInput{
		ScanIDs: []string{scanID},
		Vendors: []*testclient.ImportDomainScanReviewVendorInput{{Ref: "vendor-1", Name: vendorName}},
		Assets:  []*testclient.ImportDomainScanReviewAssetInput{},
	}
}

func createProcessingDomainScan(ctx context.Context, t *testing.T, orgID string) *generated.Scan {
	t.Helper()

	created, err := suite.Client.DB.Scan.Create().
		SetOwnerID(orgID).
		SetTarget("poll-" + strings.ToLower(ulids.New().String()) + ".example.com").
		SetScanType(enums.ScanTypeDomain).
		SetStatus(enums.ScanStatusProcessing).
		SetPerformedBy("third-party-pentest").
		Save(ctx)
	assert.NilError(t, err)

	return created
}

func domainScanStatus(ctx context.Context, t *testing.T, scanID string) enums.ScanStatus {
	t.Helper()

	found, err := suite.Client.DB.Scan.Get(ctx, scanID)
	assert.NilError(t, err)

	return found.Status
}

func domainScanVendorExists(ctx context.Context, t *testing.T, orgID, name string) bool {
	t.Helper()

	exists, err := suite.Client.DB.Entity.Query().
		Where(entity.OwnerIDEQ(orgID), entity.NameEQ(name)).
		Exist(ctx)
	assert.NilError(t, err)

	return exists
}

func domainScanPlatformLinked(ctx context.Context, t *testing.T, name, scanID string) bool {
	t.Helper()

	exists, err := suite.Client.DB.Platform.Query().
		Where(
			platform.NameEQ(name),
			platform.HasGeneratedScansWith(scan.ID(scanID)),
		).
		Exist(ctx)
	assert.NilError(t, err)

	return exists
}

func domainScanGroupNotifications(ctx context.Context, t *testing.T, orgID string) []*generated.Notification {
	t.Helper()

	found, err := suite.Client.DB.Notification.Query().
		Where(notification.OwnerIDEQ(orgID), notification.ObjectTypeEQ("scan.created")).
		All(ctx)
	assert.NilError(t, err)

	return found
}
