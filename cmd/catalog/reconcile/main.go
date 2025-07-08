package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/theopenlane/core/config"
	entdb "github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/entitlements/reconciler"
	"github.com/theopenlane/core/pkg/entitlements"
)

func main() {
	if err := app().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func app() *cli.Command {
	return &cli.Command{
		Name:  "reconcile",
		Usage: "reconcile billing data with Stripe",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Value:   "./config/.config.yaml",
				Usage:   "config file path",
				Sources: cli.EnvVars("CORE_CONFIG"),
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "print actions without making changes",
				Value: true,
			},
		},
		Action: run,
	}
}

func run(ctx context.Context, c *cli.Command) error {
	cfgLoc := c.String("config")
	cfg, err := config.Load(&cfgLoc)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	dbClient, err := entdb.New(ctx, cfg.DB, nil)
	if err != nil {
		return fmt.Errorf("ent client: %w", err)
	}

	defer dbClient.Close()

	stripeClient, err := entitlements.NewStripeClient(
		entitlements.WithAPIKey(cfg.Entitlements.PrivateStripeKey),
		entitlements.WithConfig(cfg.Entitlements),
	)

	if err != nil {
		return fmt.Errorf("stripe client: %w", err)
	}

	options := []reconciler.Option{
		reconciler.WithDB(dbClient),
		reconciler.WithStripeClient(stripeClient),
	}

	if c.Bool("dry-run") {
		options = append(options, reconciler.WithDryRun(nil))
	}

	recon, err := reconciler.New(options...)
	if err != nil {
		return err
	}

	return recon.Reconcile(ctx)
}
