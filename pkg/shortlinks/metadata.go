package shortlinks

// Purpose names the flow that generated a link; the link service groups clicks by it
type Purpose string

// Purposes for links that carry a signed token
const (
	// PurposeAssessmentAccess is the access link exposed on a system-owned assessment
	PurposeAssessmentAccess Purpose = "assessment_access"
	// PurposeQuestionnaireAuth is the questionnaire login link emailed to a recipient
	PurposeQuestionnaireAuth Purpose = "questionnaire_auth"
	// PurposeTrustCenterNDARequest is the NDA signing link for a trust center access request
	PurposeTrustCenterNDARequest Purpose = "trust_center_nda_request"
	// PurposeTrustCenterNDASigned is the trust center link sent once an NDA is signed
	PurposeTrustCenterNDASigned Purpose = "trust_center_nda_signed"
	// PurposeTrustCenterAuth is the time-limited trust center login link
	PurposeTrustCenterAuth Purpose = "trust_center_auth"
)

// Purposes for tracking-only links to public or console pages that carry no credential
const (
	// PurposeWelcome is the console link in the welcome email
	PurposeWelcome Purpose = "welcome"
	// PurposeOrgInviteJoined is the console link sent once an invite is accepted
	PurposeOrgInviteJoined Purpose = "org_invite_joined"
	// PurposeTrustCenterNDAApproval is the console link sent to an NDA approver
	PurposeTrustCenterNDAApproval Purpose = "trust_center_nda_approval"
	// PurposeOrgDeletionNotice is the billing link in an organization deletion notice
	PurposeOrgDeletionNotice Purpose = "org_deletion_notice"
	// PurposeTrustCenterUpdate is the trust center link in a published-post notification
	PurposeTrustCenterUpdate Purpose = "trust_center_update"
	// PurposeSubprocessorNotification is the trust center link in a subprocessor change notification
	PurposeSubprocessorNotification Purpose = "subprocessor_notification"
)

// Metadata is the attribution the link service records on every click of a link. Core owns
// both ends of this contract, so new attribution is a new field here and a matching column in
// the service, never a free-form key
type Metadata struct {
	// OrganizationID is the organization the link was generated for
	OrganizationID string `json:"organization_id,omitempty"`
	// UserID is the user the link was generated for, when the flow has one
	UserID string `json:"user_id,omitempty"`
	// RecipientEmail is the address the link was delivered to, when a single recipient is known
	RecipientEmail string `json:"recipient_email,omitempty"`
	// TrustCenterID is the trust center the link belongs to, for trust center flows
	TrustCenterID string `json:"trust_center_id,omitempty"`
	// Purpose names the flow that generated the link
	Purpose Purpose `json:"purpose,omitempty"`
}
