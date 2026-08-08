package notifications

import (
	"context"
	"fmt"
	"slices"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"golang.org/x/mod/semver"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/pkg/logx"
)

type standardControl struct {
	ID                         string  `json:"id"`
	OwnerID                    string  `json:"owner_id"`
	RefCode                    string  `json:"ref_code"`
	Title                      *string `json:"title,omitempty"`
	ReferenceFrameworkRevision *string `json:"reference_framework_revision"`
}

type standardSubcontrol struct {
	ControlID string `json:"control_id"`
	standardControl
}

// handleStandardMutation processes standard mutations and creates notifications
// for org admins when a system-owned standard revision is bumped up
func handleStandardMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	standardID := payload.EntityID
	allowCtx := inv.Context

	std, err := inv.Client.Standard.Get(allowCtx, standardID)
	if err != nil {
		return fmt.Errorf("failed to query standard: %w", err)
	}

	if !std.SystemOwned {
		return nil
	}

	var controls []standardControl

	err = inv.Client.Control.Query().
		Where(
			control.StandardID(standardID),
			control.OwnerIDNotNil(),
		).
		Select(
			control.FieldID,
			control.FieldOwnerID,
			control.FieldRefCode,
			control.FieldTitle,
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
		return c.OwnerID != "" && models.IsMajorMinorBump(lo.FromPtrOr(c.ReferenceFrameworkRevision, ""), std.Revision)
	})

	if len(filteredControls) == 0 {
		return nil
	}

	groups := lo.GroupBy(filteredControls, func(c standardControl) string {
		return c.OwnerID
	})

	controlIDs := lo.Map(filteredControls, func(c standardControl, _ int) string {
		return c.ID
	})

	subcontrols, err := fetchAffectedSubcontrols(allowCtx, client, controlIDs, std.Revision)
	if err != nil {
		return err
	}

	subcontrolsByControlID := lo.GroupBy(subcontrols, func(s standardSubcontrol) string {
		return s.ControlID
	})

	lo.ForEach(lo.Entries(groups), func(entry lo.Entry[string, []standardControl], _ int) {
		orgID := entry.Key
		controls := entry.Value
		logCtx := logx.WithFields(inv.Context, map[string]any{"org_id": orgID})

		ids, err := fetchOrgAdminsAndOwners(allowCtx, inv.Client, orgID)
		if err != nil {
			logx.FromContext(logCtx).Error().Err(err).Msg("failed to get org admin and owner IDs")

			return
		}

		if len(ids) == 0 {
			return
		}

		oldRevision := pickOldestRevision(controls)
		changeType := models.DetectSemverBump(oldRevision, std.Revision)
		orgSubcontrols := fetchSubcontrolsOwnedByControl(controls, subcontrolsByControlID)

		data := map[string]any{
			"url":                        getURLPathForObject(standardID, generated.TypeStandard),
			"standard_id":                standardID,
			"standard_short_name":        std.ShortName,
			"old_revision":               oldRevision,
			"new_revision":               std.Revision,
			"change_type":                changeType,
			"affected_controls_count":    len(controls),
			"affected_subcontrols_count": len(orgSubcontrols),
			"affected_controls":          controls,
			"affected_subcontrols":       orgSubcontrols,
			"diff_available":             oldRevision != "",
			"diff_input": map[string]any{
				"standard_id":  standardID,
				"old_revision": oldRevision,
				"new_revision": std.Revision,
			},
			"accept_all_input": map[string]any{
				"standard_id": standardID,
			},
			"accept_all_available":        true,
			"acceptance_updates_revision": std.Revision,
			"url":                         entityops.ConsoleObjectPath(generated.TypeStandard, standardID),
			"standard_id":                 standardID,
			"standard_short_name":         std.ShortName,
			"old_revision":                value.revision,
			"new_revision":                std.Revision,
			"change_type":                 detectVersionBump(value.revision, std.Revision),
			"affected_controls_count":     value.controlCount,
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

		notificationCtx := auth.WithCaller(inv.Context, &auth.Caller{
			SubjectID:      ids[0],
			OrganizationID: orgID,
		})

		if err := newNotificationCreation(notificationCtx, inv.Client, ids, notifInput); err != nil {
			logx.FromContext(logCtx).Error().Err(err).Msg("failed to create standard update notification")

			return
		}
	})

	return nil
}

func fetchAffectedSubcontrols(ctx context.Context, client *generated.Client, ids []string, revision string) ([]standardSubcontrol, error) {
	if len(ids) == 0 {
		return []standardSubcontrol{}, nil
	}

	var controls []standardSubcontrol

	err := client.Subcontrol.Query().
		Where(subcontrol.ControlIDIn(ids...)).
		Select(
			subcontrol.FieldID,
			subcontrol.FieldControlID,
			subcontrol.FieldRefCode,
			subcontrol.FieldTitle,
			subcontrol.FieldReferenceFrameworkRevision,
		).
		Scan(ctx, &controls)
	if err != nil {
		return nil, fmt.Errorf("failed to query subcontrols for standard update: %w", err)
	}

	affected := lo.Filter(controls, func(s standardSubcontrol, _ int) bool {
		return models.IsMajorMinorBump(lo.FromPtrOr(s.ReferenceFrameworkRevision, ""), revision)
	})

	slices.SortFunc(affected, func(a, b standardSubcontrol) int {
		if a.ControlID == b.ControlID {
			if a.RefCode < b.RefCode {
				return -1
			}

			if a.RefCode > b.RefCode {
				return 1
			}

			return 0
		}

		if a.ControlID < b.ControlID {
			return -1
		}

		return 1
	})

	return affected, nil
}

func fetchSubcontrolsOwnedByControl(controls []standardControl, grouped map[string][]standardSubcontrol) []standardSubcontrol {
	subcontrols := []standardSubcontrol{}

	for _, c := range controls {
		subcontrols = append(subcontrols, grouped[c.ID]...)
	}

	return subcontrols
}

func pickOldestRevision(controls []standardControl) string {
	revisions := lo.FilterMap(controls, func(c standardControl, _ int) (string, bool) {
		revision := lo.FromPtrOr(c.ReferenceFrameworkRevision, "")

		return revision, revision != ""
	})

	if len(revisions) == 0 {
		return ""
	}

	semver.Sort(revisions)

	return revisions[0]
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
