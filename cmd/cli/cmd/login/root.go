//go:build cli

package login

import (
	"context"
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/term"

	"github.com/theopenlane/core/cmd/cli/cmd"
	goclient "github.com/theopenlane/go-client"
	models "github.com/theopenlane/shared/openapi"
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
		// Check if 2FA is required but user hasn't enabled it for SSO
		if status.OrgTFAEnforced && !status.UserTFAEnabled {
			fmt.Println("\nTwo-factor authentication is required by your organization.")
			fmt.Println("Please login to the Openlane console and enable 2FA on your profile page before using the CLI.")
			return nil, fmt.Errorf("2FA required but not enabled on user account")
		}

		tokens, err := ssoAuth(ctx, client, status.OrganizationID)
		if err != nil {
			return nil, err
		}

		// If 2FA is enabled for the user, require verification after SSO login
		if status.UserTFAEnabled {
			if err := handle2FAVerification(ctx, client); err != nil {
				return nil, err
			}
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

func passwordAuth(ctx context.Context, client *goclient.OpenlaneClient, username string) (*oauth2.Token, error) {
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

	// Extract tokens from initial login response
	tokens := &oauth2.Token{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    resp.TokenType,
	}

	// If 2FA is enabled, require verification before allowing CLI usage
	if resp.TFAEnabled {
		if err := handle2FAVerification(ctx, client); err != nil {
			return nil, err
		}
	}

	return tokens, nil
}

// handle2FAVerification prompts the user for their 2FA code and validates it
func handle2FAVerification(ctx context.Context, client *goclient.OpenlaneClient) error {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("\nEnter your 2FA code (attempt %d of %d): ", attempt, maxAttempts)

		code, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read 2FA code: %w", err)
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
		tfaResp, err := client.ValidateTOTP(ctx, tfaReq)
		if err != nil {
			if attempt == maxAttempts {
				return fmt.Errorf("2FA verification failed after %d attempts: %w", maxAttempts, err)
			}

			fmt.Printf("Invalid 2FA code: %v\n", err)

			continue
		}

		if !tfaResp.Success {
			if attempt == maxAttempts {
				return fmt.Errorf("2FA verification failed after %d attempts", maxAttempts)
			}

			fmt.Println("Invalid 2FA code")

			continue
		}

		// 2FA verification successful
		fmt.Println("2FA verification successful")

		return nil
	}

	return fmt.Errorf("2FA verification failed after %d attempts", maxAttempts)
}
