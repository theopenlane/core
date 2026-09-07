package domainscan

import (
	"net/url"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"

	"github.com/theopenlane/core/v2/pkg/urlx"
)

func formatFaviconURL(domain string) string {
	host, err := urlx.NormalizeHostname(domain)
	if err != nil {
		return ""
	}

	return "https://www.google.com/s2/favicons?domain=" + url.QueryEscape(host) + "&sz=128"
}

func buildBranding(result *url_scanner.ScanGetResponse, profile *BrandDesignProfile) *Branding {
	branding := &Branding{}
	if profile != nil {
		branding.Favicon.URL = profile.FaviconURL
		branding.LogoURL = profile.LogoURL
		branding.PrimaryColor = profile.PrimaryColor
		branding.Font = profile.Font
		branding.ForegroundColor = profile.ForegroundColor
		branding.BackgroundColor = profile.BackgroundColor
		branding.AccentColor = profile.AccentColor
		branding.SecondaryBackgroundColor = profile.SecondaryBackgroundColor
		branding.SecondaryForegroundColor = profile.SecondaryForegroundColor
	}

	if result != nil {
		branding.Favicon.URL = formatFaviconURL(result.Task.URL)
	}

	if branding.Favicon.URL == "" && (profile == nil || profile.IsEmpty()) {
		return nil
	}

	return branding
}
