package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

const (
	// messageIDSeparator delimits the signed content, so a message id cannot contain it
	messageIDSeparator = "."
	// reasonMessageIDInvalid is the terminal reason for an unusable message id
	reasonMessageIDInvalid = "webhook message id is empty or contains a full stop"
	// reasonDestinationBlocked is the terminal reason for a non-public destination
	reasonDestinationBlocked = "webhook destination is not a public address"
	// reasonDeliveryRejected is the terminal reason for a response that will not change on retry
	reasonDeliveryRejected = "webhook destination rejected the delivery"
)

// deliver posts the body once with the Standard Webhooks id and timestamp headers, signed when a signing secret is configured; the job queue retries
func deliver(ctx context.Context, _ types.OperationRequest, c *Client, cfg DeliverOperation) (json.RawMessage, error) {
	if cfg.ID == "" || strings.Contains(cfg.ID, messageIDSeparator) {
		return nil, types.Terminal(ErrMessageIDInvalid, reasonMessageIDInvalid)
	}

	timestamp := time.Now().Unix()

	opts := []httpsling.Option{
		httpsling.Body([]byte(cfg.Body)),
		httpsling.Header(headerWebhookID, cfg.ID),
		httpsling.Header(headerWebhookTimestamp, strconv.FormatInt(timestamp, 10)),
	}

	if len(c.signingKey) > 0 {
		name, value := c.signature(cfg.ID, timestamp, cfg.Body)

		opts = append(opts, httpsling.Header(name, value))
	}

	resp, err := c.requester.SendWithContext(ctx, opts...)
	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck
	}

	// redirects are not followed, so a 3xx will not change on retry; a timeout or rate limit may
	switch {
	case errors.Is(err, urlx.ErrNonPublicDestination):
		return nil, types.Terminal(fmt.Errorf("%w: %w", ErrDeliveryFailed, err), reasonDestinationBlocked)
	case err != nil:
		return nil, fmt.Errorf("%w: %w", ErrDeliveryFailed, err)
	case httpsling.IsSuccess(resp):
		return providerkit.EncodeResult(DeliverResult{StatusCode: resp.StatusCode}, ErrResultEncode)
	case resp.StatusCode == http.StatusRequestTimeout, resp.StatusCode == http.StatusTooManyRequests, resp.StatusCode >= http.StatusInternalServerError:
		return nil, fmt.Errorf("%w: status %d", ErrDeliveryFailed, resp.StatusCode)
	default:
		return nil, types.Terminal(fmt.Errorf("%w: status %d", ErrDeliveryRejected, resp.StatusCode), reasonDeliveryRejected)
	}
}
