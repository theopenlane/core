package server

import (
	"context"
	"fmt"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/rs/zerolog/log"
)

// NewOpenAPISpec creates a new OpenAPI 3.1.0 specification based on the configured go interfaces and the operation types appended within the individual handlers
func NewOpenAPISpec() (*openapi3.T, error) {
	schemas := make(openapi3.Schemas)
	responses := make(openapi3.ResponseBodies)
	parameters := make(openapi3.ParametersMap)
	requestbodies := make(openapi3.RequestBodies)
	securityschemes := make(openapi3.SecuritySchemes)
	examples := make(openapi3.Examples)

	internalServerError := openapi3.NewResponse().
		WithDescription("Internal Server Error")
	responses["InternalServerError"] = &openapi3.ResponseRef{Value: internalServerError}

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
			WithName("X-API-Key").
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

	spec := &openapi3.T{
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
				URL:         "https://api.theopenlane.io",
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
	}

	// Merge SCIM OpenAPI spec
	if err := mergeSCIMSpec(spec); err != nil {
		log.Warn().Err(err).Msg("failed to merge SCIM spec, continuing without it")
	}

	return spec, nil
}

var (
	// ErrFailedToGetFilePath is returned when runtime.Caller fails to get the current file path
	ErrFailedToGetFilePath = fmt.Errorf("failed to get current file path")
	// ErrSCIMSpecNotFound is returned when the SCIM spec file is not found
	ErrSCIMSpecNotFound = fmt.Errorf("SCIM spec file not found")
)

// mergeSCIMSpec loads the SCIM OpenAPI specification and merges it into the main spec
func mergeSCIMSpec(mainSpec *openapi3.T) error {
	// Find the specs directory relative to this file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ErrFailedToGetFilePath
	}

	// Navigate from server -> httpserve to get to specs directory
	httpserveDir := filepath.Dir(filepath.Dir(filename))
	scimSpecPath := filepath.Join(httpserveDir, "specs", "scim.yaml")

	// Check if SCIM spec file exists
	if _, err := os.Stat(scimSpecPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrSCIMSpecNotFound, scimSpecPath)
	}

	// Load SCIM spec
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	scimSpec, err := loader.LoadFromFile(scimSpecPath)
	if err != nil {
		return fmt.Errorf("failed to load SCIM spec: %w", err)
	}

	// Validate SCIM spec
	if err := scimSpec.Validate(context.Background()); err != nil {
		return fmt.Errorf("SCIM spec validation failed: %w", err)
	}

	log.Info().Str("path", scimSpecPath).Msg("loaded SCIM OpenAPI spec")

	// Merge all spec components
	mergePaths(mainSpec, scimSpec)
	mergeComponents(mainSpec, scimSpec)
	mergeTags(mainSpec, scimSpec)

	log.Info().
		Int("paths", len(scimSpec.Paths.Map())).
		Int("schemas", len(scimSpec.Components.Schemas)).
		Msg("successfully merged SCIM spec into main OpenAPI spec")

	return nil
}

// mergePaths merges paths from source spec into main spec
func mergePaths(mainSpec, sourceSpec *openapi3.T) {
	if sourceSpec.Paths == nil {
		return
	}

	if mainSpec.Paths == nil {
		mainSpec.Paths = openapi3.NewPaths()
	}

	for path, pathItem := range sourceSpec.Paths.Map() {
		if pathItem != nil {
			mainSpec.Paths.Set(path, pathItem)
		}
	}
}

// mergeComponents merges all component types from source spec into main spec
func mergeComponents(mainSpec, sourceSpec *openapi3.T) {
	if sourceSpec.Components == nil {
		return
	}

	if mainSpec.Components == nil {
		mainSpec.Components = &openapi3.Components{}
	}

	mergeSchemas(mainSpec, sourceSpec)
	mergeResponses(mainSpec, sourceSpec)
	mergeParameters(mainSpec, sourceSpec)
	mergeRequestBodies(mainSpec, sourceSpec)
	mergeSecuritySchemes(mainSpec, sourceSpec)
}

// mergeSchemas merges schema definitions
func mergeSchemas(mainSpec, sourceSpec *openapi3.T) {
	if sourceSpec.Components.Schemas == nil {
		return
	}

	if mainSpec.Components.Schemas == nil {
		mainSpec.Components.Schemas = make(openapi3.Schemas)
	}

	for name, schema := range sourceSpec.Components.Schemas {
		mainSpec.Components.Schemas[name] = schema
	}
}

// mergeResponses merges response definitions
func mergeResponses(mainSpec, sourceSpec *openapi3.T) {
	if sourceSpec.Components.Responses == nil {
		return
	}

	if mainSpec.Components.Responses == nil {
		mainSpec.Components.Responses = make(openapi3.ResponseBodies)
	}

	for name, response := range sourceSpec.Components.Responses {
		mainSpec.Components.Responses[name] = response
	}
}

