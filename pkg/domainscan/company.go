package domainscan

import (
	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
)

// buildMeta collects scan-level metadata (radar rank, categories, geolocation), returning nil
// when none of it was present in the scan, or result is nil
func buildMeta(result *url_scanner.ScanGetResponse) *Meta {
	if result == nil {
		return nil
	}

	var meta Meta

	if len(result.Meta.Processors.RadarRank.Data) > 0 && result.Meta.Processors.RadarRank.Data[0].Rank > 0 {
		meta.Rank = int(result.Meta.Processors.RadarRank.Data[0].Rank)
	}

	if len(result.Meta.Processors.URLCategories.Data) > 0 {
		urlCategories := make([]string, 0, len(result.Meta.Processors.URLCategories.Data))
		for _, c := range result.Meta.Processors.URLCategories.Data {
			urlCategories = append(urlCategories, c.Name)
		}

		meta.URLCategories = urlCategories
	}

	if len(result.Meta.Processors.DomainCategories.Data) > 0 {
		domainCategories := make([]string, 0, len(result.Meta.Processors.DomainCategories.Data))
		for _, c := range result.Meta.Processors.DomainCategories.Data {
			domainCategories = append(domainCategories, c.Name)
		}

		meta.DomainCategories = domainCategories
	}

	if len(result.Meta.Processors.Geoip.Data) > 0 {
		geo := result.Meta.Processors.Geoip.Data[0]

		meta.Geolocation = &Geolocation{
			City:        geo.Geoip.City,
			Country:     geo.Geoip.Country,
			CountryName: geo.Geoip.CountryName,
			Region:      geo.Geoip.Region,
			Latitude:    geo.Geoip.Ll[0],
			Longitude:   geo.Geoip.Ll[1],
		}
	}

	if meta.IsEmpty() {
		return nil
	}

	return &meta
}

// buildPlatform shapes the scanned company's profile, using field names that mirror Openlane's Platform object (name/description)
func buildPlatform(enrichment Enrichment) *Platform {
	company := enrichment.Company
	if company == nil {
		return nil
	}

	platform := Platform{
		Name:             company.Name,
		Description:      company.Description,
		Industry:         company.Industry,
		Location:         company.Location,
		EmployeeRange:    company.EmployeeRange,
		FoundedYear:      company.FoundedYear,
		EstimatedRevenue: company.EstimatedRevenue,
		SSOSupported:     company.SSOSupported,
		MFASupported:     company.MFASupported,
		SocialLinks:      company.SocialLinks,
		StatusPageURL:    company.StatusPageURL,
		Customers:        company.Customers,
		ProvidedServices: company.ProvidedServices,
		AuthMethods:      authMethods(company),
	}

	return &platform
}

// buildSystems shapes each of the company's systems, using field names that
// mirror Openlane's SystemDetail object (system_name/description)
func buildSystems(enrichment Enrichment) []SystemEntry {
	if enrichment.Company == nil {
		return nil
	}

	systems := make([]SystemEntry, 0, len(enrichment.Company.Systems))

	for _, s := range enrichment.Company.Systems {
		description := s.FullDescription
		if description == "" {
			description = s.Summary
		}

		systems = append(systems, SystemEntry{SystemName: s.Name, Description: description})
	}

	return systems
}

// buildComplianceSection reports the company's compliance posture gathered by the domainscan enrichment
func buildComplianceSection(enrichment Enrichment) *Compliance {
	compliance := enrichment.Compliance
	if compliance == nil {
		return nil
	}

	section := Compliance{
		Frameworks:          compliance.Frameworks,
		IsSOC2:              compliance.SOC2Certified,
		Controls:            compliance.Controls,
		TrustCenterHostedBy: compliance.TrustCenterHostedBy,
		Documents:           compliance.Documents,
	}

	return &section
}

// authMethod names one authentication method a company's product advertises support for
type authMethod string

const (
	authMethodSSO         authMethod = "sso"
	authMethodSocial      authMethod = "social"
	authMethodCredentials authMethod = "credentials"
	authMethodPasskeys    authMethod = "passkeys"
)

// authMethods reports which authentication methods the company's product advertises support
// for, from the LLM-derived CompanyProfile booleans
func authMethods(company *CompanyProfile) []string {
	var methods []string

	if company.SSOSupported {
		methods = append(methods, string(authMethodSSO))
	}

	if company.SocialLoginSupported {
		methods = append(methods, string(authMethodSocial))
	}

	if company.CredentialsSupported {
		methods = append(methods, string(authMethodCredentials))
	}

	if company.PasskeySupported {
		methods = append(methods, string(authMethodPasskeys))
	}

	return methods
}
