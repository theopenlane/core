package keystore

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	credentialspb "cloud.google.com/go/iam/credentials/apiv1/credentialspb"
	"github.com/aws/aws-sdk-go-v2/aws"
	sts "github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/googleapis/gax-go/v2"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type stubTokenStore struct {
	bundle      *TokenBundle
	recordCalls []OAuthTokens
}

func (s *stubTokenStore) LoadTokens(context.Context, string, string) (*TokenBundle, error) {
	return s.bundle, nil
}

func (s *stubTokenStore) RecordMint(_ context.Context, payload OAuthTokens) error {
	s.recordCalls = append(s.recordCalls, payload)
	return nil
}

type stubValidator struct {
	info *IntegrationUserInfo
	err  error
}

func (s stubValidator) Validate(context.Context, string, *ProviderRuntime) (*IntegrationUserInfo, error) {
	return s.info, s.err
}

func TestBrokerRefreshOAuth(t *testing.T) {
	store := &stubTokenStore{
		bundle: &TokenBundle{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresAt:    ptrTime(time.Now().Add(30 * time.Minute)),
			Attributes: map[string]string{
				"custom": "value",
			},
			Metadata: map[string]any{
				"environment": "prod",
			},
			Scopes: []string{"repo"},
		},
	}
	registry := Registry{
		"github": {
			Spec: ProviderSpec{
				Name:     "github",
				AuthType: AuthTypeOAuth2,
			},
			OAuthConfig: &oauth2.Config{
				ClientID:     "id",
				ClientSecret: "secret",
				Scopes:       []string{"repo"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://example.com/auth",
					TokenURL: "https://example.com/token",
				},
			},
			Validator: stubValidator{
				info: &IntegrationUserInfo{
					ID:       "123",
					Email:    "user@example.com",
					Username: "octocat",
				},
			},
		},
	}

	b := newBroker(store, &registry).WithClock(func() time.Time {
		return time.Unix(1_600_000_000, 0)
	})

	creds, err := b.MintOAuthToken(context.Background(), "org-1", "github")
	if err != nil {
		t.Fatalf("MintOAuthToken returned error: %v", err)
	}
	if creds.AccessToken != "access-token" {
		t.Fatalf("expected access token to be reused, got %s", creds.AccessToken)
	}
	if creds.ProviderUsername != "octocat" {
		t.Fatalf("expected provider username to be set, got %s", creds.ProviderUsername)
	}
	if creds.Metadata["custom"] != "value" {
		t.Fatalf("expected metadata to include custom attribute, got %v", creds.Metadata)
	}
	if len(store.recordCalls) != 1 {
		t.Fatalf("expected one persistence call, got %d", len(store.recordCalls))
	}
	if got := store.recordCalls[0].Attributes["custom"]; got != "value" {
		t.Fatalf("expected attribute to be preserved, got %s", got)
	}
	if !reflect.DeepEqual(store.recordCalls[0].Scopes, []string{"repo"}) {
		t.Fatalf("expected scopes to be recorded, got %v", store.recordCalls[0].Scopes)
	}
	if env := store.recordCalls[0].Metadata["environment"]; env != "prod" {
		t.Fatalf("expected metadata environment to be persisted, got %v", env)
	}
}

func TestBrokerGitHubApp(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	store := &stubTokenStore{
		bundle: &TokenBundle{
			Attributes: map[string]string{
				"private_key":     string(privateKeyPEM),
				"installation_id": "9001",
				"app_id":          "1234",
			},
		},
	}

	base := time.Unix(1_600_000_000, 0).UTC()
	baseURL := "https://integration.test"
	expiry := base.Add(45 * time.Minute)

	httpClient := httpClientFunc(func(req *http.Request) (*http.Response, error) {
		expected := fmt.Sprintf("%s/app/installations/9001/access_tokens", baseURL)
		if req.URL.String() != expected {
			return nil, fmt.Errorf("unexpected url: %s", req.URL.String())
		}
		if got := req.Header.Get("Authorization"); !strings.HasPrefix(got, "Bearer ") {
			return nil, fmt.Errorf("expected authorization header, got %s", got)
		}
		resp := map[string]any{
			"token":      "installation-token",
			"expires_at": expiry.Format(time.RFC3339),
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(mustJSONMarshal(t, resp))),
			Header:     make(http.Header),
		}, nil
	})

	registry := Registry{
		"ghe": {
			Spec: ProviderSpec{
				Name:     "ghe",
				AuthType: AuthTypeGitHubApp,
				GitHubApp: &GitHubAppSpec{
					BaseURL: baseURL,
				},
			},
		},
	}

	b := newBroker(store, &registry).
		WithHTTPClient(httpClient).
		WithClock(func() time.Time { return base })

	creds, err := b.MintOAuthToken(context.Background(), "org-1", "ghe")
	if err != nil {
		t.Fatalf("MintOAuthToken returned error: %v", err)
	}
	if creds.AccessToken != "installation-token" {
		t.Fatalf("unexpected access token: %s", creds.AccessToken)
	}
	if creds.ExpiresAt == nil {
		t.Fatalf("expected expiry to be set")
	}
}

