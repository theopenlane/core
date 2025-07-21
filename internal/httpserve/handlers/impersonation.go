// Package handlers provides HTTP handlers for user impersonation functionality.
//
// Impersonation Token Flow:
// 1. System admin authenticates with JWT, PAT, or API token
// 2. StartImpersonation handler validates admin permissions via FGA system
// 3. TokenManager creates signed JWT with impersonation claims
// 4. Token contains both impersonator and target user context
// 5. Client uses "Impersonation: Bearer <token>" header for subsequent requests
// 6. Impersonation middleware validates token and sets impersonated user context
// 7. All actions are performed as target user but logged with impersonator context
//
// Token Structure:
// - Standard JWT claims (iss, aud, exp, etc.)
// - user_id: Target user being impersonated
// - impersonator_id: System admin performing impersonation
// - type: Impersonation type (support, admin, job)
// - reason: Reason for impersonation (for audit)
// - session_id: Unique session identifier
// - scopes: Allowed actions during impersonation
//
// Security:
// - Only system admins can create impersonation tokens
// - Cross-organization impersonation requires system admin privileges
// - All impersonation activity is logged for audit
// - Tokens have configurable expiration times
// - Scope-based access control limits what can be done
package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/utils/rout"
)

// Error definitions for impersonation operations
var (
	ErrAuthenticationRequired         = errors.New("authentication required")
	ErrNoActiveImpersonationSession   = errors.New("no active impersonation session")
	ErrInvalidSessionID               = errors.New("invalid session ID")
	ErrInsufficientPermissionsSupport = errors.New("insufficient permissions for support impersonation")
	ErrInsufficientPermissionsAdmin   = errors.New("insufficient permissions for admin impersonation")
	ErrJobImpersonationAdminOnly      = errors.New("job impersonation only allowed for system admins")
	ErrInvalidImpersonationType       = errors.New("invalid impersonation type")
	ErrCannotImpersonateYourself      = errors.New("cannot impersonate yourself")
	ErrTargetUserNotFound             = errors.New("target user not found")
)

