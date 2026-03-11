package emailruntime

import (
	"fmt"
	"maps"

	"github.com/theopenlane/newman/compose"
)

// TemplateURLKey identifies a URL field under the .URLS template object.
// It is an alias for compose.URLKey; constants are defined in the compose package.
type TemplateURLKey = compose.URLKey

const (
	// TemplateURLRoot identifies the root application URL.
	TemplateURLRoot = compose.URLKeyRoot
	// TemplateURLProduct identifies the product home URL.
	TemplateURLProduct = compose.URLKeyProduct
	// TemplateURLDocs identifies the documentation URL.
	TemplateURLDocs = compose.URLKeyDocs
	// TemplateURLVerify identifies the email verification URL.
	TemplateURLVerify = compose.URLKeyVerify
	// TemplateURLInvite identifies the organization invite URL.
	TemplateURLInvite = compose.URLKeyInvite
	// TemplateURLPasswordReset identifies the password reset URL.
	TemplateURLPasswordReset = compose.URLKeyPasswordReset
	// TemplateURLVerifySubscriber identifies the subscriber verification URL.
	TemplateURLVerifySubscriber = compose.URLKeyVerifySubscriber
	// TemplateURLVerifyBilling identifies the billing verification URL.
	TemplateURLVerifyBilling = compose.URLKeyVerifyBilling
	// TemplateURLQuestionnaire identifies the questionnaire URL.
	TemplateURLQuestionnaire = compose.URLKeyQuestionnaire
)

// TemplateData is a typed builder for composing template data overrides.
// It keeps URL overrides strongly typed and supports tokenized URL composition.
type TemplateData struct {
	fields       map[string]any
	urls         map[TemplateURLKey]string
	tokenizedURL map[TemplateURLKey]string
}

// NewTemplateData creates an empty template data builder.
func NewTemplateData() *TemplateData {
	return &TemplateData{
		fields:       make(map[string]any),
		urls:         make(map[TemplateURLKey]string),
		tokenizedURL: make(map[TemplateURLKey]string),
	}
}

// WithField sets a top-level template field override.
func (d *TemplateData) WithField(key string, value any) *TemplateData {
	d.fields[key] = value

	return d
}

// WithFields sets multiple top-level template field overrides.
func (d *TemplateData) WithFields(fields map[string]any) *TemplateData {
	maps.Copy(d.fields, fields)

	return d
}

// WithURL overrides a URL field under .URLS with a fully resolved URL.
func (d *TemplateData) WithURL(key TemplateURLKey, value string) *TemplateData {
	d.urls[key] = value

	return d
}

// WithTokenURL composes a URL field under .URLS by appending a token to the configured base URL.
func (d *TemplateData) WithTokenURL(key TemplateURLKey, token string) *TemplateData {
	d.tokenizedURL[key] = token

	return d
}

// Build materializes the final map[string]any template data using compose.Config and compose.Recipient.
func (d *TemplateData) Build(config compose.Config, recipient compose.Recipient) (map[string]any, error) {
	overrides := make(map[string]any, len(d.fields)+1)
	maps.Copy(overrides, d.fields)

	urlBases := config.URLS.ToMap()
	urlOverrides := make(map[string]any, len(d.urls)+len(d.tokenizedURL))

	for key, value := range d.urls {
		urlOverrides[string(key)] = value
	}

	for key, token := range d.tokenizedURL {
		baseURL, ok := urlBases[key]
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnsupportedTemplateURLKey, key)
		}

		tokenURL, err := compose.AddTokenToURL(baseURL, token)
		if err != nil {
			return nil, err
		}

		urlOverrides[string(key)] = tokenURL
	}

	if len(urlOverrides) > 0 {
		overrides["URLS"] = compose.BuildTemplateURLs(config, urlOverrides)
	}

	return compose.BuildTemplateData(config, recipient, overrides), nil
}
