package operations

import (
	"context"
	"encoding/json"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/mapx"
)

// RunResult captures the terminal state of one run
type RunResult struct {
	// Status is the terminal run status
	Status enums.IntegrationRunStatus
	// Summary is the optional summary text stored on the run
	Summary string
	// Error is the optional terminal error text stored on the run
	Error string
	// Metrics is the structured metrics payload stored on the run
	Metrics map[string]any
}

// operationKind classifies an operation's execution shape as an IntegrationOperationKind
func operationKind(operation types.OperationRegistration) enums.IntegrationOperationKind {
	if operation.IngestHandle != nil {
		return enums.IntegrationOperationKindSync
	}

	return enums.IntegrationOperationKindPush
}

// CreatePendingRun inserts one pending run record for a dispatched operation
func CreatePendingRun(ctx context.Context, db *ent.Client, installation *ent.Integration, operation types.OperationRegistration, runType enums.IntegrationRunType, config json.RawMessage) (*ent.IntegrationRun, error) {
	if installation == nil {
		return nil, ErrInstallationIDRequired
	}

	configMap, err := jsonx.ToMap(config)
	if err != nil {
		return nil, err
	}

	return db.IntegrationRun.Create().
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		SetOperationName(operation.Name).
		SetOperationKind(operationKind(operation)).
		SetRunType(runType).
		SetStatus(enums.IntegrationRunStatusPending).
		SetOperationConfig(configMap).
		Save(ctx)
}

// metricRetryOf is the metrics key recording the run a re-executed attempt continues from
const metricRetryOf = "retry_of"

// RetryRun inserts a fresh pending run for a re-executed attempt of the given run, carrying its
// operation, run type, and config, so every attempt writes records under its own run id
func RetryRun(ctx context.Context, db *ent.Client, run *ent.IntegrationRun) (*ent.IntegrationRun, error) {
	return db.IntegrationRun.Create().
		SetOwnerID(run.OwnerID).
		SetIntegrationID(run.IntegrationID).
		SetOperationName(run.OperationName).
		SetNillableOperationKind(lo.EmptyableToPtr(run.OperationKind)).
		SetRunType(run.RunType).
		SetStatus(enums.IntegrationRunStatusPending).
		SetOperationConfig(run.OperationConfig).
		SetMetrics(map[string]any{metricRetryOf: run.ID}).
		Save(ctx)
}

// MarkRunRunning transitions one run to running
func MarkRunRunning(ctx context.Context, db *ent.Client, runID string) error {
	if runID == "" {
		return ErrRunIDRequired
	}

	return db.IntegrationRun.UpdateOneID(runID).
		SetStatus(enums.IntegrationRunStatusRunning).
		SetStartedAt(time.Now()).
		Exec(ctx)
}

// CompleteRun writes the final run outcome
func CompleteRun(ctx context.Context, db *ent.Client, runID string, startedAt time.Time, result RunResult) error {
	if runID == "" {
		return ErrRunIDRequired
	}

	duration := time.Since(startedAt)
	status := result.Status
	if status == "" {
		status = enums.IntegrationRunStatusSuccess
	}

	return db.IntegrationRun.UpdateOneID(runID).
		SetStatus(status).
		SetSummary(result.Summary).
		SetError(result.Error).
		SetMetrics(mapx.DeepCloneMapAny(result.Metrics)).
		SetDurationMs(int(duration.Milliseconds())).
		SetFinishedAt(time.Now()).
		Exec(ctx)
}

// LastSuccessfulRunAt returns the finish time of the most recent successful run for the given
// integration and operation, or nil if no successful run exists yet
func LastSuccessfulRunAt(ctx context.Context, db *ent.Client, integrationID, operationName string) (*time.Time, error) {
	run, err := db.IntegrationRun.Query().
		Where(
			integrationrun.IntegrationIDEQ(integrationID),
			integrationrun.OperationNameEQ(operationName),
			integrationrun.StatusEQ(enums.IntegrationRunStatusSuccess),
			integrationrun.FinishedAtNotNil(),
		).
		Order(integrationrun.ByFinishedAt(sql.OrderDesc())).
		Select(integrationrun.FieldFinishedAt).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return run.FinishedAt, nil
}
