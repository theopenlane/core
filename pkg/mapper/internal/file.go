package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func safeFileName(input string) string {
	output := input
	removeCharacters := []string{".", "_", " ", "(", ")"}

	for _, c := range removeCharacters {
		output = strings.ReplaceAll(output, c, "")
	}

	return strings.ToLower(output)
}

type MetadataAttributes struct {
	Framework string `json:"framework"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
}

type ControlMetadata struct {
	Attributes MetadataAttributes `json:"metadataAttributes"`
}

func generateMetadata(filename, framework, controlID, controlTitle, controlSummary string) error {
	attributes := MetadataAttributes{
		Framework: framework,
		ID:        controlID,
		Title:     controlTitle,
		Summary:   controlSummary,
	}

	jsonData, err := json.MarshalIndent(ControlMetadata{attributes}, "", "    ")
	if err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s.metadata.json", filename))
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}

func fetchFile(fetchLink string, filename string) error {
	parsedURL, err := url.ParseRequestURI(fetchLink)
	if err != nil {
		return fmt.Errorf("invalid ISO link: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get link: %w", err)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = fmt.Errorf("failed to close response body: %w", cerr)
		}
	}()

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	defer func() {
		if cerr := out.Close(); cerr != nil {
			err = fmt.Errorf("failed to close file: %w", cerr)
		}
	}()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to copy response body to file: %w", err)
	}

	return nil
}
