package httpsling

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// startFileUploadServer starts a mock server to test file uploads
func startFileUploadServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20) // Limit: 10MB
		if err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)

			return
		}

		// Collect file upload details
		uploads := make(map[string][]string)

		for key, files := range r.MultipartForm.File {
			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					http.Error(w, "Failed to open file", http.StatusInternalServerError)
					return
				}

				defer file.Close() //nolint: errcheck

				// Read file content (for demonstration; in real tests, might hash or skip)
				content, err := io.ReadAll(file)
				if err != nil {
					http.Error(w, "Failed to read file content", http.StatusInternalServerError)

					return
				}

				// Store file details (e.g., filename and a snippet of content for verification)
				contentSnippet := string(content)
				if len(contentSnippet) > 10 {
					contentSnippet = contentSnippet[:10] + "..."
				}

				uploads[key] = append(uploads[key], fmt.Sprintf("%s: %s", fileHeader.Filename, contentSnippet))
			}
		}

		// Respond with details of the uploaded files in JSON format
		w.Header().Set(HeaderContentType, ContentTypeJSON)

		if encoder := json.NewEncoder(w); encoder != nil {
			if err = encoder.Encode(uploads); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Failed to create JSON encoder", http.StatusInternalServerError)
		}
	}))
}

func TestFiles(t *testing.T) {
	server := startFileUploadServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	fileContent1 := strings.NewReader("File content 1")
	fileContent2 := strings.NewReader("File content 2")

	resp, err := client.Post("/").
		Files(
			&File{Name: "file1", FileName: "test1.txt", Content: io.NopCloser(fileContent1)},
			&File{Name: "file2", FileName: "test2.txt", Content: io.NopCloser(fileContent2)},
		).
		Send(context.Background())

	assert.NoError(t, err, "No error expected on sending request")

	var uploads map[string][]string
	err = resp.ScanJSON(&uploads)
	assert.NoError(t, err, "Expect no error on parsing response")

	// Validate the file uploads
	assert.Contains(t, uploads, "file1", "file1 should be present in the uploads")
	assert.Contains(t, uploads, "file2", "file2 should be present in the uploads")
}
func TestFile(t *testing.T) {
	server := startFileUploadServer() // Start the mock file upload server
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Simulate a file's content
	fileContent := strings.NewReader("This is the file content")

	// Send a request with a single file
	resp, err := client.Post("/").
		File("file", "single.txt", io.NopCloser(fileContent)).
		Send(context.Background())
	assert.NoError(t, err, "No error expected on sending request")

	// Parse the server's JSON response
	var uploads map[string][]string
	err = resp.ScanJSON(&uploads)
	assert.NoError(t, err, "Expect no error on parsing response")

	// Check if the server received the file correctly
	assert.Contains(t, uploads, "file", "The file should be present in the uploads")
	assert.Contains(t, uploads["file"][0], "single.txt", "The file name should be correctly received")
}

func TestDelFile(t *testing.T) {
	server := startFileUploadServer() // Start the mock file upload server

	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Simulate file contents
	fileContent1 := strings.NewReader("File content 1")
	fileContent2 := strings.NewReader("File content 2")

	// Prepare the request with two files, then delete one before sending
	resp, err := client.Post("/").
		Files(
			&File{Name: "file1", FileName: "file1.txt", Content: io.NopCloser(fileContent1)},
			&File{Name: "file2", FileName: "file2.txt", Content: io.NopCloser(fileContent2)},
		).
		DelFile("file1"). // Remove the first file
		Send(context.Background())
	assert.NoError(t, err, "No error expected on sending request")

	// Parse the server's JSON response
	var uploads map[string][]string

	err = resp.ScanJSON(&uploads)

	assert.NoError(t, err, "Expect no error on parsing response")

	// Validate that only the second file was uploaded
	assert.NotContains(t, uploads, "file1", "file1 should have been removed from the uploads")
	assert.Contains(t, uploads, "file2", "file2 should be present in the uploads")
}
