package graphapi_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
)

type OrganizationBuilder struct {
	client *client

	// Fields
	Name           string
	DisplayName    string
	Description    *string
	OrgID          string
	ParentOrgID    string
	PersonalOrg    bool
	AllowedDomains []string
}

type OrganizationCleanup struct {
	client *client

	// Fields
	ID string
}

type GroupBuilder struct {
	client *client

	// Fields
	Name              string
	Owner             string
	ControlEditorsIDs []string
	ProgramEditorsIDs []string
}

type GroupCleanup struct {
	client *client

	// Fields
	ID string
}

type UserBuilder struct {
	client *client

	// Fields
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type UserCleanup struct {
	client *client

	// Fields
	ID string
}

type TFASettingBuilder struct {
	client *client

	totpAllowed *bool
}

type OrgMemberBuilder struct {
	client *client

	// Fields
	UserID string
	OrgID  string
	Role   string
}

type OrgMemberCleanup struct {
	client *client

	// Fields
	ID string
}

type GroupMemberBuilder struct {
	client *client

	// Fields
	UserID  string
	GroupID string
	Role    string
}

type GroupMemberCleanup struct {
	client *client

	// Fields
	ID string
}

type InviteBuilder struct {
	client *client

	// Fields
	Recipient string
	Role      string
}

type InviteCleanup struct {
	client *client

	// Fields
	ID string
}

type PersonalAccessTokenBuilder struct {
	client *client

	// Fields
	Name            string
	Token           string
	Abilities       []string
	Description     string
	ExpiresAt       *time.Time
	OwnerID         string
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
	OwnerID     string
}

type SubscriberBuilder struct {
	client *client

	// Fields
	Email string
	OrgID string
}

type EntityBuilder struct {
	client *client

	// Fields
	Name        string
	DisplayName string
	TypeID      string
	Description string
	Owner       string
}

type EntityCleanup struct {
	client *client

	// Fields
	ID string
}

type EntityTypeBuilder struct {
	client *client

	// Fields
	Name string
}

type EntityTypeCleanup struct {
	client *client

	// Fields
	ID string
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

type ContactCleanup struct {
	client *client

	// Fields
	ID string
}

type TaskBuilder struct {
	client *client

	// Fields
	Title          string
	Description    string
	Status         enums.TaskStatus
	AssigneeID     string
	Due            time.Time
	OrganizationID string
	GroupID        string
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
	Name      string
	ProgramID string
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

// Faker structs with random injected data
type Faker struct {
	Name string
}

func randomName(t *testing.T) string {
	var f Faker
	err := gofakeit.Struct(&f)
	require.NoError(t, err)

	return f.Name
}

// MustNew organization builder is used to create, without authz checks, orgs in the database
func (o *OrganizationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Organization {
	// no auth, so allow policy
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// add client to context
	// the organization hook expects the client to be in the context
	// which happens automatically when using the graph resolvers
	ctx = ent.NewContext(ctx, o.client.db)

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

// MustDelete is used to cleanup, without authz checks, orgs in the database
func (o *OrganizationCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	o.client.db.Organization.DeleteOneID(o.ID).ExecX(ctx)
}

// MustNew user builder is used to create, without authz checks, users in the database
func (u *UserBuilder) MustNew(ctx context.Context, t *testing.T) *ent.User {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

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
		SetSetting(userSetting).
		Save(ctx)
	require.NoError(t, err)

	_, err = user.Edges.Setting.DefaultOrg(ctx)
	require.NoError(t, err)

	return user
}

// MustDelete is used to cleanup, without authz checks, users in the database
func (u *UserCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	u.client.db.User.DeleteOneID(u.ID).ExecX(ctx)
}

// MustNew tfa settings builder is used to create, without authz checks, tfa settings in the database
func (tf *TFASettingBuilder) MustNew(ctx context.Context, t *testing.T, userID string) *ent.TFASetting {
	if tf.totpAllowed == nil {
		tf.totpAllowed = lo.ToPtr(true)
	}

	return tf.client.db.TFASetting.Create().
		SetTotpAllowed(*tf.totpAllowed).
		SetOwnerID(userID).
		SaveX(ctx)
}

// MustNew org members builder is used to create, without authz checks, org members in the database
func (om *OrgMemberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.OrgMembership {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if om.OrgID == "" {
		org := (&OrganizationBuilder{client: om.client}).MustNew(ctx, t)
		om.OrgID = org.ID
	}

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
		SetOrganizationID(om.OrgID).
		SetRole(*role).
		SaveX(ctx)

	return orgMembers
}

// MustDelete is used to cleanup, without authz checks, org members in the database
func (om *OrgMemberCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	om.client.db.OrgMembership.DeleteOneID(om.ID).ExecX(ctx)
}

// MustNew group builder is used to create, without authz checks, groups in the database
func (g *GroupBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Group {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if g.Name == "" {
		g.Name = randomName(t)
	}

	owner := g.Owner

	// default to test user 1's organization
	if owner == "" {
		owner = testUser1.OrganizationID
	}

	mutation := g.client.db.Group.Create().SetName(g.Name).SetOwnerID(owner)

	if len(g.ControlEditorsIDs) > 0 {
		mutation.AddControlEditorIDs(g.ControlEditorsIDs...)
	}

	if len(g.ProgramEditorsIDs) > 0 {
		mutation.AddProgramEditorIDs(g.ProgramEditorsIDs...)
	}

	group := mutation.SaveX(ctx)

	return group
}

// MustDelete is used to cleanup, without authz checks, groups in the database
func (g *GroupCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	err := g.client.db.Group.DeleteOneID(g.ID).Exec(ctx)
	require.NoError(t, err)
}

// MustNew invite builder is used to create, without authz checks, invites in the database
func (i *InviteBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Invite {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

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

// MustDelete is used to cleanup, without authz checks, invites in the database
func (i *InviteCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	i.client.db.Invite.DeleteOneID(i.ID).ExecX(ctx)
}

// MustNew subscriber builder is used to create, without authz checks, subscribers in the database
func (i *SubscriberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Subscriber {
	reqCtx := privacy.DecisionContext(ctx, privacy.Allow)

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
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if pat.Name == "" {
		pat.Name = gofakeit.AppName()
	}

	if pat.Description == "" {
		pat.Description = gofakeit.HipsterSentence(5)
	}

	if pat.OwnerID == "" {
		owner := (&UserBuilder{client: pat.client}).MustNew(ctx, t)
		pat.OwnerID = owner.ID
	}

	if pat.OrganizationIDs == nil {
		org := (&OrganizationBuilder{client: pat.client}).MustNew(ctx, t)
		pat.OrganizationIDs = []string{org.ID}
	}

	request := pat.client.db.PersonalAccessToken.Create().
		SetName(pat.Name).
		SetOwnerID(pat.OwnerID).
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
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

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

	if at.OwnerID != "" {
		request.SetOwnerID(at.OwnerID)
	}

	if at.ExpiresAt != nil {
		request.SetExpiresAt(*at.ExpiresAt)
	}

	token := request.SaveX(ctx)

	return token
}

// MustNew user builder is used to create, without authz checks, group members in the database
func (gm *GroupMemberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.GroupMembership {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if gm.GroupID == "" {
		group := (&GroupBuilder{client: gm.client}).MustNew(ctx, t)
		gm.GroupID = group.ID
	}

	if gm.UserID == "" {
		orgMember := (&OrgMemberBuilder{client: gm.client, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)
		gm.UserID = orgMember.UserID
	}

	groupMember := gm.client.db.GroupMembership.Create().
		SetUserID(gm.UserID).
		SetGroupID(gm.GroupID).
		SaveX(ctx)

	return groupMember
}

// MustDelete is used to cleanup, without authz checks, group members in the database
func (gm *GroupMemberCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	gm.client.db.GroupMembership.DeleteOneID(gm.ID).ExecX(ctx)
}

// MustNew entity type builder is used to create, without authz checks, entity types in the database
func (e *EntityTypeBuilder) MustNew(ctx context.Context, t *testing.T) *ent.EntityType {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if e.Name == "" {
		e.Name = randomName(t)
	}

	entityType := e.client.db.EntityType.Create().
		SetName(e.Name).
		SaveX(ctx)

	return entityType
}

// MustDelete is used to cleanup, without authz checks, entities in the database
func (e *EntityTypeCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	e.client.db.EntityType.DeleteOneID(e.ID).ExecX(ctx)
}

// MustNew entity builder is used to create, without authz checks, entities in the database
func (e *EntityBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Entity {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

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

// MustDelete is used to cleanup, without authz checks, entities in the database
func (e *EntityCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	e.client.db.Entity.DeleteOneID(e.ID).ExecX(ctx)
}

// MustNew contact builder is used to create, without authz checks, contacts in the database
func (c *ContactBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Contact {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

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

// MustDelete is used to cleanup, without authz checks, contacts in the database
func (c *ContactCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	c.client.db.Contact.DeleteOneID(c.ID).ExecX(ctx)
}

// MustNew task builder is used to create, without authz checks, tasks in the database
func (c *TaskBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Task {
	// add client to context, required for hooks that expect the client to be in the context
	ctx = ent.NewContext(ctx, c.client.db)

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if c.Title == "" {
		c.Title = gofakeit.AppName()
	}

	if c.Description == "" {
		c.Description = gofakeit.HipsterSentence(5)
	}

	taskCreate := c.client.db.Task.Create().
		SetTitle(c.Title).
		SetDescription(c.Description)

	if c.Status != "" {
		taskCreate.SetStatus(c.Status)
	}

	if c.AssigneeID != "" {
		taskCreate.SetAssigneeID(c.AssigneeID)
	}

	if !c.Due.IsZero() {
		taskCreate.SetDue(c.Due)
	}

	if c.OrganizationID != "" {
		taskCreate.SetOwnerID(c.OrganizationID)
	}

	if c.GroupID != "" {
		taskCreate.AddGroupIDs(c.GroupID)
	}

	task := taskCreate.SaveX(ctx)

	return task
}

// MustNew program builder is used to create, without authz checks, programs in the database
func (p *ProgramBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Program {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

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
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if pm.ProgramID == "" {
		program := (&ProgramBuilder{client: pm.client}).MustNew(ctx, t)
		pm.ProgramID = program.ID
	}

	if pm.UserID == "" {
		// first create an org member
		orgMember := (&OrgMemberBuilder{client: pm.client, OrgID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)
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
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// add client to context
	ctx = ent.NewContext(ctx, p.client.db)

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
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// add client to context
	ctx = ent.NewContext(ctx, p.client.db)

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
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// add client to context
	ctx = ent.NewContext(ctx, r.client.db)

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
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// add client to context
	ctx = ent.NewContext(ctx, c.client.db)

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
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// add client to context
	ctx = ent.NewContext(ctx, n.client.db)

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
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// add client to context
	ctx = ent.NewContext(ctx, c.client.db)

	if c.Name == "" {
		c.Name = gofakeit.UUID()
	}

	mutation := c.client.db.Control.Create().
		SetRefCode(c.Name)

	if c.ProgramID != "" {
		mutation.AddProgramIDs(c.ProgramID)
	}

	control := mutation.
		SaveX(ctx)

	return control
}

// MustNew subcontrol builder is used to create, without authz checks, subcontrols in the database
func (s *SubcontrolBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Subcontrol {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// add client to context
	ctx = ent.NewContext(ctx, s.client.db)

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
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// add client to context
	ctx = ent.NewContext(ctx, c.client.db)

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
