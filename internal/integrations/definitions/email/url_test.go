package email

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/generated"
)

func TestTrustCenterBaseURL_DefaultDomain(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "acme",
	}

	u := trustCenterBaseURL(tc, "trustcenter.example.com")

	assert.Equal(t, "https", u.Scheme)
	assert.Equal(t, "trustcenter.example.com", u.Host)
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

	assert.Equal(t, "https", u.Scheme)
	assert.Equal(t, "trust.acme.com", u.Host)
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

	assert.Equal(t, "https", u.Scheme)
	// NormalizeHostname strips the scheme, leaving just the hostname
	assert.Equal(t, "trust.acme.com", u.Host)
}

func TestTrustCenterNDAURL_PathConstruction(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "acme-corp",
	}

	u := trustCenterNDAURL(tc, "trustcenter.example.com")

	assert.Equal(t, "https", u.Scheme)
	assert.Equal(t, "trustcenter.example.com", u.Host)
	assert.Equal(t, "/acme-corp/access/sign-nda", u.Path)
}

func TestTrustCenterNDAURL_EmptySlug(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "",
	}

	u := trustCenterNDAURL(tc, "trustcenter.example.com")

	assert.Equal(t, "//access/sign-nda", u.Path)
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

	assert.Equal(t, "trust.secure.io", u.Host)
	assert.Equal(t, "/secure/access/sign-nda", u.Path)
}

func TestTrustCenterNDAURL_FullString(t *testing.T) {
	tc := &generated.TrustCenter{
		Slug: "myorg",
	}

	u := trustCenterNDAURL(tc, "trustcenter.example.com")

	assert.Equal(t, "https://trustcenter.example.com/myorg/access/sign-nda", u.String())
}
