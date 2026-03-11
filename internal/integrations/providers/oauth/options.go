package oauth

import (
	"maps"
	"slices"
	"strings"

	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
)

// buildAuthURLOpts constructs rp.AuthURLOpt values from a scope list and extra auth parameters
func buildAuthURLOpts(scopes []string, params map[string]string) []rp.AuthURLOpt {
	opts := make([]rp.AuthURLOpt, 0, len(params)+1)

	if len(scopes) > 0 {
		opts = append(opts, asAuthCodeOption[rp.AuthURLOpt](oauth2.SetAuthURLParam("scope", strings.Join(scopes, " "))))
	}

	opts = append(opts, mapAuthCodeOptions[rp.AuthURLOpt](params)...)

	return opts
}

// buildCodeExchangeOpts constructs rp.CodeExchangeOpt values from extra token parameters
func buildCodeExchangeOpts(params map[string]string) []rp.CodeExchangeOpt {
	return mapAuthCodeOptions[rp.CodeExchangeOpt](params)
}

// mapAuthCodeOptions converts a string map into a slice of auth code options of the given type
func mapAuthCodeOptions[T ~func() []oauth2.AuthCodeOption](params map[string]string) []T {
	if len(params) == 0 {
		return nil
	}

	opts := make([]T, 0, len(params))

	for _, key := range slices.Sorted(maps.Keys(params)) {
		opts = append(opts, asAuthCodeOption[T](oauth2.SetAuthURLParam(key, params[key])))
	}

	return opts
}

// asAuthCodeOption wraps a single oauth2.AuthCodeOption into the target option function type
func asAuthCodeOption[T ~func() []oauth2.AuthCodeOption](option oauth2.AuthCodeOption) T {
	return T(func() []oauth2.AuthCodeOption {
		return []oauth2.AuthCodeOption{option}
	})
}
