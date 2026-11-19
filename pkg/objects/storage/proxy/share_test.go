package proxy

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/golang-jwt/jwt/v5"

	"github.com/theopenlane/iam/tokens"
)

const (
	testIssuer        = "http://localhost:17608"
	testShareID       = "01HZZSHARE000000000000000"
	testRecipient     = "bob@example.com"
	testExpiredOffset = -2 * time.Hour
)

func TestNormalizeRecipientEmail(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "bob@example.com", NormalizeRecipientEmail("  Bob@Example.COM \n"))
	assert.Equal(t, "", NormalizeRecipientEmail("   "))
}

func TestRecipientHash(t *testing.T) {
	t.Parallel()

	// the hash is taken over the normalized address so differently cased inputs agree
	assert.Equal(t, RecipientHash("bob@example.com"), RecipientHash(" BOB@example.com "))
	assert.Assert(t, is.Len(RecipientHash("bob@example.com"), 64))
	assert.Assert(t, RecipientHash("bob@example.com") != RecipientHash("alice@example.com"))
}

func TestComposeShareURL(t *testing.T) {
	t.Parallel()

	url := ComposeShareURL("https://api.example.com/v1/files", "01HZZSHARE000000000000000", "a.b/c+d")
	assert.Equal(t, "https://api.example.com/v1/files/share/01HZZSHARE000000000000000?token=a.b%2Fc%2Bd", url)
}

func TestShareTokenRoundTrip(t *testing.T) {
	t.Parallel()

	token := ComposeShareToken("01HZZORG00000000000000000", "c2lnbmF0dXJl")

	orgID, signature, ok := SplitShareToken(token)
	assert.Assert(t, ok)
	assert.Equal(t, "01HZZORG00000000000000000", orgID)
	assert.Equal(t, "c2lnbmF0dXJl", signature)

	_, _, ok = SplitShareToken("c2lnbmF0dXJl")
	assert.Assert(t, !ok)
}

func TestNewShareVerifiedToken(t *testing.T) {
	t.Parallel()

	tm := newTestTokenManager(t)
	before := time.Now()

	tokenString, err := NewShareVerifiedToken(tm, testIssuer, testShareID, RecipientHash(testRecipient))
	assert.NilError(t, err)

	claims := jwt.MapClaims{}

	_, _, err = jwt.NewParser().ParseUnverified(tokenString, claims)
	assert.NilError(t, err)

	assert.Equal(t, ShareVerifiedTokenType, claims[ClaimType])
	assert.Equal(t, testShareID, claims[ClaimShareID])
	assert.Equal(t, RecipientHash(testRecipient), claims[ClaimRecipient])

	audience, err := claims.GetAudience()
	assert.NilError(t, err)
	assert.DeepEqual(t, jwt.ClaimStrings{DownloadAudience}, audience)

	issuer, err := claims.GetIssuer()
	assert.NilError(t, err)
	assert.Equal(t, testIssuer, issuer)

	expiresAt, err := claims.GetExpirationTime()
	assert.NilError(t, err)
	assert.Assert(t, expiresAt.Time.Sub(before.Add(ShareVerifiedDuration)).Abs() <= 5*time.Second)
}

func TestNewShareVerifiedToken_RequiresTokenManager(t *testing.T) {
	t.Parallel()

	_, err := NewShareVerifiedToken(nil, testIssuer, testShareID, RecipientHash(testRecipient))
	assert.ErrorIs(t, err, ErrTokenManagerRequired)
}

func TestVerifyShareVerifiedToken(t *testing.T) {
	t.Parallel()

	tm := newTestTokenManager(t)

	tokenString, err := NewShareVerifiedToken(tm, testIssuer, testShareID, RecipientHash(testRecipient))
	assert.NilError(t, err)

	shareID, recipient, err := VerifyShareVerifiedToken(tm, testIssuer, tokenString)
	assert.NilError(t, err)
	assert.Equal(t, testShareID, shareID)
	assert.Equal(t, RecipientHash(testRecipient), recipient)
}

func TestVerifyShareVerifiedToken_Rejections(t *testing.T) {
	t.Parallel()

	tm := newTestTokenManager(t)
	other := newTestTokenManager(t)

	// mint signs a token with the share-verified claims, letting each case override one of them
	mint := func(signer *tokens.TokenManager, overrides ...tokens.ConfigOpt) string {
		opts := append([]tokens.ConfigOpt{
			tokens.WithSignedTokenUse(),
			tokens.WithAudience(DownloadAudience),
			tokens.WithIssuer(testIssuer),
			tokens.WithAccessDuration(ShareVerifiedDuration),
			tokens.WithClaim(ClaimType, ShareVerifiedTokenType),
			tokens.WithClaim(ClaimShareID, testShareID),
			tokens.WithClaim(ClaimRecipient, RecipientHash(testRecipient)),
		}, overrides...)

		tokenString, err := signer.CreateSignedToken(opts...)
		assert.NilError(t, err)

		return tokenString
	}

	session, _, err := tm.CreateTokenPair(&tokens.Claims{UserID: "01HZZTESTUSER00000000000000", OrgID: "01HZZTESTORG000000000000000"})
	assert.NilError(t, err)

	testCases := []struct {
		name        string
		token       string
		expectedErr error
	}{
		{name: "signed by another key", token: mint(other), expectedErr: ErrShareVerifiedTokenInvalid},
		{name: "expired", token: mint(tm, tokens.WithAccessDuration(testExpiredOffset)), expectedErr: ErrShareVerifiedTokenInvalid},
		{name: "wrong audience", token: mint(tm, tokens.WithAudience("http://localhost:17608")), expectedErr: ErrShareVerifiedTokenInvalid},
		{name: "wrong issuer", token: mint(tm, tokens.WithIssuer("https://issuer.example.com")), expectedErr: ErrShareVerifiedTokenInvalid},
		{name: "wrong type", token: mint(tm, tokens.WithClaim(ClaimType, "share")), expectedErr: ErrShareVerifiedTokenType},
		{name: "session token", token: session, expectedErr: ErrShareVerifiedTokenInvalid},
		{name: "not a token", token: "not-a-token", expectedErr: ErrShareVerifiedTokenInvalid},
		{name: "empty", token: "", expectedErr: ErrShareVerifiedTokenInvalid},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := VerifyShareVerifiedToken(tm, testIssuer, tc.token)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}
