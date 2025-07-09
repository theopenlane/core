package corejobs

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gocarina/gocsv"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// ExportContentArgs for the worker to process the custom domain
type ExportContentArgs struct {
	// ID of the export
	ExportID string `json:"export_id,omitempty"`
}

type ExportWorkerConfig struct {
	OpenlaneAPIHost  string `koanf:"openlaneAPIHost" json:"openlaneAPIHost" jsonschema:"required description=the openlane api host"`
	OpenlaneAPIToken string `koanf:"openlaneAPIToken" json:"openlaneAPIToken" jsonschema:"required description=the openlane api token"`
}

// Kind satisfies the river.Job interface
func (ExportContentArgs) Kind() string { return "export_content" }

// ExportContentWorker creates a custom hostname in cloudflare, and
// creates and updates the records in our system
type ExportContentWorker struct {
	river.WorkerDefaults[ExportContentArgs]

	Config ExportWorkerConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for exporting"`

	olClient olclient.OpenlaneClient
}

// WithOpenlaneClient sets the Openlane client for the worker
// and returns the worker for method chaining
func (w *ExportContentWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *ExportContentWorker {
	w.olClient = cl
	return w
}

// Work satisfies the river.Worker interface for the export content worker
// it creates a csv, uploads it and associates it with the export
func (w *ExportContentWorker) Work(ctx context.Context, job *river.Job[ExportContentArgs]) error {
	log.Info().Str("export_id", job.Args.ExportID).Msg("exporting content")

	if job.Args.ExportID == "" {
		return newMissingRequiredArg("export_id", ExportContentArgs{}.Kind())
	}

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

	export, err := w.olClient.GetExportByID(ctx, job.Args.ExportID)
	if err != nil {
		log.Error().Err(err).Str("export_id", job.Args.ExportID).Msg("failed to get export")
		return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed)
	}

	switch export.Export.ExportType {
	case enums.ExportTypeControl:
		return w.exportControls(ctx, job.Args.ExportID)
	default:
		log.Error().Str("export_type", string(export.Export.ExportType)).Msg("unsupported export type")
		return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed)
	}
}

func (w *ExportContentWorker) exportControls(ctx context.Context, exportID string) error {
	controls, err := w.olClient.GetAllControls(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch controls")
		return w.updateExportStatus(ctx, exportID, enums.ExportStatusFailed)
	}

	if len(controls.Controls.Edges) == 0 {
		log.Info().Msg("no controls found for export")
		return w.updateExportStatus(ctx, exportID, enums.ExportStatusFailed)
	}

	controlNodes := make([]*openlaneclient.GetAllControls_Controls_Edges_Node, 0, len(controls.Controls.Edges))
	for _, edge := range controls.Controls.Edges {
		controlNodes = append(controlNodes, edge.Node)
	}

	csvData, err := gocsv.MarshalBytes(&controlNodes)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal controls to CSV")
		return w.updateExportStatus(ctx, exportID, enums.ExportStatusFailed)
	}

	filename := fmt.Sprintf("controls_export_%s_%s.csv", exportID, time.Now().Format("20060102_150405"))
	reader := bytes.NewReader(csvData)

	upload := &graphql.Upload{
		File:        reader,
		Filename:    filename,
		Size:        int64(len(csvData)),
		ContentType: "text/csv",
	}

	updateInput := openlaneclient.UpdateExportInput{
		Status: &enums.ExportStatusReady,
	}

	_, err = w.olClient.UpdateExport(ctx, exportID, updateInput, []*graphql.Upload{upload})
	if err != nil {
		log.Error().Err(err).Msg("failed to update export with file")
		return w.updateExportStatus(ctx, exportID, enums.ExportStatusFailed)
	}

	return nil
}

func (w *ExportContentWorker) updateExportStatus(ctx context.Context, exportID string, status enums.ExportStatus) error {
	updateInput := openlaneclient.UpdateExportInput{
		Status: &status,
	}

	_, err := w.olClient.UpdateExport(ctx, exportID, updateInput, nil)
	if err != nil {
		log.Error().Err(err).
			Str("export_id", exportID).
			Str("status", string(status)).
			Msg("failed to update export status")
		return err
	}

	log.Info().
		Str("export_id", exportID).
		Str("status", string(status)).
		Msg("export status updated")

	return nil
}
