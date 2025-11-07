//go:build cli

package groupmembers

import (
	"context"
	"errors"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildCreateInput() (openlaneclient.CreateGroupMembershipInput, error) {
	var input openlaneclient.CreateGroupMembershipInput

	input.GroupID = cmd.Config.String("group-id")
	if input.GroupID == "" {
		return input, cmd.NewRequiredFieldMissingError("group id")
	}

	input.UserID = cmd.Config.String("user-id")
	if input.UserID == "" {
		return input, cmd.NewRequiredFieldMissingError("user id")
	}

	role := cmd.Config.String("role")
	if role != "" {
		enumRole, err := speccli.ParseRole(role)
		if err != nil {
			return input, err
		}
		input.Role = &enumRole
	}

	return input, nil
}

func createGroupMembership(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.CreateGroupMembership, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	input, err := buildCreateInput()
	if err != nil {
		return nil, err
	}

	return client.CreateGroupMembership(ctx, input)
}

func buildUpdateInput(ctx context.Context, client *openlaneclient.OpenlaneClient) (string, openlaneclient.UpdateGroupMembershipInput, error) {
	groupID := cmd.Config.String("group-id")
	if groupID == "" {
		return "", openlaneclient.UpdateGroupMembershipInput{}, cmd.NewRequiredFieldMissingError("group id")
	}

	userID := cmd.Config.String("user-id")
	if userID == "" {
		return "", openlaneclient.UpdateGroupMembershipInput{}, cmd.NewRequiredFieldMissingError("user id")
	}

	role := cmd.Config.String("role")
	if role == "" {
		return "", openlaneclient.UpdateGroupMembershipInput{}, cmd.NewRequiredFieldMissingError("role")
	}

	enumRole, err := speccli.ParseRole(role)
	if err != nil {
		return "", openlaneclient.UpdateGroupMembershipInput{}, err
	}

	membershipID, err := lookupGroupMembershipID(ctx, client, groupID, userID)
	if err != nil {
		return "", openlaneclient.UpdateGroupMembershipInput{}, err
	}

	input := openlaneclient.UpdateGroupMembershipInput{
		Role: &enumRole,
	}

	return membershipID, input, nil
}

func updateGroupMembership(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.UpdateGroupMembership, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	id, input, err := buildUpdateInput(ctx, client)
	if err != nil {
		return nil, err
	}

	return client.UpdateGroupMembership(ctx, id, input)
}

func deleteGroupMembership(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.DeleteGroupMembership, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	groupID := cmd.Config.String("group-id")
	if groupID == "" {
		return nil, cmd.NewRequiredFieldMissingError("group id")
	}

	userID := cmd.Config.String("user-id")
	if userID == "" {
		return nil, cmd.NewRequiredFieldMissingError("user id")
	}

	membershipID, err := lookupGroupMembershipID(ctx, client, groupID, userID)
	if err != nil {
		return nil, err
	}

	return client.DeleteGroupMembership(ctx, membershipID)
}

func fetchGroupMemberships(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	id := cmd.Config.String("id")
	if id != "" {
		return client.GetGroupMembershipByID(ctx, id)
	}

	groupID := cmd.Config.String("group-id")
	if groupID != "" {
		where := openlaneclient.GroupMembershipWhereInput{
			GroupID: &groupID,
		}
		return client.GetGroupMemberships(ctx, cmd.First, cmd.Last, &where)
	}

	return client.GetAllGroupMemberships(ctx)
}

func lookupGroupMembershipID(ctx context.Context, client *openlaneclient.OpenlaneClient, groupID, userID string) (string, error) {
	where := openlaneclient.GroupMembershipWhereInput{
		GroupID: &groupID,
		UserID:  &userID,
	}

	result, err := client.GetGroupMemberships(ctx, cmd.First, cmd.Last, &where)
	if err != nil {
		return "", err
	}

	if len(result.GroupMemberships.Edges) != 1 || result.GroupMemberships.Edges[0].Node == nil {
		return "", errors.New("error getting existing relation")
	}

	return result.GroupMemberships.Edges[0].Node.ID, nil
}
