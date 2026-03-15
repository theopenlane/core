package operations

import (
	"context"
	"encoding/json"
	"maps"
	"time"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/jsonx"
)

// RunStore persists operation run records
type RunStore struct {
	db *ent.Client
}

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

// NewRunStore constructs the run store
func NewRunStore(db *ent.Client) (*RunStore, error) {
	if db == nil {
		return nil, ErrRunStoreRequired
	}

	return &RunStore{db: db}, nil
}

// CreatePending inserts one pending run record for a dispatch request
func (s *RunStore) CreatePending(ctx context.Context, installation *ent.Integration, req DispatchRequest) (*ent.IntegrationRun, error) {
	if installation == nil {
		return nil, ErrInstallationIDRequired
	}

	config, err := rawMessageToMap(req.Config)
	if err != nil {
		return nil, err
	}

	runType := req.RunType
	if runType == "" {
		runType = enums.IntegrationRunTypeManual
	}

	return s.db.IntegrationRun.Create().
		SetOwnerID(installation.OwnerID).
		SetIntegrationID(installation.ID).
		SetOperationName(string(req.Operation)).
		SetRunType(runType).
		SetStatus(enums.IntegrationRunStatusPending).
		SetOperationConfig(config).
		Save(ctx)
}

// Get resolves one persisted run record
func (s *RunStore) Get(ctx context.Context, runID string) (*ent.IntegrationRun, error) {
	if runID == "" {
		return nil, ErrRunIDRequired
	}

	return s.db.IntegrationRun.Get(ctx, runID)
}

// MarkRunning transitions one run to running
func (s *RunStore) MarkRunning(ctx context.Context, runID string) error {
	if runID == "" {
		return ErrRunIDRequired
	}

	return s.db.IntegrationRun.UpdateOneID(runID).
		SetStatus(enums.IntegrationRunStatusRunning).
		SetStartedAt(time.Now()).
		Exec(ctx)
}

// Complete writes the final run outcome
func (s *RunStore) Complete(ctx context.Context, runID string, startedAt time.Time, result RunResult) error {
	if runID == "" {
		return ErrRunIDRequired
	}

	duration := time.Since(startedAt)
	status := result.Status
	if status == "" {
		status = enums.IntegrationRunStatusSuccess
	}

	return s.db.IntegrationRun.UpdateOneID(runID).
		SetStatus(status).
		SetSummary(result.Summary).
		SetError(result.Error).
		SetMetrics(cloneMap(result.Metrics)).
		SetDurationMs(int(duration.Milliseconds())).
		SetFinishedAt(time.Now()).
		Exec(ctx)
}

// rawMessageToMap normalizes one raw JSON payload into a persisted map
func rawMessageToMap(value json.RawMessage) (map[string]any, error) {
	if len(value) == 0 {
		return map[string]any{}, nil
	}

	var out map[string]any
	if err := jsonx.UnmarshalIfPresent(value, &out); err != nil {
		return nil, err
	}

	if out == nil {
		return map[string]any{}, nil
	}

	return out, nil
}

// cloneMap copies one JSON-style map
func cloneMap(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}

	return maps.Clone(value)
}
