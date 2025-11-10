package keystore

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/theopenlane/httpsling"
)

func withHTTPRequester(t *testing.T, doer httpsling.Doer) func() {
	t.Helper()
	original := defaultHTTPRequester
	defaultHTTPRequester = httpsling.MustNew(httpsling.WithDoer(doer))
	return func() {
		defaultHTTPRequester = original
	}
}

func TestDefaultUserInfoValidator_Bearer(t *testing.T) {
	restore := withHTTPRequester(t, httpsling.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET method, got %s", req.Method)
		}
		if got := req.Header.Get(httpsling.HeaderAuthorization); got != "Bearer token-123" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		body := `{"id":"42","email":"user@example.com","login":"theuser"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioNopCloser(body),
			Header:     make(http.Header),
		}, nil
	}))
	defer restore()

	rt := &ProviderRuntime{
		Spec: ProviderSpec{
			UserInfo: &UserInfoSpec{
				URL:       "https://example.com/me",
				Method:    GET,
				AuthStyle: AuthHeaderStyleBearer,
				IDPath:    "id",
				EmailPath: "email",
				LoginPath: "login",
			},
		},
	}

	info, err := DefaultUserInfoValidator{}.Validate(context.Background(), "token-123", rt)
	if err != nil {
		t.Fatalf("validate returned error: %v", err)
	}
	if info.ID != "42" || info.Email != "user@example.com" || info.Username != "theuser" {
		t.Fatalf("unexpected info: %+v", info)
	}
	if len(info.Raw) == 0 {
		t.Fatalf("expected raw payload to be captured")
	}
}

func TestFetchPrimaryEmail(t *testing.T) {
	restore := withHTTPRequester(t, httpsling.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET method, got %s", req.Method)
		}
		body := `[{"email":"unverified@example.com","primary":false,"verified":false},{"email":"verified@example.com","primary":true,"verified":true}]`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioNopCloser(body),
			Header:     make(http.Header),
		}, nil
	}))
	defer restore()

	ui := &UserInfoSpec{
		AuthStyle:         AuthHeaderStyleBearer,
		SecondaryEmailURL: "https://example.com/emails",
	}
	email, err := fetchPrimaryEmail(context.Background(), "token-xyz", ui)
	if err != nil {
		t.Fatalf("fetchPrimaryEmail returned error: %v", err)
	}
	if email != "verified@example.com" {
		t.Fatalf("expected verified email, got %s", email)
	}
}

func ioNopCloser(body string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(body))
}
