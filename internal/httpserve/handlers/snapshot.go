package handlers

import (
	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/browser_rendering"
	"github.com/cloudflare/cloudflare-go/v7/option"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/pkg/logx"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/httpsling"
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

	out := &models.SnapshotReply{
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
	params := browser_rendering.SnapshotNewParams{}
	params.AccountID = cloudflare.F(h.CloudflareConfig.AccountID)
	params.CacheTTL = cloudflare.Float(snapshotCacheTTL)

	body := browser_rendering.SnapshotNewParamsBody{}
	body.URL = cloudflare.F(in.URL)
	body.ScreenshotOptions = cloudflare.F[interface{}](
		browser_rendering.SnapshotNewParamsBodyObjectScreenshotOptions{
			Type: cloudflare.F(browser_rendering.SnapshotNewParamsBodyObjectScreenshotOptionsTypePNG),
		},
	)

	if in.WaitForSelector != "" {
		body.WaitForSelector = cloudflare.F[interface{}](
			browser_rendering.SnapshotNewParamsBodyObjectWaitForSelector{
				Selector: cloudflare.F(in.WaitForSelector),
			},
		)
	}

	params.Body = body

	return params
}
