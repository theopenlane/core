//go:build cli

package reset

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	models "github.com/theopenlane/core/pkg/openapi"
)

var command = &cobra.Command{
	Use:   "reset",
	Short: "reset a user password",
	Run: func(cmd *cobra.Command, args []string) {
		err := resetPassword(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)

	command.Flags().StringP("token", "t", "", "reset token")
	command.Flags().StringP("password", "p", "", "new password of the user")
}

// validateReset validates the required fields for the command
func validateReset() (*models.ResetPasswordRequest, error) {
	password := cmd.Config.String("password")
	if password == "" {
		return nil, cmd.NewRequiredFieldMissingError("password")
	}

	token := cmd.Config.String("token")
	if token == "" {
		return nil, cmd.NewRequiredFieldMissingError("token")
	}

	return &models.ResetPasswordRequest{
		Password: password,
		Token:    token,
	}, nil
}

func resetPassword(ctx context.Context) error {
	// setup http client
	client, err := cmd.SetupClient(ctx)
	cobra.CheckErr(err)

	input, err := validateReset()
	cobra.CheckErr(err)

	reply, err := client.ResetPassword(ctx, input)
	cobra.CheckErr(err)

	s, err := json.Marshal(reply)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}
