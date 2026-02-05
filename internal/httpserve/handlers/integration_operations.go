package handlers

import (
	"context"
	"maps"
	"strings"
	"time"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/runner"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	emitEnqueueTimeout                = 200 * time.Millisecond
	vulnerabilityCollectOperationName = "vulnerabilities.collect"
)

// RunIntegrationOperation queues a provider-published operation for async execution
func (h *Handler) RunIntegrationOperation(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleIntegrationOperationPayload, openapi.IntegrationOperationResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if h.IntegrationRegistry == nil {
		return h.InternalServerError(ctx, errIntegrationRegistryNotConfigured, openapiCtx)
	}
	if h.IntegrationOperations == nil {
		return h.InternalServerError(ctx, errIntegrationOperationsNotConfigured, openapiCtx)
	}
	if h.EventEmitter == nil {
		return h.InternalServerError(ctx, errIntegrationEmitterNotConfigured, openapiCtx)
	}
	if h.DBClient == nil || h.IntegrationStore == nil {
		return h.InternalServerError(ctx, errIntegrationStoreNotConfigured, openapiCtx)
	}

	requestCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(requestCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapiCtx)
	}

	providerType := types.ProviderTypeFromString(req.Provider)
	if providerType == types.ProviderUnknown {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	operationName := types.OperationName(strings.TrimSpace(req.Body.Operation))
	if operationName == "" {
		return h.BadRequest(ctx, rout.MissingField("operation"), openapiCtx)
	}

	operationConfig := cloneOperationConfig(req.Body.Config)
	integrationRecord, err := h.IntegrationStore.EnsureIntegration(requestCtx, user.OrganizationID, providerType)
	if err != nil {
		logx.FromContext(requestCtx).Warn().Err(err).Str("provider", string(providerType)).Msg("failed to load integration config")
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	if !operationDescriptorRegistered(h.IntegrationRegistry, providerType, operationName) {
		return h.BadRequest(ctx, keystore.ErrOperationNotRegistered, openapiCtx)
	}

	merged, err := helpers.ResolveOperationConfig(&integrationRecord.Config, string(operationName), req.Body.Config)
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}
	operationConfig = merged
	if operationName == vulnerabilityCollectOperationName {
		operationConfig = ensureIncludePayloads(operationConfig)
	}

	systemCtx := privacy.DecisionContext(requestCtx, privacy.Allow)
	runBuilder := h.DBClient.IntegrationRun.Create().
		SetOwnerID(user.OrganizationID).
		SetIntegrationID(integrationRecord.ID).
		SetOperationName(string(operationName)).
		SetRunType(enums.IntegrationRunTypeManual).
		SetStatus(enums.IntegrationRunStatusPending)
	if operationConfig != nil {
		runBuilder.SetOperationConfig(operationConfig)
	}

	runRecord, err := runBuilder.Save(systemCtx)
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	payload := soiree.IntegrationOperationRequestedPayload{
		RunID:     runRecord.ID,
		OrgID:     user.OrganizationID,
		Provider:  string(providerType),
		Operation: string(operationName),
		Force:     req.Body.Force,
	}
	eventClient := &runner.OperationEventContext{
		DB:         h.DBClient,
		Operations: h.IntegrationOperations,
	}
	eventID, emitErr := emitIntegrationOperationEvent(context.WithoutCancel(requestCtx), h.EventEmitter, payload, eventClient)
	if emitErr != nil {
		_ = h.DBClient.IntegrationRun.UpdateOneID(runRecord.ID).
			SetStatus(enums.IntegrationRunStatusFailed).
			SetError(emitErr.Error()).
			Exec(systemCtx)
		return h.InternalServerError(ctx, wrapIntegrationError("enqueue operation", emitErr), openapiCtx)
	}
	if eventID != "" {
		_ = h.DBClient.IntegrationRun.UpdateOneID(runRecord.ID).
			SetEventID(eventID).
			Exec(systemCtx)
	}

	out := openapi.IntegrationOperationResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  string(providerType),
		Operation: string(operationName),
		Status:    "queued",
		Summary:   "Integration operation queued",
		Details: map[string]any{
			"run_id":   runRecord.ID,
			"event_id": eventID,
			"status":   enums.IntegrationRunStatusPending.String(),
		},
	}

	return h.Success(ctx, out)
}

func cloneOperationConfig(input map[string]any) map[string]any {
	return maps.Clone(input)
}

// ensureIncludePayloads forces include_payloads for operation execution
func ensureIncludePayloads(config map[string]any) map[string]any {
	if config == nil {
		config = map[string]any{}
	}
	config["include_payloads"] = true
	return config
}

// operationDescriptorRegistered checks if the operation is registered for the provider
func operationDescriptorRegistered(reg ProviderRegistry, provider types.ProviderType, name types.OperationName) bool {
	if reg == nil {
		return false
	}
	for _, descriptor := range reg.OperationDescriptors(provider) {
		if descriptor.Name == name {
			return true
		}
	}
	return false
}

// emitIntegrationOperationEvent emits an async integration operation event
func emitIntegrationOperationEvent(ctx context.Context, emitter soiree.Emitter, payload soiree.IntegrationOperationRequestedPayload, client any) (string, error) {
	if emitter == nil {
		return "", errIntegrationEmitterNotConfigured
	}

	event, err := soiree.IntegrationOperationRequestedTopic.Wrap(payload)
	if err != nil {
		return "", err
	}

	event.SetContext(ctx)
	if client != nil {
		event.SetClient(client)
	}

	eventID := soiree.EventID(event)
	if eventID == "" {
		props := event.Properties()
		if props == nil {
			props = soiree.NewProperties()
		}
		eventID = ulids.New().String()
		props[soiree.PropertyEventID] = eventID
		event.SetProperties(props)
	}

	errCh := emitter.Emit(soiree.IntegrationOperationRequestedTopic.Name(), event)
	enqueueErr := drainEmitErrors(errCh, emitEnqueueTimeout)
	if enqueueErr != nil {
		return eventID, enqueueErr
	}

	return eventID, nil
}

// drainEmitErrors reads errors until close or timeout and returns the first error
func drainEmitErrors(errCh <-chan error, timeout time.Duration) error {
	if errCh == nil {
		return nil
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	var firstErr error
	for {
		select {
		case err, ok := <-errCh:
			if !ok {
				return firstErr
			}
			if err != nil && firstErr == nil {
				firstErr = err
			}
		case <-timer.C:
			return firstErr
		}
	}
}
