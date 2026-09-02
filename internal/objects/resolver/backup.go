package resolver

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/v2/internal/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// backupDestination is the resolved destination for one source provider's backup
type backupDestination struct {
	// Provider is the destination backend type
	Provider storage.ProviderType
	// Config is the destination provider configuration the backup writes with
	Config storage.ProviderConfigs
	// ReadFromBackup serves reads from this destination instead of the source
	ReadFromBackup bool
}

// backupDestinations resolves the backup destination for every source provider that declares an
// enabled backup target. Sources without a configured backup are absent from the map
func backupDestinations(config storage.ProviderConfig) (map[storage.ProviderType]backupDestination, error) {
	sources := config.Providers.ByType()

	destinations := map[storage.ProviderType]backupDestination{}

	for source, sourceCfg := range sources {
		destination, hasBackup := sourceCfg.BackupDestination(source)
		if !hasBackup {
			continue
		}

		backupCfg := sourceCfg.Backup

		destinationCfg, ok := sources[destination]
		if !ok {
			return nil, fmt.Errorf("%w: %s", errUnsupportedProvider, destination)
		}

		if !destinationCfg.Enabled {
			log.Warn().Str("source", string(source)).Str("destination", string(destination)).Msg("backup destination provider is not enabled; source will proceed without a backup")

			continue
		}

		destinationCfg.Bucket = storage.BackupBucket(destinationCfg.Bucket)

		if backupCfg.Region != "" {
			destinationCfg.Region = backupCfg.Region
		}

		destinations[source] = backupDestination{
			Provider:       destination,
			Config:         destinationCfg,
			ReadFromBackup: backupCfg.ReadFromBackup,
		}
	}

	return destinations, nil
}

// buildBackupProviders resolves a backup provider for every source that declares an enabled Backup target
func buildBackupProviders(resolver *providerResolver, config storage.ProviderConfig) (map[storage.ProviderType]objects.BackupTarget, error) {
	destinations, err := backupDestinations(config)
	if err != nil {
		return nil, err
	}

	backups := map[storage.ProviderType]objects.BackupTarget{}

	for source, destination := range destinations {
		provider, err := resolveBackupProvider(resolver, source)
		if err != nil {
			if destination.Config.EnsureAvailable {
				return nil, fmt.Errorf("error building backup provider for %s: %w", source, err)
			}

			log.Warn().Err(err).Str("source", string(source)).Str("destination", string(destination.Provider)).Msg("backup provider unavailable, source will proceed without a backup")

			continue
		}

		backups[source] = objects.BackupTarget{
			Provider:       provider,
			ReadFromBackup: destination.ReadFromBackup,
		}

		log.Debug().Str("source", string(source)).Str("destination", string(destination.Provider)).Str("bucket", destination.Config.Bucket).Bool("read_from_backup", destination.ReadFromBackup).Msg("configured backup target for storage provider")
	}

	return backups, nil
}

// resolveBackupProvider resolves and builds the backup destination for one source provider by
// asking the rule chain with the backup hint applied
func resolveBackupProvider(resolver *providerResolver, source storage.ProviderType) (storage.Provider, error) {
	ctx := objects.WithBackupSourceHint(context.Background(), source)

	res, ok := resolver.Resolve(ctx).Get()
	if !ok || res.Builder == nil {
		return nil, fmt.Errorf("%w: %s", errUnsupportedProvider, source)
	}

	if marker, hasMarker := res.Config.Extra(storage.BackupTargetExtraKey); !hasMarker || marker != true {
		return nil, fmt.Errorf("%w: %s resolved a live provider", errUnsupportedProvider, source)
	}

	return res.Builder.Build(ctx, res.Output, res.Config)
}
