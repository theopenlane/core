package token

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

type WebauthnCreationContextKey struct {
}

func NewWebauthnCreationContextKeyWithEmail() WebauthnCreationContextKey {
	return WebauthnCreationContextKey{}
}

func NewContextWithWebauthnCreationContextKey(parent context.Context) context.Context {
	return contextx.With(parent, &WebauthnCreationContextKey{})
}

func WebauthnCreationContextKeyFromContext(ctx context.Context) *WebauthnCreationContextKey {
	w, ok := contextx.From[*WebauthnCreationContextKey](ctx)
	if !ok {
		return nil
	}

	return w
}
