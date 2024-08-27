# Slinging HTTP

The `httpsling` library simplifies the way you make HTTP httpsling. It's intended to provide an easy-to-use interface for sending requests and handling responses, reducing the boilerplate code typically associated with the `net/http` package.

## Overview

Creating a new HTTP client and making a request should be straightforward:

```go
package main

import (
    "github.com/theopenlane/core/pkg/httpsling"
    "log"
)

func main() {
    // Create a client using a base URL
    client := httpsling.URL("http://mattisthebest.com")

    // Alternatively, create a client with custom configuration
    client = httpsling.Create(&httpsling.Config{
        BaseURL: "http://mattisthebest.com",
        Timeout: 30 * time.Second,
    })

    // Perform a GET request
    resp, err := client.Get("/resource")
    if err != nil {
        log.Fatal(err)
    }

    defer resp.Close()

    log.Println(resp.String())
}
```

## Client

The `Client` struct is your gateway to making HTTP requests. You can configure it to your needs, setting default headers, cookies, timeout durations, etc.

```go
client := httpsling.URL("http://mattisthebest.com")

// Or, with full configuration
client = httpsling.Create(&httpsling.Config{
    BaseURL: "http://mattisthebest.com",
    Timeout: 5 * time.Second,
    Headers: &http.Header{
        HeaderContentType: []string{ContentTypeJSON},
    },
})
```

### Initializing the Client

You can start by creating a `Client` with specific configurations using the `Create` method:

```go
client := httpsling.Create(&httpsling.Config{
    BaseURL: "https://the.cats.meow.com",
    Timeout: 30 * time.Second,
    Headers: &http.Header{
        HeaderAuthorization: []string{"Bearer YOUR_ACCESS_TOKEN"},
        HeaderContentType: []string{ContentTypeJSON},
    },
    Cookies: map[string]string{
        "session_token": "YOUR_SESSION_TOKEN",
    },
    TLSConfig: &tls.Config{
        InsecureSkipVerify: true,
    },
    MaxRetries: 3,
    RetryStrategy: httpsling.ExponentialBackoffStrategy(1*time.Second, 2, 30*time.Second),
    RetryIf: httpsling.DefaultRetryIf,
})
```

This setup creates a `Client` tailored for your API communication, including base URL, request timeout, default headers, and cookies

### Configuring with Set Methods

Alternatively, you can use `Set` methods for a more dynamic configuration approach:

```go
client := httpsling.URL("https://the.cats.meow.com").
    SetDefaultHeader(HeaderAuthorization, "Bearer YOUR_ACCESS_TOKEN").
    SetDefaultHeader(HeaderContentType, ContentTypeJSON).
    SetDefaultCookie("session_token", "YOUR_SESSION_TOKEN").
    SetTLSConfig(&tls.Config{InsecureSkipVerify: true}).
    SetMaxRetries(3).
    SetRetryStrategy(httpsling.ExponentialBackoffStrategy(1*time.Second, 2, 30*time.Second)).
    SetRetryIf(httpsling.DefaultRetryIf).
    SetProxy("http://localhost:8080")
```

### Configuring BaseURL

Set the base URL for all requests:

```go
client.SetBaseURL("https://the.cats.meow.com")
```

### Setting Headers

Set default headers for all requests:

```go
client.SetDefaultHeader(HeaderAuthorization, "Bearer YOUR_ACCESS_TOKEN")
client.SetDefaultHeader(HeaderContentType, ContentTypeJSON)
```

Bulk set default headers:

```go
headers := &http.Header{
    HeaderAuthorization: []string{"Bearer YOUR_ACCESS_TOKEN"},
    HeaderContentType:  []string{ContentTypeJSON},
}
client.SetDefaultHeaders(headers)
```

Add or remove a header:

```go
client.AddDefaultHeader("X-Custom-Header", "Value1")
client.DelDefaultHeader("X-Unneeded-Header")
```

### Managing Cookies

Set default cookies for all requests:

