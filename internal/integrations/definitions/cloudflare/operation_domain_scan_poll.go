package cloudflare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/url_scanner"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// DomainScanPoll retrieves a previously submitted URL Scanner result by scan ID
type DomainScanPoll struct {
	// ScanResultID is the scan ID returned by DomainScanSubmit
	ScanResultID string `json:"scanResultId"`
}

// DomainScanPollResult carries the raw URL Scanner result plus any task-level errors
type DomainScanPollResult struct {
	// Result is the raw URL Scanner result payload
	Result *url_scanner.ScanGetResponse `json:"result"`
	// TaskErrors lists task-level errors reported alongside the result, if any
	TaskErrors ScanTaskErrors `json:"taskErrors,omitempty"`
	// NotReady is true when Cloudflare reported the scan isn't available yet
	NotReady bool `json:"notReady,omitempty"`
}

// ScanTaskError is a single error reported in a URL Scanner task's error list
type ScanTaskError struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
	Code    int    `json:"code"`
}

// Error returns a human-readable representation of the scan task error
func (e ScanTaskError) Error() string {
	parts := []string{}
	if e.Name != "" {
		parts = append(parts, e.Name)
	}

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if e.Detail != "" {
		parts = append(parts, e.Detail)
	}

	if e.Code != 0 {
		parts = append(parts, fmt.Sprintf("code %d", e.Code))
	}

	if len(parts) == 0 {
		return "unknown scan task error"
	}

	return strings.Join(parts, ": ")
}

// ScanTaskErrors is a list of ScanTaskError, satisfying the error interface when non-empty
type ScanTaskErrors []ScanTaskError

// Error joins every task error's message with "; "
func (e ScanTaskErrors) Error() string {
	messages := make([]string, 0, len(e))
	for _, taskErr := range e {
		messages = append(messages, taskErr.Error())
	}

	return strings.Join(messages, "; ")
}

// DomainScanPollOp is the operation ref for polling a submitted URL Scanner result
//
//nolint:revive
var DomainScanPollOp = types.OperationPayloadOf[DomainScanPoll]().
	Handles(cloudflareClient, runDomainScanPoll).
	Policy(types.ExecutionPolicy{SkipRunRecord: true}).
	CustomerSelectable(false).
	Internal()

// runDomainScanPoll retrieves the configured URL Scanner result and encodes it
func runDomainScanPoll(ctx context.Context, _ types.OperationRequest, client *CloudflareClient, cfg DomainScanPoll) (json.RawMessage, error) {
	result, err := DomainScanPoll{}.Run(ctx, client, cfg)
	if err != nil {
		return nil, err
	}

	return providerkit.EncodeResult(result, ErrResultEncode)
}

// Run retrieves a URL Scanner result by scan ID
func (DomainScanPoll) Run(ctx context.Context, client *CloudflareClient, cfg DomainScanPoll) (DomainScanPollResult, error) {
	result, taskErrors, err := getScanResultOnce(ctx, client, cfg.ScanResultID)
	if err == nil {
		return DomainScanPollResult{Result: result, TaskErrors: taskErrors}, nil
	}

	var apiErr *cf.Error
	if errors.As(err, &apiErr) && (apiErr.StatusCode == http.StatusBadRequest || apiErr.StatusCode == http.StatusNotFound) {
		return DomainScanPollResult{NotReady: true}, nil
	}

	logx.FromContext(ctx).Error().Err(err).Str("scan_result_id", cfg.ScanResultID).Str("account_id", client.Config.AccountID).Msg("cloudflare: error fetching domain scan result")

	return DomainScanPollResult{}, fmt.Errorf("%w: %w", ErrDomainScanResultFailed, err)
}

// getScanResultOnce makes a single attempt at retrieving a URL Scanner result by scan ID
func getScanResultOnce(ctx context.Context, client *CloudflareClient, scanID string) (*url_scanner.ScanGetResponse, ScanTaskErrors, error) {
	result, err := client.URLScanner.Scans.Get(ctx, scanID, url_scanner.ScanGetParams{AccountID: cf.F(client.Config.AccountID)})
	if err != nil {
		return nil, nil, err
	}

	taskErrors := struct {
		Task struct {
			Errors ScanTaskErrors `json:"errors"`
		} `json:"task"`
	}{}
	if err := json.Unmarshal([]byte(result.JSON.RawJSON()), &taskErrors); err != nil {
		return nil, nil, err
	}

	return result, taskErrors.Task.Errors, nil
}
