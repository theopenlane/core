package handlers

import (
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// EmailTestSendRequest is the request body for the test email send endpoint
type EmailTestSendRequest struct {
	// To is the recipient email address
	To string `json:"to" jsonschema:"required,description=Recipient email address"`
	// Name filters to a single dispatcher; empty sends all registered dispatchers
	Name string `json:"name,omitempty" jsonschema:"description=Dispatcher name to send (empty sends all)"`
}

// EmailTestSendResult is the per-dispatcher outcome in the response
type EmailTestSendResult struct {
	// Name is the dispatcher catalog key
	Name string `json:"name"`
	// Status is OK, SKIP, or FAIL
	Status string `json:"status"`
	// Error is the failure reason when status is FAIL or SKIP
	Error string `json:"error,omitempty"`
}

// EmailTestSendResponse is the response from the test email send endpoint
type EmailTestSendResponse struct {
	// Results contains the outcome for each dispatcher attempted
	Results []EmailTestSendResult `json:"results"`
}

// ExampleEmailTestSendRequest is used for OpenAPI schema registration
var ExampleEmailTestSendRequest = EmailTestSendRequest{
	To: "test@example.com",
}

// EmailTestSendHandler renders and sends test emails through registered dispatchers
// using the runtime's already-initialized email client. Only available when IsDev is true
func (h *Handler) EmailTestSendHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	if !h.IsDev {
		return h.BadRequest(ctx, ErrDevModeRequired, openapi)
	}

	if h.IntegrationsRuntime == nil {
		return h.InternalServerError(ctx, ErrIntegrationsNotConfigured, openapi)
	}

	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, ExampleEmailTestSendRequest, EmailTestSendResponse{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	requestCtx := ctx.Request().Context()

	if req.To == "" {
		return h.BadRequest(ctx, ErrRecipientRequired, openapi)
	}

	client, ok := h.IntegrationsRuntime.Registry().RuntimeClient(email.DefinitionID.ID())
	if !ok {
		return h.InternalServerError(ctx, ErrEmailClientNotAvailable, openapi)
	}

	emailClient, ok := client.(*email.Client)
	if !ok {
		return h.InternalServerError(ctx, ErrEmailClientNotAvailable, openapi)
	}

	ops := email.AllEmailOperations()
	if req.Name != "" {
		found := false

		for _, op := range ops {
			if op.Name == req.Name {
				ops = []types.OperationRegistration{op}
				found = true

				break
			}
		}

		if !found {
			return h.BadRequest(ctx, ErrDispatcherNotFound, openapi)
		}
	}

	testTag := newman.Tag{Name: email.TagIsTest, Value: "true"}
	results := make([]EmailTestSendResult, 0, len(ops))

	for _, op := range ops {
		dispatcher, ok := email.DispatcherByKey(op.Name)
		if !ok {
			results = append(results, EmailTestSendResult{Name: op.Name, Status: "SKIP", Error: "dispatcher not found"})
			continue
		}

		payload := email.TestFixture(op.Name, req.To)
		if payload == nil {
			results = append(results, EmailTestSendResult{Name: op.Name, Status: "SKIP", Error: "no fixture defined"})
			continue
		}

		logx.FromContext(requestCtx).Info().Str("dispatcher", op.Name).Str("to", req.To).Msg("sending test email")

		sendErr := dispatcher.SendByKey(requestCtx, types.OperationRequest{DB: h.IntegrationsRuntime.DB()}, emailClient, payload, newman.WithTag(testTag))
		if sendErr != nil {
			logx.FromContext(requestCtx).Error().Err(sendErr).Str("dispatcher", op.Name).Msg("test email failed")
			results = append(results, EmailTestSendResult{Name: op.Name, Status: "FAIL", Error: sendErr.Error()})

			continue
		}

		results = append(results, EmailTestSendResult{Name: op.Name, Status: "OK"})
	}

	return ctx.JSON(http.StatusOK, EmailTestSendResponse{Results: results})
}
