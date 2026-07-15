package serveropts

import (
	"bytes"
	"context"

	"github.com/rs/zerolog/log"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/objects/upload"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// backfillBypassCaps lets the backfill write organizations and memberships without a request caller while
// skipping the org-filter, FGA, and managed-group guards the membership hooks would otherwise apply
const backfillBypassCaps = auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation | auth.CapBypassManagedGroup

// WithBackfill runs one-time, idempotent startup backfills for data introduced by recent migrations:
// organization slug names, owner SSO exemptions, trust center update templates, and file md5 hashes.
// It is gated by the Backfill.Enabled config flag and runs in the background so it never blocks
// server startup
func WithBackfill(ctx context.Context, dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil || !s.Config.Settings.Backfill.Enabled {
			return
		}

		go func() {
			backfillCtx := privacy.DecisionContext(ctx, privacy.Allow)
			backfillCtx = auth.WithCaller(backfillCtx, &auth.Caller{Capabilities: backfillBypassCaps})

			backfillOrganizationSlugs(backfillCtx, dbClient)
			backfillOwnerSSOExemptions(backfillCtx, dbClient)
			backfillFileMD5Hashes(backfillCtx, dbClient)
		}()
	})
}

// backfillOrganizationSlugs derives slug_name from the organization name for organizations that pre-date
// the field, matching the kebab-case format applied to newly created organizations
func backfillOrganizationSlugs(ctx context.Context, dbClient *ent.Client) {
	orgs, err := dbClient.Organization.Query().
		Where(organization.Or(organization.SlugNameIsNil(), organization.SlugNameEQ(""))).
		All(ctx)
	if err != nil {
		log.Error().Err(err).Msg("backfill: failed to query organizations missing a slug name")
		return
	}

	updated := 0

	for _, org := range orgs {
		if err := dbClient.Organization.UpdateOneID(org.ID).
			SetSlugName(strcase.KebabCase(org.Name)).
			Exec(ctx); err != nil {
			log.Error().Err(err).Str("organization_id", org.ID).Msg("backfill: failed to set organization slug name")

			continue
		}

		updated++
	}

	log.Info().Int("updated", updated).Int("candidates", len(orgs)).Msg("backfill: organization slug names populated")
}

// backfillOwnerSSOExemptions marks existing organization owners as SSO exempt so the exemption is explicit
// on the membership record, matching how owners are seeded for newly created organizations
func backfillOwnerSSOExemptions(ctx context.Context, dbClient *ent.Client) {
	updated, err := dbClient.OrgMembership.Update().
		Where(orgmembership.RoleEQ(enums.RoleOwner), orgmembership.SSOExempt(false)).
		SetSSOExempt(true).
		SetSSOExemptReason("organization owner").
		Save(ctx)
	if err != nil {
		log.Error().Err(err).Msg("backfill: failed to set owner SSO exemptions")
		return
	}

	log.Info().Int("updated", updated).Msg("backfill: owner SSO exemptions populated")
}

func backfillFileMD5Hashes(ctx context.Context, dbClient *ent.Client) {
	if dbClient.ObjectManager == nil {
		log.Warn().Msg("backfill: object manager is nil, skipping backfill for file md5 hashes")
		return
	}

	const batchSize = 100

	totalFiles := 0
	updatedCounter := 0
	failedCounter := 0
	lastKnownID := ""

	for {
		query := dbClient.File.Query().
			Where(file.Or(file.Md5HashIsNil(), file.Md5HashEQ(""))).
			Order(file.ByID()).
			Limit(batchSize)

		if lastKnownID != "" {
			query = query.Where(file.IDGT(lastKnownID))
		}

		files, err := query.All(ctx)
		if err != nil {
			log.Error().Err(err).Msg("backfill: failed to query files missing md5 hash")
			return
		}

		if len(files) == 0 {
			break
		}

		totalFiles += len(files)

		for _, f := range files {
			lastKnownID = f.ID

			file := interceptors.StorageFileFromEnt(f)

			if file == nil || file.ProviderHints == nil || file.ProviderHints.KnownProvider == "" {
				failedCounter++
				log.Error().Str("file_id", f.ID).
					Msg("backfill: file storage provider is missing, skipping md5 hash")

				continue
			}

			downloaded, err := dbClient.ObjectManager.Download(ctx, nil, file, &storage.DownloadOptions{})
			if err != nil {
				failedCounter++
				log.Error().Err(err).Str("file_id", f.ID).
					Str("storage_key", f.StoreKey).
					Str("storage_path", f.StoragePath).
					Msg("backfill: failed to download file for md5 hash")

				continue
			}

			if downloaded == nil {
				failedCounter++
				log.Error().Str("file_id", f.ID).Msg("backfill: downloaded file is nil")

				continue
			}

			computedHash, err := upload.ComputeMD5Hash(bytes.NewReader(downloaded.File))
			if err != nil {
				failedCounter++
				log.Error().Err(err).Str("file_id", f.ID).Msg("backfill: failed to calculate md5 hash")

				continue
			}

			if err := dbClient.File.UpdateOneID(f.ID).SetMd5Hash(string(computedHash)).Exec(ctx); err != nil {
				failedCounter++
				log.Error().Err(err).Str("file_id", f.ID).Msg("backfill: failed to set file md5 hash")

				continue
			}

			updatedCounter++
		}
	}

	log.Info().Int("updated_files", updatedCounter).
		Int("failed_files", failedCounter).
		Int("total_files_without_hash", totalFiles).
		Msg("backfill: file md5 hashes populated in database")
}
