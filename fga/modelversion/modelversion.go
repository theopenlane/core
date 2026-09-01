// Package modelversion identifies which authorization model in the fga store matches the model
// shipped with this build
//
// the model carries a generated marker type, model_version_<hash>, where the hash covers every
// model file listed in fga.mod (see task fga:generate:version). fgax asks the matcher whether a
// model it found in the store is the one this build expects, and creates a new model when nothing
// matches, so a model change reaches a store that already holds an older model
package modelversion

import (
	"context"
	"strings"

	openfga "github.com/openfga/go-sdk"
	"github.com/theopenlane/iam/fgax"

	fgamodel "github.com/theopenlane/core/v2/fga/model"
	"github.com/theopenlane/logx"
)

// markerPrefix is the prefix shared by every version marker type, used to pick the marker out of a
// deployed model for logging
const markerPrefix = "model_version"

// Matcher returns a matcher accepting only a model that carries this build's version marker
func Matcher(ctx context.Context) (fgax.ModelMatcher, error) {
	marker, err := fgamodel.VersionMarker()
	if err != nil {
		return nil, err
	}

	return func(model openfga.AuthorizationModel) bool {
		for _, td := range model.GetTypeDefinitions() {
			if td.GetType() == marker {
				logx.FromContext(ctx).Debug().Str("deployed_model_version", td.GetType()).Msg("marker version for deployed model matches, no new model needs to be created")

				return true
			}

			// log model version
			if strings.Contains(td.GetType(), markerPrefix) {
				logx.FromContext(ctx).Debug().Str("model_version_needed", marker).Str("deployed_model_version", td.GetType()).Msg("marker versions do not match, continuing")
			}
		}

		logx.FromContext(ctx).Info().Str("model_version", marker).Msg("deployed model version not found, creating new model")

		return false
	}, nil
}
