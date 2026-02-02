package shortlinks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
)

const (
	// defaultEndpointURL is the hosted shortlink service API endpoint
	defaultEndpointURL = "https://admin.s.theopenlane.io/api/links"
	// defaultRequestTimeout is the default timeout for requests to the shortlink service
	defaultRequestTimeout = 10 * time.Second
	// headerAccessClientID is the header for the access client ID
	headerAccessClientID = "Cf-Access-Client-Id"
	// headerAccessClientSecret is the header for the access client secret
	headerAccessClientSecret = "Cf-Access-Client-Secret"
)

// createLinkRequest represents the payload for creating a shortlink
type createLinkRequest struct {
	// URL is the original URL to shorten (the target URL)
	URL string `json:"url"`
	// Slug is an optional custom slug for the shortlink
	Slug string `json:"slug,omitempty"`
}

// responseError represents an error response from the shortlinks API
type responseError struct {
	// StatusCode is that 302'ing up your biz
	StatusCode int
	// Status is the HTTP status text
	Status string
}

// Error formats a shortlinks response error
func (e *responseError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("shortlinks: request failed (%s)", e.Status)
}

// Create issues a POST to the hosted shortlink API and returns the short URL
func Create(ctx context.Context, clientID, clientSecret, url, slug string) (string, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" || clientSecret == "" {
		return "", ErrMissingAuthenticationParams
	}

	url = strings.TrimSpace(url)
	if url == "" {
		return "", ErrMissingURL
	}

	payload := createLinkRequest{
		URL:  url,
		Slug: strings.TrimSpace(slug),
	}

	opts := []httpsling.Option{
		httpsling.Post(defaultEndpointURL),
		httpsling.Body(payload),
		httpsling.ContentType(httpsling.ContentTypeJSON),
		httpsling.Accept(httpsling.ContentTypeJSON),
		httpsling.Header(headerAccessClientID, clientID),
		httpsling.Header(headerAccessClientSecret, clientSecret),
		httpsling.Client(httpclient.Timeout(defaultRequestTimeout)),
	}

	resp, err := httpsling.ReceiveWithContext(ctx, nil, opts...)
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", ErrEmptyResponse
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read shortlink response: %w", err)
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return "", &responseError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}

	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return "", ErrEmptyResponseBody
	}

	var response struct {
		ShortURL      string `json:"shortUrl"`
		ShortURLSnake string `json:"short_url"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("decode shortlink response: %w", err)
	}

	shortURL := strings.TrimSpace(response.ShortURL)
	if shortURL == "" {
		return "", ErrMissingShortURL
	}

	return shortURL, nil
}
