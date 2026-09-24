// Package modelarmor screens documents through Google Cloud Model Armor before they reach a model
package modelarmor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/auth"
	armor "cloud.google.com/go/modelarmor/apiv1"
	"cloud.google.com/go/modelarmor/apiv1/modelarmorpb"
	"google.golang.org/api/option"
)

const (
	// templateLocationIndex is where the location sits in projects/x/locations/y/templates/z
	templateLocationIndex = 3
	// templateNameParts is how many segments a full template resource name has
	templateNameParts = 6
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
	// Blocked is true when any filter matched
	Blocked bool
	// Filters names the filters that matched
	Filters []string
}

// New builds a client for the template, authenticating with the given credentials or
// application default credentials when nil; sanitize calls are served from the template's region
func New(ctx context.Context, template string, creds *auth.Credentials) (*Client, error) {
	parts := strings.Split(template, "/")
	if len(parts) != templateNameParts || parts[0] != "projects" || parts[2] != "locations" || parts[4] != "templates" {
		return nil, ErrTemplateInvalid
	}

	opts := []option.ClientOption{option.WithEndpoint(fmt.Sprintf("modelarmor.%s.rep.googleapis.com:443", parts[templateLocationIndex]))}
	if creds != nil {
		opts = append(opts, option.WithAuthCredentials(creds))
	}

	api, err := armor.NewClient(ctx, opts...)
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

// ScreenDocument submits the pdf to Model Armor and reports whether any filter matched
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

	result := resp.GetSanitizationResult()
	verdict := Verdict{Blocked: result.GetFilterMatchState() == modelarmorpb.FilterMatchState_MATCH_FOUND}

	for name, filter := range result.GetFilterResults() {
		if filterMatched(filter) {
			verdict.Filters = append(verdict.Filters, name)
		}
	}

	return verdict, nil
}

// filterMatched reports whether the filter's own result carries a match, whichever kind it is
func filterMatched(filter *modelarmorpb.FilterResult) bool {
	states := []modelarmorpb.FilterMatchState{
		filter.GetRaiFilterResult().GetMatchState(),
		filter.GetPiAndJailbreakFilterResult().GetMatchState(),
		filter.GetMaliciousUriFilterResult().GetMatchState(),
		filter.GetCsamFilterFilterResult().GetMatchState(),
		filter.GetVirusScanFilterResult().GetMatchState(),
		filter.GetSdpFilterResult().GetInspectResult().GetMatchState(),
		filter.GetSdpFilterResult().GetDeidentifyResult().GetMatchState(),
	}

	for _, state := range states {
		if state == modelarmorpb.FilterMatchState_MATCH_FOUND {
			return true
		}
	}

	return false
}
