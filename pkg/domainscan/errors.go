package domainscan

import "errors"

var (
	errBrandingBotChallenge = errors.New("Branding extraction blocked by the website's bot protection.")
	// ErrInvalidDomain is returned when its not able to determine the apex domain
	ErrInvalidDomain = errors.New("could not determine domain")
)
