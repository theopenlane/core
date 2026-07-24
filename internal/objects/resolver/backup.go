package resolver

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/objects/storage"
)

// buildBackupProviders constructs a backup provider instance for every top-level provider that declares an enabled Backup target.
// Providers without a configured backup are absent from the map
func buildBackupProviders(config storage.ProviderConfig, builders providerBuilders, runtime serviceOptions) (map[storage.ProviderType]storage.Provider, error) {
	sources := map[storage.ProviderType]storage.ProviderConfigs{
		storage.S3Provider:       config.Providers.S3,
		storage.R2Provider:       config.Providers.R2,
		storage.DiskProvider:     config.Providers.Disk,
		storage.DatabaseProvider: config.Providers.Database,
	}

	backups := map[storage.ProviderType]storage.Provider{}

	for source, sourceCfg := range sources {
		backupCfg := sourceCfg.Backup
		// only build a backup when the source provider and the backup target are both enabled
		if !sourceCfg.Enabled || backupCfg == nil || !backupCfg.Enabled {
			continue
		}

		provider, err := buildBackupProvider(*backupCfg, builders, runtime)
		if err != nil {
			if backupCfg.EnsureAvailable {
				return nil, fmt.Errorf("building backup provider for %s: %w", source, err)
			}

			log.Warn().Err(err).
				Str("source", string(source)).
				Str("destination", string(backupCfg.Provider)).
				Msg("backup provider unavailable; source will proceed without a backup")

			continue
		}

		backups[source] = provider
	}

	return backups, nil
}

// buildBackupProvider constructs a single backup destination provider from its configuration
func buildBackupProvider(backupCfg storage.BackupConfig, builders providerBuilders, runtime serviceOptions) (storage.Provider, error) {
	builder := builderFor(builders, backupCfg.Provider)
	if builder == nil {
		return nil, fmt.Errorf("%w: %s", errUnsupportedProvider, backupCfg.Provider)
	}

	options, creds, err := providerOptionsFromProviderConfig(backupCfg.Provider, backupCfg.ProviderConfigs(), runtime)
	if err != nil {
		return nil, err
	}

	return builder.Build(context.Background(), creds, options)
}
