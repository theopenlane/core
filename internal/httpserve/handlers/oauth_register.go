package handlers

import (
	"context"
	"strings"

	ph "github.com/posthog/posthog-go"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"
	"golang.org/x/oauth2"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/providers/github"
	"github.com/theopenlane/iam/providers/google"

	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// OauthRegister returns the TokenResponse for a verified authenticated external oauth user
func (h *Handler) OauthRegister(ctx echo.Context) error {
	var in models.OauthTokenRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	ctxWithToken := token.NewContextWithOauthTooToken(ctx.Request().Context(), in.Email)

	// create oauth2 token from request input
	tok := &oauth2.Token{
		AccessToken: in.ClientToken,
	}

	// verify the token provided to ensure the user is valid
	if err := h.verifyClientToken(ctxWithToken, in.AuthProvider, tok, in.Email); err != nil {
		return h.InvalidInput(ctx, err)
	}

	// check if users exists and create if not, updates last seen of existing user
	user, err := h.CheckAndCreateUser(ctxWithToken, in.Name, in.Email, enums.AuthProvider(strings.ToUpper(in.AuthProvider)), in.Image)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	if err := h.addDefaultOrgToUserQuery(ctxWithToken, user); err != nil {
		return h.InternalServerError(ctx, err)
	}

	// create claims for verified user
	auth, err := h.AuthManager.GenerateOauthAuthSession(ctx.Request().Context(), ctx.Response().Writer, user, in)
	if err != nil {
		log.Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, err)
	}

	props := ph.NewProperties().
		Set("user_id", user.ID).
		Set("email", user.Email).
		Set("organization_id", user.Edges.Setting.Edges.DefaultOrg.ID). // user is logged into their default org
		Set("auth_provider", in.AuthProvider)

	h.AnalyticsClient.Event("user_authenticated", props)

	out := models.LoginReply{
		Reply:    rout.Reply{Success: true},
		Message:  "success",
		AuthData: *auth,
	}

	// Return the access token
	return h.Success(ctx, out)
}

// verifyClientToken verifies the provided access token from an external oauth2 provider is valid and matches the user's email
// supported providers are Github and Google
func (h *Handler) verifyClientToken(ctx context.Context, provider string, token *oauth2.Token, email string) error {
	switch strings.ToLower(provider) {
	case githubProvider:
		config := h.getGithubOauth2Config()
		cc := github.ClientConfig{IsEnterprise: false, IsMock: h.IsTest}

		return github.VerifyClientToken(ctx, token, config, email, &cc)
	case googleProvider:
		config := h.getGoogleOauth2Config()
		return google.VerifyClientToken(ctx, token, config, email)
	default:
		return ErrInvalidProvider
	}
}