```go
client.SetDefaultCookie("session_id", "123456")
```

Bulk set default cookies:

```go
cookies := map[string]string{
    "session_id": "123456",
    "preferences": "dark_mode=true",
}
client.SetDefaultCookies(cookies)
```

Remove a default cookie:

```go
client.DelDefaultCookie("session_id")
```

This approach simplifies managing base URLs, headers, and cookies across all requests made with the client, ensuring consistency.

### Configuring Timeouts

Define a global timeout for all requests to prevent indefinitely hanging operations:

```go
client := httpsling.Create(&httpsling.Config{
    Timeout: 15 * time.Second,
})
```

### TLS Configuration

Custom TLS configurations can be applied for enhanced security measures, such as loading custom certificates:

```go
tlsConfig := &tls.Config{InsecureSkipVerify: true}
client.SetTLSConfig(tlsConfig)
```

## Requests

The library provides a `RequestBuilder` to construct and dispatch HTTP httpsling. Here are examples of performing various types of requests, including adding query parameters, setting headers, and attaching a body to your httpsling.

#### GET Request

```go
resp, err := client.Get("/path").
    Query("search", "query").
    Header(HeaderAccept, ContentTypeJSON).
    Send(context.Background())
```

#### POST Request

```go
resp, err := client.Post("/path").
    Header(HeaderContentType, ContentTypeJSON).
    JSONBody(map[string]interface{}{"key": "value"}).
    Send(context.Background())
```

#### PUT Request

```go
resp, err := client.Put("/stff/{stuff_id}").
    PathParam("stuff_id", "123456").
    JSONBody(map[string]interface{}{"updatedKey": "newValue"}).
    Send(context.Background())
```

#### DELETE Request

```go
resp, err := client.Delete("/stffs/{stuff_id}").
    PathParam("stuff_id", "123456meowmeow").
    Send(context.Background())
```

### Retry Mechanism

Automatically retry requests on failure with customizable strategies:

```go
client.SetMaxRetries(3)
client.SetRetryStrategy(httpsling.ExponentialBackoffStrategy(1*time.Second, 2, 30*time.Second))
client.SetRetryIf(func(req *http.Request, resp *http.Response, err error) bool {
	// Only retry for 500 Internal Server Error
	return resp.StatusCode == http.StatusInternalServerError
})
```

### Configuring Retry Strategies

#### Applying a Default Backoff Strategy

For consistent delay intervals between retries:

```go
client.SetRetryStrategy(httpsling.DefaultBackoffStrategy(5 * time.Second))
```

#### Utilizing a Linear Backoff Strategy

To increase delay intervals linearly with each retry attempt:

```go
client.SetRetryStrategy(httpsling.LinearBackoffStrategy(1 * time.Second))
```

#### Employing an Exponential Backoff Strategy

For exponential delay increases between attempts, with an option to cap the delay:

```go
client.SetRetryStrategy(httpsling.ExponentialBackoffStrategy(1*time.Second, 2, 30*time.Second))
```

### Customizing Retry Conditions

Define when retries should be attempted based on response status codes or errors:

```go
client.SetRetryIf(func(req *http.Request, resp *http.Response, err error) bool {
    return resp.StatusCode == http.StatusInternalServerError || err != nil
})
```

### Setting Maximum Retry Attempts

To limit the number of retries, use the `SetMaxRetries` method:

```go
client.SetMaxRetries(3)
```

### Proxy Configuration

Route requests through a proxy server:

```go
client.SetProxy("http://localhost:8080")
```

### Authentication

Supports various authentication methods:

- **Basic Auth**:

```go
client.SetAuth(httpsling.BasicAuth{
  Username: "user",
  Password: "pass",
})
```

- **Bearer Token**:

```go
client.SetAuth(httpsling.BearerAuth{
    Token: "YOUR_ACCESS_TOKEN",
})
```

### Query Parameters

Add query parameters to your request using `Query`, `Queries`, `QueriesStruct`, or remove them with `DelQuery`

