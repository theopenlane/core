package hooks

import (
	"context"
	"encoding/json"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
)

// sendSystemEmail marshals the input and executes a system email operation via
// the integration runtime on the ent client
func sendSystemEmail(ctx context.Context, client *generated.Client, operationName string, input any) error {
	rt := intruntime.FromClient(ctx, client)
	if rt == nil {
		return nil
	}

	config, err := json.Marshal(input)
	if err != nil {
		return err
	}

	_, err = rt.ExecuteRuntimeOperation(ctx, email.DefinitionID.ID(), operationName, config)

	return err
}
