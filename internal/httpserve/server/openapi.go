package server

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"

	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/utils/rout"
)

// NewOpenAPISpec creates a new OpenAPI 3.1.0 specification based on the configured go interfaces and the operation types appended within the individual handlers
func NewOpenAPISpec() (*openapi3.T, error) {
	schemas := make(openapi3.Schemas)
	responses := make(openapi3.ResponseBodies)
	parameters := make(openapi3.ParametersMap)
	requestbodies := make(openapi3.RequestBodies)
	securityschemes := make(openapi3.SecuritySchemes)
	examples := make(openapi3.Examples)

	// Schemas are now dynamically registered by handlers using the schema registry

	errorResponse := &openapi3.SchemaRef{
		Ref: "#/components/schemas/ErrorReply",
	}

	_, err := openapi3gen.NewSchemaRefForValue(&rout.StatusError{}, schemas)
	if err != nil {
		return nil, err
	}

	internalServerError := openapi3.NewResponse().
		WithDescription("Internal Server Error").
		WithContent(openapi3.NewContentWithJSONSchemaRef(errorResponse))
	responses["InternalServerError"] = &openapi3.ResponseRef{Value: internalServerError}

	badRequest := openapi3.NewResponse().
		WithDescription("Bad Request").
		WithContent(openapi3.NewContentWithJSONSchemaRef(errorResponse))
	responses["BadRequest"] = &openapi3.ResponseRef{Value: badRequest}

	unauthorized := openapi3.NewResponse().
		WithDescription("Unauthorized").
		WithContent(openapi3.NewContentWithJSONSchemaRef(errorResponse))
	responses["Unauthorized"] = &openapi3.ResponseRef{Value: unauthorized}

	conflict := openapi3.NewResponse().
		WithDescription("Conflict").
		WithContent(openapi3.NewContentWithJSONSchemaRef(errorResponse))
	responses["Conflict"] = &openapi3.ResponseRef{Value: conflict}

	notFound := openapi3.NewResponse().
		WithDescription("Not Found").
		WithContent(openapi3.NewContentWithJSONSchemaRef(errorResponse))
	responses["NotFound"] = &openapi3.ResponseRef{Value: notFound}

	created := openapi3.NewResponse().
		WithDescription("Created").
		WithContent(openapi3.NewContentWithJSONSchemaRef(errorResponse))
	responses["Created"] = &openapi3.ResponseRef{Value: created}

	invalidInput := openapi3.NewResponse().
		WithDescription("Invalid Input").
		WithContent(openapi3.NewContentWithJSONSchemaRef(errorResponse))
	responses["InvalidInput"] = &openapi3.ResponseRef{Value: invalidInput}

	toomanyrequests := openapi3.NewResponse().
		WithDescription("Too Many Requests").
		WithContent(openapi3.NewContentWithJSONSchemaRef(errorResponse))
	responses["TooManyRequests"] = &openapi3.ResponseRef{Value: toomanyrequests}

	securityschemes["bearer"] = &openapi3.SecuritySchemeRef{
		Value: openapi3.NewSecurityScheme().
			WithType("http").
			WithScheme("bearer").
			WithBearerFormat("JWT").
			WithDescription("Bearer authentication, the token must be a valid API token"),
	}

	securityschemes["apiKey"] = &openapi3.SecuritySchemeRef{
		Value: openapi3.NewSecurityScheme().
			WithType("apiKey").
			WithIn("header").
			WithDescription("API Key authentication, the key must be a valid API key"),
	}

	securityschemes["basic"] = &openapi3.SecuritySchemeRef{
		Value: openapi3.NewSecurityScheme().
			WithType("http").
			WithScheme("basic").
			WithDescription("Username and Password based authentication"),
	}

	securityschemes["oauth2"] = &openapi3.SecuritySchemeRef{
		Value: (*OAuth2)(&OAuth2{
			AuthorizationURL: "https://api.theopenlane.io/oauth2/authorize",
			TokenURL:         "https://api.theopenlane.io/oauth2/token",
			RefreshURL:       "https://api.theopenlane.io/oauth2/refresh",
			Scopes: map[string]string{
				"read":  "Read access",
				"write": "Write access",
			},
		}).Scheme(),
	}

	securityschemes["openIdConnect"] = &openapi3.SecuritySchemeRef{
		Value: (*OpenID)(&OpenID{
			ConnectURL: "https://api.theopenlane.io/.well-known/openid-configuration",
		}).Scheme(),
	}

	return &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:       "Openlane OpenAPI 3.0.0 Specifications",
			Description: "Openlane's API services are designed to provide a simple and easy to use interface for interacting with the Openlane platform. This API is designed to be used by both internal and external clients to interact with the Openlane platform.",
			Version:     "v1.0.0",
			Contact: &openapi3.Contact{
				Name:  "Openlane",
				Email: "support@theopenlane.io",
				URL:   "https://theopenlane.io",
			},
			License: &openapi3.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0",
			},
		},
		Paths: openapi3.NewPaths(),
		Servers: openapi3.Servers{
			&openapi3.Server{
				Description: "Openlane API Server",
				URL:         "https://api.theopenlane.io/v1",
			},
		},
		ExternalDocs: &openapi3.ExternalDocs{
			Description: "Documentation for Openlane's API services",
			URL:         "https://docs.theopenlane.io",
		},

		Components: &openapi3.Components{
			Schemas:         schemas,
			Responses:       responses,
			Parameters:      parameters,
			RequestBodies:   requestbodies,
			SecuritySchemes: securityschemes,
			Examples:        examples,
		},
	}, nil
}

