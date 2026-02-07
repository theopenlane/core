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

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/httpsling"
)

func (suite *HandlerTestSuite) TestTFAValidate() {
	t := suite.T()

	// add login handler
	// Create operation for ValidateTOTP
	operation := suite.createImpersonationOperation("ValidateTOTP", "Validate TOTP code")
	suite.registerTestHandler("POST", "2fa/validate", operation, suite.h.ValidateTOTP)

	// Create a fresh TFA-enabled user per subtest to avoid time-window and reuse
	// collisions between tests.
	newTFAUser := func(t *testing.T) (context.Context, string, []string) {
		t.Helper()

		user := suite.userBuilderWithInput(context.Background(), &userInput{
			confirmedUser: true,
			tfaEnabled:    true,
		})

		tfaSetting := suite.db.TFASetting.Create().
			SetTotpAllowed(true).
			SetOwnerID(user.ID).
			SaveX(user.UserCtx)

		updateTFASetting, err := suite.db.TFASetting.UpdateOne(tfaSetting).
			SetVerified(true).
			Save(user.UserCtx)
		require.NoError(t, err)
		require.NotNil(t, updateTFASetting.TfaSecret)

		return user.UserCtx, *updateTFASetting.TfaSecret, updateTFASetting.RecoveryCodes
	}

	type testCase struct {
		name           string
		generateCode   bool
		code           string
		recoveryIndex  int
		expectedErr    string
		expectedStatus int
	}

	testCases := []testCase{
		{
			name:           "empty totp code",
			expectedStatus: http.StatusBadRequest,
			expectedErr:    "totp_code is required",
			recoveryIndex:  -1,
		},
		{
			name:           "happy path",
			generateCode:   true,
			expectedStatus: http.StatusOK,
			recoveryIndex:  -1,
		},
		{
			name:           "invalid totp code",
			code:           "123456",
			expectedStatus: http.StatusBadRequest,
			expectedErr:    "incorrect code provided",
			recoveryIndex:  -1,
		},
		{
			name:           "recovery code",
			recoveryIndex:  0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "recovery code, 2nd code",
			recoveryIndex:  7,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, secret, recoveryCodes := newTFAUser(t)

			code := tc.code
			var err error
			if tc.generateCode {
				code, err = totpGenerator(secret)
				require.NoError(t, err)
			}

			tfaJSON := models.TFARequest{
				TOTPCode: code,
			}

			if tc.recoveryIndex >= 0 {
				require.Less(t, tc.recoveryIndex, len(recoveryCodes))
				tfaJSON.RecoveryCode = recoveryCodes[tc.recoveryIndex]
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
			suite.e.ServeHTTP(recorder, req.WithContext(ctx))

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

	// Keep recovery code reuse in a single subtest to assert stateful behavior deterministically.
	t.Run("cannot reuse the same recovery code", func(t *testing.T) {
		ctx, _, recoveryCodes := newTFAUser(t)
		require.NotEmpty(t, recoveryCodes)
		recoveryCode := recoveryCodes[0]

		send := func() (*httptest.ResponseRecorder, *models.TFAReply) {
			tfaJSON := models.TFARequest{
				RecoveryCode: recoveryCode,
			}

			body, err := json.Marshal(tfaJSON)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/2fa/validate", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			recorder := httptest.NewRecorder()
			suite.e.ServeHTTP(recorder, req.WithContext(ctx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.TFAReply
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			return recorder, out
		}

		recorder, out := send()
		assert.Equal(t, http.StatusOK, recorder.Code)
		require.NotNil(t, out)
		assert.True(t, out.Success)

		recorder, out = send()
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		require.NotNil(t, out)
		require.NotNil(t, out.Error)
		assert.Contains(t, out.Error, "invalid code provided")
	})
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
