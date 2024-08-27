package httpsling

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequestCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate a long-running operation
		fmt.Fprintln(w, "This response may never be sent")
	}))

	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel() // Ensure resources are cleaned up

	// Cancel the request after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	// Attempt to make a request that will be canceled
	_, err := client.Get("/").Send(ctx)
	if err == nil {
		t.Errorf("Expected an error due to cancellation, but got none")
	}

	// Check if the error is due to context cancellation
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

// TestSendMethodQuery checks the Send method for handling query parameters.
func TestSendMethodQuery(t *testing.T) {
	// Start a test HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with the full URL received, including query parameters.
		fmt.Fprintln(w, r.URL.String())
	}))
	defer server.Close()

	// Define a client with the test server's URL.
	client := Create(&Config{BaseURL: server.URL})

	tests := []struct {
		name          string
		url           string            // URL to request, may include query params
		additionalQPS map[string]string // Query params added via Query method
		expectedURL   string            // Expected URL path and query received by the server
	}{
		{
			name:        "URL only",
			url:         "/test?param1=value1",
			expectedURL: "/test?param1=value1",
		},
		{
			name:          "Method only",
			url:           "/test",
			additionalQPS: map[string]string{"param2": "value2"},
			expectedURL:   "/test?param2=value2",
		},
		{
			name:          "URL and Method",
			url:           "/test?param1=value1",
			additionalQPS: map[string]string{"param2": "value2"},
			expectedURL:   "/test?param1=value1&param2=value2",
		},
		{
			name:          "Method overwrites URL",
			url:           "/test?param1=value1",
			additionalQPS: map[string]string{"param1": "value2"},
			expectedURL:   "/test?param1=value2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new RequestBuilder for each test case.
			rb := client.NewRequestBuilder(http.MethodGet, tc.url)

			// If there are additional query params defined, add them.
			if tc.additionalQPS != nil {
				for key, value := range tc.additionalQPS {
					rb.Queries(map[string][]string{key: {value}})
				}
			}

			// Send the request.
			resp, err := rb.Send(context.Background())
			assert.NoError(t, err)

			// Read the response body.
			bodyBytes, err := io.ReadAll(resp.RawResponse.Body)
			assert.NoError(t, err)

			body := string(bodyBytes)

			// The body should contain the expected URL path and query.
			assert.Contains(t, body, tc.expectedURL, "The server did not receive the expected URL.")
		})
	}
}

type testAddress struct {
	Postcode string `url:"postcode"`
	City     string `url:"city"`
}

type testQueryStruct struct {
	Name       string      `url:"name"`
	Occupation string      `url:"occupation,omitempty"`
	Age        int         `url:"age"`
	IsActive   bool        `url:"is_active,int"`
	Tags       []string    `url:"tags,comma"`
	Address    testAddress `url:"addr"`
}

func TestQueryStructWithClient(t *testing.T) {
	// Start a test HTTP server that JSON-encodes and echoes back the query parameters received
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()

		w.Header().Set(HeaderContentType, ContentTypeJSON)

		if encoder := json.NewEncoder(w); encoder != nil {
			if err := encoder.Encode(queryParams); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Failed to create JSON encoder", http.StatusInternalServerError)
		}
	}))

	defer server.Close()

	// Create an instance of the client, pointing to the test server
	client := Create(&Config{BaseURL: server.URL})

	// Define the struct to be used for query parameters
	exampleStruct := testQueryStruct{
		Name:       "John Snow",
		Occupation: "Developer",
		Age:        30,
		IsActive:   true,
		Tags:       []string{"go", "programming"},
		Address: testAddress{
			Postcode: "1234",
			City:     "GoCity",
		},
	}

	// Send a request to the server using the client and the struct for query parameters
	resp, err := client.NewRequestBuilder(http.MethodGet, "/").QueriesStruct(exampleStruct).Send(context.Background())
	assert.NoError(t, err)

	// Read and verify the response
	var response map[string][]string
	err = resp.ScanJSON(&response)
	assert.NoError(t, err)

	// Now we can assert the values directly
	assert.Contains(t, response, "name")
	assert.Equal(t, []string{"John Snow"}, response["name"])
	assert.Contains(t, response, "occupation")
	assert.Equal(t, []string{"Developer"}, response["occupation"])
	assert.Contains(t, response, "age")
	assert.Equal(t, []string{"30"}, response["age"])
	assert.Contains(t, response, "is_active")
	assert.Equal(t, []string{"1"}, response["is_active"])
	assert.Contains(t, response, "tags")
	assert.Equal(t, []string{"go,programming"}, response["tags"])

	err = resp.Close()
	assert.NoError(t, err)
}

