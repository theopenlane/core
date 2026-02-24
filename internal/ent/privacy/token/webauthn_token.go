package token

import (
	"context"
)

type WebauthnCreationContextKey struct {
}

func NewWebauthnCreationContextKeyWithEmail() WebauthnCreationContextKey {
	return WebauthnCreationContextKey{}
}

func NewContextWithWebauthnCreationContextKey(parent context.Context) context.Context {
	return webauthnCreationContextKey.Set(parent, &WebauthnCreationContextKey{})
}

func WebauthnCreationContextKeyFromContext(ctx context.Context) *WebauthnCreationContextKey {
	w, ok := webauthnCreationContextKey.Get(ctx)
	if !ok {
		return nil
	}

	return w
}
