package proxy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/theopenlane/iam/tokens"
)

const (
	// DownloadAudience is the audience claim carried by every share-verified token; it must differ from the
	// session audience so a share-verified token is never accepted as a session and vice versa
	DownloadAudience = "openlane-files"
	// ShareVerifiedTokenType is the typ claim carried by the token proving a recipient verified their email for a share
	ShareVerifiedTokenType = "share-verified"
	// ShareVerifiedDuration is the lifetime of a share-verified token
	ShareVerifiedDuration = 15 * time.Minute
	// MaxShareDuration is the longest lifetime a share may be created with
	MaxShareDuration = 30 * 24 * time.Hour
	// MaxShareFiles is the most files a share link may grant
	MaxShareFiles = 20
	// UnlimitedShareUses is the max uses value that places no cap on downloads through a share link
	UnlimitedShareUses = 0
	// sharePathSegment is the path segment under the storage links base URL that serves shares
	sharePathSegment = "share"
	// shareTokenSeparator separates the organization id from the signature in a share token
	shareTokenSeparator = "."
	// ClaimType is the claim naming the token type
	ClaimType = "typ"
	// ClaimShareID is the claim naming the share link a share-verified token belongs to
	ClaimShareID = "sid"
	// ClaimRecipient is the claim carrying the sha256 hex of the normalized recipient email
	ClaimRecipient = "rcp"
)

// NormalizeRecipientEmail returns the trimmed, lowercased form of a recipient email
func NormalizeRecipientEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// RecipientHash returns the sha256 hex of the normalized recipient email
func RecipientHash(email string) string {
	sum := sha256.Sum256([]byte(NormalizeRecipientEmail(email)))

	return hex.EncodeToString(sum[:])
}

// ComposeShareToken builds the share token carried by the share URL from the owning organization id and the signature
func ComposeShareToken(orgID, signature string) string {
	return orgID + shareTokenSeparator + signature
}

// SplitShareToken returns the organization id and the signature carried by a share token
func SplitShareToken(token string) (string, string, bool) {
	return strings.Cut(token, shareTokenSeparator)
}

// ComposeShareURL builds the share URL under the storage links base URL carrying the share token
func ComposeShareURL(baseURL, shareID, tokenString string) string {
	return fmt.Sprintf("%s/%s/%s?token=%s", baseURL, sharePathSegment, url.PathEscape(shareID), url.QueryEscape(tokenString))
}

// NewShareVerifiedToken mints the token proving the recipient of a share verified their email
func NewShareVerifiedToken(tm *tokens.TokenManager, issuer, shareID, recipientHash string) (string, error) {
	if tm == nil {
		return "", ErrTokenManagerRequired
	}

	opts := []tokens.ConfigOpt{
		tokens.WithSignedTokenUse(),
		tokens.WithAudience(DownloadAudience),
		tokens.WithAccessDuration(ShareVerifiedDuration),
		tokens.WithClaim(ClaimType, ShareVerifiedTokenType),
		tokens.WithClaim(ClaimShareID, shareID),
		tokens.WithClaim(ClaimRecipient, recipientHash),
	}

	if issuer != "" {
		opts = append(opts, tokens.WithIssuer(issuer))
	}

	return tm.CreateSignedToken(opts...)
}

// VerifyShareVerifiedToken checks the signature, issuer, audience, expiry, and type of a share-verified token and
// returns its share id and recipient hash
func VerifyShareVerifiedToken(tm *tokens.TokenManager, issuer, tokenString string) (string, string, error) {
	opts := []tokens.ConfigOpt{tokens.WithAudience(DownloadAudience)}

	if issuer != "" {
		opts = append(opts, tokens.WithIssuer(issuer))
	}

	claims, err := tm.VerifySignedToken(tokenString, opts...)
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrShareVerifiedTokenInvalid, err)
	}

	if claims[ClaimType] != ShareVerifiedTokenType {
		return "", "", ErrShareVerifiedTokenType
	}

	shareID, _ := claims[ClaimShareID].(string)
	recipient, _ := claims[ClaimRecipient].(string)

	return shareID, recipient, nil
}
