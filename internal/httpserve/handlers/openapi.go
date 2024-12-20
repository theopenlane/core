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

// AddRequestBody is used to add a request body definition to the OpenAPI schema
func (h *Handler) AddRequestBody(name string, body interface{}, op *openapi3.Operation) {
	request := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + name}))
	op.RequestBody = &openapi3.RequestBodyRef{Value: request}

	request.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
	request.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(body)}
}

// AddQueryParameter is used to add a query parameter definition to the OpenAPI schema (e.g ?name=value)
func (h *Handler) AddQueryParameter(name string, paramName string, body interface{}, op *openapi3.Operation) {
	schemaRef := &openapi3.SchemaRef{Ref: "#/components/schemas/" + name}
	param := openapi3.NewQueryParameter(paramName)

	param.Schema = schemaRef
	op.AddParameter(param)
}

// AddPathParameter is used to add a path parameter definition to the OpenAPI schema (e.g. /users/{id})
func (h *Handler) AddPathParameter(name string, paramName string, body interface{}, op *openapi3.Operation) {
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
