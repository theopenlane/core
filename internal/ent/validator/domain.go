package validator

import (
	"net/url"
	"regexp"
	"slices"

	"github.com/theopenlane/utils/rout"
)

var (
	validSchemes = []string{"http", "https"}
	domainMaxLen = 255
	urlMaxLen    = 2048
	domainRegexp = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)
)

// ValidateDomains validates a list of domains and returns an error if any of them are invalid
func ValidateDomains() func(domains []string) error {
	return func(domains []string) error {
		for _, domain := range domains {
			// ensure the domain is not too long
			if len(domain) > domainMaxLen || len(domain) == 0 {
				return rout.InvalidField("domains")
			}

			if err := validateURL(domain); err != nil {
				return rout.InvalidField("domains")
			}
		}

		return nil
	}
}

// ValidateURL validates a url and returns an error if it is invalid
func ValidateURL() func(u string) error {
	return func(u string) error {
		// ensure the domain is not too long
		if len(u) > urlMaxLen || len(u) == 0 {
			return rout.InvalidField("url")
		}

		// parse the url
		if err := validateURL(u); err != nil {
			return rout.InvalidField("url")
		}

		return nil
	}
}

func validateURL(inputURL string) error {
	// parse the url
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return rout.InvalidField("url")
	}

	// if the scheme is empty, add http:// to the domain and try again
	if parsedURL.Scheme == "" {
		parsedURL, err = url.Parse("http://" + inputURL)
		if err != nil {
			return rout.InvalidField("url")
		}
	}

	// ensure the host is not empty
	if parsedURL.Host == "" {
		return rout.InvalidField("url")
	}

	// only allow http and https schemes
	if parsedURL.Scheme != "" && !slices.Contains(validSchemes, parsedURL.Scheme) {
		return rout.InvalidField("url")
	}

	// ensure the host is a valid domain
	valid := domainRegexp.MatchString(parsedURL.Host)
	if !valid {
		return rout.InvalidField("url")
	}

	return nil
}
