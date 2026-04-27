package awssecurityhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/configservice"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// IngestHandle adapts IAM directory sync to the ingest operation registration boundary
func (d CheckSync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequestConfig(configServiceClient, checkSyncOperation, ErrOperationConfigInvalid, func(ctx context.Context, _ types.OperationRequest, client *configservice.Client, cfg CheckSync) ([]types.IngestPayloadSet, error) {
		if cfg.Disable {
			logx.FromContext(ctx).Debug().Msg("aws_iam: check sync is disabled")

			return nil, nil
		}

		return d.Run(ctx, client, cfg)
	})
}

// Run collects AWS IAM users, and optionally groups and memberships
func (CheckSync) Run(_ context.Context, _ *configservice.Client, _ CheckSync) ([]types.IngestPayloadSet, error) {

	return nil, nil
}
