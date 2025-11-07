//go:build cli

package invite

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildCreateInvite() (openlaneclient.CreateInviteInput, error) {
	var input openlaneclient.CreateInviteInput

	input.Recipient = cmd.Config.String("email")
	if input.Recipient == "" {
		return input, speccli.RequiredFieldMissing("email")
	}

	role := cmd.Config.String("role")
	if role == "" {
		role = "member"
	}

	enumRole, err := speccli.ParseRole(role)
	if err != nil {
		return input, err
	}

	input.Role = enumRole

	return input, nil
}

func buildDeleteInvite() (string, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", speccli.RequiredFieldMissing("id")
	}

	return id, nil
}

func buildAcceptInvite() (models.InviteRequest, error) {
	var input models.InviteRequest

	input.Token = cmd.Config.String("token")
	if input.Token == "" {
		return input, speccli.RequiredFieldMissing("token")
	}

	return input, nil
}

func createInvite(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.CreateInvite, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	input, err := buildCreateInvite()
	if err != nil {
		return nil, err
	}

	return client.CreateInvite(ctx, input)
}

func deleteInvite(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.DeleteInvite, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id, err := buildDeleteInvite()
	if err != nil {
		return nil, err
	}

	return client.DeleteInvite(ctx, id)
}

func acceptInvite(ctx context.Context) ([]byte, error) {
	client, err := cmd.SetupClientWithAuth(ctx)
	if err != nil {
		return nil, err
	}

	input, err := buildAcceptInvite()
	if err != nil {
		return nil, err
	}

	resp, err := client.AcceptInvite(ctx, &input)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	if err := cmd.StoreToken(&oauth2.Token{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}); err != nil {
		return nil, err
	}

	cmd.StoreSessionCookies(client)

	return payload, nil
}
