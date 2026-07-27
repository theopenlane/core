package notifications

import (
	"context"
	"fmt"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/program"
	"github.com/theopenlane/core/internal/ent/generated/programmembership"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

func handleProgramMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	props := ctx.Envelope.Headers.Properties
	if !isProgramReady(payload, props) {
		return nil
	}

	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	if client == nil {
		return ErrFailedToGetClient
	}

	programID, ok := eventqueue.MutationEntityID(payload, props)
	if !ok {
		return ErrEntityIDNotFound
	}

	if programID == "" {
		return ErrEntityIDNotFound
	}

	if err := addNotificationForAuditor(ctx.Context, client, programID); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).
			Str("program_id", programID).
			Msg("failed to send program ready for auditor notification")
		return err
	}

	return nil
}

func addNotificationForAuditor(ctx context.Context, client *generated.Client, id string) error {
	program, err := client.Program.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to query program: %w", err)
	}

	if err := inviteAuditor(ctx, client, program); err != nil {
		return err
	}

	ids, err := getProgramAuditorUserIDs(ctx, client, id)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		ids, err = getOrgAuditorUserIDs(ctx, client, program.OwnerID)
		if err != nil {
			return err
		}
	}

	if len(ids) == 0 {
		logx.FromContext(ctx).Warn().Str("program_id", id).Str("org_id", program.OwnerID).Msg("no auditors found for program ready notification")
		return nil
	}

	return newNotificationCreation(ctx, client, ids, buildNotificationInputForAuditor(program))
}

func inviteAuditor(ctx context.Context, client *generated.Client, programEntity *generated.Program) error {
	email := strings.TrimSpace(programEntity.AuditorEmail)
	if email == "" {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	exists, err := client.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(programEntity.OwnerID),
			orgmembership.HasUserWith(user.EmailEqualFold(email)),
		).
		Exist(allowCtx)
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
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	var userIDs []string
	err := client.ProgramMembership.Query().
		Where(
			programmembership.ProgramID(programID),
			programmembership.RoleEQ(enums.RoleAuditor),
		).
		Select(programmembership.FieldUserID).
		Scan(allowCtx, &userIDs)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func getOrgAuditorUserIDs(ctx context.Context, client *generated.Client, orgID string) ([]string, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	var userIDs []string
	err := client.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(orgID),
			orgmembership.RoleEQ(enums.RoleAuditor),
		).
		Select(orgmembership.FieldUserID).
		Scan(allowCtx, &userIDs)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func isProgramReady(payload eventqueue.MutationGalaPayload, props map[string]string) bool {
	status, ok := eventqueue.MutationValue(payload, program.FieldStatus)
	if !ok {
		if value := eventqueue.MutationStringFromProperties(props, program.FieldStatus); value != "" {
			status = value
			ok = true
		}
	}

	if ok {
		status, ok := eventqueue.ParseEnum(status, enums.ToProgramStatus, enums.ProgramStatusInvalid)
		if ok && status == enums.ProgramStatusReadyForAuditor {
			return true
		}
	}

	ready, ok := eventqueue.MutationValue(payload, program.FieldAuditorReady)
	if !ok {
		return false
	}

	isReady, ok := ready.(bool)
	return ok && isReady
}

func buildNotificationInputForAuditor(programEntity *generated.Program) *generated.CreateNotificationInput {
	dataMap := map[string]any{
		"program_id": programEntity.ID,
		"url":        getURLPathForObject(programEntity.ID, generated.TypeProgram),
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
