package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/theopenlane/httpsling"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/validator"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

const (
	// testSigningSecret is a plain signing secret, used as the key as-is
	testSigningSecret = "test-signing-secret"
	// testStandardSecret is a Standard Webhooks secret whose key is the base64-decoded remainder
	testStandardSecret = "whsec_MfKQ9r8GKYqrTwjUPD8ILPZIo2LaLaSw"
	// testMessageID is the message id every test delivery carries
	testMessageID = "msg_2KWPBgLlAfxdpx2AI54pPJ85f4W"
	// testBody is the JSON encoding of testPayload that the destination should receive
	testBody = `{"event":"control.updated","id":"ctl_123"}`
	// testSignature is the hex HMAC-SHA256 of testBody alone under testSigningSecret, computed outside this package
	testSignature = "e38cc9e359b4e494b5e0022f21f7b53f5e7a28016807984ae7323a0a1a3db52c"
	// testSignatureBase64 is testSignature's bytes in standard base64
	testSignatureBase64 = "44zJ41m05JS14AIvIfe1P156KAFoB5hK5zI6Cho9tSw="
	// testLoopbackURL is a non-public destination no test server listens on
	testLoopbackURL = "http://127.0.0.1:8080/hook"
	// testPublicURL is a public destination that is never dialed
	testPublicURL = "https://hooks.example.com/openlane"
)

// testPayload is the delivery body every test posts
var testPayload = json.RawMessage(testBody)

// capturedRequest is the request a test destination received
type capturedRequest struct {
	// method is the received HTTP method
	method string
	// header is the received request header
	header http.Header
	// body is the received request body
	body string
}

// newTestDestination starts a destination that records the request and answers with status
func newTestDestination(t *testing.T, status int, cred CredentialSchema) (*Client, *capturedRequest) {
	t.Helper()

	captured := &capturedRequest{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}

		captured.method = r.Method
		captured.header = r.Header.Clone()
		captured.body = string(body)

		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)

	cred.URL = server.URL

	client, err := newClient(context.Background(), cred, true)
	assert.NilError(t, err)

	return client, captured
}

// standardSignature computes the Standard Webhooks signature with the standard library, independently of the package
func standardSignature(t *testing.T, key []byte, captured *capturedRequest) string {
	t.Helper()

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(captured.header.Get("webhook-id") + "." + captured.header.Get("webhook-timestamp") + "." + captured.body))

	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// TestDeliverFollowsStandardWebhooks verifies the id, timestamp and signature headers of the default scheme
func TestDeliverFollowsStandardWebhooks(t *testing.T) {
	t.Parallel()

	standardKey, err := base64.StdEncoding.DecodeString("MfKQ9r8GKYqrTwjUPD8ILPZIo2LaLaSw")
	assert.NilError(t, err)

	tests := []struct {
		name   string
		secret string
		key    []byte
	}{
		{name: "whsec secret is base64 decoded", secret: testStandardSecret, key: standardKey},
		{name: "plain secret is used as-is", secret: testSigningSecret, key: []byte(testSigningSecret)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, captured := newTestDestination(t, http.StatusOK, CredentialSchema{SigningSecret: tt.secret})

			before := time.Now().Unix()

			raw, err := deliver(context.Background(), types.OperationRequest{}, client, DeliverOperation{ID: testMessageID, Body: testPayload})
			assert.NilError(t, err)

			var result DeliverResult
			assert.NilError(t, json.Unmarshal(raw, &result))
			assert.Check(t, is.Equal(http.StatusOK, result.StatusCode))
			assert.Check(t, is.Equal(http.MethodPost, captured.method))
			assert.Check(t, is.Equal(testBody, captured.body))
			assert.Check(t, is.Equal("application/json", captured.header.Get("Content-Type")))
			assert.Check(t, is.Equal(testMessageID, captured.header.Get("webhook-id")))

			timestamp, err := strconv.ParseInt(captured.header.Get("webhook-timestamp"), 10, 64)
			assert.NilError(t, err)
			assert.Check(t, timestamp >= before && timestamp <= time.Now().Unix())

			assert.Check(t, is.Equal(standardSignature(t, tt.key, captured), captured.header.Get("webhook-signature")))
		})
	}
}

// TestDeliverCustomSignatureHeader verifies a destination's own header replaces the standard signature and signs the body alone
func TestDeliverCustomSignatureHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cred      CredentialSchema
		wantValue string
	}{
		{
			name:      "bare hex by default",
			cred:      CredentialSchema{SigningSecret: testSigningSecret, SignatureHeader: "X-Vendor-Signature"},
			wantValue: testSignature,
		},
		{
			name:      "prefixed",
			cred:      CredentialSchema{SigningSecret: testSigningSecret, SignatureHeader: "X-Vendor-Signature", SignatureFormat: "sha256={signature}"},
			wantValue: "sha256=" + testSignature,
		},
		{
			name:      "base64",
			cred:      CredentialSchema{SigningSecret: testSigningSecret, SignatureHeader: "X-Vendor-Signature", SignatureEncoding: "base64"},
			wantValue: testSignatureBase64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, captured := newTestDestination(t, http.StatusOK, tt.cred)

			_, err := deliver(context.Background(), types.OperationRequest{}, client, DeliverOperation{ID: testMessageID, Body: testPayload})
			assert.NilError(t, err)
			assert.Check(t, is.Equal(tt.wantValue, captured.header.Get("X-Vendor-Signature")))
			assert.Check(t, is.Len(captured.header.Values("webhook-signature"), 0))
			assert.Check(t, is.Equal(testMessageID, captured.header.Get("webhook-id")))
		})
	}
}

