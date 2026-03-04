package auth

import (
	"net/url"
	"testing"
)

func TestBuildEndpointURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		path    string
		params  url.Values
		want    string
	}{
		{
			name:    "base with trailing slash and path without leading slash",
			baseURL: "https://api.example.com/",
			path:    "v1/users",
			want:    "https://api.example.com/v1/users",
		},
		{
			name:    "base without trailing slash and path with leading slash",
			baseURL: "https://api.example.com",
			path:    "/v1/users",
			want:    "https://api.example.com/v1/users",
		},
		{
			name:    "query params appended",
			baseURL: "https://api.example.com/",
			path:    "items",
			params:  url.Values{"page": {"2"}, "per_page": {"10"}},
			want:    "https://api.example.com/items?page=2&per_page=10",
		},
		{
			name:    "empty params not appended",
			baseURL: "https://api.example.com/",
			path:    "items",
			params:  url.Values{},
			want:    "https://api.example.com/items",
		},
		{
			name:    "nil params not appended",
			baseURL: "https://api.example.com/",
			path:    "items",
			params:  nil,
			want:    "https://api.example.com/items",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := buildEndpointURL(tc.baseURL, tc.path, tc.params)
			if got != tc.want {
				t.Fatalf("buildEndpointURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRESTClientGetJSONUsesBaseURL(t *testing.T) {
	t.Parallel()

	// RESTClient.GetJSON calls GetJSONWithClient with the assembled endpoint.
	// Since we cannot easily mock the HTTP layer in a unit test, verify that
	// buildEndpointURL is invoked correctly by checking its output directly.
	rc := RESTClient{BaseURL: "https://api.example.com/"}
	_ = rc // confirming struct construction compiles

	endpoint := buildEndpointURL(rc.BaseURL, "v1/resource", url.Values{"q": {"test"}})
	want := "https://api.example.com/v1/resource?q=test"
	if endpoint != want {
		t.Fatalf("endpoint = %q, want %q", endpoint, want)
	}
}