// mergeParameters merges parameter definitions
func mergeParameters(mainSpec, sourceSpec *openapi3.T) {
	if sourceSpec.Components.Parameters == nil {
		return
	}

	if mainSpec.Components.Parameters == nil {
		mainSpec.Components.Parameters = make(openapi3.ParametersMap)
	}

	for name, param := range sourceSpec.Components.Parameters {
		mainSpec.Components.Parameters[name] = param
	}
}

// mergeRequestBodies merges request body definitions
func mergeRequestBodies(mainSpec, sourceSpec *openapi3.T) {
	if sourceSpec.Components.RequestBodies == nil {
		return
	}

	if mainSpec.Components.RequestBodies == nil {
		mainSpec.Components.RequestBodies = make(openapi3.RequestBodies)
	}

	for name, reqBody := range sourceSpec.Components.RequestBodies {
		mainSpec.Components.RequestBodies[name] = reqBody
	}
}

// mergeSecuritySchemes merges security scheme definitions without overriding existing ones
func mergeSecuritySchemes(mainSpec, sourceSpec *openapi3.T) {
	if sourceSpec.Components.SecuritySchemes == nil {
		return
	}

	if mainSpec.Components.SecuritySchemes == nil {
		mainSpec.Components.SecuritySchemes = make(openapi3.SecuritySchemes)
	}

	for name, scheme := range sourceSpec.Components.SecuritySchemes {
		// Only add if not already present
		if _, exists := mainSpec.Components.SecuritySchemes[name]; !exists {
			mainSpec.Components.SecuritySchemes[name] = scheme
		}
	}
}

// mergeTags merges tag definitions from source spec into main spec
func mergeTags(mainSpec, sourceSpec *openapi3.T) {
	if len(sourceSpec.Tags) == 0 {
		return
	}

	// Check if any source tags already exist in main spec
	for _, sourceTag := range sourceSpec.Tags {
		tagExists := false

		for _, mainTag := range mainSpec.Tags {
			if mainTag.Name == sourceTag.Name {
				tagExists = true
				break
			}
		}

		if !tagExists {
			mainSpec.Tags = append(mainSpec.Tags, sourceTag)
		}
	}
}

// customizer is a customizer function that allows for the modification of the generated schemas
// this is used to ignore fields that are not required in the OAS specification
// and to add additional metadata to the schema such as descriptions and examples
var customizer = openapi3gen.SchemaCustomizer(func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if tag.Get("exclude") != "" && tag.Get("exclude") == "true" {
		return &openapi3gen.ExcludeSchemaSentinel{}
	}

	if tag.Get("description") != "" {
		schema.Description = tag.Get("description")
	}

	if tag.Get("example") != "" {
		schema.Example = tag.Get("example")
	}

	// For top-level structs, try to extract description from Go doc comments
	if (name == "_root" || tag == "") && schema.Description == "" && t.Kind() == reflect.Struct {
		if desc := getTypeDescription(t); desc != "" {
			schema.Description = desc
		}
	}

	return nil
})

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
		Type:        "oauth2",
		Description: "OAuth 2.0 authorization code flow for secure API access",
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
		Description:      "OpenID Connect authentication for secure API access",
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

var (
	docCache = make(map[string]string)
	docMutex sync.RWMutex
)

// getTypeDescription extracts the Go doc comment for a given type
func getTypeDescription(t reflect.Type) string {
	if t.PkgPath() == "" {
		return ""
	}

	// Create cache key
	cacheKey := t.PkgPath() + "." + t.Name()

	// Check cache first
	docMutex.RLock()

	if desc, exists := docCache[cacheKey]; exists {
		docMutex.RUnlock()
		return desc
	}

	docMutex.RUnlock()

	// Get the source file path for this type
	pkgPath := t.PkgPath()

	// Find the Go source files for this package
	fset := token.NewFileSet()

	// Try to parse the package directory
	pkgs, err := parser.ParseDir(fset, getPackageDir(pkgPath), nil, parser.ParseComments)
	if err != nil {
		return ""
	}

	// Look through all packages
	for _, pkg := range pkgs {
		// Create doc package
		docPkg := doc.New(pkg, "./", 0)

		// Look for the type in the package
		for _, typ := range docPkg.Types {
			if typ.Name == t.Name() {
				// Cache and return the description
				description := strings.TrimSpace(typ.Doc)

				docMutex.Lock()

				docCache[cacheKey] = description

				docMutex.Unlock()

				return description
			}
		}
	}

	return ""
}

// getPackageDir attempts to find the directory for a given package path
func getPackageDir(pkgPath string) string {
	// For local packages, try to find the source directory
	if strings.Contains(pkgPath, "github.com/theopenlane/core") {
		// Get the current file's directory
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			return ""
		}

		// Navigate to the project root and then to the package
		projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename))) // Go up from server -> httpserve -> internal -> core

		// Convert package path to file path
		relativePath := strings.TrimPrefix(pkgPath, "github.com/theopenlane/core/")

		return filepath.Join(projectRoot, relativePath)
	}

	return ""
}
