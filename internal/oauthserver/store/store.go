// Package store loads, consumes, and revokes the OAuth tokens kept in the database
package store

import (
	"context"
	"errors"
	"time"

	iamoauth "github.com/theopenlane/iam/oauthserver"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/oauthtoken"
)

var (
	// ErrGrantRevoked is returned when a token's grant has been revoked or deleted
	ErrGrantRevoked = errors.New("this application's access has been revoked")
	// ErrTokenConsumed is returned when a single-use or revoked token is presented again
	ErrTokenConsumed = errors.New("oauth token already used or revoked")
	// ErrTokenExpired is returned when a token is past its expiry
	ErrTokenExpired = errors.New("oauth token expired")
)

// Load returns the row for a presented token of the given kind, verified against the stored hash
func Load(ctx context.Context, query *ent.OauthTokenQuery, issuer *iamoauth.TokenIssuer, raw string, kind iamoauth.TokenKind) (*ent.OauthToken, error) {
	presented, id, err := issuer.Lookup(raw)
	if err != nil || presented != kind {
		return nil, iamoauth.ErrTokenInvalid
	}

	row, err := query.Where(oauthtoken.ID(id.String()), oauthtoken.KindEQ(kind)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, iamoauth.ErrTokenInvalid
		}

		return nil, err
	}

	// the kind compared here comes from the stored row, never from the presented token
	stored := iamoauth.StoredToken{ID: row.ID, Kind: row.Kind, Hash: row.TokenHash, KeyVersion: row.KeyVersion}
	if err := issuer.Verify(raw, stored); err != nil {
		return nil, iamoauth.ErrTokenInvalid
	}

	return row, nil
}

// Redeem loads a token with its grant and checks it is unused, unexpired, and under an active grant
// a consumed token is returned with ErrTokenConsumed so callers can revoke its grant
func Redeem(ctx context.Context, client *ent.Client, issuer *iamoauth.TokenIssuer, raw string, kind iamoauth.TokenKind) (*ent.OauthToken, error) {
	token, err := Load(ctx, client.OauthToken.Query().WithOauthGrant(), issuer, raw, kind)
	if err != nil {
		return nil, err
	}

	switch {
	case token.ConsumedAt != nil:
		return token, ErrTokenConsumed
	case time.Now().After(token.ExpiresAt):
		return nil, ErrTokenExpired
	case !GrantActive(token.Edges.OauthGrant):
		return nil, ErrGrantRevoked
	}

	return token, nil
}

// GrantActive reports whether a grant exists and has not been revoked or deleted
func GrantActive(grant *ent.OauthGrant) bool {
	return grant != nil && grant.IsActive && grant.RevokedAt == nil && grant.DeletedAt.IsZero()
}

// Consume marks a single-use token consumed and reports false if another request consumed it first
func Consume(ctx context.Context, client *ent.Client, id string) (bool, error) {
	n, err := client.OauthToken.Update().
		Where(
			oauthtoken.ID(id),
			oauthtoken.ConsumedAtIsNil(),
		).
		SetConsumedAt(time.Now()).
		Save(ctx)

	return n == 1, err
}

// RevokeGrantTokens consumes every outstanding code, access token, and refresh token under a grant
func RevokeGrantTokens(ctx context.Context, client *ent.Client, grantID string) error {
	return client.OauthToken.Update().
		Where(
			oauthtoken.OauthGrantID(grantID),
			oauthtoken.ConsumedAtIsNil(),
		).
		SetConsumedAt(time.Now()).
		Exec(ctx)
}
