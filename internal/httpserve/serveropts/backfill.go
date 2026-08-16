package serveropts

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	intobvs "github.com/theopenlane/core/internal/integrations/observability"
	"github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/logx"
)

// backfillBypassCaps lets the backfill write organizations and memberships without a request caller while
// skipping the org-filter, FGA, and managed-group guards the membership hooks would otherwise apply
const backfillBypassCaps = auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation | auth.CapBypassManagedGroup

// maxExactExternalID is the largest float64 that can still hold every integer exactly (2^53)
const maxExactExternalID = float64(1 << 53)

// integrationReconfigureObjectType is the notification object type for a misconfigured installation
const integrationReconfigureObjectType = "integration.reconfiguration.required"

// WithBackfill runs a one-time, non-blocking, config-gated, idempotent startup backfills
// use-cases for this are things a db migration can't easily handle, computed data or fields, or repairs
func WithBackfill(ctx context.Context, dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil || !s.Config.Settings.Backfill.Enabled {
			return
		}

		rt := s.Config.Handler.IntegrationsRuntime

		go func() {
			backfillCtx := privacy.DecisionContext(ctx, privacy.Allow)
			backfillCtx = auth.WithCaller(backfillCtx, &auth.Caller{Capabilities: backfillBypassCaps})

			BackfillDirectoryDuplicates(backfillCtx, dbClient)

			backfillDirectoryExternalIDs(backfillCtx, dbClient)

			backfillManagedGroupMemberships(backfillCtx, dbClient)

			if rt != nil {
				backfillIntegrationConfiguration(backfillCtx, dbClient, rt)
			}
		}()
	})
}

// backfillIntegrationConfiguration flags connected installations whose stored user input no longer
// satisfies their definition, so the owner is told to reconnect rather than left with a cycle that
// fails on every run
func backfillIntegrationConfiguration(ctx context.Context, dbClient *ent.Client, rt *runtime.Runtime) {
	installations, err := dbClient.Integration.Query().
		Where(integration.StatusEQ(enums.IntegrationStatusConnected)).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query connected integrations")

		return
	}

	var flagged int

	for _, installation := range installations {
		def, ok := rt.Registry().Definition(installation.DefinitionID)
		if !ok {
			continue
		}

		installCtx := intobvs.WithInstallation(ctx, installation)

		if err := rt.ValidateUserInput(installCtx, def, installation.Config.ClientConfig); err == nil {
			continue
		}

		if err := markIntegrationErrored(installCtx, dbClient, installation, def.DisplayName); err != nil {
			logx.FromContext(installCtx).Error().Err(err).Msg("backfill: failed flagging misconfigured integration")

			continue
		}

		flagged++
	}

	logx.FromContext(ctx).Info().Int("flagged_misconfigured", flagged).Int("reviewed", len(installations)).Msg("backfill: connected integrations reviewed")
}

