package handlers

import (
	"fmt"
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/shortlinks"
	"github.com/theopenlane/logx"
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
func (h *Handler) EmailTestSendHandler(ctx echo.Context) error {
	if !h.IsDev {
		return h.BadRequest(ctx, ErrDevModeRequired)
	}

	if h.IntegrationsRuntime == nil {
		return h.InternalServerError(ctx, ErrIntegrationsNotConfigured)
	}

	req, err := BindAndValidate[EmailTestSendRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	requestCtx := ctx.Request().Context()

	if req.To == "" {
		return h.BadRequest(ctx, ErrRecipientRequired)
	}

	client, ok := h.IntegrationsRuntime.Registry().RuntimeClient(email.DefinitionID.ID())
	if !ok {
		return h.InternalServerError(ctx, ErrEmailClientNotAvailable)
	}

	emailClient, ok := client.(*email.Client)
	if !ok {
		return h.InternalServerError(ctx, ErrEmailClientNotAvailable)
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
			return h.BadRequest(ctx, ErrDispatcherNotFound)
		}
	}

	testTag := newman.Tag{Name: email.TagIsTest, Value: email.TagTrueValue}
	results := make([]EmailTestSendResult, 0, len(ops))

	for _, op := range ops {
		dispatcher, ok := email.DispatcherByKey(op.Name)
		if !ok {
			results = append(results, EmailTestSendResult{Name: op.Name, Status: "SKIP", Error: "dispatcher not found"})
			continue
		}

		fixtures := email.TestFixture(op.Name, req.To)
		if len(fixtures) == 0 {
			results = append(results, EmailTestSendResult{Name: op.Name, Status: "SKIP", Error: "no fixture defined"})
			continue
		}

		for i, payload := range fixtures {
			resultName := op.Name
			if i > 0 {
				resultName = fmt.Sprintf("%s #%d", op.Name, i+1)
			}

			logx.FromContext(requestCtx).Info().Str("dispatcher", resultName).Str("to", req.To).Msg("sending test email")

			// test sends must not mint real tracking links; PreHooks still run for org lookups
			sendErr := dispatcher.SendByKey(shortlinks.ContextWithoutShortlinks(requestCtx), types.OperationRequest{DB: h.IntegrationsRuntime.DB()}, emailClient, payload, newman.WithTag(testTag))
			if sendErr != nil {
				logx.FromContext(requestCtx).Error().Err(sendErr).Str("dispatcher", resultName).Msg("test email failed")
				results = append(results, EmailTestSendResult{Name: resultName, Status: "FAIL", Error: sendErr.Error()})

				continue
			}

			results = append(results, EmailTestSendResult{Name: resultName, Status: "OK"})
		}
	}

	return ctx.JSON(http.StatusOK, EmailTestSendResponse{Results: results})
}
