package graphapi

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/utils/rout"
)

// getTrustCenterID retrieves the trust center ID for the organization associated with the context
// if the provided trustCenterID is nil. If trustCenterID is not nil, it returns nil.
// this allows requests to skip providing the trust center ID in a request where it is needed
// later for authorization checks.
func getTrustCenterID(ctx context.Context, trustCenterID *string, object string) (*string, error) {
	// check if the organization has a trust center and set the id
	orgTrustCenterID, err := withTransactionalMutation(ctx).TrustCenter.Query().OnlyID(ctx)
	if err != nil {
		if trustCenterID == nil {
			if generated.IsNotFound(err) {
				return nil, rout.NewMissingRequiredFieldError("trustCenterID")
			}

			return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: object})
		}

		logx.FromContext(ctx).Error().Str("provided_trust_center_id", *trustCenterID).Str("org_trust_center_id", orgTrustCenterID).Msg("mismatch between provided id and organization trust center id")
		return nil, parseRequestError(ctx, privacy.Deny, common.Action{Action: common.ActionCreate, Object: object})
	}

	if trustCenterID == nil && orgTrustCenterID == "" {
		return nil, rout.NewMissingRequiredFieldError("trustCenterID")
	}

	if trustCenterID != nil && orgTrustCenterID != *trustCenterID {
		logx.FromContext(ctx).Error().Str("provided_trust_center_id", *trustCenterID).Str("org_trust_center_id", orgTrustCenterID).Msg("mismatch between provided id and organization trust center id")
		return nil, privacy.Deny
	}

	logx.FromContext(ctx).Debug().Str("trust center id", orgTrustCenterID).Msg("trustCenterID not provided, using organization's trust center")

	return &orgTrustCenterID, nil
}
