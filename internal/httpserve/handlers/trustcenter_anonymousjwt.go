package handlers

import (
	"fmt"
	"net/url"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/models"
)

// TODO: Add this to configuration
const defaultTrustCenterDomain = "trust.openlane.com"

func (h *Handler) CreateTrustCenterAnonymousJWT(ctx echo.Context) error {
	fmt.Println("IN HERE")
	referer := ctx.Request().URL.Query().Get("referer")
	fmt.Printf("%+v\n", referer)
	fmt.Println("got here")

	// 1. create the auth allowContext with the TrustCenterContext
	reqCtx := ctx.Request().Context()
	// Allow database queries for trust center lookup without authentication
	// allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

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

	// var trustCenter *generated.TrustCenter
	orgID := "01JX8AJMVHPYEREXQDWSF8P4N4"

	// // 3. check if the URL is the "default trust center domain"
	// if hostname == defaultTrustCenterDomain {
	// 	// 4. if we have the default trust center domain, then we require the PATH of the url to be the "slug"
	// 	slug := strings.Trim(parsedURL.Path, "/")
	// 	if slug == "" {
	// 		return h.BadRequest(ctx, ErrMissingSlugInPath)
	// 	}

	// 	// 4a. query the database for trust centers with the slug and the default hostname
	// 	trustCenter, err = h.DBClient.TrustCenter.Query().
	// 		Where(trustcenter.SlugEQ(slug)).
	// 		Where(trustcenter.Not(trustcenter.HasCustomDomain())).
	// 		First(allowCtx)
	// 	if err != nil {
	// 		if generated.IsNotFound(err) {
	// 			return h.Unauthorized(ctx, ErrTrustCenterNotFound)
	// 		}
	// 		return h.InternalServerError(ctx, err)
	// 	}
	// } else {
	// 	// 5. if not default trust center, all we care about is the hostname.
	// 	// 5a. query the database for trust centers with the hostname
	// 	trustCenter, err = h.DBClient.TrustCenter.Query().
	// 		Where(trustcenter.HasCustomDomainWith(
	// 			customdomain.CnameRecordEQ(hostname),
	// 		)).
	// 		First(allowCtx)
	// 	if err != nil {
	// 		if generated.IsNotFound(err) {
	// 			return h.Unauthorized(ctx, ErrTrustCenterNotFound)
	// 		}
	// 		return h.InternalServerError(ctx, err)
	// 	}
	// }

	auth, err := h.AuthManager.GenerateAnonymousAuthSession(reqCtx, ctx.Response().Writer, orgID)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	response := models.CreateTrustCenterAnonymousJWTResponse{
		AuthData: *auth,
	}

	return h.Success(ctx, response)
}