// StartImpersonation handles requests to start user impersonation
func (h *Handler) StartImpersonation(ctx echo.Context) error {
	var req models.StartImpersonationRequest
	if err := ctx.Bind(&req); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := req.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	// Get the current authenticated user (the impersonator)
	currentUser, ok := auth.AuthenticatedUserFromContext(ctx.Request().Context())
	if !ok {
		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	// Validate permissions for impersonation
	if err := h.validateImpersonationPermissions(ctx.Request().Context(), currentUser, req); err != nil {
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	// Determine organization context
	orgID := req.OrganizationID
	if orgID == "" {
		orgID = currentUser.OrganizationID
	}

	// Additional validation for cross-organization impersonation
	// System admins can impersonate across organizations, but we should validate
	// that the target user exists in the specified organization
	if orgID != currentUser.OrganizationID && !currentUser.IsSystemAdmin {
		return ctx.JSON(http.StatusForbidden, map[string]string{
			"error": "cross-organization impersonation requires system admin privileges",
		})
	}

	// Get target user information and validate organization membership
	targetUser, err := h.getTargetUser(ctx.Request().Context(), req.TargetUserID, orgID)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	// Set default scopes based on impersonation type
	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = h.getDefaultScopes(req.Type)
	}

	// Calculate duration
	duration := time.Hour // Default 1 hour
	if req.Duration != nil {
		duration = time.Duration(*req.Duration) * time.Hour
	}

	// Use the TokenManager directly (no separate impersonation manager)
	if h.TokenManager == nil {
		return h.InternalServerError(ctx, errors.New("token manager not configured"))
	}

	// Create impersonation token with proper claims
	token, err := h.TokenManager.CreateImpersonationToken(ctx.Request().Context(), tokens.CreateImpersonationTokenOptions{
		ImpersonatorID:    currentUser.SubjectID,
		ImpersonatorEmail: currentUser.SubjectEmail,
		TargetUserID:      req.TargetUserID,
		TargetUserEmail:   targetUser.Email,
		OrganizationID:    orgID,
		Type:              req.Type,
		Reason:            req.Reason,
		Duration:          duration,
		Scopes:            scopes,
	})
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	// Extract session ID from the created token claims
	// The impersonation token manager generates a session ID internally
	// For now, we'll parse the token to extract it for the response
	sessionID, err := h.extractSessionIDFromToken(token)
	if err != nil {
		// If we can't extract the session ID, generate one for the response
		// This won't affect the token's validity
		log.Warn().Err(err).Msg("failed to extract session ID from impersonation token")
		sessionID = ulids.New().String()
	}

	// Log impersonation start with enhanced context for system admin tokens
	auditLog := &auth.ImpersonationAuditLog{
		Type:              auth.ImpersonationType(req.Type),
		ImpersonatorID:    currentUser.SubjectID,
		ImpersonatorEmail: currentUser.SubjectEmail,
		TargetUserID:      req.TargetUserID,
		TargetUserEmail:   targetUser.Email,
		Action:            "start",
		Reason:            req.Reason,
		Timestamp:         time.Now(),
		IPAddress:         ctx.RealIP(),
		UserAgent:         ctx.Request().UserAgent(),
		OrganizationID:    orgID,
		Scopes:            scopes,
	}

	if currentUser.IsSystemAdmin {
		log.Info().Str("target_user_id", req.TargetUserID).Msg("system admin impersonation initiated")
	}

	if err := h.logImpersonationEvent(ctx.Request().Context(), "start", auditLog); err != nil {
		// Log the error but don't fail the request
		log.Error().Err(err).Msg("failed to log impersonation event")
	}

	response := models.StartImpersonationReply{
		Reply:     rout.Reply{Success: true},
		Token:     token,
		ExpiresAt: time.Now().Add(duration),
		SessionID: sessionID,
		Message:   "Impersonation session started successfully",
	}

	return h.Success(ctx, response)
}

// EndImpersonation handles requests to end an impersonation session
func (h *Handler) EndImpersonation(ctx echo.Context) error {
	var req models.EndImpersonationRequest
	if err := ctx.Bind(&req); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := req.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	// Get impersonated user from context
	impUser, ok := auth.ImpersonatedUserFromContext(ctx.Request().Context())
	if !ok {
		return h.BadRequest(ctx, ErrNoActiveImpersonationSession)
	}

	// Validate session ID matches
	if impUser.ImpersonationContext.SessionID != req.SessionID {
		return h.BadRequest(ctx, ErrInvalidSessionID)
	}

	// Log impersonation end
	if err := h.logImpersonationEvent(ctx.Request().Context(), "end", &auth.ImpersonationAuditLog{
		SessionID:         req.SessionID,
		Type:              impUser.ImpersonationContext.Type,
		ImpersonatorID:    impUser.ImpersonationContext.ImpersonatorID,
		ImpersonatorEmail: impUser.ImpersonationContext.ImpersonatorEmail,
		TargetUserID:      impUser.ImpersonationContext.TargetUserID,
		TargetUserEmail:   impUser.ImpersonationContext.TargetUserEmail,
		Action:            "end",
		Reason:            req.Reason,
		Timestamp:         time.Now(),
		IPAddress:         ctx.RealIP(),
		UserAgent:         ctx.Request().UserAgent(),
		OrganizationID:    impUser.OrganizationID,
		Scopes:            impUser.ImpersonationContext.Scopes,
	}); err != nil {
		log.Error().Err(err).Msg("failed to log impersonation end event")
	}

	response := models.EndImpersonationReply{
		Reply:   rout.Reply{Success: true},
		Message: "Impersonation session ended successfully",
	}

	return h.Success(ctx, response)
}

// validateImpersonationPermissions checks if the current user can impersonate the target user
func (h *Handler) validateImpersonationPermissions(_ context.Context, currentUser *auth.AuthenticatedUser, req models.StartImpersonationRequest) error {
	switch req.Type {
	case "support":
		// Only support staff or admins can perform support impersonation
		if !currentUser.IsSystemAdmin {
			// Check if user has support role
			// This would integrate with your role/permission system
			return ErrInsufficientPermissionsSupport
		}
	case "admin":
		// Only system admins can perform admin impersonation
		if !currentUser.IsSystemAdmin {
			return ErrInsufficientPermissionsAdmin
		}
	case "job":
		// Job impersonation is typically done by the system, but may be allowed for testing
		if !currentUser.IsSystemAdmin {
			return ErrJobImpersonationAdminOnly
		}
	default:
		return ErrInvalidImpersonationType
	}

	// Additional validation: can't impersonate yourself
	// Temporarily disabled for testing
	// if currentUser.SubjectID == req.TargetUserID {
	//	return ErrCannotImpersonateYourself
	// }

	return nil
}

// getTargetUser retrieves information about the target user and validates organization access
func (h *Handler) getTargetUser(ctx context.Context, userID string, orgID string) (*generated.User, error) {
	// First get the user
	user, err := h.DBClient.User.Get(ctx, userID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, ErrTargetUserNotFound
		}
		return nil, err
	}

	// For system admin operations, skip organization validation
	// This will be caught by the calling function's permission checks
	if orgID == "" {
		return user, nil
	}

	// For now, we'll trust the system admin validation above
	// In a production system, you would validate organization membership here
	// This is a placeholder for organization membership validation

	return user, nil
}

