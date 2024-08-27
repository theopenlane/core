package validator

import (
	"net/url"
	"regexp"
	"slices"

	"github.com/theopenlane/utils/rout"
)

var (
	validSchemes = []string{"http", "https"}
	urlMaxLen    = 255
	domainRegexp = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)
)

// ValidateDomains validates a list of domains and returns an error if any of them are invalid
func ValidateDomains() func(domains []string) error {
	return func(domains []string) error {
		for _, domain := range domains {
			// ensure the domain is not too long
			if len(domain) > urlMaxLen || len(domain) == 0 {
				return rout.InvalidField("domains")
			}

			// parse the domain
			u, err := url.Parse(domain)
			if err != nil {
				return rout.InvalidField("domains")
			}

			// if the scheme is empty, add http:// to the domain and try again
			if u.Scheme == "" {
				u, err = url.Parse("http://" + domain)
				if err != nil {
					return rout.InvalidField("domains")
				}
			}

			// ensure the host is not empty
			if u.Host == "" {
				return rout.InvalidField("domains")
			}

			// only allow http and https schemes
			if u.Scheme != "" && !slices.Contains(validSchemes, u.Scheme) {
				return rout.InvalidField("domains")
			}

			// ensure the host is a valid domain
			valid := domainRegexp.MatchString(u.Host)
			if !valid {
				return rout.InvalidField("domains")
			}
		}

		return nil
	}
}
