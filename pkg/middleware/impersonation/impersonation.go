package impersonation

import (
	"net/http"
	"slices"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
)

// Middleware handles detection and processing of impersonation tokens
type Middleware struct {
	tokenManager *tokens.TokenManager
}

// New creates a new impersonation middleware
func New(tokenManager *tokens.TokenManager) *Middleware {
	return &Middleware{
		tokenManager: tokenManager,
	}
}

// Process is the middleware function that processes impersonation tokens
func (m *Middleware) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		// Check for impersonation token in Authorization header
		tokenString, err := auth.GetImpersonationToken(c)
		if err != nil {
			return next(c)
		}

		// Validate the impersonation token
		claims, err := m.tokenManager.ValidateImpersonationToken(ctx, tokenString)
		if err != nil {
			log.Warn().Err(err).Str("ip", c.RealIP()).Msg("invalid impersonation token")

			return echo.NewHTTPError(http.StatusUnauthorized, "invalid impersonation token")
		}

		// Create the impersonated caller context
		impersonatedCaller, err := m.createImpersonatedCaller(claims)
		if err != nil {
			log.Error().Err(err).Msg("failed to create impersonated user context")

			return echo.NewHTTPError(http.StatusInternalServerError, "failed to process impersonation")
		}

		// Set the impersonated caller in the context.
		ctx = auth.WithCaller(ctx, impersonatedCaller)
		c.SetRequest(c.Request().WithContext(ctx))

		// Log the impersonation action
		m.logImpersonationAccess(claims, c)

		return next(c)
	}
}

// createImpersonatedCaller creates a Caller with impersonation context from impersonation claims.
func (m *Middleware) createImpersonatedCaller(claims *tokens.ImpersonationClaims) (*auth.Caller, error) {
	caller := &auth.Caller{
		SubjectID:          claims.UserID,
		SubjectEmail:       claims.TargetUserEmail,
		OrganizationID:     claims.OrgID,
		AuthenticationType: auth.JWTAuthentication,
	}

	// Create the impersonation context
	impersonationContext := &auth.ImpersonationContext{
		Type:              auth.ImpersonationType(claims.Type),
		ImpersonatorID:    claims.ImpersonatorID,
		ImpersonatorEmail: claims.ImpersonatorEmail,
		TargetUserID:      claims.UserID,
		TargetUserEmail:   claims.TargetUserEmail,
		Reason:            claims.Reason,
		StartedAt:         claims.IssuedAt.Time,
		ExpiresAt:         claims.ExpiresAt.Time,
		SessionID:         claims.SessionID,
		Scopes:            claims.Scopes,
	}

	caller.Impersonation = impersonationContext

	return caller, nil
}

// logImpersonationAccess logs when an impersonation token is used
func (m *Middleware) logImpersonationAccess(claims *tokens.ImpersonationClaims, _ echo.Context) {
	log.Info().Str("impersonator", claims.ImpersonatorID).Str("target", claims.UserID).Msg("impersonation token used")
}

// RequireImpersonationScope creates middleware that requires specific impersonation scopes
func RequireImpersonationScope(requiredScope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			caller, ok := auth.CallerFromContext(ctx)
			if !ok || caller == nil || !caller.IsImpersonated() {
				// Not impersonated, proceed normally
				return next(c)
			}

			// Check if the impersonation has the required scope
			if !caller.CanPerformAction(requiredScope) {
				return echo.NewHTTPError(http.StatusForbidden, "impersonation scope insufficient for this action")
			}

			return next(c)
		}
	}
}

// BlockImpersonation creates middleware that blocks impersonated users from certain endpoints
func BlockImpersonation() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			caller, ok := auth.CallerFromContext(ctx)
			if ok && caller != nil && caller.IsImpersonated() {
				return echo.NewHTTPError(http.StatusForbidden, "action not allowed during impersonation session")
			}

			return next(c)
		}
	}
}

// AllowOnlyImpersonationType creates middleware that only allows specific impersonation types
func AllowOnlyImpersonationType(allowedTypes ...auth.ImpersonationType) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			caller, ok := auth.CallerFromContext(ctx)
			if !ok || caller == nil || !caller.IsImpersonated() {
				// Not impersonated, proceed normally
				return next(c)
			}

			// Check if the impersonation type is allowed
			if slices.Contains(allowedTypes, caller.Impersonation.Type) {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusForbidden, "impersonation type not allowed for this endpoint")
		}
	}
}

// SystemAdminUserContextMiddleware handles system admin tokens with user context headers
// It detects when a system admin token is used with X-User-ID and X-Organization-ID headers
// and sets up the user context to run GraphQL as that specific user
func SystemAdminUserContextMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// Check if there's already an authenticated caller (from previous middleware)
			caller, hasUser := auth.CallerFromContext(ctx)
			if !hasUser || caller == nil {
				// No authenticated caller, continue normally
				return next(c)
			}

			// Only proceed if the current user is a system admin
			if !caller.Has(auth.CapSystemAdmin) {
				return next(c)
			}

			// Check for user context headers
			if !auth.HasUserContextHeaders(c) {
				return next(c)
			}

			targetUserID, targetOrgID := auth.GetUserContextHeaders(c)

			// Replace the caller in the context with the target user
			ctx = auth.WithCaller(ctx, &auth.Caller{
				SubjectID:          targetUserID,
				OrganizationID:     targetOrgID,
				OrganizationIDs:    []string{targetOrgID},
				AuthenticationType: auth.PATAuthentication,
			})
			// Preserve the original admin caller on the active caller lineage.
			ctx = auth.WithOriginalSystemAdminCaller(ctx, caller)
			c.SetRequest(c.Request().WithContext(ctx))

			// Log the system admin user context switch
			log.Info().Str("admin", caller.SubjectID).Str("user", targetUserID).Msg("system admin user context")

			return next(c)
		}
	}
}
