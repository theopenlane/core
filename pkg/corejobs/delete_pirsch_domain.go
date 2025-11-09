package corejobs

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/corejobs/internal/pirsch"
)

// DeletePirschDomainArgs for the worker to delete a Pirsch domain
type DeletePirschDomainArgs struct {
	// PirschDomainID is the ID of the Pirsch domain to delete
	PirschDomainID string `json:"pirsch_domain_id"`
}

// Kind satisfies the river.Job interface
func (DeletePirschDomainArgs) Kind() string { return "delete_pirsch_domain" }

// DeletePirschDomainWorker deletes a domain from Pirsch
type DeletePirschDomainWorker struct {
	river.WorkerDefaults[DeletePirschDomainArgs]

	Config PirschDomainConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for pirsch domain deletion"`

	pirschClient pirsch.Client
}

// WithPirschClient sets the Pirsch client for the worker
// and returns the worker for method chaining
func (w *DeletePirschDomainWorker) WithPirschClient(cl pirsch.Client) *DeletePirschDomainWorker {
	w.pirschClient = cl
	return w
}

// Work satisfies the river.Worker interface for the delete pirsch domain worker
// it deletes a domain from Pirsch
func (w *DeletePirschDomainWorker) Work(ctx context.Context, job *river.Job[DeletePirschDomainArgs]) error {
	log.Info().
		Str("pirsch_domain_id", job.Args.PirschDomainID).
		Msg("deleting pirsch domain")

	if job.Args.PirschDomainID == "" {
		return newMissingRequiredArg("pirsch_domain_id", DeletePirschDomainArgs{}.Kind())
	}

	if w.pirschClient == nil {
		w.pirschClient = pirsch.NewClient(w.Config.PirschClientID, w.Config.PirschClientSecret)
	}

	// Delete the domain from Pirsch
	if err := w.pirschClient.DeleteDomain(ctx, job.Args.PirschDomainID); err != nil {
		log.Error().
			Err(err).
			Str("pirsch_domain_id", job.Args.PirschDomainID).
			Msg("failed to delete pirsch domain")
		return err
	}

	log.Info().
		Str("pirsch_domain_id", job.Args.PirschDomainID).
		Msg("successfully deleted pirsch domain")

	return nil
}
