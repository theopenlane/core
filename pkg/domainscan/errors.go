package domainscan

import "errors"

var (
	errBrandingBotChallenge = errors.New("branding extraction blocked by the website's bot protection")
	errBrandingCDP          = errors.New("branding CDP")
	errBrandingProbe        = errors.New("branding probe failed")
	errBrandingProbeNoValue = errors.New("branding probe returned no value")
	errBrandingNavigation   = errors.New("branding navigation failed")
	errBrandingNoDocument   = errors.New("branding navigation did not load a document")
	errBrandingNoWebsite    = errors.New("branding navigation did not reach a website")
	// ErrInvalidDomain is returned when its not able to determine the apex domain
	ErrInvalidDomain = errors.New("could not determine domain")
)
