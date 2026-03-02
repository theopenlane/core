package oauth

import (
	"maps"
	"slices"
	"strings"

	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
)

func buildAuthURLOpts(scopes []string, params map[string]string) []rp.AuthURLOpt {
	options := make([]rp.AuthURLOpt, 0, len(params)+1)
	if len(scopes) > 0 {
		options = append(options, asAuthCodeOption[rp.AuthURLOpt](oauth2.SetAuthURLParam("scope", strings.Join(scopes, " "))))
	}
	options = append(options, mapAuthCodeOptions[rp.AuthURLOpt](params)...)

	return options
}

func buildCodeExchangeOpts(params map[string]string) []rp.CodeExchangeOpt {
	return mapAuthCodeOptions[rp.CodeExchangeOpt](params)
}

func mapAuthCodeOptions[T ~func() []oauth2.AuthCodeOption](params map[string]string) []T {
	if len(params) == 0 {
		return nil
	}

	options := make([]T, 0, len(params))
	for _, key := range slices.Sorted(maps.Keys(params)) {
		options = append(options, asAuthCodeOption[T](oauth2.SetAuthURLParam(key, params[key])))
	}

	return options
}

func asAuthCodeOption[T ~func() []oauth2.AuthCodeOption](option oauth2.AuthCodeOption) T {
	return T(func() []oauth2.AuthCodeOption {
		return []oauth2.AuthCodeOption{option}
	})
}
