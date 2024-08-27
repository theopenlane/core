package tokens

import (
	"context"
	"fmt"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// JWKSValidator provides public verification that JWT tokens have been issued by the
// authentication service by checking that the tokens have been signed using
// public keys from a JSON Web Key Set (JWKS). The validator then returns
// specific claims if the token is in fact valid.
type JWKSValidator struct {
	validator
	keys jwk.Set
}

// NewJWKSValidator is a constructor for creating a new instance of the `JWKSValidator`
// struct. It takes in a `jwk.Set` containing the JSON Web Key Set (JWKS), as well as the audience and issuer strings.
// It initializes a new `JWKSValidator` with the provided JWKS, audience, and issuer
func NewJWKSValidator(keys jwk.Set, audience, issuer string) *JWKSValidator {
	validator := &JWKSValidator{
		validator: validator{
			audience: audience,
			issuer:   issuer,
		},
		keys: keys,
	}
	validator.validator.keyFunc = validator.keyFunc

	return validator
}

// keyFunc is a jwt.KeyFunc that selects the RSA public key from the list of managed
// internal keys based on the kid in the token header
func (v *JWKSValidator) keyFunc(token *jwt.Token) (publicKey interface{}, err error) {
	// Fetch the kid from the header
	kid, ok := token.Header["kid"]
	if !ok {
		return nil, ErrTokenMissingKid
	}

	key, found := v.keys.LookupKeyID(kid.(string))
	if !found {
		return nil, ErrUnknownSigningKey
	}

	// Per JWT security notice: do not forget to validate alg is expected
	if token.Method.Alg() != key.Algorithm().String() {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"]) //nolint:err113
	}

	// Extract the raw public key from the key material and return it.
	if err = key.Raw(&publicKey); err != nil {
		return nil, fmt.Errorf("could not extract raw key: %w", err)
	}

	return publicKey, nil
}

// CachedJWKSValidator struct is a type that extends the functionality of the `JWKSValidator`
// struct. It adds caching capabilities to the JWKS validation process. It includes
// a `cache` field of type `*jwk.Cache` to store and retrieve the JWKS, an `endpoint` field to
// specify the endpoint from which to fetch the JWKS, and embeds the `JWKSValidator` struct to
// inherit its methods and fields. The `CachedJWKSValidator` struct also includes additional methods
// `Refresh` and`keyFunc` to handle the caching logic
type CachedJWKSValidator struct {
	JWKSValidator
	cache    *jwk.Cache
	endpoint string
}

// NewCachedJWKSValidator function is a constructor for creating a new instance of the
// `CachedJWKSValidator` struct. It takes in a `context.Context`, a `*jwk.Cache`, an endpoint string,
// an audience string, and an issuer string
func NewCachedJWKSValidator(ctx context.Context, cache *jwk.Cache, endpoint, audience, issuer string) (validator *CachedJWKSValidator, err error) {
	validator = &CachedJWKSValidator{
		cache:    cache,
		endpoint: endpoint,
	}

	var keys jwk.Set

	if keys, err = cache.Refresh(ctx, endpoint); err != nil {
		return nil, err
	}

	validator.JWKSValidator = *NewJWKSValidator(keys, audience, issuer)
	validator.validator.keyFunc = validator.keyFunc

	return validator, nil
}

// Refresh method in the `CachedJWKSValidator` struct is responsible for refreshing the JWKS
// (JSON Web Key Set) cache. It takes in a `context.Context` as a parameter and returns an error if
// the refresh process fails
func (v *CachedJWKSValidator) Refresh(ctx context.Context) (err error) {
	if v.JWKSValidator.keys, err = v.cache.Refresh(ctx, v.endpoint); err != nil {
		return fmt.Errorf("could not refresh cache from %s: %w", v.endpoint, err)
	}

	return nil
}

// keyFunc method in the `CachedJWKSValidator` struct is responsible for retrieving the public
// key from the JWKS cache based on the `kid` (key ID) in the token header. It first retrieves the
// JWKS from the cache using the `cache.Get` method. Then, it calls the `keyFunc` method of the embedded `JWKSValidator` struct to perform the actual key retrieval and validation. If the JWKS
// cannot be retrieved from the cache, an error is returned
func (v *CachedJWKSValidator) keyFunc(token *jwt.Token) (publicKey interface{}, err error) {
	if v.JWKSValidator.keys, err = v.cache.Get(context.Background(), v.endpoint); err != nil {
		return nil, fmt.Errorf("could not retrieve JWKS from cache: %w", err)
	}

	return v.JWKSValidator.keyFunc(token)
}
