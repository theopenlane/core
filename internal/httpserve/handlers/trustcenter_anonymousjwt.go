package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golang-jwt/jwt/v5"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/pkg/models"
)

// TODO: Add this to configuration
const defaultTrustCenterDomain = "trust.openlane.com"

func (h *Handler) CreateTrustCenterAnonymousJWT(ctx echo.Context) error {
	var in models.CreateTrustCenterAnonymousJWTRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	// 1. create the auth allowContext with the TrustCenterContext
	reqCtx := ctx.Request().Context()
	// Allow database queries for trust center lookup without authentication
	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	// 2. parse the URL out of the `in`
	if in.Referer == "" {
		return h.BadRequest(ctx, ErrMissingReferer)
	}

	parsedURL, err := url.Parse(in.Referer)
	if err != nil {
		return h.BadRequest(ctx, ErrInvalidRefererURL)
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return h.BadRequest(ctx, ErrInvalidRefererURL)
	}

	var trustCenter *generated.TrustCenter

	// 3. check if the URL is the "default trust center domain"
	if hostname == defaultTrustCenterDomain {
		// 4. if we have the default trust center domain, then we require the PATH of the url to be the "slug"
		slug := strings.Trim(parsedURL.Path, "/")
		if slug == "" {
			return h.BadRequest(ctx, ErrMissingSlugInPath)
		}

		// 4a. query the database for trust centers with the slug and the default hostname
		trustCenter, err = h.DBClient.TrustCenter.Query().
			Where(trustcenter.SlugEQ(slug)).
			Where(trustcenter.Not(trustcenter.HasCustomDomain())).
			First(allowCtx)
		if err != nil {
			if generated.IsNotFound(err) {
				return h.Unauthorized(ctx, ErrTrustCenterNotFound)
			}
			return h.InternalServerError(ctx, err)
		}
	} else {
		// 5. if not default trust center, all we care about is the hostname.
		// 5a. query the database for trust centers with the hostname
		trustCenter, err = h.DBClient.TrustCenter.Query().
			Where(trustcenter.HasCustomDomainWith(
				customdomain.CnameRecordEQ(hostname),
			)).
			First(allowCtx)
		if err != nil {
			if generated.IsNotFound(err) {
				return h.Unauthorized(ctx, ErrTrustCenterNotFound)
			}
			return h.InternalServerError(ctx, err)
		}
	}

	// 6. if we find a trust center, then we create a JWT for it, with the trust center ID in the claims
	claims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: trustCenter.ID,
		},
		// For anonymous trust center access, we don't have a user or org
		// The trust center ID is stored in the Subject field
	}

	// Create access token for the trust center
	accessToken, err := h.TokenManager.CreateAccessToken(claims)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	signedToken, err := h.TokenManager.Sign(accessToken)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	response := models.CreateTrustCenterAnonymousJWTResponse{
		AccessToken: signedToken,
		TokenType:   "Bearer",
	}

	return h.Success(ctx, response)
}

// BindCreateTrustCenterAnonymousJWT binds the trust center anonymous JWT creation endpoint to the OpenAPI schema
func (h *Handler) BindCreateTrustCenterAnonymousJWT() *openapi3.Operation {
	createJWT := openapi3.NewOperation()
	createJWT.Description = "Create an anonymous JWT token for a trust center based on the referer URL. For default trust center domains, the slug is extracted from the path. For custom domains, the hostname is used to identify the trust center."
	createJWT.Tags = []string{"trustcenter", "authentication"}
	createJWT.OperationID = "CreateTrustCenterAnonymousJWT"
	createJWT.Security = &openapi3.SecurityRequirements{} // No authentication required

	// Add query parameter for referer
	createJWT.Parameters = openapi3.Parameters{
		&openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				Name:        "referer",
				In:          "query",
				Required:    true,
				Description: "The referer URL to identify the trust center",
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    &openapi3.Types{"string"},
						Example: "https://trust.openlane.com/my-trust-center",
					},
				},
			},
		},
	}

	// Add response schemas
	h.AddResponse("CreateTrustCenterAnonymousJWTResponse", "success", models.CreateTrustCenterAnonymousJWTResponse{
		AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		TokenType:   "Bearer",
	}, createJWT, http.StatusOK)

	createJWT.AddResponse(http.StatusBadRequest, badRequest())
	createJWT.AddResponse(http.StatusUnauthorized, unauthorized())
	createJWT.AddResponse(http.StatusInternalServerError, internalServerError())

	return createJWT
}
