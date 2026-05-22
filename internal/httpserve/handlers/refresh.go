package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/logx"
)

// RefreshHandler allows users to refresh their access token using their refresh token
func (h *Handler) RefreshHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleRefreshRequest, models.ExampleRefreshSuccessResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Skip actual handler logic during OpenAPI registration
	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	// verify the refresh token
	claims, err := h.TokenManager.Verify(req.RefreshToken)
	if err != nil {
		diag := refreshTokenDiagnostics(req.RefreshToken, err)
		logx.FromContext(reqCtx).Error().Err(err).Str("token_error_category", diag.category).Str("token_kid", diag.kid).Str("token_alg", diag.alg).Str("token_parse_error", diag.parseError).Str("token_exp", diag.exp).Str("token_iat", diag.iat).Str("token_nbf", diag.nbf).Msg("error verifying refresh token")

		return h.BadRequest(ctx, ErrUnableToVerifyToken, openapi)
	}

	// check user in the database, sub == claims subject and ensure only one record is returned
	user, err := h.getUserDetailsByID(reqCtx, claims.Subject)
	if err != nil {
		if ent.IsNotFound(err) {
			logx.FromContext(reqCtx).Info().Str("userID", user.ID).Msg("user not found during token refresh")
			return h.NotFound(ctx, ErrProcessingRequest, openapi)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// ensure the user is still active
	if user.Edges.Setting.Status != "ACTIVE" {
		logx.FromContext(reqCtx).Info().Str("userID", user.ID).Msg("user not active during token refresh")

		return h.NotFound(ctx, ErrProcessingRequest, openapi)
	}

	// get modules on refresh
	modules, err := rule.GetFeaturesForSpecificOrganization(reqCtx, claims.OrgID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error obtaining org features for claims, skipping modules in JWT")
	}

	claims.Modules = modules

	// UserID is not on the refresh token, so we need to set it now
	claims.UserID = user.ID

	accessToken, refreshToken, err := h.TokenManager.CreateTokenPair(claims)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error creating token pair")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// set cookies on request with the access and refresh token
	auth.SetAuthCookies(ctx.Response().Writer, accessToken, refreshToken, *h.SessionConfig.CookieConfig)

	// set sessions in response
	if _, err = h.SessionConfig.CreateAndStoreSession(reqCtx, ctx.Response().Writer, user.ID); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error storing session")

		return err
	}

	out := &models.RefreshReply{
		Reply:   rout.Reply{Success: true},
		Message: "success",
		AuthData: models.AuthData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}

	return h.Success(ctx, out, openapi)
}

type refreshTokenLogDiagnostics struct {
	category   string
	kid        string
	alg        string
	parseError string
	exp        string
	iat        string
	nbf        string
}

// refreshTokenDiagnostics extracts safe refresh token metadata for verification failure logs
func refreshTokenDiagnostics(refreshToken string, verifyErr error) refreshTokenLogDiagnostics {
	diag := refreshTokenLogDiagnostics{category: refreshTokenErrorCategory(verifyErr)}

	token, _, err := new(jwt.Parser).ParseUnverified(refreshToken, jwt.MapClaims{})
	if err != nil {
		diag.parseError = err.Error()

		return diag
	}

	if kid, ok := token.Header["kid"].(string); ok {
		diag.kid = kid
	}

	if alg, ok := token.Header["alg"].(string); ok {
		diag.alg = alg
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return diag
	}

	diag.exp = jwtClaimTime(claims, "exp")
	diag.iat = jwtClaimTime(claims, "iat")
	diag.nbf = jwtClaimTime(claims, "nbf")

	return diag
}

// refreshTokenErrorCategory returns a stable category for refresh token verification failures
func refreshTokenErrorCategory(err error) string {
	switch {
	case tokenErrorIs(err, tokens.ErrUnknownSigningKey):
		return "unknown_signing_key"
	case tokenErrorIs(err, tokens.ErrTokenMissingKid):
		return "missing_kid"
	case tokenErrorIs(err, tokens.ErrTokenSignatureInvalid):
		return "signature_invalid"
	case tokenErrorIs(err, tokens.ErrTokenNotValidYet):
		return "not_valid_yet"
	case tokenErrorIs(err, tokens.ErrTokenUsedBeforeIssued):
		return "used_before_issued"
	case tokenErrorIs(err, tokens.ErrTokenExpired):
		return "expired"
	case tokenErrorIs(err, tokens.ErrTokenInvalidAudience):
		return "invalid_audience"
	case tokenErrorIs(err, tokens.ErrTokenInvalidIssuer):
		return "invalid_issuer"
	case tokenErrorIs(err, tokens.ErrTokenInvalidClaims):
		return "invalid_claims"
	case tokenErrorIs(err, tokens.ErrTokenMalformed):
		return "malformed"
	case tokenErrorIs(err, tokens.ErrTokenUnverifiable):
		return "unverifiable"
	default:
		return "verification_failed"
	}
}

// tokenErrorIs reports whether err matches target by wrapping or by token error text
func tokenErrorIs(err, target error) bool {
	if err == nil || target == nil {
		return false
	}

	return errors.Is(err, target) || strings.Contains(err.Error(), target.Error())
}

// jwtClaimTime returns an RFC3339 timestamp for a numeric JWT time claim
func jwtClaimTime(claims jwt.MapClaims, name string) string {
	value, ok := claims[name]
	if !ok {
		return ""
	}

	var unixSeconds float64
	switch v := value.(type) {
	case float64:
		unixSeconds = v
	case json.Number:
		parsed, err := v.Float64()
		if err != nil {
			return fmt.Sprintf("%v", value)
		}

		unixSeconds = parsed
	default:
		return fmt.Sprintf("%v", value)
	}

	return time.Unix(int64(unixSeconds), 0).UTC().Format(time.RFC3339)
}
