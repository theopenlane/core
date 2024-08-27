package tokens_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/tokens"
	ulids "github.com/theopenlane/utils/ulids"
)

var rusty = "rusty.shackleford@gmail.com"

func TestVerificationToken(t *testing.T) {
	// Test that the verification token is created correctly
	token, err := tokens.NewVerificationToken(rusty)
	require.NoError(t, err, "could not create verification token")
	require.Equal(t, rusty, token.Email)
	require.True(t, token.ExpiresAt.After(time.Now()))
	require.Len(t, token.Nonce, 64)

	// Test signing a token
	signature, secret, err := token.Sign()
	require.NoError(t, err, "failed to sign token")
	require.NotEmpty(t, signature)
	require.Len(t, secret, 128)
	require.True(t, bytes.HasPrefix(secret, token.Nonce))

	// Signing again should produce a different signature
	differentSig, differentSecret, err := token.Sign()
	require.NoError(t, err, "failed to sign token")
	require.NotEqual(t, signature, differentSig, "expected different signatures")
	require.NotEqual(t, secret, differentSecret, "expected different secrets")

	// Verification should fail if the token is missing an email address
	verify := &tokens.VerificationToken{
		SigningInfo: tokens.SigningInfo{
			ExpiresAt: time.Now().AddDate(0, 0, 7),
		},
	}
	require.ErrorIs(t, verify.Verify(signature, secret), tokens.ErrTokenMissingEmail, "expected error when token is missing email address")

	// Verification should fail if the token is expired
	verify.Email = rusty
	verify.ExpiresAt = time.Now().AddDate(0, 0, -1)
	require.ErrorIs(t, verify.Verify(signature, secret), tokens.ErrTokenExpired, "expected error when token is expired")

	// Verification should fail if the email is different
	verify.Email = "sfunk@gmail.com"
	verify.ExpiresAt = token.ExpiresAt
	require.ErrorIs(t, verify.Verify(signature, secret), tokens.ErrTokenInvalid, "expected error when email is different")

	// Verification should fail if the signature is not decodable
	verify.Email = rusty
	require.Error(t, verify.Verify("^&**(", secret), "expected error when signature is not decodable")

	// Verification should fail if the signature was created with a different secret
	require.ErrorIs(t, verify.Verify(differentSig, secret), tokens.ErrTokenInvalid, "expected error when signature was created with a different secret")

	// Should error if the secret has the wrong length
	require.ErrorIs(t, verify.Verify(signature, nil), tokens.ErrInvalidSecret, "expected error when secret is nil")
	require.ErrorIs(t, verify.Verify(signature, []byte("wronglength")), tokens.ErrInvalidSecret, "expected error when secret is the wrong length")

	// Verification should fail if the wrong secret is used
	require.ErrorIs(t, verify.Verify(signature, differentSecret), tokens.ErrTokenInvalid, "expected error when wrong secret is used")

	// Successful verification
	require.NoError(t, verify.Verify(signature, secret), "expected successful verification")
}

func TestResetToken(t *testing.T) {
	t.Run("Valid Reset Token", func(t *testing.T) {
		// Test that the reset token is created correctly
		id := ulids.New()
		token, err := tokens.NewResetToken(id)
		require.NoError(t, err, "could not create reset token")

		// Test signing a token
		signature, secret, err := token.Sign()
		require.NoError(t, err, "failed to sign token")

		// Signing again should produce a different signature
		differentSig, differentSecret, err := token.Sign()
		require.NoError(t, err, "failed to sign token")
		require.NotEqual(t, signature, differentSig, "expected different signatures")
		require.NotEqual(t, secret, differentSecret, "expected different secrets")

		// Should be able to verify the token
		require.NoError(t, token.Verify(signature, secret), "expected successful verification")
	})

	t.Run("Missing ID", func(t *testing.T) {
		// Should fail to create a token without an ID
		_, err := tokens.NewResetToken(ulids.Null)
		require.ErrorIs(t, err, tokens.ErrMissingUserID, "expected error when token is missing ID")
	})

	t.Run("Token Missing User ID", func(t *testing.T) {
		// Token with missing user ID should be an error
		token := &tokens.ResetToken{}
		require.ErrorIs(t, token.Verify("", nil), tokens.ErrTokenMissingUserID, "expected error when token is missing ID")
	})

	t.Run("Token Expired", func(t *testing.T) {
		// Token that is expired should be an error
		token := &tokens.ResetToken{
			SigningInfo: tokens.SigningInfo{
				ExpiresAt: time.Now().AddDate(0, 0, -1),
			},
			UserID: ulids.New(),
		}
		require.ErrorIs(t, token.Verify("", nil), tokens.ErrTokenExpired, "expected error when token is expired")
	})

	t.Run("Wrong User ID", func(t *testing.T) {
		// Sign a valid token
		token, err := tokens.NewResetToken(ulids.New())
		require.NoError(t, err, "could not create reset token")
		signature, secret, err := token.Sign()
		require.NoError(t, err, "failed to sign token")

		// Verification should fail if the user ID is different
		token.UserID = ulids.New()
		require.ErrorIs(t, token.Verify(signature, secret), tokens.ErrTokenInvalid, "expected error when user ID is different")
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		// Sign a valid token
		token, err := tokens.NewResetToken(ulids.New())
		require.NoError(t, err, "could not create reset token")
		_, secret, err := token.Sign()
		require.NoError(t, err, "failed to sign token")

		// Verification should fail if the signature is not decodable
		require.Error(t, token.Verify("^&**(", secret), "expected error when signature is not decodable")

		// Verification should fail if the signature was created with a different secret
		otherToken, err := tokens.NewResetToken(token.UserID)
		require.NoError(t, err, "could not create reset token")
		otherSig, _, err := otherToken.Sign()
		require.NoError(t, err, "failed to sign token")
		require.ErrorIs(t, token.Verify(otherSig, secret), tokens.ErrTokenInvalid, "expected error when signature was created with a different secret")
	})

	t.Run("Invalid Secret", func(t *testing.T) {
		// Sign a valid token
		token, err := tokens.NewResetToken(ulids.New())
		require.NoError(t, err, "could not create reset token")
		signature, _, err := token.Sign()
		require.NoError(t, err, "failed to sign token")

		// Should error if the secret has the wrong length
		require.ErrorIs(t, token.Verify(signature, nil), tokens.ErrInvalidSecret, "expected error when secret is nil")
		require.ErrorIs(t, token.Verify(signature, []byte("wronglength")), tokens.ErrInvalidSecret, "expected error when secret is the wrong length")

		// Verification should fail if the wrong secret is used
		otherToken, err := tokens.NewResetToken(token.UserID)
		require.NoError(t, err, "could not create reset token")
		_, otherSecret, err := otherToken.Sign()
		require.NoError(t, err, "failed to sign token")
		require.ErrorIs(t, token.Verify(signature, otherSecret), tokens.ErrTokenInvalid, "expected error when wrong secret is used")
	})
}

