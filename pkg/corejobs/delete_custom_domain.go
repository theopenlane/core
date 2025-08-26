package corejobs

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/cloudflare/cloudflare-go/v5"
	"github.com/cloudflare/cloudflare-go/v5/custom_hostnames"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"

	intcloudflare "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
)

// DeleteCustomDomainArgs for the worker to process the custom domain
type DeleteCustomDomainArgs struct {
	// ID of the custom domain in our system
	CustomDomainID             string `json:"custom_domain_id"`
	DNSVerificationID          string `json:"dns_verification_id"`
	CloudflareCustomHostnameID string `json:"cloudflare_custom_hostname_id"`
	CloudflareZoneID           string `json:"cloudflare_zone_id"`
}

// Kind satisfies the river.Job interface
func (DeleteCustomDomainArgs) Kind() string { return "delete_custom_domain" }

// DeleteCustomDomainWorker delete the custom hostname from cloudflare and
// updates the records in our system
type DeleteCustomDomainWorker struct {
	river.WorkerDefaults[DeleteCustomDomainArgs]

	Config CustomDomainConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for custom domain deletion"`

	cfClient intcloudflare.Client
	olClient olclient.OpenlaneClient
}

// WithCloudflareClient sets the Cloudflare client for the worker
// and returns the worker for method chaining
func (w *DeleteCustomDomainWorker) WithCloudflareClient(cl intcloudflare.Client) *DeleteCustomDomainWorker {
	w.cfClient = cl
	return w
}

// WithOpenlaneClient sets the Openlane client for the worker
// and returns the worker for method chaining
func (w *DeleteCustomDomainWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *DeleteCustomDomainWorker {
	w.olClient = cl
	return w
}

// Work satisfies the river.Worker interface for the delete custom domain worker
// it deletes a custom domain for an organization
func (w *DeleteCustomDomainWorker) Work(ctx context.Context, job *river.Job[DeleteCustomDomainArgs]) error {
	log.Info().Str(
		"custom_domain_id", job.Args.CustomDomainID,
	).Str(
		"dns_verification_id", job.Args.DNSVerificationID,
	).Str(
		"cloudflare_custom_hostname_id", job.Args.CloudflareCustomHostnameID,
	).Str(
		"cloudflare_zone_id", job.Args.CloudflareZoneID,
	).Msg("deleting custom domain")

	if w.cfClient == nil {
		w.cfClient = intcloudflare.NewClient(w.Config.CloudflareAPIKey)
	}

	if w.olClient == nil {
		cl, err := getOpenlaneClient(w.Config)
		if err != nil {
			return err
		}

		w.olClient = cl
	}

	if job.Args.CloudflareCustomHostnameID != "" {
		_, err := w.cfClient.CustomHostnames().Delete(ctx, job.Args.CloudflareCustomHostnameID, custom_hostnames.CustomHostnameDeleteParams{
			ZoneID: cloudflare.F(job.Args.CloudflareZoneID),
		})
		if err != nil {
			var apierr *cloudflare.Error
			if errors.As(err, &apierr) {
				if apierr.StatusCode != http.StatusNotFound {
					return err
				}
			} else {
				return err
			}
		}
	}

	if job.Args.DNSVerificationID != "" {
		_, err := w.olClient.DeleteDNSVerification(ctx, job.Args.DNSVerificationID)
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return err
			}
		}
	}

	if job.Args.CustomDomainID != "" {
		_, err := w.olClient.DeleteCustomDomain(ctx, job.Args.CustomDomainID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return nil
			}

			return err
		}
	}

	return nil
}
