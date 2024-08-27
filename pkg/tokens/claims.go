package tokens

import (
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/utils/ulids"
)

// Claims implements custom claims and extends the `jwt.RegisteredClaims` struct; we will store user-related elements here (and thus in the JWT Token) for reference / validation
type Claims struct {
	jwt.RegisteredClaims
	// UserID is the internal generated mapping ID for the user
	UserID string `json:"user_id,omitempty"`
	// OrgID is the internal generated mapping ID for the organization the JWT token is valid for
	OrgID string `json:"org,omitempty"`
}

// ParseUserID returns the ID of the user from the Subject of the claims
func (c Claims) ParseUserID() ulid.ULID {
	userID, err := ulid.Parse(c.UserID)
	if err != nil {
		return ulids.Null
	}

	return userID
}

// ParseOrgID parses and return the organization ID from the `OrgID` field of the claims
func (c Claims) ParseOrgID() ulid.ULID {
	orgID, err := ulid.Parse(c.OrgID)
	if err != nil {
		return ulids.Null
	}

	return orgID
}
