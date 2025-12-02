//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/config"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/ent/entdb"
	"github.com/theopenlane/ent/entitlements/reconciler"
	"github.com/theopenlane/ent/generated"
	_ "github.com/theopenlane/ent/generated/runtime"
)

// main is the entry point for the reconcile CLI application
func main() {
	if err := app().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// app creates and configures the CLI application with all commands and global flags
func app() *cli.Command {
	return &cli.Command{
		Name:  "reconcile",
		Usage: "reconcile billing data with Stripe",
		Description: `Manages Stripe subscription reconciliation and analysis.

Examples:
  # Dry-run by default (preview changes without making them)
  reconcile orgs
  reconcile update-cancel-behavior

  # Actually make changes
  reconcile orgs --dry-run=false
  reconcile create-schedules --dry-run=false

  # Use custom config
  reconcile --config /path/to/config.yaml orgs

  # Use production catalog (default is sandbox)
  reconcile report-missing-products --catalog ./pkg/catalog/catalog.yaml`,
		Commands: []*cli.Command{
			{
				Name:   "orgs",
				Usage:  "reconcile organization subscriptions",
				Action: run,
				Flags:  []cli.Flag{},
			},
			{
				Name:   "update-cancel-behavior",
				Usage:  "update subscriptions from pause to cancel behavior",
				Action: updateCancelBehavior,
				Flags:  []cli.Flag{},
			},
			{
				Name:   "create-schedules",
				Usage:  "create subscription schedules for subscriptions without schedules",
				Action: createSchedules,
				Flags:  []cli.Flag{},
			},
			{
				Name:   "report-missing-products",
				Usage:  "report subscriptions with products not in catalog",
				Action: reportMissingProducts,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "catalog",
						Value: "./pkg/catalog/catalog_sandbox.yaml",
						Usage: "path to catalog file",
					},
					&cli.StringFlag{
						Name:  "output",
						Value: "table",
						Usage: "output format (table, json)",
					},
				},
			},
			{
				Name:   "update-personal-org-metadata",
				Usage:  "update customer metadata for personal organizations",
				Action: updatePersonalOrgMetadata,
				Flags:  []cli.Flag{},
			},
			{
				Name:   "analyze-stripe-mismatches",
				Usage:  "analyze Stripe customers vs internal system for mismatches",
				Action: analyzeStripeMismatches,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "output",
						Value: "table",
						Usage: "output format (table, json)",
					},
					&cli.StringFlag{
						Name:  "action",
						Usage: "action to take on mismatched records (cleanup-orphaned-customers)",
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Value:   "./config/.config.yaml",
				Usage:   "config file path",
				Sources: cli.EnvVars("CORE_CONFIG"),
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "preview changes without making them (global default)",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "enable debug logging",
			},
			&cli.StringSliceFlag{
				Name:  "org-ids",
				Usage: "specific organization IDs to reconcile (only valid with 'orgs' command)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return cli.ShowCommandHelp(ctx, c, "")
		},
	}
}

// createReconciler creates a reconciler with Stripe client configuration only
func createReconciler(c *cli.Command) (*reconciler.Reconciler, error) {
	cfgLoc := c.Root().String("config")
	cfg, err := config.Load(&cfgLoc)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	stripeClient, err := entitlements.NewStripeClient(
		entitlements.WithAPIKey(cfg.Entitlements.PrivateStripeKey),
		entitlements.WithConfig(cfg.Entitlements),
	)
	if err != nil {
		return nil, fmt.Errorf("stripe client: %w", err)
	}

	options := []reconciler.Option{
		reconciler.WithStripeClient(stripeClient),
	}

	if c.Root().Bool("dry-run") {
		options = append(options, reconciler.WithDryRun(nil))
	}

	return reconciler.New(options...)
}

// createReconcilerWithDB creates a reconciler with both database and Stripe clients for operations requiring database access
func createReconcilerWithDB(c *cli.Command) (*reconciler.Reconciler, error) {
	cfgLoc := c.Root().String("config")
	cfg, err := config.Load(&cfgLoc)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Validate job queue config
	if cfg.JobQueue.ConnectionURI == "" {
		return nil, fmt.Errorf("missing required job queue connection URI in config")
	}

	// Create database client
	jobOpts := []riverqueue.Option{
		riverqueue.WithConnectionURI(cfg.JobQueue.ConnectionURI),
	}

	// Create basic context for database client creation
	ctx := context.Background()

	stripeClient, err := entitlements.NewStripeClient(
		entitlements.WithAPIKey(cfg.Entitlements.PrivateStripeKey),
		entitlements.WithConfig(cfg.Entitlements),
	)
	if err != nil {
		return nil, fmt.Errorf("stripe client: %w", err)
	}

	// add the ent config options
	entOpts := []generated.Option{
		generated.EntConfig(&cfg.EntConfig),
		generated.EntitlementManager(stripeClient),
	}

	dbClient, err := entdb.New(ctx, cfg.DB, jobOpts, entOpts...)
	if err != nil {
		return nil, fmt.Errorf("database client: %w", err)
	}

	options := []reconciler.Option{
		reconciler.WithDB(dbClient),
		reconciler.WithStripeClient(stripeClient),
	}

	if c.Root().Bool("dry-run") {
		options = append(options, reconciler.WithDryRun(nil))
	}

	return reconciler.New(options...)
}

