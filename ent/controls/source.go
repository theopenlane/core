package controls

import (
	"context"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/shared/logx"
)

// CheckSourceAllowed checks if the source is allowed to be modified
// if restrictedSource is nil, all sources are allowed
// if restrictedSource is set, only objects with a different source are allowed to be modified
func CheckSourceAllowed(ctx context.Context, restrictedSource *enums.ControlSource) bool {
	if restrictedSource == nil {
		return true
	}

	id := graphutils.GetStringInputVariableByName(ctx, "id")
	if id == nil {
		logx.FromContext(ctx).Error().Msg("no id found in context for externalReadOnly directive")
		return true
	}

	// now get the object from the database
	client := generated.FromContext(ctx)
	if client == nil {
		logx.FromContext(ctx).Error().Msg("no ent client found in context for externalReadOnly directive")
		return true
	}

	var objSource *enums.ControlSource
	obj, err := client.Control.Get(ctx, *id)
	if err == nil {
		objSource = &obj.Source
	} else {
		obj, err := client.Subcontrol.Get(ctx, *id)
		if err != nil {
			logx.FromContext(ctx).Error().Msg("failed to check for object source in externalReadOnly directive")

			return true
		}

		objSource = &obj.Source
	}

	// only allow if the source is different than the one on the object
	// the specified source is not allowed to make changes
	return *objSource != *restrictedSource
}
