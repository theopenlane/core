package graphapi

import (
	"context"

	"github.com/theopenlane/core/common/jobspec"

	"github.com/theopenlane/core/v2/internal/graphapi/common"
)

func enqueueTrustCenterPreviewDomain(ctx context.Context, trustCenterID string, r *mutationResolver) error {
	_, err := r.db.Job.Insert(ctx, jobspec.CreatePreviewDomainArgs{
		TrustCenterID:            trustCenterID,
		TrustCenterPreviewZoneID: r.trustCenterPreviewZoneID,
		TrustCenterCnameTarget:   r.trustCenterCnameTarget,
	}, nil)
	if err != nil {
		return parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "trustcenterpreviewdomain"})
	}

	return nil
}
