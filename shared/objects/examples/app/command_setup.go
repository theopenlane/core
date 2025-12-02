//go:build examples

package app

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/theopenlane/shared/objects/examples/config"
	"github.com/theopenlane/shared/objects/examples/openlane"
)

func setupCommand() *cli.Command {
	return &cli.Command{
		Name:  "setup",
		Usage: "Initialize Openlane configuration for examples",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "api",
				Usage: "Openlane API base URL",
				Value: "http://localhost:17608",
			},
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Force re-initialization even if configuration exists",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg := setupConfig{
				API:   cmd.String("api"),
				Force: cmd.Bool("force"),
			}
			out := cmd.Writer
			if out == nil {
				out = os.Stdout
			}
			return runSetup(ctx, out, cfg)
		},
	}
}

type setupConfig struct {
	API   string
	Force bool
}

func runSetup(ctx context.Context, out io.Writer, cfg setupConfig) error {
	fmt.Fprintln(out, "=== Openlane Examples Setup ===")

	baseURL, err := url.Parse(cfg.API)
	if err != nil {
		return fmt.Errorf("invalid API URL: %w", err)
	}

	if cfg.Force && config.ConfigExists() {
		fmt.Fprintln(out, "Removing existing configuration...")
		if err := config.DeleteConfig(); err != nil {
			return fmt.Errorf("delete existing config: %w", err)
		}
	}

	if config.ConfigExists() && !cfg.Force {
		fmt.Fprintln(out, "Configuration already exists. Use --force to re-initialize.")
		loadedConfig, err := config.Load(nil)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		printConfig(out, &loadedConfig.Openlane)
		return nil
	}

	config, err := openlane.SetupWorkflow(ctx, out, baseURL)
	if err != nil {
		return fmt.Errorf("setup workflow: %w", err)
	}

	fmt.Fprintln(out, "\n=== Setup Summary ===")
	printConfig(out, config)

	return nil
}

func printConfig(out io.Writer, cfg *config.OpenlaneConfig) {
	fmt.Fprintf(out, "Email: %s\n", cfg.Email)
	fmt.Fprintf(out, "Organization ID: %s\n", cfg.OrganizationID)
	if len(cfg.PAT) > 16 {
		fmt.Fprintf(out, "PAT: %s...%s\n", cfg.PAT[:8], cfg.PAT[len(cfg.PAT)-8:])
	} else {
		fmt.Fprintln(out, "PAT: [configured]")
	}
	fmt.Fprintf(out, "Base URL: %s\n", cfg.BaseURL)
}
