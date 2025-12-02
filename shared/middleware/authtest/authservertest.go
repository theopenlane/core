package authtest

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/theopenlane/iam/tokens"
)

const (
	Audience = "http://localhost:17608"
	Issuer   = "http://localhost:17608"
)

// Server implements an endpoint to host JWKS public keys and also provides simple
// functionality to create access and refresh tokens that would be authenticated.
type Server struct {
	srv    *httptest.Server
	mux    *http.ServeMux
	tokens *tokens.TokenManager
	URL    *url.URL
}

// NewServer starts and returns a new authtest server. The caller should call Close
// when finished, to shut it down.
func NewServer() (s *Server, err error) {
	// Setup routes for the mux
	s = &Server{}
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/.well-known/jwks.json", s.JWKS)

	// Setup httptest Server
	s.srv = httptest.NewServer(s.mux)
	s.URL, _ = url.Parse("http://localhost:17608/.well-known/jwks.json")

	// Create token manager
	keys := map[string]string{}

	// Checks for the file in the root of this repo
	privFileName := "../../../../private_key.pem"

	conf := tokens.Config{
		Audience:        Audience,
		Issuer:          Issuer,
		AccessDuration:  1 * time.Hour,
		RefreshDuration: 2 * time.Hour, //nolint:mnd
		RefreshOverlap:  -15 * time.Minute,
	}

	// if the file isn't there, generate with a new key
	if _, err := os.Stat(privFileName); err != nil {
		var key ed25519.PrivateKey

		if _, key, err = ed25519.GenerateKey(rand.Reader); err != nil {
			return nil, err
		}

		if s.tokens, err = tokens.NewWithKey(key, conf); err != nil {
			return nil, err
		}
	} else {
		// This is the same KID that the task file uses, so your server and generated
		// tokens will have same ID
		keys["01HHAS67AM73778S0QEZ3CEAGE"] = fmt.Sprintf("%v", privFileName)
		conf.Keys = keys

		if s.tokens, err = tokens.New(conf); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Server) Close() {
	s.srv.Close()
}

func (s *Server) JWKS(w http.ResponseWriter, _ *http.Request) {
	keys, err := s.tokens.Keys()
	if err != nil {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error())) // nolint: errcheck

		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(keys) // nolint: errcheck
}

func (s *Server) KeysURL() string {
	return s.URL.ResolveReference(&url.URL{Path: "/.well-known/jwks.json"}).String()
}

// CreateToken creates a token without overwriting the claims, which is useful for
// creating tokens with specific not before and expiration times for testing.
func (s *Server) CreateToken(claims *tokens.Claims) (tks string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return s.tokens.Sign(token)
}

func (s *Server) CreateAccessToken(claims *tokens.Claims) (tks string, err error) {
	var token *jwt.Token

	if token, err = s.tokens.CreateAccessToken(claims); err != nil {
		return "", err
	}

	return s.tokens.Sign(token)
}

func (s *Server) CreateTokenPair(claims *tokens.Claims) (accessToken, refreshToken string, err error) {
	return s.tokens.CreateTokenPair(claims)
}
