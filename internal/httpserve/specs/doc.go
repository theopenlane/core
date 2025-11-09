// Package specs contains OpenAPI specifications that are merged into the main API spec
// The scim.yaml file contains the SCIM 2.0 OpenAPI specification which is automatically
// loaded and merged into the main OpenAPI specification during server startup
// The openlane.openapi.json file contains the fully composed OpenAPI specification that
// is served from /api-docs and kept in source control for linting and diffing
package specs
