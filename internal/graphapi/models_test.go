package graphapi_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

type OrganizationBuilder struct {
	client *client

	// Fields
	Name           string
	DisplayName    string
	Description    *string
	ParentOrgID    string
	PersonalOrg    bool
	AllowedDomains []string
}

type GroupBuilder struct {
	client *client

	// Fields
	Name              string
	ControlEditorsIDs []string
	ProgramEditorsIDs []string
}

type UserBuilder struct {
	client *client

	// Fields
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type TFASettingBuilder struct {
	client *client

	totpAllowed *bool
}

type WebauthnBuilder struct {
	client *client
}

type OrgMemberBuilder struct {
	client *client

	// Fields
	UserID string
	Role   string
}

type GroupMemberBuilder struct {
	client *client

	// Fields
	UserID  string
	GroupID string
	Role    string
}

type InviteBuilder struct {
	client *client

	// Fields
	Recipient string
	Role      string
}

type PersonalAccessTokenBuilder struct {
	client *client

	// Fields
	Name            string
	Token           string
	Abilities       []string
	Description     string
	ExpiresAt       *time.Time
	OrganizationIDs []string
}

type APITokenBuilder struct {
	client *client

	// Fields
	Name        string
	Token       string
	Scopes      []string
	Description string
	ExpiresAt   *time.Time
}

type SubscriberBuilder struct {
	client *client

	// Fields
	Email string
}

type EntityBuilder struct {
	client *client

	// Fields
	Name        string
	DisplayName string
	TypeID      string
	Description string
}

type EntityTypeBuilder struct {
	client *client

	// Fields
	Name string
}

type ContactBuilder struct {
	client *client

	// Fields
	Name    string
	Email   string
	Address string
	Phone   string
	Title   string
	Company string
	Status  enums.UserStatus
}

type TaskBuilder struct {
	client *client

	// Fields
	Title      string
	Details    string
	Status     enums.TaskStatus
	AssigneeID string
	Due        time.Time
	GroupID    string
}

type ProgramBuilder struct {
	client *client

	// Fields
	Name string

	// Create Edges
	WithProcedure bool
	WithPolicy    bool
	// Add Permissions
	EditorIDs       string
	BlockedGroupIDs string
}

type ProgramMemberBuilder struct {
	client *client

	// Fields
	UserID    string
	ProgramID string
	Role      string
}

type ProcedureBuilder struct {
	client *client

	// Fields
	Name    string
	GroupID string
}

type InternalPolicyBuilder struct {
	client *client

	// Fields
	Name    string
	GroupID string
}

type RiskBuilder struct {
	client *client

	// Fields
	Name      string
	ProgramID string
}

type ControlObjectiveBuilder struct {
	client *client

	// Fields
	Name      string
	ProgramID string
}

type NarrativeBuilder struct {
	client *client

	// Fields
	Name      string
	ProgramID string
}

type ControlBuilder struct {
	client *client

	// Fields
	Name                    string
	ProgramID               string
	StandardID              string
	ControlOwnerID          string
	ControlViewerGroupID    string
	ControlEditorGroupID    string
	ControlImplementationID string
	// AllFields will set all direct fields on the control with random data
	AllFields bool
}

type SubcontrolBuilder struct {
	client *client

	// Fields
	Name      string
	ControlID string
}

type EvidenceBuilder struct {
	client *client

	// Fields
	Name      string
	ProgramID string
	ControlID string
}

type StandardBuilder struct {
	client *client

	// Fields
	Name      string
	Framework string
	IsPublic  bool
}

type NoteBuilder struct {
	client *client

	// Fields
	Text    string
	TaskID  string
	FileIDs []string
}

type ControlImplementationBuilder struct {
	client *client

	// Fields
	Details            string
	ImplementationDate time.Time
	ControlIDs         []string
	SubcontrolIDs      []string
}

type MappableDomainBuilder struct {
	client *client

	// Fields
	Name string
}

// Faker structs with random injected data
type Faker struct {
	Name string
}

func randomName(t *testing.T) string {
	var f Faker
	err := gofakeit.Struct(&f)
	require.NoError(t, err)

	var b strings.Builder
	for _, r := range f.Name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String() + "_" + gofakeit.UUID()
}

// DeleteClient is an interface for deleting entities
// client must implement DeleteOneID method that has ExecX method
type DeleteClient[T DeleteExec] interface {
	DeleteOneID(string) T
}

// DeleteExec is an interface for executing delete operations
// that will panic if an error occurs
type DeleteExec interface {
	ExecX(ctx context.Context)
	Exec(ctx context.Context) error
}

// Cleanup is a struct for cleaning up entities
type Cleanup[T DeleteExec] struct {
	client DeleteClient[T]

	// Fields
	ID  string
	IDs []string
}

// MustDelete deletes the entities without authz checks for the type
// this should normally look like this:
// type: generated.OrganizationDeleteOne (replace Organization with the entity you want to delete)
// client: suite.client.db.Organization (replace Organization with the entity you want to delete)
//
//	(&Cleanup[*generated.OrganizationDeleteOne]{
//		client: suite.client.db.Organization,
//		ID: resp.CreateOrganization.Organization.ID}).
//		MustDelete(testUser1.UserCtx, suite)
func (c *Cleanup[DeleteExec]) MustDelete(ctx context.Context, suite *GraphTestSuite) {
	// add client to context for hooks that expect the client to be in the context
	ctx = setContext(ctx, suite.client.db)

	for _, id := range c.IDs {
		err := c.client.DeleteOneID(id).Exec(ctx)
		require.NoError(suite.T(), err)
	}

	if c.ID != "" {
		err := c.client.DeleteOneID(c.ID).Exec(ctx)
		require.NoError(suite.T(), err)
	}
}

// setContext is a helper function to set the context for the client
// setting privacy to allow and adding the client to the context
func setContext(ctx context.Context, db *ent.Client) context.Context {
	return ent.NewContext(rule.WithInternalContext(ctx), db)
}

// MustNew organization builder is used to create, without authz checks, orgs in the database
func (o *OrganizationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Organization {
	// no auth, so allow policy
	ctx = setContext(ctx, o.client.db)

	if o.Name == "" {
		o.Name = randomName(t)
	}

	if o.DisplayName == "" {
		o.DisplayName = gofakeit.LetterN(40)
	}

	if o.Description == nil {
		desc := gofakeit.HipsterSentence(10)
		o.Description = &desc
	}

	m := o.client.db.Organization.Create().SetName(o.Name).SetDescription(*o.Description).SetDisplayName(o.DisplayName).SetPersonalOrg(o.PersonalOrg)

	if o.ParentOrgID != "" {
		m.SetParentID(o.ParentOrgID)
	}

	org, err := m.Save(ctx)
	require.NoError(t, err)

	if o.AllowedDomains != nil {
		orgSetting, err := org.Setting(ctx)
		require.NoError(t, err)

		err = orgSetting.Update().SetAllowedEmailDomains(o.AllowedDomains).Exec(ctx)
		require.NoError(t, err)
	}

	return org
}

// MustNew user builder is used to create, without authz checks, users in the database
func (u *UserBuilder) MustNew(ctx context.Context, t *testing.T) *ent.User {
	ctx = setContext(ctx, u.client.db)

	if u.FirstName == "" {
		u.FirstName = gofakeit.FirstName()
	}

	if u.LastName == "" {
		u.LastName = gofakeit.LastName()
	}

	if u.Email == "" {
		u.Email = gofakeit.Email()
	}

	if u.Password == "" {
		u.Password = gofakeit.Password(true, true, true, true, false, 20)
	}

	// create user setting
	userSetting, err := u.client.db.UserSetting.Create().Save(ctx)
	require.NoError(t, err)

	user, err := u.client.db.User.Create().
		SetFirstName(u.FirstName).
		SetLastName(u.LastName).
		SetEmail(u.Email).
		SetPassword(u.Password).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SetSetting(userSetting).
		Save(ctx)
	require.NoError(t, err)

	_, err = user.Edges.Setting.DefaultOrg(ctx)
	require.NoError(t, err)

	return user
}

// MustNew tfa settings builder is used to create, without authz checks, tfa settings in the database
func (tf *TFASettingBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TFASetting {
	if tf.totpAllowed == nil {
		tf.totpAllowed = lo.ToPtr(true)
	}

	return tf.client.db.TFASetting.Create().
		SetTotpAllowed(*tf.totpAllowed).
		SaveX(ctx)
}

// MustNew webauthn settings builder is used to create passkeys without the browser setup process
func (w *WebauthnBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Webauthn {
	uuidBytes, err := uuid.NewUUID()
	require.NoError(t, err)

	return w.client.db.Webauthn.Create().
		SetAaguid(models.ToAAGUID(uuidBytes[:])).
		SetAttestationType("type").
		SetBackupEligible(true).
		SetBackupState(true).
		SetSignCount(10).
		SetCredentialID([]byte(uuid.NewString())).
		SetTransports([]string{uuid.NewString()}).
		SaveX(ctx)
}

// MustNew org members builder is used to create, without authz checks, org members in the database
func (om *OrgMemberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.OrgMembership {
	ctx = setContext(ctx, om.client.db)

	if om.UserID == "" {
		user := (&UserBuilder{client: om.client}).MustNew(ctx, t)
		om.UserID = user.ID
	}

	role := enums.ToRole(om.Role)
	if role == &enums.RoleInvalid {
		role = &enums.RoleMember
	}

	orgMembers := om.client.db.OrgMembership.Create().
		SetUserID(om.UserID).
		SetRole(*role).
		SaveX(ctx)

	return orgMembers
}

// MustNew group builder is used to create, without authz checks, groups in the database
func (g *GroupBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Group {
	ctx = setContext(ctx, g.client.db)

	if g.Name == "" {
		g.Name = randomName(t)
	}

	mutation := g.client.db.Group.Create().SetName(g.Name)

	if len(g.ControlEditorsIDs) > 0 {
		mutation.AddControlEditorIDs(g.ControlEditorsIDs...)
	}

	if len(g.ProgramEditorsIDs) > 0 {
		mutation.AddProgramEditorIDs(g.ProgramEditorsIDs...)
	}

	group, err := mutation.Save(ctx)
	require.NoError(t, err)

	return group
}

// MustNew invite builder is used to create, without authz checks, invites in the database
func (i *InviteBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Invite {
	ctx = setContext(ctx, i.client.db)

	// create user if not provided
	rec := i.Recipient

	if rec == "" {
		rec = gofakeit.Email()
	}

	inviteQuery := i.client.db.Invite.Create().
		SetRecipient(rec)

	if i.Role != "" {
		inviteQuery.SetRole(*enums.ToRole(i.Role))
	}

	invite := inviteQuery.SaveX(ctx)

	return invite
}

// MustNew subscriber builder is used to create, without authz checks, subscribers in the database
func (i *SubscriberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Subscriber {
	reqCtx := setContext(ctx, i.client.db)

	// create user if not provided
	rec := i.Email

	if rec == "" {
		rec = gofakeit.Email()
	}

	sub := i.client.db.Subscriber.Create().
		SetEmail(rec).
		SetActive(true).SaveX(reqCtx)

	return sub
}

// MustNew personal access tokens builder is used to create, without authz checks, personal access tokens in the database
func (pat *PersonalAccessTokenBuilder) MustNew(ctx context.Context, t *testing.T) *ent.PersonalAccessToken {
	ctx = setContext(ctx, pat.client.db)

	if pat.Name == "" {
		pat.Name = gofakeit.AppName()
	}

	if pat.Description == "" {
		pat.Description = gofakeit.HipsterSentence(5)
	}

	if pat.OrganizationIDs == nil {
		org := (&OrganizationBuilder{client: pat.client}).MustNew(ctx, t)
		pat.OrganizationIDs = []string{org.ID}
	}

	request := pat.client.db.PersonalAccessToken.Create().
		SetName(pat.Name).
		SetDescription(pat.Description).
		AddOrganizationIDs(pat.OrganizationIDs...)

	if pat.ExpiresAt != nil {
		request.SetExpiresAt(*pat.ExpiresAt)
	}

	token := request.SaveX(ctx)

	return token
}

// MustNew api tokens builder is used to create, without authz checks, api tokens in the database
func (at *APITokenBuilder) MustNew(ctx context.Context, t *testing.T) *ent.APIToken {
	ctx = setContext(ctx, at.client.db)

	if at.Name == "" {
		at.Name = gofakeit.AppName()
	}

	if at.Description == "" {
		at.Description = gofakeit.HipsterSentence(5)
	}

	if at.Scopes == nil {
		at.Scopes = []string{"read", "write"}
	}

	request := at.client.db.APIToken.Create().
		SetName(at.Name).
		SetDescription(at.Description).
		SetScopes(at.Scopes)

	if at.ExpiresAt != nil {
		request.SetExpiresAt(*at.ExpiresAt)
	}

	token := request.SaveX(ctx)

	return token
}

// MustNew user builder is used to create, without authz checks, group members in the database
func (gm *GroupMemberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.GroupMembership {
	ctx = setContext(ctx, gm.client.db)

	if gm.GroupID == "" {
		group := (&GroupBuilder{client: gm.client}).MustNew(ctx, t)
		gm.GroupID = group.ID
	}

	if gm.UserID == "" {
		orgMember := (&OrgMemberBuilder{client: gm.client}).MustNew(testUser1.UserCtx, t)
		gm.UserID = orgMember.UserID
	}

	groupMember, err := gm.client.db.GroupMembership.Create().
		SetUserID(gm.UserID).
		SetGroupID(gm.GroupID).
		Save(ctx)
	require.NoError(t, err)

	return groupMember
}

// MustNew entity type builder is used to create, without authz checks, entity types in the database
func (e *EntityTypeBuilder) MustNew(ctx context.Context, t *testing.T) *ent.EntityType {
	ctx = setContext(ctx, e.client.db)

	if e.Name == "" {
		e.Name = randomName(t)
	}

	entityType := e.client.db.EntityType.Create().
		SetName(e.Name).
		SaveX(ctx)

	return entityType
}

// MustNew entity builder is used to create, without authz checks, entities in the database
func (e *EntityBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Entity {
	ctx = setContext(ctx, e.client.db)

	if e.Name == "" {
		e.Name = gofakeit.AppName()
	}

	if e.DisplayName == "" {
		e.DisplayName = e.Name
	}

	if e.Description == "" {
		e.Description = gofakeit.HipsterSentence(5)
	}

	if e.TypeID == "" {
		et := (&EntityTypeBuilder{client: e.client}).MustNew(ctx, t)
		e.TypeID = et.ID
	}

	entity := e.client.db.Entity.Create().
		SetName(e.Name).
		SetDisplayName(e.DisplayName).
		SetEntityTypeID(e.TypeID).
		SetDescription(e.Description).
		SaveX(ctx)

	return entity
}

// MustNew contact builder is used to create, without authz checks, contacts in the database
func (c *ContactBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Contact {
	ctx = setContext(ctx, c.client.db)

	if c.Name == "" {
		c.Name = gofakeit.AppName()
	}

	if c.Email == "" {
		c.Email = gofakeit.Email()
	}

	if c.Phone == "" {
		c.Phone = gofakeit.Phone()
	}

	if c.Address == "" {
		address := gofakeit.Address()
		c.Address = fmt.Sprintf("%s, %s, %s, %s", address.Street, address.City, address.State, address.Zip)
	}

	if c.Title == "" {
		c.Title = gofakeit.JobTitle()
	}

	if c.Company == "" {
		c.Company = gofakeit.Company()
	}

	entity := c.client.db.Contact.Create().
		SetFullName(c.Name).
		SetEmail(c.Email).
		SetPhoneNumber(c.Phone).
		SetAddress(c.Address).
		SetTitle(c.Title).
		SetCompany(c.Company).
		SaveX(ctx)

	return entity
}

// MustNew task builder is used to create, without authz checks, tasks in the database
func (c *TaskBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Task {
	ctx = setContext(ctx, c.client.db)

	if c.Title == "" {
		c.Title = gofakeit.AppName()
	}

	if c.Details == "" {
		c.Details = gofakeit.HipsterSentence(5)
	}

	taskCreate := c.client.db.Task.Create().
		SetTitle(c.Title).
		SetDetails(c.Details)

	if c.Status != "" {
		taskCreate.SetStatus(c.Status)
	}

	if c.AssigneeID != "" {
		taskCreate.SetAssigneeID(c.AssigneeID)
	}

	if !c.Due.IsZero() {
		taskCreate.SetDue(models.DateTime(c.Due))
	}

	if c.GroupID != "" {
		taskCreate.AddGroupIDs(c.GroupID)
	}

	task := taskCreate.SaveX(ctx)

	return task
}

// MustNew program builder is used to create, without authz checks, programs in the database
func (p *ProgramBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Program {
	ctx = setContext(ctx, p.client.db)

	if p.Name == "" {
		p.Name = gofakeit.AppName()
	}

	mutation := p.client.db.Program.Create().
		SetName(p.Name)

	if p.WithProcedure {
		procedure := (&ProcedureBuilder{client: p.client, Name: gofakeit.AppName()}).MustNew(ctx, t)
		mutation.AddProcedureIDs(procedure.ID)
	}

	if p.WithPolicy {
		policy := (&InternalPolicyBuilder{client: p.client, Name: gofakeit.AppName()}).MustNew(ctx, t)
		mutation.AddInternalPolicyIDs(policy.ID)
	}

	if p.EditorIDs != "" {
		mutation.AddEditorIDs(p.EditorIDs)
	}

	if p.BlockedGroupIDs != "" {
		mutation.AddBlockedGroupIDs(p.BlockedGroupIDs)
	}

	program := mutation.
		SaveX(ctx)

	return program
}

// MustNew user builder is used to create, without authz checks, program members in the database
func (pm *ProgramMemberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.ProgramMembership {
	ctx = setContext(ctx, pm.client.db)

	if pm.ProgramID == "" {
		program := (&ProgramBuilder{client: pm.client}).MustNew(ctx, t)
		pm.ProgramID = program.ID
	}

	if pm.UserID == "" {
		// first create an org member
		orgMember := (&OrgMemberBuilder{client: pm.client}).MustNew(testUser1.UserCtx, t)
		pm.UserID = orgMember.UserID
	}

	mutation := pm.client.db.ProgramMembership.Create().
		SetUserID(pm.UserID).
		SetProgramID(pm.ProgramID)

	if pm.Role != "" {
		mutation.SetRole(*enums.ToRole(pm.Role))
	}

	programMember := mutation.
		SaveX(ctx)

	return programMember
}

// MustNew procedure builder is used to create, without authz checks, procedures in the database
func (p *ProcedureBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Procedure {
	ctx = setContext(ctx, p.client.db)

	if p.Name == "" {
		p.Name = gofakeit.AppName()
	}

	mutation := p.client.db.Procedure.Create().
		SetName(p.Name)

	if p.GroupID != "" {
		mutation.AddEditorIDs(p.GroupID)
	}

	procedure := mutation.
		SaveX(ctx)

	return procedure
}

// MustNew policy builder is used to create, without authz checks, policies in the database
func (p *InternalPolicyBuilder) MustNew(ctx context.Context, t *testing.T) *ent.InternalPolicy {
	ctx = setContext(ctx, p.client.db)

	if p.Name == "" {
		p.Name = gofakeit.AppName()
	}

	policy := p.client.db.InternalPolicy.Create().
		SetName(p.Name).
		SaveX(ctx)

	return policy
}

// MustNew risk builder is used to create, without authz checks, risks in the database
func (r *RiskBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Risk {
	ctx = setContext(ctx, r.client.db)

	if r.Name == "" {
		r.Name = gofakeit.AppName()
	}

	mutation := r.client.db.Risk.Create().
		SetName(r.Name)

	if r.ProgramID != "" {
		mutation.AddProgramIDs(r.ProgramID)
	}

	risk := mutation.
		SaveX(ctx)

	return risk
}

// MustNew control objective builder is used to create, without authz checks, control objectives in the database
func (c *ControlObjectiveBuilder) MustNew(ctx context.Context, t *testing.T) *ent.ControlObjective {
	ctx = setContext(ctx, c.client.db)

	if c.Name == "" {
		c.Name = gofakeit.AppName()
	}

	mutation := c.client.db.ControlObjective.Create().
		SetName(c.Name)

	if c.ProgramID != "" {
		mutation.AddProgramIDs(c.ProgramID)
	}

	co := mutation.
		SaveX(ctx)

	return co
}

// MustNew narrative builder is used to create, without authz checks, narratives in the database
func (n *NarrativeBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Narrative {
	ctx = setContext(ctx, n.client.db)

	if n.Name == "" {
		n.Name = gofakeit.AppName()
	}

	mutation := n.client.db.Narrative.Create().
		SetName(n.Name)

	if n.ProgramID != "" {
		mutation.AddProgramIDs(n.ProgramID)
	}

	narrative := mutation.
		SaveX(ctx)

	return narrative
}

// MustNew control builder is used to create, without authz checks, controls in the database
func (c *ControlBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Control {
	ctx = setContext(ctx, c.client.db)

	if c.Name == "" {
		c.Name = gofakeit.UUID()
	}

	mutation := c.client.db.Control.Create().
		SetRefCode(c.Name)

	if c.ProgramID != "" {
		mutation.AddProgramIDs(c.ProgramID)
	}

	if c.StandardID != "" {
		mutation.SetStandardID(c.StandardID)
	}

	if c.ControlOwnerID != "" {
		mutation.SetControlOwnerID(c.ControlOwnerID)
	}

	if c.ControlViewerGroupID != "" {
		mutation.AddViewerIDs(c.ControlViewerGroupID)
	}

	if c.ControlEditorGroupID != "" {
		mutation.AddEditorIDs(c.ControlEditorGroupID)
	}

	if c.ControlImplementationID != "" {
		mutation.AddControlImplementationIDs(c.ControlImplementationID)
	}

	if c.AllFields {
		mutation.SetDescription(gofakeit.HipsterSentence(5)).
			SetCategory(gofakeit.Adjective()).
			SetCategoryID("A").
			SetSubcategory(gofakeit.Adjective()).
			SetControlType(enums.ControlTypeDetective).
			SetExampleEvidence([]models.ExampleEvidence{
				{
					DocumentationType: "Documentation",
					Description:       gofakeit.HipsterSentence(5),
				},
			}).
			SetImplementationGuidance([]models.ImplementationGuidance{
				{
					ReferenceID: "A",
					Guidance: []string{
						gofakeit.HipsterSentence(5),
						gofakeit.HipsterSentence(5),
					},
				},
			}).
			SetMappedCategories([]string{"Governance", "Risk Management"}).
			SetTags([]string{"tag1", "tag2"}).
			SetSource(enums.ControlSourceFramework).
			SetReferences([]models.Reference{
				{
					Name: gofakeit.HipsterSentence(5),
					URL:  gofakeit.URL(),
				},
			})
	}

	control := mutation.
		SaveX(ctx)

	return control
}

// MustNew subcontrol builder is used to create, without authz checks, subcontrols in the database
func (s *SubcontrolBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Subcontrol {
	ctx = setContext(ctx, s.client.db)

	if s.Name == "" {
		s.Name = gofakeit.UUID()
	}

	mutation := s.client.db.Subcontrol.Create().
		SetRefCode(s.Name)

	if s.ControlID == "" {
		control := (&ControlBuilder{client: s.client}).MustNew(ctx, t)
		s.ControlID = control.ID
	}

	mutation.SetControlID(s.ControlID)

	sc := mutation.
		SaveX(ctx)

	return sc
}

// MustNew control builder is used to create, without authz checks, controls in the database
func (c *EvidenceBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Evidence {
	ctx = setContext(ctx, c.client.db)

	if c.Name == "" {
		c.Name = gofakeit.AppName()
	}

	mutation := c.client.db.Evidence.Create().
		SetName(c.Name)

	if c.ProgramID != "" {
		mutation.AddProgramIDs(c.ProgramID)
	}

	if c.ControlID != "" {
		mutation.AddControlIDs(c.ControlID)
	}

	control := mutation.
		SaveX(ctx)

	return control
}

// MustNew standard builder is used to create, without authz checks, standards in the database
func (s *StandardBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Standard {
	ctx = setContext(ctx, s.client.db)

	if s.Name == "" {
		s.Name = gofakeit.AppName()
	}

	if s.Framework == "" {
		s.Framework = "MITB Framework"
	}

	Standard := s.client.db.Standard.Create().
		SetName(s.Name).
		SetFramework(s.Framework).
		SetIsPublic(s.IsPublic).
		SaveX(ctx)

	return Standard
}

// MustNew note builder is used to create, without authz checks, notes in the database
func (n *NoteBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Note {
	ctx = setContext(ctx, n.client.db)

	if n.Text == "" {
		n.Text = gofakeit.HipsterSentence(10)
	}

	mutation := n.client.db.Note.Create().
		SetText(n.Text)

	if n.TaskID != "" {
		mutation.SetTaskID(n.TaskID)
	}

	if len(n.FileIDs) > 0 {
		mutation.AddFileIDs(n.FileIDs...)
	}

	note := mutation.SaveX(ctx)

	return note
}

// MustNew controlImplementation builder is used to create, without authz checks, controlImplementations in the database
func (e *ControlImplementationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.ControlImplementation {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if e.Details == "" {
		e.Details = gofakeit.Paragraph(3, 4, 300, "<br />")
	}

	if e.ImplementationDate.IsZero() {
		e.ImplementationDate = time.Now()
	}

	mutation := e.client.db.ControlImplementation.Create().
		SetDetails(e.Details).
		SetImplementationDate(e.ImplementationDate)

	if len(e.ControlIDs) > 0 {
		mutation.AddControlIDs(e.ControlIDs...)
	}

	if len(e.SubcontrolIDs) > 0 {
		mutation.AddSubcontrolIDs(e.SubcontrolIDs...)
	}

	controlImplementation, err := mutation.
		Save(ctx)

	require.NoError(t, err)

	return controlImplementation
}

// MustNew mappable domain builder is used to create, without authz checks, mappable domains in the database
func (e *MappableDomainBuilder) MustNew(ctx context.Context, t *testing.T) *ent.MappableDomain {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if e.Name == "" {
		e.Name = gofakeit.DomainName()
	}

	mappableDomain, err := e.client.db.MappableDomain.Create().
		SetName(e.Name).
		Save(ctx)
	require.NoError(t, err)

	return mappableDomain
}

// CustomDomainBuilder is used to create custom domains
type CustomDomainBuilder struct {
	client *client

	// Fields
	CnameRecord      string
	MappableDomainID string
	OwnerID          string
}

// MustNew custom domain builder is used to create, without authz checks, custom domains in the database
func (c *CustomDomainBuilder) MustNew(ctx context.Context, t *testing.T, status *enums.CustomDomainStatus) *ent.CustomDomain {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if c.CnameRecord == "" {
		c.CnameRecord = gofakeit.DomainName()
	}

	if c.MappableDomainID == "" {
		mappableDomain := (&MappableDomainBuilder{client: c.client}).MustNew(ctx, t)
		c.MappableDomainID = mappableDomain.ID
	}

	if c.OwnerID == "" {
		// Use the organization ID from the test user
		c.OwnerID = testUser1.OrganizationID
	}

	if status == nil {
		status = lo.ToPtr(enums.CustomDomainStatusPending)
	}

	customDomain, err := c.client.db.CustomDomain.Create().
		SetCnameRecord(c.CnameRecord).
		SetMappableDomainID(c.MappableDomainID).
		SetOwnerID(c.OwnerID).
		SetStatus(*status).
		Save(ctx)
	require.NoError(t, err)

	return customDomain
}
