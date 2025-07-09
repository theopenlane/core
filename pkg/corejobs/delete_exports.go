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
	OpenlaneAPIHost  string `koanf:"openlaneAPIHost" json:"openlaneAPIHost" jsonschema:"required description=the openlane api host"`
	OpenlaneAPIToken string `koanf:"openlaneAPIToken" json:"openlaneAPIToken" jsonschema:"required description=the openlane api token"`

	// CutoffDuration defines the tolerance for exports. If you set 30 minutes, all exports older than 30 minutes
	// at the time of job execution will be deleted
	CutoffDuration time.Duration `koanf:"cutoffDuration" json:"cutoffDuration" jsonschema:"required,default=30m description=how long do you want exports to exist before they are deleted"`
}

// Kind satisfies the river.Job interface
func (DeleteExportContentArgs) Kind() string { return "delete_export_content" }

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
func (w *DeleteExportContentWorker) Work(ctx context.Context, job *river.Job[DeleteExportContentArgs]) error {

	if w.olClient == nil {
		cl, err := getOpenlaneClient(CustomDomainConfig{
			OpenlaneAPIHost:  w.Config.OpenlaneAPIHost,
			OpenlaneAPIToken: w.Config.OpenlaneAPIToken,
		})
		if err != nil {
			return err
		}

		w.olClient = cl
	}

	cutOffTime := time.Now().Add(-w.Config.CutoffDuration)

	exports, err := w.olClient.GetExports(ctx, nil, nil, &openlaneclient.ExportWhereInput{
		CreatedAtLte: &cutOffTime,
		Status:       &enums.ExportStatusReady,
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
