package operations

import (
	"context"
	"time"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
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

// CreatePendingRun inserts one pending run record for a dispatch request
func CreatePendingRun(ctx context.Context, db *ent.Client, installation *ent.Integration, req DispatchRequest) (*ent.IntegrationRun, error) {
	if installation == nil {
		return nil, ErrInstallationIDRequired
	}

	config, err := jsonx.ToMap(req.Config)
	if err != nil {
		return nil, err
	}

	return db.IntegrationRun.Create().
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		SetOperationName(req.Operation).
		SetRunType(req.RunType).
		SetStatus(enums.IntegrationRunStatusPending).
		SetOperationConfig(config).
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
