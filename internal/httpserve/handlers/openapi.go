package handlers

import (
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/theopenlane/httpsling"
)

// badRequest is a wrapper for openaAPI bad request response
func badRequest() *openapi3.Response {
	return openapi3.NewResponse().
		WithDescription("Bad Request").
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/responses/BadRequest"}))
}

// internalServerError is a wrapper for openaAPI internal server error response
func internalServerError() *openapi3.Response {
	return openapi3.NewResponse().
		WithDescription("Internal Server Error").
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/responses/InternalServerError"}))
}

// notFound is a wrapper for openaAPI not found response
func notFound() *openapi3.Response {
	return openapi3.NewResponse().
		WithDescription("Not Found").
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/responses/NotFound"}))
}

// created is a wrapper for openaAPI created response
func created() *openapi3.Response {
	return openapi3.NewResponse().
		WithDescription("Created").
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/responses/Created"}))
}

// conflict is a wrapper for openaAPI conflict response
func conflict() *openapi3.Response {
	return openapi3.NewResponse().
		WithDescription("Conflict").
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/responses/Conflict"}))
}

// unauthorized is a wrapper for openaAPI unauthorized response
func unauthorized() *openapi3.Response {
	return openapi3.NewResponse().
		WithDescription("Unauthorized").
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/responses/Unauthorized"}))
}

// invalidInput is a wrapper for openaAPI invalidInput response
func invalidInput() *openapi3.Response {
	return openapi3.NewResponse().
		WithDescription("Invalid Input").
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/responses/InvalidInput"}))
}

// tooManyRequests is a wrapper for openaAPI too many requests response
func tooManyRequests() *openapi3.Response {
	return openapi3.NewResponse().
		WithDescription("Too Many Requests").
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/responses/TooManyRequests"}))
}

// AddRequestBody is used to add a request body definition to the OpenAPI schema
func (h *Handler) AddRequestBody(name string, body interface{}, op *openapi3.Operation) {
	request := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + name}))
	op.RequestBody = &openapi3.RequestBodyRef{Value: request}

	request.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
	request.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(body)}
}

// AddQueryParameter is used to add a query parameter definition to the OpenAPI schema (e.g ?name=value)
func (h *Handler) AddQueryParameter(name string, paramName string, op *openapi3.Operation) {
	schemaRef := &openapi3.SchemaRef{Ref: "#/components/schemas/" + name}
	param := openapi3.NewQueryParameter(paramName)

	param.Schema = schemaRef
	op.AddParameter(param)
}

// AddPathParameter is used to add a path parameter definition to the OpenAPI schema (e.g. /users/{id})
func (h *Handler) AddPathParameter(name string, paramName string, op *openapi3.Operation) {
	schemaRef := &openapi3.SchemaRef{Ref: "#/components/schemas/" + name}
	param := openapi3.NewPathParameter(paramName)

	param.Schema = schemaRef
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