func TestHeaderManipulationMethods(t *testing.T) {
	// Start a test HTTP server that checks received headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		assert.Equal(t, ContentTypeJSON, r.Header.Get(HeaderContentType))
		assert.Equal(t, "Bearer token", r.Header.Get(HeaderAuthorization))
		assert.Empty(t, r.Header.Get("X-Deprecated-Header"))

		fmt.Fprintln(w, "Headers received")
	}))

	defer server.Close()

	// Create an instance of the client, pointing to the test server
	rq := Create(&Config{BaseURL: server.URL}).Get("/test-headers")
	rq.Headers(http.Header{HeaderContentType: []string{ContentTypeJSON}})
	rq.AddHeader(HeaderAuthorization, "Bearer token")
	rq.Header("X-Modified-Header", "NewValue")
	rq.AddHeader("X-Deprecated-Header", "OldValue")
	rq.DelHeader("X-Deprecated-Header")

	// Send the request
	resp, err := rq.Send(context.Background())
	assert.NoError(t, err)

	// Read and verify the response
	responseBody, err := io.ReadAll(resp.RawResponse.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(responseBody), "Headers received")
}

func TestUserAgentMethod(t *testing.T) {
	// Start a test HTTP server that checks received User-Agent header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check User-Agent header
		assert.Equal(t, "MyCustomUserAgent", r.Header.Get(HeaderUserAgent))

		fmt.Fprintln(w, "User-Agent received")
	}))
	defer server.Close()

	// Create an instance of the client, pointing to the test server
	client := Create(&Config{BaseURL: server.URL})
	rq := client.Get("/test-user-agent")

	// Set the User-Agent header using the UserAgent method
	rq.UserAgent("MyCustomUserAgent")

	// Send the request
	resp, err := rq.Send(context.Background())
	assert.NoError(t, err)

	// Read and verify the response
	responseBody, err := io.ReadAll(resp.RawResponse.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(responseBody), "User-Agent received")
}

func TestContentTypeMethod(t *testing.T) {
	// Start a test HTTP server that checks received Content-Type header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Content-Type header
		assert.Equal(t, ContentTypeJSON, r.Header.Get(HeaderContentType))

		fmt.Fprintln(w, "Content-Type received")
	}))

	defer server.Close()

	// Create an instance of the client, pointing to the test server
	client := Create(&Config{BaseURL: server.URL})
	rq := client.Get("/test-content-type")

	// Set the Content-Type header using the ContentType method
	rq.ContentType(ContentTypeJSON)

	// Send the request
	resp, err := rq.Send(context.Background())
	assert.NoError(t, err)

	// Read and verify the response
	responseBody, err := io.ReadAll(resp.RawResponse.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(responseBody), "Content-Type received")
}

func TestAcceptMethod(t *testing.T) {
	// Start a test HTTP server that checks received Accept header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Accept header
		assert.Equal(t, ContentTypeXML, r.Header.Get(HeaderAccept))

		fmt.Fprintln(w, "Accept received")
	}))

	defer server.Close()

	// Create an instance of the client, pointing to the test server
	client := Create(&Config{BaseURL: server.URL})
	rq := client.Get("/test-accept")

	// Set the Accept header using the Accept method
	rq.Accept(ContentTypeXML)

	// Send the request
	resp, err := rq.Send(context.Background())
	assert.NoError(t, err)

	// Read and verify the response
	responseBody, err := io.ReadAll(resp.RawResponse.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(responseBody), "Accept received")
}

