// Package modelarmor screens documents through Google Cloud Model Armor before they reach a model
package modelarmor

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	armor "cloud.google.com/go/modelarmor/apiv1"
	"cloud.google.com/go/modelarmor/apiv1/modelarmorpb"
	"google.golang.org/api/option"
)

const (
	// templateLocationIndex is where the location sits in projects/x/locations/y/templates/z
	templateLocationIndex = 3
	// templateNameParts is how many segments a full template resource name has
	templateNameParts = 6
	// modelarmourendpoint is the model armour api endpoint
	modelarmorendpoint = "modelarmor.%s.rep.googleapis.com:443"
)

var (
	// ErrTemplateInvalid is returned when the template is not a full resource name
	ErrTemplateInvalid = errors.New("modelarmor: template must be projects/<project>/locations/<location>/templates/<template>")
	// ErrDocumentBlocked is returned when Model Armor matched a filter against the document
	ErrDocumentBlocked = errors.New("the uploaded file contains content that looks like instructions to an AI system and was rejected")
	// ErrSanitizeFailed is returned when the sanitize request itself failed
	ErrSanitizeFailed = errors.New("modelarmor: sanitize request failed")
)

// Client screens documents against one Model Armor template
type Client struct {
	api      *armor.Client
	template string
}

// Verdict is the outcome of screening one document
type Verdict struct {
	// Blocked is true when a filter that must stop the upload matched
	Blocked bool
	// Filters names the matched filters that stop the upload
	Filters []string
	// Reported names the matched filters that do not stop the upload; sensitive data detection is
	// reported so the caller can act on it under its own data policy rather than rejecting the file
	Reported []string
}

// New builds a client for the template, authenticating with application default credentials so
// the workload's own identity is used; sanitize calls are served from the template's region
func New(ctx context.Context, template string) (*Client, error) {
	parts := strings.Split(template, "/")
	if len(parts) != templateNameParts || parts[0] != "projects" || parts[2] != "locations" || parts[4] != "templates" {
		return nil, ErrTemplateInvalid
	}

	endpoint := option.WithEndpoint(fmt.Sprintf(modelarmorendpoint, parts[templateLocationIndex]))

	api, err := armor.NewClient(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	return &Client{api: api, template: template}, nil
}

// Template returns the template the client screens against
func (c *Client) Template() string {
	return c.template
}

// Close releases the underlying connection
func (c *Client) Close() error {
	return c.api.Close()
}

// ScreenDocument submits the pdf to Model Armor and reports which filters matched; only the filters
// that make a document unsafe to send to a model stop the upload
func (c *Client) ScreenDocument(ctx context.Context, pdf []byte) (Verdict, error) {
	resp, err := c.api.SanitizeUserPrompt(ctx, &modelarmorpb.SanitizeUserPromptRequest{
		Name: c.template,
		UserPromptData: &modelarmorpb.DataItem{
			DataItem: &modelarmorpb.DataItem_ByteItem{
				ByteItem: &modelarmorpb.ByteDataItem{ByteDataType: modelarmorpb.ByteDataItem_PDF, ByteData: pdf},
			},
		},
	})
	if err != nil {
		return Verdict{}, fmt.Errorf("%w: %w", ErrSanitizeFailed, err)
	}

	var verdict Verdict

	// the overall match state is set by any filter, so each filter's own result decides the outcome
	for name, filter := range resp.GetSanitizationResult().GetFilterResults() {
		switch {
		case blockingMatch(filter):
			verdict.Blocked = true
			verdict.Filters = append(verdict.Filters, name)
		case sensitiveDataMatch(filter):
			verdict.Reported = append(verdict.Reported, name)
		}
	}

	slices.Sort(verdict.Filters)
	slices.Sort(verdict.Reported)

	return verdict, nil
}

// blockingMatch reports whether the filter found something that makes the document unsafe to send
// to a model, which is every filter except sensitive data detection
func blockingMatch(filter *modelarmorpb.FilterResult) bool {
	return matched(
		filter.GetPiAndJailbreakFilterResult().GetMatchState(),
		filter.GetRaiFilterResult().GetMatchState(),
		filter.GetMaliciousUriFilterResult().GetMatchState(),
		filter.GetCsamFilterFilterResult().GetMatchState(),
		filter.GetVirusScanFilterResult().GetMatchState(),
	)
}

// sensitiveDataMatch reports whether the filter found sensitive data, which a report legitimately
// contains and which says nothing about whether the document is trying to instruct a model
func sensitiveDataMatch(filter *modelarmorpb.FilterResult) bool {
	return matched(
		filter.GetSdpFilterResult().GetInspectResult().GetMatchState(),
		filter.GetSdpFilterResult().GetDeidentifyResult().GetMatchState(),
	)
}

// matched reports whether any of the states carries a match
func matched(states ...modelarmorpb.FilterMatchState) bool {
	for _, state := range states {
		if state == modelarmorpb.FilterMatchState_MATCH_FOUND {
			return true
		}
	}

	return false
}