// markIntegrationErrored flags one installation as misconfigured and notifies the owning organization
func markIntegrationErrored(ctx context.Context, dbClient *ent.Client, installation *ent.Integration, displayName string) error {
	if err := dbClient.Integration.UpdateOneID(installation.ID).
		SetStatus(enums.IntegrationStatusErrored).
		Exec(ctx); err != nil {
		return err
	}

	logx.FromContext(ctx).Warn().Msg("backfill: integration is missing required configuration, marked errored")

	_, err := dbClient.Notification.Create().
		SetOwnerID(installation.OwnerID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType(integrationReconfigureObjectType).
		SetTitle(fmt.Sprintf("%s needs to be reconnected", displayName)).
		SetBody(fmt.Sprintf("The %s integration is missing required configuration and has stopped syncing. Reconnect it and supply the required settings to resume.", displayName)).
		SetData(map[string]any{
			"integration_id": installation.ID,
			"definition_id":  installation.DefinitionID,
		}).
		SetTopic(enums.NotificationTopicIntegration).
		Save(ctx)

	return err
}

// backfillManagedGroupMemberships restores memberships in the system managed groups that were
// removed by the formerly unscoped org membership cascade delete, which wiped a user's group
// memberships across every organization when they left a single one. The default groups
// (Admins, Viewers, All Members) are derived from each member's role in the organization; each
// member's personal managed group is matched by its "<display name> - <user id>" naming
// convention. Only missing rows are added, so the pass is idempotent
func backfillManagedGroupMemberships(ctx context.Context, dbClient *ent.Client) {
	orgs, err := dbClient.Organization.Query().
		Where(
			organization.DeletedAtIsNil(),
			organization.PersonalOrg(false),
		).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query organizations for managed group membership repair")

		return
	}

	restored := 0

	for _, org := range orgs {
		managedGroups, err := dbClient.Group.Query().
			Where(
				group.OwnerID(org.ID),
				group.IsManaged(true),
			).
			All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("organization_id", org.ID).Msg("backfill: failed to query managed groups")

			continue
		}

		if len(managedGroups) == 0 {
			continue
		}

		groupsByName := make(map[string]*ent.Group, len(managedGroups))
		personalGroupsByUserID := make(map[string]*ent.Group)
		groupIDs := make([]string, 0, len(managedGroups))

		for _, g := range managedGroups {
			groupIDs = append(groupIDs, g.ID)

			switch g.Name {
			case hooks.AdminsGroup, hooks.ViewersGroup, hooks.AllMembersGroup:
				groupsByName[g.Name] = g
			default:
				// personal managed groups are named "<display name> - <user id>";
				// the user id suffix is the only stable identifier since tags do not
				// include it and display names can change
				if idx := strings.LastIndex(g.Name, " - "); idx >= 0 {
					personalGroupsByUserID[g.Name[idx+len(" - "):]] = g
				}
			}
		}

		members, err := dbClient.OrgMembership.Query().
			Where(orgmembership.OrganizationID(org.ID)).
			All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("organization_id", org.ID).Msg("backfill: failed to query org memberships")

			continue
		}

		existing, err := dbClient.GroupMembership.Query().
			Where(groupmembership.GroupIDIn(groupIDs...)).
			All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("organization_id", org.ID).Msg("backfill: failed to query managed group memberships")

			continue
		}

		present := make(map[string]struct{}, len(existing))
		for _, gm := range existing {
			present[gm.GroupID+"|"+gm.UserID] = struct{}{}
		}

		for _, m := range members {
			targetGroups := []*ent.Group{groupsByName[hooks.AllMembersGroup]}

			switch m.Role {
			case enums.RoleMember:
				targetGroups = append(targetGroups, groupsByName[hooks.ViewersGroup])
			case enums.RoleAdmin, enums.RoleSuperAdmin, enums.RoleOwner:
				targetGroups = append(targetGroups, groupsByName[hooks.AdminsGroup])
			}

			targetGroups = append(targetGroups, personalGroupsByUserID[m.UserID])

			for _, g := range targetGroups {
				if g == nil {
					continue
				}

				if _, ok := present[g.ID+"|"+m.UserID]; ok {
					continue
				}

				role := enums.RoleMember
				if err := dbClient.GroupMembership.Create().
					SetInput(ent.CreateGroupMembershipInput{
						Role:    &role,
						UserID:  m.UserID,
						GroupID: g.ID,
					}).
					Exec(ctx); err != nil {
					logx.FromContext(ctx).Error().Err(err).Str("organization_id", org.ID).Str("user_id", m.UserID).Str("group", g.Name).Msg("backfill: failed to restore managed group membership")

					continue
				}

				restored++
			}
		}
	}

	logx.FromContext(ctx).Info().Int("restored_memberships", restored).Int("organizations", len(orgs)).Msg("backfill: managed group memberships repaired")
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

