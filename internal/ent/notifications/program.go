package notifications

import (
	"context"
	"fmt"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/program"
	"github.com/theopenlane/core/internal/ent/generated/programmembership"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/logx"
)

func handleProgramMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	if !isProgramReady(payload) {
		return nil
	}

	if err := addNotificationForAuditor(inv.Context, inv.Client, payload.EntityID); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to send program ready for auditor notification")
		return err
	}

	return nil
}

func addNotificationForAuditor(ctx context.Context, client *generated.Client, id string) error {
	program, found, err := entityops.LoadEntity(ctx, id, client.Program.Get)
	switch {
	case err != nil:
		return fmt.Errorf("failed to query program: %w", err)
	case !found:
		return nil
	}

	if err := inviteAuditor(ctx, client, program); err != nil {
		return err
	}

	ids, err := getProgramAuditorUserIDs(ctx, client, id)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		ids, err = orgUserIDsByRole(ctx, client, program.OwnerID, enums.RoleAuditor)
		if err != nil {
			return err
		}
	}

	if len(ids) == 0 {
		logx.FromContext(ctx).Warn().Str("org_id", program.OwnerID).Msg("no auditors found for program ready notification")
		return nil
	}

	return entityops.CreateNotifications(ctx, client, ids, buildNotificationInputForAuditor(program))
}

func inviteAuditor(ctx context.Context, client *generated.Client, programEntity *generated.Program) error {
	email := strings.TrimSpace(programEntity.AuditorEmail)
	if email == "" {
		return nil
	}

	exists, err := client.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(programEntity.OwnerID),
			orgmembership.HasUserWith(user.EmailEqualFold(email)),
		).
		Exist(ctx)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	role := enums.RoleAuditor
	input := generated.CreateInviteInput{
		Recipient: email,
		Role:      &role,
		OwnerID:   &programEntity.OwnerID,
	}

	_, err = client.Invite.Create().SetInput(input).Save(ctx)

	return err
}

func getProgramAuditorUserIDs(ctx context.Context, client *generated.Client, programID string) ([]string, error) {
	var userIDs []string
	err := client.ProgramMembership.Query().
		Where(
			programmembership.ProgramID(programID),
			programmembership.RoleEQ(enums.RoleAuditor),
		).
		Select(programmembership.FieldUserID).
		Scan(ctx, &userIDs)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func isProgramReady(payload entityops.MutationPayload) bool {
	if raw, ok := payload.Value(program.FieldStatus); ok {
		if status, ok := entityops.ParseEnum(raw, enums.ToProgramStatus, enums.ProgramStatusInvalid); ok && status == enums.ProgramStatusReadyForAuditor {
			return true
		}
	}

	ready, ok := payload.Value(program.FieldAuditorReady)
	if !ok {
		return false
	}

	isReady, ok := ready.(bool)
	return ok && isReady
}

func buildNotificationInputForAuditor(programEntity *generated.Program) *generated.CreateNotificationInput {
	dataMap := map[string]any{
		"program_id": programEntity.ID,
		"url":        entityops.ConsoleObjectPath(generated.TypeProgram, programEntity.ID),
	}

	topic := enums.NotificationTopicApproval
	body := "Program is ready for auditor review"
	if programEntity.Name != "" {
		body = fmt.Sprintf("%s is ready for auditor review", programEntity.Name)
	}

	return &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeOrganization,
		Title:            "Program ready for auditor",
		Body:             body,
		Data:             dataMap,
		OwnerID:          &programEntity.OwnerID,
		Topic:            &topic,
		ObjectType:       generated.TypeProgram,
	}
}
