package handlers_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func (suite *HandlerTestSuite) TestBindStartOAuthFlowHandler() {
	t := suite.T()

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Test the OpenAPI binding function
	operation := suite.h.BindStartOAuthFlowHandler()

	// Verify basic operation properties
	assert.NotNil(t, operation)
	assert.Equal(t, "Start OAuth integration flow for third-party providers", operation.Description)
	assert.Equal(t, "StartOAuthFlow", operation.OperationID)
	assert.Contains(t, operation.Tags, "oauth")
	assert.Contains(t, operation.Tags, "integrations")

	// Verify security requirements
	assert.NotNil(t, operation.Security)
	assert.Greater(t, len(*operation.Security), 0)

	// Verify request body
	assert.NotNil(t, operation.RequestBody)

	// Verify responses
	assert.NotNil(t, operation.Responses)
	assert.Contains(t, operation.Responses, "200")
	assert.Contains(t, operation.Responses, "400")
	assert.Contains(t, operation.Responses, "401")
	assert.Contains(t, operation.Responses, "500")
}

func (suite *HandlerTestSuite) TestBindHandleOAuthCallbackHandler() {
	t := suite.T()

	ctx := echocontext.NewTestEchoContext().Request().Context()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// Test the OpenAPI binding function
	operation := suite.h.BindHandleOAuthCallbackHandler()

	// Verify basic operation properties
	assert.NotNil(t, operation)
	assert.Equal(t, "Handle OAuth callback and store integration tokens", operation.Description)
	assert.Equal(t, "HandleOAuthCallback", operation.OperationID)
	assert.Contains(t, operation.Tags, "oauth")
	assert.Contains(t, operation.Tags, "integrations")

	// Verify security requirements
	assert.NotNil(t, operation.Security)
	assert.Greater(t, len(*operation.Security), 0)

	// Verify request body
	assert.NotNil(t, operation.RequestBody)

	// Verify responses
	assert.NotNil(t, operation.Responses)
	assert.Contains(t, operation.Responses, "200")
	assert.Contains(t, operation.Responses, "400")
	assert.Contains(t, operation.Responses, "401")
	assert.Contains(t, operation.Responses, "500")
}

func (suite *HandlerTestSuite) TestOAuthBindingConsistency() {
	t := suite.T()

	// Test that both binding functions return valid operations
	startOp := suite.h.BindStartOAuthFlowHandler()
	callbackOp := suite.h.BindHandleOAuthCallbackHandler()

	require.NotNil(t, startOp)
	require.NotNil(t, callbackOp)

	// Both should have the same tags
	assert.Equal(t, startOp.Tags, callbackOp.Tags)

	// Both should have security requirements
	assert.NotNil(t, startOp.Security)
	assert.NotNil(t, callbackOp.Security)

	// Both should have request bodies
	assert.NotNil(t, startOp.RequestBody)
	assert.NotNil(t, callbackOp.RequestBody)

	// Both should have standard HTTP responses
	expectedResponses := []string{"200", "400", "401", "500"}
	for _, code := range expectedResponses {
		assert.Contains(t, startOp.Responses, code, "StartOAuthFlow should have response code %s", code)
		assert.Contains(t, callbackOp.Responses, code, "HandleOAuthCallback should have response code %s", code)
	}
}
