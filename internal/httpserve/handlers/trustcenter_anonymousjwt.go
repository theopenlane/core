package handlers

import (
	"net/url"
	"strings"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/pkg/models"
)

// TODO: Add this to configuration This will allow for folks to test out their
// trust center before they have custom domains, e.g.
// trust.openlane.com/catcafe
const defaultTrustCenterDomain = "trust.openlane.com"

func (h *Handler) CreateTrustCenterAnonymousJWT(ctx echo.Context) error {
	referer := ctx.Request().URL.Query().Get("referer")

	// 1. create the auth allowContext with the TrustCenterContext
	reqCtx := ctx.Request().Context()
	// Allow database queries for trust center lookup without authentication
	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.TrustCenterContextKey{})

	// 2. parse the URL out of the `in`
	if referer == "" {
		return h.BadRequest(ctx, ErrMissingReferer)
	}

	parsedURL, err := url.Parse(referer)
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
		pathSegments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(pathSegments) == 0 || pathSegments[0] == "" {
			return h.BadRequest(ctx, ErrMissingSlugInPath)
		}
		slug := pathSegments[0]

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

	auth, err := h.AuthManager.GenerateAnonymousTrustCenterSession(reqCtx, ctx.Response().Writer, trustCenter.OwnerID, trustCenter.ID)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	response := models.CreateTrustCenterAnonymousJWTResponse{
		AuthData: *auth,
	}

	return h.Success(ctx, response)
}
