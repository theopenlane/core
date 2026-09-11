package domainscan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/browser_rendering"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

const (
	brandingCDPReadLimit      = 8 << 20
	brandingEvaluationTimeout = 10 * time.Second
	brandingBrowserTimeout    = 75 * time.Second
	brandingCleanupTimeout    = 10 * time.Second
	brandingViewportWidth     = 1440
	brandingViewportHeight    = 900

	brandingPageInfoScript = `(async () => {
  await new Promise(resolve => setTimeout(resolve, 3000));
  await Promise.race([document.fonts.ready, new Promise(resolve => setTimeout(resolve, 3000))]);
  return {title: document.title, url: location.href};
})()`
)

type brandDesignCDP struct {
	conn             *websocket.Conn
	nextID           int
	currentSessionID string
	loadedIDs        map[string]bool
	loadedDocuments  map[string]browser_rendering.JsonNewResponseEnvelopeMeta
}

type brandDesignCDPMessaging struct {
	ID        int             `json:"id"`
	SessionID string          `json:"sessionId"`
	Method    string          `json:"method"`
	Params    json.RawMessage `json:"params"`
	Result    json.RawMessage `json:"result"`
	Error     *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (b *brandDesignCDP) callMethod(ctx context.Context, method string, params any, result any) error {
	b.nextID++

	command := map[string]any{"id": b.nextID, "method": method, "params": params}
	if b.currentSessionID != "" {
		command["sessionId"] = b.currentSessionID
	}

	if err := wsjson.Write(ctx, b.conn, command); err != nil {
		return fmt.Errorf("branding CDP %s: %w", method, err)
	}

	for {
		message, err := b.readResponse(ctx)
		if err != nil {
			return fmt.Errorf("branding CDP %s: %w", method, err)
		}

		if message.ID != b.nextID {
			continue
		}

		if message.Error != nil {
			return fmt.Errorf("%w %s (%d): %s", errBrandingCDP, method, message.Error.Code, message.Error.Message)
		}

		if result == nil {
			return nil
		}

		return json.Unmarshal(message.Result, result)
	}
}

func (b *brandDesignCDP) readResponse(ctx context.Context) (*brandDesignCDPMessaging, error) {
	var message brandDesignCDPMessaging
	if err := wsjson.Read(ctx, b.conn, &message); err != nil {
		return nil, err
	}

	if message.SessionID != b.currentSessionID {
		return &message, nil
	}

	switch message.Method {
	case "Page.lifecycleEvent":
		var event struct {
			LoaderID string `json:"loaderId"`
			Name     string `json:"name"`
		}

		if err := json.Unmarshal(message.Params, &event); err != nil {
			return nil, err
		}

		if event.Name == "load" {
			b.loadedIDs[event.LoaderID] = true
		}

	case "Network.responseReceived":
		var event struct {
			FrameID  string `json:"frameId"`
			Type     string `json:"type"`
			Response struct {
				URL     string            `json:"url"`
				Status  float64           `json:"status"`
				Headers map[string]string `json:"headers"`
			} `json:"response"`
		}
		if err := json.Unmarshal(message.Params, &event); err != nil {
			return nil, err
		}

		if event.Type == "Document" {
			b.loadedDocuments[event.FrameID] = browser_rendering.JsonNewResponseEnvelopeMeta{
				FinalURL: event.Response.URL,
				Status:   event.Response.Status,
				Headers:  event.Response.Headers,
			}
		}
	}

	return &message, nil
}

func (b *brandDesignCDP) evaluatePageData(ctx context.Context, expression string, destination any) error {
	var evaluation struct {
		Result struct {
			Value json.RawMessage `json:"value"`
		} `json:"result"`
		ExceptionDetails json.RawMessage `json:"exceptionDetails"`
	}

	params := map[string]any{
		"expression":    expression,
		"returnByValue": true,
		"awaitPromise":  true,
		"timeout":       brandingEvaluationTimeout.Milliseconds(),
		// sites like stripe.com block out with csp and we never get the data
		"allowUnsafeEvalBlockedByCSP": true,
	}

	if err := b.callMethod(ctx, "Runtime.evaluate", params, &evaluation); err != nil {
		return err
	}

	if len(evaluation.ExceptionDetails) > 0 && string(evaluation.ExceptionDetails) != "null" {
		return fmt.Errorf("%w: %s", errBrandingProbe, evaluation.ExceptionDetails)
	}

	if len(evaluation.Result.Value) == 0 || string(evaluation.Result.Value) == "null" {
		return errBrandingProbeNoValue
	}

	return json.Unmarshal(evaluation.Result.Value, destination)
}

func (c *Config) browserBranding(ctx context.Context, domain string) (*BrandDesignProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, brandingBrowserTimeout)
	defer cancel()

	target, err := urlx.Parse(domain)
	if err != nil {
		return nil, err
	}

	if target.Scheme != "http" && target.Scheme != "https" {
		return nil, errBrandingURLScheme
	}

	client := cloudflare.NewClient(c.clientOptions()...)

	devToolBrowser, err := client.BrowserRendering.Devtools.Browser.New(ctx, browser_rendering.DevtoolBrowserNewParams{
		AccountID: cloudflare.F(c.AccountID),
	})
	if err != nil {
		return nil, err
	}

	// make sure to clean up the session
	defer func() {
		deleteCtx, deleteCancel := context.WithTimeout(context.WithoutCancel(ctx), brandingCleanupTimeout)
		defer deleteCancel()

		if _, err := client.BrowserRendering.Devtools.Browser.Delete(deleteCtx, devToolBrowser.SessionID, browser_rendering.DevtoolBrowserDeleteParams{
			AccountID: cloudflare.F(c.AccountID),
		}); err != nil {
			logx.FromContext(ctx).Warn().Err(err).Msg("domainscan: failed closing branding browser")
		}
	}()

	endpoint := fmt.Sprintf("wss://api.cloudflare.com/client/v4/accounts/%s/browser-rendering/devtools/browser/%s",
		url.PathEscape(c.AccountID), url.PathEscape(devToolBrowser.SessionID))

	conn, _, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{ //nolint:bodyclose
		HTTPHeader: http.Header{"Authorization": {"Bearer " + c.APIToken}},
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := conn.CloseNow(); err != nil {
			logx.FromContext(ctx).Warn().Err(err).Msg("domainscan: failed closing branding websocket")
		}
	}()
	conn.SetReadLimit(brandingCDPReadLimit)

	cdp := &brandDesignCDP{conn: conn, loadedIDs: map[string]bool{}, loadedDocuments: map[string]browser_rendering.JsonNewResponseEnvelopeMeta{}}

	var page struct {
		TargetID string `json:"targetId"`
	}

	if err := cdp.callMethod(ctx, "Target.createTarget", map[string]any{"url": "about:blank"}, &page); err != nil {
		return nil, err
	}

	var attached struct {
		SessionID string `json:"sessionId"`
	}

	if err := cdp.callMethod(ctx, "Target.attachToTarget", map[string]any{"targetId": page.TargetID, "flatten": true}, &attached); err != nil {
		return nil, err
	}

	cdp.currentSessionID = attached.SessionID

	for _, method := range []string{"Page.enable", "Network.enable", "Runtime.enable", "Log.enable"} {
		if err := cdp.callMethod(ctx, method, map[string]any{}, nil); err != nil {
			return nil, err
		}
	}

	if err := cdp.callMethod(ctx, "Page.setLifecycleEventsEnabled", map[string]any{"enabled": true}, nil); err != nil {
		return nil, err
	}

	if err := cdp.callMethod(ctx, "Emulation.setDeviceMetricsOverride", map[string]any{
		"width": brandingViewportWidth, "height": brandingViewportHeight, "deviceScaleFactor": 1, "mobile": false,
	}, nil); err != nil {
		return nil, err
	}

	navigationCtx, navigationCancel := context.WithTimeout(ctx, browserNavigationTimeout*time.Millisecond)
	defer navigationCancel()
	var navigation struct {
		FrameID   string `json:"frameId"`
		LoaderID  string `json:"loaderId"`
		ErrorText string `json:"errorText"`
	}
	if err := cdp.callMethod(navigationCtx, "Page.navigate", map[string]any{"url": target.String()}, &navigation); err != nil {
		return nil, err
	}
	if navigation.ErrorText != "" {
		return nil, fmt.Errorf("%w: %s", errBrandingNavigation, navigation.ErrorText)
	}

	if navigation.LoaderID == "" {
		return nil, errBrandingNoDocument
	}

	for !cdp.loadedIDs[navigation.LoaderID] {
		if _, err := cdp.readResponse(navigationCtx); err != nil {
			return nil, err
		}
	}

	var pageInfo struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}

	if err := cdp.evaluatePageData(ctx, brandingPageInfoScript, &pageInfo); err != nil {
		return nil, err
	}

	meta := cdp.loadedDocuments[navigation.FrameID]
	meta.Title = pageInfo.Title
	meta.FinalURL = pageInfo.URL

	if isBotChallenge(meta) {
		return nil, errBrandingBotChallenge
	}

	if meta.Status >= http.StatusBadRequest {
		return &BrandDesignProfile{Error: fmt.Sprintf("Branding extraction failed: website returned HTTP %d.", int(meta.Status))}, nil
	}

	if !strings.HasPrefix(pageInfo.URL, "https://") && !strings.HasPrefix(pageInfo.URL, "http://") {
		return nil, errBrandingNoWebsite
	}

	var branding BrandDesignProfile
	if err := cdp.evaluatePageData(ctx, brandStyleProbeScript, &branding); err != nil {
		return nil, err
	}

	return &branding, nil
}

// isBotChallenge checks to see if we cannot access the domain
func isBotChallenge(meta browser_rendering.JsonNewResponseEnvelopeMeta) bool {
	for name, value := range meta.Headers {
		if strings.EqualFold(name, "cf-mitigated") && strings.EqualFold(strings.TrimSpace(value), "challenge") {
			return true
		}
	}

	return (meta.Status == http.StatusForbidden || meta.Status == http.StatusTooManyRequests) &&
		strings.EqualFold(strings.TrimSpace(meta.Title), "Just a moment...")
}