```go
// Add a single query parameter
request.Query("search", "query")

// Add multiple query parameters
request.Queries(url.Values{"sort": []string{"date"}, "limit": []string{"10"}})

// Add query parameters from a struct
type queryParams struct {
    Sort  string `url:"sort"`
    Limit int    `url:"limit"`
}
request.QueriesStruct(queryParams{Sort: "date", Limit: 10})

// Remove one or more query parameters
request.DelQuery("sort", "limit")
```

### Headers

Set request headers using `Header`, `Headers`, or related methods

```go
request.Header(HeaderAuthorization, "Bearer YOUR_ACCESS_TOKEN")
request.Headers(http.Header{HeaderContentType: []string{ContentTypeJSON}})

// Convenient methods for common headers
request.ContentType(ContentTypeJSON)
request.Accept(ContentTypeJSON)
request.UserAgent("MyCustomClient/1.0")
request.Referer("https://example.com")
```

### Cookies

Add cookies to your request using `Cookie`, `Cookies`, or remove them with `DelCookie`.

```go
// Add a single cookie
request.Cookie("session_token", "YOUR_SESSION_TOKEN")

// Add multiple cookies at once
request.Cookies(map[string]string{
    "session_token": "YOUR_SESSION_TOKEN",
    "user_id": "12345",
})

// Remove one or more cookies
request.DelCookie("session_token", "user_id")

```

### Body Content

Specify the request body directly with `Body` or use format-specific methods like `JSONBody`, `XMLBody`, `YAMLBody`, `TextBody`, or `RawBody` for appropriate content types.

```go
// Setting JSON body
request.JSONBody(map[string]interface{}{"key": "value"})

// Setting XML body
request.XMLBody(myXmlStruct)

// Setting YAML body
request.YAMLBody(myYamlStruct)

// Setting text body
request.TextBody("plain text content")

// Setting raw body
request.RawBody([]byte("raw data"))
```

### Timeout and Retries

Configure request-specific timeout and retry strategies:

```go
request.Timeout(10 * time.Second).MaxRetries(3)
```

### Sending Requests

The `Send(ctx)` method executes the HTTP request built with the Request builder. It requires a `context.Context` argument, allowing you to control request cancellation and timeouts.

```go
resp, err := request.Send(context.Background())
if err != nil {
    log.Fatalf("Request failed: %v", err)
}
// Process response...
```
### Advanced Features

#### Handling Cancellation

To cancel a request, simply use the context's cancel function. This is particularly useful for long-running requests that you may want to abort if they take too long or if certain conditions are met.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel() // Ensures resources are freed up after the operation completes or times out

// Cancel the request if it hasn't completed within the timeout
resp, err := request.Send(ctx)
if errors.Is(err, context.Canceled) {
    log.Println("Request was canceled")
}
```

#### HTTP Client Customization

Directly customize the underlying `http.Client`:

```go
customHTTPClient := &http.Client{Timeout: 20 * time.Second}
client.SetHTTPClient(customHTTPClient)
```

#### Path Parameters

To insert or modify path parameters in your URL, use `PathParam` for individual parameters or `PathParams` for multiple. For removal, use `DelPathParam`.

```go
// Setting a single path parameter
request.PathParam("userId", "123")

// Setting multiple path parameters at once
request.PathParams(map[string]string{"userId": "123", "postId": "456"})

// Removing path parameters
request.DelPathParam("userId", "postId")
```

When using `client.Get("/users/{userId}/posts/{postId}")`, replace `{userId}` and `{postId}` with actual values by using `PathParams` or `PathParam`.

#### Form Data

For `application/x-www-form-urlencoded` content, utilize `FormField` for individual fields or `FormFields` for multiple.

```go
// Adding individual form field
request.FormField("name", "John Snow")

// Setting multiple form fields at once
fields := map[string]interface{}{"name": "John", "age": "30"}
request.FormFields(fields)
```

#### File Uploads

To include files in a `multipart/form-data` request, specify each file's form field name, file name, and content using `File` or add multiple files with `Files`.

```go
// Adding a single file
file, _ := os.Open("path/to/file")
request.File("profile_picture", "filename.jpg", file)

