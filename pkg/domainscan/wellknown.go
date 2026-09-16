package domainscan

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/sync/errgroup"

	"github.com/theopenlane/core/v2/pkg/urlx"
)

// maxWellKnownBodyBytes caps how much of a well-known file is read. These files are
// conventionally small; anything larger is either misconfigured or hostile
const maxWellKnownBodyBytes int64 = 256 * 1024

// knownAICrawlers are the user-agent tokens published by AI training, retrieval, and
// agent crawlers, matched case-insensitively against robots.txt user-agent groups
var knownAICrawlers = []string{
	"GPTBot",
	"OAI-SearchBot",
	"ChatGPT-User",
	"ClaudeBot",
	"Claude-Web",
	"anthropic-ai",
	"Google-Extended",
	"CCBot",
	"PerplexityBot",
	"Perplexity-User",
	"Bytespider",
	"Applebot-Extended",
	"Amazonbot",
	"meta-externalagent",
	"FacebookBot",
	"Diffbot",
	"Omgilibot",
	"ImagesiftBot",
	"cohere-ai",
	"Timpibot",
	"YouBot",
	"PetalBot",
	"CloudflareBrowserRenderingCrawler",
}

// contentSignalKeys are the recognized keys within a Content-Signal directive
var contentSignalKeys = []string{"search", "ai-train", "ai-input", "use"}

// SecurityTxt is the parsed result of probing a domain for RFC 9116 security.txt
type SecurityTxt struct {
	// Present reports whether a security.txt was served
	Present bool `json:"present"`
	// URL is where it was found, since the well-known path takes precedence over the legacy root path
	URL string `json:"url,omitempty"`
	// Contacts are the Contact field values, in the order published
	Contacts []string `json:"contacts,omitempty"`
	// Policy is the Policy field value, linking the vulnerability disclosure policy
	Policy string `json:"policy,omitempty"`
	// Encryption is the Encryption field value, linking a public key
	Encryption string `json:"encryption,omitempty"`
	// PreferredLanguages are the Preferred-Languages field values
	PreferredLanguages []string `json:"preferred_languages,omitempty"`
	// Expires is the Expires field value, after which RFC 9116 says the file should not be relied on
	Expires *time.Time `json:"expires,omitempty"`
	// Expired reports whether Expires is in the past
	Expired bool `json:"expired,omitempty"`
}

// RobotsTxt is the parsed result of probing a domain for robots.txt, focused on
// what it says about AI crawlers rather than general crawl directives
type RobotsTxt struct {
	// Present reports whether a robots.txt was served
	Present bool `json:"present"`
	// ContentSignals are the parsed Content-Signal directives, e.g. {"ai-train": "no"}
	ContentSignals map[string]string `json:"content_signals,omitempty"`
	// DisallowedAICrawlers are the known AI crawlers fully disallowed (Disallow: /)
	DisallowedAICrawlers []string `json:"disallowed_ai_crawlers,omitempty"`
	// NamedAICrawlers are the known AI crawlers named at all, whether allowed or disallowed
	NamedAICrawlers []string `json:"named_ai_crawlers,omitempty"`
	// Sitemaps are the declared sitemap URLs
	Sitemaps []string `json:"sitemaps,omitempty"`
}

// StatesAIPosition reports whether the file expresses any position on AI crawlers,
// through either named crawler rules or Content-Signal directives
func (r RobotsTxt) StatesAIPosition() bool {
	return len(r.NamedAICrawlers) > 0 || len(r.ContentSignals) > 0
}

// LLMsTxt is the parsed result of probing a domain for an llms.txt summary file
type LLMsTxt struct {
	// Present reports whether /llms.txt was served
	Present bool `json:"present"`
	// URL is where it was found
	URL string `json:"url,omitempty"`
	// Title is the H1 project or site name, the spec's only required section
	Title string `json:"title,omitempty"`
	// Summary is the blockquote summary immediately following the H1, if present
	Summary string `json:"summary,omitempty"`
	// SectionCount is the number of H2-delimited sections
	SectionCount int `json:"section_count,omitempty"`
	// FullPresent reports whether the companion /llms-full.txt was also served
	FullPresent bool `json:"full_present,omitempty"`
}

