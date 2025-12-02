package graphapi

import (
	"context"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/shared/logx"
	"github.com/theopenlane/utils/rout"
)

// getTrustCenterID retrieves the trust center ID for the organization associated with the context
// if the provided trustCenterID is nil. If trustCenterID is not nil, it returns nil.
// this allows requests to skip providing the trust center ID in a request where it is needed
// later for authorization checks.
func getTrustCenterID(ctx context.Context, trustCenterID *string, object string) (*string, error) {
	if trustCenterID == nil {
		// check if the organization has a trust center and set the id
		trustCenterID, err := withTransactionalMutation(ctx).TrustCenter.Query().OnlyID(ctx)
		if err != nil {
			if generated.IsNotFound(err) {
				return nil, rout.NewMissingRequiredFieldError("trustCenterID")
			}

			return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: object})
		}

		if trustCenterID == "" {
			return nil, rout.NewMissingRequiredFieldError("trustCenterID")
		}

		logx.FromContext(ctx).Debug().Str("trust center id", trustCenterID).Msg("trustCenterID not provided, using organization's trust center")

		return &trustCenterID, nil
	}

	return nil, nil
}