// Adding multiple files
request.Files(file1, file2)
```

### Authentication

Apply authentication methods directly to the request:

```go
request.Auth(httpsling.BasicAuth{
   Username: "user",
   Password: "pass",
})
```

## Middleware

Add custom middleware to process the request or response:

```go
request.AddMiddleware(func(next httpsling.MiddlewareHandlerFunc) httpsling.MiddlewareHandlerFunc {
    return func(req *http.Request) (*http.Response, error) {
        // Custom logic before request
        resp, err := next(req)
        // Custom logic after response
        return resp, err
    }
})
```

### Understanding Middleware

Middleware functions wrap around HTTP requests, allowing pre- and post-processing of requests and responses. They can modify requests before they are sent, examine responses, and decide whether to modify them, retry the request, or take other actions.

### Client-Level Middleware

Client-level middleware is applied to all requests made by a client. It's ideal for cross-cutting concerns like logging, error handling, and metrics collection.

**Adding Middleware to a Client:**

```go
client := httpsling.Create(&httpsling.Config{BaseURL: "https://the.cats.meow.com"})
client.AddMiddleware(func(next httpsling.MiddlewareHandlerFunc) httpsling.MiddlewareHandlerFunc {
    return func(req *http.Request) (*http.Response, error) {
        // Pre-request manipulation
        fmt.Println("Request URL:", req.URL)

        // Proceed with the request
        resp, err := next(req)

        // Post-response manipulation
        if err == nil {
            fmt.Println("Response status:", resp.Status)
        }

        return resp, err
    }
})
```

### Request-Level Middleware

Request-level middleware applies only to individual httpsling. This is useful for request-specific concerns, such as request tracing or modifying the request based on dynamic context.

**Adding Middleware to a Request:**

```go
request := client.NewRequestBuilder(MethodGet, "/path").AddMiddleware(func(next httpsling.MiddlewareHandlerFunc) httpsling.MiddlewareHandlerFunc {
    return func(req *http.Request) (*http.Response, error) {
        // Modify the request here
        req.Header.Add("X-Request-ID", "12345")

        // Proceed with the modified request
        return next(req)
    }
})
```

### Implementing Custom Middleware

Custom middleware can perform a variety of tasks, such as authentication, logging, and metrics. Here's a simple logging middleware example:

```go
func loggingMiddleware(next httpsling.MiddlewareHandlerFunc) httpsling.MiddlewareHandlerFunc {
    return func(req *http.Request) (*http.Response, error) {
        log.Printf("Requesting %s %s", req.Method, req.URL)
        resp, err := next(req)
        if err != nil {
            log.Printf("Request to %s failed: %v", req.URL, err)
        } else {
            log.Printf("Received %d response from %s", resp.StatusCode, req.URL)
        }
        return resp, err
    }
}
```

### Integrating OpenTelemetry Middleware

OpenTelemetry middleware can be used to collect tracing and metrics for your requests if you're into that sort of thing. Below is an example of how to set up a basic trace for an HTTP request:

**Implementing OpenTelemetry Middleware:**

```go
func openTelemetryMiddleware(next httpsling.MiddlewareHandlerFunc) httpsling.MiddlewareHandlerFunc {
    return func(req *http.Request) (*http.Response, error) {
        ctx, span := otel.Tracer("requests").Start(req.Context(), req.URL.Path)
        defer span.End()

        // Add trace ID to request headers if needed
        traceID := span.SpanContext().TraceID().String()
        req.Header.Set("X-Trace-ID", traceID)

        resp, err := next(req)

        // Set span attributes based on response
        if err == nil {
            span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
        } else {
            span.RecordError(err)
        }

        return resp, err
    }
}
```


## Responses

Handling responses is necessary in determining the outcome of your HTTP requests - the library has some built-in response code validators and other tasty things.

```go
type APIResponse struct {
    Data string `json:"data"`
}

