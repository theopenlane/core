package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// Standard Webhooks (https://www.standardwebhooks.com) header names, signature version and secret prefix
const (
	// headerWebhookID carries the message id, the same on every attempt, for the destination to deduplicate on
	headerWebhookID = "webhook-id"
	// headerWebhookTimestamp carries the unix time of the attempt
	headerWebhookTimestamp = "webhook-timestamp"
	// headerWebhookSignature carries the Standard Webhooks signature
	headerWebhookSignature = "webhook-signature"
	// standardSignatureVersion precedes the base64 signature in the Standard Webhooks header
	standardSignatureVersion = "v1,"
	// standardSecretPrefix marks a Standard Webhooks secret, whose remainder is the base64 key
	standardSecretPrefix = "whsec_"
)

// settings for a destination that dictates its own signature header
const (
	// signaturePlaceholder is replaced by the encoded signature in the signature format
	signaturePlaceholder = "{signature}"
	// encodingHex writes the signature as lowercase hex
	encodingHex = "hex"
	// encodingBase64 writes the signature as standard base64
	encodingBase64 = "base64"
)

// signingKey returns the HMAC key for a secret: the decoded remainder of a whsec_ secret, otherwise the secret's bytes
func signingKey(secret string) ([]byte, error) {
	encoded, standard := strings.CutPrefix(secret, standardSecretPrefix)
	if !standard {
		return []byte(secret), nil
	}

	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSigningSecretInvalid, err)
	}

	return key, nil
}

// signature returns the signature header name and value for one attempt
func (c *Client) signature(id string, timestamp int64, body []byte) (string, string) {
	if c.signatureHeader == "" {
		signed := hmacSHA256(c.signingKey, fmt.Appendf(nil, "%s.%d.%s", id, timestamp, body))

		return headerWebhookSignature, standardSignatureVersion + base64.StdEncoding.EncodeToString(signed)
	}

	// a destination with its own header signs the body alone
	signed := hmacSHA256(c.signingKey, body)

	encoded := hex.EncodeToString(signed)
	if c.signatureEncoding == encodingBase64 {
		encoded = base64.StdEncoding.EncodeToString(signed)
	}

	return c.signatureHeader, strings.ReplaceAll(c.signatureFormat, signaturePlaceholder, encoded)
}

// hmacSHA256 returns the HMAC-SHA256 of content under key
func hmacSHA256(key, content []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(content)

	return mac.Sum(nil)
}