// directoryDupKey is one (integration_id, external_id) identity with its row count, scanned
// from the duplicate-detection group-by
type directoryDupKey struct {
	IntegrationID string `json:"integration_id"`
	ExternalID    string `json:"external_id"`
	Count         int    `json:"count"`
}

// minForkedRowCount is the row count at which an identity is considered forked
const minForkedRowCount = 2

// BackfillDirectoryDuplicates removes forked directory rows sharing an (integration_id,
// external_id) identity: concurrent syncs raced the same record into multiple rows, which the
// run-scoped unique index permitted, and the ingest lookup fails "not singular" on every
// subsequent sync until only one row remains. The earliest row survives; the extras and their
// memberships are deleted rather than repointed, because membership foreign keys are immutable
// and the next complete directory sync rebuilds active memberships against the survivor
func BackfillDirectoryDuplicates(ctx context.Context, dbClient *ent.Client) {
	var groupKeys []directoryDupKey
	if err := dbClient.DirectoryGroup.Query().
		GroupBy(directorygroup.FieldIntegrationID, directorygroup.FieldExternalID).
		Aggregate(ent.Count()).
		Scan(ctx, &groupKeys); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to scan directory groups for duplicate identities")

		return
	}

	groupsRemoved := 0

	for _, key := range groupKeys {
		if key.Count < minForkedRowCount {
			continue
		}

		rows, err := dbClient.DirectoryGroup.Query().
			Where(directorygroup.IntegrationID(key.IntegrationID), directorygroup.ExternalID(key.ExternalID)).
			Order(ent.Asc(directorygroup.FieldID)).
			All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("integration_id", key.IntegrationID).Str("external_id", key.ExternalID).Msg("backfill: failed to load duplicate directory groups")

			continue
		}

		for _, loser := range rows[1:] {
			if _, err := dbClient.DirectoryMembership.Delete().
				Where(directorymembership.DirectoryGroupID(loser.ID)).
				Exec(ctx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("directory_group_id", loser.ID).Msg("backfill: failed to delete memberships of duplicate directory group")

				continue
			}

			if err := dbClient.DirectoryGroup.DeleteOneID(loser.ID).Exec(ctx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("directory_group_id", loser.ID).Msg("backfill: failed to delete duplicate directory group")

				continue
			}

			groupsRemoved++
		}
	}

	var accountKeys []directoryDupKey
	if err := dbClient.DirectoryAccount.Query().
		GroupBy(directoryaccount.FieldIntegrationID, directoryaccount.FieldExternalID).
		Aggregate(ent.Count()).
		Scan(ctx, &accountKeys); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to scan directory accounts for duplicate identities")

		return
	}

	accountsRemoved := 0

	for _, key := range accountKeys {
		if key.Count < minForkedRowCount {
			continue
		}

		rows, err := dbClient.DirectoryAccount.Query().
			Where(directoryaccount.IntegrationID(key.IntegrationID), directoryaccount.ExternalID(key.ExternalID)).
			Order(ent.Asc(directoryaccount.FieldID)).
			All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("integration_id", key.IntegrationID).Str("external_id", key.ExternalID).Msg("backfill: failed to load duplicate directory accounts")

			continue
		}

		for _, loser := range rows[1:] {
			if _, err := dbClient.DirectoryMembership.Delete().
				Where(directorymembership.DirectoryAccountID(loser.ID)).
				Exec(ctx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("directory_account_id", loser.ID).Msg("backfill: failed to delete memberships of duplicate directory account")

				continue
			}

			if err := dbClient.DirectoryAccount.DeleteOneID(loser.ID).Exec(ctx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("directory_account_id", loser.ID).Msg("backfill: failed to delete duplicate directory account")

				continue
			}

			accountsRemoved++
		}
	}

	logx.FromContext(ctx).Info().Int("groups_removed", groupsRemoved).Int("accounts_removed", accountsRemoved).Msg("backfill: forked directory records removed")
}
