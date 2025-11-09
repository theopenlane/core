package corejobs

import (
	"context"
	"errors"
	"strings"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/corejobs/internal/pirsch"
)

var (
	// ErrMissingPirschDomainID is returned when a trust center does not have a pirsch domain ID
	ErrMissingPirschDomainID = errors.New("trust center does not have a pirsch domain ID")
)

// UpdatePirschDomainArgs for the worker to update the pirsch domain
type UpdatePirschDomainArgs struct {
	// TrustCenterID is the ID of the trust center to update the Pirsch domain for
	TrustCenterID string `json:"trust_center_id"`
}

// Kind satisfies the river.Job interface
func (UpdatePirschDomainArgs) Kind() string { return "update_pirsch_domain" }

// UpdatePirschDomainWorker updates a domain in Pirsch using the Pirsch API
type UpdatePirschDomainWorker struct {
	river.WorkerDefaults[UpdatePirschDomainArgs]

	Config PirschDomainConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for pirsch domain update"`

	pirschClient pirsch.Client
	olClient     olclient.OpenlaneClient
}

// WithPirschClient sets the Pirsch client for the worker
// and returns the worker for method chaining
func (w *UpdatePirschDomainWorker) WithPirschClient(cl pirsch.Client) *UpdatePirschDomainWorker {
	w.pirschClient = cl
	return w
}

// WithOpenlaneClient sets the Openlane client for the worker
// and returns the worker for method chaining
func (w *UpdatePirschDomainWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *UpdatePirschDomainWorker {
	w.olClient = cl
	return w
}

// Work satisfies the river.Worker interface for the update pirsch domain worker
// it updates a domain in Pirsch with the hostname and subdomain from the trust center's custom domain
func (w *UpdatePirschDomainWorker) Work(ctx context.Context, job *river.Job[UpdatePirschDomainArgs]) error {
	log.Debug().Str("trust_center_id", job.Args.TrustCenterID).Msg("updating pirsch domain")

	if job.Args.TrustCenterID == "" {
		return newMissingRequiredArg("trust_center_id", UpdatePirschDomainArgs{}.Kind())
	}

	if w.olClient == nil {
		cl, err := w.Config.getOpenlaneClient()
		if err != nil {
			return err
		}

		w.olClient = cl
	}

	if w.pirschClient == nil {
		w.pirschClient = pirsch.NewClient(w.Config.PirschClientID, w.Config.PirschClientSecret)
	}

	// Get the trust center
	trustCenter, err := w.olClient.GetTrustCenterByID(ctx, job.Args.TrustCenterID)
	if err != nil {
		return err
	}

	log.Debug().
		Str("trust_center_id", trustCenter.TrustCenter.ID).
		Str("owner_id", *trustCenter.TrustCenter.OwnerID).
		Msg("got trust center")

	// Check if trust center has a pirsch domain ID
	if trustCenter.TrustCenter.PirschDomainID == nil || *trustCenter.TrustCenter.PirschDomainID == "" {
		log.Error().
			Str("trust_center_id", job.Args.TrustCenterID).
			Msg("trust center does not have a pirsch domain ID")
		return ErrMissingPirschDomainID
	}

	pirschDomainID := *trustCenter.TrustCenter.PirschDomainID

	// Check if trust center has a custom domain
	if trustCenter.TrustCenter.CustomDomain == nil {
		log.Error().
			Str("trust_center_id", job.Args.TrustCenterID).
			Msg("trust center does not have a custom domain")
		return ErrTrustCenterNoCustomDomain
	}

	customDomainHostname := trustCenter.TrustCenter.CustomDomain.CnameRecord

	log.Debug().
		Str("custom_domain_hostname", customDomainHostname).
		Msg("got custom domain hostname")

	// Parse the hostname to extract subdomain and base hostname
	// e.g., "trust.example.com" -> subdomain: "trust", hostname: "example.com"
	parts := strings.SplitN(customDomainHostname, ".", hostnamePartCount)
	if len(parts) != hostnamePartCount {
		log.Error().
			Str("hostname", customDomainHostname).
			Msg("invalid hostname format")
		return ErrInvalidHostname
	}

	subdomain := parts[0]
	hostname := parts[1]

	log.Debug().
		Str("subdomain", subdomain).
		Str("hostname", hostname).
		Str("pirsch_domain_id", pirschDomainID).
		Msg("parsed domain information")

	// Update the hostname in Pirsch
	if err := w.pirschClient.UpdateHostname(ctx, pirschDomainID, hostname); err != nil {
		log.Error().
			Err(err).
			Str("pirsch_domain_id", pirschDomainID).
			Str("hostname", hostname).
			Msg("failed to update pirsch domain hostname")
		return err
	}

	log.Info().
		Str("pirsch_domain_id", pirschDomainID).
		Str("hostname", hostname).
		Msg("successfully updated pirsch domain hostname")

	// Update the subdomain in Pirsch
	if err := w.pirschClient.UpdateSubdomain(ctx, pirschDomainID, subdomain); err != nil {
		log.Error().
			Err(err).
			Str("pirsch_domain_id", pirschDomainID).
			Str("subdomain", subdomain).
			Msg("failed to update pirsch domain subdomain")
		return err
	}

	log.Info().
		Str("pirsch_domain_id", pirschDomainID).
		Str("subdomain", subdomain).
		Msg("successfully updated pirsch domain subdomain")

	log.Info().
		Str("trust_center_id", job.Args.TrustCenterID).
		Str("pirsch_domain_id", pirschDomainID).
		Msg("successfully updated pirsch domain")

	return nil
}
