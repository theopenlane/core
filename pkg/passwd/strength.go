package passwd

import (
	"strings"
	"unicode"
)

type PasswordStrength uint8

const (
	Weak PasswordStrength = iota
	Poor
	Fair
	Moderate
	Strong
	Excellent
)

// See: https://nordpass.com/most-common-passwords-list/
var invalidPasswords = map[string]struct{}{
	"password":    {},
	"12345678":    {},
	"123456789":   {},
	"1234567890":  {},
	"12345678910": {},
	"987654321":   {},
	"qweqrtyu":    {},
	"qwerty":      {},
	"azerty":      {},
	"guest":       {},
	"abcd1234":    {},
	"iloveyou":    {},
	"col123456":   {},
	"110110jp":    {},
	"groupd2013":  {},
	"liman1000":   {},
	"123123123":   {},
	"9136668099":  {},
	"11111111":    {},
	"1qaz2wsx":    {},
	"password1":   {},
	"luzit2000":   {},
	"asdfghjkl":   {},
	"football":    {},
	"samsung":     {},
	"qazwsxedc":   {},
}

// Strength is currently a very simple password strength algorithm that simply checks
// the length and contents of a password to ensure that reasonable passwords are added.
// In the future this algorithm can be strengthened with zxcvbn algorithms.
// TODO: implement dictionary word, spatial closeness, and l33t strength algorithms.
// See: https://nulab.com/learn/software-development/password-strength/
func Strength(password string) PasswordStrength {
	if len(password) < 8 || isInvalid(password) {
		return Weak
	}

	strength := Poor
	if len(password) >= 12 { //nolint:mnd
		strength++
	}

	if hasUpper(password) && hasLower(password) {
		strength++
	}

	if strings.ContainsAny(password, "0123456789") {
		strength++
	}

	if strings.ContainsAny(password, " !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~") {
		strength++
	}

	return strength
}

func isInvalid(password string) bool {
	password = strings.TrimSpace(strings.ToLower(password))
	_, ok := invalidPasswords[password]

	return ok
}

func hasUpper(password string) bool {
	for _, r := range password {
		if unicode.IsUpper(r) {
			return true
		}
	}

	return false
}

func hasLower(password string) bool {
	for _, r := range password {
		if unicode.IsLower(r) {
			return true
		}
	}

	return false
}
