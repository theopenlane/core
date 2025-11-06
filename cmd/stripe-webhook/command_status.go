//go:build clistripe

package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/stripe/stripe-go/v83"
	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/utils/cli/tables"
)

// statusCommand is the CLI action that shows migration progress for webhooks
func statusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "show webhook migration status",
		Action: func(ctx context.Context, c *cli.Command) error {
			client, err := getStripeClient(c.Root())
			if err != nil {
				return err
			}

			repoRoot := getRepoRoot(c.Root())
			currentVersion, _, err := readDefaultAPIVersions(repoRoot)
			if err != nil {
				return err
			}

			webhookURL := getWebhookURL(c.Root())
			if webhookURL != "" {
				state, err := client.GetWebhookMigrationState(ctx, webhookURL,
					entitlements.WithCurrentVersion(currentVersion))
				if err != nil {
					return err
				}
				return printMigrationStatus(webhookURL, state)
			}

			endpoints, err := client.ListWebhookEndpoints(ctx)
			if err != nil {
				return err
			}

			bases := uniqueWebhookBases(endpoints)
			if len(bases) == 0 {
				fmt.Fprintln(outWriter, "No Stripe webhook endpoints found.")
				return nil
			}

			for i, base := range bases {
				state, err := client.GetWebhookMigrationState(ctx, base,
					entitlements.WithCurrentVersion(currentVersion))
				if err != nil {
					return err
				}
				if err := printMigrationStatus(base, state); err != nil {
					return err
				}
				if i < len(bases)-1 {
					fmt.Fprintln(outWriter)
				}
			}

			return nil
		},
	}
}

// printMigrationStatus is the report formatter for a single webhook base URL
func printMigrationStatus(target string, state *entitlements.WebhookMigrationState) error {
	fmt.Fprintf(outWriter, "\nWebhook Migration Status (%s)\n", target)
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
	fmt.Fprintf(outWriter, "Next Action: %s\n", nextAction)

	return nil
}

// uniqueWebhookBases is the helper that deduplicates webhook endpoints by base URL
func uniqueWebhookBases(endpoints []*stripe.WebhookEndpoint) []string {
	seen := make(map[string]bool)
	var bases []string

	for _, endpoint := range endpoints {
		base := cleanBaseURL(endpoint.URL)
		if base == "" {
			continue
		}
		if !seen[base] {
			seen[base] = true
			bases = append(bases, base)
		}
	}

	return bases
}

// cleanBaseURL is the sanitizer that strips query and fragment data from a URL
func cleanBaseURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String()
}
