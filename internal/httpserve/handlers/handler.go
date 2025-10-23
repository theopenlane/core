package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type User struct {
	ID    string
	Email string
	Name  string
}

type Reply struct {
	Success bool `json:"success"`
}

type SSOStatusReply struct {
	Reply
	Enforced       bool   `json:"enforced"`
	OrganizationID string `json:"organization_id"`
	IsOrgOwner     bool   `json:"is_org_owner"`
}

type DB interface {
	GetUserByEmail(email string) (User, error)
	GetUserDefaultOrgID(userID string) (string, error)
	FetchSSOStatus(orgID, userID string) (SSOStatusReply, error)
}

type Handler struct {
	DB DB
}

// WebfingerHandler example
func (h *Handler) WebfingerHandler(c echo.Context) error {
	resource := c.QueryParam("resource")
	if !strings.HasPrefix(resource, "acct:") {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid resource")
	}

	email := strings.TrimPrefix(resource, "acct:")
	user, err := h.DB.GetUserByEmail(email)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	orgID, err := h.DB.GetUserDefaultOrgID(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "org not found")
	}

	sso, _ := h.DB.FetchSSOStatus(orgID, user.ID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"subject": resource,
		"links": []map[string]interface{}{
			{
				"rel":  "http://openid.net/specs/connect/1.0/issuer",
				"href": "https://example.com",
				"sso":  sso,
			},
		},
	})
}
