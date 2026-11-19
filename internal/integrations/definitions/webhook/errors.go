package webhook

import "errors"

var (
	// ErrURLInvalid indicates the destination url is unparsable or not allowed
	ErrURLInvalid = errors.New("webhook: destination url invalid")
	// ErrHTTPSRequired indicates the destination url does not use https
	ErrHTTPSRequired = errors.New("webhook: destination url must use https")
	// ErrHeaderNameInvalid indicates a static header name is invalid
	ErrHeaderNameInvalid = errors.New("webhook: header name invalid")
	// ErrSigningSecretInvalid indicates a whsec_ signing secret is not valid base64
	ErrSigningSecretInvalid = errors.New("webhook: signing secret invalid")
	// ErrMessageIDInvalid indicates the message id is empty or contains a full stop
	ErrMessageIDInvalid = errors.New("webhook: message id invalid")
	// ErrSignatureFormatInvalid indicates the signature format has no signature placeholder
	ErrSignatureFormatInvalid = errors.New("webhook: signature format must contain {signature}")
	// ErrSignatureEncodingInvalid indicates the signature encoding is not hex or base64
	ErrSignatureEncodingInvalid = errors.New("webhook: signature encoding must be hex or base64")
	// ErrClientBuild indicates the webhook client could not be constructed
	ErrClientBuild = errors.New("webhook: unable to build client")
	// ErrDeliveryFailed indicates a delivery failure that may succeed on retry
	ErrDeliveryFailed = errors.New("webhook: delivery failed")
	// ErrDeliveryRejected indicates the destination rejected the delivery
	ErrDeliveryRejected = errors.New("webhook: delivery rejected by destination")
	// ErrResultEncode indicates the delivery result could not be serialized
	ErrResultEncode = errors.New("webhook: result encode failed")
)
