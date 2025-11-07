//go:build cli

package orgmembers

import (
	"context"
	"errors"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildCreateOrgMembershipInput() (openlaneclient.CreateOrgMembershipInput, error) {
	var input openlaneclient.CreateOrgMembershipInput

	input.UserID = cmd.Config.String("user-id")
	if input.UserID == "" {
		return input, cmd.NewRequiredFieldMissingError("user id")
	}

	if orgID := cmd.Config.String("org-id"); orgID != "" {
		input.OrganizationID = orgID
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

func createOrgMembership(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.CreateOrgMembership, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	input, err := buildCreateOrgMembershipInput()
	if err != nil {
		return nil, err
	}

	return client.CreateOrgMembership(ctx, input)
}

func lookupOrgMembership(ctx context.Context, client *openlaneclient.OpenlaneClient, userID, orgID string) (string, error) {
	where := openlaneclient.OrgMembershipWhereInput{
		UserID: &userID,
	}

	if orgID != "" {
		where.OrganizationID = &orgID
	}

	result, err := client.GetOrgMemberships(ctx, cmd.First, cmd.Last, &where)
	if err != nil {
		return "", err
	}

	if len(result.OrgMemberships.Edges) != 1 || result.OrgMemberships.Edges[0].Node == nil {
		return "", errors.New("error getting existing relation")
	}

	return result.OrgMemberships.Edges[0].Node.ID, nil
}

func buildUpdateOrgMembershipInput(ctx context.Context, client *openlaneclient.OpenlaneClient) (string, openlaneclient.UpdateOrgMembershipInput, error) {
	userID := cmd.Config.String("user-id")
	if userID == "" {
		return "", openlaneclient.UpdateOrgMembershipInput{}, cmd.NewRequiredFieldMissingError("user id")
	}

	orgID := cmd.Config.String("org-id")
	role := cmd.Config.String("role")
	if role == "" {
		return "", openlaneclient.UpdateOrgMembershipInput{}, cmd.NewRequiredFieldMissingError("role")
	}

	enumRole, err := speccli.ParseRole(role)
	if err != nil {
		return "", openlaneclient.UpdateOrgMembershipInput{}, err
	}

	membershipID, err := lookupOrgMembership(ctx, client, userID, orgID)
	if err != nil {
		return "", openlaneclient.UpdateOrgMembershipInput{}, err
	}

	input := openlaneclient.UpdateOrgMembershipInput{
		Role: &enumRole,
	}

	return membershipID, input, nil
}

func updateOrgMembership(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.UpdateOrgMembership, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	id, input, err := buildUpdateOrgMembershipInput(ctx, client)
	if err != nil {
		return nil, err
	}

	return client.UpdateOrgMembership(ctx, id, input)
}

func deleteOrgMembership(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.DeleteOrgMembership, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	userID := cmd.Config.String("user-id")
	if userID == "" {
		return nil, cmd.NewRequiredFieldMissingError("user id")
	}

	orgID := cmd.Config.String("org-id")

	membershipID, err := lookupOrgMembership(ctx, client, userID, orgID)
	if err != nil {
		return nil, err
	}

	return client.DeleteOrgMembership(ctx, membershipID)
}

func fetchOrgMemberships(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}

	userID := cmd.Config.String("user-id")
	orgID := cmd.Config.String("org-id")

	if userID == "" && orgID == "" {
		return client.GetOrgMemberships(ctx, cmd.First, cmd.Last, nil)
	}

	where := openlaneclient.OrgMembershipWhereInput{}
	if userID != "" {
		where.UserID = &userID
	}
	if orgID != "" {
		where.OrganizationID = &orgID
	}

	return client.GetOrgMemberships(ctx, cmd.First, cmd.Last, &where)
}
