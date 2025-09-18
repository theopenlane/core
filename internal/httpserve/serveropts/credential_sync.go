package serveropts

import (
	"context"
	"crypto/md5" //nolint:gosec // MD5 is used only for checksum comparison, not security
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// ErrNoActiveIntegration is returned when no active system integration is found for a provider
var ErrNoActiveIntegration = errors.New("no active system integration found")

// SystemOrganizationID is a special organization ID used for system integrations
// This should be set to a unique value for each test run
var SystemOrganizationID = "01101101011010010111010001100010"

// CredentialSyncService manages synchronization between config file credentials and database records
type CredentialSyncService struct {
	entClient     *generated.Client
	clientService *cp.ClientService[storage.Provider]
	config        *storage.ProviderConfigs
}

// NewCredentialSyncService creates a new credential synchronization service
func NewCredentialSyncService(entClient *generated.Client, clientService *cp.ClientService[storage.Provider], config *storage.ProviderConfigs) *CredentialSyncService {
	return &CredentialSyncService{
		entClient:     entClient,
		clientService: clientService,
		config:        config,
	}
}

// SyncConfigCredentials synchronizes config file credentials with database records on startup
func (css *CredentialSyncService) SyncConfigCredentials(ctx context.Context) error {
	providers := map[string]storage.ProviderCredentials{
		string(storage.S3Provider):   css.config.S3,
		string(storage.R2Provider):   css.config.CloudflareR2,
		string(storage.GCSProvider):  css.config.GCS,
		string(storage.DiskProvider): css.config.Disk,
	}

	for providerType, configCreds := range providers {
		if !configCreds.Enabled {
			continue
		}

		if err := css.syncProvider(ctx, providerType, configCreds); err != nil {
			return err
		}
	}

	return nil
}

// syncProvider synchronizes a single provider's credentials
func (css *CredentialSyncService) syncProvider(ctx context.Context, providerType string, configCreds storage.ProviderCredentials) error {
	// Get current active system integration for this provider using ent client
	integrations, err := css.entClient.Integration.Query().
		Where(
			integration.OwnerIDEQ(SystemOrganizationID),
			integration.KindEQ(providerType),
		).
		WithSecrets().
		All(ctx)
	if err != nil {
		return err
	}

	// Find active integration with non-expired credentials
	var activeInteg *generated.Integration
	for _, integ := range integrations {
		// Check if credentials match config
		if css.CredentialsMatch(integ.Edges.Secrets[0], configCreds) {
			zerolog.Ctx(ctx).Debug().Msgf("credentials already up to date for provider %s integration %s", providerType, integ.ID)
			return nil
		}
		// If not matched, keep track of the first integration (for rotation)
		if activeInteg == nil {
			activeInteg = integ
		}
	}

	// Create new integration + hush for updated config credentials
	newInteg, err := css.createSystemIntegration(ctx, providerType, configCreds)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("failed to create system integration for provider %s", providerType)
		return err
	}

	// If we had an existing integration, mark it as superseded (but don't expire yet)
	if activeInteg != nil {
		zerolog.Ctx(ctx).Info().Msgf("rotated system credentials for provider %s from integration %s to %s", providerType, activeInteg.ID, newInteg.ID)
	}

	return nil
}

// CredentialsMatch checks if the stored credentials match the config credentials
func (css *CredentialSyncService) CredentialsMatch(secret *generated.Hush, configCreds storage.ProviderCredentials) bool {
	// Generate hash of config credentials for comparison
	configHash := css.GenerateCredentialHash(configCreds)

	// Check if stored secret has the same hash
	storedHash := css.GenerateCredentialHashFromSet(secret.CredentialSet)
	return configHash == storedHash
}

// GenerateCredentialHash creates a hash of credential data for comparison
func (css *CredentialSyncService) GenerateCredentialHash(creds storage.ProviderCredentials) string {
	normalized := cp.StructToCredentials(creds)

	// Remove metadata fields that shouldn't affect credential matching
	delete(normalized, "region")
	delete(normalized, "bucket")
	delete(normalized, "enabled")
	delete(normalized, "credentials_json")

	data, _ := json.Marshal(normalized)
	hash := md5.Sum(data) //nolint:gosec // MD5 is used only for checksum comparison, not security
	return hex.EncodeToString(hash[:])
}

// GenerateCredentialHashFromSet creates a hash from a CredentialSet
func (css *CredentialSyncService) GenerateCredentialHashFromSet(credSet models.CredentialSet) string {
	normalized := cp.StructToCredentials(credSet)

	data, _ := json.Marshal(normalized)
	hash := md5.Sum(data) //nolint:gosec // MD5 is used only for checksum comparison, not security
	return hex.EncodeToString(hash[:])
}

// createSystemIntegration creates a new system integration and hush record
func (css *CredentialSyncService) createSystemIntegration(ctx context.Context, providerType string, configCreds storage.ProviderCredentials) (*generated.Integration, error) {
	credSet := models.CredentialSet{
		AccessKeyID:     configCreds.AccessKeyID,
		SecretAccessKey: configCreds.SecretAccessKey,
		Endpoint:        configCreds.Endpoint,
		ProjectID:       configCreds.ProjectID,
		AccountID:       configCreds.AccountID,
	}

	// Create metadata for the integration
	metadata := map[string]any{
		"region":          configCreds.Region,
		"bucket":          configCreds.Bucket,
		"source":          "system_config",
		"synchronized_at": time.Now(),
	}

	// Create hush record first
	hush, err := css.entClient.Hush.Create().
		SetName(fmt.Sprintf("%s_system_credentials", providerType)).
		SetMetadata(metadata).
		SetCredentialSet(credSet).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create hush record: %w", err)
	}

	// Create integration record
	integration, err := css.entClient.Integration.Create().
		SetName(fmt.Sprintf("System %s Storage", providerType)).
		SetDescription(fmt.Sprintf("System-level %s storage integration", providerType)).
		SetKind(providerType).
		SetIntegrationType("storage").
		SetOwnerID(SystemOrganizationID).
		SetMetadata(metadata).
		AddSecrets(hush).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration record: %w", err)
	}

	zerolog.Ctx(ctx).Info().Msgf("created system integration %s for config credentials for provider %s", integration.ID, providerType)

	return integration, nil
}

// GetActiveSystemProvider returns the active system provider for a given type
func (css *CredentialSyncService) GetActiveSystemProvider(ctx context.Context, providerType string) (*generated.Integration, error) {
	integrations, err := css.entClient.Integration.Query().
		Where(
			integration.OwnerIDEQ(SystemOrganizationID),
			integration.KindEQ(providerType),
		).
		WithSecrets().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Find the most recently created integration
	var newest *generated.Integration
	var newestTime time.Time

	for _, integ := range integrations {
		if syncTimeStr, ok := integ.Metadata["synchronized_at"].(string); ok {
			if syncTime, err := time.Parse(time.RFC3339, syncTimeStr); err == nil {
				if newest == nil || syncTime.After(newestTime) {
					newest = integ
					newestTime = syncTime
				}
			}
		}
	}

	if newest == nil {
		return nil, fmt.Errorf("%w for provider: %s", ErrNoActiveIntegration, providerType)
	}

	return newest, nil
}