// TestDeliverUnsignedWithoutSecret verifies a destination that only needs an API key header gets no signature
func TestDeliverUnsignedWithoutSecret(t *testing.T) {
	t.Parallel()

	client, captured := newTestDestination(t, http.StatusOK, CredentialSchema{Headers: map[string]string{"X-Api-Key": "key-123"}})

	_, err := deliver(context.Background(), types.OperationRequest{}, client, DeliverOperation{ID: testMessageID, Body: testPayload})
	assert.NilError(t, err)
	assert.Check(t, is.Equal("key-123", captured.header.Get("X-Api-Key")))
	assert.Check(t, is.Equal(testMessageID, captured.header.Get("webhook-id")))
	assert.Check(t, is.Len(captured.header.Values("webhook-signature"), 0))
}

// TestDeliverRejectsUnusableMessageID verifies an id that would break the signed content is refused without a request
func TestDeliverRejectsUnusableMessageID(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"", "msg.with.stops"} {
		client, captured := newTestDestination(t, http.StatusOK, CredentialSchema{})

		_, err := deliver(context.Background(), types.OperationRequest{}, client, DeliverOperation{ID: id, Body: testPayload})
		assert.Check(t, is.ErrorIs(err, ErrMessageIDInvalid))

		_, terminal := types.TerminalFrom(err)
		assert.Check(t, terminal)
		assert.Check(t, is.Equal("", captured.method))
	}
}

func TestDeliverSendsStaticHeaders(t *testing.T) {
	t.Parallel()

	client, captured := newTestDestination(t, http.StatusOK, CredentialSchema{Headers: map[string]string{"X-Tenant": "acme", "Authorization": "Bearer token-123"}})

	_, err := deliver(context.Background(), types.OperationRequest{}, client, DeliverOperation{ID: testMessageID, Body: testPayload})
	assert.NilError(t, err)
	assert.Check(t, is.Equal("acme", captured.header.Get("X-Tenant")))
	assert.Check(t, is.Equal("Bearer token-123", captured.header.Get("Authorization")))
}

func TestDeliverFailureClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   int
		terminal bool
		sentinel error
	}{
		{name: "bad request is terminal", status: http.StatusBadRequest, terminal: true, sentinel: ErrDeliveryRejected},
		{name: "not found is terminal", status: http.StatusNotFound, terminal: true, sentinel: ErrDeliveryRejected},
		{name: "redirect is terminal", status: http.StatusFound, terminal: true, sentinel: ErrDeliveryRejected},
		{name: "server error retries", status: http.StatusInternalServerError, terminal: false, sentinel: ErrDeliveryFailed},
		{name: "unavailable retries", status: http.StatusServiceUnavailable, terminal: false, sentinel: ErrDeliveryFailed},
		{name: "rate limited retries", status: http.StatusTooManyRequests, terminal: false, sentinel: ErrDeliveryFailed},
		{name: "request timeout retries", status: http.StatusRequestTimeout, terminal: false, sentinel: ErrDeliveryFailed},
		{name: "gone is terminal", status: http.StatusGone, terminal: true, sentinel: ErrDeliveryRejected},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, _ := newTestDestination(t, tt.status, CredentialSchema{})

			_, err := deliver(context.Background(), types.OperationRequest{}, client, DeliverOperation{ID: testMessageID, Body: testPayload})
			assert.Check(t, is.ErrorIs(err, tt.sentinel))

			_, terminal := types.TerminalFrom(err)
			assert.Check(t, is.Equal(tt.terminal, terminal))
		})
	}
}

func TestDeliverTransportErrorRetries(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.NotFoundHandler())
	url := server.URL

	server.Close()

	client, err := newClient(context.Background(), CredentialSchema{URL: url, SigningSecret: testSigningSecret}, true)
	assert.NilError(t, err)

	_, err = deliver(context.Background(), types.OperationRequest{}, client, DeliverOperation{ID: testMessageID, Body: testPayload})
	assert.Check(t, is.ErrorIs(err, ErrDeliveryFailed))

	_, terminal := types.TerminalFrom(err)
	assert.Check(t, !terminal)
}

func TestDeliverNonPublicDestinationIsTerminal(t *testing.T) {
	t.Parallel()

	httpClient, err := urlx.NewHTTPClient(urlx.PublicOnly())
	assert.NilError(t, err)

	requester, err := httpsling.New(httpsling.WithHTTPClient(httpClient), httpsling.URL(testLoopbackURL), httpsling.Post())
	assert.NilError(t, err)

	client := &Client{requester: requester}

	_, err = deliver(context.Background(), types.OperationRequest{}, client, DeliverOperation{ID: testMessageID, Body: testPayload})
	assert.Check(t, is.ErrorIs(err, urlx.ErrNonPublicDestination))

	_, terminal := types.TerminalFrom(err)
	assert.Check(t, terminal)
}

