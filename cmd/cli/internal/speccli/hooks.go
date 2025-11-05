//go:build cli

package speccli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/pkg/openlaneclient"
)

// CreatePreHook allows overriding create command execution before default handling.
type CreatePreHook func(ctx context.Context, cmd *cobra.Command, client *openlaneclient.OpenlaneClient) (handled bool, out OperationOutput, err error)
