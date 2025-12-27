# OAuth Integration System

This document describes the OAuth integration system that enables secure connections between Openlane and third-party services like GitHub, Slack, and other OAuth providers.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [OAuth Flow](#oauth-flow)
4. [Security Model](#security-model)
5. [Token Management](#token-management)
6. [Token Invalidation Detection](#token-invalidation-detection)
7. [API Endpoints](#api-endpoints)
8. [External System Integration](#external-system-integration)
9. [Configuration](#configuration)
10. [Testing](#testing)
11. [Troubleshooting](#troubleshooting)

## Overview

The OAuth integration system provides a secure, organization-scoped mechanism for users to authorize Openlane to access their third-party accounts (GitHub, Slack, etc.). This enables features like:

- Repository access and management via GitHub
- Team communication through Slack
- Automated workflows across integrated services
- Secure API calls on behalf of users

### Key Features

* Organization-Scoped: All tokens are stored per organization, ensuring proper data isolation
* Secure Token Storage: OAuth tokens encrypted and stored using the Hush secrets management system
* Token Lifecycle Management: Automatic refresh handling and expiry detection
* Invalidation Detection: Proactive detection of revoked authorizations
* Provider Extensibility: Easy addition of new OAuth providers

## Architecture

### System Components

```mermaid
graph TB
    subgraph "Frontend Layer"
        UI[OAuth Testing Interface]
    end

    subgraph "API Layer"
        AUTH[Authentication Middleware]
        OAUTH[OAuth Flow Handlers]
        MGMT[Token Management]
    end

    subgraph "Storage Layer"
        DB[(Core Database)]
        HUSH[(Hush Secrets)]
    end

    subgraph "External Services"
        GITHUB[GitHub OAuth]
        SLACK[Slack OAuth]
    end

    UI --> AUTH
    AUTH --> OAUTH
    OAUTH --> MGMT
    MGMT --> DB
    MGMT --> HUSH
    OAUTH --> GITHUB
    OAUTH --> SLACK
```

### OAuth Flow Sequence

```mermaid
sequenceDiagram
    participant User
    participant UI as Frontend UI
    participant API as OAuth Handler
    participant Provider as OAuth Provider
    participant DB as Database
    participant Hush as Hush Secrets

    User->>UI: Click "Connect GitHub"
    UI->>API: POST /v1/integrations/oauth/start
    API->>API: Generate secure state
    API->>API: Set OAuth cookies
    API-->>UI: Return OAuth URL
    UI->>Provider: Redirect to OAuth URL
    Provider->>User: Show authorization page
    User->>Provider: Grant permissions
    Provider->>API: Redirect to callback with code
    API->>API: Validate state parameter
    API->>Provider: Exchange code for tokens
    Provider-->>API: Return access/refresh tokens
    API->>Provider: Validate tokens with API call
    Provider-->>API: Return user info
    API->>DB: Create integration entity
    API->>Hush: Store encrypted tokens
    API-->>UI: Redirect with success
    UI->>User: Show connected status
```

### Data Flow

1. Authorization Request: User initiates OAuth flow through UI
1. State Generation: Server creates secure state parameter with org context
1. Provider Redirect: User redirected to OAuth provider (GitHub, Slack)
1. User Authorization: User grants permissions on provider's site
1. Callback Processing: Provider redirects back with authorization code
1. Token Exchange: Server exchanges code for access/refresh tokens
1. Secure Storage: Tokens encrypted and stored with organization scope
1. API Access: Stored tokens used for subsequent API calls

## OAuth Flow

### 1. Initiate OAuth Flow

**Endpoint**: `POST /v1/integrations/oauth/start`

```json
{
  "provider": "github",
  "scopes": ["repo", "user:email"],
  "redirectUri": "https://app.openlane.io/integrations/callback"
}
```

**Process**:
1. Validate authenticated user and organization context
1. Generate cryptographically secure state parameter containing:
   * Organization ID
   * Provider name
   * Random entropy (16 bytes)
   * Base64 URL-encoded for safety
1. Build OAuth authorization URL with appropriate scopes
1. Return authorization URL for client redirect

**Security Considerations**:
* State parameter prevents CSRF attacks
* Organization context embedded in state ensures proper scoping
* Random entropy prevents state guessing attacks
* HTTPS required for all redirects

### 2. OAuth Provider Authorization

User is redirected to the OAuth provider (e.g., `github.com/login/oauth/authorize`) where they:
1. Review requested permissions
1. Grant or deny authorization
1. Are redirected back to Openlane with authorization code

### 3. Handle OAuth Callback

**Endpoint**: `POST /v1/integrations/oauth/callback`

```json
{
  "provider": "github",
  "code": "gho_16C7e42F292c6912E7710c838347Ae178B4a",
  "state": "eyJvcmdJRCI6IjAxSE..."
}
```

**Process**:
1. **State Validation**: Decode and verify state parameter
   * Extract organization ID and provider
   * Validate format and authenticity
1. **Code Exchange**: Exchange authorization code for tokens
   * Use provider's token endpoint
   * Obtain access token, refresh token, and expiry
1. **Token Validation**: Verify token with provider API
   * Make test API call to validate token
   * Extract user information from provider
1. **Secure Storage**: Store tokens in Hush secrets system
   * Encrypt tokens at rest
   * Associate with integration entity
   * Store provider user metadata

## Security Model

### Authentication & Authorization

* **User Authentication**: All OAuth flows require authenticated Openlane users
* **Organization Scoping**: Tokens are always associated with user's current organization
* **Privacy Bypass**: System operations use privacy bypass for secure token storage
* **Audit Trail**: All token operations logged for security monitoring

### Token Security

#### Storage Security
* **Encryption at Rest**: All tokens encrypted using Hush secrets system
* **Access Control**: Tokens only accessible by organization members
* **Immutable Updates**: Token updates require delete/recreate pattern
* **Secret Rotation**: Support for token refresh and rotation

#### Transmission Security
* **HTTPS Only**: All OAuth communications over TLS
* **State Parameters**: CSRF protection through cryptographic state
* **Secure Redirects**: Validated redirect URIs prevent open redirect attacks
* **Token Masking**: Tokens never exposed in logs or error messages

#### Provider-Specific Security

##### GitHub
```go
// Default scopes for API access
Scopes: []string{"read:user", "user:email", "repo"}

// Token validation
GET https://api.github.com/user
Authorization: token {access_token}
```

##### Slack (when configured)
```go
// Default scopes for team communication
Scopes: []string{"channels:read", "chat:write", "users:read"}

// Token validation
POST https://slack.com/api/auth.test
Authorization: Bearer {access_token}
```

### Security Best Practices

1. **Minimal Scopes**: Request only necessary permissions
2. **Token Expiry**: Respect and enforce token expiration
3. **Refresh Rotation**: Rotate refresh tokens when possible
4. **Error Handling**: Never expose tokens in error responses
5. **Audit Logging**: Log all token operations for monitoring

## Token Management

### Token Storage Schema

Tokens are stored in the Hush secrets system with the following structure:

```
Integration Entity:
├── id: "01HX..."
├── owner_id: "{organization_id}"
├── name: "GitHub Integration (username)"
├── kind: "github"
└── description: "OAuth integration with GitHub for {username}"

Associated Hush Secrets:
├── {provider}_access_token: "encrypted_access_token"
├── {provider}_refresh_token: "encrypted_refresh_token"
├── {provider}_expires_at: "2024-12-31T23:59:59Z"
├── {provider}_provider_user_id: "12345"
├── {provider}_provider_username: "johndoe"
└── {provider}_provider_email: "john@example.com"
```

### Token Lifecycle

#### Creation
1. OAuth callback successfully processed
2. Integration entity created in database
3. Tokens encrypted and stored in Hush
4. Provider user metadata cached

#### Refresh
```go
// Automatic refresh when tokens expire
func (h *Handler) RefreshIntegrationToken(ctx context.Context, orgID, provider string) error {
    // 1. Retrieve current tokens
    // 2. Use refresh token to get new access token
    // 3. Validate new token with provider
    // 4. Update stored tokens atomically
    // 5. Return fresh token data
}
```

#### Deletion
1. Remove all associated Hush secrets
2. Delete integration entity
3. Audit log deletion event

## Token Invalidation Detection

### Proactive Detection Mechanisms

#### 1. API Call Monitoring
Every API call made with stored tokens includes response monitoring:

```go
// Example: GitHub API call with invalidation detection
func (s *GitHubService) makeAPICall(token string, endpoint string) (*Response, error) {
    resp, err := http.Get(endpoint, headers{"Authorization": "token " + token})

    switch resp.StatusCode {
    case 401:
        // Token invalid - mark integration as invalid
        s.markIntegrationInvalid("github", "Token unauthorized")
        return nil, ErrTokenInvalid
    case 403:
        if isRateLimited(resp) {
            // Rate limited - temporary issue
            return nil, ErrRateLimit
        }
        // Forbidden - possibly revoked permissions
        s.markIntegrationSuspicious("github", "API access forbidden")
    }

    return resp, err
}
```

#### 2. Periodic Token Validation
Background service that validates tokens periodically:

```go
// Runs every 4 hours to validate active integrations
func (s *IntegrationService) validateActiveTokens() {
    integrations := s.getActiveIntegrations()

    for _, integration := range integrations {
        if err := s.validateIntegrationToken(integration); err != nil {
            s.handleTokenInvalidation(integration, err)
        }
    }
}
```

#### 3. Webhook Integration (Future Enhancement)
For providers that support webhooks, register for token revocation events:

```go
// GitHub App webhook handler
func (h *WebhookHandler) handleInstallationDeleted(event *GitHubEvent) {
    // Mark all tokens for this installation as invalid
    h.invalidateTokensForInstallation(event.Installation.ID)
}
```

### Invalidation Response Strategies

#### 1. Immediate Invalidation
When token is confirmed invalid:
```go
func (s *IntegrationService) markIntegrationInvalid(orgID, provider, reason string) {
    // 1. Update integration status to "invalid"
    // 2. Add invalidation timestamp and reason
    // 3. Trigger UI notification
    // 4. Log security event
    // 5. Optionally attempt token refresh if refresh token available
}
```

#### 2. Graceful Degradation
For suspected but unconfirmed invalidation:
```go
func (s *IntegrationService) markIntegrationSuspicious(orgID, provider, reason string) {
    // 1. Update integration status to "warning"
    // 2. Increase validation frequency
    // 3. Show warning in UI
    // 4. Continue attempting API calls with backoff
}
```

### User Notification System

#### UI Indicators
Integration status displayed in real-time:

```typescript
interface IntegrationStatus {
  provider: string;
  connected: boolean;
  status: 'connected' | 'expired' | 'invalid' | 'warning';
  tokenValid: boolean;
  tokenExpired: boolean;
  lastValidated: Date;
  message: string;
  actions: ('reconnect' | 'refresh' | 'configure')[];
}
```

#### Status Messages
```json
{
  "github": {
    "status": "invalid",
    "message": "GitHub integration has been revoked. Please reconnect to continue using GitHub features.",
    "actions": ["reconnect"],
    "lastError": "Token validation failed with 401 Unauthorized",
    "detectedAt": "2024-01-15T10:30:00Z"
  }
}
```

#### Notification Channels
1. **In-App Notifications**: Real-time status updates in UI
2. **Email Alerts**: For critical integration failures
3. **Slack/Teams**: For team administrators
4. **API Events**: Webhook notifications to client applications

### Auto-Recovery Mechanisms

#### 1. Automatic Token Refresh
```go
func (s *IntegrationService) attemptTokenRefresh(integration *Integration) error {
    if !integration.HasRefreshToken() {
        return ErrNoRefreshToken
    }

    newToken, err := s.refreshOAuthToken(integration)
    if err != nil {
        // Refresh failed - user must re-authorize
        s.markIntegrationInvalid(integration.OrgID, integration.Provider,
            "Refresh token invalid - re-authorization required")
        return err
    }

    // Refresh successful - update stored tokens
    s.updateStoredTokens(integration, newToken)
    s.markIntegrationValid(integration.OrgID, integration.Provider)
    return nil
}
```

#### 2. Smart Retry Logic
```go
func (s *APIClient) makeRequestWithRetry(req *Request) (*Response, error) {
    resp, err := s.makeRequest(req)

    if err != nil && isTokenError(err) {
        // Attempt token refresh
        if refreshErr := s.refreshToken(); refreshErr == nil {
            // Retry with new token
            return s.makeRequest(req)
        }
    }

    return resp, err
}
```

## API Endpoints

### OAuth Flow Endpoints

#### Start OAuth Flow
```http
POST /v1/integrations/oauth/start
Content-Type: application/json
Authorization: Bearer {user_session_token}

{
  "provider": "github",
  "scopes": ["repo", "gist"],
  "redirectUri": "https://app.openlane.io/callback"
}
```

**Response**:
```json
{
  "success": true,
  "authUrl": "https://github.com/login/oauth/authorize?client_id=...",
  "state": "eyJvcmdJRCI6..."
}
```

#### Handle OAuth Callback
```http
POST /v1/integrations/oauth/callback
Content-Type: application/json

{
  "provider": "github",
  "code": "gho_16C7e42F292c6912E7710c838347Ae178B4a",
  "state": "eyJvcmdJRCI6..."
}
```

**Response**:
```json
{
  "success": true,
  "message": "Successfully connected GitHub integration",
  "integration": {
    "id": "01HX...",
    "name": "GitHub Integration (johndoe)",
    "provider": "github"
  }
}
```

### Token Management Endpoints

#### Get Integration Token
```http
GET /v1/integrations/{provider}/token
Authorization: Bearer {user_session_token}
```

**Response**:
```json
{
  "success": true,
  "provider": "github",
  "token": {
    "provider": "github",
    "accessToken": "gho_16C7e42F292c6912E7710c838347Ae178B4a",
    "refreshToken": "ghr_1B4a6cdad02f005f329FW1qn4a1...",
    "expiresAt": "2024-12-31T23:59:59Z",
    "providerUserId": "12345",
    "providerUsername": "johndoe"
  }
}
```

#### Check Integration Status
```http
GET /v1/integrations/{provider}/status
Authorization: Bearer {user_session_token}
```

**Response**:
```json
{
  "success": true,
  "provider": "github",
  "connected": true,
  "status": "connected",
  "tokenValid": true,
  "tokenExpired": false,
  "lastValidated": "2024-01-15T10:30:00Z",
  "message": "GitHub integration is connected and active"
}
```

#### Refresh Integration Token
```http
POST /v1/integrations/{provider}/refresh
Authorization: Bearer {user_session_token}
```

#### List All Integrations
```http
GET /v1/integrations
Authorization: Bearer {user_session_token}
```

#### Delete Integration
```http
DELETE /v1/integrations/{integration_id}
Authorization: Bearer {user_session_token}
```

## External System Integration

### GitHub Integration

#### Configuration
```go
type IntegrationProviderConfig struct {
    ClientID     string `json:"clientId" koanf:"clientId"`
    ClientSecret string `json:"clientSecret" koanf:"clientSecret"`
    ClientEndpoint string `json:"clientEndpoint" koanf:"clientEndpoint"`
    Scopes       []string `json:"scopes" koanf:"scopes"`
}
```

#### Default Scopes
- `read:user` - Read user profile information
- `user:email` - Access user email addresses
- `repo` - Full repository access (read/write)

#### API Usage Examples
```go
// List user repositories
func (g *GitHubClient) ListRepositories(token string) ([]*Repository, error) {
    req := &http.Request{
        URL: "https://api.github.com/user/repos",
        Headers: map[string]string{
            "Authorization": "token " + token,
            "Accept": "application/vnd.github.v3+json",
        },
    }

    resp, err := g.makeRequest(req)
    if err != nil {
        return nil, g.handleAPIError(err)
    }

    return parseRepositories(resp.Body)
}

// Create repository webhook
func (g *GitHubClient) CreateWebhook(token, repo, webhookURL string) error {
    // Implementation for webhook creation
}
```

### Slack Integration

#### Configuration
```go
type IntegrationProviderConfig struct {
    ClientID     string `json:"clientId" koanf:"clientId"`
    ClientSecret string `json:"clientSecret" koanf:"clientSecret"`
    ClientEndpoint string `json:"clientEndpoint" koanf:"clientEndpoint"`
    Scopes       []string `json:"scopes" koanf:"scopes"`
}
```

#### Default Scopes
- `channels:read` - View channels in workspace
- `chat:write` - Send messages
- `users:read` - View people in workspace

### Adding New Providers

To add a new OAuth provider:

1. **Add Provider Configuration**:
```go
// Add to IntegrationOauthProviderConfig
type IntegrationOauthProviderConfig struct {
    // ... existing providers
    NewProvider IntegrationProviderConfig `json:"newprovider" koanf:"newprovider"`
}
```

2. **Implement Provider Interface**:
```go
func (h *Handler) getNewProviderConfig() *oauth2.Config {
    return &oauth2.Config{
        ClientID:     h.IntegrationOauthProvider.NewProvider.ClientID,
        ClientSecret: h.IntegrationOauthProvider.NewProvider.ClientSecret,
        RedirectURL:  fmt.Sprintf("%s/v1/integrations/oauth/callback",
                       h.IntegrationOauthProvider.NewProvider.ClientEndpoint),
        Endpoint:     newprovider.Endpoint,
        Scopes:       h.IntegrationOauthProvider.NewProvider.Scopes,
    }
}

func (h *Handler) validateNewProviderToken(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error) {
    // Implementation for token validation with provider API
}
```

3. **Update Provider Registry**:
```go
func (h *Handler) getIntegrationProviders() map[string]IntegrationProvider {
    providers := map[string]IntegrationProvider{
        // existing providers...
        "newprovider": {
            Name:   "newprovider",
            Config: h.getNewProviderConfig(),
            Validate: func(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error) {
                return h.validateNewProviderToken(ctx, token)
            },
        },
    }
    return providers
}
```

## Configuration

### OAuth App Setup

Before using the OAuth integration system, you must create OAuth applications with each provider and configure the server with the credentials.

#### GitHub App Configuration

**Important**: The OAuth integration system uses **separate OAuth configuration** from social login.

1. **Create or Update GitHub OAuth App**:
   - Go to GitHub Settings → Developer settings → OAuth Apps
   - You can either:
     - **Option A**: Create a new OAuth app specifically for integrations
     - **Option B**: Update your existing OAuth app (used for social login) to include both callback URLs
   - **Authorization callback URLs** should include:
     ```
     http://localhost:17608/v1/integrations/oauth/callback  # For integrations
     # And if sharing with social login:
     http://localhost:17608/v1/github/callback              # For social login
     ```

1. **Integration-Specific Configuration**:

The system uses a **separate configuration section** for integrations:
```yaml
# config.yaml - Integration OAuth configuration (separate from social login)
integrationOauthProvider:
  github:
    clientId: "your_github_client_id"        # Can be same as social login
    clientSecret: "your_github_client_secret" # Can be same as social login
    clientEndpoint: "http://localhost:17608" # Base URL for callbacks
    scopes: ["read:user", "user:email", "repo"] # Extended scopes for API access
  slack:
    clientId: "your_slack_client_id"         # Slack app configuration
    clientSecret: "your_slack_client_secret"
    clientEndpoint: "http://localhost:17608"
    scopes: ["channels:read", "chat:write", "users:read"]
```

**Key Differences from Social Login**:

- **Separate configuration section**: `integrationOauthProvider` vs `auth.providers`
- **Different callback URL**: `/v1/integrations/oauth/callback` vs `/v1/github/callback`
- **Extended scopes**: Includes `repo` for API access vs basic profile scopes
- **Different token storage**: Organization-scoped in Hush vs session-based
- **Different use case**: Long-term API access vs short-term user authentication

#### Environment Variables (Alternative)

```bash
# Integration OAuth Provider Configuration (separate from social login)
INTEGRATION_OAUTH_GITHUB_CLIENT_ID=your_github_client_id
INTEGRATION_OAUTH_GITHUB_CLIENT_SECRET=your_github_client_secret
INTEGRATION_OAUTH_GITHUB_CLIENT_ENDPOINT=http://localhost:17608

# Slack Integration
INTEGRATION_OAUTH_SLACK_CLIENT_ID=your_slack_client_id
INTEGRATION_OAUTH_SLACK_CLIENT_SECRET=your_slack_client_secret
INTEGRATION_OAUTH_SLACK_CLIENT_ENDPOINT=http://localhost:17608

# Security Settings
OAUTH_STATE_EXPIRY=300  # State parameter expiry in seconds
OAUTH_TOKEN_VALIDATION_INTERVAL=14400  # 4 hours
```

### **Security Note: Client Credentials**

**OAuth client secrets are NEVER exposed to the frontend**. The HTML testing interface works by:

1. **Frontend calls server**: `POST /v1/integrations/oauth/start`
1. **Server uses stored credentials**: Creates OAuth URL with server-side client ID
1. **Server returns OAuth URL**: Frontend redirects user to provider
1. **Provider redirects back**: With authorization code to server callback
1. **Server exchanges code**: Using server-side client secret

This ensures that:
- Client secrets remain secure on the server
- Credentials can be rotated without frontend changes
- No sensitive data transmitted to browsers
- Proper separation of concerns

### Database Schema

#### Integration Table
```sql
CREATE TABLE integrations (
    id VARCHAR PRIMARY KEY,
    owner_id VARCHAR NOT NULL,  -- Organization ID
    name VARCHAR NOT NULL,
    description TEXT,
    kind VARCHAR NOT NULL,      -- Provider name
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    INDEX idx_owner_kind (owner_id, kind)
);
```

#### Hush Secrets (for tokens)
```sql
CREATE TABLE hush (
    id VARCHAR PRIMARY KEY,
    owner_id VARCHAR NOT NULL,
    name VARCHAR NOT NULL,
    description TEXT,
    kind VARCHAR NOT NULL,
    secret_name VARCHAR NOT NULL,
    secret_value TEXT NOT NULL,  -- Encrypted
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    INDEX idx_owner_secret (owner_id, secret_name)
);
```

## Testing

### Test Environment Setup

#### Prerequisites
Before testing OAuth integrations, you need:

1. **GitHub OAuth App** (for GitHub integration testing):
   - Create at: https://github.com/settings/applications/new
   - Application name: "Openlane Local Development - Integrations"
   - Homepage URL: `http://localhost:17608`
   - Authorization callback URL: `http://localhost:17608/v1/integrations/oauth/callback`
   - Copy the Client ID and Client Secret

2. **Server Configuration**:

```bash
# Set environment variables or update config.yaml
export INTEGRATION_OAUTH_GITHUB_CLIENT_ID="your_github_client_id"
export INTEGRATION_OAUTH_GITHUB_CLIENT_SECRET="your_github_client_secret"
export INTEGRATION_OAUTH_GITHUB_CLIENT_ENDPOINT="http://localhost:17608"
```

Or in your `config.yaml`:

```yaml
integrationOauthProvider:
  github:
    clientId: "your_github_client_id"
    clientSecret: "your_github_client_secret"
    clientEndpoint: "http://localhost:17608"
```

#### Start Development Environment

```bash
# Start full development stack with OAuth configuration
task run-dev

# This starts:
# - Core API server on http://localhost:17608
# - Database with seeded data
# - OAuth handlers configured with your GitHub app
# - All necessary services
```

#### OAuth Testing Interface

Navigate to: `http://localhost:17608/internal/testutils/integrations/index.html`

**Important**: The testing interface does NOT contain OAuth credentials. It works by:

1. Making API calls to your local server (`localhost:17608`)
1. Server uses its configured OAuth credentials
1. Server returns OAuth URLs for redirection
1. No sensitive credentials exposed to browser

Features:

- **Authentication Status**: Shows current user login state
- **Integration Management**: Start OAuth flows, check status, get tokens
- **API Testing**: Test GitHub/Slack API calls with stored tokens
- **Activity Logging**: Real-time log of all actions and responses

#### Testing Workflow

1. **Authenticate**: Login to your local Openlane instance
2. **Configure OAuth App**: Ensure GitHub OAuth app is configured on server
3. **Test Integration**: Click "Connect GitHub" to start OAuth flow
4. **Verify Storage**: Check that tokens are stored and accessible
5. **Test API Calls**: Use stored tokens for GitHub API calls

#### Unit Tests

```bash
# Run OAuth integration tests
go test ./internal/httpserve/handlers/ -run TestOAuth

# Run OAuth model validation tests
go test ./common/models/ -run TestOAuth

# Run OpenAPI binding tests
go test ./internal/httpserve/handlers/ -run TestBind.*OAuth
```

### Test Coverage

The test suite covers:

- **Input Validation**: All request validation scenarios
- **Error Handling**: OAuth errors, token invalidation, network failures
- **State Management**: CSRF protection, state validation
- **Token Lifecycle**: Creation, refresh, expiry, deletion
- **API Integration**: Provider API calls and error handling
- **Security**: Authorization, encryption, audit logging

## Troubleshooting

### Common Issues

#### "Invalid OAuth State Parameter"

**Cause**: State parameter validation failed
**Solutions**:
- Check that state hasn't expired (5 minute default)
- Verify organization context hasn't changed
- Ensure HTTPS is used for all redirects

#### "Provider Not Configured"

**Cause**: OAuth provider configuration missing
**Solutions**:
- Verify environment variables are set
- Check OAuth app configuration in provider settings
- Ensure redirect URIs match exactly

#### "Token Validation Failed"

**Cause**: Provider rejected the token
**Solutions**:
- Check if user revoked access in provider settings
- Verify OAuth app hasn't been deleted/suspended
- Attempt token refresh if refresh token available

#### "Integration Not Found"

**Cause**: Integration deleted or organization mismatch
**Solutions**:
- Verify user is in correct organization
- Check if integration was manually deleted
- Re-run OAuth flow to recreate integration

### Security Monitoring

#### Audit Events to Monitor

- Multiple failed OAuth attempts
- Token validation failures
- Unusual API access patterns
- Integration deletions
- Token refresh failures

#### Alert Conditions

```yaml
# Example monitoring rules
alerts:
  - name: oauth_failures
    condition: oauth_errors > 10 per minute
    action: notify_security_team

  - name: token_invalidation_spike
    condition: token_validation_failures > 5 per hour per org
    action: investigate_provider_issues

  - name: suspicious_token_access
    condition: token_access outside business_hours
    action: log_security_event
```

### Debugging Tools

#### OAuth Flow Debugging

```bash
# Enable OAuth debug logging
export OAUTH_DEBUG=true

# View OAuth flow logs
tail -f logs/oauth.log | grep "oauth_flow"
```

#### Token Validation Testing

```bash
# Test token manually
curl -H "Authorization: token $GITHUB_TOKEN" \
     https://api.github.com/user
```

#### Integration Status Check

```bash
# Check integration status via API
curl -H "Authorization: Bearer $SESSION_TOKEN" \
     http://localhost:17608/v1/integrations/github/status
```

### Support and Maintenance

#### Regular Maintenance Tasks

1. **Token Cleanup**: Remove expired/invalid tokens monthly
1. **Provider Updates**: Monitor OAuth provider API changes
1. **Security Review**: Quarterly review of OAuth configurations
1. **Performance Monitoring**: Track API call success rates

#### Emergency Procedures

1. **Mass Token Invalidation**: If provider security breach detected
1. **Service Degradation**: Fallback procedures for provider outages
1. **Data Breach Response**: Token rotation and user notification procedures
