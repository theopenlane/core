package tokens

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/theopenlane/utils/ulids"
)

const (
	nonceLength                 = 64
	keyLength                   = 64
	expirationDays              = 7
	resetTokenExpirationMinutes = 15
	inviteExpirationDays        = 14
)

// NewVerificationToken creates a token struct from an email address that expires
// in 7 days
func NewVerificationToken(email string) (token *VerificationToken, err error) {
	if email == "" {
		return nil, ErrMissingEmail
	}

	token = &VerificationToken{
		Email: email,
	}

	if token.SigningInfo, err = NewSigningInfo(time.Hour * 24 * expirationDays); err != nil {
		return nil, err
	}

	return token, nil
}

// VerificationToken packages an email address with random data and an expiration
// time so that it can be serialized and hashed into a token which can be sent to users
type VerificationToken struct {
	Email string `msgpack:"email"`
	SigningInfo
}

// Sign creates a base64 encoded string from the token data so that it can be sent to
// users as part of a URL. The returned secret should be stored in the database so that
// the string can be recomputed when verifying a user provided token.
func (t *VerificationToken) Sign() (string, []byte, error) {
	data, err := msgpack.Marshal(t)
	if err != nil {
		return "", nil, err
	}

	return t.signData(data)
}

// Verify checks that a token was signed with the secret and is not expired
func (t *VerificationToken) Verify(signature string, secret []byte) (err error) {
	if t.Email == "" {
		return ErrTokenMissingEmail
	}

	if t.IsExpired() {
		return ErrTokenExpired
	}

	if len(secret) != nonceLength+keyLength {
		return ErrInvalidSecret
	}

	// Serialize the struct with the nonce from the secret
	t.Nonce = secret[0:nonceLength]

	var data []byte

	if data, err = msgpack.Marshal(t); err != nil {
		return err
	}

	return t.verifyData(data, signature, secret)
}

// NewResetToken creates a token struct from a user ID that expires in 15 minutes
func NewResetToken(id ulid.ULID) (token *ResetToken, err error) {
	if ulids.IsZero(id) {
		return nil, ErrMissingUserID
	}

	token = &ResetToken{
		UserID: id,
	}

	if token.SigningInfo, err = NewSigningInfo(time.Minute * resetTokenExpirationMinutes); err != nil {
		return nil, err
	}

	return token, nil
}

// ResetToken packages a user ID with random data and an expiration time so that it can
// be serialized and hashed into a token which can be sent to users
type ResetToken struct {
	UserID ulid.ULID `msgpack:"user_id"`
	SigningInfo
}

// Sign creates a base64 encoded string from the token data so that it can be sent to
// users as part of a URL. The returned secret should be stored in the database so that
// the string can be recomputed when verifying a user provided token
func (t *ResetToken) Sign() (string, []byte, error) {
	data, err := msgpack.Marshal(t)
	if err != nil {
		return "", nil, err
	}

	return t.signData(data)
}

// Verify checks that a token was signed with the secret and is not expired
func (t *ResetToken) Verify(signature string, secret []byte) (err error) {
	if ulids.IsZero(t.UserID) {
		return ErrTokenMissingUserID
	}

	if t.IsExpired() {
		return ErrTokenExpired
	}

	if len(secret) != nonceLength+keyLength {
		return ErrInvalidSecret
	}

	// Serialize the struct with the nonce from the secret
	t.Nonce = secret[0:nonceLength]

	var data []byte

	if data, err = msgpack.Marshal(t); err != nil {
		return err
	}

	return t.verifyData(data, signature, secret)
}

// NewSigningInfo creates new signing info with a time expiration
func NewSigningInfo(expires time.Duration) (info SigningInfo, err error) {
	if expires == 0 {
		return info, ErrExpirationIsRequired
	}

	info = SigningInfo{
		ExpiresAt: time.Now().UTC().Add(expires).Truncate(time.Microsecond),
		Nonce:     make([]byte, nonceLength),
	}

	if _, err = rand.Read(info.Nonce); err != nil {
		return info, ErrFailedSigning
	}

	return info, nil
}

