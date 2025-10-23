package corejobs

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/custom_hostnames"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"

	intcloudflare "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
)

// ValidateCustomDomainArgs for the worker to process the custom domain
type ValidateCustomDomainArgs struct {
	CustomDomainID string `json:"custom_domain_id"`
}

// Kind satisfies the river.Job interface
func (ValidateCustomDomainArgs) Kind() string { return "validate_custom_domain" }

// ValidateCustomDomainWorker checks cloudflare custom domain(s), and updates
// the status in our system
type ValidateCustomDomainWorker struct {
	river.WorkerDefaults[ValidateCustomDomainArgs]

	Config CustomDomainConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for custom domain validation"`

	cfClient intcloudflare.Client
	olClient olclient.OpenlaneClient
}

func (w *ValidateCustomDomainWorker) WithCloudflareClient(cl intcloudflare.Client) *ValidateCustomDomainWorker {
	w.cfClient = cl
	return w
}

func (w *ValidateCustomDomainWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *ValidateCustomDomainWorker {
	w.olClient = cl
	return w
}

// ValidateCustomDomainConfig contains the configuration for the worker
// Work satisfies the river.Worker interface for the validate custom domain worker.
// It validates custom domains by checking their status in Cloudflare and updating
// our system with the current verification status.
func (w *ValidateCustomDomainWorker) Work(ctx context.Context, job *river.Job[ValidateCustomDomainArgs]) error {
	// Initialize Cloudflare client if not already set
	if w.cfClient == nil {
		w.cfClient = intcloudflare.NewClient(w.Config.CloudflareAPIKey)
	}

	// Initialize Openlane client if not already set
	if w.olClient == nil {
		cl, err := w.Config.getOpenlaneClient()
		if err != nil {
			return err
		}

		w.olClient = cl
	}

	// Determine which custom domains to process - either a specific one or all
	var customDomains []*openlaneclient.CustomDomain

	if job.Args.CustomDomainID != "" {
		// If a specific custom domain ID is provided, fetch just that one
		customDomain, err := w.olClient.GetCustomDomainByID(ctx, job.Args.CustomDomainID)
		if err != nil {
			return err
		}

		cd := &openlaneclient.CustomDomain{
			ID:                customDomain.GetCustomDomain().ID,
			OwnerID:           customDomain.GetCustomDomain().OwnerID,
			CnameRecord:       customDomain.GetCustomDomain().CnameRecord,
			MappableDomainID:  customDomain.GetCustomDomain().MappableDomainID,
			DNSVerificationID: customDomain.GetCustomDomain().DNSVerificationID,
			MappableDomain: &openlaneclient.MappableDomain{
				ID:     customDomain.GetCustomDomain().MappableDomain.ID,
				Name:   customDomain.GetCustomDomain().MappableDomain.Name,
				ZoneID: customDomain.GetCustomDomain().MappableDomain.ZoneID,
			},
		}

		if customDomain.GetCustomDomain().DNSVerification != nil {
			cd.DNSVerification = &openlaneclient.DNSVerification{
				ID:                          customDomain.GetCustomDomain().DNSVerification.ID,
				CloudflareHostnameID:        customDomain.GetCustomDomain().DNSVerification.CloudflareHostnameID,
				DNSTxtRecord:                customDomain.GetCustomDomain().DNSVerification.DNSTxtRecord,
				DNSTxtValue:                 customDomain.GetCustomDomain().DNSVerification.DNSTxtValue,
				DNSVerificationStatus:       customDomain.GetCustomDomain().DNSVerification.DNSVerificationStatus,
				DNSVerificationStatusReason: customDomain.GetCustomDomain().DNSVerification.DNSVerificationStatusReason,
				AcmeChallengePath:           customDomain.GetCustomDomain().DNSVerification.AcmeChallengePath,
				ExpectedAcmeChallengeValue:  customDomain.GetCustomDomain().DNSVerification.ExpectedAcmeChallengeValue,
				AcmeChallengeStatus:         customDomain.GetCustomDomain().DNSVerification.AcmeChallengeStatus,
				AcmeChallengeStatusReason:   customDomain.GetCustomDomain().DNSVerification.AcmeChallengeStatusReason,
			}
		}

		customDomains = append(customDomains, cd)
	} else {
		// Otherwise, fetch all custom domains
		log.Debug().Msg("No custom domain ID provided, would fetch all domains here")

		cds, err := w.olClient.GetAllCustomDomains(ctx)
		if err != nil {
			return err
		}

		for _, cd := range cds.GetCustomDomains().Edges {
			customDomain := &openlaneclient.CustomDomain{
				ID:                cd.Node.ID,
				OwnerID:           cd.Node.OwnerID,
				CnameRecord:       cd.Node.CnameRecord,
				MappableDomainID:  cd.Node.MappableDomain.ID,
				DNSVerificationID: cd.Node.DNSVerificationID,
				MappableDomain: &openlaneclient.MappableDomain{
					ID:     cd.Node.MappableDomain.ID,
					Name:   cd.Node.MappableDomain.Name,
					ZoneID: cd.Node.MappableDomain.ZoneID,
				},
			}
			if cd.Node.DNSVerification != nil {
				customDomain.DNSVerification = &openlaneclient.DNSVerification{
					ID:                          cd.Node.DNSVerification.ID,
					CloudflareHostnameID:        cd.Node.DNSVerification.CloudflareHostnameID,
					DNSTxtRecord:                cd.Node.DNSVerification.DNSTxtRecord,
					DNSTxtValue:                 cd.Node.DNSVerification.DNSTxtValue,
					DNSVerificationStatus:       cd.Node.DNSVerification.DNSVerificationStatus,
					DNSVerificationStatusReason: cd.Node.DNSVerification.DNSVerificationStatusReason,
					AcmeChallengePath:           cd.Node.DNSVerification.AcmeChallengePath,
					ExpectedAcmeChallengeValue:  cd.Node.DNSVerification.ExpectedAcmeChallengeValue,
					AcmeChallengeStatus:         cd.Node.DNSVerification.AcmeChallengeStatus,
					AcmeChallengeStatusReason:   cd.Node.DNSVerification.AcmeChallengeStatusReason,
				}
			}

			customDomains = append(customDomains, customDomain)
		}
	}

	log.Info().Int("num_custom_domains", len(customDomains)).Msg("Processing custom domains")

	// Process each custom domain
	for _, customDomain := range customDomains {
		// Skip domains without verification IDs
		if customDomain.DNSVerification == nil {
			log.Debug().
				Str("custom_domain_id", customDomain.ID).
				Msg("No DNS verification ID found for custom domain")

			continue
		}

		dnsVerification := customDomain.DNSVerification

		log.Info().
			Str("custom_domain_id", customDomain.ID).
			Str("cname_record", customDomain.CnameRecord).
			Msg("Processing custom domain")

		mappableDomain := customDomain.MappableDomain
		if mappableDomain == nil {
			log.Error().
				Str("custom_domain_id", customDomain.ID).
				Msg("No mappable domain found for custom domain")

			continue
		}

		zoneID := mappableDomain.ZoneID
		cloudflareHostnameID := dnsVerification.CloudflareHostnameID

		// Fetch the current status from Cloudflare
		customHostname, err := w.cfClient.CustomHostnames().Get(ctx, cloudflareHostnameID, custom_hostnames.CustomHostnameGetParams{
			ZoneID: cloudflare.F(zoneID),
		})
		if err != nil {
			log.Error().Err(err).Msg("error getting custom hostname ID")

			continue
		}

		// Prepare updates based on Cloudflare's current status
		hasUpdates := false
		dnsVerificationUpdate := openlaneclient.UpdateDNSVerificationInput{}

		// Extract ACME challenge details if available and not already stored
		if dnsVerification.AcmeChallengePath == nil && len(customHostname.SSL.ValidationRecords) > 0 {
			acmeChallengeURL, err := url.Parse(customHostname.SSL.ValidationRecords[0].HTTPURL)
			if err != nil {
				log.Error().Err(err).Msg("Unable to parse acme challenge url")

				continue
			}

			spl := strings.Split(acmeChallengeURL.Path, "/")
			dnsVerificationUpdate.AcmeChallengePath = &spl[len(spl)-1]
			dnsVerificationUpdate.ExpectedAcmeChallengeValue = &customHostname.SSL.ValidationRecords[0].HTTPBody
			hasUpdates = true
		}

		// Update SSL verification status if changed
		if string(dnsVerification.AcmeChallengeStatus) != string(customHostname.SSL.Status) {
			dnsVerificationUpdate.AcmeChallengeStatus = lo.ToPtr(enums.SSLVerificationStatus(customHostname.SSL.Status))
			hasUpdates = true
		}

		// Update DNS verification status if changed
		if string(dnsVerification.DNSVerificationStatus) != string(customHostname.Status) {
			dnsVerificationUpdate.DNSVerificationStatus = lo.ToPtr(enums.DNSVerificationStatus(customHostname.Status))
			hasUpdates = true
		}

		// Update DNS verification error reasons if present
		if len(customHostname.VerificationErrors) > 0 {
			verifyErrors := strings.Join(customHostname.VerificationErrors, ", ")

			if dnsVerification.DNSVerificationStatusReason == nil || *dnsVerification.DNSVerificationStatusReason != verifyErrors {
				dnsVerificationUpdate.DNSVerificationStatusReason = &verifyErrors
				hasUpdates = true
			}
		}

		// Update SSL verification error reasons if present
		if len(customHostname.SSL.ValidationErrors) > 0 {
			verifyErrors := ""
			for _, validationErr := range customHostname.SSL.ValidationErrors {
				verifyErrors = fmt.Sprintf("%s, %s", verifyErrors, validationErr.Message)
			}

			if dnsVerification.AcmeChallengeStatusReason == nil || *dnsVerification.AcmeChallengeStatusReason != verifyErrors {
				dnsVerificationUpdate.AcmeChallengeStatusReason = &verifyErrors
				hasUpdates = true
			}
		}

		// Apply updates if any changes were detected
		if hasUpdates {
			log.Debug().Str("custom_domain_id", customDomain.ID).Msg("Updating DNS verification")

			_, err := w.olClient.UpdateDNSVerification(ctx, *customDomain.DNSVerificationID, dnsVerificationUpdate)
			if err != nil {
				log.Error().Err(err).Msg("error updating dns verification")
			}
		}
	}

	return nil
}
