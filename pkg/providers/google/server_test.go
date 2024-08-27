package google

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/theopenlane/core/pkg/testutils"
)

// newGoogleTestServer creates a httptest.Server which mocks the Google
// Userinfo endpoint and a client to holler
func newGoogleTestServer(jsonData string) (*http.Client, *httptest.Server) {
	client, mux, server := testutils.TestServer()
	mux.HandleFunc("/oauth2/v2/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jsonData)
	})

	return client, server
}
