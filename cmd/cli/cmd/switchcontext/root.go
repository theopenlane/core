//go:build cli

package switchcontext

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildSwitchRequest() (*models.SwitchOrganizationRequest, error) {
	target := cmdpkg.Config.String("target-org")
	if target == "" {
		return nil, cmdpkg.NewRequiredFieldMissingError("target organization")
	}

	return &models.SwitchOrganizationRequest{TargetOrganizationID: target}, nil
}

func switchOrganization(ctx context.Context, client *openlaneclient.OpenlaneClient) (*models.SwitchOrganizationReply, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	input, err := buildSwitchRequest()
	if err != nil {
		return nil, err
	}

	resp, err := client.Switch(ctx, input)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}
	if err := cmdpkg.StoreToken(token); err != nil {
		return nil, err
	}

	cmdpkg.StoreSessionCookies(client)

	return resp, nil
}
