package keygen

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idxbits  = 6
	idxmask  = 1<<idxbits - 1
	idxmax   = 63 / idxbits
)

// Defaults for the length of key IDs and secrets
const (
	KeyIDLength  = 32
	SecretLength = 64
	ByteLength   = 8
)

// Alpha generates a random string of n characters that only includes upper and
// lowercase letters (no symbols or digits)
func Alpha(n int) string {
	return generate(n, alphabet)
}

// AlphaNumeric generates a random string of n characters that includes upper and
// lowercase letters and the digits 0-9
func AlphaNumeric(n int) string {
	return generate(n, alphanum)
}

// KeyID returns a random ID that is of a fixed length with only alpha characters
func KeyID() string {
	return Alpha(KeyIDLength)
}

// Secret returns a random string of a fixed length with alpha-numeric characters
func Secret() string {
	return AlphaNumeric(SecretLength)
}

// PrefixedSecret returns a prefixed random string of a fixed length with alpha-numeric characters
func PrefixedSecret(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, Secret())
}

// generate is a helper function to create a random string of n characters from the
// character set defined by chars. It uses as efficient a method of generation as
// possible, using a string builder to prevent multiple allocations and a 6 bit mask
// to select 10 random letters at a time to add to the string. This method would be far
// faster if it used math/rand src and the Int63() function, but for API key generation
// it is important to use a cryptographically random generator.
//
// See: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func generate(n int, chars string) string {
	sb := strings.Builder{}
	sb.Grow(n)

	for i, cache, remain := n-1, CryptoRandInt(), idxmax; i >= 0; {
		if remain == 0 {
			cache, remain = CryptoRandInt(), idxmax
		}

		if idx := int(cache & idxmask); idx < len(chars) { // nolint:gosec
			sb.WriteByte(chars[idx])

			i--
		}

		cache >>= idxbits
		remain--
	}

	return sb.String()
}

// CryptoRandInt function generates a random 64-bit unsigned integer using cryptographic methods
func CryptoRandInt() uint64 {
	buf := make([]byte, ByteLength)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("cannot generate random number: %w", err))
	}

	return binary.BigEndian.Uint64(buf)
}

// HashFromBytes returns a SHA-256 checksum of the input
func HashFromBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return fmt.Sprintf("%x", sum)
}

// Hash returns a SHA-256 checksum of a string
func Hash(value string) string {
	return HashFromBytes([]byte(value))
}

// GenerateRandomBytes returns random bytes
func GenerateRandomBytes(size int) []byte {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

// GenerateRandomString returns a random string
func GenerateRandomString(size int) string {
	return base64.URLEncoding.EncodeToString(GenerateRandomBytes(size))
}

// GenerateRandomStringHex returns a random hexadecimal string
func GenerateRandomStringHex(size int) string {
	return hex.EncodeToString(GenerateRandomBytes(size))
}

// HashInput function takes an input and generates a bcrypt hash
func HashInput(input string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(input), bcrypt.DefaultCost)

	return string(bytes), err
}

// GenerateSHA256Hmac generates a SHA-256 HMAC by using the secret as the key and
// the data as the message
func GenerateSHA256Hmac(secret string, data []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}
