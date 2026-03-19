package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/samber/do/v2"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// ExecuteOperation runs one integration operation inline without run tracking.
func (r *Runtime) ExecuteOperation(ctx context.Context, installation *ent.Integration, operation types.OperationRegistration, credential types.CredentialSet, config json.RawMessage) (json.RawMessage, error) {
	if installation == nil {
		return nil, ErrInstallationRequired
	}

	return r.executeResolvedOperation(ctx, installation, operation, credential, config, false, operations.IngestOptions{
		Source: integrationgenerated.IntegrationIngestSourceOperation,
	})
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
				"response": jsonx.DecodeAnyOrNil(response),
			}
		}
		if completeErr := operations.CompleteRun(ctx, db, envelope.RunID, startedAt, result); completeErr != nil {
			return errors.Join(execErr, completeErr)
		}

		return execErr
	}

	if err := operations.MarkRunRunning(ctx, db, envelope.RunID); err != nil {
		return err
	}

	ingestOptions := operations.IngestOptions{
		Source: integrationgenerated.IntegrationIngestSourceOperation,
		RunID:  envelope.RunID,
	}
	if envelope.WorkflowMeta != nil {
		ingestOptions.Source = integrationgenerated.IntegrationIngestSourceWorkflow
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

	response, err := r.executeResolvedOperation(ctx, installation, operation, credential, envelope.Config, envelope.ForceClientRebuild, ingestOptions)
	if err != nil {
		return failRun(err, response)
	}

	return operations.CompleteRun(ctx, db, envelope.RunID, startedAt, operations.RunResult{
		Status:  enums.IntegrationRunStatusSuccess,
		Summary: "operation completed",
		Metrics: map[string]any{
			"response": jsonx.DecodeAnyOrNil(response),
		},
	})
}

// executeResolvedOperation executes the given operation with the input integration and registered Operation
func (r *Runtime) executeResolvedOperation(ctx context.Context, installation *ent.Integration, operation types.OperationRegistration, credential types.CredentialSet, config json.RawMessage, clientForce bool, ingestOptions operations.IngestOptions) (json.RawMessage, error) {
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

	req := types.OperationRequest{
		Integration: installation,
		Credential:  credential,
		Client:      client,
		Config:      jsonx.CloneRawMessage(config),
	}

	if operation.IngestHandle != nil {
		payloadSets, err := operation.IngestHandle(ctx, req)
		if err != nil {
			return nil, err
		}

		if err := operations.EmitPayloadSets(ctx, operations.IngestContext{
			Registry:     r.Registry(),
			DB:           do.MustInvoke[*ent.Client](r.injector),
			Runtime:      do.MustInvoke[*gala.Gala](r.injector),
			Installation: installation,
		}, operation.Name, operation.Ingest, payloadSets, ingestOptions); err != nil {
			return nil, err
		}

		return nil, nil
	}

	response, err := operation.Handle(ctx, req)
	if err != nil {
		return response, err
	}

	return response, nil
}
