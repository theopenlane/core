package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/httpsling"
)

func (suite *HandlerTestSuite) TestTFALogin() {
	t := suite.T()

	suite.registerTestHandler("POST", "login", suite.h.LoginHandler)
	suite.registerTestHandler("POST", "login/2fa/validate", suite.h.ValidateLoginTOTP)

	validPassword := "ssw444up3rs3cu7e!"

	user := suite.userBuilderWithInput(context.Background(), &userInput{
		password:      validPassword,
		confirmedUser: true,
		tfaEnabled:    true,
	})

	setting := suite.db.TFASetting.Create().
		SetTotpAllowed(true).
		SetOwnerID(user.ID).
		SetVerified(true).
		SaveX(user.UserCtx)

	require.NotNil(t, setting.TfaSecret)

	reqBody, err := json.Marshal(models.LoginRequest{
		Username: user.UserInfo.Email,
		Password: validPassword,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(string(reqBody)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	var out models.LoginResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))

	// no auth data should be generated but we want to make sure tfa is enforced
	assert.True(t, out.Success)
	assert.True(t, out.TFAEnabled)
	assert.Empty(t, out.AuthData)

	cookies := res.Cookies()

	code, err := totpGenerator(*setting.TfaSecret)
	require.NoError(t, err)
	reqBody, err = json.Marshal(models.TFARequest{TOTPCode: code})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/login/2fa/validate", strings.NewReader(string(reqBody)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	recorder = httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	res = recorder.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))
	assert.True(t, out.Success)
	assert.NotEmpty(t, out.Session)
}
