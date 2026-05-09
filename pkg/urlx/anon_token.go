package urlx

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/theopenlane/core/pkg/shortlinks"
	"github.com/theopenlane/iam/tokens"
)

// AnonTokenRequest holds the parameters for generating an anonymous JWT and
// embedding it in a URL
type AnonTokenRequest struct {
	// Prefix is prepended to SubjectID to form the JWT subject (e.g. "anon-tc-")
	Prefix string
	// SubjectID is the domain-specific identifier (e.g. request ID, ULID)
	SubjectID string
	// OrgID is the organization the token is scoped to
	OrgID string
	// Email is the recipient email address embedded in the token
	Email string
	// Duration is the access token lifetime
	Duration time.Duration
	// ExtraClaims is an optional callback to set domain-specific claim fields
	// (e.g. AssessmentID, TrustCenterID) on the token before signing
	ExtraClaims func(*tokens.Claims)
}

// AnonTokenResult holds the generated access token and the final URL containing it
type AnonTokenResult struct {
	// AccessToken is the signed JWT string
	AccessToken string
	// URL is the full URL with the token embedded as a query parameter
	URL string
}

// GenerateAnonTokenURL creates a short-lived anonymous JWT from the request
// parameters, appends it to baseURL as a query parameter, and optionally
// shortens the result
func GenerateAnonTokenURL(ctx context.Context, tm *tokens.TokenManager, sl *shortlinks.Client, baseURL url.URL, req AnonTokenRequest) (*AnonTokenResult, error) {
	subject := fmt.Sprintf("%s%s", req.Prefix, req.SubjectID)

	claims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: subject,
		},
		UserID: subject,
		OrgID:  req.OrgID,
		Email:  req.Email,
	}

	if req.ExtraClaims != nil {
		req.ExtraClaims(claims)
	}

	accessToken, _, err := tm.CreateTokenPair(claims, tokens.WithAccessDuration(req.Duration))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenCreationFailed, err)
	}

	tokenURL, err := BuildTokenURL(ctx, sl, baseURL, accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrURLConstructionFailed, err)
	}

	return &AnonTokenResult{
		AccessToken: accessToken,
		URL:         tokenURL,
	}, nil
}
