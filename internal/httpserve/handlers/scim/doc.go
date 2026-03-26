// Package scim provides SCIM 2.0 (RFC 7644) compliant handlers for user and group provisioning.
//
// # Authentication and Authorization
//
// SCIM endpoints use Bearer token authentication with API tokens. The API token
// middleware handles authentication and sets the organization context based on the
// token's owner_id. SCIM operations follow the same authorization rules as other
// API endpoints.
//
// SCIM installations are always handled as directory-scoped ingest: requests are ingested
// into DirectoryAccount, DirectoryGroup, and DirectoryMembership records scoped
// to the resolved integration installation.
//
// # Handler Implementation
//
// The package uses the elimity-com/scim library which provides:
//   - RFC-compliant schema definitions for User and Group resources
//   - Request parsing and validation
//   - Patch operation handling
//   - List/filter/pagination support
//
// The DirectoryUserHandler and DirectoryGroupHandler implement the scim.ResourceHandler
// interface and translate SCIM resources into integration-scoped directory records.
//
// # Context Flow
//
// Request context flows through the following middleware chain:
//  1. Base middleware (transaction, logging, etc.)
//  2. Authentication middleware (validates bearer token, creates authenticated user context)
//  3. SCIM handlers (operate within the token's organization + integration scope)
package scim
