//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/entdb"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	systemdef "github.com/theopenlane/core/v2/internal/integrations/definitions/system"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// runFileOrphanSweep runs the sweep through the runtime from a bare context
func runFileOrphanSweep(t *testing.T, sweep systemdef.FileOrphanSweep) int {
	t.Helper()

	config, err := json.Marshal(sweep)
	assert.NilError(t, err)

	raw, err := suite.IntegrationsRT.ExecuteRuntimeOperation(context.Background(), systemdef.DefinitionID.ID(), systemdef.FileOrphanOp.Name(), config)
	assert.NilError(t, err)

	var result integrationtypes.ScheduledCycleResult
	assert.NilError(t, json.Unmarshal(raw, &result))

	return result.Processed
}

func TestFileOrphanSweep(t *testing.T) {
	user := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	ctx := th.SetInternalContext(user.UserCtx, suite.Client.DB)

	s3 := storage.S3Provider.String()
	r2 := storage.R2Provider.String()

	orphanS3 := seedFile(t, ctx, "orphan-s3.txt", s3, relocateSippySource, nil)
	orphanR2 := seedFile(t, ctx, "orphan-r2.txt", r2, relocateTargetBucket, nil)
	recent := seedFile(t, ctx, "recent-orphan.txt", r2, relocateTargetBucket, nil)

	referenced := seedFile(t, ctx, "referenced.txt", r2, relocateTargetBucket, nil)
	attachContact(t, ctx, referenced.ID)

	// held by a foreign key on the files table rather than an edge table
	diagram := seedFile(t, ctx, "diagram.png", r2, relocateTargetBucket, nil)
	attachPlatformDiagram(t, ctx, diagram.ID)

	softParent := seedFile(t, ctx, "soft-parent.txt", r2, relocateTargetBucket, nil)
	parent := attachContact(t, ctx, softParent.ID)
	assert.NilError(t, suite.Client.DB.Contact.DeleteOneID(parent.ID).Exec(ctx))

	// a soft deleted file keeps its references
	softReferenced := seedFile(t, ctx, "soft-referenced.txt", r2, relocateTargetBucket, nil)
	attachContact(t, ctx, softReferenced.ID)
	assert.NilError(t, suite.Client.DB.File.DeleteOneID(softReferenced.ID).Exec(ctx))

	softOrphan := seedFile(t, ctx, "soft-orphan.txt", r2, relocateTargetBucket, nil)
	assert.NilError(t, suite.Client.DB.File.DeleteOneID(softOrphan.ID).Exec(ctx))

	assert.Check(t, !suite.Client.DeletedStorageKeys.Has(softOrphan.StoragePath), "soft delete keeps the object")

	backdate(t, ctx, orphanS3.ID, orphanR2.ID, referenced.ID, diagram.ID, softParent.ID, softReferenced.ID, softOrphan.ID)

	exists := func(id string) bool {
		ok, err := suite.Client.DB.File.Query().Where(file.ID(id)).Exist(entx.SkipSoftDelete(ctx))
		assert.NilError(t, err)

		return ok
	}

	kept := []*generated.File{recent, referenced, diagram, softParent, softReferenced}
	swept := []*generated.File{orphanS3, orphanR2, softOrphan}

	t.Run("dry run deletes nothing", func(t *testing.T) {
		processed := runFileOrphanSweep(t, systemdef.FileOrphanSweep{DryRun: true})

		assert.Check(t, processed >= len(swept), "every aged orphan is reported")

		for _, seeded := range append(kept, swept...) {
			assert.Check(t, exists(seeded.ID), seeded.ProvidedFileName)
			assert.Check(t, !suite.Client.DeletedStorageKeys.Has(seeded.StoragePath), seeded.ProvidedFileName)
		}
	})

	t.Run("aged orphans are hard deleted with their objects", func(t *testing.T) {
		processed := runFileOrphanSweep(t, systemdef.FileOrphanSweep{})

		assert.Check(t, processed >= len(swept))

		for _, seeded := range swept {
			assert.Check(t, !exists(seeded.ID), seeded.ProvidedFileName)
			assert.Check(t, suite.Client.DeletedStorageKeys.Has(seeded.StoragePath), seeded.ProvidedFileName)
		}
	})

	t.Run("recent, referenced and soft deleted parent files are kept with their objects", func(t *testing.T) {
		for _, seeded := range kept {
			assert.Check(t, exists(seeded.ID), seeded.ProvidedFileName)
			assert.Check(t, !suite.Client.DeletedStorageKeys.Has(seeded.StoragePath), seeded.ProvidedFileName)
		}
	})

	t.Run("rerun leaves the kept rows alone", func(t *testing.T) {
		runFileOrphanSweep(t, systemdef.FileOrphanSweep{})

		for _, seeded := range kept {
			assert.Check(t, exists(seeded.ID), seeded.ProvidedFileName)
		}
	})
}

// attachPlatformDiagram references the file from a platform diagram removed after the test
func attachPlatformDiagram(t *testing.T, ctx context.Context, fileID string) {
	t.Helper()

	p, err := suite.Client.DB.Platform.Create().SetName("sweep " + fileID).AddArchitectureDiagramIDs(fileID).Save(ctx)
	assert.NilError(t, err)

	t.Cleanup(func() {
		if err := suite.Client.DB.Platform.DeleteOneID(p.ID).Exec(entx.SkipSoftDelete(ctx)); err != nil && !generated.IsNotFound(err) {
			t.Error(err)
		}
	})
}

// backdate ages rows past the sweep's minimum age; the audit hook stamps created_at on create
func backdate(t *testing.T, ctx context.Context, ids ...string) {
	t.Helper()

	db := entdb.SQLDB(suite.Client.DB)
	createdAt := time.Now().Add(-systemdef.FileOrphanMinAge - time.Hour)

	for _, id := range ids {
		_, err := db.ExecContext(ctx, "UPDATE files SET created_at = $1 WHERE id = $2", createdAt, id)
		assert.NilError(t, err)
	}
}
