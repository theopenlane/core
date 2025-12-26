//go:build cli

package invite

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/theopenlane/cli/cmd"
	models "github.com/theopenlane/common/openapi"
)

var acceptCmd = &cobra.Command{
	Use:   "accept",
	Short: "accept an invite to join an organization",
	Run: func(cmd *cobra.Command, args []string) {
		err := accept(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(acceptCmd)

	acceptCmd.Flags().StringP("token", "t", "", "invite token")
}

// acceptValidation validates the input for the accept command
func acceptValidation() (input models.InviteRequest, err error) {
	input.Token = cmd.Config.String("token")
	if input.Token == "" {
		return input, cmd.NewRequiredFieldMissingError("token")
	}

	return input, nil
}

// accept an invite to join an organization
func accept(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)

	var s []byte

	input, err := acceptValidation()
	cobra.CheckErr(err)

	resp, err := client.AcceptInvite(ctx, &input)
	cobra.CheckErr(err)

	s, err = json.Marshal(resp)
	cobra.CheckErr(err)

	if err := cmd.StoreToken(&oauth2.Token{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}); err != nil {
		cobra.CheckErr(err)
	}

	cmd.StoreSessionCookies(client)

	return cmd.JSONPrint(s)
}
