package handlers

import (
	"net/url"
	"strings"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/pkg/domain"
	"github.com/theopenlane/core/pkg/logx"
)

func (h *Handler) CreateTrustCenterAnonymousJWT(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		response := models.CreateTrustCenterAnonymousJWTResponse{}
		return h.Success(ctx, response, openapi)
	}

	referer := ctx.Request().URL.Query().Get("referer")

	// 1. create the auth allowContext with a bootstrap trust center caller
	reqCtx := ctx.Request().Context()
	// Allow database queries for trust center lookup without authentication
	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	allowCtx = auth.WithCaller(allowCtx, auth.NewTrustCenterBootstrapCaller(""))

	// 2. parse the URL out of the `in`
	if referer == "" {
		return h.BadRequest(ctx, ErrMissingReferer, openapi)
	}

	parsedURL, err := url.Parse(referer)
	if err != nil {
		return h.BadRequest(ctx, ErrInvalidRefererURL, openapi)
	}

	hostname := parsedURL.Hostname()
	normalizedHost, err := domain.NormalizeHostname(hostname)
	if err != nil {
		return h.BadRequest(ctx, ErrInvalidRefererURL, openapi)
	}

	normalizedDefaultDomain, err := domain.NormalizeHostname(h.DefaultTrustCenterDomain)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("invalid default trust center domain")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	var trustCenter *generated.TrustCenter

	// 3. check if the URL is the "default trust center domain"
	if normalizedHost == normalizedDefaultDomain {
		// 4. if we have the default trust center domain, then we require the PATH of the url to be the "slug"
		pathSegments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(pathSegments) == 0 || pathSegments[0] == "" {
			return h.BadRequest(ctx, ErrMissingSlugInPath, openapi)
		}

		slug := pathSegments[0]

		// 4a. query the database for trust centers with the slug and the default hostname
		trustCenter, err = h.DBClient.TrustCenter.Query().
			Where(trustcenter.SlugEQ(slug)).
			Only(allowCtx)
		if err != nil {
			if generated.IsNotFound(err) {
				return h.Unauthorized(ctx, ErrTrustCenterNotFound)
			}

			logx.FromContext(reqCtx).Error().Err(err).Msg("error querying trust center")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}
	} else {
		// 5. if not default trust center, all we care about is the hostname.
		// 5a. query the database for trust centers with the hostname
		domainPredicate := customdomain.Or(
			customdomain.CnameRecordEqualFold(normalizedHost),
			customdomain.CnameRecordEqualFold(normalizedHost+"."),
		)
		trustCenter, err = h.DBClient.TrustCenter.Query().
			Where(trustcenter.Or(
				trustcenter.HasCustomDomainWith(domainPredicate),
				trustcenter.HasPreviewDomainWith(domainPredicate),
			)).
			Only(allowCtx)
		if err != nil {
			if generated.IsNotFound(err) {
				return h.Unauthorized(ctx, ErrTrustCenterNotFound)
			}

			logx.FromContext(reqCtx).Error().Err(err).Msg("error querying trust center by custom domain")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}
	}

	auth, err := h.AuthManager.GenerateAnonymousTrustCenterSession(reqCtx, ctx.Response().Writer, trustCenter.OwnerID, trustCenter.ID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	response := models.CreateTrustCenterAnonymousJWTResponse{
		AuthData: *auth,
	}

	return h.Success(ctx, response, openapi)
}