type httpClientFunc func(*http.Request) (*http.Response, error)

func (f httpClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

type fakeIAMClient struct {
	req     *credentialspb.GenerateAccessTokenRequest
	token   string
	expires time.Time
}

func (f *fakeIAMClient) GenerateAccessToken(_ context.Context, req *credentialspb.GenerateAccessTokenRequest, _ ...gax.CallOption) (*credentialspb.GenerateAccessTokenResponse, error) {
	f.req = req
	return &credentialspb.GenerateAccessTokenResponse{
		AccessToken: f.token,
		ExpireTime:  timestamppb.New(f.expires),
	}, nil
}

func (f *fakeIAMClient) Close() error { return nil }

func TestBrokerWorkloadIdentity(t *testing.T) {
	base := time.Unix(1_700_000_000, 0).UTC()
	bundle := &TokenBundle{
		Attributes: map[string]string{
			"audience":            "audience-value",
			"serviceAccountEmail": "svc@example.iam.gserviceaccount.com",
			"scopes":              "https://www.googleapis.com/auth/cloud-platform",
		},
	}
	store := &stubTokenStore{bundle: bundle}
	registry := Registry{
		"gcp_scc": {
			Spec: ProviderSpec{
				Name:     "gcp_scc",
				AuthType: AuthTypeWorkloadIdentity,
				WorkloadIdentity: &WorkloadIdentitySpec{
					TokenLifetime: 15 * time.Minute,
				},
			},
		},
	}

	stsResponse := map[string]any{
		"access_token": "sts-access-token",
		"expires_in":   3600,
		"token_type":   "Bearer",
	}
	httpClient := httpClientFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == stsEndpoint {
			body := readBody(t, req)
			if !strings.Contains(body, "audience=audience-value") {
				t.Fatalf("expected audience in STS payload, got %s", body)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(mustJSONMarshal(t, stsResponse))),
				Header:     make(http.Header),
			}, nil
		}
		return nil, &url.Error{Op: "post", URL: req.URL.String(), Err: errors.New("unexpected url")}
	})

	fakeClient := &fakeIAMClient{
		token:   "impersonated-token",
		expires: base.Add(30 * time.Minute),
	}

	b := newBroker(store, &registry).
		WithHTTPClient(httpClient).
		WithIAMClientFactory(func(context.Context, oauth2.TokenSource) (iamCredentialsClient, error) {
			return fakeClient, nil
		}).
		WithClock(func() time.Time { return base })

	b.WorkloadIdentityIssuer = WorkloadIdentityIssuerFunc(func(context.Context, string, *ProviderSpec, map[string]string) (*WorkloadIdentitySubjectToken, error) {
		return &WorkloadIdentitySubjectToken{Token: "subject-token", Type: "urn:ietf:params:oauth:token-type:id_token"}, nil
	})

	creds, err := b.MintOAuthToken(context.Background(), "org-1", "gcp_scc")
	if err != nil {
		t.Fatalf("MintOAuthToken returned error: %v", err)
	}
	if creds.AccessToken != "impersonated-token" {
		t.Fatalf("unexpected access token %s", creds.AccessToken)
	}
	if creds.ExpiresAt == nil || !creds.ExpiresAt.Equal(fakeClient.expires) {
		t.Fatalf("expected expiry %v, got %v", fakeClient.expires, creds.ExpiresAt)
	}
	if fakeClient.req == nil {
		t.Fatalf("expected IAM client to be invoked")
	}
	if got := strings.Join(fakeClient.req.Scope, ","); !strings.Contains(got, "cloud-platform") {
		t.Fatalf("expected scopes to contain cloud-platform, got %s", got)
	}
}

