package corejobs

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
	intcloudflare "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/keygen"
)

// CreatePreviewDomainArgs for the worker to process the preview domain creation
type CreatePreviewDomainArgs struct {
	// TrustCenterID is the ID of the trust center to create a preview domain for
	TrustCenterID string `json:"trust_center_id"`
	// TrustCenterPreviewZoneID is the cloudflare zone id for the trust center preview domain
	TrustCenterPreviewZoneID string `json:"trust_center_preview_zone_id"`
	// TrustCenterCnameTarget is the cname target for the trust center
	TrustCenterCnameTarget string `json:"trust_center_cname_target"`
}

// Kind satisfies the river.Job interface
func (CreatePreviewDomainArgs) Kind() string { return "create_preview_domain" }

type PreviewDomainConfig struct {
	// embed OpenlaneConfig to reuse validation and client creation logic
	OpenlaneConfig `koanf:",squash" jsonschema:"description=the openlane API configuration for pirsch domain management"`

	// Enabled indicates whether the preview domain worker is enabled
	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the preview domain worker is enabled"`

	// CloudflareAPIKey is the api key for cloudflare
	CloudflareAPIKey string `koanf:"cloudflareapikey" json:"cloudflareapikey" jsonschema:"required description=the cloudflare api key" sensitive:"true"`
}

type CreatePreviewDomainWorker struct {
	river.WorkerDefaults[CreatePreviewDomainArgs]

	Config PreviewDomainConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for preview domain creation"`

	cfClient intcloudflare.Client
	olClient olclient.OpenlaneClient
}

func (w *CreatePreviewDomainWorker) WithCloudflareClient(cl intcloudflare.Client) *CreatePreviewDomainWorker {
	w.cfClient = cl
	return w
}

func (w *CreatePreviewDomainWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *CreatePreviewDomainWorker {
	w.olClient = cl
	return w
}

func (w *CreatePreviewDomainWorker) Work(ctx context.Context, job *river.Job[CreatePreviewDomainArgs]) error {
	if job.Args.TrustCenterID == "" {
		return newMissingRequiredArg("trust_center_id", CreatePreviewDomainArgs{}.Kind())
	}

	if job.Args.TrustCenterPreviewZoneID == "" {
		return newMissingRequiredArg("trust_center_preview_zone_id", CreatePreviewDomainArgs{}.Kind())
	}

	if job.Args.TrustCenterCnameTarget == "" {
		return newMissingRequiredArg("trust_center_cname_target", CreatePreviewDomainArgs{}.Kind())
	}

	if w.cfClient == nil {
		w.cfClient = intcloudflare.NewClient(w.Config.CloudflareAPIKey)
	}

	if w.olClient == nil {
		cl, err := w.Config.getOpenlaneClient()
		if err != nil {
			return err
		}

		w.olClient = cl
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

	zone, err := w.cfClient.Zones().Get(ctx, zones.ZoneGetParams{
		ZoneID: cloudflare.F(job.Args.TrustCenterPreviewZoneID),
	})
	if err != nil {
		return err
	}

	log.Debug().
		Str("zone_id", zone.ID).
		Str("zone_name", zone.Name).
		Msg("got zone")

	mappableDomain, err := w.olClient.GetMappableDomains(ctx, nil, nil, &openlaneclient.MappableDomainWhereInput{
		Name: &job.Args.TrustCenterCnameTarget,
	})
	if err != nil {
		return err
	}

	if len(mappableDomain.MappableDomains.Edges) == 0 {
		return fmt.Errorf("no mappable domain found for %s", job.Args.TrustCenterCnameTarget)
	}

	log.Debug().
		Str("mappable_domain_id", mappableDomain.MappableDomains.Edges[0].Node.ID).
		Msg("got mappable domain")

	previewDomain := createPreviewDomain(ctx, zone.Name, *trustCenter.TrustCenter.Slug)

	customDomain, err := w.olClient.CreateCustomDomain(ctx, openlaneclient.CreateCustomDomainInput{
		CnameRecord:      previewDomain,
		MappableDomainID: mappableDomain.MappableDomains.Edges[0].Node.ID,
		OwnerID:          trustCenter.TrustCenter.OwnerID,
	})
	if err != nil {
		return err
	}
	log.Debug().
		Str("custom_domain_id", customDomain.CreateCustomDomain.CustomDomain.ID).
		Msg("created custom domain")

	_, err = w.olClient.UpdateTrustCenter(ctx, job.Args.TrustCenterID, openlaneclient.UpdateTrustCenterInput{
		PreviewDomainID: &customDomain.CreateCustomDomain.CustomDomain.ID,
		PreviewStatus:   &enums.TrustCenterPreviewStatusProvisioning,
	})
	if err != nil {
		return err
	}

	record, err := w.cfClient.Record().New(ctx, dns.RecordNewParams{
		ZoneID: cloudflare.F(job.Args.TrustCenterPreviewZoneID),
		Body: dns.CNAMERecordParam{
			Name:    cloudflare.F(previewDomain),
			Content: cloudflare.F(job.Args.TrustCenterCnameTarget),
			Type:    cloudflare.F(dns.CNAMERecordTypeCNAME),
		},
	})
	if err != nil {
		return err
	}

	log.Debug().
		Str("record_id", record.ID).
		Str("record_name", record.Name).
		Str("record_target", record.Content).
		Msg("created record")

	return nil
}

// create domain like $slug-$randomstring.$zoneName
func createPreviewDomain(ctx context.Context, zoneName string, trustCenterSlug string) string {
	randomString := keygen.AlphaNumeric(9)
	return fmt.Sprintf("%s-%s.%s", strings.ToLower(trustCenterSlug), strings.ToLower(randomString), zoneName)
}