// run executes the main organization reconciliation command
func run(ctx context.Context, c *cli.Command) error {
	recon, err := createReconcilerWithDB(c)
	if err != nil {
		return err
	}

	orgIDs := c.Root().StringSlice("org-ids")

	result, err := recon.Reconcile(ctx, orgIDs)
	if err != nil {
		return err
	}

	if len(result.Actions) == 0 {
		fmt.Println("No organizations require reconciliation actions")
		return nil
	}

	fmt.Printf("Found %d organizations requiring reconciliation\n", len(result.Actions))

	// Output as table
	printOrgsTable([]string{"ORGANIZATION_ID", "ORGANIZATION_NAME", "ACTION"}, func() {
		for _, action := range result.Actions {
			fmt.Printf("%-30s %-30s %-70s\n",
				truncate(action.OrgID, 30),
				truncate(action.OrgName, 30),
				truncate(action.Action, 70))
		}
	})

	return nil
}

// updateCancelBehavior updates subscription cancel behavior from pause to cancel
func updateCancelBehavior(ctx context.Context, c *cli.Command) error {
	recon, err := createReconciler(c)
	if err != nil {
		return err
	}

	orgIDs := c.Root().StringSlice("org-ids")

	result, err := recon.UpdateSubscriptionsCancelBehavior(ctx, orgIDs)
	if err != nil {
		return err
	}

	if len(result.Actions) == 0 {
		fmt.Println("No subscriptions require cancel behavior updates")
		return nil
	}

	fmt.Printf("Found %d subscriptions requiring cancel behavior updates\n", len(result.Actions))

	printOrgsTable([]string{"SUBSCRIPTION_ID", "ORGANIZATION_NAME", "ACTION"}, func() {
		for _, action := range result.Actions {
			fmt.Printf("%-30s %-30s %-70s\n",
				truncate(action.OrgID, 30),
				truncate(action.OrgName, 30),
				truncate(action.Action, 70))
		}
	})

	return nil
}

// createSchedules creates subscription schedules for subscriptions that don't have them
func createSchedules(ctx context.Context, c *cli.Command) error {
	recon, err := createReconciler(c)
	if err != nil {
		return err
	}

	orgIDs := c.Root().StringSlice("org-ids")

	result, err := recon.CreateMissingSubscriptionSchedules(ctx, orgIDs)
	if err != nil {
		return err
	}

	if len(result.Actions) == 0 {
		fmt.Println("No subscriptions require schedule creation")
		return nil
	}

	fmt.Printf("Found %d subscriptions requiring schedule creation\n", len(result.Actions))

	printOrgsTable([]string{"SUBSCRIPTION_ID", "ORGANIZATION_NAME", "ACTION"}, func() {
		for _, action := range result.Actions {
			fmt.Printf("%-30s %-30s %-70s\n",
				truncate(action.OrgID, 30),
				truncate(action.OrgName, 30),
				truncate(action.Action, 70))
		}
	})

	return nil
}

// reportMissingProducts generates a report of subscriptions with products not found in the catalog
func reportMissingProducts(ctx context.Context, c *cli.Command) error {
	recon, err := createReconciler(c)
	if err != nil {
		return err
	}

	catalogPath := c.String("catalog")
	report, err := recon.ReportSubscriptionsWithMissingProducts(ctx, catalogPath)
	if err != nil {
		return err
	}

	if len(report) == 0 {
		fmt.Println("All active subscriptions have products that exist in the catalog")
		return nil
	}

	fmt.Printf("Found %d subscriptions with missing products\n", len(report))

	outputFormat := c.String("output")
	switch outputFormat {
	case "json":
		// Output as JSON
		for _, item := range report {
			fmt.Printf(`{"subscription_id":"%s","customer_id":"%s","product_id":"%s","product_name":"%s","status":"%s","organization_id":"%s"}`+"\n",
				item.SubscriptionID, item.CustomerID, item.ProductID, item.ProductName, item.Status, item.OrganizationID)
		}
	default:
		// Output as table
		printProductTable([]string{"SUBSCRIPTION_ID", "CUSTOMER_ID", "PRODUCT_ID", "PRODUCT_NAME", "STATUS", "ORGANIZATION_ID"}, func() {
			for _, item := range report {
				fmt.Printf("%-20s %-20s %-20s %-30s %-10s %-30s\n",
					truncate(item.SubscriptionID, 20),
					truncate(item.CustomerID, 20),
					truncate(item.ProductID, 20),
					truncate(item.ProductName, 30),
					truncate(item.Status, 10),
					truncate(item.OrganizationID, 30))
			}
		})
	}

	return nil
}

