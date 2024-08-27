package httpsling

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	yaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

// startTestHTTPServer starts a test HTTP server that responds to various endpoints for testing purposes
func startTestHTTPServer() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/test-get", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "GET response")
	})

	handler.HandleFunc("/test-post", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "POST response")
	})

	handler.HandleFunc("/test-put", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "PUT response")
	})

	handler.HandleFunc("/test-delete", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "DELETE response")
	})

	handler.HandleFunc("/test-patch", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "PATCH response")
	})

	handler.HandleFunc("/test-status-code", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated) // 201
		fmt.Fprintln(w, `Created`)
	})

	handler.HandleFunc("/test-headers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "TestValue")
		fmt.Fprintln(w, `Headers test`)
	})

	handler.HandleFunc("/test-cookies", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "test-cookie", Value: "cookie-value"})
		fmt.Fprintln(w, `Cookies test`)
	})

	handler.HandleFunc("/test-body", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "This is the response body.")
	})

	handler.HandleFunc("/test-empty", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler.HandleFunc("/test-json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeJSON)
		fmt.Fprintln(w, `{"message": "This is a JSON response", "status": true}`)
	})

	handler.HandleFunc("/test-xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeXML)
		fmt.Fprintln(w, `<Response><Message>This is an XML response</Message><Status>true</Status></Response>`)
	})

	handler.HandleFunc("/test-text", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeText)
		fmt.Fprintln(w, `This is a text response`)
	})

	handler.HandleFunc("/test-pdf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, "application/pdf")
		fmt.Fprintln(w, `This is a PDF response`)
	})

	handler.HandleFunc("/test-redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/test-redirected", http.StatusFound)
	})

	handler.HandleFunc("/test-redirected", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Redirected")
	})

	handler.HandleFunc("/test-failure", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})

	return httptest.NewServer(handler)
}

// testRoundTripperFunc type is an adapter to allow the use of ordinary functions as http.RoundTrippers
type testRoundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip executes a single HTTP transaction
func (f testRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// TestSetHTTPClient verifies that SetHTTPClient correctly sets a custom http.Client and uses it for subsequent httpsling, specifically checking for cookie modifications
func TestSetHTTPClient(t *testing.T) {
	// Create a mock server that inspects incoming httpsling for a specific cookie
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for the presence of a specific cookie
		cookie, err := r.Cookie("X-Custom-Test-Cookie")
		if err != nil || cookie.Value != "true" {
			// If the cookie is missing or not as expected, respond with a 400 Bad Request
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// If the cookie is present and correct, respond with a 200
		w.WriteHeader(http.StatusOK)
	}))

	defer mockServer.Close()

	// Create a new instance of your Client
	client := Create(&Config{
		BaseURL: mockServer.URL,
	})

	// Define a custom transport that adds a custom cookie to all outgoing httpsling
	customTransport := testRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// Add the custom cookie to the request
		req.AddCookie(&http.Cookie{Name: "X-Custom-Test-Cookie", Value: "true"})
		// Proceed with the default transport after adding the cookie
		return http.DefaultTransport.RoundTrip(req)
	})

	// Set the custom http.Client with the custom transport to your Client
	client.SetHTTPClient(&http.Client{
		Transport: customTransport,
	})

	// Send a request using the custom http.Client
	resp, err := client.Get("/test").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	defer resp.Close() //nolint: errcheck

	// Verify that the server responded with a 200 OK, indicating the custom cookie was successfully added.
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("Expected status code 200, got %d. Indicates custom cookie was not recognized by the server.", resp.StatusCode())
	}
}

func TestClientURL(t *testing.T) {
	client := URL("http://localhost:8080")
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.BaseURL)
}

func TestClientGetRequest(t *testing.T) {
	server := startTestHTTPServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})
	resp, err := client.Get("/test-get").Send(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "GET response\n", resp.String())
}

func TestClientPostRequest(t *testing.T) {
	server := startTestHTTPServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})
	resp, err := client.Post("/test-post").Body(map[string]interface{}{"key": "value"}).Send(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "POST response\n", resp.String())
}

func TestClientPutRequest(t *testing.T) {
	server := startTestHTTPServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})
	resp, err := client.Put("/test-put").Body(map[string]interface{}{"key": "value"}).Send(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "PUT response\n", resp.String())
}

func TestClientDeleteRequest(t *testing.T) {
	server := startTestHTTPServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})
	resp, err := client.Delete("/test-delete").Send(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "DELETE response\n", resp.String())
}

