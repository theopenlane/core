package hooks

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// sendSystemEmail marshals the input and executes a system email operation via
// the process-wide integration runtime
func sendSystemEmail(ctx context.Context, operationName string, input any) error {
	rt := intruntime.Default()
	if rt == nil {
		return nil
	}

	config, err := json.Marshal(input)
	if err != nil {
		return err
	}

	_, err = rt.Dispatch(ctx, types.DispatchRequest{
		DefinitionID: email.DefinitionID.ID(),
		Operation:    operationName,
		Config:       config,
		RunType:      enums.IntegrationRunTypeEvent,
		Runtime:      true,
	})

	return err
}