// updatePersonalOrgMetadata updates Stripe customer metadata for personal organizations
func updatePersonalOrgMetadata(ctx context.Context, c *cli.Command) error {
	recon, err := createReconcilerWithDB(c)
	if err != nil {
		return err
	}

	result, err := recon.UpdatePersonalOrgMetadata(ctx)
	if err != nil {
		return err
	}

	if len(result.Actions) == 0 {
		fmt.Println("No personal organizations require metadata updates")
		return nil
	}

	fmt.Printf("Found %d personal organizations requiring metadata updates\n", len(result.Actions))

	printOrgsTable([]string{"ORGANIZATION_ID", "ORGANIZATION_NAME", "ACTION"}, func() {
		for _, action := range result.Actions {
			fmt.Printf("%-30s %-30s %-70s\n",
				truncate(action.OrgID, 30),
				truncate(action.OrgName, 30),
				truncate(action.Action, 70))
		}
	})

	return nil
}

// analyzeStripeMismatches performs comprehensive analysis of Stripe customers vs internal system data
func analyzeStripeMismatches(ctx context.Context, c *cli.Command) error {
	recon, err := createReconcilerWithDB(c)
	if err != nil {
		return err
	}

	action := c.String("action")
	report, err := recon.AnalyzeStripeSystemMismatches(ctx, action)
	if err != nil {
		return err
	}

	// If action was performed, report is empty and we're done
	if action != "" {
		return nil
	}

	fmt.Printf("Found %d mismatches\n", len(report))

	if len(report) == 0 {
		fmt.Println("No mismatches found between Stripe and internal system")
		return nil
	}

	outputFormat := c.String("output")
	switch outputFormat {
	case "json":
		// Output as JSON
		for _, item := range report {
			jsonData := fmt.Sprintf(`{"customer_id":"%s","organization_id":"%s", "organization_name":"%s", "mismatch_type":"%s","description":"%s","stripe_data":"%s","internal_data":"%s"`,
				item.CustomerID, item.OrganizationID, item.OrganizationName, item.MismatchType, item.Description, item.StripeData, item.InternalData)
			if len(item.SubscriptionIssues) > 0 {
				jsonData += `,"subscription_issues_count":` + fmt.Sprintf("%d", len(item.SubscriptionIssues))
			}
			jsonData += "}"
			fmt.Println(jsonData)
		}
	default:
		// Output as table
		printTable([]string{"CUSTOMER_ID", "ORGANIZATION_ID", "ORGANIZATION_NAME", "MISMATCH_TYPE", "DESCRIPTION", "SUBSCRIPTION_ISSUES"}, func() {
			for _, item := range report {
				subscriptionIssueCount := len(item.SubscriptionIssues)
				fmt.Printf("%-20s %-30s %-25s %-25s %-50s %-18s\n",
					truncate(item.CustomerID, 20),
					truncate(item.OrganizationID, 30),
					truncate(item.OrganizationName, 25),
					truncate(item.MismatchType, 25),
					truncate(item.Description, 50),
					fmt.Sprintf("%d", subscriptionIssueCount))
			}
		})
	}

	return nil
}

// printTable prints a simple table with headers and calls the provided function to print rows
func printTable(headers []string, printRows func()) {
	// Print header
	for i, header := range headers {
		width := getColumnWidth(i)
		fmt.Printf("%-*s ", width, header)
	}

	fmt.Println()

	// Print separator
	for i := range headers {
		width := getColumnWidth(i)
		for j := 0; j < width; j++ {
			fmt.Print("-")
		}
		fmt.Print(" ")
	}

	fmt.Println()

	// Print rows
	printRows()
}

// getColumnWidth returns the width for each column
func getColumnWidth(col int) int {
	widths := []int{20, 30, 25, 50, 18}
	if col < len(widths) {
		return widths[col]
	}

	return 20
}

// truncate truncates a string to the specified length
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}

	return s[:length-3] + "..."
}

// printProductTable prints a simple table for product reports
func printProductTable(headers []string, printRows func()) {
	// Print header
	widths := []int{20, 20, 20, 30, 10, 30}
	for i, header := range headers {
		width := widths[i]
		fmt.Printf("%-*s ", width, header)
	}

	fmt.Println()

	// Print separator
	for _, width := range widths {
		for j := 0; j < width; j++ {
			fmt.Print("-")
		}
		fmt.Print(" ")
	}

	fmt.Println()

	// Print rows
	printRows()
}

// printOrgsTable prints a simple table for organization reports
func printOrgsTable(headers []string, printRows func()) {
	// Print header
	widths := []int{30, 30, 70}
	for i, header := range headers {
		width := widths[i]
		fmt.Printf("%-*s ", width, header)
	}

	fmt.Println()

	// Print separator
	for _, width := range widths {
		for j := 0; j < width; j++ {
			fmt.Print("-")
		}
		fmt.Print(" ")
	}

	fmt.Println()

	// Print rows
	printRows()
}
