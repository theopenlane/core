package hooks

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/internal/integrations/operations"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
)

// sendSystemSlack marshals the input and executes a system Slack operation via
// the integration runtime on the ent client
func sendSystemSlack(ctx context.Context, client *generated.Client, operationName string, input any) error {
	rt := intruntime.FromClient(ctx, client)
	if rt == nil {
		return nil
	}

	config, err := json.Marshal(input)
	if err != nil {
		return err
	}

	_, err = rt.Dispatch(ctx, operations.DispatchRequest{
		DefinitionID: slack.DefinitionID.ID(),
		Operation:    operationName,
		Config:       config,
		RunType:      enums.IntegrationRunTypeEvent,
		Runtime:      true,
	})

	return err
}