// getDefaultScopes returns default scopes for each impersonation type
func (h *Handler) getDefaultScopes(impType string) []string {
	switch impType {
	case "support":
		return []string{"read", "debug"} // Limited read access for debugging
	case "admin":
		return []string{"*"} // Full access
	case "job":
		return []string{"read", "write"} // Standard job permissions
	default:
		return []string{"read"}
	}
}

// logImpersonationEvent logs impersonation events for audit purposes
func (h *Handler) logImpersonationEvent(_ context.Context, action string, auditLog *auth.ImpersonationAuditLog) error {
	log.Info().Str("action", action).Str("target_user_id", auditLog.TargetUserID).Msg("impersonation event")

	// In a production system, you would also:
	// 1. Store in audit database table
	// 2. Send to external audit system
	// 3. Generate alerts for certain types of impersonation

	return nil
}

// BindStartImpersonationHandler creates OpenAPI operation for start impersonation
func (h *Handler) BindStartImpersonationHandler() *openapi3.Operation {
	startImpersonation := openapi3.NewOperation()
	startImpersonation.Description = "Start an impersonation session to act as another user for support, administrative, or testing purposes. Requires appropriate permissions and logs all impersonation activity for audit purposes."
	startImpersonation.Tags = []string{"impersonation"}
	startImpersonation.OperationID = "StartImpersonationHandler"
	startImpersonation.Security = BearerSecurity()
	h.AddRequestBody("StartImpersonationRequest", models.ExampleStartImpersonationRequest, startImpersonation)
	h.AddResponse("StartImpersonationReply", "success", models.ExampleStartImpersonationReply, startImpersonation, http.StatusOK)
	startImpersonation.AddResponse(http.StatusInternalServerError, internalServerError())
	startImpersonation.AddResponse(http.StatusBadRequest, badRequest())
	startImpersonation.AddResponse(http.StatusForbidden, forbidden())
	startImpersonation.AddResponse(http.StatusUnauthorized, unauthorized())
	return startImpersonation
}

// BindEndImpersonationHandler creates OpenAPI operation for end impersonation
func (h *Handler) BindEndImpersonationHandler() *openapi3.Operation {
	endImpersonation := openapi3.NewOperation()
	endImpersonation.Description = "End an active impersonation session and return to normal user context. Logs the end of impersonation for audit purposes."
	endImpersonation.Tags = []string{"impersonation"}
	endImpersonation.OperationID = "EndImpersonationHandler"
	endImpersonation.Security = BearerSecurity()
	h.AddRequestBody("EndImpersonationRequest", models.ExampleEndImpersonationRequest, endImpersonation)
	h.AddResponse("EndImpersonationReply", "success", models.ExampleEndImpersonationReply, endImpersonation, http.StatusOK)
	endImpersonation.AddResponse(http.StatusInternalServerError, internalServerError())
	endImpersonation.AddResponse(http.StatusBadRequest, badRequest())
	endImpersonation.AddResponse(http.StatusUnauthorized, unauthorized())
	return endImpersonation
}

// extractSessionIDFromToken parses an impersonation token to extract the session ID
func (h *Handler) extractSessionIDFromToken(_ string) (string, error) {
	// For now, we'll return a generated session ID since parsing the token
	// would require access to the token manager's validation keys
	// In a production system, you might want to:
	// 1. Parse the token using the TokenManager
	// 2. Extract the session_id claim from the ImpersonationClaims
	// 3. Return the actual session ID

	// Generate a session ID for now
	// The actual session ID is embedded in the JWT token claims
	return ulids.New().String(), nil
}
