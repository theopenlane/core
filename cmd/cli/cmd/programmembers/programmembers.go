//go:build cli

package programmembers

import (
	"context"
	"errors"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildCreateMembershipInput() (openlaneclient.CreateProgramMembershipInput, error) {
	var input openlaneclient.CreateProgramMembershipInput

	input.ProgramID = cmd.Config.String("program-id")
	if input.ProgramID == "" {
		return input, cmd.NewRequiredFieldMissingError("program id")
	}

	input.UserID = cmd.Config.String("user-id")
	if input.UserID == "" {
		return input, cmd.NewRequiredFieldMissingError("user id")
	}

	role := cmd.Config.String("role")
	if role == "" {
		role = "member"
	}

	enumRole, err := speccli.ParseRole(role)
	if err != nil {
		return input, err
	}

	input.Role = &enumRole

	return input, nil
}

func createProgramMembership(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.CreateProgramMembership, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	input, err := buildCreateMembershipInput()
	if err != nil {
		return nil, err
	}

	return client.CreateProgramMembership(ctx, input)
}

func lookupProgramMembership(ctx context.Context, client *openlaneclient.OpenlaneClient, programID, userID string) (string, error) {
	where := openlaneclient.ProgramMembershipWhereInput{
		ProgramID: &programID,
		UserID:    &userID,
	}

	result, err := client.GetProgramMemberships(ctx, cmd.First, cmd.Last, &where)
	if err != nil {
		return "", err
	}

	if len(result.ProgramMemberships.Edges) != 1 || result.ProgramMemberships.Edges[0].Node == nil {
		return "", errors.New("error getting existing relation")
	}

	return result.ProgramMemberships.Edges[0].Node.ID, nil
}

func buildUpdateMembershipInput(ctx context.Context, client *openlaneclient.OpenlaneClient) (string, openlaneclient.UpdateProgramMembershipInput, error) {
	programID := cmd.Config.String("program-id")
	if programID == "" {
		return "", openlaneclient.UpdateProgramMembershipInput{}, cmd.NewRequiredFieldMissingError("program id")
	}

	userID := cmd.Config.String("user-id")
	if userID == "" {
		return "", openlaneclient.UpdateProgramMembershipInput{}, cmd.NewRequiredFieldMissingError("user id")
	}

	role := cmd.Config.String("role")
	if role == "" {
		return "", openlaneclient.UpdateProgramMembershipInput{}, cmd.NewRequiredFieldMissingError("role")
	}

	enumRole, err := speccli.ParseRole(role)
	if err != nil {
		return "", openlaneclient.UpdateProgramMembershipInput{}, err
	}

	membershipID, err := lookupProgramMembership(ctx, client, programID, userID)
	if err != nil {
		return "", openlaneclient.UpdateProgramMembershipInput{}, err
	}

	input := openlaneclient.UpdateProgramMembershipInput{
		Role: &enumRole,
	}

	return membershipID, input, nil
}

func updateProgramMembership(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.UpdateProgramMembership, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	id, input, err := buildUpdateMembershipInput(ctx, client)
	if err != nil {
		return nil, err
	}

	return client.UpdateProgramMembership(ctx, id, input)
}

func deleteProgramMembership(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.DeleteProgramMembership, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	programID := cmd.Config.String("program-id")
	if programID == "" {
		return nil, cmd.NewRequiredFieldMissingError("program id")
	}

	userID := cmd.Config.String("user-id")
	if userID == "" {
		return nil, cmd.NewRequiredFieldMissingError("user id")
	}

	membershipID, err := lookupProgramMembership(ctx, client, programID, userID)
	if err != nil {
		return nil, err
	}

	return client.DeleteProgramMembership(ctx, membershipID)
}

func fetchProgramMemberships(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	id := cmd.Config.String("id")
	if id != "" {
		return client.GetProgramMembershipByID(ctx, id)
	}

	programID := cmd.Config.String("program-id")
	if programID != "" {
		where := openlaneclient.ProgramMembershipWhereInput{
			ProgramID: &programID,
		}
		return client.GetProgramMemberships(ctx, cmd.First, cmd.Last, &where)
	}

	return client.GetAllProgramMemberships(ctx)
}
