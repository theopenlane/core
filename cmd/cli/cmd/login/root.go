package login

import (
	"context"
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/term"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var command = &cobra.Command{
	Use:   "login",
	Short: "authenticate with the API using password credentials",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := login(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	cmd.RootCmd.AddCommand(command)

	command.Flags().StringP("username", "u", "", "username (email) to authenticate with password auth")
}

func login(ctx context.Context) (*oauth2.Token, error) {
	// setup http client
	client, err := cmd.SetupClient(ctx)
	cobra.CheckErr(err)

	username := cmd.Config.String("username")
	if username == "" {
		return nil, cmd.NewRequiredFieldMissingError("username")
	}

	tokens, err := passwordAuth(ctx, client, username)
	cobra.CheckErr(err)

	fmt.Println("\nAuthentication Successful!")

	err = cmd.StoreToken(tokens)
	cobra.CheckErr(err)

	cmd.StoreSessionCookies(client)

	fmt.Println("auth tokens successfully stored in keychain")

	return tokens, nil
}

func passwordAuth(ctx context.Context, client *openlaneclient.OpenLaneClient, username string) (*oauth2.Token, error) {
	// read password from terminal if not set in environment variable
	password := cmd.Config.String("password")

	if password == "" {
		fmt.Print("Password: ")

		bytepw, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, err
		}

		password = string(bytepw)
	}

	login := models.LoginRequest{
		Username: username,
		Password: password,
	}

	resp, err := client.Login(ctx, &login)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    resp.TokenType,
	}, nil
}
