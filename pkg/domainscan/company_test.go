package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestBuildPlatform(t *testing.T) {
	t.Run("nil company returns nil", func(t *testing.T) {
		got := buildPlatform(Enrichment{})

		assert.Check(t, got == nil)
	})

	t.Run("includes optional fields only when present", func(t *testing.T) {
		enrichment := Enrichment{
			Company: &CompanyProfile{
				Name:             "Acme",
				Description:      "Widgets",
				Industry:         "Manufacturing",
				Location:         "Remote",
				EmployeeRange:    "11-50",
				FoundedYear:      "2020",
				EstimatedRevenue: "$1M-$10M",
				SSOSupported:     true,
				SocialLinks:      SocialLinks{GitHub: "https://github.com/acme"},
				StatusPageURL:    "https://status.acme.com",
				Customers:        []string{"Globex"},
			},
		}

		got := buildPlatform(enrichment)

		want := &Platform{
			Name:             "Acme",
			Description:      "Widgets",
			Industry:         "Manufacturing",
			Location:         "Remote",
			EmployeeRange:    "11-50",
			FoundedYear:      "2020",
			EstimatedRevenue: "$1M-$10M",
			SSOSupported:     true,
			SocialLinks:      SocialLinks{GitHub: "https://github.com/acme"},
			StatusPageURL:    "https://status.acme.com",
			Customers:        []string{"Globex"},
			AuthMethods:      []string{"sso"},
		}

		assert.Check(t, is.DeepEqual(want, got))
	})

	t.Run("auth_methods lists every advertised method", func(t *testing.T) {
		enrichment := Enrichment{
			Company: &CompanyProfile{
				Name:                 "Acme",
				SSOSupported:         true,
				SocialLoginSupported: true,
				CredentialsSupported: true,
				PasskeySupported:     true,
			},
		}

		got := buildPlatform(enrichment)

		assert.Check(t, is.DeepEqual([]string{"sso", "social", "credentials", "passkeys"}, got.AuthMethods))
	})

	t.Run("auth_methods empty when none advertised", func(t *testing.T) {
		enrichment := Enrichment{
			Company: &CompanyProfile{Name: "Acme"},
		}

		got := buildPlatform(enrichment)

		assert.Check(t, is.Len(got.AuthMethods, 0))
	})
}

func TestBuildSystems(t *testing.T) {
	t.Run("nil company returns nil", func(t *testing.T) {
		got := buildSystems(Enrichment{})

		assert.Check(t, got == nil)
	})

	t.Run("prefers full description, falls back to summary", func(t *testing.T) {
		enrichment := Enrichment{
			Company: &CompanyProfile{
				Systems: []System{
					{Name: "Widget", Summary: "short", FullDescription: "long"},
					{Name: "Gadget", Summary: "short only"},
				},
			},
		}

		got := buildSystems(enrichment)

		want := []SystemEntry{
			{SystemName: "Widget", Description: "long"},
			{SystemName: "Gadget", Description: "short only"},
		}

		assert.Check(t, is.DeepEqual(want, got))
	})
}

func TestBuildComplianceSection(t *testing.T) {
	t.Run("nil compliance returns nil", func(t *testing.T) {
		got := buildComplianceSection(Enrichment{})

		assert.Check(t, got == nil)
	})

	t.Run("includes optional fields only when present", func(t *testing.T) {
		enrichment := Enrichment{
			Compliance: &CompliancePage{
				Frameworks:          []string{"SOC 2"},
				SOC2Certified:       true,
				Controls:            []string{"MFA enforced"},
				TrustCenterHostedBy: "Vanta",
				Documents:           []TrustDocument{{Name: "SOC 2 Report"}},
			},
		}

		got := buildComplianceSection(enrichment)

		want := &Compliance{
			Frameworks:          []string{"SOC 2"},
			IsSOC2:              true,
			Controls:            []string{"MFA enforced"},
			TrustCenterHostedBy: "Vanta",
			Documents:           []TrustDocument{{Name: "SOC 2 Report"}},
		}

		assert.Check(t, is.DeepEqual(want, got))
	})
}

func TestBuildRegistrar(t *testing.T) {
	t.Run("nil registrar returns nil", func(t *testing.T) {
		got := buildRegistrar(Enrichment{})

		assert.Check(t, got == nil)
	})

	t.Run("includes optional fields only when present", func(t *testing.T) {
		enrichment := Enrichment{
			Registrar: &RegistrarInfo{
				Registrar:   "Cloudflare, Inc.",
				CreatedDate: "2020-01-01T00:00:00Z",
				Nameservers: []string{"abdullah.ns.cloudflare.com"},
				DNSSEC:      true,
			},
		}

		got := buildRegistrar(enrichment)

		want := &Registrar{
			DNSSEC:      true,
			Registrar:   "Cloudflare, Inc.",
			CreatedDate: "2020-01-01T00:00:00Z",
			Nameservers: []string{"abdullah.ns.cloudflare.com"},
		}

		assert.Check(t, is.DeepEqual(want, got))
	})
}
