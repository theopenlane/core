package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// ExecuteOperation runs one integration operation inline without run tracking
func (r *Runtime) ExecuteOperation(ctx context.Context, installation *ent.Integration, operation types.OperationRegistration, credentialOverrides types.CredentialBindings, config json.RawMessage) (json.RawMessage, error) {
	if installation == nil {
		return nil, ErrInstallationRequired
	}

	if len(config) > 0 {
		if err := validatePayload(operation.ConfigSchema, config, ErrOperationConfigInvalid); err != nil {
			return nil, err
		}
	}

	return r.executeResolvedOperation(ctx, installation, operation, credentialOverrides, config, false, operations.IngestOptions{
		Source: integrationgenerated.IntegrationIngestSourceOperation,
	})
}

// HandleOperation executes one queued operation envelope through the runtime-managed dependencies
func (r *Runtime) HandleOperation(ctx context.Context, envelope operations.Envelope) error {
	startedAt := time.Now()
	db := r.DB()

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
		ingestOptions.WorkflowMeta = envelope.WorkflowMeta
	}

	installation, err := db.Integration.Get(ctx, envelope.InstallationID)
	if err != nil {
		return failRun(err, nil)
	}

	operation, err := r.Registry().Operation(installation.DefinitionID, envelope.Operation)
	if err != nil {
		return failRun(err, nil)
	}

	response, err := r.executeResolvedOperation(ctx, installation, operation, nil, envelope.Config, envelope.ForceClientRebuild, ingestOptions)
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
func (r *Runtime) executeResolvedOperation(ctx context.Context, installation *ent.Integration, operation types.OperationRegistration, credentialOverrides types.CredentialBindings, config json.RawMessage, clientForce bool, ingestOptions operations.IngestOptions) (json.RawMessage, error) {
	var (
		client       any
		err          error
		credentials  types.CredentialBindings
		registration types.ClientRegistration
	)

	if operation.ClientRef.Valid() {
		registration, err = r.Registry().Client(installation.DefinitionID, operation.ClientRef)
		if err != nil {
			return nil, err
		}

		credentials, err = r.LoadCredentials(ctx, installation, registration.CredentialRefs)
		if err != nil {
			return nil, err
		}

		credentials = mergeCredentials(credentials, credentialOverrides)

		client, err = r.Keystore().BuildClient(ctx, installation, registration, credentials, config, clientForce)
		if err != nil {
			return nil, err
		}
	} else {
		credentials = mergeCredentials(nil, credentialOverrides)
	}

	req := types.OperationRequest{
		Integration: installation,
		Credential:  singleCredential(credentials, registration.CredentialRefs),
		Credentials: credentials,
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
			DB:           r.DB(),
			Runtime:      r.Gala(),
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

// mergeCredentials combines two credential binding sets, with overrides replacing matching entries in current
func mergeCredentials(current, overrides types.CredentialBindings) types.CredentialBindings {
	if len(current) == 0 && len(overrides) == 0 {
		return nil
	}

	merged := make(types.CredentialBindings, 0, len(current)+len(overrides))
	merged = append(merged, current...)

	for _, override := range overrides {
		replaced := false
		for index := range merged {
			if merged[index].Ref == override.Ref {
				merged[index] = override
				replaced = true
				break
			}
		}

		if !replaced {
			merged = append(merged, override)
		}
	}

	return merged
}

// singleCredential returns the resolved credential when exactly one ref is declared, or an empty set otherwise
func singleCredential(credentials types.CredentialBindings, refs []types.CredentialSlotID) types.CredentialSet {
	if len(refs) != 1 {
		return types.CredentialSet{}
	}

	credential, _ := credentials.Resolve(refs[0])

	return credential
}
