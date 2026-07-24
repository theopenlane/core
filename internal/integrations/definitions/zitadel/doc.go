// Package zitadel provides an integration definition for Zitadel
// (https://zitadel.com), an open-source identity and access management
// platform designed for cloud-native and self-hosted deployments.
//
// # Overview
//
// This integration connects a Zitadel instance to OpenLane, enabling directory
// sync of users for identity posture and access governance workflows.
//
// # Authentication
//
// The integration supports two authentication modes, selected at connect time:
//
//   - Personal Access Token (PAT): a token generated from the Zitadel admin
//     console, passed as a static OAuth2 bearer token on every API request.
//   - OAuth2 Client Credentials: a Zitadel service user Client ID and Client
//     Secret, exchanged for an access token via the client-credentials grant.
//
// Both modes are handled through the Zitadel Go SDK's unified client.
//
// # Directory Sync
//
// The directory sync operation pulls the following resources from the connected
// Zitadel instance and normalizes them into OpenLane's internal directory schemas:
//
//   - DirectoryAccount: Zitadel users (both human and machine user types)
//
// Note: Group sync is not yet implemented for Zitadel as the GroupService
// is not yet available in the Zitadel Go SDK (github.com/zitadel/zitadel-go/v3).
// Group sync will be added in a future update when SDK support is available.
//
// # Configuration
//
// The following options are available when connecting a Zitadel instance:
//
//   - Domain: The Zitadel instance domain (e.g. my-instance.zitadel.cloud). Uses TLS by
//     default; prefix with http:// for a non-TLS self-hosted instance.
//   - Token: A Personal Access Token generated from the Zitadel admin console (PAT mode)
//   - ClientID / ClientSecret: A service user's client credentials (OAuth mode)
//   - PrimaryDirectory: Marks this installation as the authoritative source for identity holder enrichment
//   - FilterExpr: An optional CEL expression applied to records before ingestion
//
// # SDK
//
// This integration uses the official Zitadel Go SDK (github.com/zitadel/zitadel-go/v3),
// which provides gRPC-based access to the Zitadel User Service API.
package zitadel