package reconcile

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/config"
	entdb "github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/entitlements/reconciler"
	"github.com/theopenlane/core/pkg/entitlements"
)

var command = &cobra.Command{
	Use:   "reconcile",
	Short: "reconcile billing data with Stripe",
	RunE: func(cmdC *cobra.Command, args []string) error {
		return run(cmdC.Context())
	},
}

var cfg = NewConfig()

func init() {
	cmd.RootCmd.AddCommand(command)

	command.Flags().StringVar(&cfg.ConfigPath, "config", cfg.ConfigPath, "config file path")
	command.Flags().BoolVar(&cfg.DryRun, "dry-run", cfg.DryRun, "print actions without making changes")
}

func run(ctx context.Context) error {
	c, err := config.Load(&cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	dbClient, err := entdb.New(ctx, c.DB, nil)
	if err != nil {
		return fmt.Errorf("ent client: %w", err)
	}
	defer dbClient.Close()

	stripeClient, err := entitlements.NewStripeClient(
		entitlements.WithAPIKey(c.Entitlements.PrivateStripeKey),
		entitlements.WithConfig(c.Entitlements),
	)
	if err != nil {
		return fmt.Errorf("stripe client: %w", err)
	}

	options := []reconciler.Option{
		reconciler.WithDB(dbClient),
		reconciler.WithStripeClient(stripeClient),
	}
	if cfg.DryRun {
		options = append(options, reconciler.WithDryRun(nil))
	}

	recon, err := reconciler.New(options...)
	if err != nil {
		return err
	}

	return recon.Reconcile(ctx)
}
