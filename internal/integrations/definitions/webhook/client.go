package webhook

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

const (
	// deliveryTimeout bounds one delivery attempt; Standard Webhooks recommends 15 to 30 seconds
	deliveryTimeout = 15 * time.Second
	// reasonSigningSecretInvalid is the terminal reason for an undecodable signing secret
	reasonSigningSecretInvalid = "webhook signing secret is invalid"
	// reasonDestinationInvalid is the terminal reason for an unusable destination url
	reasonDestinationInvalid = "webhook destination url is invalid or not a public address"
)

// clientBuilder returns the webhook client constructor
func clientBuilder(allowPrivate bool) func(context.Context, types.ConnectionRequest[CredentialSchema]) (*Client, error) {
	return func(ctx context.Context, req types.ConnectionRequest[CredentialSchema]) (*Client, error) {
		return newClient(ctx, req.Credential, allowPrivate)
	}
}

// newClient builds the client; redirects are not followed so every hop passes the destination check
func newClient(ctx context.Context, cred CredentialSchema, allowPrivate bool) (*Client, error) {
	destination, err := urlx.ParseAbsolute(cred.URL)
	if err != nil {
		return nil, types.Terminal(fmt.Errorf("%w: %w", ErrURLInvalid, err), reasonDestinationInvalid)
	}

	key, err := signingKey(cred.SigningSecret)
	if err != nil {
		return nil, types.Terminal(err, reasonSigningSecretInvalid)
	}

	httpClient, err := urlx.NewCustomerHTTPClient(cred.URL, allowPrivate, httpclient.Timeout(deliveryTimeout), httpclient.NoRedirects())
	if err != nil {
		return nil, types.Terminal(fmt.Errorf("%w: %w", ErrURLInvalid, err), reasonDestinationInvalid)
	}

	requester, err := httpsling.New(
		httpsling.WithHTTPClient(httpClient),
		httpsling.URL(cred.URL),
		httpsling.Post(),
		httpsling.HeadersFromMap(cred.Headers),
		httpsling.ContentType(httpsling.ContentTypeJSON),
		httpsling.Accept(httpsling.ContentTypeJSON),
	)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("webhook: failed to create requester")

		return nil, ErrClientBuild
	}

	return &Client{
		requester:         requester,
		host:              destination.Host,
		signingKey:        key,
		signatureHeader:   cred.SignatureHeader,
		signatureFormat:   lo.CoalesceOrEmpty(cred.SignatureFormat, signaturePlaceholder),
		signatureEncoding: lo.CoalesceOrEmpty(cred.SignatureEncoding, encodingHex),
	}, nil
}
