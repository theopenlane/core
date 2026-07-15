package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/utils/rout"

	models "github.com/theopenlane/core/common/openapi"

	"github.com/theopenlane/core/pkg/jsonx"
)

// ErrorReplySchemaName is the component schema name for the error body (rout.Reply) returned by all handler error helpers
const ErrorReplySchemaName = "ErrorReply"

// errorReplySchemaRef is the reference to the shared error response schema
const errorReplySchemaRef = "#/components/schemas/" + ErrorReplySchemaName

// Security Requirements for common authentication patterns
var (
	// AuthenticatedSecurity for endpoints requiring authentication, satisfied by
	// either a bearer JWT or an API key; the requirements are listed under the
	// "or" context, meaning only one of them needs to be met. If you wanted to
	// require more than one at once you would list them all under a single
	// SecurityRequirement rather than individual ones
	AuthenticatedSecurity = &openapi3.SecurityRequirements{
		openapi3.SecurityRequirement{
			"bearer": []string{},
		},
		openapi3.SecurityRequirement{
			"apiKey": []string{},
		},
	}
	// PublicSecurity for public endpoints with no authentication
	PublicSecurity = &openapi3.SecurityRequirements{}
)

// AuthEndpointDesc creates a description for authenticated endpoints
func AuthEndpointDesc(action, resource string) string {
	return fmt.Sprintf("%s %s. Requires authentication.", action, resource)
}

// NewErrorResponse creates a response for the given status code whose body references the shared ErrorReply schema
func NewErrorResponse(statusCode int) *openapi3.Response {
	statusText := http.StatusText(statusCode)

	response := openapi3.NewResponse().
		WithDescription(statusText).
		WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{Ref: errorReplySchemaRef}))

	response.Content.Get(httpsling.ContentTypeJSON).Examples = map[string]*openapi3.ExampleRef{
		"error": {Value: openapi3.NewExample(normalizeExampleValue(rout.Reply{Success: false, Error: strings.ToLower(statusText)}))},
	}

	return response
}

// RedirectResponse builds the 302 response; redirects carry no body, only a Location header
func RedirectResponse() *openapi3.Response {
	response := openapi3.NewResponse().WithDescription(http.StatusText(http.StatusFound))
	response.Headers = openapi3.Headers{
		"Location": &openapi3.HeaderRef{Value: &openapi3.Header{Parameter: openapi3.Parameter{
			Description: "Destination URL of the redirect",
			Schema:      openapi3.NewStringSchema().NewRef(),
		}}},
	}

	return response
}

// RegisterRequestBody attaches the request body schema and the model's registered examples for
// the given request model instance to the operation; used only during spec generation
func RegisterRequestBody(op *openapi3.Operation, registry SchemaRegistry, instance any) error {
	schemaRef, err := registry.GetOrRegister(instance)
	if err != nil {
		return err
	}

	request := openapi3.NewRequestBody().
		WithDescription("Request body").
		WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
	op.RequestBody = &openapi3.RequestBodyRef{Value: request}

	request.Content.Get(httpsling.ContentTypeJSON).Examples = exampleRefs(instance)

	return nil
}

// RegisterSuccessResponse attaches a success response schema and the model's registered examples
// for the given response model instance to the operation; used only during spec generation
func RegisterSuccessResponse(op *openapi3.Operation, registry SchemaRegistry, status int, instance any) error {
	schemaRef, err := registry.GetOrRegister(instance)
	if err != nil {
		return err
	}

	response := openapi3.NewResponse().
		WithDescription(http.StatusText(status)).
		WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
	op.AddResponse(status, response)

	response.Content.Get(httpsling.ContentTypeJSON).Examples = exampleRefs(instance)

	return nil
}

// exampleRefs renders the model's explicitly registered examples as named example refs: the
// curated set from ExampleSetProvider when implemented, else one success example from
// ExampleProvider, and none otherwise — a zero-value instance is not a meaningful example and its
// nil slices would not even validate against the schema
func exampleRefs(instance any) map[string]*openapi3.ExampleRef {
	var examples map[string]any

	switch provider := pointerTo(instance).(type) {
	case models.ExampleSetProvider:
		examples = provider.ExampleSet()
	case models.ExampleProvider:
		examples = map[string]any{"success": provider.ExampleResponse()}
	default:
		return nil
	}

	refs := make(map[string]*openapi3.ExampleRef, len(examples))
	for name, example := range examples {
		refs[name] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(example))}
	}

	return refs
}

// pointerTo returns an addressable pointer to a copy of the value so pointer-receiver interface
// implementations can be detected on instances stored as any
func pointerTo(value any) any {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Pointer {
		return value
	}

	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)

	return ptr.Interface()
}

// RegisterParameters adds query and path parameters derived from the request model's struct tags
// to the operation; pathParams holds the parameter names present in the route's path template so a
// model shared by routes with and without a path parameter only declares it where the template has
// it. Used only during spec generation
func RegisterParameters(op *openapi3.Operation, instance any, pathParams map[string]bool) {
	typ := reflect.TypeOf(instance)
	if typ == nil {
		return
	}

	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return
	}

	for field := range typ.Fields() {
		field := field
		description := field.Tag.Get("description")
		exampleTag := field.Tag.Get("example")

		if queryTag := field.Tag.Get("query"); queryTag != "" {
			param := openapi3.NewQueryParameter(queryTag).WithSchema(openapi3.NewStringSchema())
			param.Description = description

			if exampleTag != "" {
				param.Example = exampleTag
			}

			op.AddParameter(param)
		}

		if paramTag := field.Tag.Get("param"); paramTag != "" && pathParams[paramTag] {
			param := openapi3.NewPathParameter(paramTag).WithSchema(openapi3.NewStringSchema())
			param.Description = description

			if exampleTag != "" {
				param.Example = exampleTag
			}

			op.AddParameter(param)
		}
	}
}

// normalizeExampleValue converts strongly-typed example objects into a generic form
// that kin-openapi can validate (maps, slices, primitives). Structs are marshaled
// to JSON and unmarshaled back into map[string]any / []any representations
func normalizeExampleValue(value any) any {
	if value == nil {
		return nil
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}

		return normalizeExampleValue(rv.Elem().Interface())
	}

	switch value.(type) {
	case map[string]any, []any,
		string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		json.Number:
		return value
	}

	var generic any

	if err := jsonx.RoundTrip(value, &generic); err != nil {
		return value
	}

	return generic
}
