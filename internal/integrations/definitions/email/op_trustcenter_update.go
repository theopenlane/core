package email

import (
	"html/template"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TrustCenterUpdateTemplate is the dispatcher registration key for the trust center post
// notification; trust center update campaigns render through it directly with no linked
// email template
const TrustCenterUpdateTemplate = "tmpl_trustcenter_update"

// DefaultTrustCenterUpdateTitle backs the post notification subject and post heading when the post
// has no title and the trust center branding carries no company name
const DefaultTrustCenterUpdateTitle = "Trust center update"

// trustCenterUpdateButtonText is the call-to-action label linking back to the trust center
const trustCenterUpdateButtonText = "View trust center"

// TrustCenterUpdateRequest is the message notifying trust center subscribers about a newly published
// post. The caller supplies only data (the post, links, and trust center branding) while the subject
// and body copy are composed by the operation, leading with the trust center's company name when
// branded. Recipient and campaign fields are populated per send by the campaign dispatch, so they are
// excluded from the reflected schema
type TrustCenterUpdateRequest struct {
	RecipientInfo   `jsonschema:"-"`
	CampaignContext `jsonschema:"-"`
	// TrustCenterBranding is the trust center's visual identity overlay; empty values fall back to the
	// Openlane system branding
	TrustCenterBranding
	// PostTitle is the published post's title, shown as the post heading and in the subject
	PostTitle string `json:"postTitle,omitempty" jsonschema:"description=Published post title"`
	// PostText is the published post's body, rendered as paragraphs under the post heading
	PostText string `json:"postText,omitempty" jsonschema:"description=Published post body text"`
	// TrustCenterURL is the public trust center link the call-to-action button navigates to
	TrustCenterURL string `json:"trustCenterURL,omitempty" jsonschema:"format=uri,description=Public trust center URL for the call-to-action button"`
	// UnsubscribeURL is the unsubscribe link shown in the footer; campaign sends carry the
	// {{ .unsubscribeToken }} placeholder interpolated per recipient at render time
	UnsubscribeURL string `json:"unsubscribeURL,omitempty" jsonschema:"description=Unsubscribe link shown in the footer"`
}

// trustCenterUpdateSchema is the reflected JSON schema for the trust center update input type and
// TrustCenterUpdateOp is the typed operation ref registered under the durable template key
var (
	trustCenterUpdateSchema = jsonx.SchemaFrom[TrustCenterUpdateRequest]()
	// TrustCenterUpdateOp is the typed operation ref for the trust center post notification
	TrustCenterUpdateOp = types.NewOperationRef[TrustCenterUpdateRequest](TrustCenterUpdateTemplate)
)

var _ = RegisterEmailOperation(Operation[TrustCenterUpdateRequest]{
	Op: TrustCenterUpdateOp, Schema: trustCenterUpdateSchema, Theme: baseTheme,
	Description: "System notification announcing a new trust center post to subscribers",
	Example: TrustCenterUpdateRequest{
		RecipientInfo: RecipientInfo{
			Email:     "jordan.avery@example.com",
			FirstName: "Jordan",
			LastName:  "Avery",
		},
		TrustCenterBranding: TrustCenterBranding{
			CompanyName: "Acme Security",
			AccentColor: "#3fc2b4",
		},
		PostTitle:      "SOC 2 Type II report published",
		PostText:       "Our latest SOC 2 Type II report is now available in the trust center.\nRequest access to review the full report.",
		TrustCenterURL: "https://trust.example.com",
	},
	Subject: func(_ RuntimeEmailConfig, req TrustCenterUpdateRequest) string {
		return trustCenterUpdateSubject(req.CompanyName, req.PostTitle)
	},
	Build: func(_ RuntimeEmailConfig, req TrustCenterUpdateRequest) render.ContentBody {
		// the subject already announces the update, so the body carries a single attribution line
		// under the post title and then the post content as it was written
		intro := "A new update was published to the trust center."
		if req.CompanyName != "" {
			intro = req.CompanyName + " has published a new update to their trust center."
		}

		body := render.ContentBody{
			Name:          req.FirstName,
			Title:         lo.CoalesceOrEmpty(req.PostTitle, DefaultTrustCenterUpdateTitle),
			Intros:        render.IntrosBlock{Paragraphs: []string{intro}},
			ContentBlocks: postContentBlocks(req.PostText),
		}

		if req.TrustCenterURL != "" {
			body.Actions = []render.Action{{
				Button: render.Button{Text: trustCenterUpdateButtonText, Link: req.TrustCenterURL},
			}}
		}

		return body
	},
	Config: func(cfg RuntimeEmailConfig, req TrustCenterUpdateRequest) RuntimeEmailConfig {
		return trustCenterEmailConfig(cfg, req.TrustCenterBranding, req.UnsubscribeURL)
	},
})

// trustCenterUpdateSubject builds a post notification subject naming the trust center's company and
// the post topic when present, so subscribers can tell who the update is from and what it covers,
// mirroring how the NDA emails lead with the trust center's company name
func trustCenterUpdateSubject(companyName, title string) string {
	subject := DefaultTrustCenterUpdateTitle
	if companyName != "" {
		subject = companyName + " Trust Center Update"
	}

	if title == "" {
		return subject
	}

	return subject + ": " + title
}

// postContentBlocks renders the post body as it was written: HTML posts (the trust center editor
// stores rendered HTML) are sanitized with the shared email scrubber and injected as-is so authored
// formatting like lists and emphasis is preserved, while plain-text posts are escaped and wrapped
// into paragraphs
func postContentBlocks(text string) []template.HTML {
	if strings.Contains(text, "<") && strings.Contains(text, ">") {
		return []template.HTML{template.HTML(EmailScrubber().Scrub(text))} //nolint:gosec
	}

	return lo.FilterMap(strings.Split(text, "\n"), func(line string, _ int) (template.HTML, bool) {
		line = strings.TrimSpace(line)
		if line == "" {
			return "", false
		}

		return template.HTML("<p>" + template.HTMLEscapeString(line) + "</p>"), true //nolint:gosec
	})
}

// TrustCenterUpdateContent converts the update request to the campaign metadata map persisted on
// automated trust center update campaigns and rendered per recipient at send time. Recipient fields
// are per-target, so the empty recipient email key is dropped from the stored content
func TrustCenterUpdateContent(req TrustCenterUpdateRequest) (map[string]any, error) {
	content, err := jsonx.ToMap(req)
	if err != nil {
		return nil, err
	}

	delete(content, "email")

	return content, nil
}
