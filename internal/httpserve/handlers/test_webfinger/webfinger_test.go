package test_webfinger

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// --- Mock structs used in Webfinger ---
type User struct {
	ID    string
	Email string
	Name  string
}

type Reply struct {
	Success bool
}

type SSOStatusResponse struct {
	Reply          Reply
	Enforced       bool
	OrganizationID string
	IsOrgOwner     bool
}

// --- Mock DB implementing only needed methods ---
type MockDB struct{}

func (db *MockDB) GetUserByEmail(email string) (User, error) {
	if email == "test@example.com" {
		return User{ID: "123", Email: email, Name: "Test User"}, nil
	}
	return User{}, errors.New("user not found")
}

func (db *MockDB) GetUserDefaultOrgID(userID string) (string, error) {
	if userID == "123" {
		return "org-1", nil
	}
	return "", errors.New("org not found")
}

func (db *MockDB) FetchSSOStatus(orgID, userID string) (SSOStatusResponse, error) {
	return SSOStatusResponse{
		Reply:          Reply{Success: true},
		Enforced:       true,
		OrganizationID: orgID,
		IsOrgOwner:     true,
	}, nil
}

// --- Minimal Handler stub ---
type Handler struct {
	DB *MockDB
}

func (h *Handler) WebfingerHandler(c echo.Context) error {
	resource := c.QueryParam("resource")

	// Validate resource
	if resource == "" || !strings.HasPrefix(resource, "acct:") {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid resource")
	}

	userEmail := resource[len("acct:"):]
	u, err := h.DB.GetUserByEmail(userEmail)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	orgID, _ := h.DB.GetUserDefaultOrgID(u.ID)
	status, _ := h.DB.FetchSSOStatus(orgID, u.ID)

	resp := map[string]interface{}{
		"subject": resource,
		"links": []interface{}{
			map[string]interface{}{
				"href":         "https://example.com/sso",
				"rel":          "self",
				"ssoStatus":    status,
				"organization": orgID,
			},
		},
	}

	return c.JSON(http.StatusOK, resp)
}

// --- Test WebfingerHandler ---
func TestWebfingerHandler(t *testing.T) {
	e := echo.New()
	h := &Handler{DB: &MockDB{}}

	cases := []struct {
		name     string
		resource string
		status   int
	}{
		{"valid user", "acct:test@example.com", http.StatusOK},
		{"nonexistent user", "acct:nouser@example.com", http.StatusNotFound},
		{"invalid resource", "invalid", http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource="+tc.resource, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.WebfingerHandler(c)

			if tc.status == http.StatusOK {
				assert.NilError(t, err)
				var resp map[string]interface{}
				assert.NilError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.Assert(t, is.Equal(tc.resource, resp["subject"]))
				links, _ := resp["links"].([]interface{})
				assert.Assert(t, len(links) > 0)
			} else {
				assert.Assert(t, err != nil)
				httpErr, ok := err.(*echo.HTTPError)
				assert.Assert(t, ok)
				assert.Assert(t, is.Equal(tc.status, httpErr.Code))
			}
		})
	}
}
