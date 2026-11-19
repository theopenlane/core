package webhook

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/net/http/httpguts"

	"github.com/theopenlane/core/v2/internal/ent/validator"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

// schemeHTTPS is the only destination scheme accepted outside local development
const schemeHTTPS = "https"

// verify records the destination host without calling it, since any request is a delivery
func verify(_ context.Context, _ types.ConnectionRequest[CredentialSchema], c *Client) (InstallationMetadata, error) {
	return InstallationMetadata{Host: c.host}, nil
}

// credentialValidator returns the connect-time credential check
func credentialValidator(allowPrivate bool) func(context.Context, types.InstallationRequest, *CredentialSchema) error {
	return func(_ context.Context, _ types.InstallationRequest, cred *CredentialSchema) error {
		return validateCredential(*cred, allowPrivate)
	}
}

// validateCredential checks the destination url, header names and signature settings
func validateCredential(cred CredentialSchema, allowPrivate bool) error {
	destination, err := urlx.ParseAbsolute(cred.URL)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrURLInvalid, err)
	}

	if !allowPrivate {
		if destination.Scheme != schemeHTTPS {
			return ErrHTTPSRequired
		}

		if _, err := urlx.ValidatePublicURL(cred.URL); err != nil {
			return fmt.Errorf("%w: %w", ErrURLInvalid, err)
		}
	}

	if _, err := signingKey(cred.SigningSecret); err != nil {
		return err
	}

	switch {
	case cred.SignatureFormat != "" && !strings.Contains(cred.SignatureFormat, signaturePlaceholder):
		return ErrSignatureFormatInvalid
	case !lo.Contains([]string{"", encodingHex, encodingBase64}, cred.SignatureEncoding):
		return ErrSignatureEncodingInvalid
	}

	// the signature header in use is checked with the static ones
	headers := lo.Assign(cred.Headers, map[string]string{lo.CoalesceOrEmpty(cred.SignatureHeader, headerWebhookSignature): ""})

	if name, found := lo.FindKeyBy(headers, func(name, _ string) bool { return !httpguts.ValidHeaderFieldName(name) }); found {
		return fmt.Errorf("%w: %q", ErrHeaderNameInvalid, name)
	}

	return validator.ValidateOutboundHeaders()(headers)
}
