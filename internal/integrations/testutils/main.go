//go:build examples

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"

	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	scimContentType = "application/scim+json"
)

type config struct {
	baseURL     string
	bearerToken string
	directory   string
}

type sender struct {
	requester   *httpsling.Requester
	baseURL     string
	bearerToken string
}

type scimResource struct {
	ID string `json:"id"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := parseConfig()
	if err != nil {
		return err
	}

	fixtureDir, err := resolveFixtureDirectory(cfg.directory)
	if err != nil {
		return err
	}

	requester, err := httpsling.New(
		httpsling.Client(httpclient.Timeout(15 * time.Second)),
	)
	if err != nil {
		return fmt.Errorf("create requester: %w", err)
	}

	client := sender{
		requester:   requester,
		baseURL:     cfg.baseURL,
		bearerToken: strings.TrimSpace(cfg.bearerToken),
	}

	ctx := context.Background()

	userPayload, err := loadTemplatePayload(filepath.Join(fixtureDir, "user.json"), nil)
	if err != nil {
		return fmt.Errorf("load user fixture: %w", err)
	}

	userResponse, err := client.post(ctx, "/Users", userPayload)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	var createdUser scimResource
	if err := jsonx.RoundTrip(userResponse, &createdUser); err != nil {
		return fmt.Errorf("decode created user: %w", err)
	}

	fmt.Printf("created user: %s\n", createdUser.ID)

	groupPayload, err := loadTemplatePayload(filepath.Join(fixtureDir, "group.json"), map[string]any{
		"UserID": createdUser.ID,
	})
	if err != nil {
		return fmt.Errorf("load group fixture: %w", err)
	}

	groupResponse, err := client.post(ctx, "/Groups", groupPayload)
	if err != nil {
		return fmt.Errorf("create group: %w", err)
	}

	var createdGroup scimResource
	if err := jsonx.RoundTrip(groupResponse, &createdGroup); err != nil {
		return fmt.Errorf("decode created group: %w", err)
	}

	fmt.Printf("created group: %s\n", createdGroup.ID)

	return nil
}

func parseConfig() (config, error) {
	var cfg config

	flag.StringVar(&cfg.baseURL, "url", "", "exact SCIM base URL ending in /v2")
	flag.StringVar(&cfg.bearerToken, "bearer-token", "", "SCIM bearer token")
	flag.StringVar(&cfg.directory, "directory", "directory_1", "fixture directory name under data/")
	flag.Parse()

	cfg.baseURL = strings.TrimSpace(cfg.baseURL)
	cfg.bearerToken = strings.TrimSpace(cfg.bearerToken)
	cfg.directory = strings.TrimSpace(cfg.directory)

	switch {
	case cfg.baseURL == "":
		return config{}, errors.New("--url is required")
	case cfg.bearerToken == "":
		return config{}, errors.New("--bearer-token is required")
	case cfg.directory == "":
		return config{}, errors.New("--directory is required")
	}

	return cfg, nil
}

func resolveFixtureDirectory(directory string) (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("resolve fixture directory")
	}

	return filepath.Join(filepath.Dir(filename), "data", directory), nil
}

func loadTemplatePayload(path string, data any) (json.RawMessage, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tpl, err := template.New(filepath.Base(path)).Option("missingkey=error").Parse(string(content))
	if err != nil {
		return nil, err
	}

	var rendered strings.Builder
	if err := tpl.Execute(&rendered, data); err != nil {
		return nil, err
	}

	return json.RawMessage(rendered.String()), nil
}

func (s sender) post(ctx context.Context, path string, payload json.RawMessage) (json.RawMessage, error) {
	targetURL := s.baseURL + path

	resp, err := s.requester.ReceiveWithContext(ctx, nil,
		httpsling.Post(targetURL),
		httpsling.Body(payload),
		httpsling.BearerAuth(s.bearerToken),
		httpsling.ContentType(scimContentType),
		httpsling.Accept(scimContentType),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if !httpsling.IsSuccess(resp) {
		return nil, fmt.Errorf("%s returned %d: %s", targetURL, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return json.RawMessage(body), nil
}
