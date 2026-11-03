//go:build cli

package login

import (
	"context"
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/term"

	"github.com/theopenlane/core/cli/cmd"
	models "github.com/theopenlane/core/common/openapi"
	openlane "github.com/theopenlane/go-client"
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

	status, err := fetchSSOStatus(ctx, client, username)
	if err == nil && status.Enforced && status.OrganizationID != "" {
		tokens, err := ssoAuth(ctx, client, status.OrganizationID)
		if err != nil {
			return nil, err
		}

		fmt.Println("\nAuthentication Successful!")
		if err := cmd.StoreToken(tokens); err == nil {
			fmt.Println("auth tokens successfully stored in keychain")
		}

		return tokens, nil
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

func passwordAuth(ctx context.Context, client *openlane.Client, username string) (*oauth2.Token, error) {
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

	// Check if 2FA is required but user hasn't enabled it
	if resp.TFASetupRequired && !resp.TFAEnabled {
		fmt.Println("\nTwo-factor authentication is required by your organization.")
		fmt.Println("Please login to the Openlane console and enable 2FA on your profile page before using the CLI.")
		return nil, fmt.Errorf("2FA required but not enabled on user account")
	}

	// If 2FA is enabled, the session is only issued after verification
	if resp.TFAEnabled {
		resp, err = handle2FAVerification(ctx, client)
		if err != nil {
			return nil, err
		}
	}

	return &oauth2.Token{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    resp.TokenType,
	}, nil
}

// handle2FAVerification prompts the user for their 2FA code and validates it, returning the issued session
func handle2FAVerification(ctx context.Context, client *openlane.Client) (*models.LoginResponse, error) {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("\nEnter your 2FA code (attempt %d of %d): ", attempt, maxAttempts)

		code, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, fmt.Errorf("failed to read 2FA code: %w", err)
		}

		codeStr := string(code)
		fmt.Println() // Add newline after password input

		if codeStr == "" {
			fmt.Println("2FA code cannot be empty")
			continue
		}

		// Create TFA request
		tfaReq := &models.TFARequest{
			TOTPCode: codeStr,
		}

		// Validate the code
		tfaResp, err := client.ValidateLoginTOTP(ctx, tfaReq)
		if err != nil {
			if attempt == maxAttempts {
				return nil, fmt.Errorf("2FA verification failed after %d attempts: %w", maxAttempts, err)
			}

			fmt.Printf("Invalid 2FA code: %v\n", err)

			continue
		}

		if !tfaResp.Success {
			if attempt == maxAttempts {
				return nil, fmt.Errorf("2FA verification failed after %d attempts", maxAttempts)
			}

			fmt.Println("Invalid 2FA code")

			continue
		}

		if tfaResp.AccessToken == "" {
			return nil, ErrNoSessionIssued
		}

		// 2FA verification successful
		fmt.Println("2FA verification successful")

		return tfaResp, nil
	}

	return nil, fmt.Errorf("2FA verification failed after %d attempts", maxAttempts)
}
