package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// backfillCmd runs the one-time, idempotent data backfills and exits
var backfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "run one-time data backfills and exit",
	Run: func(cmd *cobra.Command, _ []string) {
		err := backfill(cmd.Context())
		cobra.CheckErr(err)
	},
}

// init registers the backfill command and its flags on the root command.
func init() {
	rootCmd.AddCommand(backfillCmd)

	backfillCmd.PersistentFlags().String("config", "./config/.config.yaml", "config file location")
}

// backfill runs service as backfill only
func backfill(ctx context.Context) error {
	return serve(ctx, true)
}
