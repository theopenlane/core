package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/stretchr/testify/require"
	echo "github.com/theopenlane/echox"
	iamwebauthn "github.com/theopenlane/iam/providers/webauthn"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/iam/auth"
)

func (suite *HandlerTestSuite) TestBeginWebauthnRegistration() {
	t := suite.T()

	suite.h.WebAuthn = iamwebauthn.NewWithConfig(iamwebauthn.ProviderConfig{
		Enabled:        true,
		DisplayName:    "Openlane",
		RelyingPartyID: "localhost",
		RequestOrigins: []string{"http://localhost"},
		Timeout:        time.Minute,
	})
	t.Cleanup(func() { suite.h.WebAuthn = nil })

	suite.registerTestHandler("POST", "registration/options", suite.h.BeginWebauthnRegistration)

	email := "passkey-" + strings.ToLower(ulids.New().String()) + "@theopenlane.io"

	body, err := json.Marshal(models.WebauthnRegistrationRequest{Email: email, Name: "Passkey User"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/registration/options", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	var out models.WebauthnBeginRegistrationResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&out))
	require.NotNil(t, out.CredentialCreation)
	require.NotEmpty(t, out.Session)

	// add privacy allow to run the query to check that the user is created
	internalCtx := auth.WithInternalOperationContext(context.Background())

	exists, err := suite.db.User.Query().Where(user.Email(email)).Exist(internalCtx)
	require.NoError(t, err)
	require.True(t, exists)
}

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
