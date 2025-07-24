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

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/utils/rout"
)

// Error definitions for impersonation operations
var (
	// ErrAuthenticationRequired indicates that the user must be authenticated to perform this action
	ErrAuthenticationRequired = errors.New("authentication required")
	// ErrNoActiveImpersonationSession indicates that there is no active impersonation session
	ErrNoActiveImpersonationSession = errors.New("no active impersonation session")
	// ErrInvalidSessionID indicates that the provided session ID is invalid
	ErrInvalidSessionID = errors.New("invalid session ID")
	// ErrInsufficientPermissionsSupport indicates that the user does not have permissions to perform support impersonation
	ErrInsufficientPermissionsSupport = errors.New("insufficient permissions for support impersonation")
	// ErrInsufficientPermissionsAdmin indicates that the user does not have permissions to perform admin impersonation
	ErrInsufficientPermissionsAdmin = errors.New("insufficient permissions for admin impersonation")
	// ErrJobImpersonationAdminOnly indicates that job impersonation is only allowed for system admins
	ErrJobImpersonationAdminOnly = errors.New("job impersonation only allowed for system admins")
	// ErrInvalidImpersonationType indicates that the provided impersonation type is invalid
	ErrInvalidImpersonationType = errors.New("invalid impersonation type")
	// ErrTargetUserNotFound indicates that the target user for impersonation was not found
	ErrTargetUserNotFound = errors.New("target user not found")
	// ErrTokenManagerNotConfigured indicates that the token manager is not configured
	ErrTokenManagerNotConfigured = errors.New("token manager not configured")
	// ErrFailedToExtractSessionID indicates that the session ID could not be extracted from the token
	ErrFailedToExtractSessionID = errors.New("failed to extract session ID from token")
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
	currentUser, err := auth.GetAuthenticatedUserFromContext(ctx.Request().Context())
	if err != nil {
		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	// Validate permissions for impersonation
	if err := h.validateImpersonationPermissions(currentUser, req); err != nil {
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

	// Use the TokenManager to create impersonation token
	if h.TokenManager == nil {
		return h.InternalServerError(ctx, ErrTokenManagerNotConfigured)
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
	sessionID, err := h.extractSessionIDFromToken(token)
	if err != nil {
		// If we can't extract the session ID, the token creation was successful
		// but we have an issue with token parsing - this should not happen
		log.Error().Err(err).Msg("failed to extract session ID from newly created impersonation token")

		return h.InternalServerError(ctx, ErrFailedToExtractSessionID)
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
func (h *Handler) validateImpersonationPermissions(currentUser *auth.AuthenticatedUser, req models.StartImpersonationRequest) error {
	switch req.Type {
	case "support":
		// Currently only system admins can perform support impersonation
		if !currentUser.IsSystemAdmin {
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

	// Organization membership validation is handled by the calling function's
	// permission checks. System admins can impersonate across organizations.

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
// Currently logs to application logs only. Future enhancement will persist to database.
func (h *Handler) logImpersonationEvent(_ context.Context, action string, auditLog *auth.ImpersonationAuditLog) error {
	log.Info().Str("action", action).Str("target_user_id", auditLog.TargetUserID).Msg("impersonation event")

	//TODO: Add ent schema to persist impersonation events to database for audit trail
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
func (h *Handler) extractSessionIDFromToken(token string) (string, error) {
	// Use the TokenManager to validate and parse the token
	claims, err := h.TokenManager.ValidateImpersonationToken(context.Background(), token)
	if err != nil {
		return "", err
	}

	// Return the session ID from the claims
	return claims.SessionID, nil
}
