package hooks

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
	"time"

	"entgo.io/ent"
	"github.com/PuerkitoBio/goquery"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	avatarFetchTimeout = 1 * time.Minute
	// 2MB
	avatarMaxBytes = 2 << 20
)

var avatarDiscoveryClient = &http.Client{
	Timeout: avatarFetchTimeout,
}

// RegisterGalaOrganizationAvatarListeners registers organization avatar discovery on Gala.
func RegisterGalaOrganizationAvatarListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry, gala.Definition[eventqueue.MutationGalaPayload]{
		Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeOrganization),
		Name:       "organization.avatar",
		Operations: []string{ent.OpCreate.String()},
		Handle:     handleOrganizationAvatarCreated,
	})
}

// handleOrganizationAvatarCreated fetches icons from the domain name instead and sets it as the remote logo url
func handleOrganizationAvatarCreated(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	orgID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || orgID == "" {
		return nil
	}

	allowCtx := workflows.AllowContext(ctx.Context)

	org, err := client.Organization.Query().
		Where(organization.IDEQ(orgID)).
		WithSetting().
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	setting, err := org.Edges.SettingOrErr()
	if err != nil || setting == nil || len(setting.Domains) == 0 {
		return nil
	}

	avatarURL, err := discoverAvatar(ctx.Context, avatarDiscoveryClient, setting.Domains)
	if err != nil {
		logx.FromContext(ctx.Context).Err(err).
			Str("organization_id", orgID).
			Msg("organization avatar discovery failed")
		return nil
	}

	if avatarURL == "" {
		return nil
	}

	return client.Organization.UpdateOneID(orgID).
		SetAvatarRemoteURL(avatarURL).
		Exec(allowCtx)
}

func discoverAvatar(ctx context.Context, client *http.Client, domains []string) (string, error) {
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
		return "", fmt.Errorf("domain returned status %d", resp.StatusCode)
	}

	reader := io.LimitReader(resp.Body, avatarMaxBytes)

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

		parts := strings.SplitN(size, "x", 2)
		if len(parts) != 2 {
			return 0, false
		}

		width, widthErr := strconv.Atoi(parts[0])
		height, heightErr := strconv.Atoi(parts[1])
		area := width * height

		return area, widthErr == nil && heightErr == nil && area > 0
	})

	return lo.Max(areas)
}