// TransportSecurity is the result of probing how strictly a domain enforces HTTPS
type TransportSecurity struct {
	// HTTPSReachable reports whether the domain answered over HTTPS
	HTTPSReachable bool `json:"https_reachable"`
	// RedirectsToHTTPS reports whether a plain HTTP request ended up on HTTPS
	RedirectsToHTTPS bool `json:"redirects_to_https"`
	// HSTS reports whether a Strict-Transport-Security header was sent over HTTPS
	HSTS bool `json:"hsts"`
	// HSTSMaxAge is the header's max-age directive in seconds
	HSTSMaxAge int64 `json:"hsts_max_age,omitempty"`
	// HSTSIncludeSubdomains reports whether the header carried includeSubDomains
	HSTSIncludeSubdomains bool `json:"hsts_include_subdomains,omitempty"`
	// HSTSPreload reports whether the header carried the preload directive
	HSTSPreload bool `json:"hsts_preload,omitempty"`
}

// WellKnown holds the well-known file and transport security probes for a domain.
// Every field is best-effort: a nil section means that probe found nothing or failed,
// which for scoring purposes is treated the same as absent
type WellKnown struct {
	// SecurityTxt is the RFC 9116 security contact probe
	SecurityTxt *SecurityTxt `json:"security_txt,omitempty"`
	// Robots is the robots.txt AI crawler probe
	Robots *RobotsTxt `json:"robots_txt,omitempty"`
	// LLMs is the llms.txt probe
	LLMs *LLMsTxt `json:"llms_txt,omitempty"`
	// Transport is the HTTPS and HSTS enforcement probe
	Transport *TransportSecurity `json:"transport,omitempty"`
}

// IsEmpty reports whether no probe produced a result
func (w WellKnown) IsEmpty() bool {
	return w.SecurityTxt == nil && w.Robots == nil && w.LLMs == nil && w.Transport == nil
}

// GetWellKnown probes domain for security.txt, robots.txt, llms.txt, and HTTPS
// enforcement, concurrently. Each probe is best-effort and independent, so a failure
// in one leaves the others intact; the returned error is non-nil only when domain
// itself could not be resolved to a registrable name
func GetWellKnown(ctx context.Context, domain string) (*WellKnown, error) {
	base, err := wellKnownBaseURL(domain)
	if err != nil {
		return nil, err
	}

	var (
		result WellKnown
		g      errgroup.Group
	)

	g.Go(func() error {
		result.SecurityTxt = probeSecurityTxt(ctx, base)

		return nil
	})

	g.Go(func() error {
		result.Robots = probeRobotsTxt(ctx, base)

		return nil
	})

	g.Go(func() error {
		result.LLMs = probeLLMsTxt(ctx, base)

		return nil
	})

	g.Go(func() error {
		result.Transport = probeTransportSecurity(ctx, base)

		return nil
	})

	_ = g.Wait() // each probe is best-effort and records absence rather than failing

	return &result, nil
}

// wellKnownBaseURL normalizes domain to an https origin at its registrable name,
// accepting a bare domain, a hostname, or a full URL
func wellKnownBaseURL(domain string) (string, error) {
	raw := strings.TrimSpace(domain)
	if raw == "" {
		return "", ErrInvalidDomain
	}

	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	parsed, err := urlx.Parse(raw)
	if err != nil {
		return "", err
	}

	host, err := publicsuffix.EffectiveTLDPlusOne(parsed.Hostname())
	if err != nil {
		return "", err
	}

	return "https://" + host, nil
}

// isPlainTextContentType reports whether ct is a text media type a well-known file
// would plausibly be served as, excluding text/html which indicates a soft 404
func isPlainTextContentType(ct string) bool {
	media := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))

	switch media {
	case "text/plain", "text/markdown", "text/x-markdown", "application/octet-stream":
		return true
	default:
		return false
	}
}

// probeSecurityTxt fetches and parses a domain's RFC 9116 security.txt, preferring the
// well-known location and falling back to the legacy document root
func probeSecurityTxt(ctx context.Context, base string) *SecurityTxt {
	for _, path := range []string{"/.well-known/security.txt", "/security.txt"} {
		body, finalURL, ok := fetchBody(ctx, base+path, maxWellKnownBodyBytes, isPlainTextContentType)
		if !ok {
			continue
		}

		parsed := parseSecurityTxt(body)
		if !parsed.Present {
			continue
		}

		parsed.URL = finalURL

		return parsed
	}

	return &SecurityTxt{Present: false}
}