func TestNewClientRejectsUnparsableDestination(t *testing.T) {
	t.Parallel()

	_, err := newClient(context.Background(), CredentialSchema{URL: "not a url", SigningSecret: testSigningSecret}, true)
	assert.Check(t, is.ErrorIs(err, ErrURLInvalid))

	_, terminal := types.TerminalFrom(err)
	assert.Check(t, terminal)
}

func TestVerifyRecordsHost(t *testing.T) {
	t.Parallel()

	client, err := newClient(context.Background(), CredentialSchema{URL: testPublicURL, SigningSecret: testSigningSecret}, false)
	assert.NilError(t, err)

	metadata, err := verify(context.Background(), types.ConnectionRequest[CredentialSchema]{}, client)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("hooks.example.com", metadata.Host))
}

func TestNewClientRejectsNonPublicDestination(t *testing.T) {
	t.Parallel()

	cred := CredentialSchema{URL: testLoopbackURL, SigningSecret: testSigningSecret}

	_, err := newClient(context.Background(), cred, false)
	assert.Check(t, is.ErrorIs(err, urlx.ErrNonPublicDestination))

	_, terminal := types.TerminalFrom(err)
	assert.Check(t, terminal)

	_, err = newClient(context.Background(), cred, true)
	assert.Check(t, is.Nil(err))
}

func TestValidateCredential(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		cred         CredentialSchema
		allowPrivate bool
		expected     error
	}{
		{name: "public https", cred: CredentialSchema{URL: testPublicURL}},
		{name: "http rejected", cred: CredentialSchema{URL: "http://hooks.example.com/openlane"}, expected: ErrHTTPSRequired},
		{name: "loopback rejected", cred: CredentialSchema{URL: "https://127.0.0.1/hook"}, expected: urlx.ErrNonPublicDestination},
		{name: "ipv6 loopback rejected", cred: CredentialSchema{URL: "https://[::1]/hook"}, expected: urlx.ErrNonPublicDestination},
		{name: "localhost rejected", cred: CredentialSchema{URL: "https://localhost/hook"}, expected: urlx.ErrNonPublicDestination},
		{name: "private network rejected", cred: CredentialSchema{URL: "https://10.0.0.5/hook"}, expected: urlx.ErrNonPublicDestination},
		{name: "cloud metadata address rejected", cred: CredentialSchema{URL: "https://169.254.169.254/latest/meta-data"}, expected: urlx.ErrNonPublicDestination},
		{name: "internal hostname rejected", cred: CredentialSchema{URL: "https://hooks.internal/openlane"}, expected: urlx.ErrNonPublicDestination},
		{name: "loopback http allowed in dev mode", cred: CredentialSchema{URL: testLoopbackURL}, allowPrivate: true},
		{name: "unparsable url", cred: CredentialSchema{URL: "not a url"}, expected: ErrURLInvalid},
		{name: "no signing secret", cred: CredentialSchema{URL: testPublicURL, Headers: map[string]string{"X-Api-Key": "key-123"}}},
		{name: "vendor signature settings", cred: CredentialSchema{URL: testPublicURL, SigningSecret: testSigningSecret, SignatureHeader: "X-Vendor-Signature", SignatureFormat: "v1={signature}", SignatureEncoding: "base64"}},
		{name: "standard webhooks secret", cred: CredentialSchema{URL: testPublicURL, SigningSecret: testStandardSecret}},
		{name: "whsec secret that is not base64", cred: CredentialSchema{URL: testPublicURL, SigningSecret: "whsec_not base64!"}, expected: ErrSigningSecretInvalid},
		{name: "signature format without placeholder", cred: CredentialSchema{URL: testPublicURL, SignatureFormat: "sha256="}, expected: ErrSignatureFormatInvalid},
		{name: "unknown signature encoding", cred: CredentialSchema{URL: testPublicURL, SignatureEncoding: "base32"}, expected: ErrSignatureEncodingInvalid},
		{name: "invalid signature header name", cred: CredentialSchema{URL: testPublicURL, SignatureHeader: "X Bad"}, expected: ErrHeaderNameInvalid},
		{name: "metadata signature header blocked", cred: CredentialSchema{URL: testPublicURL, SignatureHeader: "Metadata-Flavor"}, expected: validator.ErrHeaderNotAllowed},
		{name: "invalid header name", cred: CredentialSchema{URL: testPublicURL, Headers: map[string]string{"X Bad": "value"}}, expected: ErrHeaderNameInvalid},
		{name: "metadata header blocked", cred: CredentialSchema{URL: testPublicURL, Headers: map[string]string{"Metadata-Flavor": "Google"}}, expected: validator.ErrHeaderNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateCredential(tt.cred, tt.allowPrivate)
			if tt.expected == nil {
				assert.Check(t, is.Nil(err))

				return
			}

			assert.Check(t, is.ErrorIs(err, tt.expected))
		})
	}
}
