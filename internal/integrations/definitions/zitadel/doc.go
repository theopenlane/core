// Package zitadel provides an integration definition for Zitadel
// (https://zitadel.com), an open-source identity and access management
// platform designed for cloud-native and self-hosted deployments.
//
// # Overview
//
// This integration connects a Zitadel instance to OpenLane using Personal
// Access Token authentication, enabling directory sync of users for identity
// posture and access governance workflows.
//
// # Authentication
//
// The integration authenticates using a Personal Access Token (PAT) generated
// from the Zitadel admin console. The token is passed as a static OAuth2
// bearer token on every API request via the Zitadel Go SDK.
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
//   - Domain: The Zitadel instance domain (e.g. https://my-instance.zitadel.cloud)
//   - Token: A Personal Access Token generated from the Zitadel admin console
//   - PrimaryDirectory: Marks this installation as the authoritative source for identity holder enrichment
//   - FilterExpr: An optional CEL expression applied to records before ingestion
//
// # SDK
//
// This integration uses the official Zitadel Go SDK (github.com/zitadel/zitadel-go/v3),
// which provides gRPC-based access to the Zitadel User Service API.
package zitadel