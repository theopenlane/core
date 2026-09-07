package operations

import (
	"context"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/contact"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// ingestRelinkAppliers converges adopted rows onto the syncing installation for the hand-written
// persistence paths, one conditional bulk update per schema; catalog-upsert schemas converge
// through the registry's bookkeeping write instead. The integration_id inequality guard keeps
// steady-state syncs write-free
var ingestRelinkAppliers = map[string]func(context.Context, *ent.Client, []string, string) error{
	entityops.SchemaContact.Snake: func(ctx context.Context, db *ent.Client, ids []string, integrationID string) error {
		return db.Contact.Update().Where(contact.IDIn(ids...), contact.IntegrationIDNEQ(integrationID)).SetIntegrationID(integrationID).Exec(ctx)
	},
	entityops.SchemaDirectoryAccount.Snake: func(ctx context.Context, db *ent.Client, ids []string, integrationID string) error {
		return db.DirectoryAccount.Update().Where(directoryaccount.IDIn(ids...), directoryaccount.IntegrationIDNEQ(integrationID)).SetIntegrationID(integrationID).Exec(ctx)
	},
	entityops.SchemaDirectoryGroup.Snake: func(ctx context.Context, db *ent.Client, ids []string, integrationID string) error {
		return db.DirectoryGroup.Update().Where(directorygroup.IDIn(ids...), directorygroup.IntegrationIDNEQ(integrationID)).SetIntegrationID(integrationID).Exec(ctx)
	},
	entityops.SchemaDirectoryMembership.Snake: func(ctx context.Context, db *ent.Client, ids []string, integrationID string) error {
		return db.DirectoryMembership.Update().Where(directorymembership.IDIn(ids...), directorymembership.IntegrationIDNEQ(integrationID)).SetIntegrationID(integrationID).Exec(ctx)
	},
}

// relinkIngestIntegration converges one row's integration linkage to the syncing installation
// without recording an ingest change or emitting mutation events; under a batch the write defers
// to the run's bulk flush so adoption bursts stay chunked
func relinkIngestIntegration(ctx context.Context, db *ent.Client, schema string, id string, integrationID string) error {
	if integrationID == "" {
		return nil
	}

	if batch := directorySyncBatchFromContext(ctx); batch != nil {
		batch.relinkIDs[schema] = append(batch.relinkIDs[schema], id)

		return nil
	}

	return ingestRelinkAppliers[schema](entityops.WithEmissionVetoed(ctx), db, []string{id}, integrationID)
}

// flushIngestRelinks applies the batch's queued relinks in chunked conditional updates
func flushIngestRelinks(ctx context.Context, db *ent.Client, batch *directorySyncBatch, integrationID string) error {
	if integrationID == "" {
		return nil
	}

	vetoCtx := entityops.WithEmissionVetoed(ctx)

	for schema, ids := range batch.relinkIDs {
		for _, chunk := range lo.Chunk(ids, directoryConfirmChunkSize) {
			if err := ingestRelinkAppliers[schema](vetoCtx, db, chunk, integrationID); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("schema", schema).Int("records", len(chunk)).Msg("ingest integration relink flush failed")

				return err
			}
		}
	}

	return nil
}

// ingestIntegrationLinked reports whether the record already carries the syncing installation's
// integrations edge; under a batch the installation's linked id set loads once per run, otherwise
// the record's edge is checked directly
func ingestIntegrationLinked(ctx context.Context, schema string, id string, loadLinked func(context.Context) ([]string, error), linkExists func(context.Context) (bool, error)) (bool, error) {
	batch := directorySyncBatchFromContext(ctx)
	if batch == nil {
		return linkExists(ctx)
	}

	set, ok := batch.linkedIntegrationIDs[schema]
	if !ok {
		ids, err := loadLinked(ctx)
		if err != nil {
			return false, err
		}

		set = lo.SliceToMap(ids, func(id string) (string, struct{}) { return id, struct{}{} })
		batch.linkedIntegrationIDs[schema] = set
	}

	_, linked := set[id]

	return linked, nil
}

// markIngestIntegrationLinked records a newly written integrations edge in the batch's linked id set
func markIngestIntegrationLinked(ctx context.Context, schema string, id string) {
	if batch := directorySyncBatchFromContext(ctx); batch != nil {
		if set, ok := batch.linkedIntegrationIDs[schema]; ok {
			set[id] = struct{}{}
		}
	}
}

// ensureIngestIntegrationLink writes the record's integrations edge when it is missing, leaving
// already-linked records untouched so unchanged syncs stay write-free; a link that fails is
// retried on the next sync because the record still reads as unlinked
func ensureIngestIntegrationLink(ctx context.Context, schema string, id string, integrationID string, loadLinked func(context.Context) ([]string, error), linkExists func(context.Context) (bool, error), link func(context.Context) error) error {
	if integrationID == "" {
		return nil
	}

	linked, err := ingestIntegrationLinked(ctx, schema, id, loadLinked, linkExists)
	if err != nil {
		return err
	}

	if linked {
		return nil
	}

	if err := link(entityops.WithEmissionVetoed(ctx)); err != nil {
		return err
	}

	markIngestIntegrationLinked(ctx, schema, id)

	return nil
}
