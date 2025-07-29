package handlers

import (
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/theopenlane/httpsling"
)

// commonResponse creates a response that references a common component
func commonResponse(statusCode int) *openapi3.Response {
	statusText := http.StatusText(statusCode)
	componentName := strings.ReplaceAll(statusText, " ", "")
	return openapi3.NewResponse().
		WithDescription(statusText).
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/responses/" + componentName}))
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

// notFound is a wrapper for OpenAPI not found response
func notFound() *openapi3.Response {
	return commonResponse(http.StatusNotFound)
}

// created is a wrapper for OpenAPI created response
func created() *openapi3.Response {
	return commonResponse(http.StatusCreated)
}

// conflict is a wrapper for OpenAPI conflict response
func conflict() *openapi3.Response {
	return commonResponse(http.StatusConflict)
}

// unauthorized is a wrapper for OpenAPI unauthorized response
func unauthorized() *openapi3.Response {
	return commonResponse(http.StatusUnauthorized)
}

// forbidden is a wrapper for OpenAPI forbidden response
func forbidden() *openapi3.Response {
	return commonResponse(http.StatusForbidden)
}

// invalidInput is a wrapper for OpenAPI invalid input response
// Note: This uses the InvalidInput component, not BadRequest
func invalidInput() *openapi3.Response {
	return openapi3.NewResponse().
		WithDescription("Invalid Input").
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/responses/InvalidInput"}))
}

// tooManyRequests is a wrapper for OpenAPI too many requests response
func tooManyRequests() *openapi3.Response {
	return commonResponse(http.StatusTooManyRequests)
}

// AddRequestBody is used to add a request body definition to the OpenAPI schema
func (h *Handler) AddRequestBody(name string, body interface{}, op *openapi3.Operation) {
	request := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + name}))
	op.RequestBody = &openapi3.RequestBodyRef{Value: request}

	request.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
	request.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(body)}
}

// AddRequestBodyWithRegistry is used to add a request body definition to the OpenAPI schema with automatic type registration
func (h *Handler) AddRequestBodyWithRegistry(body interface{}, op *openapi3.Operation, registry interface {
	GetOrRegister(any) (*openapi3.SchemaRef, error)
}) error {
	schemaRef, err := registry.GetOrRegister(body)
	if err != nil {
		return err
	}

	request := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
	op.RequestBody = &openapi3.RequestBodyRef{Value: request}

	request.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
	request.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(body)}

	return nil
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
	response.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(body)}
}

// AddResponseWithRegistry is used to add a response definition to the OpenAPI schema with automatic type registration
func (h *Handler) AddResponseWithRegistry(description string, body interface{}, op *openapi3.Operation, status int, registry interface {
	GetOrRegister(any) (*openapi3.SchemaRef, error)
}) error {
	schemaRef, err := registry.GetOrRegister(body)
	if err != nil {
		return err
	}

	response := openapi3.NewResponse().
		WithDescription(description).
		WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
	op.AddResponse(status, response)

	response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
	response.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(body)}

	return nil
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