func TestRefererMethod(t *testing.T) {
	// Start a test HTTP server that checks received Referer header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Referer header
		assert.Equal(t, "https://example.com", r.Header.Get(HeaderReferer))

		fmt.Fprintln(w, "Referer received")
	}))

	defer server.Close()

	// Create an instance of the client, pointing to the test server
	client := Create(&Config{BaseURL: server.URL})
	rq := client.Get("/test-referer")

	// Set the Referer header
	rq.Referer("https://example.com")

	// Send the request
	resp, err := rq.Send(context.Background())
	assert.NoError(t, err)

	// Read and verify the response
	responseBody, err := io.ReadAll(resp.RawResponse.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(responseBody), "Referer received")
}

func TestCookieManipulationMethods(t *testing.T) {
	// Start a test HTTP server that checks received cookies
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check cookies
		cookie1, err1 := r.Cookie("SessionID")
		assert.NoError(t, err1)
		assert.Equal(t, "12345", cookie1.Value)

		cookie2, err2 := r.Cookie("AuthToken")
		assert.NoError(t, err2)
		assert.Equal(t, "abcdef", cookie2.Value)

		// Ensure the deleted cookie is not present
		_, err3 := r.Cookie("DeletedCookie")
		assert.Error(t, err3) // We expect an error because the cookie should not be present

		fmt.Fprintln(w, "Cookies received")
	}))

	defer server.Close()

	// Create an instance of the client, pointing to the test server
	rq := Create(&Config{BaseURL: server.URL}).Get("/test-cookies")
	// Using SetCookies to set multiple cookies at once
	rq.Cookies(map[string]string{
		"SessionID":     "12345",
		"AuthToken":     "abcdef",
		"DeletedCookie": "should-be-deleted",
	})
	// Demonstrate individual cookie manipulation
	rq.Cookie("SingleCookie", "single-value")
	// Removing a previously set cookie
	rq.DelCookie("DeletedCookie")

	// Send the request
	resp, err := rq.Send(context.Background())
	assert.NoError(t, err)

	// Read and verify the response
	responseBody, err := io.ReadAll(resp.RawResponse.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(responseBody), "Cookies received")
}

func TestPathParameterMethods(t *testing.T) {
	// Start a test HTTP server that checks the received path for correctness
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path is as expected
		expectedPath := "/users/johnDoe/posts/123"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		fmt.Fprintln(w, "Path parameters received correctly")
	}))

	defer server.Close()

	// Create an instance of the client, pointing to the test server
	client := Create(&Config{BaseURL: server.URL})
	rq := client.Get("/users/{userId}/posts/{postId}")

	// Using PathParams to set multiple path params at once
	rq.PathParams(map[string]string{
		"postId": "123",
	})

	// Demonstrate individual path parameter manipulation
	rq.PathParam("userId", "johnDoe").PathParam("hello", "world")
	rq.DelPathParam("hello")

	// Send the request
	resp, err := rq.Send(context.Background())
	assert.NoError(t, err)

	// Read and verify the response
	responseBody, err := io.ReadAll(resp.RawResponse.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(responseBody), "Path parameters received correctly")
}

func startEchoServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)

		w.Header().Set(HeaderContentType, ContentTypeJSON)

		response := map[string]string{
			"body":        string(bodyBytes),
			"contentType": r.Header.Get(HeaderContentType),
		}

		if encoder := json.NewEncoder(w); encoder != nil {
			if err := encoder.Encode(response); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Failed to create JSON encoder", http.StatusInternalServerError)
		}
	}))
}

