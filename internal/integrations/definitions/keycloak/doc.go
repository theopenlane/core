// Package keycloak provides an integration definition for Keycloak
// (https://www.keycloak.org), an open-source identity and access management
// solution designed for self-hosted and on-premise deployments.
//
// # Overview
//
// This integration connects a Keycloak instance to OpenLane using OAuth2
// client credentials authentication, enabling directory sync of users, groups,
// and group memberships for identity posture and access governance workflows.
//
// # Authentication
//
// The integration authenticates using OAuth2 client credentials (client ID and
// client secret) scoped to a specific realm. An access token is acquired at
// connection time and automatically refreshed before expiry during long-running
// sync operations.
//
// # Directory Sync
//
// The directory sync operation pulls the following resources from the connected
// Keycloak realm and normalizes them into OpenLane's internal directory schemas:
//
//   - DirectoryAccount: Keycloak users within the configured realm
//   - DirectoryGroup: Keycloak groups with full representation
//   - DirectoryMembership: Group membership relationships between accounts and groups
//
// Service accounts (identified by the presence of serviceAccountClientId) are
// mapped to the SERVICE account type. Group sync can be disabled via the
// DisableGroupSync option in UserInput.
//
// # Configuration
//
// The following options are available when connecting a Keycloak instance:
//
//   - BaseURL: The base URL of the Keycloak instance (e.g. https://keycloak.mycompany.com)
//   - Realm: The Keycloak realm to sync (e.g. master or your-organization-realm)
//   - ClientID: The client ID with permissions to read realm users and groups
//   - ClientSecret: The client secret for the configured client
//   - DisableGroupSync: When true, only users are synced; groups and memberships are skipped
//   - PrimaryDirectory: Marks this installation as the authoritative source for identity holder enrichment
//   - FilterExpr: An optional CEL expression applied to records before ingestion
//
// # SDK
//
// This integration uses the gocloak library (github.com/Nerzal/gocloak/v13),
// a well-maintained Go client for the Keycloak Admin REST API.
package keycloak
