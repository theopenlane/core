package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/theopenlane/core/internal/integrations"
)

const DefaultSchemaVersion = "v1"

var (
	supportedSchemaVersions = map[string]struct{}{
		DefaultSchemaVersion: {},
	}
	envPlaceholderExpr = regexp.MustCompile(`\$\{[A-Za-z0-9_]+\}`)
)

func (s *ProviderSpec) supportsSchemaVersion() bool {
	if s == nil {
		return false
	}
	version := s.SchemaVersion
	if version == "" {
		version = DefaultSchemaVersion
	}
	_, ok := supportedSchemaVersions[version]
	return ok
}

func (s *ProviderSpec) interpolatePlaceholders(lookup EnvLookup) error {
	if s == nil || lookup == nil {
		return nil
	}

	if err := interpolateStringTargets(lookup,
		&s.Name,
		&s.DisplayName,
		&s.Category,
		&s.LogoURL,
		&s.DocsURL,
		&s.SchemaVersion,
	); err != nil {
		return err
	}

	if err := interpolateStringMap(s.Labels, lookup); err != nil {
		return err
	}

	if err := interpolateAnyMap(s.Metadata, lookup); err != nil {
		return err
	}
	if err := interpolateAnyMap(s.Defaults, lookup); err != nil {
		return err
	}

	if err := interpolateAnyMap(s.CredentialsSchema, lookup); err != nil {
		return err
	}

	if s.OAuth != nil {
		if err := interpolateStringTargets(lookup,
			&s.OAuth.ClientID,
			&s.OAuth.ClientSecret,
			&s.OAuth.AuthURL,
			&s.OAuth.TokenURL,
			&s.OAuth.OIDCDiscovery,
			&s.OAuth.RedirectURI,
		); err != nil {
			return err
		}
		if err := interpolateStringSliceInPlace(s.OAuth.Scopes, lookup); err != nil {
			return err
		}
		if err := interpolateStringSliceInPlace(s.OAuth.AdditionalHosts, lookup); err != nil {
			return err
		}
		if err := interpolateStringMap(s.OAuth.AuthParams, lookup); err != nil {
			return err
		}
		if err := interpolateStringMap(s.OAuth.TokenParams, lookup); err != nil {
			return err
		}
	}

	if s.APIKey != nil {
		if err := interpolateStringTargets(lookup,
			&s.APIKey.KeyLabel,
			&s.APIKey.HeaderName,
			&s.APIKey.QueryParam,
		); err != nil {
			return err
		}
	}

	if s.UserInfo != nil {
		if err := interpolateStringTargets(lookup,
			&s.UserInfo.URL,
			&s.UserInfo.Method,
			&s.UserInfo.AuthStyle,
			&s.UserInfo.AuthHeader,
			&s.UserInfo.IDPath,
			&s.UserInfo.EmailPath,
			&s.UserInfo.LoginPath,
			&s.UserInfo.SecondaryEmailURL,
		); err != nil {
			return err
		}
	}

	if s.WorkloadIdentity != nil {
		if err := interpolateStringTargets(lookup,
			&s.WorkloadIdentity.Audience,
			&s.WorkloadIdentity.TargetServiceAccount,
			&s.WorkloadIdentity.SubjectTokenType,
		); err != nil {
			return err
		}
		if err := interpolateStringSliceInPlace(s.WorkloadIdentity.Scopes, lookup); err != nil {
			return err
		}
	}

	if s.GitHubApp != nil {
		if err := interpolateStringTargets(lookup, &s.GitHubApp.BaseURL); err != nil {
			return err
		}
	}

	if s.AWSSTS != nil {
		if err := interpolateStringTargets(lookup,
			&s.AWSSTS.RoleARN,
			&s.AWSSTS.SessionName,
			&s.AWSSTS.Region,
			&s.AWSSTS.ExternalID,
		); err != nil {
			return err
		}
	}

	return nil
}

func interpolateStringTargets(lookup EnvLookup, targets ...*string) error {
	if lookup == nil {
		return nil
	}
	for _, target := range targets {
		if target == nil || *target == "" {
			continue
		}
		replaced, err := interpolateString(*target, lookup)
		if err != nil {
			return err
		}
		*target = replaced
	}
	return nil
}

func interpolateStringSliceInPlace(values []string, lookup EnvLookup) error {
	if lookup == nil || len(values) == 0 {
		return nil
	}
	for idx, value := range values {
		if value == "" {
			continue
		}
		replaced, err := interpolateString(value, lookup)
		if err != nil {
			return err
		}
		values[idx] = replaced
	}
	return nil
}

func interpolateStringMap(values map[string]string, lookup EnvLookup) error {
	if lookup == nil || len(values) == 0 {
		return nil
	}
	for key, value := range values {
		if value == "" {
			continue
		}
		replaced, err := interpolateString(value, lookup)
		if err != nil {
			return err
		}
		values[key] = replaced
	}
	return nil
}

func interpolateAnyMap(values map[string]any, lookup EnvLookup) error {
	if lookup == nil || len(values) == 0 {
		return nil
	}
	for key, raw := range values {
		interpolated, err := interpolateAny(raw, lookup)
		if err != nil {
			return err
		}
		values[key] = interpolated
	}
	return nil
}

func interpolateAnySlice(values []any, lookup EnvLookup) error {
	if lookup == nil || len(values) == 0 {
		return nil
	}
	for idx, raw := range values {
		interpolated, err := interpolateAny(raw, lookup)
		if err != nil {
			return err
		}
		values[idx] = interpolated
	}
	return nil
}

func interpolateAny(value any, lookup EnvLookup) (any, error) {
	switch typed := value.(type) {
	case string:
		return interpolateString(typed, lookup)
	case map[string]any:
		if err := interpolateAnyMap(typed, lookup); err != nil {
			return nil, err
		}
		return typed, nil
	case []any:
		if err := interpolateAnySlice(typed, lookup); err != nil {
			return nil, err
		}
		return typed, nil
	default:
		return value, nil
	}
}

func interpolateString(value string, lookup EnvLookup) (string, error) {
	if lookup == nil || value == "" {
		return value, nil
	}
	if !strings.Contains(value, "${") {
		return value, nil
	}

	var interpolationErr error
	replaced := envPlaceholderExpr.ReplaceAllStringFunc(value, func(segment string) string {
		if interpolationErr != nil {
			return segment
		}
		key := strings.TrimSuffix(strings.TrimPrefix(segment, "${"), "}")
		if key == "" {
			interpolationErr = fmt.Errorf("integrations/config: empty env placeholder in %q", value)
			return segment
		}
		resolved, ok := lookup(key)
		if !ok {
			interpolationErr = fmt.Errorf("%w: %s", integrations.ErrEnvVarNotDefined, key)
			return segment
		}
		return resolved
	})
	if interpolationErr != nil {
		return "", interpolationErr
	}
	return replaced, nil
}
