package hooks

import (
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/gala"
)

func init() { registerListeners(TemplateMappingListeners) }

type TemplateMappingRequest struct {
	ControlIDs []string `json:"control_ids"`
	ProgramID  string   `json:"program_id,omitempty"`
}

var TemplateMappingTopic = gala.NamespacedTopic[TemplateMappingRequest](gala.System, "template.mappings.requested")

func TemplateMappingListeners() []gala.Registration {
	return []gala.Registration{
		gala.Definition[TemplateMappingRequest]{
			Topic: TemplateMappingTopic,
			Caller: func(restored *auth.Caller, _ TemplateMappingRequest) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
			Handle: handleTemplateMappings,
		},
	}
}

func handleTemplateMappings(ctx gala.HandlerContext, req TemplateMappingRequest) error {
	client := generated.FromContext(ctx.Context)
	if client == nil {
		return ErrClientResolveFailed
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx.Context)
	if err != nil {
		return err
	}

	var program *generated.Program

	if req.ProgramID != "" {

		program, err = client.Program.Get(ctx.Context, req.ProgramID)
		if generated.IsNotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return controls.CloneTemplateMappings(ctx.Context, client, req.ControlIDs, orgID, program)
}
