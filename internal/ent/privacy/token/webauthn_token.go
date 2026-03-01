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
	ctx := webauthnCreationContextKey.Set(parent, &WebauthnCreationContextKey{})

	return withTokenContextBypassCaller(ctx)
}

func WebauthnCreationContextKeyFromContext(ctx context.Context) *WebauthnCreationContextKey {
	w, ok := webauthnCreationContextKey.Get(ctx)
	if !ok {
		return nil
	}

	return w
}
