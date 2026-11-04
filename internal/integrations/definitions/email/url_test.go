package email

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func TestTrustCenterBaseURL_DefaultDomain(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "acme",
	}

	u := trustCenterBaseURL(tc, "trustcenter.example.com")

	assert.Check(t, is.Equal("https", u.Scheme))
	assert.Check(t, is.Equal("trustcenter.example.com", u.Host))
	assert.Check(t, is.Equal("/acme", u.Path))
}

func TestTrustCenterBaseURL_CustomDomain(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "acme",
		Edges: generated.TrustCenterEdges{
			CustomDomain: &generated.CustomDomain{
				CnameRecord: "trust.acme.com",
			},
		},
	}

	u := trustCenterBaseURL(tc, "trustcenter.example.com")

	assert.Check(t, is.Equal("https", u.Scheme))
	assert.Check(t, is.Equal("trust.acme.com", u.Host))
}

func TestTrustCenterBaseURL_CustomDomainWithScheme(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "acme",
		Edges: generated.TrustCenterEdges{
			CustomDomain: &generated.CustomDomain{
				CnameRecord: "https://trust.acme.com",
			},
		},
	}

	u := trustCenterBaseURL(tc, "trustcenter.example.com")

	assert.Check(t, is.Equal("https", u.Scheme))
	// NormalizeHostname strips the scheme, leaving just the hostname
	assert.Check(t, is.Equal("trust.acme.com", u.Host))
}

func TestTrustCenterNDAURL_PathConstruction(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "acme-corp",
	}

	u := trustCenterNDAURL(tc, "trustcenter.example.com")

	assert.Check(t, is.Equal("https", u.Scheme))
	assert.Check(t, is.Equal("trustcenter.example.com", u.Host))
	assert.Check(t, is.Equal("/acme-corp/access/sign-nda", u.Path))
}

func TestTrustCenterNDAURL_EmptySlug(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "",
	}

	u := trustCenterNDAURL(tc, "trustcenter.example.com")

	assert.Check(t, is.Equal("//access/sign-nda", u.Path))
}

func TestTrustCenterNDAURL_CustomDomain(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "secure",
		Edges: generated.TrustCenterEdges{
			CustomDomain: &generated.CustomDomain{
				CnameRecord: "trust.secure.io",
			},
		},
	}

	u := trustCenterNDAURL(tc, "default.example.com")

	assert.Check(t, is.Equal("trust.secure.io", u.Host))
	assert.Check(t, is.Equal("/secure/access/sign-nda", u.Path))
}

func TestTrustCenterNDAURL_FullString(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "myorg",
	}

	u := trustCenterNDAURL(tc, "trustcenter.example.com")

	assert.Check(t, is.Equal("https://trustcenter.example.com/myorg/access/sign-nda", u.String()))
}
