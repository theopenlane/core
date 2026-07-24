package serveropts

import (
	"context"
	"math"
	"strconv"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// backfillBypassCaps lets the backfill write organizations and memberships without a request caller while
// skipping the org-filter, FGA, and managed-group guards the membership hooks would otherwise apply
const backfillBypassCaps = auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation | auth.CapBypassManagedGroup

// maxExactExternalID is the largest float64 that can still hold every integer exactly (2^53)
const maxExactExternalID = float64(1 << 53)

// WithBackfill runs a one-time, non-blocking, config-gated, idempotent startup backfills
// use-cases for this are things a db migration can't easily handle, computed data or fields, or repairs
func WithBackfill(ctx context.Context, dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil || !s.Config.Settings.Backfill.Enabled {
			return
		}

		go func() {
			backfillCtx := privacy.DecisionContext(ctx, privacy.Allow)
			backfillCtx = auth.WithCaller(backfillCtx, &auth.Caller{Capabilities: backfillBypassCaps})

			backfillDirectoryExternalIDs(backfillCtx, dbClient)
		}()
	})
}

// backfillDirectoryExternalIDs rewrites directory account and group external ids that the CEL double
// conversion stored in scientific notation (e.g. "1.47884153e+08" back to "147884153"); the "e+"
// contains query is just a cheap prefilter, the strict parse in decimalExternalID decides what
// actually gets touched, so values like emails that happen to contain "e+" are never rewritten
func backfillDirectoryExternalIDs(ctx context.Context, dbClient *ent.Client) {
	accounts, err := dbClient.DirectoryAccount.Query().
		Where(directoryaccount.ExternalIDContains("e+")).
		All(ctx)
	if err != nil {
		log.Error().Err(err).Msg("backfill: failed to query directory accounts with scientific notation external ids")
		return
	}

	accountsCorrected := 0

	for _, account := range accounts {
		corrected, ok := decimalExternalID(account.ExternalID)
		if !ok {
			continue
		}

		conflict, err := dbClient.DirectoryAccount.Query().
			Where(
				directoryaccount.OwnerID(account.OwnerID),
				directoryaccount.ExternalID(corrected),
				directoryaccount.IDNEQ(account.ID),
			).
			Exist(ctx)
		if err != nil {
			log.Error().Err(err).Str("directory_account_id", account.ID).Msg("backfill: failed to check directory account external id conflict")

			continue
		}

		if conflict {
			log.Warn().Str("directory_account_id", account.ID).Str("external_id", corrected).Msg("backfill: corrected external id already held by another directory account, skipping")

			continue
		}

		// external_id is immutable, so the fix has to go through the sql modifier
		if err := dbClient.DirectoryAccount.UpdateOneID(account.ID).
			Modify(func(u *sql.UpdateBuilder) {
				u.Set(directoryaccount.FieldExternalID, corrected)
			}).
			Exec(ctx); err != nil {
			log.Error().Err(err).Str("directory_account_id", account.ID).Msg("backfill: failed to correct directory account external id")

			continue
		}

		accountsCorrected++
	}

	groups, err := dbClient.DirectoryGroup.Query().
		Where(directorygroup.ExternalIDContains("e+")).
		All(ctx)
	if err != nil {
		log.Error().Err(err).Msg("backfill: failed to query directory groups with scientific notation external ids")
		return
	}

	groupsCorrected := 0

	for _, group := range groups {
		corrected, ok := decimalExternalID(group.ExternalID)
		if !ok {
			continue
		}

		conflict, err := dbClient.DirectoryGroup.Query().
			Where(
				directorygroup.OwnerID(group.OwnerID),
				directorygroup.ExternalID(corrected),
				directorygroup.IDNEQ(group.ID),
			).
			Exist(ctx)
		if err != nil {
			log.Error().Err(err).Str("directory_group_id", group.ID).Msg("backfill: failed to check directory group external id conflict")

			continue
		}

		if conflict {
			log.Warn().Str("directory_group_id", group.ID).Str("external_id", corrected).Msg("backfill: corrected external id already held by another directory group, skipping")

			continue
		}

		// external_id is immutable, so the fix has to go through the sql modifier
		if err := dbClient.DirectoryGroup.UpdateOneID(group.ID).
			Modify(func(u *sql.UpdateBuilder) {
				u.Set(directorygroup.FieldExternalID, corrected)
			}).
			Exec(ctx); err != nil {
			log.Error().Err(err).Str("directory_group_id", group.ID).Msg("backfill: failed to correct directory group external id")

			continue
		}

		groupsCorrected++
	}

	log.Info().Int("accounts_corrected", accountsCorrected).Int("groups_corrected", groupsCorrected).Msg("backfill: directory external id notation corrected")
}

// decimalExternalID converts a scientific notation external id back to plain digits, refusing
// anything that isn't a whole number float64 can represent exactly
func decimalExternalID(externalID string) (string, bool) {
	value, err := strconv.ParseFloat(externalID, 64)
	if err != nil || value != math.Trunc(value) || math.Abs(value) >= maxExactExternalID {
		return "", false
	}

	return strconv.FormatFloat(value, 'f', -1, 64), true
}
