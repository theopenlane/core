package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/stripe/stripe-go/v83"
	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/utils/cli/tables"
)

var (
	outWriter io.Writer = os.Stdout
	// ErrUnknownMigrationStep is returned when an unknown migration step is requested
	ErrUnknownMigrationStep = errors.New("unknown migration step")
	// ErrUnknownMigrationStage is returned when an unknown migration stage is encountered
	ErrUnknownMigrationStage = errors.New("unknown migration stage")
)

func main() {
	if err := webhookApp().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// webhookApp creates the main webhook management CLI application
func webhookApp() *cli.Command {
	app := &cli.Command{
		Name:  "stripe-webhook",
		Usage: "manage Stripe webhook endpoints and API version migrations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "stripe-key",
				Usage:   "Stripe API key",
				Sources: cli.EnvVars("STRIPE_API_KEY", "STRIPE_SECRET_KEY"),
			},
			&cli.StringFlag{
				Name:    "webhook-url",
				Usage:   "webhook URL to manage",
				Sources: cli.EnvVars("STRIPE_WEBHOOK_URL"),
			},
		},
		Commands: []*cli.Command{
			listCommand(),
			statusCommand(),
			migrateCommand(),
		},
	}

	return app
}

// getStripeClient creates a new Stripe client from CLI context
func getStripeClient(c *cli.Command) (*entitlements.StripeClient, error) {
	apiKey := c.String("stripe-key")
	if apiKey == "" {
		return nil, entitlements.ErrMissingAPIKey
	}

	return entitlements.NewStripeClient(
		entitlements.WithAPIKey(apiKey),
	)
}

// getWebhookURL retrieves the webhook URL from CLI context
func getWebhookURL(c *cli.Command) string {
	return c.String("webhook-url")
}

// listCommand returns the list subcommand
func listCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "list all Stripe webhook endpoints",
		Action: func(ctx context.Context, c *cli.Command) error {
			client, err := getStripeClient(c.Root())
			if err != nil {
				return err
			}

			endpoints, err := client.ListWebhookEndpoints(ctx)
			if err != nil {
				return err
			}

			writer := tables.NewTableWriter(outWriter, "ID", "URL", "Status", "API Version", "Events")
			for _, endpoint := range endpoints {
				if err := writer.AddRow(
					endpoint.ID,
					endpoint.URL,
					string(endpoint.Status),
					endpoint.APIVersion,
					fmt.Sprintf("%d events", len(endpoint.EnabledEvents)),
				); err != nil {
					return err
				}
			}
			if err := writer.Render(); err != nil {
				return err
			}

			return nil
		},
	}
}

// statusCommand returns the status subcommand
func statusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "show webhook migration status",
		Action: func(ctx context.Context, c *cli.Command) error {
			client, err := getStripeClient(c.Root())
			if err != nil {
				return err
			}

			webhookURL := getWebhookURL(c.Root())
			if webhookURL == "" {
				return entitlements.ErrWebhookURLRequired
			}

			state, err := client.GetWebhookMigrationState(ctx, webhookURL)
			if err != nil {
				return err
			}

			fmt.Fprintf(outWriter, "\nWebhook Migration Status\n")
			fmt.Fprintf(outWriter, "Current SDK Version: %s\n", stripe.APIVersion)
			fmt.Fprintf(outWriter, "Migration Stage: %s\n", state.MigrationStage)
			fmt.Fprintf(outWriter, "Can Migrate: %t\n\n", state.CanMigrate)

			if state.OldWebhook != nil {
				fmt.Fprintf(outWriter, "Old Webhook:\n")
				writer := tables.NewTableWriter(outWriter, "Field", "Value")
				if err := writer.AddRow("ID", state.OldWebhook.ID); err != nil {
					return err
				}
				if err := writer.AddRow("URL", state.OldWebhook.URL); err != nil {
					return err
				}
				if err := writer.AddRow("Status", string(state.OldWebhook.Status)); err != nil {
					return err
				}
				if err := writer.AddRow("API Version", state.OldWebhook.APIVersion); err != nil {
					return err
				}
				if err := writer.AddRow("Event Count", fmt.Sprintf("%d", len(state.OldWebhook.EnabledEvents))); err != nil {
					return err
				}
				if err := writer.Render(); err != nil {
					return err
				}
				fmt.Fprintln(outWriter)
			}

			if state.NewWebhook != nil {
				fmt.Fprintf(outWriter, "New Webhook:\n")
				writer := tables.NewTableWriter(outWriter, "Field", "Value")
				if err := writer.AddRow("ID", state.NewWebhook.ID); err != nil {
					return err
				}
				if err := writer.AddRow("URL", state.NewWebhook.URL); err != nil {
					return err
				}
				if err := writer.AddRow("Status", string(state.NewWebhook.Status)); err != nil {
					return err
				}
				if err := writer.AddRow("API Version", state.NewWebhook.APIVersion); err != nil {
					return err
				}
				if err := writer.AddRow("Event Count", fmt.Sprintf("%d", len(state.NewWebhook.EnabledEvents))); err != nil {
					return err
				}
				if err := writer.Render(); err != nil {
					return err
				}
				fmt.Fprintln(outWriter)
			}

			nextAction := entitlements.GetNextMigrationAction(state.MigrationStage)
			fmt.Fprintf(outWriter, "Next Action: %s\n\n", nextAction)

			return nil
		},
	}
}

