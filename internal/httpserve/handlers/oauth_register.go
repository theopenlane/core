package handlers

import (
	"context"
	"errors"
	"strings"

	echo "github.com/theopenlane/echox"
	"golang.org/x/oauth2"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/providers/github"
	"github.com/theopenlane/iam/providers/google"

	"github.com/theopenlane/common/enums"
	models "github.com/theopenlane/common/openapi"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	entval "github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/logx"
)

// OauthRegister returns the TokenResponse for a verified authenticated external oauth user
func (h *Handler) OauthRegister(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleOauthTokenRequest, models.ExampleLoginSuccessResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	ctxWithToken := token.NewContextWithOauthTooToken(reqCtx, in.Email)

	// create oauth2 token from request input
	tok := &oauth2.Token{
		AccessToken: in.ClientToken,
	}

	// verify the token provided to ensure the user is valid
	if err := h.verifyClientToken(ctxWithToken, in.AuthProvider, tok, in.Email); err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// check if users exists and create if not, updates last seen of existing user
	user, err := h.CheckAndCreateUser(ctxWithToken, in.Name, in.Email, enums.AuthProvider(strings.ToUpper(in.AuthProvider)), in.Image)
	if err != nil {
		if errors.Is(err, entval.ErrEmailNotAllowed) {
			logx.FromContext(reqCtx).Error().Err(err).Str("email", in.Email).Msg("email not allowed during registration")

			return h.InvalidInput(ctx, err, openapi)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// set user to verified
	if !user.Edges.Setting.EmailConfirmed {
		if err := h.setEmailConfirmed(ctxWithToken, user); err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("unable to set email as verified")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}
	}

	// create claims for verified user
	auth, err := h.AuthManager.GenerateOauthAuthSession(reqCtx, ctx.Response().Writer, user, *in)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	out := models.LoginReply{
		Reply:      rout.Reply{Success: true},
		TFAEnabled: user.Edges.Setting.IsTfaEnabled,
		Message:    "success",
		AuthData:   *auth,
	}

	// Return the access token
	return h.Success(ctx, out, openapi)
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
