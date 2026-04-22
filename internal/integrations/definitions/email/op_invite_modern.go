package email

import (
	"html/template"
	"strings"

	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// InviteModernRequest is the input for the modern-themed variant of the organization
// invite operation. It mirrors InviteRequest field-for-field; a distinct type is required
// because OperationSchema keys on the input type name
type InviteModernRequest struct {
	RecipientInfo
	// InviterName is the name of the person who sent the invite
	InviterName string `json:"inviter_name" jsonschema:"required,description=Name of the person sending the invite"`
	// OrgName is the organization being invited to
	OrgName string `json:"org_name" jsonschema:"required,description=Organization name"`
	// Role is the invited role
	Role string `json:"role,omitempty" jsonschema:"description=Invited role"`
	// Token is the invite token appended to the invite URL
	Token string `json:"token" jsonschema:"required,description=Invite token"`
}

// inviteModernSchema is the reflected JSON schema for the modern invite input type
// and InviteModernOp is the typed operation ref used for catalog dispatch
var (
	inviteModernSchema, InviteModernOp = providerkit.OperationSchema[InviteModernRequest]()
)

// inviteModernEmail renders the same content as inviteEmail using the modern-message theme
// so the two layouts can be compared side-by-side
var inviteModernEmail = EmailOperation[InviteModernRequest]{
	Op: InviteModernOp, Schema: inviteModernSchema, Theme: modernMessageTheme,
	Description: "Modern-themed variant of the organization invite email for side-by-side theme comparison",
	Subject: func(cfg RuntimeEmailConfig, req InviteModernRequest) string {
		return "Join Your Teammate " + req.InviterName + " on " + cfg.CompanyName + "!"
	},
	Build: func(cfg RuntimeEmailConfig, req InviteModernRequest) render.ContentBody {
		inviteURL := cfg.ProductURL + "/invite?token=" + req.Token

		intro := "You're in — let's build trust without the busywork. " + req.InviterName + " has invited you to collaborate in " + cfg.CompanyName + ", as part of the <b>" + req.OrgName + "</b> organization"
		if req.Role != "" {
			intro += " with the role of " + strings.ToUpper(req.Role)
		}

		intro += "."

		return render.ContentBody{
			Preheader: "You've been invited to join " + cfg.CompanyName,
			Name:      req.FirstName,
			Title:     "You've been invited to join " + cfg.CompanyName + "!",
			Intros: render.IntrosBlock{
				Unsafe: []template.HTML{
					template.HTML(intro),
					"To get started (and verify your email), click the link below:",
				},
			},
			Actions: []render.Action{{
				Button: render.Button{Text: "Accept Invite", Link: inviteURL},
			}},
			Outros: render.OutrosBlock{
				Paragraphs: []string{
					"Or, if you're feeling old school, copy and paste this link into your browser: " + inviteURL,
					"This link expires in 7 days — but don't worry, if it does, you'll get a fresh one when you try to verify later.",
				},
			},
		}
	},
}
