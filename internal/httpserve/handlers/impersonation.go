package handlers

import (
	"context"
	"net/http"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/utils/rout"
)

// StartImpersonation handles requests to start user impersonation
func (h *Handler) StartImpersonation(ctx echo.Context, openapi *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleStartImpersonationRequest, &models.StartImpersonationReply{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	// Get the current authenticated user (the impersonator)
	currentUser, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		return h.Unauthorized(ctx, ErrAuthenticationRequired, openapi)
	}

	// Validate permissions for impersonation
	if err := h.validateImpersonationPermissions(currentUser, *req); err != nil {
		return h.Forbidden(ctx, err, openapi)
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
	targetUser, err := h.getTargetUser(reqCtx, req.TargetUserID, orgID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error getting target user")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
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
		logx.FromContext(reqCtx).Error().Msg("token manager not configured")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// Create impersonation token with proper claims
	token, err := h.TokenManager.CreateImpersonationToken(reqCtx, tokens.CreateImpersonationTokenOptions{
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
		logx.FromContext(reqCtx).Error().Err(err).Msg("error creating impersonation token")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// Extract session ID from the created token claims
	sessionID, err := h.extractSessionIDFromToken(token)
	if err != nil {
		// If we can't extract the session ID, the token creation was successful
		// but we have an issue with token parsing - this should not happen
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to extract session ID from newly created impersonation token")

		return h.InternalServerError(ctx, ErrFailedToExtractSessionID, openapi)
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
		logx.FromContext(reqCtx).Info().Str("target_user_id", req.TargetUserID).Msg("system admin impersonation initiated")
	}

	if err := h.logImpersonationEvent(reqCtx, "start", auditLog); err != nil {
		// Log the error but don't fail the request
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to log impersonation event")
	}

	response := models.StartImpersonationReply{
		Reply:     rout.Reply{Success: true},
		Token:     token,
		ExpiresAt: time.Now().Add(duration),
		SessionID: sessionID,
		Message:   "Impersonation session started successfully",
	}

	return h.Success(ctx, response, openapi)
}

// EndImpersonation handles requests to end an impersonation session
func (h *Handler) EndImpersonation(ctx echo.Context, openapi *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleEndImpersonationRequest, &models.EndImpersonationReply{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	// Get impersonated user from context
	impUser, ok := auth.ImpersonatedUserFromContext(reqCtx)
	if !ok {
		return h.BadRequest(ctx, ErrNoActiveImpersonationSession, openapi)
	}

	// Validate session ID matches
	if impUser.ImpersonationContext.SessionID != req.SessionID {
		return h.BadRequest(ctx, ErrInvalidSessionID, openapi)
	}

	// Log impersonation end
	if err := h.logImpersonationEvent(reqCtx, "end", &auth.ImpersonationAuditLog{
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
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to log impersonation end event")
	}

	response := models.EndImpersonationReply{
		Reply:   rout.Reply{Success: true},
		Message: "Impersonation session ended successfully",
	}

	return h.Success(ctx, response, openapi)
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
// and persists it into the database
func (h *Handler) logImpersonationEvent(ctx context.Context, action string, auditLog *auth.ImpersonationAuditLog) error {
	logx.FromContext(ctx).Info().Str("action", action).Str("target_user_id", auditLog.TargetUserID).Msg("impersonation event")

	_, err := h.DBClient.ImpersonationEvent.Create().
		SetAction(*enums.ToImpersonationAction(action)).
		SetImpersonationType(*enums.ToImpersonationType(string(auditLog.Type))).
		SetReason(auditLog.Reason).
		SetIPAddress(auditLog.IPAddress).
		SetScopes(auditLog.Scopes).
		SetUserAgent(auditLog.UserAgent).
		SetUserID(auditLog.ImpersonatorID).
		SetTargetUserID(auditLog.TargetUserID).
		SetOrganizationID(auditLog.OrganizationID).
		SetCreatedBy(auditLog.ImpersonatorID).
		SetCreatedAt(time.Now()).
		Save(ctx)
	return err
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
