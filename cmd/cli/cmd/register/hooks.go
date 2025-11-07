//go:build cli

package register

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
)

func registerUserHook(_ *speccli.PrimarySpec) speccli.PrimaryPreHook {
	return func(ctx context.Context, _ *cobra.Command) error {
		payload, err := executeRegister(ctx)
		if err != nil {
			return err
		}

		return speccli.PrintJSON(payload)
	}
}
