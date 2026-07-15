package domainscan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// crtSHURL is the certificate transparency log search endpoint used to discover subdomains
// from issued certificates without ever contacting the target domain's own infrastructure
const crtSHURL = "https://crt.sh/"

// crtSHRequestTimeout bounds how long a single crt.sh query may take. crt.sh is a free,
// unauthenticated, best-effort public service that can be slow or degraded; without its own
// timeout a hung request would sit on the shared enrichment budget other lookups need
const crtSHRequestTimeout = 10 * time.Second

// maxCertSubdomains bounds how many distinct subdomains are kept from a crt.sh response
const maxCertSubdomains = 200

// crtSHEntry is a single row from crt.sh's JSON output
type crtSHEntry struct {
	// NameValue holds the certificate's common name and SANs, newline-separated
	NameValue string `json:"name_value"`
}

// certTransparencySubdomains queries crt.sh for certificates issued for apex and returns the distinct subdomain hostnames found in their SANs
func certTransparencySubdomains(ctx context.Context, apex string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, crtSHRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, crtSHURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("q", "%."+apex)
	q.Set("output", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w - %d", ErrUnexpectedErrorCode, resp.StatusCode)
	}

	var entries []crtSHEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	return parseCertTransparencySubdomains(entries, apex), nil
}

// parseCertTransparencySubdomains extracts, filters, and dedupes the subdomains of apex found
// across entries' newline-separated SAN lists, capped at maxCertSubdomains
func parseCertTransparencySubdomains(entries []crtSHEntry, apex string) []string {
	seen := make(map[string]struct{})

	var hosts []string

	for _, entry := range entries {
		for _, name := range strings.Split(entry.NameValue, "\n") {
			name = strings.ToLower(strings.TrimSpace(name))
			name = strings.TrimPrefix(name, "*.")

			if name == "" || name == apex || !strings.HasSuffix(name, "."+apex) {
				continue
			}

			if _, ok := seen[name]; ok {
				continue
			}

			seen[name] = struct{}{}
			hosts = append(hosts, name)

			if len(hosts) >= maxCertSubdomains {
				sort.Strings(hosts)
				return hosts
			}
		}
	}

	sort.Strings(hosts)

	return hosts
}
