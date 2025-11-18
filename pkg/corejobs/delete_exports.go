package corejobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// DeleteExportContentArgs for the worker to process deletion of exports
type DeleteExportContentArgs struct {
}

type DeleteExportWorkerConfig struct {
	// embed OpenlaneConfig to reuse validation and client creation logic
	OpenlaneConfig `koanf:",squash" jsonschema:"description=the openlane API configuration for deleting exports"`

	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the delete export worker is enabled"`

	Interval time.Duration `koanf:"interval" json:"interval" jsonschema:"required,default=10m description=the interval at which to run the delete export worker"`

	// CutoffDuration defines the tolerance for exports. If you set 30 minutes, all exports older than 30 minutes
	// at the time of job execution will be deleted
	CutoffDuration time.Duration `koanf:"cutoffduration" json:"cutoffduration" jsonschema:"required,default=30m description=how long do you want exports to exist before they are deleted"`
}

// Kind satisfies the river.Job interface
func (DeleteExportContentArgs) Kind() string { return "delete_export_content" }

// InsertOpts provides the insertion options for the delete export content job
func (DeleteExportContentArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{MaxAttempts: 3} //nolint:mnd
}

// DeleteExportContentWorker deletes exports that are older than the configured cutoff duration
type DeleteExportContentWorker struct {
	river.WorkerDefaults[DeleteExportContentArgs]

	Config DeleteExportWorkerConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for deleting exports"`

	olClient olclient.OpenlaneClient
}

// WithOpenlaneClient sets the Openlane client for the worker
// and returns the worker for method chaining
func (w *DeleteExportContentWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *DeleteExportContentWorker {
	w.olClient = cl
	return w
}

// Work satisfies the river.Worker interface for the delete export worker
// it deletes exports that are older than the configured cutoff duration
func (w *DeleteExportContentWorker) Work(ctx context.Context, _ *river.Job[DeleteExportContentArgs]) error {
	if w.olClient == nil {
		cl, err := w.Config.getOpenlaneClient()
		if err != nil {
			return err
		}

		w.olClient = cl
	}

	cutOffTime := time.Now().Add(-w.Config.CutoffDuration)

	exports, err := w.olClient.GetExports(ctx, nil, nil, &openlaneclient.ExportWhereInput{
		CreatedAtLte: &cutOffTime,
		StatusIn: []enums.ExportStatus{
			enums.ExportStatusNodata,
			enums.ExportStatusReady,
		},
	})
	if err != nil {
		return err
	}

	if len(exports.Exports.Edges) == 0 {
		return nil
	}

	var ids = make([]string, 0, len(exports.Exports.Edges))

	for _, export := range exports.Exports.Edges {
		ids = append(ids, export.Node.ID)
	}

	_, err = w.olClient.DeleteBulkExport(ctx, ids)

	return err
}
