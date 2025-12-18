package graphapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	echo "github.com/theopenlane/echox"
)

const (
	queryEndpoint        = "/query"
	historyQueryEndpoint = "/history/query"
)

// CheckGraphReadRequest checks if the incoming GraphQL request is a read-only query via POST
// this can be used to conditionally skip middleware like CSRF for read-only queries
func CheckGraphReadRequest(c echo.Context) bool {
	req := c.Request()

	if req.URL.Path != queryEndpoint && req.URL.Path != historyQueryEndpoint {
		return false
	}

	var queryStr string

	switch req.Method {
	case http.MethodOptions, http.MethodGet:
		return true
	case http.MethodPost:
		if req.Body == nil {
			// no body, unable to determine, so do not skip
			return false
		}

		// read and buffer the body
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return false
		}

		// restore body so downstream handlers/middleware can read it
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		var payload struct {
			Query string `json:"query"`
		}

		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			// if its not valid, just return false
			return false
		}

		queryStr = payload.Query
	default:
		return false
	}

	return detectGraphQLOperationType(queryStr)
}

// detectGraphQLOperationType inspects the GraphQL query string to determine the operation type
func detectGraphQLOperationType(query string) bool {
	s := strings.TrimSpace(query)

	// empty query cannot be determined as nothing was provided in the request
	if s == "" {
		return false
	}

	sLower := strings.ToLower(s)

	switch {
	case strings.HasPrefix(sLower, "mutation "):
		return false
	case strings.HasPrefix(sLower, "subscription "):
		// skip subscriptions as they are read-only as well
		return true
	case strings.HasPrefix(sLower, "query "):
		return true
	default:
		// anonymous operation like `{ me { id } }` is implicitly a query
		return true
	}
}
