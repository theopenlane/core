# User Impersonation

User impersonation allows authorized administrators to act on behalf of other users for support, administrative, or testing purposes. This feature is critical for providing customer support and debugging user-specific issues while maintaining a complete audit trail.

## Overview

The impersonation system provides a secure way for administrators to temporarily assume the identity of another user while maintaining full accountability through audit logging. All actions performed during an impersonation session are executed with the target user's permissions but are logged with both the impersonator's and target user's context.

## Impersonation Token Flow

The impersonation process follows these steps:

1. **Authentication**: The system administrator authenticates using their standard credentials (JWT, Personal Access Token, or API token)

1. **Permission Validation**: The `StartImpersonation` handler validates that the requesting user has system administrator privileges

1. **Token Creation**: The `TokenManager` creates a signed JWT containing impersonation claims that include both the impersonator and target user context

1. **Token Usage**: The client includes the impersonation token in subsequent requests using the `Authorization: Impersonation <token>` header (where token contains only alphanumeric characters, underscores, hyphens, and periods)

1. **Middleware Processing**: The impersonation middleware intercepts requests, validates the impersonation token, and sets the impersonated user context

1. **Request Execution**: All API requests are executed with the target user's permissions and organizational context

1. **Audit Logging**: Every action is logged with complete context including who performed the action, who they were impersonating, and why

## Token Structure

Impersonation tokens are JWTs that contain standard claims plus additional impersonation-specific claims:

### Standard JWT Claims

- `iss` (issuer): Token issuer identification
- `aud` (audience): Intended token recipient
- `exp` (expiration): Token expiration timestamp
- `iat` (issued at): Token creation timestamp
- `jti` (JWT ID): Unique token identifier

### Impersonation Claims

- `user_id`: The ID of the target user being impersonated
- `target_user_email`: Email address of the target user
- `impersonator_id`: The ID of the system administrator performing the impersonation
- `impersonator_email`: Email address of the impersonator
- `org_id`: Organization ID for the impersonation context
- `type`: The type of impersonation (support, admin, or job)
- `reason`: A required field explaining why impersonation is necessary (minimum 10 characters)
- `session_id`: A unique identifier for this impersonation session
- `scopes`: An array of allowed actions during the impersonation session

## Impersonation Types

### Support Impersonation

- **Purpose**: Customer support and debugging
- **Default Scopes**: `["read", "debug"]`
- **Permissions Required**: System administrator privileges
- **Use Cases**: Viewing user data, reproducing issues, investigating problems

### Admin Impersonation

- **Purpose**: Administrative tasks and system management
- **Default Scopes**: `["*"]` (full access)
- **Permissions Required**: System administrator privileges
- **Use Cases**: Correcting data issues, performing administrative actions on behalf of users

### Job Impersonation

- **Purpose**: System jobs and automated processes
- **Default Scopes**: `["read", "write"]`
- **Permissions Required**: System administrator privileges
- **Use Cases**: Batch processing, automated maintenance, system migrations

## Security Considerations

### Access Control

- Only users with system administrator privileges can initiate impersonation sessions
- Cross-organization impersonation requires system admin privileges
- Impersonation type determines the level of access granted

### Audit Trail

- All impersonation sessions are logged with:
  - Session start and end times
  - Impersonator and target user details
  - Reason for impersonation
  - IP address and user agent
  - **Note**: Currently logs to application logs only. Database persistence for audit trails is planned for future implementation.

### Token Security

- Tokens are cryptographically signed and validated
- Configurable expiration times (default: 1 hour)
- Session IDs enable tracking and revocation
- Tokens cannot be refreshed - new sessions must be explicitly created

### Scope Restrictions

- Actions are limited by the scopes assigned to the impersonation session
- Certain sensitive operations may be blocked entirely during impersonation
- Middleware can enforce additional restrictions based on impersonation type

## API Usage

### Starting an Impersonation Session

```http
POST /v1/impersonation/start
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "target_user_id": "user_123",
  "type": "support",
  "reason": "Investigating reported login issue",
  "duration": 2,  // hours
  "scopes": ["read", "debug"],
  "organization_id": "org_456"  // optional, defaults to current org
}
```

Response:
```json
{
  "success": true,
  "token": "imp_...",
  "expires_at": "2024-01-15T14:00:00Z",
  "session_id": "sess_789",
  "message": "Impersonation session started successfully"
}
```

### Using the Impersonation Token

Include the impersonation token in the `Authorization` header with `Impersonation` scheme:

```http
GET /v1/user/profile
Authorization: Impersonation imp_...
```

### Ending an Impersonation Session

```http
POST /v1/impersonation/end
Authorization: Impersonation imp_...
Content-Type: application/json

{
  "session_id": "sess_789",
  "reason": "Support task completed"
}
```

## Middleware Integration

The impersonation middleware can be configured to:

- **Block Impersonation**: Prevent impersonated users from accessing certain endpoints
- **Require Specific Scopes**: Ensure the impersonation session has necessary permissions
- **Filter by Type**: Allow only specific impersonation types for certain operations

Example middleware usage:

```go
// Block all impersonation
e.POST("/sensitive-action", handler, impersonation.BlockImpersonation())

// Require specific scope
e.GET("/admin-data", handler, impersonation.RequireImpersonationScope("admin:read"))

// Allow only support impersonation
e.GET("/debug-info", handler, impersonation.AllowOnlyImpersonationType(auth.SupportImpersonation))
```

## Best Practices

1. **Always provide a clear reason** when starting an impersonation session
2. **Use the minimum necessary scope** for the task at hand
3. **End sessions promptly** when the task is complete
4. **Monitor audit logs** regularly for unusual impersonation activity
5. **Implement alerts** for sensitive operations performed under impersonation
6. **Test thoroughly** with impersonation tokens in non-production environments
7. **Document** all standard impersonation procedures for your support team

## Current Limitations

The following limitations exist in the current implementation:

- **Permission System**: All impersonation types (support, admin, job) currently require system administrator privileges. There is no separate "support staff" role implemented.
- **Self-Impersonation**: Validation to prevent users from impersonating themselves is currently disabled for testing purposes.
- **Audit Persistence**: Impersonation events are currently logged to application logs only. Database persistence for complete audit trails is planned for future implementation.
- **Organization Validation**: The system does not actively validate organization membership for target users, relying instead on permission checks in the calling functions.

## Auditing

- Maintaining audit logs of all impersonation activity (currently in application logs)
- Requiring explicit reasons for each impersonation session (minimum 10 characters)
- Implementing time-limited sessions with no refresh capability
- Providing clear attribution of all actions to both the impersonator and target user
- **Note**: Full database audit persistence is planned for future implementation to enhance compliance capabilities.
