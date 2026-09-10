//go:build test

package eventstest_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterwatermarkconfig"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
)

// watermarkJobCount counts enqueued watermark jobs for one document so assertions stay
// isolated from jobs enqueued for other documents in the shared river tables
func watermarkJobCount(ctx context.Context, t *testing.T, docID string) int {
	t.Helper()

	var count int

	err := suite.Client.DB.Job.GetPool().QueryRow(ctx,
		"SELECT count(*) FROM river_job WHERE kind = $1 AND args->>'trust_center_document_id' = $2",
		(jobspec.WatermarkDocArgs{}).Kind(), docID).Scan(&count)
	assert.NilError(t, err)

	return count
}

func TestTrustCenterWatermarkListeners(t *testing.T) {
	ctx := context.Background()

	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter
	dbCtx := privacy.DecisionContext(th.SetContext(tcOrg.Owner.UserCtx, suite.Client.DB), privacy.Allow)

	watermarkConfig, err := suite.Client.DB.TrustCenterWatermarkConfig.Query().
		Where(trustcenterwatermarkconfig.TrustCenterID(trustCenter.ID)).
		Only(dbCtx)
	assert.NilError(t, err)

	err = suite.Client.DB.TrustCenterWatermarkConfig.UpdateOne(watermarkConfig).
		SetIsEnabled(true).
		Exec(dbCtx)
	assert.NilError(t, err)

	docKind := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_doc",
	}).MustNew(tcOrg.Owner.UserCtx, t)

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.TrustCenterWatermarkListeners())
	assert.NilError(t, err)
	t.Cleanup(setup.Teardown)

	var doc *generated.TrustCenterDoc

	t.Run("create with watermarking enabled enqueues job", func(t *testing.T) {
		doc = (&th.TrustCenterDocBuilder{
			Client:        suite.Client,
			TrustCenterID: trustCenter.ID,
		}).MustNew(tcOrg.Owner.UserCtx, t)

		assert.Assert(t, doc.WatermarkingEnabled)
		assert.Assert(t, doc.OriginalFileID != nil)

		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(1, watermarkJobCount(ctx, t, doc.ID)))
	})

	t.Run("update without original file change does not enqueue", func(t *testing.T) {
		err := suite.Client.DB.TrustCenterDoc.UpdateOneID(doc.ID).
			SetTitle("Watermark Listener Doc Renamed").
			Exec(dbCtx)
		assert.NilError(t, err)

		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(1, watermarkJobCount(ctx, t, doc.ID)))
	})

	t.Run("update changing original file enqueues", func(t *testing.T) {
		replacement := (&th.FileBuilder{
			Client: suite.Client,
			Name:   "watermark-listener-replacement.pdf",
		}).MustNew(tcOrg.Owner.UserCtx, t)

		err := suite.Client.DB.TrustCenterDoc.UpdateOneID(doc.ID).
			SetOriginalFileID(replacement.ID).
			Exec(dbCtx)
		assert.NilError(t, err)

		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(2, watermarkJobCount(ctx, t, doc.ID)))
	})

	t.Run("watermarking disabled doc acks without job", func(t *testing.T) {
		disabledDoc, err := suite.Client.DB.TrustCenterDoc.Create().
			SetTitle("Watermark Listener Disabled Doc").
			SetTrustCenterDocKindName(docKind.Name).
			SetTrustCenterID(trustCenter.ID).
			SetWatermarkingEnabled(false).
			Save(dbCtx)
		assert.NilError(t, err)

		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(0, watermarkJobCount(ctx, t, disabledDoc.ID)))
	})

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}