// customizer is a customizer function that allows for the modification of the generated schemas
// this is used to ignore fields that are not required in the OAS specification
// and to add additional metadata to the schema such as descriptions and examples
var customizer = openapi3gen.SchemaCustomizer(func(_ string, _ reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if tag.Get("exclude") != "" && tag.Get("exclude") == "true" {
		return &openapi3gen.ExcludeSchemaSentinel{}
	}

	if tag.Get("description") != "" {
		schema.Description = tag.Get("description")
	}

	if tag.Get("example") != "" {
		schema.Example = tag.Get("example")
	}

	return nil
})

// openAPISchemas is a mapping of types to auto generate schemas for - these specifically live under the OAS "schema" type so that we can simply make schemaRef's to them and not have to define them all individually in the OAS paths
var openAPISchemas = map[string]any{
	"LoginRequest":                      &models.LoginRequest{},
	"LoginReply":                        &models.LoginReply{},
	"ForgotPasswordRequest":             &models.ForgotPasswordRequest{},
	"ForgotPasswordReply":               &models.ForgotPasswordReply{},
	"ResetPasswordRequest":              &models.ResetPasswordRequest{},
	"ResetPasswordReply":                &models.ResetPasswordReply{},
	"RefreshRequest":                    &models.RefreshRequest{},
	"RefreshReply":                      &models.RefreshReply{},
	"RegisterRequest":                   &models.RegisterRequest{},
	"RegisterReply":                     &models.RegisterReply{},
	"ResendEmailRequest":                &models.ResendRequest{},
	"ResendEmailReply":                  &models.ResendReply{},
	"VerifyRequest":                     &models.VerifyRequest{},
	"VerifyReply":                       &models.VerifyReply{},
	"SwitchOrganizationRequest":         &models.SwitchOrganizationRequest{},
	"SwitchOrganizationReply":           &models.SwitchOrganizationReply{},
	"VerifySubscriptionRequest":         &models.VerifySubscribeRequest{},
	"VerifySubscriptionReply":           &models.VerifySubscribeReply{},
	"InviteRequest":                     &models.InviteRequest{},
	"InviteReply":                       &models.InviteReply{},
	"ErrorReply":                        &rout.StatusError{},
	"AccountAccessRequest":              &models.AccountAccessRequest{},
	"AccountAccessReply":                &models.AccountAccessReply{},
	"AccountRolesRequest":               &models.AccountRolesRequest{},
	"AccountRolesReply":                 &models.AccountRolesReply{},
	"AccountRolesOrganizationRequest":   &models.AccountRolesOrganizationRequest{},
	"AccountRolesOrganizationReply":     &models.AccountRolesOrganizationReply{},
	"AccountFeaturesReply":              &models.AccountFeaturesReply{},
	"JobRunnerRegistrationRequest":      &models.JobRunnerRegistrationRequest{},
	"JobRunnerRegistrationReply":        &models.JobRunnerRegistrationReply{},
	"SSOStatusReply":                    &models.SSOStatusReply{},
	"TFARequest":                        &models.TFARequest{},
	"TFAReply":                          &models.TFAReply{},
	"WebauthnRegistrationRequest":       &models.WebauthnRegistrationRequest{},
	"WebauthnBeginRegistrationResponse": &models.WebauthnBeginRegistrationResponse{},
}

// OAuth2 is a struct that represents an OAuth2 security scheme
type OAuth2 struct {
	AuthorizationURL string
	TokenURL         string
	RefreshURL       string
	Scopes           map[string]string
}

// Scheme returns the OAuth2 security scheme
func (i *OAuth2) Scheme() *openapi3.SecurityScheme {
	return &openapi3.SecurityScheme{
		Type: "oauth2",
		Flows: &openapi3.OAuthFlows{
			AuthorizationCode: &openapi3.OAuthFlow{
				AuthorizationURL: i.AuthorizationURL,
				TokenURL:         i.TokenURL,
				RefreshURL:       i.RefreshURL,
				Scopes:           i.Scopes,
			},
		},
	}
}

// OpenID is a struct that represents an OpenID Connect security scheme
type OpenID struct {
	ConnectURL string
}

// Scheme returns the OpenID Connect security scheme
func (i *OpenID) Scheme() *openapi3.SecurityScheme {
	return &openapi3.SecurityScheme{
		Type:             "openIdConnect",
		OpenIdConnectUrl: i.ConnectURL,
	}
}

// APIKey is a struct that represents an API Key security scheme
type APIKey struct {
	Name string
}

// Scheme returns the API Key security scheme
func (k *APIKey) Scheme() *openapi3.SecurityScheme {
	return &openapi3.SecurityScheme{
		Type: "http",
		In:   "header",
		Name: k.Name,
	}
}

// Basic is a struct that represents a Basic Auth security scheme
type Basic struct {
	Username string
	Password string
}

// Scheme returns the Basic Auth security scheme
func (b *Basic) Scheme() *openapi3.SecurityScheme {
	return &openapi3.SecurityScheme{
		Type:   "http",
		Scheme: "basic",
	}
}
