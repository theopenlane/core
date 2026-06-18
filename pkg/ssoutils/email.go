package ssoutils

import "strings"

// EmailDomain returns the domain portion of an email address (e.g. "theopenlane.io" from "user@theopenlane.io")
func EmailDomain(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return ""
	}

	return email[at+1:]
}