func TestInviteToken(t *testing.T) {
	t.Run("Valid Reset Token", func(t *testing.T) {
		// Test that the reset token is created correctly
		orgID := ulids.New()
		token, err := tokens.NewOrgInvitationToken(rusty, orgID)
		require.NoError(t, err, "could not create reset token")

		// Test signing a token
		signature, secret, err := token.Sign()
		require.NoError(t, err, "failed to sign token")

		// Signing again should produce a different signature
		differentSig, differentSecret, err := token.Sign()
		require.NoError(t, err, "failed to sign token")
		require.NotEqual(t, signature, differentSig, "expected different signatures")
		require.NotEqual(t, secret, differentSecret, "expected different secrets")

		// Should be able to verify the token
		require.NoError(t, token.Verify(signature, secret), "expected successful verification")
	})

	t.Run("Missing ID", func(t *testing.T) {
		// Should fail to create a token without an ID
		_, err := tokens.NewOrgInvitationToken(rusty, ulids.Null)
		require.ErrorIs(t, err, tokens.ErrInviteTokenMissingOrgID, "invite token is missing org id")
	})

	t.Run("Missing Email", func(t *testing.T) {
		// Should fail to create a token without an ID
		_, err := tokens.NewOrgInvitationToken("", ulids.New())
		require.ErrorIs(t, err, tokens.ErrInviteTokenMissingEmail, "invite token is missing email")
	})

	t.Run("Token Expired", func(t *testing.T) {
		// Token that is expired should be an error
		token := &tokens.OrgInviteToken{
			SigningInfo: tokens.SigningInfo{
				ExpiresAt: time.Now().AddDate(0, 0, -1),
			},
			OrgID: ulids.New(),
			Email: rusty,
		}
		require.ErrorIs(t, token.Verify("", nil), tokens.ErrTokenExpired, "expected error when token is expired")
	})

	t.Run("Wrong Org ID", func(t *testing.T) {
		// Sign a valid token
		token, err := tokens.NewOrgInvitationToken(rusty, ulids.New())
		require.NoError(t, err, "could not create reset token")
		signature, secret, err := token.Sign()
		require.NoError(t, err, "failed to sign token")

		// Verification should fail if the user ID is different
		token.OrgID = ulids.New()
		require.ErrorIs(t, token.Verify(signature, secret), tokens.ErrTokenInvalid, "expected error when user ID is different")
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		// Sign a valid token
		token, err := tokens.NewOrgInvitationToken(rusty, ulids.New())
		require.NoError(t, err, "could not create reset token")
		_, secret, err := token.Sign()
		require.NoError(t, err, "failed to sign token")

		// Verification should fail if the signature is not decodable
		require.Error(t, token.Verify("^&**(", secret), "expected error when signature is not decodable")

		// Verification should fail if the signature was created with a different secret
		otherToken, err := tokens.NewOrgInvitationToken(rusty, token.OrgID)
		require.NoError(t, err, "could not create reset token")
		otherSig, _, err := otherToken.Sign()
		require.NoError(t, err, "failed to sign token")
		require.ErrorIs(t, token.Verify(otherSig, secret), tokens.ErrTokenInvalid, "expected error when signature was created with a different secret")
	})

	t.Run("Invalid Secret", func(t *testing.T) {
		// Sign a valid token
		token, err := tokens.NewOrgInvitationToken(rusty, ulids.New())
		require.NoError(t, err, "could not create reset token")
		signature, _, err := token.Sign()
		require.NoError(t, err, "failed to sign token")

		// Should error if the secret has the wrong length
		require.ErrorIs(t, token.Verify(signature, nil), tokens.ErrInvalidSecret, "expected error when secret is nil")
		require.ErrorIs(t, token.Verify(signature, []byte("wronglength")), tokens.ErrInvalidSecret, "expected error when secret is the wrong length")

		// Verification should fail if the wrong secret is used
		otherToken, err := tokens.NewOrgInvitationToken(rusty, token.OrgID)
		require.NoError(t, err, "could not create reset token")
		_, otherSecret, err := otherToken.Sign()
		require.NoError(t, err, "failed to sign token")
		require.ErrorIs(t, token.Verify(signature, otherSecret), tokens.ErrTokenInvalid, "expected error when wrong secret is used")
	})
}
