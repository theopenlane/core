package handlers_test

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/theopenlane/common/openapi"
	"github.com/theopenlane/httpsling"
)

func (suite *HandlerTestSuite) TestTFAValidate() {
	t := suite.T()

	// add login handler
	// Create operation for ValidateTOTP
	operation := suite.createImpersonationOperation("ValidateTOTP", "Validate TOTP code")
	suite.registerTestHandler("POST", "2fa/validate", operation, suite.h.ValidateTOTP)

	tfaSetting := suite.db.TFASetting.Create().
		SetTotpAllowed(true).
		SetOwnerID(testUser1.ID).
		SaveX(testUser1.UserCtx)

	updateTFASetting, err := suite.db.TFASetting.UpdateOne(tfaSetting).
		SetVerified(true).
		Save(testUser1.UserCtx)

	require.NoError(t, err)

	testCases := []struct {
		name           string
		generateCode   bool
		code           string
		recoveryCode   string
		ctx            context.Context
		expectedErr    string
		expectedStatus int
	}{
		{
			name:           "empty totp code",
			ctx:            testUser1.UserCtx,
			expectedStatus: http.StatusBadRequest,
			expectedErr:    "totp_code is required",
		},
		{
			name:           "happy path",
			ctx:            testUser1.UserCtx,
			generateCode:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid totp code",
			ctx:            testUser1.UserCtx,
			code:           "123456",
			expectedStatus: http.StatusBadRequest,
			expectedErr:    "incorrect code provided",
		},
		{
			name:           "recovery code",
			ctx:            testUser1.UserCtx,
			recoveryCode:   updateTFASetting.RecoveryCodes[0],
			expectedStatus: http.StatusOK,
		},
		{
			name:           "cannot reuse the same recovery code",
			ctx:            testUser1.UserCtx,
			recoveryCode:   updateTFASetting.RecoveryCodes[0],
			expectedStatus: http.StatusBadRequest,
			expectedErr:    "invalid code provided",
		},
		{
			name:           "recovery code, 2nd code",
			ctx:            testUser1.UserCtx,
			recoveryCode:   updateTFASetting.RecoveryCodes[7],
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code := tc.code

			if tc.generateCode {
				code, err = totpGenerator(*updateTFASetting.TfaSecret)
				require.NoError(t, err)
			}

			tfaJSON := models.TFARequest{
				TOTPCode: code,
			}

			if tc.recoveryCode != "" {
				tfaJSON.RecoveryCode = tc.recoveryCode
			}

			body, err := json.Marshal(tfaJSON)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/2fa/validate", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(tc.ctx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.TFAReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusOK {
				assert.True(t, out.Success)
			} else {
				require.NotNil(t, out.Error)
				assert.Contains(t, out.Error, tc.expectedErr)
			}
		})
	}
}

// totpGenerator generates a TOTP code using the secret
// this is used for testing the TFA validation handler
func totpGenerator(secret string) (string, error) {
	decryptedSecret, err := decrypt(secret)
	if err != nil {
		return "", err
	}

	code, err := totp.GenerateCode(decryptedSecret, time.Now())
	if err != nil {
		return "", err
	}

	return code, nil
}

// decrypt decrypts an encrypted string using a versioned secret
// this is based off the original decrypt function in totp.go
// but has been modified to work with the test cases
func decrypt(encryptedTxt string) (string, error) {
	// Split and parse the version prefix
	v := strings.Split(encryptedTxt, ":")[0]
	encryptedTxt = strings.TrimPrefix(encryptedTxt, fmt.Sprintf("%s:", v))

	secret := otpManagerSecret

	key := sha256.Sum256([]byte(secret.Key))

	decoded, err := base64.StdEncoding.DecodeString(encryptedTxt)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(decoded) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short") //nolint:err113
	}

	nonce := decoded[:gcm.NonceSize()]
	cipherText := decoded[gcm.NonceSize():]

	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
