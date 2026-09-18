package domainscan

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestParseSecurityTxt(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		wantPresent    bool
		wantContacts   []string
		wantPolicy     string
		wantLanguages  []string
		wantHasExpires bool
		wantExpired    bool
	}{
		{
			name:        "empty body is not present",
			body:        "",
			wantPresent: false,
		},
		{
			name:        "body without a contact field is not present",
			body:        "Policy: https://example.com/policy\nExpires: 2099-01-01T00:00:00Z",
			wantPresent: false,
		},
		{
			name: "full file parses every field",
			body: `# comment line
Contact: mailto:security@example.com
Contact: https://example.com/report
Encryption: https://example.com/pgp.txt
Policy: https://example.com/security-policy
Preferred-Languages: en, fr
Expires: 2099-09-09T00:00:00Z`,
			wantPresent:    true,
			wantContacts:   []string{"mailto:security@example.com", "https://example.com/report"},
			wantPolicy:     "https://example.com/security-policy",
			wantLanguages:  []string{"en", "fr"},
			wantHasExpires: true,
			wantExpired:    false,
		},
		{
			name:           "past expires marks the file expired",
			body:           "Contact: mailto:security@example.com\nExpires: 2001-01-01T00:00:00Z",
			wantPresent:    true,
			wantContacts:   []string{"mailto:security@example.com"},
			wantHasExpires: true,
			wantExpired:    true,
		},
		{
			name:         "field names are case insensitive",
			body:         "CONTACT: mailto:sec@example.com",
			wantPresent:  true,
			wantContacts: []string{"mailto:sec@example.com"},
		},
		{
			name:         "unparsable expires is ignored without dropping the file",
			body:         "Contact: mailto:sec@example.com\nExpires: next tuesday",
			wantPresent:  true,
			wantContacts: []string{"mailto:sec@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSecurityTxt(tt.body)

			assert.Equal(t, got.Present, tt.wantPresent)
			assert.DeepEqual(t, got.Contacts, tt.wantContacts)
			assert.Equal(t, got.Policy, tt.wantPolicy)
			assert.DeepEqual(t, got.PreferredLanguages, tt.wantLanguages)
			assert.Equal(t, got.Expires != nil, tt.wantHasExpires)
			assert.Equal(t, got.Expired, tt.wantExpired)
		})
	}
}

func TestParseSecurityTxtExpiresRoundTrip(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)

	got := parseSecurityTxt("Contact: mailto:sec@example.com\nExpires: " + future)

	assert.Assert(t, got.Expires != nil)
	assert.Equal(t, got.Expired, false)
}

func TestParseRobotsTxt(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		wantDisallowed []string
		wantNamed      []string
		wantSignals    map[string]string
		wantSitemaps   []string
		wantStatesAI   bool
	}{
		{
			name:         "empty robots states no AI position",
			body:         "",
			wantStatesAI: false,
		},
		{
			name:         "generic wildcard rules name no AI crawlers",
			body:         "User-agent: *\nDisallow: /admin\nSitemap: https://example.com/sitemap.xml",
			wantSitemaps: []string{"https://example.com/sitemap.xml"},
			wantStatesAI: false,
		},
		{
			name:           "named AI crawler with full disallow is recorded",
			body:           "User-agent: GPTBot\nDisallow: /",
			wantDisallowed: []string{"GPTBot"},
			wantNamed:      []string{"GPTBot"},
			wantStatesAI:   true,
		},
		{
			name:         "named AI crawler with path scoped disallow is named but not blocked",
			body:         "User-agent: ClaudeBot\nDisallow: /private",
			wantNamed:    []string{"ClaudeBot"},
			wantStatesAI: true,
		},
		{
			name:           "consecutive user-agent lines share one group",
			body:           "User-agent: GPTBot\nUser-agent: CCBot\nUser-agent: Bytespider\nDisallow: /",
			wantDisallowed: []string{"Bytespider", "CCBot", "GPTBot"},
			wantNamed:      []string{"Bytespider", "CCBot", "GPTBot"},
			wantStatesAI:   true,
		},
		{
			name:         "content signal directives are parsed and lowercased",
			body:         "User-agent: *\nContent-Signal: search=yes, ai-train=no, use=reference\nAllow: /",
			wantSignals:  map[string]string{"search": "yes", "ai-train": "no", "use": "reference"},
			wantStatesAI: true,
		},
		{
			name:         "unknown content signal keys are dropped",
			body:         "Content-Signal: search=yes,made-up-key=maybe",
			wantSignals:  map[string]string{"search": "yes"},
			wantStatesAI: true,
		},
		{
			name:           "crawler names match case insensitively but report canonical casing",
			body:           "User-agent: gptbot\nDisallow: /",
			wantDisallowed: []string{"GPTBot"},
			wantNamed:      []string{"GPTBot"},
			wantStatesAI:   true,
		},
		{
			name:           "comments are stripped before parsing",
			body:           "# block the bots\nUser-agent: GPTBot # openai\nDisallow: / # everything",
			wantDisallowed: []string{"GPTBot"},
			wantNamed:      []string{"GPTBot"},
			wantStatesAI:   true,
		},
		{
			name: "openlane style file with signals and named crawlers",
			body: `User-agent: *
Content-Signal: search=yes,ai-train=no,use=reference
Allow: /

User-agent: ClaudeBot
Disallow: /

User-agent: GPTBot
Disallow: /

User-agent: Google-Extended
Disallow: /`,
			wantDisallowed: []string{"ClaudeBot", "GPTBot", "Google-Extended"},
			wantNamed:      []string{"ClaudeBot", "GPTBot", "Google-Extended"},
			wantSignals:    map[string]string{"search": "yes", "ai-train": "no", "use": "reference"},
			wantStatesAI:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRobotsTxt(tt.body)

			assert.Equal(t, got.Present, true)
			assert.DeepEqual(t, got.DisallowedAICrawlers, tt.wantDisallowed)
			assert.DeepEqual(t, got.NamedAICrawlers, tt.wantNamed)
			assert.DeepEqual(t, got.ContentSignals, tt.wantSignals)
			assert.DeepEqual(t, got.Sitemaps, tt.wantSitemaps)
			assert.Equal(t, got.StatesAIPosition(), tt.wantStatesAI)
		})
	}
}

