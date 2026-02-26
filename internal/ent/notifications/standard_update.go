package notifications

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// handleStandardMutation processes standard mutations and creates notifications
// for org admins when a system-owned standard revision is bumped up
func handleStandardMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if !isUpdateOperation(payload.Operation) {
		return nil
	}

	if !eventqueue.MutationFieldChanged(payload, standard.FieldRevision) {
		return nil
	}

	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	props := ctx.Envelope.Headers.Properties

	standardID, ok := eventqueue.MutationEntityID(payload, props)
	if !ok {
		return ErrEntityIDNotFound
	}

	allowCtx := ctx.Context

	std, err := client.Standard.Get(allowCtx, standardID)
	if err != nil {
		return fmt.Errorf("failed to query standard: %w", err)
	}

	if !std.SystemOwned {
		return nil
	}

	type standardControl struct {
		OwnerID                    string  `json:"owner_id"`
		ReferenceFrameworkRevision *string `json:"reference_framework_revision"`
	}

	var controls []standardControl

	err = client.Control.Query().
		Where(
			control.StandardID(standardID),
			control.OwnerIDNotNil(),
		).
		Select(
			control.FieldOwnerID,
			control.FieldReferenceFrameworkRevision,
		).
		Scan(allowCtx, &controls)
	if err != nil {
		return fmt.Errorf("failed to query controls for standard: %w", err)
	}

	if len(controls) == 0 {
		return nil
	}

	filteredControls := lo.Filter(controls, func(c standardControl, _ int) bool {
		return c.OwnerID != ""
	})

	type organizations struct {
		revision     string
		controlCount int
	}

	groups := lo.GroupBy(filteredControls, func(c standardControl) string {
		return c.OwnerID
	})

	orgMap := lo.MapValues(groups, func(cs []standardControl, _ string) organizations {
		return organizations{
			revision:     lo.FromPtrOr(cs[0].ReferenceFrameworkRevision, ""),
			controlCount: len(cs),
		}
	})

	significantOrgs := lo.PickBy(orgMap, func(_ string, info organizations) bool {
		return detectVersionBump(info.revision, std.Revision) != ""
	})

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	lo.ForEach(lo.Entries(significantOrgs), func(entry lo.Entry[string, organizations], _ int) {
		orgID := entry.Key
		value := entry.Value

		ids, err := fetchOrgAdminsAndOwners(allowCtx, client, orgID)
		if err != nil {
			logx.FromContext(ctx.Context).Error().Err(err).
				Str("org_id", orgID).
				Msg("failed to get org admin and owner IDs")

			return
		}

		if len(ids) == 0 {
			return
		}

		data := map[string]any{
			"url":                     getURLPathForObject(consoleURL, standardID, generated.TypeStandard),
			"standard_id":             standardID,
			"standard_short_name":     std.ShortName,
			"old_revision":            value.revision,
			"new_revision":            std.Revision,
			"change_type":             detectVersionBump(value.revision, std.Revision),
			"affected_controls_count": value.controlCount,
		}

		topic := enums.NotificationTopicStandardUpdate
		notifInput := &generated.CreateNotificationInput{
			NotificationType: enums.NotificationTypeOrganization,
			Title:            fmt.Sprintf("%s update available", std.ShortName),
			Body:             fmt.Sprintf("%s has been updated to %s", std.ShortName, std.Revision),
			Data:             data,
			OwnerID:          &orgID,
			Topic:            &topic,
			ObjectType:       generated.TypeStandard,
		}

		notificationCtx := auth.WithAuthenticatedUser(ctx.Context, &auth.AuthenticatedUser{
			SubjectID:       ids[0],
			OrganizationID:  orgID,
			OrganizationIDs: []string{orgID},
		})

		if err := newNotificationCreation(notificationCtx, client, ids, notifInput); err != nil {
			logx.FromContext(ctx.Context).Error().Err(err).
				Str("org_id", orgID).
				Msg("failed to create standard update notification")

			return
		}
	})

	return nil
}

func fetchOrgAdminsAndOwners(ctx context.Context, client *generated.Client, orgID string) ([]string, error) {
	var ids []string

	err := client.OrgMembership.Query().
		Where(
			orgmembership.OrganizationIDEQ(orgID),
			orgmembership.RoleIn(enums.RoleOwner, enums.RoleAdmin),
		).
		Select(orgmembership.FieldUserID).
		Scan(ctx, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func detectVersionBump(oldRevision, newRevision string) string {
	oldVersion, err := models.ToSemverVersion(&oldRevision)
	if err != nil {
		return "major"
	}

	newVersion, err := models.ToSemverVersion(&newRevision)
	if err != nil {
		return "major"
	}

	if oldVersion.Major != newVersion.Major {
		return "major"
	}

	if oldVersion.Minor != newVersion.Minor {
		return "minor"
	}

	return ""
}