func TestFormFields(t *testing.T) {
	server := startEchoServer() // Starts a mock HTTP server that echoes back received requests
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Example form data using a map
	formData := map[string]string{
		"name": "Jane Doe",
		"age":  "32",
	}

	resp, err := client.Post("/").
		FormFields(formData). // Using FormFields to set form data
		Send(context.Background())
	assert.NoError(t, err, "Request should not fail")

	var response map[string]string
	err = resp.Scan(&response)
	assert.NoError(t, err, "Response should be parsed without error")

	// Validates that the form data was correctly encoded and sent in the request body
	expectedEncodedFormData := url.Values{"name": {"Jane Doe"}, "age": {"32"}}.Encode()

	assert.Equal(t, expectedEncodedFormData, response["body"], "The body content should match the encoded form data")
	assert.Equal(t, ContentTypeForm, response["contentType"], "The content type should be application/x-www-form-urlencoded")
}

func TestFormField(t *testing.T) {
	server := startEchoServer() // Simulated HTTP server
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	resp, err := client.Post("/").
		FormField("name", "John Snow"). // Adding a single form field
		Send(context.Background())
	assert.NoError(t, err, "No error expected on sending request")

	var response map[string]string
	err = resp.Scan(&response)
	assert.NoError(t, err, "Parsing response should not error")

	// Validate that the single form field was correctly encoded and sent
	expectedEncodedFormData := "name=John+Snow"
	assert.Equal(t, expectedEncodedFormData, response["body"], "The body content should match the single form field")
}

func TestDelFormField(t *testing.T) {
	server := startEchoServer() // Setup mock server
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Set initial form fields
	initialFormData := map[string]string{
		"name": "Jane Doe",
		"age":  "32",
	}

	// Delete the "age" field before sending
	resp, err := client.Post("/").
		FormFields(initialFormData).
		DelFormField("age"). // Removing an existing form field
		Send(context.Background())
	assert.NoError(t, err, "Expect no error on request send")

	var response map[string]string
	err = resp.Scan(&response)
	assert.NoError(t, err, "Expect no error on response parse")

	// Validates that the "age" field was correctly removed
	expectedEncodedFormData := "name=Jane+Doe"
	assert.Equal(t, expectedEncodedFormData, response["body"], "The body should match after deleting a field")
}

func TestBody(t *testing.T) {
	server := startEchoServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Example body data
	bodyData := url.Values{"key": []string{"value"}}
	encodedData := bodyData.Encode()

	resp, err := client.Post("/").
		Body(bodyData).
		ContentType(ContentTypeForm).
		Send(context.Background())

	assert.NoError(t, err)

	var response map[string]string
	err = resp.Scan(&response)
	assert.NoError(t, err)

	// Asserts
	assert.Equal(t, encodedData, response["body"], "The body content should match.")
	assert.Equal(t, ContentTypeForm, response["contentType"], "The content type should be set correctly.")
}

func TestJSONBody(t *testing.T) {
	server := startEchoServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Example JSON data
	jsonData := map[string]interface{}{"name": "John Snow", "age": 30}
	jsonDataStr, _ := json.Marshal(jsonData)

	resp, err := client.Post("/").
		JSONBody(jsonData).
		Send(context.Background())
	assert.NoError(t, err)

	var response map[string]string
	err = resp.Scan(&response)
	assert.NoError(t, err)

	// Asserts
	assert.JSONEq(t, string(jsonDataStr), response["body"], "The body content should match.")
	assert.Equal(t, ContentTypeJSON, response["contentType"], "The content type should be set to application/json.")
}

func TestXMLBody(t *testing.T) {
	server := startEchoServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Example XML data
	xmlData := struct {
		XMLName xml.Name `xml:"Person"`
		Name    string   `xml:"Name"`
		Age     int      `xml:"Age"`
	}{Name: "Jane Doe", Age: 32}
	xmlDataStr, _ := xml.Marshal(xmlData)

	resp, err := client.Post("/").
		XMLBody(xmlData).
		Send(context.Background())
	assert.NoError(t, err)

	var response map[string]string
	err = resp.Scan(&response)
	assert.NoError(t, err)

	// Asserts
	assert.Equal(t, string(xmlDataStr), strings.TrimSpace(response["body"]), "The body content should match.")
	assert.Equal(t, ContentTypeXML, response["contentType"], "The content type should be set to application/xml.")
}

