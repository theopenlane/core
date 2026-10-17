package hooks

import (
	"fmt"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// init registers the template mapping listeners so gala setup picks them up automatically
func init() { registerListeners(TemplateMappingListeners) }

// TemplateMappingRequest asks for mappings from source template controls to be cloned,
// optionally filtered to a program's framework
type TemplateMappingRequest struct {
	ControlIDs []string `json:"control_ids"`
	ProgramID  string   `json:"program_id,omitempty"`
}

// TemplateMappingTopic carries explicit requests to clone recommended mappings after controls are cloned
var TemplateMappingTopic = gala.NamespacedTopic[TemplateMappingRequest](gala.System, "control.template_mappings.requested")

// TemplateMappingListeners clones recommended template mappings into the caller's organization
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

// handleTemplateMappings clones mappings named by an explicit request
func handleTemplateMappings(hctx gala.HandlerContext, req TemplateMappingRequest) error {

	client := generated.FromContext(hctx.Context)
	if client == nil {
		return fmt.Errorf("%w: unable to handle control template mapping request", ErrClientResolveFailed)
	}

	if len(req.ControlIDs) == 0 {
		return nil
	}

	var program *generated.Program

	if req.ProgramID != "" {
		var (
			ok  bool
			err error
		)

		program, ok, err = entityops.LoadEntity(hctx.Context, req.ProgramID, client.Program.Get)
		if err != nil || !ok {
			return err
		}
	}

	return controls.CloneTemplateMappings(hctx.Context, client, req.ControlIDs, hctx.Caller.OrganizationID, program)
}
