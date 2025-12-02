//go:build generate

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/theopenlane/ent/hush/crypto"
)

const minDisableArgs = 2

// main is the entry point for the hush CLI application
func main() {
	if err := hushApp().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// hushApp creates the CLI application for keyset management
func hushApp() *cli.Command {
	app := &cli.Command{
		Name:  "hush",
		Usage: "Manage Tink encryption keysets for envelope encryption",
		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "Generate a new keyset",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "quiet",
						Usage: "output only the keyset without additional text",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "export",
						Usage: "output in export format",
						Value: false,
					},
				},
				Action: func(_ context.Context, c *cli.Command) error {
					return generateKeyset(c.Bool("quiet"), c.Bool("export"))
				},
			},
			{
				Name:      "rotate",
				Usage:     "Rotate keys in an existing keyset",
				ArgsUsage: "<keyset>",
				Action: func(_ context.Context, c *cli.Command) error {
					if c.Args().Len() < 1 {
						return cli.Exit("Please provide the current keyset as an argument", 1)
					}
					return rotateKeyset(c.Args().Get(0))
				},
			},
			{
				Name:      "info",
				Usage:     "Show keyset information",
				ArgsUsage: "<keyset>",
				Action: func(_ context.Context, c *cli.Command) error {
					if c.Args().Len() < 1 {
						return cli.Exit("Please provide the keyset as an argument", 1)
					}
					return showKeysetInfo(c.Args().Get(0))
				},
			},
			{
				Name:      "add",
				Usage:     "Add a new key to the keyset (without making it primary)",
				ArgsUsage: "<keyset>",
				Action: func(_ context.Context, c *cli.Command) error {
					if c.Args().Len() < 1 {
						return cli.Exit("Please provide the current keyset as an argument", 1)
					}
					return addKeyToKeyset(c.Args().Get(0))
				},
			},
			{
				Name:      "disable",
				Usage:     "Disable old keys in the keyset",
				ArgsUsage: "<keyset> <keep-count>",
				Action: func(_ context.Context, c *cli.Command) error {
					if c.Args().Len() < minDisableArgs {
						return cli.Exit("Please provide the keyset and keep-count as arguments", 1)
					}

					keepCount := 0
					if _, err := fmt.Sscanf(c.Args().Get(1), "%d", &keepCount); err != nil {
						return cli.Exit(fmt.Sprintf("Invalid keep-count: %v", err), 1)
					}

					return disableOldKeys(c.Args().Get(0), keepCount)
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "keyset",
				Usage:   "Tink keyset from environment variable",
				Sources: cli.EnvVars("OPENLANE_TINK_KEYSET"),
			},
		},
	}

	return app
}

// generateKeyset generates a new Tink keyset and prints it
func generateKeyset(quiet, export bool) error {
	keyset, err := crypto.GenerateTinkKeyset()
	if err != nil {
		return fmt.Errorf("failed to generate keyset: %w", err)
	}

	if quiet {
		fmt.Print(keyset)
		return nil
	}

	if export {
		fmt.Printf("export OPENLANE_TINK_KEYSET=%s\n", keyset)
		return nil
	}

	// Default behavior (same as current)
	fmt.Println(keyset)
	return nil
}

// rotateKeyset rotates the keys in the provided keyset and prints the new keyset
func rotateKeyset(currentKeyset string) error {
	newKeyset, err := crypto.RotateKeyset(currentKeyset)
	if err != nil {
		return fmt.Errorf("failed to rotate keyset: %w", err)
	}

	fmt.Println(newKeyset)

	return nil
}

// showKeysetInfo retrieves and displays information about the provided keyset
func showKeysetInfo(keysetData string) error {
	info, err := crypto.GetKeysetInfo(keysetData)
	if err != nil {
		return fmt.Errorf("failed to get keyset info: %w", err)
	}

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format keyset info: %w", err)
	}

	fmt.Println(string(jsonData))

	return nil
}

// addKeyToKeyset adds a new key to the existing keyset without making it primary
func addKeyToKeyset(currentKeyset string) error {
	newKeyset, err := crypto.AddKeyToKeyset(currentKeyset)
	if err != nil {
		return fmt.Errorf("failed to add key to keyset: %w", err)
	}

	fmt.Println(newKeyset)

	return nil
}

// disableOldKeys disables old keys in the keyset, keeping the specified number of keys active
func disableOldKeys(currentKeyset string, keepCount int) error {
	newKeyset, err := crypto.DisableOldKeys(currentKeyset, keepCount)
	if err != nil {
		return fmt.Errorf("failed to disable old keys: %w", err)
	}

	fmt.Println(newKeyset)

	return nil
}
