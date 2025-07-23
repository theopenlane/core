# OAuth Integration Testing Interface

This HTML interface provides comprehensive end-to-end testing for OAuth integrations with GitHub and Slack.

## Setup and Usage

### 0. Prerequisites: OAuth App Configuration

**⚠️ Important**: Before testing, you must configure OAuth applications:

#### GitHub OAuth App Setup

**⚠️ Important**: OAuth integrations use your **existing GitHub OAuth app** (configured for social login).

1. **Update Existing OAuth App**: Go to your existing GitHub OAuth app settings
2. **Add Integration Callback URL**: Ensure your app has BOTH callback URLs:
   ```
   http://localhost:17608/v1/github/callback              # For social login (existing)
   http://localhost:17608/v1/integrations/oauth/callback  # For integrations (add this)
   ```
3. **No New Configuration Needed**: Your existing `config.yaml` already has the credentials:

```yaml
# config.yaml (your existing social login config works!)
auth:
  providers:
    github:
      clientId: "your_github_client_id"        # ← Already configured
      clientSecret: "your_github_client_secret" # ← Already configured
      clientEndpoint: "http://localhost:17608" # ← Already configured
      redirectUrl: "/v1/github/callback"       # ← For social login
```

The integration system automatically uses this existing configuration.

### 1. Start the Local Development Environment

```bash
# Start the full stack with seeded data and OAuth config
task run-dev
```

This will start:
- Core API server on `http://localhost:17608`
- Database with seeded users and organizations
- OAuth handlers configured with your GitHub app
- All necessary services

### 2. Open the Testing Interface

**Option A - Nginx (Recommended):**
```bash
# Start the OAuth test UI
task oauth-test-ui

# Open in browser
open http://localhost:3004/
```

**Option B - Direct from server:**
Navigate to: `http://localhost:17608/pkg/testutils/integrations/index.html`

**🔒 Security Note**: This HTML file contains NO OAuth credentials. It works by:
1. Making API calls to your local server (`localhost:17608`)
2. Server uses its configured OAuth app credentials
3. Server returns OAuth URLs for browser redirection
4. No sensitive information is exposed to the browser

**Do NOT serve this file locally** (`file://`) as it won't be able to make API calls to the server.

### 3. Authenticate

Before testing integrations, you need to be authenticated:

1. **Option A**: Use the login link in the interface
2. **Option B**: Login via existing flow at `http://localhost:17608/v1/github/login`
3. **Option C**: Use seeded credentials if available

### 4. Test OAuth Integration Flows

#### GitHub Integration:
1. Click **"Connect GitHub"** button
2. You'll be redirected to GitHub OAuth (or local mock)
3. Authorize the application
4. You'll be redirected back with the integration created
5. Use **"Check Status"** to verify connection
6. Use **"Get Token"** to retrieve stored tokens
7. Use **"Test /user API Call"** to verify tokens work

#### Slack Integration:
1. Click **"Connect Slack"** button
2. Complete OAuth flow (requires Slack app configuration)
3. Test using similar buttons as GitHub

## Features

### 🔐 **Authentication Status**
- Shows current user authentication state
- Displays user name and email when authenticated

### 🔗 **Integration Management**
- Start OAuth flows for GitHub/Slack
- Check integration connection status
- Retrieve and display stored tokens (with preview)
- Refresh expired tokens
- Disconnect/delete integrations

### 📋 **Integration Listing**
- View all configured integrations for the organization
- Shows integration names, providers, and IDs

### 🧪 **API Testing**
- Test GitHub API calls using stored tokens
- Test Slack API calls using stored tokens
- Verify tokens are working correctly

### 📝 **Activity Logging**
- Real-time log of all actions and responses
- Success/error/info message types
- Timestamps for all activities

## API Endpoints Tested

The interface tests these OAuth integration endpoints:

- `POST /v1/integrations/oauth/start` - Start OAuth flow
- `POST /v1/integrations/oauth/callback` - Handle OAuth callback
- `GET /v1/integrations/:provider/status` - Check integration status
- `GET /v1/integrations/:provider/token` - Get stored tokens
- `POST /v1/integrations/:provider/refresh` - Refresh tokens
- `GET /v1/integrations` - List all integrations
- `DELETE /v1/integrations/:id` - Delete integration

## OAuth Flow Testing

### Complete End-to-End Flow:
1. ✅ **Start Flow**: Interface initiates OAuth with proper scopes
2. ✅ **Redirect**: User redirected to OAuth provider
3. ✅ **Callback**: OAuth callback processed and tokens stored
4. ✅ **Verification**: Tokens validated and integration created
5. ✅ **API Access**: Stored tokens used for actual API calls
6. ✅ **Management**: Status checking, token refresh, disconnection

### Verification Points:
- ✅ Integration created in database
- ✅ Tokens stored securely in Hush secrets
- ✅ Organization-specific scoping
- ✅ Token expiry handling
- ✅ API authentication working
- ✅ Clean disconnection/deletion

## Configuration Requirements

### GitHub Integration:
- Requires `OauthProvider.Github` configuration in your settings
- OAuth app configured with callback URL: `http://localhost:17608/v1/integrations/oauth/callback`

### Slack Integration:
- Requires Slack app configuration (currently placeholder)
- When implemented, needs `OauthProvider.Slack` configuration

## Troubleshooting

### Common Issues:

1. **"Not authenticated"**: Login first using `/v1/github/login`
2. **"Provider not configured"**: Check OAuth provider configuration
3. **"Invalid redirect URI"**: Ensure OAuth app callback URL matches
4. **"Token validation failed"**: Check OAuth app permissions/scopes

### Debug Tips:

- Check the Activity Log for detailed error messages
- Use browser dev tools to inspect network requests
- Verify authentication token in URL params after login
- Check that `task run-dev` is running and healthy

## Expected Behavior

With a working setup, you should see:

1. ✅ **Authentication successful** - Shows user info
2. ✅ **GitHub integration connects** - OAuth flow completes
3. ✅ **Status shows "Connected"** - Integration verified
4. ✅ **Tokens retrieved** - Access/refresh tokens displayed
5. ✅ **API test passes** - GitHub /user call succeeds
6. ✅ **Integration listed** - Shows in "All Integrations" section

This provides complete verification that your OAuth integration handlers are working correctly end-to-end!