var apiResp APIResponse
if err := resp.ScanJSON(&apiResp); err != nil {
    log.Fatal(err)
}

log.Printf("Status Code: %d\n", resp.StatusCode())
log.Printf("Response Data: %s\n", apiResp.Data)
```

### Parsing Response Body

By leveraging the `Scan`, `ScanJSON`, `ScanXML`, and `ScanYAML` methods, you can decode responses based on the `Content-Type`

#### JSON Responses

Given a JSON response, you can unmarshal it directly into a Go struct using either the specific `ScanJSON` method or the generic `Scan` method, which automatically detects the content type:

```go
var jsonData struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// Unmarshal using ScanJSON
if err := response.ScanJSON(&jsonData); err != nil {
    log.Fatalf("Error unmarshalling JSON: %v", err)
}

// Alternatively, unmarshal using Scan
if err := response.Scan(&jsonData); err != nil {
    log.Fatalf("Error unmarshalling response: %v", err)
}
```

#### XML Responses

For XML responses, use `ScanXML` or `Scan` to decode into a Go struct. Here's an example assuming the response contains XML data:

```go
var xmlData struct {
    Name string `xml:"name"`
    Age  int    `xml:"age"`
}

// Unmarshal using ScanXML
if err := response.ScanXML(&xmlData); err != nil {
    log.Fatalf("Error unmarshalling XML: %v", err)
}

// Alternatively, unmarshal using Scan
if err := response.Scan(&xmlData); err != nil {
    log.Fatalf("Error unmarshalling response: %v", err)
}
```

#### YAML Responses

YAML content is similarly straightforward to handle. The `ScanYAML` or `Scan` method decodes the YAML response into the specified Go struct:

```go
var yamlData struct {
    Name string `yaml:"name"`
    Age  int    `yaml:"age"`
}

// Unmarshal using ScanYAML
if err := response.ScanYAML(&yamlData); err != nil {
    log.Fatalf("Error unmarshalling YAML: %v", err)
}

// Alternatively, unmarshal using Scan
if err := response.Scan(&yamlData); err != nil {
    log.Fatalf("Error unmarshalling response: %v", err)
}
```

### Storing Response Content

For saving the response body to a file or streaming it to an `io.Writer`:

- **Save**: Write the response body to a designated location

    ```go
    // Save response to a file
    if err := response.Save("downloaded_file.txt"); err != nil {
        log.Fatalf("Failed to save file: %v", err)
    }
    ```

### Evaluating Response Success

To assess whether the HTTP request was successful:

- **IsSuccess**: Check if the status code signifies a successful response

    ```go
    if response.IsSuccess() {
        fmt.Println("The request succeeded hot diggity dog")
    }
    ```


## Enabling Logging

To turn on logging, you must explicitly initialize and set a `Logger` in the client configuration. Here's how to create and use the `DefaultLogger`, which logs to `os.Stderr` by default, and is configured to log errors only:

```go
logger := httpsling.NewDefaultLogger(os.Stderr, slog.LevelError)
client := httpsling.Create(&httpsling.Config{
    Logger: logger,
})
```

Or, for an already instantiated client:

```go
client.SetLogger(httpsling.NewDefaultLogger(os.Stderr, slog.LevelError))
```

### Adjusting Log Levels

Adjusting the log level is straightforward. After defining your logger, simply set the desired level. This allows you to control the verbosity of the logs based on your requirements.

```go
logger := httpsling.NewDefaultLogger(os.Stderr, httpsling.LevelError)
logger.SetLevel(httpsling.LevelInfo) // Set to Info level to capture more detailed logs

client := httpsling.Create(&httpsling.Config{
    Logger: logger,
})
```

The available log levels are:

- `LevelDebug`
- `LevelInfo`
- `LevelWarn`
- `LevelError`

### Implementing a Custom Logger

For more advanced scenarios where you might want to integrate with an existing logging system or format logs differently, implement the `Logger` interface. This requires methods for each level of logging (`Debugf`, `Infof`, `Warnf`, `Errorf`) and a method to set the log level (`SetLevel`).

Here is a simplified example:

```go
type MyLogger struct {
    // Include your custom logging mechanism here
}

