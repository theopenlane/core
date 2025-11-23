package corejobs

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/riverqueue/river"
	intcloudflare "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
)

type DeletePreviewDomainArgs struct {
	// CustomDomainID is the ID of the custom domain to delete
	CustomDomainID string `json:"custom_domain_id"`
	// TrustCenterPreviewZoneID is the cloudflare zone id for the trust center preview domain
	TrustCenterPreviewZoneID string `json:"trust_center_preview_zone_id"`
}

func (DeletePreviewDomainArgs) Kind() string { return "delete_preview_domain" }

type DeletePreviewDomainConfig struct {
	// embed OpenlaneConfig to reuse validation and client creation logic
	OpenlaneConfig `koanf:",squash" jsonschema:"description=the openlane API configuration for preview domain deletion"`

	// Enabled indicates whether the preview domain worker is enabled
	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the preview domain worker is enabled"`

	// CloudflareAPIKey is the api key for cloudflare
	CloudflareAPIKey string `koanf:"cloudflareapikey" json:"cloudflareapikey" jsonschema:"required description=the cloudflare api key" sensitive:"true"`
}

type DeletePreviewDomainWorker struct {
	river.WorkerDefaults[DeletePreviewDomainArgs]

	Config DeletePreviewDomainConfig

	cfClient intcloudflare.Client
	olClient olclient.OpenlaneClient
}

func (w *DeletePreviewDomainWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *DeletePreviewDomainWorker {
	w.olClient = cl
	return w
}

func (w *DeletePreviewDomainWorker) WithCloudflareClient(cl intcloudflare.Client) *DeletePreviewDomainWorker {
	w.cfClient = cl
	return w
}

func (w *DeletePreviewDomainWorker) Work(ctx context.Context, job *river.Job[DeletePreviewDomainArgs]) error {
	if job.Args.CustomDomainID == "" {
		return newMissingRequiredArg("custom_domain_id", DeletePreviewDomainArgs{}.Kind())
	}
	if job.Args.TrustCenterPreviewZoneID == "" {
		return newMissingRequiredArg("trust_center_preview_zone_id", DeletePreviewDomainArgs{}.Kind())
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

	customDomain, err := w.olClient.GetCustomDomainByID(ctx, job.Args.CustomDomainID)
	if err != nil {
		return err
	}

	listRes, err := w.cfClient.Record().List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(job.Args.TrustCenterPreviewZoneID),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(customDomain.CustomDomain.CnameRecord),
		}),
	})
	if err != nil {
		return err
	}

	for _, record := range listRes.Result {
		if record.Type == dns.RecordResponseTypeCNAME {
			_, err := w.cfClient.Record().Delete(ctx, record.ID, dns.RecordDeleteParams{
				ZoneID: cloudflare.F(job.Args.TrustCenterPreviewZoneID),
			})
			if err != nil {
				return err
			}
		}
	}

	_, err = w.olClient.DeleteCustomDomain(ctx, job.Args.CustomDomainID)
	if err != nil {
		return err
	}

	return nil
}
