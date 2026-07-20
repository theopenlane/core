package domainscan

import (
	"encoding/json"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
)

// buildBranding extracts the site favicon from the scan's page data, returning nil when no
// favicon was captured or result is nil
func buildBranding(result *url_scanner.ScanGetResponse) *Branding {
	if result == nil {
		return nil
	}

	type pageData struct {
		Favicon Favicon `json:"favicon"`
	}

	favicon := result.Page.JSON.RawJSON()
	if favicon == "" {
		return nil
	}

	var page pageData
	if err := json.Unmarshal([]byte(favicon), &page); err != nil || page.Favicon.Hash == "" {
		return nil
	}

	return &Branding{Favicon: page.Favicon}
}
