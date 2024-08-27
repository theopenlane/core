package passwd

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/crypto/argon2"
)

// ===========================================================================
// Derived Key Algorithm
// ===========================================================================

// Argon2 constants for the derived key (dk) algorithm
// See: https://cryptobook.nakov.com/mac-and-key-derivation/argon2
const (
	dkAlg  = "argon2id"        // the derived key algorithm
	dkTime = uint32(1)         // draft RFC recommends time = 1
	dkMem  = uint32(64 * 1024) // draft RFC recommends memory as ~64MB (or as much as possible)
	dkProc = uint8(2)          // can be set to the number of available CPUs
	dkSLen = 16                // the length of the salt to generate per user
	dkKLen = uint32(32)        // the length of the derived key (32 bytes is the required key size for AES-256)
)

// Argon2 variables for the derived key (dk) algorithm
var (
	dkParse = regexp.MustCompile(`^\$(?P<alg>[\w\d]+)\$v=(?P<ver>\d+)\$m=(?P<mem>\d+),t=(?P<time>\d+),p=(?P<procs>\d+)\$(?P<salt>[\+\/\=a-zA-Z0-9]+)\$(?P<key>[\+\/\=a-zA-Z0-9]+)$`)
)

// CreateDerivedKey creates an encoded derived key with a random hash for the password.
func CreateDerivedKey(password string) (string, error) {
	if password == "" {
		return "", ErrCannotCreateDK
	}

	salt := make([]byte, dkSLen)
	if _, err := rand.Read(salt); err != nil {
		return "", ErrCouldNotGenerate
	}

	dk := argon2.IDKey([]byte(password), salt, dkTime, dkMem, dkProc, dkKLen)
	b64salt := base64.StdEncoding.EncodeToString(salt)
	b64dk := base64.StdEncoding.EncodeToString(dk)

	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s", dkAlg, argon2.Version, dkMem, dkTime, dkProc, b64salt, b64dk), nil
}

// VerifyDerivedKey checks that the submitted password matches the derived key.
func VerifyDerivedKey(dk, password string) (bool, error) {
	if dk == "" || password == "" {
		return false, ErrUnableToVerify
	}

	dkb, salt, t, m, p, err := ParseDerivedKey(dk)
	if err != nil {
		return false, err
	}

	vdk := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(dkb))) // nolint:gosec

	return bytes.Equal(dkb, vdk), nil
}

// ParseDerivedKey returns the parts of the encoded derived key string.
func ParseDerivedKey(encoded string) (dk, salt []byte, time, memory uint32, threads uint8, err error) {
	if !dkParse.MatchString(encoded) {
		return nil, nil, 0, 0, 0, ErrCannotParseDK
	}

	parts := dkParse.FindStringSubmatch(encoded)

	if len(parts) != 8 { //nolint:mnd
		return nil, nil, 0, 0, 0, ErrCannotParseEncodedEK
	}

	// check the algorithm
	if parts[1] != dkAlg {
		return nil, nil, 0, 0, 0, newParseError("dkAlg", parts[1], dkAlg)
	}

	// check the version
	if version, err := strconv.Atoi(parts[2]); err != nil || version != argon2.Version {
		return nil, nil, 0, 0, 0, newParseError("version", parts[2], fmt.Sprintf("%d", argon2.Version))
	}

	var (
		time64    uint64
		memory64  uint64
		threads64 uint64
	)

	if memory64, err = strconv.ParseUint(parts[3], 10, 32); err != nil {
		return nil, nil, 0, 0, 0, newParseError("memory", parts[3], err.Error())
	}

	memory = uint32(memory64) // nolint:gosec

	if time64, err = strconv.ParseUint(parts[4], 10, 32); err != nil {
		return nil, nil, 0, 0, 0, newParseError("time", parts[4], err.Error())
	}

	time = uint32(time64) // nolint:gosec

	if threads64, err = strconv.ParseUint(parts[5], 10, 8); err != nil {
		return nil, nil, 0, 0, 0, newParseError("threads", parts[5], err.Error())
	}

	threads = uint8(threads64) // nolint:gosec

	if salt, err = base64.StdEncoding.DecodeString(parts[6]); err != nil {
		return nil, nil, 0, 0, 0, newParseError("salt", parts[6], err.Error())
	}

	if dk, err = base64.StdEncoding.DecodeString(parts[7]); err != nil {
		return nil, nil, 0, 0, 0, newParseError("dk", parts[7], err.Error())
	}

	return dk, salt, time, memory, threads, nil
}

func IsDerivedKey(s string) bool {
	return dkParse.MatchString(s)
}
