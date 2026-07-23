package domainscan

// mergeStrings unions two string slices, dropping empty values and duplicates
// while preserving first-seen order
func mergeStrings(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	merged := make([]string, 0, len(a)+len(b))

	for _, s := range append(append([]string{}, a...), b...) {
		if s == "" {
			continue
		}

		if _, ok := seen[s]; ok {
			continue
		}

		seen[s] = struct{}{}
		merged = append(merged, s)
	}

	return merged
}

// mergeCompanyProfiles combines the results of probing a company's homepage and its common marketing subpaths
// into a single CompanyProfile, preferring the first non-empty scalar value seen and unioning list/boolean fields
// so a detail mentioned on only one page (e.g. SSO on /pricing) still surfaces
func mergeCompanyProfiles(pages ...*CompanyProfile) *CompanyProfile {
	merged := &CompanyProfile{}

	var systems []System

	seenSystems := map[string]struct{}{}

	for _, p := range pages {
		if p == nil {
			continue
		}

		if merged.Name == "" {
			merged.Name = p.Name
		}

		if merged.Description == "" {
			merged.Description = p.Description
		}

		if merged.Industry == "" {
			merged.Industry = p.Industry
		}

		if merged.Location == "" {
			merged.Location = p.Location
		}

		if merged.EmployeeRange == "" {
			merged.EmployeeRange = p.EmployeeRange
		}

		if merged.FoundedYear == "" {
			merged.FoundedYear = p.FoundedYear
		}

		if merged.EstimatedRevenue == "" {
			merged.EstimatedRevenue = p.EstimatedRevenue
		}

		if merged.StatusPageURL == "" {
			merged.StatusPageURL = p.StatusPageURL
		}

		merged.SocialLinks = mergeSocialLinks(merged.SocialLinks, p.SocialLinks)

		for _, s := range p.Systems {
			if s.Name == "" {
				continue
			}

			if _, ok := seenSystems[s.Name]; ok {
				continue
			}

			seenSystems[s.Name] = struct{}{}
			systems = append(systems, s)
		}

		merged.Customers = mergeStrings(merged.Customers, p.Customers)
		merged.Technologies = mergeStrings(merged.Technologies, p.Technologies)
		merged.ProvidedServices = mergeStrings(merged.ProvidedServices, p.ProvidedServices)
		merged.SubdomainLinks = mergeStrings(merged.SubdomainLinks, p.SubdomainLinks)
		merged.SSOSupported = merged.SSOSupported || p.SSOSupported
		merged.MFASupported = merged.MFASupported || p.MFASupported
		merged.SocialLoginSupported = merged.SocialLoginSupported || p.SocialLoginSupported
		merged.CredentialsSupported = merged.CredentialsSupported || p.CredentialsSupported
		merged.PasskeySupported = merged.PasskeySupported || p.PasskeySupported
	}

	merged.Systems = systems

	return merged
}

// mergeSocialLinks fills in any empty field of a from b, preferring a's existing values
func mergeSocialLinks(a, b SocialLinks) SocialLinks {
	if a.LinkedIn == "" {
		a.LinkedIn = b.LinkedIn
	}

	if a.Twitter == "" {
		a.Twitter = b.Twitter
	}

	if a.GitHub == "" {
		a.GitHub = b.GitHub
	}

	if a.Discord == "" {
		a.Discord = b.Discord
	}

	if a.Instagram == "" {
		a.Instagram = b.Instagram
	}

	if a.YouTube == "" {
		a.YouTube = b.YouTube
	}

	if a.Facebook == "" {
		a.Facebook = b.Facebook
	}

	return a
}
