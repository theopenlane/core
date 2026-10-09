package hooks

import (
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/gala"
)

func init() { registerListeners(TemplateMappingListeners) }

type TemplateMappingRequest struct {
	ControlIDs []string `json:"control_ids"`
	ProgramID  string   `json:"program_id,omitempty"`
}

func TemplateMappingListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaControl,
			Operations: []string{entityops.OpCreate, entityops.OpUpdate, entityops.OpUpdateOne},
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
			Handle: handleTemplateMappings,
		},
	}
}

func handleTemplateMappings(inv entityops.Invocation, _ entityops.MutationPayload) error {
	oc, ok := gala.OperationContextFromContext(inv.Context)
	if !ok || len(oc.Attributes) == 0 {
		return nil
	}

	req, err := gala.DecodeAttributes[TemplateMappingRequest](oc)
	if err != nil {
		return err
	}

	if len(req.ControlIDs) == 0 {
		return nil
	}

	var program *generated.Program

	if req.ProgramID != "" {

		program, err = inv.Client.Program.Get(inv.Context, req.ProgramID)
		if generated.IsNotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return controls.CloneTemplateMappings(inv.Context, inv.Client, req.ControlIDs, inv.Caller.OrganizationID, program)
}
