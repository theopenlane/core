package handlers_test

import (
	"net/http/httptest"

	"github.com/stretchr/testify/require"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/sessions"

	"github.com/theopenlane/shared/enums"
)

func (suite *HandlerTestSuite) TestHasValidSSOSession() {
	t := suite.T()
	ctx := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())

	user := suite.userBuilder(ctx.Request().Context())

	set := map[string]any{
		sessions.UserTypeKey: enums.AuthProviderOIDC.String(),
		sessions.UserIDKey:   user.ID,
	}

	c, err := suite.h.SessionConfig.SaveAndStoreSession(ctx.Request().Context(), ctx.Response().Writer, set, user.ID)
	require.NoError(t, err)

	token, err := sessions.SessionToken(c)
	require.NoError(t, err)

	ctx.Request().AddCookie(sessions.NewDevSessionCookie(token))

	ok := suite.h.HasValidSSOSession(ctx, user.ID)
	require.True(t, ok)
}

func (suite *HandlerTestSuite) TestHasValidSSOSessionInvalid() {
	t := suite.T()
	ctx := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())

	user := suite.userBuilder(ctx.Request().Context())

	// no cookie set
	ok := suite.h.HasValidSSOSession(ctx, user.ID)
	require.False(t, ok)
}
