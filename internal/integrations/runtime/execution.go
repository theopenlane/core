package runtime

import (
	"context"
	"encoding/json"
	"time"

	"github.com/samber/do/v2"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/jsonx"
)

// ExecuteOperation runs one integration operation inline without run tracking.
func (r *Runtime) ExecuteOperation(ctx context.Context, installation *ent.Integration, operation types.OperationRegistration, credential types.CredentialSet, config json.RawMessage) (json.RawMessage, error) {
	if installation == nil {
		return nil, ErrInstallationRequired
	}

	return r.executeResolvedOperation(ctx, installation, operation, credential, config, false)
}

// HandleOperation executes one queued operation envelope through the runtime-managed dependencies.
func (r *Runtime) HandleOperation(ctx context.Context, envelope operations.Envelope) error {
	startedAt := time.Now()
	db := do.MustInvoke[*ent.Client](r.injector)
	failRun := func(execErr error, response json.RawMessage) error {
		result := operations.RunResult{
			Status: enums.IntegrationRunStatusFailed,
			Error:  execErr.Error(),
		}
		if len(response) > 0 {
			result.Metrics = map[string]any{
				"response": decodeResponse(response),
			}
		}
		_ = operations.CompleteRun(ctx, db, envelope.RunID, startedAt, result)
		return execErr
	}

	if err := operations.MarkRunRunning(ctx, db, envelope.RunID); err != nil {
		return err
	}

	installation, err := db.Integration.Get(ctx, envelope.InstallationID)
	if err != nil {
		return failRun(err, nil)
	}

	operation, err := r.Registry().Operation(installation.DefinitionID, envelope.Operation)
	if err != nil {
		return failRun(err, nil)
	}

	credential, found, err := r.LoadCredential(ctx, installation)
	if err != nil {
		return failRun(err, nil)
	}
	if !found {
		credential = types.CredentialSet{}
	}

	response, err := r.executeResolvedOperation(ctx, installation, operation, credential, envelope.Config, envelope.ClientForce)
	if err != nil {
		return failRun(err, response)
	}

	return operations.CompleteRun(ctx, db, envelope.RunID, startedAt, operations.RunResult{
		Status:  enums.IntegrationRunStatusSuccess,
		Summary: "operation completed",
		Metrics: map[string]any{
			"response": decodeResponse(response),
		},
	})
}

func (r *Runtime) executeResolvedOperation(ctx context.Context, installation *ent.Integration, operation types.OperationRegistration, credential types.CredentialSet, config json.RawMessage, clientForce bool) (json.RawMessage, error) {
	var (
		client any
		err    error
	)

	if operation.ClientRef.Valid() {
		registration, err := r.Registry().Client(installation.DefinitionID, operation.ClientRef)
		if err != nil {
			return nil, err
		}

		client, err = do.MustInvoke[*keystore.Store](r.injector).BuildClient(ctx, installation, registration, credential, config, clientForce)
		if err != nil {
			return nil, err
		}
	}

	response, err := operation.Handle(ctx, types.OperationRequest{
		Integration: installation,
		Credential:  credential,
		Client:      client,
		Config:      jsonx.CloneRawMessage(config),
	})
	if err != nil {
		return response, err
	}

	if len(operation.Ingest) > 0 {
		if err := operations.ProcessIngest(ctx, r.Registry(), do.MustInvoke[*ent.Client](r.injector), installation, operation, response); err != nil {
			return nil, err
		}
	}

	return response, nil
}

func decodeResponse(value json.RawMessage) any {
	if len(value) == 0 {
		return map[string]any{}
	}

	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return string(value)
	}

	return decoded
}
