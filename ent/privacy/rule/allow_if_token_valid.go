package rule

import (
	"context"
	"reflect"

	"entgo.io/ent/entql"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/privacy/token"
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
	_, ok := contextx.From[T](ctx)
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
				return privacy.Deny
			}

			actualToken, ok := contextx.From[T](ctx)

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
		switch reflect.TypeOf(tokenType) {
		case reflect.TypeOf(&token.VerifyToken{}):
			if ContextHasPrivacyTokenOfType[*token.VerifyToken](ctx) {
				return true
			}
		case reflect.TypeOf(&token.OrgInviteToken{}):
			if ContextHasPrivacyTokenOfType[*token.OrgInviteToken](ctx) {
				return true
			}
		case reflect.TypeOf(&token.SignUpToken{}):
			if ContextHasPrivacyTokenOfType[*token.SignUpToken](ctx) {
				return true
			}
		case reflect.TypeOf(&token.OauthTooToken{}):
			if ContextHasPrivacyTokenOfType[*token.OauthTooToken](ctx) {
				return true
			}
		case reflect.TypeOf(&token.ResetToken{}):
			if ContextHasPrivacyTokenOfType[*token.ResetToken](ctx) {
				return true
			}
		case reflect.TypeOf(&token.JobRunnerRegistrationToken{}):
			if ContextHasPrivacyTokenOfType[*token.JobRunnerRegistrationToken](ctx) {
				return true
			}
		}
	}

	return false
}