func TestFormWithUrlValues(t *testing.T) {
	server := startEchoServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Example form data
	formData := url.Values{
		"name": []string{"Jane Doe"},
		"age":  []string{"32"},
	}

	resp, err := client.Post("/").
		Form(formData).
		Send(context.Background())
	assert.NoError(t, err)

	var response map[string]string
	err = resp.Scan(&response)
	assert.NoError(t, err)

	// Asserts
	assert.Equal(t, formData.Encode(), response["body"], "The body content should match.")
	assert.Equal(t, ContentTypeForm, response["contentType"], "The content type should be set correctly.")
}

func TestTextBody(t *testing.T) {
	server := startEchoServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Example text data
	textData := "This is a plain text body."

	resp, err := client.Post("/").
		TextBody(textData).
		Send(context.Background())
	assert.NoError(t, err)

	var response map[string]string
	err = resp.Scan(&response)
	assert.NoError(t, err)

	// Asserts
	assert.Equal(t, textData, response["body"], "The body content should match.")
	assert.Equal(t, ContentTypeText, response["contentType"], "The content type should be set to text/plain.")
}

func TestRawBody(t *testing.T) {
	server := startEchoServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Example raw data
	rawData := []byte("This is raw byte data.")

	resp, err := client.Post("/").
		RawBody(rawData).
		ContentType("application/octet-stream"). // Explicitly set content type
		Send(context.Background())
	assert.NoError(t, err)

	var response map[string]string
	err = resp.Scan(&response)
	assert.NoError(t, err)

	// Asserts
	assert.Equal(t, string(rawData), response["body"], "The body content should match.")
	assert.Equal(t, "application/octet-stream", response["contentType"], "The content type should be set to application/octet-stream.")
}

func TestRequestLevelRetries(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			// Simulate a server error on the first request
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// Succeed on subsequent attempts
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Success")
		}
	}))

	defer server.Close()

	// Set up a request builder with retry configuration
	client := Create(&Config{BaseURL: server.URL})
	rq := client.Get("/")
	rq.MaxRetries(2) // Allow up to 2 retries
	rq.RetryStrategy(func(attempt int) time.Duration { return 10 * time.Millisecond })
	rq.RetryIf(func(req *http.Request, resp *http.Response, err error) bool {
		// Retry on server error
		return resp.StatusCode == http.StatusInternalServerError
	})

	// Send the request
	_, err := rq.Send(context.Background())
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Verify that the retry logic was applied
	expectedAttempts := int32(2)
	if requestCount != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, requestCount)
	}
}

func TestFormWithNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeJSON)
		// Ensure a valid JSON response is sent back for all scenarios
		response := map[string]interface{}{
			"status": "received",
			"body":   "empty or nil form",
		}
		if encoder := json.NewEncoder(w); encoder != nil {
			if err := encoder.Encode(response); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Failed to create JSON encoder", http.StatusInternalServerError)
		}
	}))

	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	resp, err := client.Post("/").Form(nil).Send(context.Background())
	assert.NoError(t, err, "No error expected on sending request with nil form")

	var response map[string]interface{}
	err = resp.ScanJSON(&response)
	assert.NoError(t, err, "Expect no error on parsing response")

	// Assert form is correctly received
	assert.Contains(t, response, "status", "Status should be present")
	assert.Contains(t, response, "body", "Body should be present")
}

func startFormHandlingServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeJSON)

		err := r.ParseMultipartForm(32 << 20) // limit to 32MB
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fields := make(map[string][]string)
		files := make(map[string][]string)

		if r.MultipartForm != nil {
			for key, values := range r.MultipartForm.Value {
				fields[key] = values
			}

			for key, fileHeaders := range r.MultipartForm.File {
				for _, fileHeader := range fileHeaders {
					files[key] = append(files[key], fileHeader.Filename)
				}
			}
		}

		response := map[string]interface{}{
			"fields": fields,
			"files":  files,
		}

		if encoder := json.NewEncoder(w); encoder != nil {
			if err := encoder.Encode(response); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Failed to create JSON encoder", http.StatusInternalServerError)
		}
	}))
}

