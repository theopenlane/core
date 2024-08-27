package server

import (
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

	generator := openapi3gen.NewGenerator(openapi3gen.UseAllExportedFields())
	for key, val := range openAPISchemas {
		ref, err := generator.NewSchemaRefForValue(val, schemas)
		if err != nil {
			return nil, err
		}

		schemas[key] = ref
	}

	errorResponse := &openapi3.SchemaRef{
		Ref: "#/components/schemas/ErrorResponse",
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

	securityschemes["bearer"] = &openapi3.SecuritySchemeRef{
		Value: openapi3.NewSecurityScheme().
			WithType("http").
			WithScheme("bearer").
			WithIn("header").
			WithDescription("Bearer authnetication, the token must be a valid Datum API token"),
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

	securityschemes["openid"] = &openapi3.SecuritySchemeRef{
		Value: (*OpenID)(&OpenID{
			ConnectURL: "https://api.theopenlane.io/.well-known/openid-configuration",
		}).Scheme(),
	}

	securityschemes["apikey"] = &openapi3.SecuritySchemeRef{
		Value: (*APIKey)(&APIKey{
			Name: "X-API-Key",
		}).Scheme(),
	}

	securityschemes["basic"] = &openapi3.SecuritySchemeRef{
		Value: (*Basic)(&Basic{
			Username: "username",
			Password: "password",
		}).Scheme(),
	}

	return &openapi3.T{
		OpenAPI: "3.1.0",
		Info: &openapi3.Info{
			Title:   "Datum OpenAPI 3.1.0 Specifications",
			Version: "v1.0.0",
			Contact: &openapi3.Contact{
				Name:  "Datum",
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
				Description: "Datum API Server",
				URL:         "https://api.theopenlane.io/v1",
			},
			&openapi3.Server{
				Description: "Datum API Server (local)",
				URL:         "http://localhost:17608/v1",
			},
		},
		ExternalDocs: &openapi3.ExternalDocs{
			Description: "Documentation for Datum's API services",
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
		Tags: openapi3.Tags{
			&openapi3.Tag{
				Name:        "schema",
				Description: "Add or update schema definitions",
			},
			&openapi3.Tag{
				Name:        "graphql",
				Description: "GraphQL query endpoints",
			},
		},
	}, nil
}

// openAPISchemas is a mapping of types to auto generate schemas for - these specifically live under the OAS "schema" type so that we can simply make schemaRef's to them and not have to define them all individually in the OAS paths
var openAPISchemas = map[string]any{
	"LoginRequest":               &models.LoginRequest{},
	"LoginResponse":              &models.LoginReply{},
	"ForgotPasswordRequest":      &models.ForgotPasswordRequest{},
	"ForgotPasswordResponse":     &models.ForgotPasswordReply{},
	"ResetPasswordRequest":       &models.ResetPasswordRequest{},
	"ResetPasswordResponse":      &models.ResetPasswordReply{},
	"RefreshRequest":             &models.RefreshRequest{},
	"RefreshResponse":            &models.RefreshReply{},
	"RegisterRequest":            &models.RegisterRequest{},
	"RegisterResponse":           &models.RegisterReply{},
	"ResendEmailRequest":         &models.ResendRequest{},
	"ResendEmailResponse":        &models.ResendReply{},
	"VerifyRequest":              &models.VerifyRequest{},
	"VerifyResponse":             &models.VerifyReply{},
	"PublishRequest":             &models.PublishRequest{},
	"PublishResponse":            &models.PublishReply{},
	"SwitchRequest":              &models.SwitchOrganizationRequest{},
	"SwitchResponse":             &models.SwitchOrganizationReply{},
	"VerifySubscriptionRequest":  &models.VerifySubscribeRequest{},
	"VerifySubscriptionResponse": &models.VerifySubscribeReply{},
	"InviteRequest":              &models.InviteRequest{},
	"InviteResponse":             &models.InviteReply{},
	"ErrorResponse":              &rout.StatusError{},
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
