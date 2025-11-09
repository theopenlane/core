package handlers

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/theopenlane/httpsling"
)

// Security Requirements for common authentication patterns
var (
	// AuthenticatedSecurity for endpoints requiring authentication
	AuthenticatedSecurity = BearerSecurity()
	// PublicSecurity for public endpoints with no authentication
	PublicSecurity = &openapi3.SecurityRequirements{}
	// AllAuthSecurity for endpoints accepting any authentication method
	AllAuthSecurity = AllSecurityRequirements()
)

// Error Response Patterns for common error combinations
var (
	// StandardAuthErrors for typical authenticated endpoints
	StandardAuthErrors = []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError}
	// PublicEndpointErrors for public endpoints
	PublicEndpointErrors = []int{http.StatusBadRequest, http.StatusInternalServerError}
	// AdminOnlyErrors for admin-only endpoints
	AdminOnlyErrors = []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError}
)

// AuthEndpointDesc creates a description for authenticated endpoints
func AuthEndpointDesc(action, resource string) string {
	return fmt.Sprintf("%s %s. Requires authentication.", action, resource)
}

func PublicEndpointDesc(action, resource string) string {
	return fmt.Sprintf("%s %s. No authentication required.", action, resource)
}

func AdminEndpointDesc(action, resource string) string {
	return fmt.Sprintf("%s %s. Requires admin privileges.", action, resource)
}

// commonResponse creates a response that references a common error schema
func commonResponse(statusCode int) *openapi3.Response {
	statusText := http.StatusText(statusCode)
	// For now, just return a simple response without schema reference
	// TODO: Add proper error schema references when StatusError schema is working
	return openapi3.NewResponse().
		WithDescription(statusText)
}

// Legacy wrapper functions for backward compatibility
// These can be gradually replaced with commonResponse(statusCode) calls

// badRequest is a wrapper for OpenAPI bad request response
func badRequest() *openapi3.Response {
	return commonResponse(http.StatusBadRequest)
}

// internalServerError is a wrapper for OpenAPI internal server error response
func internalServerError() *openapi3.Response {
	return commonResponse(http.StatusInternalServerError)
}

// unauthorized is a wrapper for OpenAPI unauthorized response
func unauthorized() *openapi3.Response {
	return commonResponse(http.StatusUnauthorized)
}

// AddRequestBody is used to add a request body definition to the OpenAPI schema
func (h *Handler) AddRequestBody(name string, body interface{}, op *openapi3.Operation) {
	request := openapi3.NewRequestBody().
		WithDescription("Request body").
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + name}))
	op.RequestBody = &openapi3.RequestBodyRef{Value: request}

	request.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
	request.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(body))}
}

// AddQueryParameter is used to add a query parameter definition to the OpenAPI schema (e.g ?name=value)
func (h *Handler) AddQueryParameter(paramName string, op *openapi3.Operation) {
	param := openapi3.NewQueryParameter(paramName).WithSchema(openapi3.NewStringSchema())

	op.AddParameter(param)
}

// AddPathParameter is used to add a path parameter definition to the OpenAPI schema (e.g. /users/{id})
func (h *Handler) AddPathParameter(paramName string, op *openapi3.Operation) {
	param := openapi3.NewPathParameter(paramName).WithSchema(openapi3.NewStringSchema())

	op.AddParameter(param)
}

// AddResponse is used to add a response definition to the OpenAPI schema
func (h *Handler) AddResponse(name string, description string, body interface{}, op *openapi3.Operation, status int) {
	response := openapi3.NewResponse().
		WithDescription(description).
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + name}))
	op.AddResponse(status, response)

	response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
	response.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(body))}
}

// bearerSecurity is used to add a bearer security definition to the OpenAPI schema
func BearerSecurity() *openapi3.SecurityRequirements {
	return &openapi3.SecurityRequirements{
		openapi3.SecurityRequirement{
			"bearer": []string{},
		},
	}
}

// oauthSecurity is used to add a oauth security definition to the OpenAPI schema
func OauthSecurity() *openapi3.SecurityRequirements {
	return &openapi3.SecurityRequirements{
		openapi3.SecurityRequirement{
			"oauth2": []string{},
		},
	}
}

// basicSecurity is used to add a basic security definition to the OpenAPI schema
func BasicSecurity() *openapi3.SecurityRequirements {
	return &openapi3.SecurityRequirements{
		openapi3.SecurityRequirement{
			"basic": []string{},
		},
	}
}

// apiKeySecurity is used to add a apiKey security definition to the OpenAPI schema
func APIKeySecurity() *openapi3.SecurityRequirements {
	return &openapi3.SecurityRequirements{
		openapi3.SecurityRequirement{
			"apiKey": []string{},
		},
	}
}

// allSecurityRequirements is used to add all security definitions to the OpenAPI schema under the "or" context,
// meaning you can satisfy 1 or any / all of these requirements but only 1 is required
// if you wanted to list the security requirements with an "and" operator (meaning more than 1 needs to be met)
// you would list them all under a single `SecurityRequirement` rather than individual ones
func AllSecurityRequirements() *openapi3.SecurityRequirements {
	return &openapi3.SecurityRequirements{
		openapi3.SecurityRequirement{
			"bearer": []string{},
		},
		openapi3.SecurityRequirement{
			"oauth2": []string{},
		},
		openapi3.SecurityRequirement{
			"basic": []string{},
		},
		openapi3.SecurityRequirement{
			"apiKey": []string{},
		},
	}
}

// AddStandardResponses adds common error responses to an OpenAPI operation
// Note: This function is now a no-op since error responses are registered
// dynamically by the error handler methods themselves when they are called.
func AddStandardResponses(operation *openapi3.Operation) {
	if operation != nil {
		operation.AddResponse(http.StatusBadRequest, badRequest())
		operation.AddResponse(http.StatusUnauthorized, unauthorized())
		operation.AddResponse(http.StatusInternalServerError, internalServerError())
	}
}
