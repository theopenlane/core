//go:build clistripe

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	if err := webhookApp().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func webhookApp() *cli.Command {
	return &cli.Command{
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
			&cli.StringFlag{
				Name:  "config-root",
				Usage: "path to the core repository root (used to update config defaults)",
				Value: ".",
			},
		},
		Commands: []*cli.Command{
			listCommand(),
			statusCommand(),
			migrateCommand(),
		},
	}
}
