package corejobs

import (
	"context"
	"errors"
	"strings"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/corejobs/internal/pirsch"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

const (
	// hostnamePartCount is the expected number of parts when splitting a hostname (subdomain + domain)
	hostnamePartCount = 2
	// defaultActiveVisitorsSeconds is the default time window for active visitors tracking in Pirsch
	defaultActiveVisitorsSeconds = 300
)

var (
	// ErrTrustCenterNoCustomDomain is returned when a trust center does not have an associated custom domain
	ErrTrustCenterNoCustomDomain = errors.New("trust center does not have a custom domain")
	// ErrInvalidHostname is returned when the custom domain hostname is invalid
	ErrInvalidHostname = errors.New("invalid custom domain hostname")
)

// CreatePirschDomainArgs for the worker to process the pirsch domain creation
type CreatePirschDomainArgs struct {
	// TrustCenterID is the ID of the trust center to create a Pirsch domain for
	TrustCenterID string `json:"trust_center_id"`
}

// Kind satisfies the river.Job interface
func (CreatePirschDomainArgs) Kind() string { return "create_pirsch_domain" }

// CreatePirschDomainWorker creates a domain in Pirsch using the Pirsch API
type CreatePirschDomainWorker struct {
	river.WorkerDefaults[CreatePirschDomainArgs]

	Config PirschDomainConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for pirsch domain creation"`

	pirschClient pirsch.Client
	olClient     olclient.OpenlaneClient
	riverClient  riverqueue.JobClient
}

// WithPirschClient sets the Pirsch client for the worker
// and returns the worker for method chaining
func (w *CreatePirschDomainWorker) WithPirschClient(cl pirsch.Client) *CreatePirschDomainWorker {
	w.pirschClient = cl
	return w
}

// WithOpenlaneClient sets the Openlane client for the worker
// and returns the worker for method chaining
func (w *CreatePirschDomainWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *CreatePirschDomainWorker {
	w.olClient = cl
	return w
}

// WithRiverClient sets the River client for the worker
// and returns the worker for method chaining
func (w *CreatePirschDomainWorker) WithRiverClient(cl riverqueue.JobClient) *CreatePirschDomainWorker {
	w.riverClient = cl
	return w
}

// Work satisfies the river.Worker interface for the create pirsch domain worker
// it creates a domain in Pirsch for analytics tracking
func (w *CreatePirschDomainWorker) Work(ctx context.Context, job *river.Job[CreatePirschDomainArgs]) error {
	log.Debug().Str("trust_center_id", job.Args.TrustCenterID).Msg("creating pirsch domain")

	if job.Args.TrustCenterID == "" {
		return newMissingRequiredArg("trust_center_id", CreatePirschDomainArgs{}.Kind())
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

	if w.riverClient == nil {
		riverClient, err := riverqueue.New(ctx, riverqueue.WithConnectionURI(w.Config.DatabaseHost))
		if err != nil {
			return err
		}

		w.riverClient = riverClient
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

	// Get the organization to use its display name
	org, err := w.olClient.GetOrganizationByID(ctx, *trustCenter.TrustCenter.OwnerID)
	if err != nil {
		return err
	}

	displayName := org.Organization.DisplayName

	log.Debug().
		Str("organization_name", displayName).
		Str("subdomain", subdomain).
		Str("hostname", hostname).
		Msg("parsed domain information")

	// Create the domain in Pirsch
	domain, err := w.pirschClient.CreateDomain(ctx, pirsch.CreateDomainRequest{
		Hostname:                    hostname,
		Subdomain:                   subdomain,
		Timezone:                    "UTC",
		DisplayName:                 displayName,
		Public:                      false,
		GroupByTitle:                false,
		ActiveVisitorsSeconds:       defaultActiveVisitorsSeconds,
		DisableScripts:              false,
		TrafficSpikeThreshold:       0,
		TrafficWarningThresholdDays: 0,
	})
	if err != nil {
		return err
	}

	log.Info().
		Str("pirsch_domain_id", domain.ID).
		Str("hostname", domain.Hostname).
		Str("subdomain", domain.Subdomain).
		Str("trust_center_id", job.Args.TrustCenterID).
		Msg("Successfully created Pirsch domain")

	if _, err = w.olClient.UpdateTrustCenter(
		ctx, trustCenter.TrustCenter.ID,
		openlaneclient.UpdateTrustCenterInput{
			PirschDomainID:           &domain.ID,
			PirschIdentificationCode: &domain.IdentificationCode,
		},
	); err != nil {
		log.Error().
			Err(err).
			Str("trust_center_id", job.Args.TrustCenterID).
			Str("pirsch_domain_id", domain.ID).
			Msg("failed to update trust center with pirsch domain information, attempting cleanup")

		// Insert a delete job to reliably clean up the pirsch domain
		_, insertErr := w.riverClient.Insert(ctx, DeletePirschDomainArgs{
			PirschDomainID: domain.ID,
		}, nil)
		if insertErr != nil {
			log.Error().
				Err(insertErr).
				Str("pirsch_domain_id", domain.ID).
				Msg("failed to insert delete_pirsch_domain job for cleanup")
		} else {
			log.Info().
				Str("pirsch_domain_id", domain.ID).
				Msg("successfully inserted delete_pirsch_domain job for cleanup")
		}

		return err
	}

	log.Info().
		Str("trust_center_id", job.Args.TrustCenterID).
		Str("pirsch_domain_id", domain.ID).
		Msg("successfully updated trust center with pirsch domain information")

	return nil
}
