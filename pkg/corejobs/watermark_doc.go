package corejobs

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog"

	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
)

// WatermarkDocArgs for the worker to process watermarking of a document
type WatermarkDocArgs struct {
	// ID of the trust center document
	TrustCenterDocumentID string `json:"trust_center_document_id"`
}

// Kind satisfies the river.Job interface
func (WatermarkDocArgs) Kind() string { return "watermark_doc" }

type WatermarkWorkerConfig struct {
	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the watermark worker is enabled"`

	OpenlaneAPIHost  string `koanf:"openlaneAPIHost" json:"openlaneAPIHost" jsonschema:"required description=the openlane api host"`
	OpenlaneAPIToken string `koanf:"openlaneAPIToken" json:"openlaneAPIToken" jsonschema:"required description=the openlane api token"`
}

// WatermarkDocWorker is the worker to process watermarking of a document
type WatermarkDocWorker struct {
	river.WorkerDefaults[WatermarkDocArgs]

	Config WatermarkWorkerConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for watermarking"`

	olClient olclient.OpenlaneClient
}

// WithOpenlaneClient sets the openlane client to use for API requests
func (w *WatermarkDocWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *WatermarkDocWorker {
	w.olClient = cl
	return w
}

// Work satisfies the river.Worker interface for the watermark doc worker
func (w *WatermarkDocWorker) Work(ctx context.Context, job *river.Job[WatermarkDocArgs]) error {
	// TODO: implement watermarking
	zerolog.Ctx(ctx).Info().Str("trust_center_document_id", job.Args.TrustCenterDocumentID).Msg("would watermark document")

	return nil
}
