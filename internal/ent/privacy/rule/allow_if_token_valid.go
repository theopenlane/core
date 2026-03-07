package rule

import (
	"context"
	"reflect"

	"entgo.io/ent/entql"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/token"
)

// AllowIfContextHasPrivacyTokenOfType allows a mutation
// to proceed if a privacy token of a specific type is found in the
// context. It checks if the actual type of the token in the context
// matches the expected type, and if so, it returns `privacy.Allow`.
// If the types do not match, it returns `privacy.Skipf` with a message
// indicating that no token was found in the context with the expected type
func AllowIfContextHasPrivacyTokenOfType[T token.PrivacyToken]() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if ContextHasPrivacyTokenOfType[T](ctx) {
			return privacy.Allowf("found token in context")
		}

		return privacy.Skipf("no token found from context")
	})
}

// ContextHasPrivacyTokenOfType checks the context for the token type and returns true if they match
func ContextHasPrivacyTokenOfType[T token.PrivacyToken](ctx context.Context) bool {
	_, ok := privacyTokenFromContext[T](ctx)
	if !ok {
		return false
	}

	return ok
}

// PrivacyToken is an interface that defines a method to get the token string
type PrivacyToken interface {
	GetToken() string
}

// AllowAfterApplyingPrivacyTokenFilter allows the mutation to proceed
// if a privacy token of a specific type is found in the context. It
// also applies a privacy filter to the token before allowing the
// mutation to proceed
func AllowAfterApplyingPrivacyTokenFilter[T PrivacyToken]() privacy.QueryMutationRule {
	type Filter interface {
		WhereToken(p entql.StringP)
	}

	return privacy.FilterFunc(
		func(ctx context.Context, f privacy.Filter) error {
			tokenFilter, ok := f.(Filter)
			if !ok {
				return privacy.Denyf("unable to cast to token filter")
			}

			actualToken, ok := privacyTokenFromContext[T](ctx)

			if ok {
				tokenFilter.WhereToken(entql.StringEQ(actualToken.GetToken()))

				return privacy.Allowf("applied privacy token filter")
			}

			return privacy.Skipf("no token found from context")
		})
}

// SkipTokenInContext checks if the context has a token of a specific type
// and returns true if it does. It supports multiple token types
// and checks for each type in the skipTypes slice. If any of the types
// match, it returns true, otherwise it returns false
// This function is useful for skipping certain rules based on the presence
// of a token in the context
func SkipTokenInContext(ctx context.Context, skipTypes []token.PrivacyToken) bool {
	for _, tokenType := range skipTypes {
		if tokenType == nil {
			continue
		}

		if _, ok := privacyTokenFromContextByType(ctx, reflect.TypeOf(tokenType)); ok {
			return true
		}
	}

	return false
}

func privacyTokenFromContext[T any](ctx context.Context) (T, bool) {
	var zero T

	tokenType := reflect.TypeOf((*T)(nil)).Elem()
	rawToken, ok := privacyTokenFromContextByType(ctx, tokenType)
	if !ok {
		return zero, false
	}

	typedToken, ok := rawToken.(T)
	if !ok {
		return zero, false
	}

	return typedToken, true
}

func privacyTokenFromContextByType(ctx context.Context, tokenType reflect.Type) (any, bool) {
	switch tokenType {
	case reflect.TypeOf(&token.VerifyToken{}):
		t := token.VerifyTokenFromContext(ctx)
		return t, t != nil
	case reflect.TypeOf(&token.OrgInviteToken{}):
		t := token.OrgInviteTokenFromContext(ctx)
		return t, t != nil
	case reflect.TypeOf(&token.SignUpToken{}):
		t := token.EmailSignUpTokenFromContext(ctx)
		return t, t != nil
	case reflect.TypeOf(&token.OauthTooToken{}):
		t := token.OauthTooTokenFromContext(ctx)
		return t, t != nil
	case reflect.TypeOf(&token.ResetToken{}):
		t := token.ResetTokenFromContext(ctx)
		return t, t != nil
	case reflect.TypeOf(&token.JobRunnerRegistrationToken{}):
		t := token.JobRunnerRegistrationTokenFromContext(ctx)
		return t, t != nil
	case reflect.TypeOf(&token.DownloadToken{}):
		t := token.DownloadTokenFromContext(ctx)
		return t, t != nil
	default:
		return nil, false
	}
}