// migrateCommand returns the migrate subcommand
func migrateCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "automated webhook API version migration",
		Description: `Manages the migration of Stripe webhooks to a new API version.

The migration follows these stages:
1. Create new webhook endpoint (disabled) with current SDK API version
2. Enable new webhook to begin dual-processing
3. Update code to process only new version events
4. Disable old webhook to complete migration
5. Cleanup old webhook after verification

Use --step to execute specific migration steps, or run without flags for automated guidance.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "step",
				Usage: "specific migration step: create, enable, disable, cleanup, rollback, promote",
			},
			&cli.BoolFlag{
				Name:  "auto",
				Usage: "automatically execute next migration step",
			},
			&cli.StringSliceFlag{
				Name:  "events",
				Usage: "webhook events to subscribe to (defaults to current webhook events)",
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
			events := c.StringSlice("events")

			if step != "" {
				return executeStep(ctx, client, webhookURL, step, events)
			}

			if auto {
				return executeNextStep(ctx, client, webhookURL, events)
			}

			return showGuidance(ctx, client, webhookURL)
		},
	}
}

const (
	stepCreate   = "create"
	stepEnable   = "enable"
	stepDisable  = "disable"
	stepCleanup  = "cleanup"
	stepRollback = "rollback"
	stepPromote  = "promote"
)

// executeStep executes a specific migration step
func executeStep(ctx context.Context, client *entitlements.StripeClient, webhookURL, step string, events []string) error {
	switch step {
	case stepCreate:
		return createNewWebhook(ctx, client, webhookURL, events)
	case stepEnable:
		return enableNewWebhook(ctx, client, webhookURL)
	case stepDisable:
		return disableOldWebhook(ctx, client, webhookURL)
	case stepCleanup:
		return cleanupOldWebhook(ctx, client, webhookURL)
	case stepRollback:
		return rollbackMigration(ctx, client, webhookURL)
	case stepPromote:
		return promoteNewWebhook(ctx, client, webhookURL)
	default:
		return fmt.Errorf("%w: %s (valid: create, enable, disable, cleanup, rollback, promote)", ErrUnknownMigrationStep, step)
	}
}

// executeNextStep automatically executes the next migration step
func executeNextStep(ctx context.Context, client *entitlements.StripeClient, webhookURL string, events []string) error {
	state, err := client.GetWebhookMigrationState(ctx, webhookURL)
	if err != nil {
		return err
	}

	stage := entitlements.MigrationStage(state.MigrationStage)

	switch stage {
	case entitlements.MigrationStageReady:
		return createNewWebhook(ctx, client, webhookURL, events)
	case entitlements.MigrationStageNewCreated:
		return enableNewWebhook(ctx, client, webhookURL)
	case entitlements.MigrationStageDualProcessing:
		fmt.Fprintln(outWriter, "Manual action required: Update your code to process only new version events")
		fmt.Fprintln(outWriter, "After code is deployed, run: stripe-webhook migrate --step disable")
		return nil
	case entitlements.MigrationStageTransitioned:
		fmt.Fprintln(outWriter, "Monitor the new webhook for stability")
		fmt.Fprintln(outWriter, "When ready to cleanup, run: stripe-webhook migrate --step cleanup")
		return nil
	case entitlements.MigrationStageComplete:
		fmt.Fprintln(outWriter, "Migration complete")
		fmt.Fprintln(outWriter, "Optional: run stripe-webhook migrate --step promote to remove version query parameter")
		return nil
	case entitlements.MigrationStageNone:
		fmt.Fprintln(outWriter, "No migration needed - webhook is already at current SDK version")
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnknownMigrationStage, state.MigrationStage)
	}
}

// showGuidance displays migration guidance without executing steps
func showGuidance(ctx context.Context, client *entitlements.StripeClient, webhookURL string) error {
	state, err := client.GetWebhookMigrationState(ctx, webhookURL)
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
	case entitlements.MigrationStageNewCreated:
		fmt.Fprintln(outWriter, "To enable new webhook: stripe-webhook migrate --step enable")
		fmt.Fprintln(outWriter, "To rollback: stripe-webhook migrate --step rollback")
	case entitlements.MigrationStageDualProcessing:
		fmt.Fprintln(outWriter, "After code update: stripe-webhook migrate --step disable")
		fmt.Fprintln(outWriter, "To rollback: stripe-webhook migrate --step rollback")
	case entitlements.MigrationStageTransitioned:
		fmt.Fprintln(outWriter, "To cleanup old webhook: stripe-webhook migrate --step cleanup")
		fmt.Fprintln(outWriter, "To rollback: stripe-webhook migrate --step rollback")
	case entitlements.MigrationStageComplete:
		fmt.Fprintln(outWriter, "To promote new webhook: stripe-webhook migrate --step promote")
	case entitlements.MigrationStageNone:
		fmt.Fprintln(outWriter, "No action needed")
	}

	return nil
}

// createNewWebhook creates a new webhook endpoint for migration
func createNewWebhook(ctx context.Context, client *entitlements.StripeClient, webhookURL string, events []string) error {
	fmt.Fprintf(outWriter, "Creating new webhook endpoint for URL: %s\n", webhookURL)

	endpoint, err := client.CreateNewWebhookForMigration(ctx, webhookURL, events)
	if err != nil {
		return err
	}

	fmt.Fprintf(outWriter, "New webhook created successfully\n")
	fmt.Fprintf(outWriter, "ID: %s\n", endpoint.ID)
	fmt.Fprintf(outWriter, "URL: %s\n", endpoint.URL)
	fmt.Fprintf(outWriter, "API Version: %s\n", endpoint.APIVersion)
	fmt.Fprintf(outWriter, "Status: %s\n", endpoint.Status)
	fmt.Fprintf(outWriter, "Secret: %s\n\n", endpoint.Secret)
	fmt.Fprintln(outWriter, "IMPORTANT: Save the webhook secret above - you will need to update your configuration")
	fmt.Fprintln(outWriter, "Next step: stripe-webhook migrate --step enable")

	return nil
}

// enableNewWebhook enables the new webhook endpoint
func enableNewWebhook(ctx context.Context, client *entitlements.StripeClient, webhookURL string) error {
	fmt.Fprintln(outWriter, "Enabling new webhook endpoint for dual-processing")

	endpoint, err := client.EnableNewWebhook(ctx, webhookURL)
	if err != nil {
		return err
	}

	fmt.Fprintf(outWriter, "New webhook enabled successfully\n")
	fmt.Fprintf(outWriter, "ID: %s\n", endpoint.ID)
	fmt.Fprintf(outWriter, "Status: %s\n\n", endpoint.Status)
	fmt.Fprintln(outWriter, "Both old and new webhooks are now active")
	fmt.Fprintln(outWriter, "Next: Update your code to process only new version events")
	fmt.Fprintln(outWriter, "After deployment: stripe-webhook migrate --step disable")

	return nil
}

// disableOldWebhook disables the old webhook endpoint
func disableOldWebhook(ctx context.Context, client *entitlements.StripeClient, webhookURL string) error {
	fmt.Fprintln(outWriter, "Disabling old webhook endpoint")

	endpoint, err := client.DisableOldWebhook(ctx, webhookURL)
	if err != nil {
		return err
	}

	fmt.Fprintf(outWriter, "Old webhook disabled successfully\n")
	fmt.Fprintf(outWriter, "ID: %s\n", endpoint.ID)
	fmt.Fprintf(outWriter, "Status: %s\n\n", endpoint.Status)
	fmt.Fprintln(outWriter, "Migration transitioned - only new webhook is active")
	fmt.Fprintln(outWriter, "Monitor for stability, then: stripe-webhook migrate --step cleanup")

	return nil
}

// cleanupOldWebhook deletes the old webhook endpoint
func cleanupOldWebhook(ctx context.Context, client *entitlements.StripeClient, webhookURL string) error {
	fmt.Fprintln(outWriter, "Deleting old webhook endpoint")

	endpoint, err := client.CleanupOldWebhook(ctx, webhookURL)
	if err != nil {
		return err
	}

	fmt.Fprintf(outWriter, "Old webhook deleted successfully\n")
	fmt.Fprintf(outWriter, "ID: %s\n\n", endpoint.ID)
	fmt.Fprintln(outWriter, "Migration complete")
	fmt.Fprintln(outWriter, "Optional: stripe-webhook migrate --step promote to remove version query parameter")

	return nil
}

// rollbackMigration rolls back the migration
func rollbackMigration(ctx context.Context, client *entitlements.StripeClient, webhookURL string) error {
	fmt.Fprintln(outWriter, "Rolling back migration")

	err := client.RollbackMigration(ctx, webhookURL)
	if err != nil {
		return err
	}

	fmt.Fprintln(outWriter, "Migration rolled back successfully")
	fmt.Fprintln(outWriter, "Old webhook re-enabled, new webhook disabled")

	return nil
}

// promoteNewWebhook promotes the new webhook by removing version query parameter
func promoteNewWebhook(ctx context.Context, client *entitlements.StripeClient, webhookURL string) error {
	fmt.Fprintln(outWriter, "Promoting new webhook")

	endpoint, err := client.PromoteNewWebhook(ctx, webhookURL)
	if err != nil {
		return err
	}

	fmt.Fprintf(outWriter, "New webhook promoted successfully\n")
	fmt.Fprintf(outWriter, "ID: %s\n", endpoint.ID)
	fmt.Fprintf(outWriter, "URL: %s\n", endpoint.URL)

	return nil
}
