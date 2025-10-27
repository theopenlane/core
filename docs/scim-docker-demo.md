# SCIM Docker Demo Setup

This directory contains a Docker Compose setup for testing and demonstrating the SCIM 2.0 provisioning capabilities of Openlane.

## Overview

The setup includes:
- A Python-based SCIM client for testing SCIM endpoints
- Example scripts for user and group provisioning workflows
- Docker Compose configuration for easy deployment

## Prerequisites

1. Openlane server running locally (default: http://localhost:17608)
2. A valid Openlane authentication token with SCIM permissions
3. Docker and Docker Compose installed

## Quick Start

The easiest way to run the SCIM demo is using the Taskfile commands, which handle token generation automatically.

### Prerequisites

1. Openlane server running locally (default: http://localhost:17608)
2. Openlane CLI built (`task go:build-cli`)
3. An authenticated user session (run `task cli:login` first)

### Option 1: Using Taskfile (Recommended)

Run the complete SCIM demo with automatic token generation:

```bash
# Full demo - starts compose, generates token, runs demo
task scim:demo:full

# Or run steps individually:
task scim:demo:up    # Start the SCIM client container
task scim:demo:run   # Run the basic demo with auto-generated token
task scim:demo:team  # Run the team provisioning demo
task scim:demo:down  # Stop the container
```

The demo will:
1. Create a test user
2. Create a test group with the user as a member
3. List all users and groups
4. Deactivate and reactivate the user
5. Clean up by deleting the test user and group

### Option 2: Manual Setup

If you prefer manual setup or need custom configuration:

```bash
# 1. Set environment variables
export SCIM_BASE_URL=http://host.docker.internal:17608/scim
export SCIM_AUTH_TOKEN=tola_your_token_here

# 2. Start the container
docker compose --profile scim-demo -f docker-compose.scim-demo.yml up -d

# 3. Run the demo
docker compose -f docker-compose.scim-demo.yml exec scim-client python /app/scripts/scim_client.py
```

## Available Taskfile Commands

| Command | Description |
|---------|-------------|
| `task scim:demo:full` | Complete workflow: start compose, generate token, run demo |
| `task scim:demo:up` | Start the SCIM client container |
| `task scim:demo:down` | Stop and remove the SCIM client container |
| `task scim:demo:run` | Run the basic SCIM demo with auto-generated token |
| `task scim:demo:team` | Run the team provisioning demo with auto-generated token |
| `task scim:demo:shell` | Open an interactive shell with auto-generated token |
| `task scim:demo:logs` | View logs from the SCIM client container |

## Manual Testing

You can use the interactive shell for manual testing:

```bash
# Open shell with auto-generated token
task scim:demo:shell

# Or manually enter the container
docker compose -f docker-compose.scim-demo.yml exec scim-client bash
```

Inside the container, use the Python SCIM client library:

```python
from scim_client import SCIMClient
import os

client = SCIMClient(os.getenv('SCIM_BASE_URL'), os.getenv('SCIM_AUTH_TOKEN'))

# Create a user
user = client.create_user('alice@example.com', 'Alice', 'Wonderland')
print(user)

# Create a group
group = client.create_group('Developers', members=[user['id']])
print(group)

# List users
users = client.list_users()
print(users)
```

## SCIM Endpoints

The following SCIM 2.0 endpoints are available:

### Users
- `POST /scim/v2/Users` - Create a user
- `GET /scim/v2/Users/{id}` - Get a user
- `GET /scim/v2/Users` - List users
- `PATCH /scim/v2/Users/{id}` - Update a user
- `PUT /scim/v2/Users/{id}` - Replace a user
- `DELETE /scim/v2/Users/{id}` - Delete a user

### Groups
- `POST /scim/v2/Groups` - Create a group
- `GET /scim/v2/Groups/{id}` - Get a group
- `GET /scim/v2/Groups` - List groups
- `PATCH /scim/v2/Groups/{id}` - Update a group
- `PUT /scim/v2/Groups/{id}` - Replace a group
- `DELETE /scim/v2/Groups/{id}` - Delete a group

### Service Provider Configuration
- `GET /scim/ServiceProviderConfig` - Get SCIM service provider configuration
- `GET /scim/Schemas` - Get supported SCIM schemas
- `GET /scim/ResourceTypes` - Get supported SCIM resource types

## Testing with cURL

You can also test the SCIM endpoints directly with cURL:

```bash
# Get service provider configuration
curl -H "Authorization: Bearer tola_your_token_here" \
     http://localhost:17608/scim/ServiceProviderConfig

# Create a user
curl -X POST \
     -H "Authorization: Bearer tola_your_token_here" \
     -H "Content-Type: application/scim+json" \
     -d '{
       "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
       "userName": "bob@example.com",
       "name": {
         "givenName": "Bob",
         "familyName": "Builder"
       },
       "displayName": "Bob Builder",
       "active": true
     }' \
     http://localhost:17608/scim/v2/Users

# List users
curl -H "Authorization: Bearer tola_your_token_here" \
     http://localhost:17608/scim/v2/Users
```

## Integration with Real IdPs

### Okta Integration

To configure Okta to provision users to Openlane:

1. In Okta Admin Console, go to Applications > Applications
2. Click "Create App Integration"
3. Select "SWA - Secure Web Authentication" or "SAML 2.0"
4. After creating, go to the "Provisioning" tab
5. Click "Configure API Integration"
6. Enter the SCIM details:
   - Base URL: `https://your-openlane-instance.com/scim2`
   - Authentication: Bearer Token
   - Token: Your Openlane token
7. Enable provisioning features (Create Users, Update User Attributes, Deactivate Users)

### Azure AD Integration

To configure Azure AD to provision users to Openlane:

1. In Azure AD, go to Enterprise Applications
2. Create a new non-gallery application
3. Go to Provisioning and set Provisioning Mode to Automatic
4. Enter SCIM connection details:
   - Tenant URL: `https://your-openlane-instance.com/scim2`
   - Secret Token: Your Openlane token
5. Test the connection and save
6. Configure attribute mappings as needed
7. Start provisioning

## Cleanup

To stop and remove the SCIM client container:

```bash
docker-compose -f docker-compose.scim-demo.yml down
```

## Troubleshooting

### Connection Refused

If you get connection errors, make sure:
- Openlane server is running
- The SCIM_BASE_URL is correct (use `host.docker.internal` for Docker Desktop on Mac/Windows)
- Your firewall allows Docker containers to access the host

### Authentication Errors

If you get 401 Unauthorized errors:
- Verify your token is valid
- Ensure your token has the necessary permissions for SCIM operations
- Check that the token is properly prefixed with `tola_` (or whatever your token prefix is)

### SCIM Endpoint Not Found

If you get 404 errors:
- Verify the SCIM routes are registered in your Openlane server
- Check that the base URL path is `/scim2` (not `/scim/v2`)
- Ensure the SCIM handler is enabled in your Openlane configuration

## Further Reading

- [SCIM 2.0 Specification](https://scim.cloud/)
- [RFC 7643 - SCIM Core Schema](https://tools.ietf.org/html/rfc7643)
- [RFC 7644 - SCIM Protocol](https://tools.ietf.org/html/rfc7644)