func (l *MyLogger) Debugf(format string, v ...any) {
    // Custom debug logging implementation
}

func (l *MyLogger) Infof(format string, v ...any) {
    // Custom info logging implementation
}

func (l *MyLogger) Warnf(format string, v ...any) {
    // Custom warn logging implementation
}

func (l *MyLogger) Errorf(format string, v ...any) {
    // Custom error logging implementation
}

func (l *MyLogger) SetLevel(level httpsling.Level) {
    // Implement setting the log level in your logger
}

// Usage
myLogger := &MyLogger{}
myLogger.SetLevel(httpsling.LevelDebug) // Example setting to Debug level

client := httpsling.Create(&httpsling.Config{
    Logger: myLogger,
})
```

## Stream Callbacks

Stream callbacks are functions that you define to handle chunks of data as they are received from the server. The Requests library supports three types of stream callbacks:

- **StreamCallback**: Invoked for each chunk of data received
- **StreamErrCallback**: Invoked when an error occurs during streaming
- **StreamDoneCallback**: Invoked once streaming is completed, regardless of whether it ended due to an error or successfully

### Configuring Stream Callbacks

To configure streaming for a request, use the `Stream` method on a `RequestBuilder` instance. This method accepts a `StreamCallback` function, which will be called with each chunk of data received from the server.

```go
streamCallback := func(data []byte) error {
    fmt.Println("Received stream data:", string(data))
    return nil // Return an error if needed to stop streaming
}

request := client.Get("/stream-endpoint").Stream(streamCallback)
```

### Handling Stream Errors

To handle errors that occur during streaming, set a `StreamErrCallback` using the `StreamErr` method on the `Response` object.

```go
streamErrCallback := func(err error) {
    fmt.Printf("Stream error: %v\n", err)
}

response, _ := request.Send(context.Background())
response.StreamErr(streamErrCallback)
```

### Completing Stream Processing

Once streaming is complete, you can use the `StreamDone` method on the `Response` object to set a `StreamDoneCallback`. This callback is invoked after the stream is fully processed, either successfully or due to an error.

```go
streamDoneCallback := func() {
    fmt.Println("Stream processing completed")
}

response.StreamDone(streamDoneCallback)
```

### Example: Consuming an SSE Stream

The following example demonstrates how to consume a Server-Sent Events (SSE) stream, processing each event as it arrives, handling errors, and performing cleanup once the stream ends.

```go
// Configure the stream callback to handle data chunks
streamCallback := func(data []byte) error {
    fmt.Println("Received stream event:", string(data))
    return nil
}

// Configure error and done callbacks
streamErrCallback := func(err error) {
    fmt.Printf("Error during streaming: %v\n", err)
}

streamDoneCallback := func() {
    fmt.Println("Stream ended")
}

// Create the streaming request
client := httpsling.Create(&httpsling.Config{BaseURL: "https://example.com"})
request := client.Get("/events").Stream(streamCallback)

// Send the request and configure callbacks
response, err := request.Send(context.Background())
if err != nil {
    fmt.Printf("Failed to start streaming: %v\n", err)
    return
}

response.StreamErr(streamErrCallback).StreamDone(streamDoneCallback)
```


## Inspirations

This library was inspired by and built upon the work of several other HTTP client libraries:

- [Dghubble/sling](https://github.com/dghubble/sling)
- [Monaco-io/request](https://github.com/monaco-io/request)
- [Go-resty/resty](https://github.com/go-resty/resty)
- [Fiber Client](https://github.com/gofiber/fiber)

Props to dghubble for a great name with `sling`, which was totally ripped off to make `httpsling` <3. I chose not to use any of these directly because I wanted to have layers of control we may need within our services echosystem.