func TestFormWithFiles(t *testing.T) {
	server := startFormHandlingServer()

	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	fileContent1 := strings.NewReader("File content 1")
	fileContent2 := strings.NewReader("File content 2")

	formData := map[string]any{
		"file1": &File{Name: "file1", FileName: "file1.txt", Content: io.NopCloser(fileContent1)},
		"file2": &File{Name: "file2", FileName: "file2.txt", Content: io.NopCloser(fileContent2)},
	}

	resp, err := client.Post("/").Form(formData).Send(context.Background())
	assert.NoError(t, err, "No error expected on sending request with files")

	var response map[string]interface{}
	err = resp.ScanJSON(&response)
	assert.NoError(t, err, "Expect no error on parsing response")

	// Assert files are correctly received
	assert.Contains(t, response["files"].(map[string]interface{}), "file1", "File1 should be present")
	assert.Contains(t, response["files"].(map[string]interface{}), "file2", "File2 should be present")
}

func TestFormWithMixedFilesAndFields(t *testing.T) {
	server := startFormHandlingServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	fileContent := strings.NewReader("File content 1")

	formData := map[string]any{
		"name": "John Snow",
		"age":  "30",
		"file": &File{Name: "file", FileName: "file.txt", Content: io.NopCloser(fileContent)},
	}

	resp, err := client.Post("/").Form(formData).Send(context.Background())
	assert.NoError(t, err, "No error expected on sending request with mixed form data")

	var response map[string]interface{}
	err = resp.Scan(&response)
	assert.NoError(t, err, "Expect no error on parsing response")

	// Assert fields and files are correctly received
	fields := response["fields"].(map[string]interface{})
	assert.Contains(t, fields, "name", "Name should be present")
	assert.Contains(t, fields, "age", "Age should be present")

	files := response["files"].(map[string]interface{})
	assert.Contains(t, files, "file", "File should be present")
}

// TestAuthRequest verifies that the Auth method correctly applies basic authentication to a request
func TestAuthRequest(t *testing.T) {
	// Expected username and password for basic authentication
	expectedUsername := "testuser"
	expectedPassword := "testpass"

	// Encode the username and password into the expected format for the Authorization header
	expectedAuthValue := "Basic " + base64.StdEncoding.EncodeToString([]byte(expectedUsername+":"+expectedPassword))

	// Set up a mock server to handle the request. This server checks the Authorization header
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the Authorization header from the incoming request
		authHeader := r.Header.Get(HeaderAuthorization)

		// Compare the Authorization header to the expected value
		if authHeader != expectedAuthValue {
			// If they don't match, respond with 401 Unauthorized to indicate a failed authentication attempt
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// If the Authorization header is correct, respond with 200 OK to indicate successful authentication
		w.WriteHeader(http.StatusOK)
	}))

	defer mockServer.Close() // Ensure the server is shut down at the end of the test

	// Initialize the HTTP client with the base URL set to the mock server's URL
	client := Create(&Config{
		BaseURL: mockServer.URL,
	})

	// Create a request to the mock server with basic authentication credentials
	resp, err := client.Get("/").Auth(BasicAuth{
		Username: expectedUsername,
		Password: expectedPassword,
	}).Send(context.Background())

	if err != nil {
		// If there's an error sending the request, fail the test
		t.Fatalf("Failed to send request: %v", err)
	}

	defer resp.Close() //nolint: errcheck

	// Check if the response status code is 200 OK, which indicates successful authentication
	if resp.StatusCode() != http.StatusOK {
		// If the status code is not 200, it indicates the Authorization header was not set correctly
		t.Errorf("Expected status code 200, got %d. Indicates Authorization header was not set correctly", resp.StatusCode())
	}
}
