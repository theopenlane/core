//go:build cli

package speccli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/pkg/openlaneclient"
)

// CreatePreHook allows overriding create command execution before default handling.
type CreatePreHook func(ctx context.Context, cmd *cobra.Command, client *openlaneclient.OpenlaneClient) (handled bool, out OperationOutput, err error)

// UpdatePreHook allows overriding update command execution before default handling.
type UpdatePreHook func(ctx context.Context, cmd *cobra.Command, client *openlaneclient.OpenlaneClient) (handled bool, out OperationOutput, err error)

// GetPreHook allows overriding get command execution before default handling.
type GetPreHook func(ctx context.Context, cmd *cobra.Command, client *openlaneclient.OpenlaneClient) (handled bool, out OperationOutput, err error)

// DeletePreHook allows overriding delete command execution before default handling.
type DeletePreHook func(ctx context.Context, cmd *cobra.Command, client *openlaneclient.OpenlaneClient) (handled bool, out OperationOutput, err error)

// PrimaryPreHook handles execution when the resource command is invoked without subcommands.
type PrimaryPreHook func(ctx context.Context, cmd *cobra.Command) error

// CreateHookFactory produces a CreatePreHook given the hydrated spec.
type CreateHookFactory func(spec *CreateSpec) CreatePreHook

// UpdateHookFactory produces an UpdatePreHook given the hydrated spec.
type UpdateHookFactory func(spec *UpdateSpec) UpdatePreHook

// GetHookFactory produces a GetPreHook given the hydrated spec.
type GetHookFactory func(spec *GetSpec) GetPreHook

// DeleteHookFactory produces a DeletePreHook given the hydrated spec.
type DeleteHookFactory func(spec *DeleteSpec) DeletePreHook

// PrimaryHookFactory produces a PrimaryPreHook given the hydrated spec.
type PrimaryHookFactory func(spec *PrimarySpec) PrimaryPreHook
