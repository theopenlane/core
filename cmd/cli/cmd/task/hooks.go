//go:build cli

package task

import (
	"context"
	"encoding/json"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func taskCSVHook(ctx context.Context, cobraCmd *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
	if cmd.InputFile == "" {
		return false, speccli.OperationOutput{}, nil
	}

	upload, err := storage.NewUploadFile(cmd.InputFile)
	if err != nil {
		return true, speccli.OperationOutput{}, err
	}

	result, err := client.CreateBulkCSVTask(ctx, graphql.Upload{
		File:        upload.RawFile,
		Filename:    upload.OriginalName,
		Size:        upload.Size,
		ContentType: upload.ContentType,
	})
	if err != nil {
		return true, speccli.OperationOutput{}, err
	}

	bulk := result.GetCreateBulkCSVTask()
	raw, err := json.Marshal(bulk.Tasks)
	if err != nil {
		return true, speccli.OperationOutput{}, err
	}

	var records []map[string]any
	if err := json.Unmarshal(raw, &records); err != nil {
		return true, speccli.OperationOutput{}, err
	}

	return true, speccli.OperationOutput{Raw: result, Records: records}, nil
}
