package corejobs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	intcloudflare "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var (
	ErrMaxSnoozesReached = errors.New("max snoozes reached")
)

type ValidatePreviewDomainArgs struct {
	// TrustCenterID is the ID of the trust center to validate the preview domain for
	TrustCenterID string `json:"trust_center_id"`
	// TrustCenterPreviewZoneID is the cloudflare zone id for the trust center preview domain
	TrustCenterPreviewZoneID string `json:"trust_center_preview_zone_id"`
}

type ValidatePreviewDomainConfig struct {
	// embed OpenlaneConfig to reuse validation and client creation logic
	OpenlaneConfig `koanf:",squash" jsonschema:"description=the openlane API configuration for preview domain validation"`

	// Enabled indicates whether the preview domain worker is enabled
	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the preview domain worker is enabled"`

	// CloudflareAPIKey is the api key for cloudflare
	CloudflareAPIKey string `koanf:"cloudflareapikey" json:"cloudflareapikey" jsonschema:"required description=the cloudflare api key" sensitive:"true"`

	// MaxSnoozes is the maximum number of times to snooze the job before giving up
	MaxSnoozes int `koanf:"maxsnoozes" json:"maxsnoozes" jsonschema:"required,default=30 description=the maximum number of times to snooze the job before giving up"`

	// SnoozeDuration is the duration to snooze the job for
	SnoozeDuration time.Duration `koanf:"snoozeduration" json:"snoozeduration" jsonschema:"required,default=5s description=the duration to snooze the job for"`
}

// Kind satisfies the river.Job interface
func (ValidatePreviewDomainArgs) Kind() string { return "validate_preview_domain" }

type ValidatePreviewDomainWorker struct {
	river.WorkerDefaults[ValidatePreviewDomainArgs]

	Config ValidatePreviewDomainConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for preview domain validation"`

	cfClient intcloudflare.Client
	olClient olclient.OpenlaneClient
}

func (w *ValidatePreviewDomainWorker) WithCloudflareClient(cl intcloudflare.Client) *ValidatePreviewDomainWorker {
	w.cfClient = cl
	return w
}

func (w *ValidatePreviewDomainWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *ValidatePreviewDomainWorker {
	w.olClient = cl
	return w
}

// ValidatePreviewDomainWorker checks preview domain, creates TXT validation records, and updates the preview status
func (w *ValidatePreviewDomainWorker) Work(ctx context.Context, job *river.Job[ValidatePreviewDomainArgs]) error {
	if job.Args.TrustCenterID == "" {
		return newMissingRequiredArg("trust_center_id", CreatePreviewDomainArgs{}.Kind())
	}

	if job.Args.TrustCenterPreviewZoneID == "" {
		return newMissingRequiredArg("trust_center_preview_zone_id", CreatePreviewDomainArgs{}.Kind())
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

	trustCenter, err := w.olClient.GetTrustCenterByID(ctx, job.Args.TrustCenterID)
	if err != nil {
		return err
	}

	if trustCenter.TrustCenter.PreviewDomain == nil {
		return newMissingRequiredArg("preview_domain", CreatePreviewDomainArgs{}.Kind())
	}

	snoozes := struct {
		Snoozes int `json:"snoozes"`
	}{}
	if err = json.Unmarshal(job.Metadata, &snoozes); err != nil {
		snoozes.Snoozes = 0
	}

	dnsVerification := trustCenter.TrustCenter.PreviewDomain.DNSVerification
	if dnsVerification == nil {
		if snoozes.Snoozes >= w.Config.MaxSnoozes {
			return ErrMaxSnoozesReached
		}
		// Snooze job
		return river.JobSnooze(w.Config.SnoozeDuration)
	}

	listRes, err := w.cfClient.Record().List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(job.Args.TrustCenterPreviewZoneID),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(dnsVerification.DNSTxtRecord),
		}),
	})
	if err != nil {
		return err
	}

	// Check if the TXT record exists
	txtRecordExists := false
	for _, record := range listRes.Result {
		// Check if this is a TXT record with the expected value
		if string(record.Type) == "TXT" && record.Content == dnsVerification.DNSTxtValue {
			txtRecordExists = true
			break
		}
	}

	// If TXT record doesn't exist, create it
	if !txtRecordExists {
		// create the txt record
		txtRecord, err := w.cfClient.Record().New(ctx, dns.RecordNewParams{
			ZoneID: cloudflare.F(job.Args.TrustCenterPreviewZoneID),
			Body: dns.TXTRecordParam{
				Name:    cloudflare.F(dnsVerification.DNSTxtRecord),
				Content: cloudflare.F(dnsVerification.DNSTxtValue),
				Type:    cloudflare.F(dns.TXTRecordTypeTXT),
			},
		})
		if err != nil {
			return err
		}

		log.Debug().
			Str("record_id", txtRecord.ID).
			Str("record_name", txtRecord.Name).
			Str("record_content", txtRecord.Content).
			Msg("created txt record")

		return nil
	}

	// check dns verification status, if good, update preview status on trust center to "ready"
	if dnsVerification.DNSVerificationStatus == enums.DNSVerificationStatusActive &&
		dnsVerification.AcmeChallengeStatus == enums.SSLVerificationStatusActive {

		_, err := w.olClient.UpdateTrustCenter(ctx, job.Args.TrustCenterID, openlaneclient.UpdateTrustCenterInput{
			PreviewStatus: lo.ToPtr(enums.TrustCenterPreviewStatusReady),
		})
		if err != nil {
			return err
		}
		log.Info().Str("trust_center_id", job.Args.TrustCenterID).Msg("updated trust center preview status to ready")

		return nil
	}

	if snoozes.Snoozes >= w.Config.MaxSnoozes {
		return ErrMaxSnoozesReached
	}
	// Snooze job
	return river.JobSnooze(w.Config.SnoozeDuration)
}
