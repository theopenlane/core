package tokens

import (
	"crypto/subtle"

	"github.com/golang-jwt/jwt/v5"
)

// Validator are able to verify that access and refresh tokens were issued by
// OpenLane and that their claims are valid (e.g. not expired).
type Validator interface {
	// Verify an access or a refresh token after parsing and return its claims
	Verify(tks string) (claims *Claims, err error)

	// Parse an access or refresh token without verifying claims (e.g. to check an expired token)
	Parse(tks string) (claims *Claims, err error)
}

// validator implements the Validator interface, allowing structs in this package to
// embed the validation code base and supply their own keyFunc; unifying functionality
type validator struct {
	audience string
	issuer   string
	keyFunc  jwt.Keyfunc
}

// Verify an access or a refresh token after parsing and return its claims.
func (v *validator) Verify(tks string) (claims *Claims, err error) {
	var token *jwt.Token

	if token, err = jwt.ParseWithClaims(tks, &Claims{}, v.keyFunc); err != nil {
		return nil, err
	}

	var ok bool

	if claims, ok = token.Claims.(*Claims); ok && token.Valid {
		if !claims.VerifyAudience(v.audience, true) {
			return nil, ErrTokenInvalidAudience
		}

		if !claims.VerifyIssuer(v.issuer, true) {
			return nil, ErrTokenInvalidIssuer
		}

		return claims, nil
	}

	return nil, ErrTokenInvalidClaims
}

// Parse an access or refresh token verifying its signature but without verifying its
// claims. This ensures that valid JWT tokens are still accepted but claims can be
// handled on a case-by-case basis; for example by validating an expired access token
// during reauthentication
func (v *validator) Parse(tks string) (claims *Claims, err error) {
	method := GetAlgorithms()
	parser := jwt.NewParser(jwt.WithValidMethods(method), jwt.WithoutClaimsValidation())
	claims = &Claims{}

	if _, err = parser.ParseWithClaims(tks, claims, v.keyFunc); err != nil {
		return nil, err
	}

	return claims, nil
}

func (c *Claims) VerifyAudience(cmp string, req bool) bool {
	return verifyAud(c.Audience, cmp, req)
}

func (c *Claims) VerifyIssuer(cmp string, req bool) bool {
	return verifyIss(c.Issuer, cmp, req)
}

func verifyIss(iss string, cmp string, required bool) bool {
	if iss == "" {
		return !required
	}

	return subtle.ConstantTimeCompare([]byte(iss), []byte(cmp)) != 0
}

func verifyAud(aud []string, cmp string, required bool) bool {
	if len(aud) == 0 {
		return !required
	}
	// use a var here to keep constant time compare when looping over a number of claims
	result := false

	var stringClaims string

	for _, a := range aud {
		if subtle.ConstantTimeCompare([]byte(a), []byte(cmp)) != 0 {
			result = true
		}

		stringClaims += a
	}

	// case where "" is sent in one or many aud claims
	if len(stringClaims) == 0 {
		return !required
	}

	return result
}
