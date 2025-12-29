//go:build examples

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"

	"github.com/theopenlane/iam/providers/github"
	"github.com/theopenlane/iam/sessions"
)

const (
	sessionName     = "example-github-app"
	sessionSecret   = "example cookie signing secret"
	sessionUserKey  = "githubID"
	sessionUsername = "githubUsername"
)

// sessionStore encodes and decodes session data stored in signed cookies
var sessionStore = sessions.NewCookieStore[any](sessions.DebugCookieConfig, []byte(sessionSecret), nil)

// Config configures the main ServeMux.
type Config struct {
	ClientID     string
	ClientSecret string
}

var DebugOnlyCookieConfig = sessions.CookieConfig{
	Name:     "temporary-cookie",
	Path:     "/",
	MaxAge:   600, // 10 min
	HTTPOnly: true,
	Secure:   false, // allows cookies to be send over HTTP
	SameSite: http.SameSiteLaxMode,
}

// New returns a new ServeMux with app routes.
func New(config *Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", profileHandler)
	mux.HandleFunc("/logout", logoutHandler)
	// 1. Register LoginHandler and CallbackHandler
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  "http://localhost:8000/v1/github/callback",
		Endpoint:     githubOAuth2.Endpoint,
	}
	// state param cookies require HTTPS by default; disable for localhost development
	stateConfig := DebugOnlyCookieConfig

	mux.Handle("/v1/github/login", github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil)))
	mux.Handle("/v1/github/callback", github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, issueSession(), nil)))
	return mux
}

// issueSession issues a cookie session after successful GitHub login
func issueSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		githubUser, err := github.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 2. Implement a success handler to issue some form of session
		session := sessionStore.New(sessionName)
		session.Set(sessionUserKey, *githubUser.ID)
		session.Set(sessionUsername, *githubUser.Login)
		if err := session.Save(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, req, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// profileHandler shows a personal profile or a login button (unauthenticated).
func profileHandler(w http.ResponseWriter, req *http.Request) {
	session, err := sessionStore.Get(req, sessionName)
	if err != nil {
		// welcome with login button
		page, _ := os.ReadFile("home.html")
		fmt.Fprint(w, string(page))
		return
	}
	// authenticated profile
	fmt.Fprintf(w, `<p>You are logged in %s!</p><form action="/logout" method="post"><input type="submit" value="Logout"></form>`, session.Get(sessionUsername))
}

// logoutHandler destroys the session on POSTs and redirects to home.
func logoutHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		sessionStore.Destroy(w, sessionName)
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

// main creates and starts a Server listening.
func main() {
	const address = "localhost:8000"
	// read credentials from environment variables if available
	config := &Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	}
	// allow consumer credential flags to override config fields
	clientID := flag.String("client-id", "", "GitHub Client ID")
	clientSecret := flag.String("client-secret", "", "GitHub Client Secret")
	flag.Parse()
	if *clientID != "" {
		config.ClientID = *clientID
	}
	if *clientSecret != "" {
		config.ClientSecret = *clientSecret
	}
	if config.ClientID == "" {
		log.Fatal("Missing GitHub Client ID")
	}
	if config.ClientSecret == "" {
		log.Fatal("Missing GitHub Client Secret")
	}

	log.Printf("Starting Server listening on %s\n", address)
	err := http.ListenAndServe(address, New(config))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
