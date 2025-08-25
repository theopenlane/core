package corejobs

import (
	"context"
	"errors"

	"github.com/cloudflare/cloudflare-go/v5"
	"github.com/cloudflare/cloudflare-go/v5/custom_hostnames"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/theopenlane/riverboat/pkg/riverqueue"

	intcloudflare "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var (
	// ErrDomainVerificationAlreadyExists is returned when a custom domain already has a verification member
	ErrDomainVerificationAlreadyExists = errors.New("custom domain already has a verification member")
)

// CreateCustomDomainArgs for the worker to process the custom domain
type CreateCustomDomainArgs struct {
	// ID of the custom domain in our system
	CustomDomainID string `json:"custom_domain_id"`
}

// Kind satisfies the river.Job interface
func (CreateCustomDomainArgs) Kind() string { return "create_custom_domain" }

// CreateCustomDomainWorker creates a custom hostname in cloudflare, and
// creates and updates the records in our system
type CreateCustomDomainWorker struct {
	river.WorkerDefaults[CreateCustomDomainArgs]

	Config CustomDomainConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for custom domain creation"`

	cfClient    intcloudflare.Client
	olClient    olclient.OpenlaneClient
	riverClient riverqueue.JobClient
}

// WithCloudflareClient sets the Cloudflare client for the worker
// and returns the worker for method chaining
func (w *CreateCustomDomainWorker) WithCloudflareClient(cl intcloudflare.Client) *CreateCustomDomainWorker {
	w.cfClient = cl
	return w
}

// WithOpenlaneClient sets the Openlane client for the worker
// and returns the worker for method chaining
func (w *CreateCustomDomainWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *CreateCustomDomainWorker {
	w.olClient = cl
	return w
}

// WithRiverClient sets the River client for the worker
// and returns the worker for method chaining
func (w *CreateCustomDomainWorker) WithRiverClient(cl riverqueue.JobClient) *CreateCustomDomainWorker {
	w.riverClient = cl
	return w
}

// Work satisfies the river.Worker interface for the create custom domain worker
// it creates a custom domain for an organization
func (w *CreateCustomDomainWorker) Work(ctx context.Context, job *river.Job[CreateCustomDomainArgs]) error {
	log.Debug().Str("custom_domain_id", job.Args.CustomDomainID).Msg("creating custom domain")

	if job.Args.CustomDomainID == "" {
		return newMissingRequiredArg("custom_domain_id", CreateCustomDomainArgs{}.Kind())
	}

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

	if w.riverClient == nil {
		riverClient, err := riverqueue.New(ctx, riverqueue.WithConnectionURI(w.Config.DatabaseHost))
		if err != nil {
			return err
		}

		w.riverClient = riverClient
	}

	// get the custom domain
	customDomain, err := w.olClient.GetCustomDomainByID(ctx, job.Args.CustomDomainID)
	if err != nil {
		return err
	}

	log.Debug().Str("custom_domain", customDomain.GetCustomDomain().ID).Msg("got custom domain")

	if customDomain.CustomDomain.DNSVerificationID != nil && *customDomain.CustomDomain.DNSVerificationID != "" {
		return ErrDomainVerificationAlreadyExists
	}

	mappableDomain, err := w.olClient.GetMappableDomainByID(ctx, customDomain.CustomDomain.MappableDomainID)
	if err != nil {
		return err
	}

	log.Debug().Str("mappable_domain", mappableDomain.GetMappableDomain().ID).Msg("got mappable domain")

	res, err := w.cfClient.CustomHostnames().New(ctx, custom_hostnames.CustomHostnameNewParams{
		ZoneID:   cloudflare.F(mappableDomain.MappableDomain.ZoneID),
		Hostname: cloudflare.F(customDomain.CustomDomain.CnameRecord),
		SSL: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSL{
			Method: cloudflare.F(custom_hostnames.DCVMethodHTTP),
			Settings: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSLSettings{
				MinTLSVersion: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSLSettingsMinTLSVersion1_0),
			}),
			Type: cloudflare.F(custom_hostnames.DomainValidationTypeDv),
		}),
	})
	if err != nil {
		return err
	}

	log.Debug().Str("cloudflare_id", res.ID).Msg("created custom hostname")

	dnsVerificationID := ""

	defer func() {
		if err != nil {
			_, insertErr := w.riverClient.Insert(ctx, DeleteCustomDomainArgs{
				CloudflareCustomHostnameID: res.ID,
				CloudflareZoneID:           mappableDomain.MappableDomain.ZoneID,
				DNSVerificationID:          dnsVerificationID,
			}, nil)
			if insertErr != nil {
				log.Error().Err(insertErr).Msg("error inserting delete_cloudflare_custom_hostname job")
			}
		}
	}()

	ownerVerifyTXT := res.OwnershipVerification.Name
	ownerVerifyValue := res.OwnershipVerification.Value

	createVerificationRes, err := w.olClient.CreateDNSVerification(ctx, openlaneclient.CreateDNSVerificationInput{
		CloudflareHostnameID: res.ID,
		DNSTxtRecord:         ownerVerifyTXT,
		DNSTxtValue:          ownerVerifyValue,
		OwnerID:              customDomain.CustomDomain.OwnerID,
	})
	if err != nil {
		return err
	}

	log.Debug().Str("dns_verification", createVerificationRes.GetCreateDNSVerification().DNSVerification.ID).Msg("created dns verification")

	dnsVerificationID = createVerificationRes.CreateDNSVerification.DNSVerification.ID

	_, err = w.olClient.UpdateCustomDomain(ctx, job.Args.CustomDomainID, openlaneclient.UpdateCustomDomainInput{
		DNSVerificationID: lo.ToPtr(createVerificationRes.CreateDNSVerification.DNSVerification.ID),
	})
	if err != nil {
		return err
	}

	log.Info().Str(
		"verification_id", createVerificationRes.CreateDNSVerification.DNSVerification.ID,
	).Str(
		"cloudflare_hostname_id", res.ID,
	).Str(
		"custom_domain_id", job.Args.CustomDomainID,
	).Msg("Successfully created and associated custom hostname")

	return nil
}
