package serveropts

import (
	"context"
	"strings"

	"github.com/samber/do/v2"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/group"
	"github.com/theopenlane/core/v2/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// backfillBypassCaps lets the backfill write organizations and memberships without a request caller while
// skipping the org-filter, FGA, and managed-group guards the membership hooks would otherwise apply
const backfillBypassCaps = auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation | auth.CapBypassManagedGroup

// backfillTopic is the gala topic backfill runs are submitted on
var backfillTopic = gala.NamespacedTopic[backfillRequest](gala.System, "startup.backfill")

// backfillUniqueKey is the run-once uniqueness key: every pod submits the same key, River
// keeps the first insert and skips the rest across live and terminal job states
const backfillUniqueKey = "startup-backfill"

// backfillRequest is the payload for a backfill run submission
type backfillRequest struct{}

// WithBackfill submits the config-gated, idempotent startup backfills as a run-once gala job:
// every pod submits the same unique key, so exactly one process executes the run.
// Use-cases are things a db migration can't easily handle, computed data or fields, or repairs
func WithBackfill(ctx context.Context, galaApp *gala.Gala) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if !s.Config.Settings.Backfill.Enabled {
			return
		}

		if _, err := gala.Register(galaApp, gala.Definition[backfillRequest]{
			Topic: backfillTopic,
			Caller: func(*auth.Caller, backfillRequest) *auth.Caller {
				return &auth.Caller{Capabilities: backfillBypassCaps}
			},
			Handle: func(handlerCtx gala.HandlerContext, _ backfillRequest) error {
				dbClient := do.MustInvoke[*ent.Client](handlerCtx.Injector)
				rt := do.MustInvoke[*runtime.Runtime](handlerCtx.Injector)

				BackfillDirectoryDuplicates(handlerCtx.Context, dbClient)
				backfillManagedGroupMemberships(handlerCtx.Context, dbClient)
				backfillIntegrationConfiguration(handlerCtx.Context, dbClient, rt)
				backfillReconcileLoops(handlerCtx.Context, dbClient, rt)

				return nil
			},
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to register listener")

			return
		}

		if _, err := galaApp.EmitWithHeaders(ctx, backfillTopic.Name, backfillRequest{}, gala.Headers{
			UniqueKey:  backfillUniqueKey,
			UniqueOnce: true,
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to submit run")
		}
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

		if err := rt.MarkIntegrationUnhealthy(installCtx, installation, "it is missing required configuration"); err != nil {
			logx.FromContext(installCtx).Error().Err(err).Msg("backfill: failed flagging misconfigured integration")

			continue
		}

		flagged++
	}

	logx.FromContext(ctx).Info().Int("flagged_misconfigured", flagged).Int("reviewed", len(installations)).Msg("backfill: connected integrations reviewed")
}

// backfillReconcileLoops collapses each connected installation's recurring loops to exactly one
// per operation: every active reconcile job is cancelled and a single fresh loop is emitted with
// insert-time uniqueness, removing duplicate loops left by historical seeding races. Emitted
// loops are unique-keyed, so re-running the backfill against a healthy state is a reset, not a
// duplication
func backfillReconcileLoops(ctx context.Context, dbClient *ent.Client, rt *runtime.Runtime) {
	installations, err := dbClient.Integration.Query().
		Where(integration.StatusEQ(enums.IntegrationStatusConnected)).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query connected integrations for loop reset")

		return
	}

	var reset int

	for _, installation := range installations {
		installCtx := intobvs.WithInstallation(ctx, installation)

		if err := rt.ResetReconcileLoops(installCtx, installation); err != nil {
			logx.FromContext(installCtx).Error().Err(err).Msg("backfill: failed resetting reconcile loops")

			continue
		}

		reset++
	}

	logx.FromContext(ctx).Info().Int("reset", reset).Int("reviewed", len(installations)).Msg("backfill: reconcile loop reset completed")
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
