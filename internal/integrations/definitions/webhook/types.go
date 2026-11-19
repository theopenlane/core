package webhook

import (
	"encoding/json"

	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the webhook integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0WEBHOOK0000000000000001")
	// installation is the installation metadata handle
	installation = types.InstallationOf[InstallationMetadata]()
	// webhookConnection is the connection handle
	webhookConnection = types.ConnectionOf[CredentialSchema]()
)

// CredentialSchema holds the destination, the headers it requires and optional signing for one installation
type CredentialSchema struct {
	// URL is the destination every delivery is posted to
	URL string `json:"url" jsonschema:"required,title=Destination URL,description=HTTPS endpoint that receives the JSON deliveries."`
	// Headers are static headers added to every delivery, such as an API key the destination requires
	Headers map[string]string `json:"headers,omitempty" jsonschema:"title=Headers,secret=true,description=Static headers added to every delivery keyed by header name. Use these for an API key or token the destination requires."`
	// SigningSecret signs each delivery with HMAC-SHA256; deliveries are unsigned when empty
	SigningSecret string `json:"signingSecret,omitempty" jsonschema:"title=Signing Secret,secret=true,description=Secret used to sign each delivery with HMAC-SHA256 following Standard Webhooks. A whsec_ secret is base64 decoded. Leave empty to send unsigned deliveries."`
	// SignatureHeader replaces the Standard Webhooks signature with a signature of the body alone in this header, for a destination that dictates its own
	SignatureHeader string `json:"signatureHeader,omitempty" jsonschema:"title=Custom Signature Header,description=Only for a destination that requires its own signature header. The body alone is signed and sent in this header in place of the Standard Webhooks signature."`
	// SignatureFormat is the custom signature header value, with {signature} replaced by the encoded signature
	SignatureFormat string `json:"signatureFormat,omitempty" jsonschema:"title=Custom Signature Format,default={signature},description=Value of the custom signature header. {signature} is replaced by the encoded signature."`
	// SignatureEncoding is how the custom signature bytes are written
	SignatureEncoding string `json:"signatureEncoding,omitempty" jsonschema:"title=Custom Signature Encoding,enum=hex,enum=base64,default=hex,description=How the custom signature bytes are written."`
}

// InstallationMetadata holds the destination details for one installation
type InstallationMetadata struct {
	// Host is the destination host deliveries are posted to
	Host string `json:"host,omitempty" jsonschema:"title=Host"`
}

// Client posts signed deliveries to one installation's destination
type Client struct {
	// requester holds the destination, static headers and safe http client
	requester *httpsling.Requester
	// signingKey is the HMAC-SHA256 key; empty sends unsigned deliveries
	signingKey []byte
	// signatureHeader is the destination's own signature header; empty signs with Standard Webhooks
	signatureHeader string
	// signatureFormat is the custom signature header value with the signature placeholder
	signatureFormat string
	// signatureEncoding is how the custom signature bytes are written
	signatureEncoding string
	// host is the destination host
	host string
}

// DeliverOperation holds per-invocation parameters for one webhook delivery
type DeliverOperation struct {
	// ID identifies the message and is the same on every attempt; sent as webhook-id for the destination to deduplicate on
	ID string `json:"id" jsonschema:"required,title=Message ID,description=Unique message id sent as the webhook-id header. It must not contain a full stop."`
	// Body is the JSON document posted to the destination as-is
	Body json.RawMessage `json:"body" jsonschema:"required,title=Body,description=JSON document posted to the destination."`
}

// DeliverResult is the outcome of one webhook delivery
type DeliverResult struct {
	// StatusCode is the HTTP status the destination returned
	StatusCode int `json:"statusCode"`
}
