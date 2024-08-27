package sessions

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/oauth2"
)

// SessionContextKey is the context key for the user claims
var SessionContextKey = &ContextKey{"SessionContextKey"}

// ContextKey is the key name for the additional context
type ContextKey struct {
	name string
}

// OhAuthTokenFromContext returns the Token from the ctx
func OhAuthTokenFromContext(ctx context.Context) (*oauth2.Token, error) {
	token, ok := ctx.Value(SessionContextKey).(*oauth2.Token)
	if !ok {
		return nil, errors.New("context missing Token")
	}

	return token, nil
}

// ContextWithToken returns a copy of ctx that stores the Token
func ContextWithToken(ctx context.Context, token *oauth2.Token) context.Context {
	return context.WithValue(ctx, SessionContextKey, token)
}

// UserIDFromContext returns the user ID from the ctx
// this function assumes the session data is stored in a string map
func UserIDFromContext(ctx context.Context) (string, error) {
	sessionDetails, ok := ctx.Value(SessionContextKey).(*Session[any])
	if !ok {
		return "", ErrInvalidSession
	}

	sessionID := sessionDetails.GetKey()

	sessionData, ok := sessionDetails.GetOk(sessionID)
	if !ok {
		return "", ErrInvalidSession
	}

	sd, ok := sessionData.(map[string]string)
	if !ok {
		return "", ErrInvalidSession
	}

	userID, ok := sd["userID"]
	if !ok {
		return "", ErrInvalidSession
	}

	return userID, nil
}

// ContextWithUserID returns a copy of ctx that stores the user ID
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	if strings.TrimSpace(userID) == "" {
		return ctx
	}

	return context.WithValue(ctx, SessionContextKey, userID)
}

// SessionToken returns the encoded session token
func SessionToken(ctx context.Context) (string, error) {
	sd := getSessionDataFromContext(ctx)
	if sd == nil {
		return "", ErrInvalidSession
	}

	sd.mu.Lock()
	defer sd.mu.Unlock()

	return sd.store.EncodeCookie(sd)
}

// addSessionDataToContext adds the session details to the context
func (s *Session[P]) addSessionDataToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, SessionContextKey, s)
}

// getSessionDataFromContext gets the session information from the context
func getSessionDataFromContext(ctx context.Context) *Session[map[string]any] {
	c, ok := ctx.Value(SessionContextKey).(*Session[map[string]any])
	if !ok {
		return nil
	}

	return c
}
