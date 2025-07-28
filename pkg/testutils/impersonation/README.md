# Impersonation Demo

This demo demonstrates the impersonation functionality for admin users to assume another user's context.

## Overview

The impersonation system allows system administrators to temporarily act as another user for:
- **Support**: Help users debug issues with limited access
- **Admin**: Full administrative operations as another user
- **Job**: Automated job execution in user context

## How to Use

### 1. Prerequisites

- OpenLane server running on `http://localhost:17608` (or update the API host)
- System admin privileges (JWT token, PAT, or API token)
- Target user ID to impersonate

### 2. Getting System Admin Token

You need a system admin token. Here are the options:

#### Option A: JWT Token (from login)
```bash
# Login as system admin user
./openlane-cli login -u admin@admin.theopenlane.io

# The JWT token is stored in ~/.config/openlane/credentials
cat ~/.config/openlane/credentials
```

#### Option B: Personal Access Token (PAT)
```bash
# Create PAT for system admin user
./openlane-cli pat create -n "impersonation-demo" -o <org-id>

# Grant system admin privileges via FGA
task cli:setup-system-admin
```

#### Option C: API Token
```bash
# Create API token
./openlane-cli token create -n "impersonation-demo" --scopes=read,write

# Grant system admin privileges via FGA
fga tuple write --store-id=<store-id> service:<token-id> system_admin system:openlane_core --api-token=<fga-token>
```

### 3. Getting Target User ID

Find a user to impersonate:

```bash
# List users
./openlane-cli user get

# Or via GraphQL
curl -X POST http://localhost:17608/query \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { users { edges { node { id email } } } }"}'
```

### 4. Using the Demo

1. **Open the Demo**: Open `index.html` in your browser
2. **Configure API**: Enter your API host and system admin token
3. **Test Connection**: Click "Test Connection" to verify your admin token works
4. **Start Impersonation**:
   - Enter the target user ID
   - Select impersonation type (admin/support/job)
   - Provide a detailed reason (required for audit)
   - Set duration (1-24 hours)
   - Optionally specify organization ID
5. **Get Impersonation Token**: After starting, you'll receive a JWT token
6. **Test Impersonated Requests**: Use the token with `Impersonation: Bearer <token>` header
7. **End Session**: End the impersonation session when done

## API Endpoints

The demo uses these endpoints:

### Start Impersonation
```
POST /v1/impersonation/start
Authorization: Bearer <system-admin-token>
Content-Type: application/json

{
  "target_user_id": "user-id-to-impersonate",
  "type": "admin|support|job",
  "reason": "Detailed reason for impersonation",
  "duration_hours": 1,
  "organization_id": "optional-org-id"
}
```

**Response:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJSUzI1NiIs...",
  "expires_at": "2024-01-01T12:00:00Z",
  "session_id": "session-uuid",
  "message": "Impersonation session started successfully"
}
```

### Using Impersonation Token
```
GET /v1/users/me
Authorization: Impersonation eyJhbGciOiJSUzI1NiIs...
```

### End Impersonation
```
POST /v1/impersonation/end
Authorization: Impersonation <impersonation-token>
Content-Type: application/json

{
  "session_id": "session-uuid",
  "reason": "Task completed"
}
```

## Security Features

- **Audit Logging**: All impersonation activity is logged
- **Time-Limited**: Tokens expire automatically (1-24 hours)
- **Scope-Based**: Different impersonation types have different permissions
- **System Admin Only**: Only system admins can create impersonation tokens
- **Session Tracking**: Each impersonation session has a unique ID

## Impersonation Types

### Admin (`admin`)
- **Permissions**: Full access (`*` scope)
- **Use Case**: Administrative operations, data fixes
- **Restrictions**: System admin only

### Support (`support`)
- **Permissions**: Limited read/debug access
- **Use Case**: Customer support, debugging user issues
- **Restrictions**: System admin or support role

### Job (`job`)
- **Permissions**: Read/write access for automated tasks
- **Use Case**: Background jobs, automated processes
- **Restrictions**: System admin only

## Troubleshooting

### Common Issues

1. **403 Forbidden**: Your token doesn't have system admin privileges
   - Verify FGA system admin tuple is set correctly
   - Use `task cli:setup-system-admin` to grant privileges

2. **401 Unauthorized**: Invalid or expired token
   - Verify your admin token is valid
   - Re-login or recreate token if expired

3. **Target User Not Found**: Invalid user ID
   - Verify the user ID exists in the system
   - Check organization membership if specified

4. **Connection Failed**: Server not running or wrong host
   - Ensure OpenLane server is running
   - Verify API host URL is correct

### Testing Without UI

You can test the API directly:

```bash
# Start impersonation
curl -X POST http://localhost:17608/v1/impersonation/start \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "target_user_id": "target-user-id",
    "type": "admin",
    "reason": "Testing impersonation functionality"
  }'

# Use impersonation token
curl -X GET http://localhost:17608/v1/users/me \
  -H "Authorization: Impersonation <impersonation-token>"

# End impersonation
curl -X POST http://localhost:17608/v1/impersonation/end \
  -H "Authorization: Impersonation <impersonation-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "session-id",
    "reason": "Testing completed"
  }'
```

## Architecture

```
[Admin User] → [Start Impersonation] → [Impersonation Token]
                     ↓
[TokenManager] → [JWT with Claims]
                     ↓
[Client] → [API Request with Impersonation Header]
                     ↓
[ImpersonationMiddleware] → [Validate Token] → [Set User Context]
                     ↓
[API Handler] → [Execute as Target User] → [Audit Log]
```

The impersonation token contains:
- Impersonator ID and email
- Target user ID and email
- Impersonation type and reason
- Session ID and expiration
- Allowed scopes

All impersonation activity is logged for compliance and security auditing.