// Package authentik provides an integration definition for Authentik
// (https://goauthentik.io), an open-source identity provider designed for
// self-hosted and on-premise deployments.
//
// # Overview
//
// This integration connects an Authentik instance to OpenLane using API token
// authentication, enabling directory sync of users, groups, and group
// memberships for identity posture and access governance workflows.
//
// # Authentication
//
// The integration authenticates using a static API token generated from the
// Authentik admin panel. The token is scoped to a service account and passed
// as a Bearer token on every API request. No OAuth2 flow or token refresh is
// required.
//
// # Directory Sync
//
// The directory sync operation pulls the following resources from the connected
// Authentik instance and normalizes them into OpenLane's internal directory
// schemas:
//
//   - DirectoryAccount: Authentik users (internal and external types)
//   - DirectoryGroup: Authentik groups with classification derived from is_superuser
//   - DirectoryMembership: Group membership relationships between accounts and groups
//
// Users of type service_account and internal_service_account are included in
// the sync but mapped to the SERVICE account type. Group sync can be disabled
// via the DisableGroupSync option in UserInput.
//
// # Configuration
//
// The following options are available when connecting an Authentik instance:
//
//   - BaseURL: The base URL of the Authentik instance (e.g. https://authentik.mycompany.com)
//   - Token: A static API token generated from the Authentik admin panel
//   - DisableGroupSync: When true, only users are synced; groups and memberships are skipped
//   - PrimaryDirectory: Marks this installation as the authoritative source for identity holder enrichment
//   - FilterExpr: An optional CEL expression applied to records before ingestion
//
// # SDK
//
// This integration uses the official Authentik Go SDK (goauthentik.io/api/v3),
// which is generated from Authentik's OpenAPI schema and versioned per Authentik
// release. This ensures breaking API changes are caught at compile time when the
// SDK version is updated.
package authentik
