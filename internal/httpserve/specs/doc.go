// Package specs contains OpenAPI specifications that are merged into the main API spec.
//
// The scim.yaml file contains the SCIM 2.0 OpenAPI specification which is automatically
// loaded and merged into the main OpenAPI specification during server startup.
//
// The merge process is handled by internal/httpserve/server/openapi.go:mergeSCIMSpec().
package specs
