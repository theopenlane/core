package microsoft

import "github.com/golang-jwt/jwt/v5"

// TokenClaims returns the claims of a Microsoft-issued JWT without verifying it
func TokenClaims(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	if _, _, err := new(jwt.Parser).ParseUnverified(token, claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// TokenClaim returns a string claim from a Microsoft-issued JWT without verifying it, empty when absent
func TokenClaim(token, name string) string {
	claims, err := TokenClaims(token)
	if err != nil {
		return ""
	}

	value, _ := claims[name].(string)

	return value
}