// parseSecurityTxt parses the field/value lines of a security.txt body. Fields are
// case-insensitive and may repeat; a body with no Contact field is not treated as
// present, since Contact is the only field RFC 9116 requires
func parseSecurityTxt(body string) *SecurityTxt {
	out := &SecurityTxt{}

	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		field, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}

		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		switch strings.ToLower(strings.TrimSpace(field)) {
		case "contact":
			out.Contacts = append(out.Contacts, value)
		case "policy":
			if out.Policy == "" {
				out.Policy = value
			}
		case "encryption":
			if out.Encryption == "" {
				out.Encryption = value
			}
		case "preferred-languages":
			for _, lang := range strings.Split(value, ",") {
				if lang = strings.TrimSpace(lang); lang != "" {
					out.PreferredLanguages = append(out.PreferredLanguages, lang)
				}
			}
		case "expires":
			if expires, err := time.Parse(time.RFC3339, value); err == nil {
				out.Expires = &expires
				out.Expired = time.Now().After(expires)
			}
		}
	}

	// RFC 9116 requires Contact; without one a researcher has nowhere to send a finding,
	// which is the whole point of the file. Such a file is reported as absent rather than
	// present-but-useless, and its other fields are dropped so a caller is never handed a
	// Policy or Expires belonging to a file we have just declared missing
	if len(out.Contacts) == 0 {
		return &SecurityTxt{}
	}

	out.Present = true

	return out
}

// probeRobotsTxt fetches and parses a domain's robots.txt
func probeRobotsTxt(ctx context.Context, base string) *RobotsTxt {
	body, _, ok := fetchBody(ctx, base+"/robots.txt", maxWellKnownBodyBytes, isPlainTextContentType)
	if !ok {
		return &RobotsTxt{Present: false}
	}

	return parseRobotsTxt(body)
}

// parseRobotsTxt walks robots.txt group by group, recording which known AI crawlers are
// named and which are fully disallowed, plus any Content-Signal directives and sitemaps.
// Consecutive User-agent lines form one group, per the robots.txt convention
func parseRobotsTxt(body string) *RobotsTxt {
	out := &RobotsTxt{Present: true, ContentSignals: map[string]string{}}

	var (
		group        []string
		inAgentBlock bool
		named        = map[string]bool{}
		disallowed   = map[string]bool{}
	)

	for _, line := range strings.Split(body, "\n") {
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		field, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}

		field = strings.ToLower(strings.TrimSpace(field))
		value = strings.TrimSpace(value)

		switch field {
		case "user-agent":
			if !inAgentBlock {
				group = nil
			}

			group = append(group, value)
			inAgentBlock = true

			if crawler, isAI := matchAICrawler(value); isAI {
				named[crawler] = true
			}

		case "disallow":
			inAgentBlock = false

			// only a bare "/" is a full block; a path-scoped disallow is not
			if value != "/" {
				continue
			}

			for _, agent := range group {
				if crawler, isAI := matchAICrawler(agent); isAI {
					disallowed[crawler] = true
				}
			}

		case "content-signal":
			inAgentBlock = false

			for key, val := range parseContentSignal(value) {
				out.ContentSignals[key] = val
			}

		case "sitemap":
			inAgentBlock = false

			if value != "" {
				out.Sitemaps = append(out.Sitemaps, value)
			}

		default:
			inAgentBlock = false
		}
	}

	out.NamedAICrawlers = sortedKeys(named)
	out.DisallowedAICrawlers = sortedKeys(disallowed)

	if len(out.ContentSignals) == 0 {
		out.ContentSignals = nil
	}

	return out
}

// matchAICrawler reports whether agent names one of knownAICrawlers, returning the
// canonical spelling so output doesn't inherit whatever casing the site used
func matchAICrawler(agent string) (string, bool) {
	agent = strings.TrimSpace(agent)

	for _, crawler := range knownAICrawlers {
		if strings.EqualFold(agent, crawler) {
			return crawler, true
		}
	}

	return "", false
}

