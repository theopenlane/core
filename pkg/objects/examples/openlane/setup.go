//go:build examples

package openlane

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

type registerResponse struct {
	Token string `json:"token"`
}

// RegisterUser registers a new user with the given email and password
func RegisterUser(ctx context.Context, baseURL *url.URL, email, password, firstName, lastName string) (string, error) {
	registerURL := baseURL.String() + "/v1/register"

	payload := map[string]string{
		"email":     email,
		"password":  password,
		"firstName": firstName,
		"lastName":  lastName,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal registration payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", registerURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send registration request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	var result registerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode registration response: %w", err)
	}

	return result.Token, nil
}

// VerifyUser verifies a user's email address using a verification token
func VerifyUser(ctx context.Context, baseURL *url.URL, token string) error {
	verifyURL := fmt.Sprintf("%s/v1/verify?token=%s", baseURL.String(), token)

	req, err := http.NewRequestWithContext(ctx, "GET", verifyURL, nil)
	if err != nil {
		return fmt.Errorf("create verify request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send verify request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("verification failed with status %d", resp.StatusCode)
	}

	return nil
}

// LoginUser logs in a user and returns an authenticated Openlane client
func LoginUser(ctx context.Context, baseURL *url.URL, email, password string) (*openlaneclient.OpenlaneClient, error) {
	config := openlaneclient.NewDefaultConfig()

	client, err := openlaneclient.New(config, openlaneclient.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("create initial client: %w", err)
	}

	loginInput := models.LoginRequest{
		Username: email,
		Password: password,
	}

	resp, err := client.Login(ctx, &loginInput)
	if err != nil {
		return nil, fmt.Errorf("login request: %w", err)
	}

	session, err := client.GetSessionFromCookieJar()
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	authClient, err := openlaneclient.New(
		config,
		openlaneclient.WithBaseURL(baseURL),
		openlaneclient.WithCredentials(openlaneclient.Authorization{
			BearerToken: resp.AccessToken,
			Session:     session,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create authenticated client: %w", err)
	}

	return authClient, nil
}

// CreateOrganization creates a new organization
func CreateOrganization(ctx context.Context, client *openlaneclient.OpenlaneClient, name, description string) (string, error) {
	input := openlaneclient.CreateOrganizationInput{
		Name:        name,
		Description: &description,
	}

	org, err := client.CreateOrganization(ctx, input, nil)
	if err != nil {
		return "", fmt.Errorf("create organization: %w", err)
	}

	return org.CreateOrganization.Organization.ID, nil
}

// GetOrganizationID retrieves the first organization ID for the current user
func GetOrganizationID(ctx context.Context, client *openlaneclient.OpenlaneClient) (string, error) {
	orgs, err := client.GetOrganizations(ctx, nil, nil, nil)
	if err != nil {
		return "", fmt.Errorf("get organizations: %w", err)
	}

	if len(orgs.Organizations.Edges) == 0 {
		return "", fmt.Errorf("no organizations found")
	}

	return orgs.Organizations.Edges[0].Node.ID, nil
}

// CreatePAT creates a personal access token for the specified organization
func CreatePAT(ctx context.Context, client *openlaneclient.OpenlaneClient, orgID, name, description string) (string, error) {
	input := openlaneclient.CreatePersonalAccessTokenInput{
		Name:            name,
		Description:     &description,
		OrganizationIDs: []string{orgID},
	}

	pat, err := client.CreatePersonalAccessToken(ctx, input)
	if err != nil {
		return "", fmt.Errorf("create personal access token: %w", err)
	}

	return pat.CreatePersonalAccessToken.PersonalAccessToken.Token, nil
}

// InitializeClient creates an Openlane client with the given token and optional organization ID
func InitializeClient(baseURL *url.URL, token, orgID string) (*openlaneclient.OpenlaneClient, error) {
	config := openlaneclient.NewDefaultConfig()

	if orgID != "" {
		config.Interceptors = append(config.Interceptors, openlaneclient.WithOrganizationHeader(orgID))
	}

	client, err := openlaneclient.New(
		config,
		openlaneclient.WithBaseURL(baseURL),
		openlaneclient.WithCredentials(openlaneclient.Authorization{
			BearerToken: token,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	return client, nil
}

// CreateUpload creates a graphql upload from a file path
func CreateUpload(filePath string) (*graphql.Upload, error) {
	uploadFile, err := storage.NewUploadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("create upload file: %w", err)
	}

	return &graphql.Upload{
		File:        uploadFile.RawFile,
		Filename:    uploadFile.OriginalName,
		Size:        uploadFile.Size,
		ContentType: uploadFile.ContentType,
	}, nil
}

// CreateEvidenceWithFile creates evidence with an attached file
func CreateEvidenceWithFile(ctx context.Context, client *openlaneclient.OpenlaneClient, name, desc string, upload *graphql.Upload) (*openlaneclient.CreateEvidence, error) {
	status := enums.EvidenceStatusSubmitted

	input := openlaneclient.CreateEvidenceInput{
		Name:        name,
		Description: &desc,
		Status:      &status,
	}

	result, err := client.CreateEvidence(ctx, input, []*graphql.Upload{upload})
	if err != nil {
		return nil, fmt.Errorf("create evidence: %w", err)
	}

	return result, nil
}