func TestParseLLMsTxt(t *testing.T) {
	tests := []struct {
		name             string
		body             string
		wantTitle        string
		wantSummary      string
		wantSectionCount int
	}{
		{
			name:      "title only",
			body:      "# Openlane",
			wantTitle: "Openlane",
		},
		{
			name:             "title, summary and sections",
			body:             "# Openlane\n\n> Compliance automation.\n\n## Docs\n\n- [Guide](https://example.com)\n\n## Optional\n\n- [Blog](https://example.com/blog)",
			wantTitle:        "Openlane",
			wantSummary:      "Compliance automation.",
			wantSectionCount: 2,
		},
		{
			name:      "only the first h1 is taken as the title",
			body:      "# First\n# Second",
			wantTitle: "First",
		},
		{
			name:             "h3 headings are not counted as sections",
			body:             "# Openlane\n## Docs\n### Nested\n",
			wantTitle:        "Openlane",
			wantSectionCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLLMsTxt(tt.body)

			assert.Equal(t, got.Present, true)
			assert.Equal(t, got.Title, tt.wantTitle)
			assert.Equal(t, got.Summary, tt.wantSummary)
			assert.Equal(t, got.SectionCount, tt.wantSectionCount)
		})
	}
}

func TestParseHSTS(t *testing.T) {
	tests := []struct {
		name            string
		header          string
		wantHSTS        bool
		wantMaxAge      int64
		wantIncludeSubs bool
		wantPreload     bool
	}{
		{
			name:     "empty header is not HSTS",
			header:   "",
			wantHSTS: false,
		},
		{
			name:       "max-age zero switches HSTS off",
			header:     "max-age=0",
			wantHSTS:   false,
			wantMaxAge: 0,
		},
		{
			name:       "plain max-age",
			header:     "max-age=31536000",
			wantHSTS:   true,
			wantMaxAge: 31536000,
		},
		{
			name:            "full directive set",
			header:          "max-age=63072000; includeSubDomains; preload",
			wantHSTS:        true,
			wantMaxAge:      63072000,
			wantIncludeSubs: true,
			wantPreload:     true,
		},
		{
			name:       "quoted max-age value",
			header:     `max-age="31536000"`,
			wantHSTS:   true,
			wantMaxAge: 31536000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TransportSecurity

			parseHSTS(tt.header, &got)

			assert.Equal(t, got.HSTS, tt.wantHSTS)
			assert.Equal(t, got.HSTSMaxAge, tt.wantMaxAge)
			assert.Equal(t, got.HSTSIncludeSubdomains, tt.wantIncludeSubs)
			assert.Equal(t, got.HSTSPreload, tt.wantPreload)
		})
	}
}

func TestWellKnownBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		want    string
		wantErr bool
	}{
		{name: "bare domain", domain: "example.com", want: "https://example.com"},
		{name: "host with subdomain reduces to registrable name", domain: "mail.corp.example.com", want: "https://example.com"},
		{name: "full url", domain: "https://www.example.co.uk/path?q=1", want: "https://example.co.uk"},
		{name: "empty domain errors", domain: "  ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := wellKnownBaseURL(tt.domain)

			if tt.wantErr {
				assert.Assert(t, err != nil)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestIsPlainTextContentType(t *testing.T) {
	assert.Assert(t, is.Equal(isPlainTextContentType("text/plain"), true))
	assert.Assert(t, is.Equal(isPlainTextContentType("text/plain; charset=utf-8"), true))
	assert.Assert(t, is.Equal(isPlainTextContentType("TEXT/MARKDOWN"), true))
	assert.Assert(t, is.Equal(isPlainTextContentType("text/html; charset=utf-8"), false))
	assert.Assert(t, is.Equal(isPlainTextContentType("application/json"), false))
}

func TestWellKnownIsEmpty(t *testing.T) {
	assert.Equal(t, WellKnown{}.IsEmpty(), true)
	assert.Equal(t, WellKnown{Robots: &RobotsTxt{}}.IsEmpty(), false)
}
