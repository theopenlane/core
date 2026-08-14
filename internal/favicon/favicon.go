package favicon

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/validator"
)

// maxPageBytes bounds how much of a landing page is read while scanning for icon links (2MB)
const maxPageBytes = 2 << 20

// Discover returns the first reachable icon URL scraped from the given domains, trying each
// domain's landing page icon links (largest first, apple-touch preferred) before falling
// back to /favicon.ico
func Discover(ctx context.Context, client *http.Client, domains []string) (string, error) {
	var errs []error

	for _, domain := range domains {
		websiteURL := normalizeURL(domain)
		if websiteURL == "" {
			continue
		}

		avatarURL, err := fetchAvatarFromURL(ctx, client, websiteURL)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if avatarURL == "" {
			continue
		}

		return avatarURL, nil
	}

	return "", errors.Join(errs...)
}

func fetchAvatarFromURL(ctx context.Context, client *http.Client, websiteURL string) (string, error) {
	if err := validator.ValidateURL()(websiteURL); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, websiteURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("domain returned status %d", resp.StatusCode) //nolint:err113
	}

	reader := io.LimitReader(resp.Body, maxPageBytes)

	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return "", err
	}

	availableIconURLs := retrieveAvailableIcons(document, resp.Request.URL)

	// /favicon.ico as a default
	faviconURL := resp.Request.URL.ResolveReference(&url.URL{Path: "/favicon.ico"}).String()

	availableIconURLs = append(availableIconURLs, faviconURL)

	for _, iconURL := range availableIconURLs {
		if !checkIconURL(ctx, client, iconURL) {
			continue
		}

		return iconURL, nil
	}

	return "", nil
}

func retrieveAvailableIcons(document *goquery.Document, pageURL *url.URL) []string {
	type icon struct {
		url  string
		size int
	}

	var appleIcons, icons, shortcutIcons []icon

	document.Find("link[href]").Each(func(_ int, selection *goquery.Selection) {
		rel := strings.ToLower(strings.TrimSpace(selection.AttrOr("rel", "")))

		href := strings.TrimSpace(selection.AttrOr("href", ""))
		if href == "" {
			return
		}

		resolvedURL, err := pageURL.Parse(href)
		if err != nil || (resolvedURL.Scheme != "http" && resolvedURL.Scheme != "https") {
			return
		}

		resolvedIcon := icon{url: resolvedURL.String(), size: getLargestIconSize(selection.AttrOr("sizes", ""))}

		switch {
		case strings.Contains(rel, "apple-touch-icon"):
			appleIcons = append(appleIcons, resolvedIcon)
		case rel == "shortcut icon":
			shortcutIcons = append(shortcutIcons, resolvedIcon)
		case lo.Contains(strings.Fields(rel), "icon"):
			icons = append(icons, resolvedIcon)
		}
	})

	// sort from the largest sized icons to the smallest
	// while keeping the priorities still. ( apple-touch, icon, shortcut-icon  )
	groups := lo.Map([][]icon{appleIcons, icons, shortcutIcons}, func(group []icon, _ int) []string {
		slices.SortStableFunc(group, func(a, b icon) int {
			return cmp.Compare(b.size, a.size)
		})

		return lo.Map(group, func(icon icon, _ int) string {
			return icon.url
		})
	})

	return lo.Flatten(groups)
}

// checkIconURL checks if the url contains a valid image
func checkIconURL(ctx context.Context, client *http.Client, iconURL string) bool {
	if err := validator.ValidateURL()(iconURL); err != nil {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, iconURL, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Accept", "image/*")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)

	return resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
}

// normalizeURL adds HTTPS when the domain has no scheme.
// if provided domain by the org during onboarding is theopenlane.io, this
// converts it to https://theopenlane.io so we can reach it successfully
func normalizeURL(domain string) string {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return ""
	}

	parsed, err := url.Parse(domain)
	if err != nil {
		return ""
	}

	if parsed.Scheme == "" {
		return "https://" + domain
	}

	parsed.Scheme = "https"

	return parsed.String()
}

func getLargestIconSize(sizes string) int {
	// defined as <link rel="icon" sizes="32x32 64x64">
	sizesCollection := strings.Fields(strings.ToLower(sizes))

	areas := lo.FilterMap(sizesCollection, func(size string, _ int) (int, bool) {
		parts := strings.SplitN(size, "x", 2) //nolint:mnd
		if len(parts) != 2 {                  //nolint:mnd
			return 0, false
		}

		width, widthErr := strconv.Atoi(parts[0])
		height, heightErr := strconv.Atoi(parts[1])
		area := width * height

		return area, widthErr == nil && heightErr == nil && area > 0
	})

	return lo.Max(areas)
}
