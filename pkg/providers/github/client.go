package github

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/go-github/v63/github"
)

// ClientConfig holds the configuration for the GitHub client
type ClientConfig struct {
	BaseURL   *url.URL
	UploadURL *url.URL

	IsEnterprise bool
	IsMock       bool
}

// GitHubInterface defines all necessary methods
// https://godoc.org/github.com/google/go-github/github#NewClient
type GitHubInterface interface {
	NewClient(httpClient *http.Client) GitHubClient
	GetConfig() *ClientConfig
	SetConfig(config *ClientConfig)
}

// GitHubClient defines all necessary methods used by the client
type GitHubClient struct {
	Users githubUserService
}

// githubUserService defines all necessary methods for the User service
type githubUserService interface {
	Get(ctx context.Context, user string) (*github.User, *github.Response, error)
	ListEmails(ctx context.Context, opts *github.ListOptions) ([]*github.UserEmail, *github.Response, error)
}

// GitHubCreator implements GitHubInterface
type GitHubCreator struct {
	Config *ClientConfig
}

// GetConfig returns the current configuration
func (g *GitHubCreator) GetConfig() *ClientConfig {
	return g.Config
}

// SetConfig sets the configuration
func (g *GitHubCreator) SetConfig(config *ClientConfig) {
	g.Config = config
}

// NewClient returns a new GitHubClient
func (g *GitHubCreator) NewClient(httpClient *http.Client) GitHubClient {
	client := github.NewClient(httpClient)

	if g.Config.BaseURL != nil {
		client.BaseURL = g.Config.BaseURL
	}

	if g.Config.UploadURL != nil {
		client.UploadURL = g.Config.UploadURL
	}

	return GitHubClient{
		Users: client.Users,
	}
}

// GitHubMock implements GitHubInterface
type GitHubMock struct {
	Config *ClientConfig
}

// GetConfig returns the current configuration
func (g *GitHubMock) GetConfig() *ClientConfig {
	return g.Config
}

// SetConfig sets the configuration
func (g *GitHubMock) SetConfig(config *ClientConfig) {
	g.Config = config
}

// NewClient returns a new mock GitHubClient
func (g *GitHubMock) NewClient(httpClient *http.Client) GitHubClient {
	return GitHubClient{
		Users: &UsersMock{},
	}
}

// UsersMock mocks UsersService
type UsersMock struct {
	githubUserService
}

// Get returns a Github user
func (u *UsersMock) Get(context.Context, string) (*github.User, *github.Response, error) {
	resp := &http.Response{StatusCode: http.StatusOK}

	return &github.User{
		Login: github.String("antman"),
		ID:    github.Int64(1),
	}, &github.Response{Response: resp}, nil
}

// ListEmails returns a mock list of Github user emails
func (u *UsersMock) ListEmails(ctx context.Context, opts *github.ListOptions) ([]*github.UserEmail, *github.Response, error) {
	resp := &http.Response{StatusCode: http.StatusOK}

	return []*github.UserEmail{
		{
			Email:   github.String("antman@theopenlane.io"),
			Primary: github.Bool(true),
		},
		{
			Email:   github.String("ant-man@avengers.com"),
			Primary: github.Bool(false),
		},
	}, &github.Response{Response: resp}, nil
}