// SigningInfo contains an expiration time and a nonce that is used to sign the token
type SigningInfo struct {
	ExpiresAt time.Time `msgpack:"expires_at"`
	Nonce     []byte    `msgpack:"nonce"`
}

func (d SigningInfo) IsExpired() bool {
	return d.ExpiresAt.Before(time.Now())
}

// Create a signature from raw data and a nonce. The resulting signature is safe to be used in a URL
func (d SigningInfo) signData(data []byte) (_ string, secret []byte, err error) {
	// Compute hash with a random 64 byte key
	key := make([]byte, keyLength)
	if _, err = rand.Read(key); err != nil {
		return "", nil, err
	}

	mac := hmac.New(sha256.New, key)
	if _, err = mac.Write(data); err != nil {
		return "", nil, err
	}

	// Include the nonce with the key so that the token can be reconstructed later
	secret = make([]byte, nonceLength+keyLength)
	copy(secret[0:nonceLength], d.Nonce)
	copy(secret[nonceLength:], key)

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), secret, nil
}

// Verify data using the signature and secret
func (d SigningInfo) verifyData(data []byte, signature string, secret []byte) (err error) {
	// Compute hash to verify the user token
	mac := hmac.New(sha256.New, secret[nonceLength:])
	if _, err = mac.Write(data); err != nil {
		return err
	}

	// Decode the user token
	var token []byte

	if token, err = base64.RawURLEncoding.DecodeString(signature); err != nil {
		return err
	}

	// Check if the recomputed token matches the user token
	if !hmac.Equal(mac.Sum(nil), token) {
		return ErrTokenInvalid
	}

	return nil
}

// NewOrgInvitationToken creates a token struct from an email address that expires
// in 14 days
func NewOrgInvitationToken(email string, orgID ulid.ULID) (token *OrgInviteToken, err error) {
	if email == "" {
		return nil, ErrInviteTokenMissingEmail
	}

	if ulids.IsZero(orgID) {
		return nil, ErrInviteTokenMissingOrgID
	}

	token = &OrgInviteToken{
		Email: email,
		OrgID: orgID,
	}

	if token.SigningInfo, err = NewSigningInfo(time.Hour * 24 * inviteExpirationDays); err != nil {
		return nil, err
	}

	return token, nil
}

// OrgInviteToken packages an email address with random data and an expiration
// time so that it can be serialized and hashed into a token which can be sent to users
type OrgInviteToken struct {
	Email string    `msgpack:"email"`
	OrgID ulid.ULID `msgpack:"organization_id"`
	SigningInfo
}

// Sign creates a base64 encoded string from the token data so that it can be sent to
// users as part of a URL. The returned secret should be stored in the database so that
// the string can be recomputed when verifying a user provided token.
func (t *OrgInviteToken) Sign() (string, []byte, error) {
	data, err := msgpack.Marshal(t)
	if err != nil {
		return "", nil, err
	}

	return t.signData(data)
}

// Verify checks that a token was signed with the secret and is not expired
func (t *OrgInviteToken) Verify(signature string, secret []byte) (err error) {
	if t.Email == "" {
		return ErrInviteTokenMissingEmail
	}

	if ulids.IsZero(t.OrgID) {
		return ErrInviteTokenMissingOrgID
	}

	if t.IsExpired() {
		return ErrTokenExpired
	}

	if len(secret) != nonceLength+keyLength {
		return ErrInvalidSecret
	}

	// Serialize the struct with the nonce from the secret
	t.Nonce = secret[0:nonceLength]

	var data []byte

	if data, err = msgpack.Marshal(t); err != nil {
		return err
	}

	return t.verifyData(data, signature, secret)
}
