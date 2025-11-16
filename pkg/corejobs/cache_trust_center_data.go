package corejobs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/riverqueue/river"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

const (
	// trustCenterPageSize is the page size for paginating trust centers
	trustCenterPageSize = int64(50)
)

var (
	// ErrStorageServiceRequired is returned when the storage service is not configured
	ErrStorageServiceRequired = errors.New("storage service is required for caching trust center data")
)

// CacheTrustCenterDataArgs for the worker to process the custom domain
type CacheTrustCenterDataArgs struct {
	TrustCenterID string `json:"trust_center_id"`
}

type CacheTrustCenterDataConfig struct {
	OpenlaneConfig `koanf:",squash" jsonschema:"description=the openlane API configuration for caching trust center data"`
	// ObjectStorage contains the configuration for object storage
	ObjectStorage storage.ProviderConfig `koanf:"objectStorage" json:"objectStorage" jsonschema:"description=the object storage configuration"`
	// Enabled indicates whether the cache trust center data worker is enabled
	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the cache trust center data worker is enabled"`
	// CacheInterval is the interval at which to cache trust center data
	CacheInterval time.Duration `koanf:"cacheInterval" json:"cacheInterval" jsonschema:"required,default=10m description=the interval at which to cache trust center data"`
}

// Kind satisfies the river.Job interface
func (CacheTrustCenterDataArgs) Kind() string { return "cache_trust_center_data" }

// CacheTrustCenterDataWorker caches trust center data to object storage
type CacheTrustCenterDataWorker struct {
	river.WorkerDefaults[CacheTrustCenterDataArgs]

	Config CacheTrustCenterDataConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for caching trust center data"`

	olClient       olclient.OpenlaneClient
	storageService *objects.Service
}

// WithOpenlaneClient sets the Openlane client for the worker
// and returns the worker for method chaining
func (w *CacheTrustCenterDataWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *CacheTrustCenterDataWorker {
	w.olClient = cl
	return w
}

// WithStorageService sets the object storage service for the worker
// and returns the worker for method chaining
func (w *CacheTrustCenterDataWorker) WithStorageService(svc *objects.Service) *CacheTrustCenterDataWorker {
	w.storageService = svc
	return w
}

func (w *CacheTrustCenterDataWorker) Work(ctx context.Context, job *river.Job[CacheTrustCenterDataArgs]) error {
	log.Info().Msg("starting cache trust center data job")

	// Initialize Openlane client if not already set
	if w.olClient == nil {
		cl, err := w.Config.getOpenlaneClient()
		if err != nil {
			return err
		}
		w.olClient = cl
	}

	// Check if storage service is set
	if w.storageService == nil {
		return ErrStorageServiceRequired
	}

	// Paginate through all trust centers with GetTrustCentersCacheData query
	pageSize := trustCenterPageSize
	var after *string
	where := &openlaneclient.TrustCenterWhereInput{
		CustomDomainIDNotNil: lo.ToPtr(true),
	}
	if job.Args.TrustCenterID != "" {
		where.ID = &job.Args.TrustCenterID
	}

	for {
		resp, err := w.olClient.GetTrustCentersCacheData(ctx, &pageSize, nil, after, nil, where)
		if err != nil {
			return err
		}

		if resp == nil || resp.TrustCenters.Edges == nil {
			log.Info().Msg("no trust centers found")
			break
		}

		// Process each trust center
		for _, edge := range resp.TrustCenters.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}

			node := edge.Node

			// Skip if no custom domain
			if node.CustomDomain == nil || node.CustomDomain.CnameRecord == "" {
				continue
			}

			// Skip if no organization ID
			if node.OwnerID == nil || *node.OwnerID == "" {
				log.Warn().Str("trust_center_id", node.ID).Msg("trust center has no organization ID, skipping")
				continue
			}

			// Marshal the trust center settings to JSON
			settingsJSON, err := json.Marshal(node)
			if err != nil {
				log.Error().Err(err).Str("custom_domain", node.CustomDomain.CnameRecord).Msg("failed to marshal trust center settings")
				continue
			}

			// Upload to object storage with organization context
			if err := w.uploadTrustCenterData(ctx, w.storageService, node.CustomDomain.CnameRecord, *node.OwnerID, settingsJSON); err != nil {
				log.Error().Err(err).Str("custom_domain", node.CustomDomain.CnameRecord).Msg("failed to upload trust center data")
				continue
			}

			log.Debug().Str("custom_domain", node.CustomDomain.CnameRecord).Msg("successfully cached trust center data")
		}

		// Check if there are more pages
		if !resp.TrustCenters.PageInfo.HasNextPage {
			break
		}

		after = resp.TrustCenters.PageInfo.EndCursor
	}

	log.Info().Msg("completed cache trust center data job")
	return nil
}

// uploadTrustCenterData uploads the trust center settings JSON to object storage
func (w *CacheTrustCenterDataWorker) uploadTrustCenterData(ctx context.Context, svc *objects.Service, customDomain string, organizationID string, settingsJSON []byte) error {
	// Inject organization ID into context for storage provider resolution
	ctx = auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{
		OrganizationID: organizationID,
	})

	// Create a reader from the JSON bytes
	reader := bytes.NewReader(settingsJSON)

	// Upload options
	uploadOpts := &storage.UploadOptions{
		FileName:          "settings.json",
		ContentType:       "application/json",
		FolderDestination: customDomain,
		FileMetadata: storage.FileMetadata{
			Key: "trust_center_settings",
		},
	}

	// Upload the file
	_, err := svc.Upload(ctx, reader, uploadOpts)
	if err != nil {
		return fmt.Errorf("failed to upload trust center settings: %w", err)
	}

	return nil
}
