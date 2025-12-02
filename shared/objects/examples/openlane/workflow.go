//go:build examples

package openlane

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/theopenlane/shared/objects/examples/config"
)

// SetupWorkflow performs the complete setup workflow and persists the configuration
func SetupWorkflow(ctx context.Context, out io.Writer, baseURL *url.URL) (*config.OpenlaneConfig, error) {
	email, password := generateCredentials()

	fmt.Fprintln(out, "Registering new user...")
	token, err := RegisterUser(ctx, baseURL, email, password, "Example", "User")
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}

	fmt.Fprintln(out, "Verifying user...")
	if err := VerifyUser(ctx, baseURL, token); err != nil {
		return nil, fmt.Errorf("verify user: %w", err)
	}

	fmt.Fprintln(out, "Logging in...")
	client, err := LoginUser(ctx, baseURL, email, password)
	if err != nil {
		return nil, fmt.Errorf("login user: %w", err)
	}

	fmt.Fprintln(out, "Creating organization...")
	orgID, err := CreateOrganization(ctx, client, "Examples Organization", "Organization for running object storage examples")
	if err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}

	fmt.Fprintln(out, "Creating personal access token...")
	pat, err := CreatePAT(ctx, client, orgID, "Examples PAT", "Personal Access Token for examples")
	if err != nil {
		return nil, fmt.Errorf("create PAT: %w", err)
	}

	cfg := &config.OpenlaneConfig{
		BaseURL:        baseURL.String(),
		Email:          email,
		Password:       password,
		Token:          token,
		OrganizationID: orgID,
		PAT:            pat,
	}

	fmt.Fprintln(out, "Saving configuration...")
	if err := config.SaveOpenlaneConfig(cfg); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	fmt.Fprintln(out, "Setup completed successfully")
	return cfg, nil
}

// EnsureSetup ensures that a valid configuration exists, creating one if necessary
func EnsureSetup(ctx context.Context, out io.Writer, baseURL *url.URL) (*config.OpenlaneConfig, error) {
	if config.ConfigExists() {
		cfg, err := config.Load(nil)
		if err != nil {
			return nil, fmt.Errorf("load existing config: %w", err)
		}
		return &cfg.Openlane, nil
	}

	return SetupWorkflow(ctx, out, baseURL)
}

func generateCredentials() (email, password string) {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)

	email = fmt.Sprintf("example-%d-%s@theopenlane.io", timestamp, randomHex[:8])
	password = generatePassword()

	return email, password
}

func generatePassword() string {
	passwordBytes := make([]byte, 16)
	rand.Read(passwordBytes)
	return hex.EncodeToString(passwordBytes)
}
