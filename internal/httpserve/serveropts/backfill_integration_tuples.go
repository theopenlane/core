package serveropts

import (
	"context"

	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationwebhook"
)

// integrationParentTupleBatchSize is the number of rows paged, and tuples deleted, per batch
const integrationParentTupleBatchSize = 100

// cleanupIntegrationParentTuples deletes parent tuples on runs and webhooks plus run org tuples, none are written or used anymore
func cleanupIntegrationParentTuples(ctx context.Context, deps backfillDeps) error {
	if err := cleanupIntegrationRunParentTuples(ctx, deps.Client); err != nil {
		return err
	}

	return cleanupIntegrationWebhookParentTuples(ctx, deps.Client)
}

// cleanupIntegrationRunParentTuples pages every integration run by ID and deletes its integration parent and org parent_context tuples
func cleanupIntegrationRunParentTuples(ctx context.Context, client *generated.Client) error {
	objectType := strcase.SnakeCase(generated.TypeIntegrationRun)

	var lastID string

	for {
		rows, err := client.IntegrationRun.Query().
			Where(integrationrun.IDGT(lastID)).
			Order(integrationrun.ByID()).
			Select(integrationrun.FieldID, integrationrun.FieldIntegrationID, integrationrun.FieldOwnerID).
			Limit(integrationParentTupleBatchSize).
			All(ctx)
		if err != nil {
			return err
		}

		if len(rows) == 0 {
			logx.FromContext(ctx).Info().Msg("backfill: integration run parent tuples removed")

			return nil
		}

		deletes := make([]fgax.TupleKey, 0, len(rows)*2) //nolint:mnd
		for _, row := range rows {
			if row.IntegrationID != "" {
				deletes = append(deletes, integrationParentTuple(row.IntegrationID, objectType, row.ID))
			}

			if row.OwnerID != "" {
				deletes = append(deletes, orgParentContextTuple(row.OwnerID, objectType, row.ID))
			}
		}

		if err := deleteIntegrationParentTuples(ctx, client, deletes); err != nil {
			return err
		}

		lastID = rows[len(rows)-1].ID
	}
}

// cleanupIntegrationWebhookParentTuples pages every integration webhook by ID and deletes its integration parent tuple
func cleanupIntegrationWebhookParentTuples(ctx context.Context, client *generated.Client) error {
	objectType := strcase.SnakeCase(generated.TypeIntegrationWebhook)

	var lastID string

	for {
		rows, err := client.IntegrationWebhook.Query().
			Where(integrationwebhook.IDGT(lastID)).
			Order(integrationwebhook.ByID()).
			Select(integrationwebhook.FieldID, integrationwebhook.FieldIntegrationID).
			Limit(integrationParentTupleBatchSize).
			All(ctx)
		if err != nil {
			return err
		}

		if len(rows) == 0 {
			logx.FromContext(ctx).Info().Msg("backfill: integration webhook parent tuples removed")

			return nil
		}

		deletes := make([]fgax.TupleKey, 0, len(rows))
		for _, row := range rows {
			if row.IntegrationID != "" {
				deletes = append(deletes, integrationParentTuple(row.IntegrationID, objectType, row.ID))
			}
		}

		if err := deleteIntegrationParentTuples(ctx, client, deletes); err != nil {
			return err
		}

		lastID = rows[len(rows)-1].ID
	}
}

// integrationParentTuple builds the integration:<id> parent <objectType>:<objectID> tuple key
func integrationParentTuple(integrationID, objectType, objectID string) fgax.TupleKey {
	return fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:   integrationID,
		SubjectType: strcase.SnakeCase(generated.TypeIntegration),
		ObjectID:    objectID,
		ObjectType:  objectType,
		Relation:    fgax.ParentRelation,
	})
}

// orgParentContextTuple builds the organization:<id> parent_context <objectType>:<objectID> tuple key
func orgParentContextTuple(orgID, objectType, objectID string) fgax.TupleKey {
	return fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:   orgID,
		SubjectType: strcase.SnakeCase(generated.TypeOrganization),
		ObjectID:    objectID,
		ObjectType:  objectType,
		Relation:    fgax.ParentContextRelation,
	})
}

// deleteIntegrationParentTuples deletes the tuples, ignoring any that are already gone so re-runs are no-ops
func deleteIntegrationParentTuples(ctx context.Context, client *generated.Client, deletes []fgax.TupleKey) error {
	if len(deletes) == 0 {
		return nil
	}

	_, err := client.Authz.WriteTupleKeys(ctx, nil, deletes, fgax.WithIgnoreDuplicateKeyError(true))

	return err
}