func TestBrokerAWSFederationStaticCredentials(t *testing.T) {
	store := &stubTokenStore{
		bundle: &TokenBundle{
			Attributes: map[string]string{
				"access_key_id":     "AKIA123",
				"secret_access_key": "secret",
				"session_token":     "session",
				"region":            "us-west-2",
			},
		},
	}

	registry := Registry{
		"aws-config": {
			Spec: ProviderSpec{
				Name:     "aws-config",
				AuthType: AuthTypeAWSSTS,
			},
		},
	}

	b := newBroker(store, &registry)
	creds, err := b.MintOAuthToken(context.Background(), "org-1", "aws-config")
	if err != nil {
		t.Fatalf("MintOAuthToken returned error: %v", err)
	}
	if creds.Metadata["access_key_id"] != "AKIA123" {
		t.Fatalf("expected access key metadata, got %v", creds.Metadata["access_key_id"])
	}
	if creds.AccessToken != "session" {
		t.Fatalf("expected session token passthrough, got %s", creds.AccessToken)
	}
	if len(store.recordCalls) != 0 {
		t.Fatalf("expected no persistence writes for aws federation, got %d", len(store.recordCalls))
	}
}

func TestBrokerAWSFederationAssumeRole(t *testing.T) {
	expiration := time.Unix(1_700_000_000, 0).UTC()
	fakeSTS := &fakeSTSClient{
		output: &sts.AssumeRoleWithWebIdentityOutput{
			Credentials: &ststypes.Credentials{
				AccessKeyId:     aws.String("AKIA456"),
				SecretAccessKey: aws.String("secret"),
				SessionToken:    aws.String("session-token"),
				Expiration:      aws.Time(expiration),
			},
			AssumedRoleUser: &ststypes.AssumedRoleUser{
				Arn: aws.String("arn:aws:sts::123456789012:assumed-role/test/Session"),
			},
		},
	}

	store := &stubTokenStore{
		bundle: &TokenBundle{
			AccessToken: "web-identity-token",
			Attributes: map[string]string{
				"role_arn":         "arn:aws:iam::123456789012:role/test",
				"session_name":     "custom-session",
				"session_duration": "5400",
				"region":           "us-east-2",
			},
		},
	}

	registry := Registry{
		"aws-config": {
			Spec: ProviderSpec{
				Name:     "aws-config",
				AuthType: AuthTypeAWSSTS,
			},
		},
	}

	b := newBroker(store, &registry).
		WithSTSClientFactory(func(region string) (stsClient, error) {
			if region != "us-east-2" {
				t.Fatalf("expected region us-east-2, got %s", region)
			}
			return fakeSTS, nil
		})

	creds, err := b.MintOAuthToken(context.Background(), "org-1", "aws-config")
	if err != nil {
		t.Fatalf("MintOAuthToken returned error: %v", err)
	}
	if fakeSTS.input == nil {
		t.Fatalf("expected STS client to be invoked")
	}
	if got := aws.ToString(fakeSTS.input.RoleArn); got != "arn:aws:iam::123456789012:role/test" {
		t.Fatalf("unexpected role arn %s", got)
	}
	if got := aws.ToInt32(fakeSTS.input.DurationSeconds); got != 5400 {
		t.Fatalf("expected duration 5400, got %d", got)
	}
	if creds.ExpiresAt == nil || !creds.ExpiresAt.Equal(expiration) {
		t.Fatalf("expected expiry %v, got %v", expiration, creds.ExpiresAt)
	}
	if creds.Metadata["assumed_role_arn"] == "" {
		t.Fatalf("expected assumed role arn metadata, got %v", creds.Metadata)
	}
	if creds.AccessToken != "session-token" {
		t.Fatalf("expected session token, got %s", creds.AccessToken)
	}
}

type fakeSTSClient struct {
	input  *sts.AssumeRoleWithWebIdentityInput
	output *sts.AssumeRoleWithWebIdentityOutput
	err    error
}

func (f *fakeSTSClient) AssumeRoleWithWebIdentity(_ context.Context, params *sts.AssumeRoleWithWebIdentityInput, _ ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	f.input = params
	return f.output, f.err
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func mustJSONMarshal(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return data
}

func readBody(t *testing.T, r *http.Request) string {
	t.Helper()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(data)
}
