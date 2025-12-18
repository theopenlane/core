//go:build clistripe

package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/stripe/stripe-go/v84"
	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/entitlements"
)

const (
	stepCreate   = "create"
	stepDisable  = "disable"
	stepRollback = "rollback"
)

var (
	// ErrAPIVersionMatchesConfig indicates the new version matches the existing configuration.
	ErrAPIVersionMatchesConfig = errors.New("new API version matches current configuration")
	// ErrDiscardVersionNotConfigured indicates the discard version is missing in configuration.
	ErrDiscardVersionNotConfigured = errors.New("discard API version is not configured")
)

// migrateCommand is the CLI entry point for managing webhook migrations
func migrateCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "guides the webhook API version migration workflow",
		Description: `Manages the migration of Stripe webhooks to a new API version.

The migration flow is:
1. Create a new webhook and enable it for dual delivery (both old and new active)
2. Deploy code that accepts both API versions
3. Disable the old webhook

Use --step to execute specific migration stages, or run without flags for guidance.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "step",
				Usage: "specific migration step: create, disable, rollback",
			},
			&cli.BoolFlag{
				Name:  "auto",
				Usage: "automatically execute the next migration step",
			},
			&cli.StringSliceFlag{
				Name:  "events",
				Usage: "webhook events to subscribe to (defaults to current webhook events)",
			},
			&cli.StringFlag{
				Name:  "new-version",
				Usage: "explicit Stripe API version to migrate to (bypasses interactive prompt)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			client, err := getStripeClient(c.Root())
			if err != nil {
				return err
			}

			webhookURL := getWebhookURL(c.Root())
			if webhookURL == "" {
				return entitlements.ErrWebhookURLRequired
			}

			step := c.String("step")
			auto := c.Bool("auto")
			opts := migrationOptions{
				Events:     c.StringSlice("events"),
				NewVersion: c.String("new-version"),
				RepoRoot:   getRepoRoot(c.Root()),
			}

			if step != "" {
				return executeStep(ctx, client, webhookURL, step, opts)
			}

			if auto {
				return executeNextStep(ctx, client, webhookURL, opts)
			}

			return showGuidance(ctx, client, webhookURL, opts)
		},
	}
}

// executeStep is the dispatcher for explicitly requested migration steps
func executeStep(ctx context.Context, client *entitlements.StripeClient, webhookURL, step string, opts migrationOptions) error {
	switch step {
	case stepCreate:
		return createNewWebhook(ctx, client, webhookURL, opts)
	case stepDisable:
		return disableOldWebhook(ctx, client, webhookURL, opts)
	case stepRollback:
		return rollbackMigration(ctx, client, webhookURL, opts)
	default:
		return fmt.Errorf("%w: %s (valid: create, disable, rollback)", ErrUnknownMigrationStep, step)
	}
}

// executeNextStep is the automation that runs the appropriate next migration stage
func executeNextStep(ctx context.Context, client *entitlements.StripeClient, webhookURL string, opts migrationOptions) error {
	currentVersion, _, err := readDefaultAPIVersions(opts.RepoRoot)
	if err != nil {
		return err
	}

	state, err := client.GetWebhookMigrationState(ctx, webhookURL,
		entitlements.WithCurrentVersion(currentVersion))
	if err != nil {
		return err
	}

	stage := entitlements.MigrationStage(state.MigrationStage)

	switch stage {
	case entitlements.MigrationStageReady:
		return createNewWebhook(ctx, client, webhookURL, opts)
	case entitlements.MigrationStageDualProcessing:
		fmt.Fprintln(outWriter, "Manual action required: Deploy code that accepts both API versions.")
		fmt.Fprintf(outWriter, "After deployment, run: stripe-webhook migrate --step disable --webhook-url %s\n", webhookURL)
		return nil
	case entitlements.MigrationStageComplete:
		fmt.Fprintln(outWriter, "Migration complete. No further automated actions needed.")
		return nil
	case entitlements.MigrationStageNone:
		fmt.Fprintln(outWriter, "No migration needed - webhook already matches the current SDK version.")
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnknownMigrationStage, state.MigrationStage)
	}
}

// showGuidance is the interactive helper that prints migration recommendations
func showGuidance(ctx context.Context, client *entitlements.StripeClient, webhookURL string, opts migrationOptions) error {
	currentVersion, _, err := readDefaultAPIVersions(opts.RepoRoot)
	if err != nil {
		return err
	}

	state, err := client.GetWebhookMigrationState(ctx, webhookURL,
		entitlements.WithCurrentVersion(currentVersion))
	if err != nil {
		return err
	}

	fmt.Fprintf(outWriter, "\nWebhook Migration Guidance\n")
	fmt.Fprintf(outWriter, "Current SDK Version: %s\n", stripe.APIVersion)
	fmt.Fprintf(outWriter, "Migration Stage: %s\n", state.MigrationStage)
	fmt.Fprintf(outWriter, "Can Migrate: %t\n\n", state.CanMigrate)

	nextAction := entitlements.GetNextMigrationAction(state.MigrationStage)
	fmt.Fprintf(outWriter, "Next Action: %s\n\n", nextAction)

	stage := entitlements.MigrationStage(state.MigrationStage)

	switch stage {
	case entitlements.MigrationStageReady:
		fmt.Fprintln(outWriter, "To create new webhook: stripe-webhook migrate --step create")
		fmt.Fprintln(outWriter, "To auto-execute: stripe-webhook migrate --auto")
	case entitlements.MigrationStageDualProcessing:
		fmt.Fprintln(outWriter, "After deploying code that accepts both API versions:")
		fmt.Fprintf(outWriter, "  stripe-webhook migrate --step disable --webhook-url %s\n", webhookURL)
		fmt.Fprintln(outWriter, "To rollback: stripe-webhook migrate --step rollback")
	case entitlements.MigrationStageComplete:
		fmt.Fprintln(outWriter, "Migration complete. No additional CLI steps are required.")
	case entitlements.MigrationStageNone:
		fmt.Fprintln(outWriter, "No action needed.")
	}

	return nil
}

// createNewWebhook is the flow that provisions a new webhook and updates defaults
func createNewWebhook(ctx context.Context, client *entitlements.StripeClient, webhookURL string, opts migrationOptions) error {
	fmt.Fprintf(outWriter, "Preparing webhook migration for URL: %s\n", webhookURL)

	oldConfigVersion, _, err := readDefaultAPIVersions(opts.RepoRoot)
	if err != nil {
		return err
	}

	newVersion := strings.TrimSpace(opts.NewVersion)
	if newVersion == "" {
		input, err := promptForNewAPIVersion(oldConfigVersion, stripe.APIVersion)
		if err != nil {
			return err
		}
		newVersion = input
	} else {
		fmt.Fprintf(outWriter, "\nUsing provided new API version: %s\n", newVersion)
	}

	if err := validateAPIVersion(newVersion); err != nil {
		return err
	}

	if oldConfigVersion != "" && oldConfigVersion == newVersion {
		return fmt.Errorf("new API version %s matches the current config version: %w", newVersion, ErrAPIVersionMatchesConfig)
	}

	proceed, err := confirmAction(fmt.Sprintf("Proceed with migrating from %s to %s", displayValue(oldConfigVersion), newVersion))
	if err != nil {
		return err
	}
	if !proceed {
		return ErrMigrationAborted
	}

	changed, err := updateAPIVersionDefaults(opts.RepoRoot, oldConfigVersion, newVersion)
	if err != nil {
		return err
	}

	if changed {
		fmt.Fprintf(outWriter, "\nUpdated default API versions in %s\n", filepath.Join(opts.RepoRoot, "pkg", "entitlements", "config.go"))
	} else {
		fmt.Fprintf(outWriter, "\nDefault API version values already reflect %s (current) and %s (discard)\n", newVersion, displayValue(oldConfigVersion))
	}

	endpoint, err := client.CreateNewWebhookForMigration(ctx, webhookURL, opts.Events, newVersion)
	if err != nil {
		return err
	}

	fmt.Fprintf(outWriter, "\nNew webhook created and enabled\n")
	fmt.Fprintf(outWriter, "ID: %s\n", endpoint.ID)
	fmt.Fprintf(outWriter, "API Version: %s\n", endpoint.APIVersion)
	fmt.Fprintf(outWriter, "Status: %s\n", endpoint.Status)
	fmt.Fprintf(outWriter, "Secret: %s\n", endpoint.Secret)

	envKey := secretEnvKeyForVersion(newVersion)
	fmt.Fprintf(outWriter, "\nStore the secret above in your secret manager using the environment key %s\n", envKey)

	fmt.Fprintln(outWriter, "\nNext steps:")
	fmt.Fprintf(outWriter, "  1. Run `task config:generate` to refresh configuration artifacts with the new defaults.\n")
	fmt.Fprintf(outWriter, "  2. Commit the regenerated files (e.g., pkg/entitlements/config.go and config outputs).\n")
	fmt.Fprintf(outWriter, "  3. Populate %s in your secret manager before deploying the new code.\n", envKey)
	fmt.Fprintf(outWriter, "  4. Deploy code that accepts both API versions.\n")
	fmt.Fprintf(outWriter, "  5. After deployment, disable the legacy webhook:\n")
	fmt.Fprintf(outWriter, "     stripe-webhook migrate --step disable --webhook-url %s\n", webhookURL)

	return nil
}

// disableOldWebhook is the step that turns off the legacy webhook using the discard version
func disableOldWebhook(ctx context.Context, client *entitlements.StripeClient, webhookURL string, opts migrationOptions) error {
	fmt.Fprintln(outWriter, "Disabling old webhook endpoint")

	_, discardVersion, err := readDefaultAPIVersions(opts.RepoRoot)
	if err != nil {
		return err
	}

	if discardVersion == "" {
		return fmt.Errorf("no discard version configured in pkg/entitlements/config.go: %w", ErrDiscardVersionNotConfigured)
	}

	endpoint, err := client.DisableWebhookByVersion(ctx, webhookURL, discardVersion)
	if err != nil {
		return err
	}

	fmt.Fprintf(outWriter, "Old webhook disabled successfully\n")
	fmt.Fprintf(outWriter, "ID: %s\n", endpoint.ID)
	fmt.Fprintf(outWriter, "Status: %s\n\n", endpoint.Status)
	fmt.Fprintln(outWriter, "Migration complete - only the new webhook is active")

	return nil
}

// rollbackMigration is the helper that reverts to the pre-migration webhook state
func rollbackMigration(ctx context.Context, client *entitlements.StripeClient, webhookURL string, opts migrationOptions) error {
	fmt.Fprintln(outWriter, "Rolling back migration")

	currentVersion, _, err := readDefaultAPIVersions(opts.RepoRoot)
	if err != nil {
		return err
	}

	err = client.RollbackMigration(ctx, webhookURL,
		entitlements.WithCurrentVersion(currentVersion))
	if err != nil {
		return err
	}

	fmt.Fprintln(outWriter, "Migration rolled back successfully")
	fmt.Fprintln(outWriter, "Old webhook re-enabled, new webhook disabled")

	return nil
}
