package handlers

import (
	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/browser_rendering"
	"github.com/cloudflare/cloudflare-go/v7/option"

	models "github.com/theopenlane/core/common/openapi"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/rout"
)

const (
	snapshotCacheTTL = 600 // 10 minutes in seconds
)

// Snapshot will take a snapshot of a provided domain
func (h *Handler) SnapshotHandler(ctx echo.Context) error {
	in, err := BindAndValidate[models.SnapshotRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	out := &models.SnapshotResponse{
		Reply: rout.Reply{Success: true},
	}

	opts := []option.RequestOption{
		option.WithAPIToken(h.CloudflareConfig.APIToken),
		option.WithHeader(httpsling.HeaderContentType, httpsling.ContentTypeJSON),
	}

	client := cloudflare.NewClient(opts...)

	resp, err := client.BrowserRendering.Snapshot.New(reqCtx, h.getSnapshotParams(in))
	if err != nil {
		logx.FromContext(reqCtx).Error().Str("url", in.URL).Err(err).Msg("failed to take snapshot")

		return h.InternalServerError(ctx, err)
	}

	out.Image = resp.Screenshot

	return h.Success(ctx, out)
}

// getSnapshotParams converts the input SnapshotRequest into Cloudflare SnapshotNewParams
// for use with the Cloudflare API
func (h *Handler) getSnapshotParams(in *models.SnapshotRequest) browser_rendering.SnapshotNewParams {
	params := browser_rendering.SnapshotNewParams{
		AccountID: cloudflare.F(h.CloudflareConfig.AccountID),
		CacheTTL:  cloudflare.Float(snapshotCacheTTL),
		URL:       cloudflare.F(in.URL),
		ScreenshotOptions: cloudflare.F(browser_rendering.SnapshotNewParamsScreenshotOptions{
			Type: cloudflare.F(browser_rendering.SnapshotNewParamsScreenshotOptionsTypePNG),
		}),
	}

	if in.WaitForSelector != "" {
		params.WaitForSelector = cloudflare.F(browser_rendering.SnapshotNewParamsWaitForSelector{
			Selector: cloudflare.F(in.WaitForSelector),
		})
	}

	return params
}