// parseContentSignal parses a Content-Signal directive value, e.g.
// "search=yes,ai-train=no,use=reference", keeping only recognized keys
func parseContentSignal(value string) map[string]string {
	out := map[string]string{}

	for _, part := range strings.Split(value, ",") {
		key, val, found := strings.Cut(part, "=")
		if !found {
			continue
		}

		key = strings.ToLower(strings.TrimSpace(key))
		val = strings.ToLower(strings.TrimSpace(val))

		if val == "" {
			continue
		}

		for _, known := range contentSignalKeys {
			if key == known {
				out[key] = val

				break
			}
		}
	}

	return out
}

// sortedKeys returns m's keys in a stable order so repeated scans of an unchanged
// site produce identical reports
func sortedKeys(m map[string]bool) []string {
	if len(m) == 0 {
		return nil
	}

	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	slices.Sort(out)

	return out
}

// probeLLMsTxt fetches and parses a domain's llms.txt, and separately notes whether the
// companion llms-full.txt exists. Neither is part of Cloudflare's agent-readiness
// assessment, so this is additive rather than duplicative
func probeLLMsTxt(ctx context.Context, base string) *LLMsTxt {
	body, finalURL, ok := fetchBody(ctx, base+"/llms.txt", maxWellKnownBodyBytes, isPlainTextContentType)
	if !ok {
		return &LLMsTxt{Present: false}
	}

	out := parseLLMsTxt(body)
	out.URL = finalURL

	if _, _, fullOK := fetchBody(ctx, base+"/llms-full.txt", maxWellKnownBodyBytes, isPlainTextContentType); fullOK {
		out.FullPresent = true
	}

	return out
}

// parseLLMsTxt reads the structure llmstxt.org defines: a required H1 naming the project,
// an optional blockquote summary, and zero or more H2-delimited sections
func parseLLMsTxt(body string) *LLMsTxt {
	out := &LLMsTxt{Present: true}

	var sawTitle bool

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "## "):
			out.SectionCount++

		case strings.HasPrefix(trimmed, "# ") && !sawTitle:
			out.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			sawTitle = true

		case strings.HasPrefix(trimmed, "> ") && sawTitle && out.Summary == "":
			out.Summary = strings.TrimSpace(strings.TrimPrefix(trimmed, "> "))
		}
	}

	return out
}

// probeTransportSecurity checks whether a domain answers over HTTPS, whether plain HTTP
// redirects there, and what Strict-Transport-Security it sends. The redirect check relies
// on the shared requester following redirects, so the final URL's scheme is the answer
func probeTransportSecurity(ctx context.Context, base string) *TransportSecurity {
	out := &TransportSecurity{}

	requester, err := scanRequester()
	if err != nil {
		return out
	}

	resp, err := requester.SendWithContext(ctx, httpsling.Head(base))
	if err == nil {
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode < http.StatusBadRequest {
			out.HTTPSReachable = true
			parseHSTS(resp.Header.Get("Strict-Transport-Security"), out)
		}
	}

	if plain, ok := strings.CutPrefix(base, "https://"); ok {
		if finalURL, reachable := urlReachable(ctx, "http://"+plain); reachable {
			out.RedirectsToHTTPS = strings.HasPrefix(strings.ToLower(finalURL), "https://")
		}
	}

	return out
}

// parseHSTS parses a Strict-Transport-Security header value onto out. A header with a
// zero or absent max-age is not treated as HSTS being in force, since max-age=0 is the
// directive for switching it off
func parseHSTS(header string, out *TransportSecurity) {
	if strings.TrimSpace(header) == "" {
		return
	}

	for _, directive := range strings.Split(header, ";") {
		directive = strings.ToLower(strings.TrimSpace(directive))

		switch {
		case strings.HasPrefix(directive, "max-age"):
			_, value, found := strings.Cut(directive, "=")
			if !found {
				continue
			}

			if age, err := strconv.ParseInt(strings.Trim(strings.TrimSpace(value), `"`), 10, 64); err == nil {
				out.HSTSMaxAge = age
			}

		case directive == "includesubdomains":
			out.HSTSIncludeSubdomains = true

		case directive == "preload":
			out.HSTSPreload = true
		}
	}

	out.HSTS = out.HSTSMaxAge > 0
}
