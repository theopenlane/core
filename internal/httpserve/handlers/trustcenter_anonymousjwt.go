package handlers

import (
	"net/url"
	"strings"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

func (h *Handler) CreateTrustCenterAnonymousJWT(ctx echo.Context) error {
	referer := ctx.Request().Referer()
	reqCtx := ctx.Request().Context()

	// validate the request
	if referer == "" {
		return h.BadRequest(ctx, ErrMissingReferer)
	}

	parsedURL, err := url.Parse(referer)
	if err != nil {
		return h.BadRequest(ctx, ErrInvalidRefererURL)
	}

	hostname := parsedURL.Hostname()
	normalizedHost, err := urlx.NormalizeHostname(hostname)
	if err != nil {
		return h.BadRequest(ctx, ErrInvalidRefererURL)
	}

	normalizedDefaultDomain, err := urlx.NormalizeHostname(h.DefaultTrustCenterDomain)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("invalid default trust center domain")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	// setup allow context to run queries on behalf of the anon user
	allowCtx := auth.WithCaller(reqCtx, auth.NewAnonBootstrapCallerOrgBypass())

	var trustCenter *generated.TrustCenter

	if normalizedHost == normalizedDefaultDomain {
		// if we have the default trust center domain, then we require the PATH of the url to be the "slug"
		pathSegments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(pathSegments) == 0 || pathSegments[0] == "" {
			return h.BadRequest(ctx, ErrMissingSlugInPath)
		}

		slug := pathSegments[0]

		//  query the database for trust centers with the slug and the default hostname
		trustCenter, err = h.DBClient.TrustCenter.Query().
			Where(trustcenter.SlugEQ(slug)).
			Only(allowCtx)
		if err != nil {
			if generated.IsNotFound(err) {
				return h.Unauthorized(ctx, ErrTrustCenterNotFound)
			}

			logx.FromContext(reqCtx).Error().Err(err).Msg("error querying trust center")

			return h.InternalServerError(ctx, ErrProcessingRequest)
		}
	} else {
		// if not default trust center, all we care about is the hostname.
		// query the database for trust centers with the hostname
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

			return h.InternalServerError(ctx, ErrProcessingRequest)
		}
	}

	auth, err := h.AuthManager.GenerateAnonymousTrustCenterSession(reqCtx, ctx.Response().Writer, trustCenter.OwnerID, trustCenter.ID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	response := models.CreateTrustCenterAnonymousJWTResponse{
		AuthData: *auth,
	}

	return h.Success(ctx, response)
}
