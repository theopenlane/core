package oauth2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	contentType     = "Content-Type"
	jsonContentType = "application/json"
)

// NewAccessTokenServer creates a httptest.Server OAuth2 provider Access Token endpoint
func NewAccessTokenServer(t *testing.T, json string) *httptest.Server {
	return NewTestServerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		w.Header().Set(contentType, jsonContentType)
		_, _ = w.Write([]byte(json))
	})
}

// NewTestServeFunc wraps httptest.Server so it can be used with funcs
func NewTestServerFunc(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}