func TestClientPatchRequest(t *testing.T) {
	server := startTestHTTPServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})
	resp, err := client.Patch("/test-patch").Body(map[string]interface{}{"key": "value"}).Send(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "PATCH response\n", resp.String())
}

func TestClientOptionsRequest(t *testing.T) {
	server := startTestHTTPServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})
	resp, err := client.Options("/test-get").Send(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.RawResponse.StatusCode)
}

func TestClientHeadRequest(t *testing.T) {
	server := startTestHTTPServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})
	resp, err := client.Head("/test-get").Send(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.RawResponse.StatusCode)
}

func TestClientTraceRequest(t *testing.T) {
	server := startTestHTTPServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})
	resp, err := client.Trace("/test-get").Send(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.RawResponse.StatusCode)
}

func TestClientCustomMethodRequest(t *testing.T) {
	server := startTestHTTPServer()
	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})
	resp, err := client.Custom("/test-get", "OPTIONS").Send(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.RawResponse.StatusCode)
}

// testSchema represents the JSON structure for testing
type testSchema struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// TestSetJSONMarshal tests custom JSON marshal functionality
func TestSetJSONMarshal(t *testing.T) {
	// Start a mock HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read body from the request
		var received testSchema
		err := json.NewDecoder(r.Body).Decode(&received)
		assert.NoError(t, err)
		assert.Equal(t, "John Snow", received.Name)
		assert.Equal(t, 30, received.Age)
	}))

	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	client.SetJSONMarshal(sonic.Marshal)

	data := testSchema{
		Name: "John Snow",
		Age:  30,
	}

	// Send a request with the custom marshaled body
	resp, err := client.Post("/").JSONBody(&data).Send(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

// TestSetJSONUnmarshal tests custom JSON unmarshal functionality
func TestSetJSONUnmarshal(t *testing.T) {
	// Mock response data
	mockResponse := `{"name":"Jane Doe","age":25}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeJSON)
		fmt.Fprintln(w, mockResponse)
	}))

	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	// Set the custom JSON unmarshal function
	client.SetJSONUnmarshal(sonic.Unmarshal)

	// Fetch and unmarshal the response
	resp, err := client.Get("/").Send(context.Background())
	assert.NoError(t, err)

	var result testSchema
	err = resp.Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, "Jane Doe", result.Name)
	assert.Equal(t, 25, result.Age)
}

type xmlTestSchema struct {
	XMLName xml.Name `xml:"Test"`
	Message string   `xml:"Message"`
	Status  bool     `xml:"Status"`
}

func TestSetXMLMarshal(t *testing.T) {
	// Mock server to check the received XML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var received xmlTestSchema
		err := xml.NewDecoder(r.Body).Decode(&received)
		assert.NoError(t, err)
		assert.Equal(t, "Test message", received.Message)
		assert.True(t, received.Status)
	}))

	defer server.Close()

	// Create your client and set the XML marshal function to use Go's default
	client := Create(&Config{BaseURL: server.URL})
	client.SetXMLMarshal(xml.Marshal)

	// Data to marshal and send
	data := xmlTestSchema{
		Message: "Test message",
		Status:  true,
	}

	// Marshal and send the data
	resp, err := client.Post("/").XMLBody(&data).Send(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSetXMLUnmarshal(t *testing.T) {
	// Mock server to send XML data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeXML)
		fmt.Fprintln(w, `<Test><Message>Response message</Message><Status>true</Status></Test>`)
	}))
	defer server.Close()

	// Create your client and set the XML unmarshal function to use go's default
	client := Create(&Config{BaseURL: server.URL})
	client.SetXMLUnmarshal(xml.Unmarshal)

	// Fetch and attempt to unmarshal the data
	resp, err := client.Get("/").Send(context.Background())
	assert.NoError(t, err)

	var result xmlTestSchema
	err = resp.Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, "Response message", result.Message)
	assert.True(t, result.Status)
}

func TestSetYAMLMarshal(t *testing.T) {
	type yamlTestSchema struct {
		Message string `yaml:"message"`
		Status  bool   `yaml:"status"`
	}

	// Mock server to check the received YAML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var received yamlTestSchema
		err := yaml.NewDecoder(r.Body).Decode(&received)
		assert.NoError(t, err)
		assert.Equal(t, "Test message", received.Message)
		assert.True(t, received.Status)
	}))

	defer server.Close()

	// Create your client and set the YAML marshal function to use goccy/go-yaml's Marshal
	client := Create(&Config{BaseURL: server.URL})
	client.SetYAMLMarshal(yaml.Marshal)

	// Data to marshal and send
	data := yamlTestSchema{
		Message: "Test message",
		Status:  true,
	}

	// Marshal and send the data
	resp, err := client.Post("/").YAMLBody(&data).Send(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSetYAMLUnmarshal(t *testing.T) {
	// Define a test schema
	type yamlTestSchema struct {
		Message string `yaml:"message"`
		Status  bool   `yaml:"status"`
	}

	// Mock server to send YAML data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, ContentTypeYAML)
		fmt.Fprintln(w, "message: Response message\nstatus: true")
	}))
	defer server.Close()

	// Create your client and set the YAML unmarshal function to use goccy/go-yaml's Unmarshal
	client := Create(&Config{BaseURL: server.URL})
	client.SetYAMLUnmarshal(yaml.Unmarshal)

	// Fetch and attempt to unmarshal the data
	resp, err := client.Get("/").Send(context.Background())
	assert.NoError(t, err)

	var result yamlTestSchema
	err = resp.Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, "Response message", result.Message)
	assert.True(t, result.Status)
}

// TestSetAuth verifies that SetAuth correctly sets the Authorization header for basic authentication
func TestSetAuth(t *testing.T) {
	// Expected username and password
	expectedUsername := "testuser"
	expectedPassword := "testpass"

	// Expected Authorization header value
	expectedAuthValue := "Basic " + base64.StdEncoding.EncodeToString([]byte(expectedUsername+":"+expectedPassword))

	// Create a mock server that checks the Authorization header
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the Authorization header from the request.
		authHeader := r.Header.Get(HeaderAuthorization)

		// Check if the Authorization header matches the expected value
		if authHeader != expectedAuthValue {
			// If not, respond with 401
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// If the header is correct, respond with 200 OK
		w.WriteHeader(http.StatusOK)
	}))

	defer mockServer.Close()

	client := Create(&Config{
		BaseURL: mockServer.URL,
	})

	// Set basic authentication using the SetBasicAuth method
	client.SetAuth(BasicAuth{
		Username: expectedUsername,
		Password: expectedPassword,
	})

	// Send the request through the client
	resp, err := client.Get("/").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Close() //nolint: errcheck

	// Check the response status code
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("Expected status code 200, got %d. Indicates Authorization header was not set correctly.", resp.StatusCode())
	}
}

func TestSetDefaultHeaders(t *testing.T) {
	// Create a mock server to check headers
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "HeaderValue" {
			t.Error("Default header 'X-Custom-Header' not found or value incorrect")
		}
	}))

	defer mockServer.Close()

	// Initialize the client and set a default header
	client := Create(&Config{BaseURL: mockServer.URL})
	client.SetDefaultHeader("X-Custom-Header", "HeaderValue")

	// Make a request to trigger the header check
	_, err := client.Get("/").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
}

func TestDelDefaultHeader(t *testing.T) {
	// Mock server to check for the absence of a specific header
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Deleted-Header") != "" {
			t.Error("Deleted default header 'X-Deleted-Header' was found in the request")
		}
	}))

	defer mockServer.Close()

	// Initialize the client, set, and then delete a default header
	client := Create(&Config{BaseURL: mockServer.URL})
	client.SetDefaultHeader("X-Deleted-Header", "ShouldBeDeleted")
	client.DelDefaultHeader("X-Deleted-Header")

	// Make a request to check for the absence of the deleted header
	_, err := client.Get("/").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
}

func TestSetDefaultContentType(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the Content-Type header
		if r.Header.Get(HeaderContentType) != ContentTypeJSON {
			t.Error("Default Content-Type header not set correctly")
		}
	}))
	defer mockServer.Close()

	client := Create(&Config{BaseURL: mockServer.URL})
	client.SetDefaultContentType(ContentTypeJSON)

	_, err := client.Get("/").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
}

func TestSetDefaultAccept(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the Accept header
		if r.Header.Get(HeaderAccept) != ContentTypeXML {
			t.Error("Default Accept header not set correctly")
		}
	}))
	defer mockServer.Close()

	client := Create(&Config{BaseURL: mockServer.URL})
	client.SetDefaultAccept(ContentTypeXML)

	_, err := client.Get("/").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
}

func TestSetDefaultUserAgent(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the User-Agent header
		if r.Header.Get(HeaderUserAgent) != "MyCustomAgent/1.0" {
			t.Error("Default User-Agent header not set correctly")
		}
	}))
	defer mockServer.Close()

	client := Create(&Config{BaseURL: mockServer.URL})
	client.SetDefaultUserAgent("MyCustomAgent/1.0")

	_, err := client.Get("/").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
}

func TestSetDefaultTimeout(t *testing.T) {
	// Create a server that delays its response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than client's timeout
	}))
	defer mockServer.Close()

	client := Create(&Config{BaseURL: mockServer.URL})
	client.SetDefaultTimeout(1 * time.Second) // Set timeout to 1 second

	_, err := client.Get("/").Send(context.Background())
	if err == nil {
		t.Fatal("Expected a timeout error, got nil")
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		// Check if the error is a timeout error
	} else {
		t.Fatalf("Expected a timeout error, got %v", err)
	}
}

func TestSetDefaultCookieJar(t *testing.T) {
	jar, _ := cookiejar.New(nil)

	// Initialize the client and set the default cookie jar nom nom nom
	client := Create(&Config{})
	client.SetCookieJar(jar)

	// Start a test HTTP server that sets a cookie
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/set-cookie" {
			http.SetCookie(w, &http.Cookie{Name: "test", Value: "cookie"})
			return
		}

		// Check for the cookie on a different endpoint
		cookie, err := r.Cookie("test")
		if err != nil {
			t.Fatal("Cookie 'test' not found in request, cookie jar not working")
		}

		if cookie.Value != "cookie" {
			t.Fatalf("Expected cookie 'test' to have value 'cookie', got '%s'", cookie.Value)
		}
	}))

	defer server.Close()

	// First request to set the cookie
	_, err := client.Get(server.URL + "/set-cookie").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	// Second request to check if the cookie is sent back
	_, err = client.Get(server.URL + "/check-cookie").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send second request: %v", err)
	}
}

func TestSetDefaultCookies(t *testing.T) {
	// Create a mock server to check cookies
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for the presence of specific cookies
		sessionCookie, err := r.Cookie("session_id")
		if err != nil || sessionCookie.Value != "abcd1234" {
			t.Error("Default cookie 'session_id' not found or value incorrect")
		}

		authCookie, err := r.Cookie("auth_token")
		if err != nil || authCookie.Value != "token1234" {
			t.Error("Default cookie 'auth_token' not found or value incorrect")
		}
	}))

	defer mockServer.Close()

	// Initialize the client and set default cookies
	client := Create(&Config{BaseURL: mockServer.URL})
	client.SetDefaultCookies(map[string]string{
		"session_id": "abcd1234",
		"auth_token": "token1234",
	})

	// Make a request to trigger the cookie check
	_, err := client.Get("/").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
}

func TestDelDefaultCookie(t *testing.T) {
	// Mock server to check for absence of a specific cookie
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("session_id")
		if err == nil {
			t.Error("Deleted default cookie 'session_id' was found in the request")
		}
	}))

	defer mockServer.Close()

	// Initialize the client, set, and then delete a default cookie
	client := Create(&Config{BaseURL: mockServer.URL})
	client.SetDefaultCookie("session_id", "abcd1234")
	client.DelDefaultCookie("session_id")

	// Make a request to check for the absence of the deleted cookie
	_, err := client.Get("/").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
}

func createTestRetryServer(t *testing.T) *httptest.Server {
	var requestCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		switch requestCount {
		case 1:
			w.WriteHeader(http.StatusInternalServerError) // Simulate server error on first attempt
		case 2:
			w.WriteHeader(http.StatusOK) // Successful on second attempt
		default:
			t.Fatalf("Unexpected number of httpsling: %d", requestCount)
		}
	}))

	return server
}

func TestSetMaxRetriesAndRetryStrategy(t *testing.T) {
	server := createTestRetryServer(t)

	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	retryCalled := false

	client.SetMaxRetries(1).SetRetryStrategy(func(attempt int) time.Duration {
		retryCalled = true
		return 10 * time.Millisecond // Short delay for testing
	})

	// Make a request to the test server
	_, err := client.Get("/").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if !retryCalled {
		t.Error("Expected retry strategy to be called, but it wasn't")
	}
}

func TestSetRetryIf(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // Always return server error
	}))

	defer server.Close()

	client := Create(&Config{BaseURL: server.URL})

	client.SetMaxRetries(2).SetRetryIf(func(req *http.Request, resp *http.Response, err error) bool {
		return resp.StatusCode == http.StatusInternalServerError
	})

	retryCount := 0

	client.SetRetryStrategy(func(int) time.Duration {
		retryCount++
		return 10 * time.Millisecond // Short delay for testing
	})

	// Make a request to the test server
	_, err := client.Get("/").Send(context.Background())
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	if retryCount != 2 {
		t.Errorf("Expected 2 retries, got %d", retryCount)
	}
}
