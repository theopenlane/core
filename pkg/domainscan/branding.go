package domainscan

import (
	"encoding/json"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
)

// buildBranding combines the URL Scanner favicon with browser-rendered design tokens
func buildBranding(result *url_scanner.ScanGetResponse, profile *BrandDesignProfile) *Branding {
	branding := &Branding{}
	if profile != nil {
		branding.PrimaryColor = profile.PrimaryColor
		branding.Font = profile.Font
		branding.ForegroundColor = profile.ForegroundColor
		branding.BackgroundColor = profile.BackgroundColor
		branding.AccentColor = profile.AccentColor
		branding.SecondaryBackgroundColor = profile.SecondaryBackgroundColor
		branding.SecondaryForegroundColor = profile.SecondaryForegroundColor
	}

	if result == nil {
		if profile == nil || profile.IsEmpty() {
			return nil
		}

		return branding
	}

	type pageData struct {
		Favicon Favicon `json:"favicon"`
	}

	favicon := result.Page.JSON.RawJSON()
	if favicon == "" {
		if profile == nil || profile.IsEmpty() {
			return nil
		}

		return branding
	}

	var page pageData
	if err := json.Unmarshal([]byte(favicon), &page); err == nil && page.Favicon.Hash != "" {
		branding.Favicon = page.Favicon
	}

	if branding.Favicon.Hash == "" && (profile == nil || profile.IsEmpty()) {
		return nil
	}

	return branding
}
