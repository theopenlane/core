package aikido

import (
	"context"
	"fmt"

	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/mapx"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// autofixIssueIDsKey is the payload key the targeted issue ids are attached under for the remediation links
	autofixIssueIDsKey = "targeted_issue_ids"
	// autofixStatusMerged is the status of an autofix whose pull request was merged, the only ones that remediated anything
	autofixStatusMerged = "pr_merged"
)

// runAutofixSync collects Aikido autofix tasks for remediation ingestion
func runAutofixSync(ctx context.Context, _ types.OperationRequest, client *APIClient, _ AutofixSync) ([]types.IngestPayloadSet, error) {
	return CollectAutofixes(ctx, client)
}

// CollectAutofixes lists the autofix history and emits one remediation envelope per merged task with the issue ids it
// targets, a preview or closed pull request remediated nothing
func CollectAutofixes(ctx context.Context, c *APIClient) ([]types.IngestPayloadSet, error) {
	tasks, err := c.listAll(ctx, pathAutofixHistory, nil, autofixPageSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAutofixFetchFailed, err)
	}

	envelopes := make([]types.MappingEnvelope, 0, len(tasks))

	for _, task := range tasks {
		if mapx.Value[string](task, "status") != autofixStatusMerged {
			continue
		}

		taskID := recordID(task)

		issueIDs, err := c.autofixIssueIDs(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("%w: task %s: %w", ErrAutofixFetchFailed, taskID, err)
		}

		task[autofixIssueIDsKey] = issueIDs

		title := mapx.Value[string](mapx.Value[map[string]any](task, "autofix_target"), "title")

		envelope, err := providerkit.MarshalEnvelope(title, mapx.PruneMapZeroAny(task), ErrPayloadEncode)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("autofix_id", taskID).Msg("aikido: failed to marshal autofix envelope")

			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaRemediation.Name,
			Envelopes: envelopes,
		},
	}, nil
}
