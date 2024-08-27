package graphapi_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	mock_fga "github.com/datumforge/fgax/mockery"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
)

type OrganizationBuilder struct {
	client *client

	// Fields
	Name        string
	DisplayName string
	Description *string
	OrgID       string
	ParentOrgID string
	PersonalOrg bool
}

type OrganizationCleanup struct {
	client *client

	// Fields
	ID string
}

type GroupBuilder struct {
	client *client

	// Fields
	Name  string
	Owner string
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

type SubscriberCleanup struct {
	client *client

	// Fields
	Email string
}

type FeatureBuilder struct {
	client *client

	// Fields
	Name        string
	Description string
	DisplayName string
}

type FeatureCleanup struct {
	client *client

	// Fields
	ID string
}

type EntitlementBuilder struct {
	client *client

	// Fields
	PlanID         string
	OrganizationID string
}

type EntitlementCleanup struct {
	client *client

	// Fields
	ID string
}

type EntitlementPlanBuilder struct {
	client *client

	// Fields
	Name        string
	Description string
	DisplayName string
	Version     string
}

type EntitlementPlanCleanup struct {
	client *client

	// Fields
	ID string
}

type EntitlementPlanFeatureBuilder struct {
	client *client

	// Fields
	PlanID    string
	FeatureID string
	MetaData  map[string]interface{}
}

type EntitlementPlanFeatureCleanup struct {
	client *client

	// Fields
	ID string
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

// MustNew organization builder is used to create, without authz checks, orgs in the database
func (o *OrganizationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Organization {
	// no auth, so allow policy
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// add client to context
	// the organization hook expects the client to be in the context
	// which happens automatically when using the graph resolvers
	ctx = ent.NewContext(ctx, o.client.db)

	if !o.PersonalOrg {
		// mock writes
		mock_fga.WriteOnce(t, o.client.fga)
	}

	if o.Name == "" {
		o.Name = gofakeit.LetterN(40)
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

	org := m.SaveX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(o.client.fga)

	return org
}

// MustDelete is used to cleanup, without authz checks, orgs in the database
func (o *OrganizationCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	o.client.db.Organization.DeleteOneID(o.ID).ExecX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(o.client.fga)
}

// MustNew user builder is used to create, without authz checks, users in the database
func (u *UserBuilder) MustNew(ctx context.Context, t *testing.T) *ent.User {
	// mock writes
	mock_fga.WriteOnce(t, u.client.fga)

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
	userSetting := u.client.db.UserSetting.Create().SaveX(ctx)

	user := u.client.db.User.Create().
		SetFirstName(u.FirstName).
		SetLastName(u.LastName).
		SetEmail(u.Email).
		SetPassword(u.Password).
		SetSetting(userSetting).
		SaveX(ctx)

	user.Edges.Setting.DefaultOrg(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(u.client.fga)

	return user
}

// MustDelete is used to cleanup, without authz checks, users in the database
func (u *UserCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	u.client.db.User.DeleteOneID(u.ID).ExecX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(u.client.fga)
}

// MustNew tfa settings builder is used to create, without authz checks, tfa settings in the database
func (tf *TFASettingBuilder) MustNew(ctx context.Context, t *testing.T, userID string) *ent.TFASetting {
	return tf.client.db.TFASetting.Create().
		SetTotpAllowed(true).
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

	// mock writes
	mock_fga.WriteOnce(t, om.client.fga)

	orgMembers := om.client.db.OrgMembership.Create().
		SetUserID(om.UserID).
		SetOrganizationID(om.OrgID).
		SetRole(*role).
		SaveX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(om.client.fga)

	return orgMembers
}

// MustDelete is used to cleanup, without authz checks, org members in the database
func (om *OrgMemberCleanup) MustDelete(ctx context.Context, t *testing.T) {
	// mock writes
	mock_fga.WriteOnce(t, om.client.fga)

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	om.client.db.OrgMembership.DeleteOneID(om.ID).ExecX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(om.client.fga)
}

// MustNew group builder is used to create, without authz checks, groups in the database
func (g *GroupBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Group {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if g.Name == "" {
		g.Name = gofakeit.AppName()
	}

	// create owner if not provided
	owner := g.Owner

	if g.Owner == "" {
		owner = testPersonalOrgID
	}

	// mock writes
	mock_fga.WriteAny(t, g.client.fga)

	mock_fga.ListAny(t, g.client.fga, []string{fmt.Sprintf("group:%s", owner)})

	group := g.client.db.Group.Create().SetName(g.Name).SetOwnerID(owner).SaveX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(g.client.fga)

	return group
}

// MustDelete is used to cleanup, without authz checks, groups in the database
func (g *GroupCleanup) MustDelete(ctx context.Context, t *testing.T) {
	mock_fga.ClearMocks(g.client.fga)

	// mock writes
	mock_fga.ReadAny(t, g.client.fga)
	mock_fga.ListAny(t, g.client.fga, []string{fmt.Sprintf("group:%s", g.ID)})

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	g.client.db.Group.DeleteOneID(g.ID).ExecX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(g.client.fga)
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

	// clear mocks before going to tests
	mock_fga.ClearMocks(i.client.fga)

	return invite
}

// MustDelete is used to cleanup, without authz checks, invites in the database
func (i *InviteCleanup) MustDelete(ctx context.Context, t *testing.T) {
	// mock writes
	mock_fga.ReadAny(t, i.client.fga)

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	i.client.db.Invite.DeleteOneID(i.ID).ExecX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(i.client.fga)
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

	// clear mocks before going to tests
	mock_fga.ClearMocks(pat.client.fga)

	return token
}

// MustNew api tokens builder is used to create, without authz checks, api tokens in the database
func (at *APITokenBuilder) MustNew(ctx context.Context, t *testing.T) *ent.APIToken {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// mock writes
	if len(at.Scopes) > 0 {
		mock_fga.WriteOnce(t, at.client.fga)
	}

	if at.Name == "" {
		at.Name = gofakeit.AppName()
	}

	if at.Description == "" {
		at.Description = gofakeit.HipsterSentence(5)
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

	// clear mocks before going to tests
	mock_fga.ClearMocks(at.client.fga)

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
		user := (&UserBuilder{client: gm.client}).MustNew(ctx, t)
		gm.UserID = user.ID
	}

	// mock writes
	mock_fga.ListAny(t, gm.client.fga, []string{fmt.Sprintf("organization:%s", testPersonalOrgID)})
	mock_fga.WriteOnce(t, gm.client.fga)

	groupMember := gm.client.db.GroupMembership.Create().
		SetUserID(gm.UserID).
		SetGroupID(gm.GroupID).
		SaveX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(gm.client.fga)

	return groupMember
}

// MustDelete is used to cleanup, without authz checks, group members in the database
func (gm *GroupMemberCleanup) MustDelete(ctx context.Context, t *testing.T) {
	// mock writes
	mock_fga.WriteOnce(t, gm.client.fga)

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	gm.client.db.GroupMembership.DeleteOneID(gm.ID).ExecX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(gm.client.fga)
}

// MustNew feature builder is used to create, without authz checks, features in the database
func (f *FeatureBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Feature {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if f.Name == "" {
		f.Name = gofakeit.AppName()
	}

	if f.Description == "" {
		f.Description = gofakeit.HipsterSentence(5)
	}

	feature := f.client.db.Feature.Create().
		SetName(f.Name).
		SetDescription(f.Description).
		SetDisplayName(f.DisplayName).
		SaveX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(f.client.fga)

	return feature
}

// MustNew plan builder is used to create, without authz checks, plans in the database
func (p *EntitlementPlanBuilder) MustNew(ctx context.Context, t *testing.T) *ent.EntitlementPlan {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if p.Name == "" {
		p.Name = gofakeit.AppName()
	}

	if p.Description == "" {
		p.Description = gofakeit.HipsterSentence(5)
	}

	if p.Version == "" {
		p.Version = fmt.Sprintf("v%d", gofakeit.Number(1, 10))
	}

	plan := p.client.db.EntitlementPlan.Create().
		SetName(p.Name).
		SetVersion(p.Version).
		SetDescription(p.Description).
		SaveX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(p.client.fga)

	return plan
}

// MustNew entitlement builder is used to create, without authz checks, entitlements in the database
func (e *EntitlementBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Entitlement {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if e.PlanID == "" {
		plan := (&EntitlementPlanBuilder{client: e.client}).MustNew(ctx, t)
		e.PlanID = plan.ID
	}

	if e.OrganizationID == "" {
		org := (&OrganizationBuilder{client: e.client}).MustNew(ctx, t)
		e.OrganizationID = org.ID
	}

	entitlement := e.client.db.Entitlement.Create().
		SetPlanID(e.PlanID).
		SetOrganizationID(e.OrganizationID).
		SaveX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(e.client.fga)

	return entitlement
}

// MustNew entitlement plan feature builder is used to create, without authz checks, plan features in the database
func (e *EntitlementPlanFeatureBuilder) MustNew(ctx context.Context, t *testing.T) *ent.EntitlementPlanFeature {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if e.PlanID == "" {
		plan := (&EntitlementPlanBuilder{client: e.client}).MustNew(ctx, t)
		e.PlanID = plan.ID
	}

	if e.FeatureID == "" {
		feature := (&FeatureBuilder{client: e.client}).MustNew(ctx, t)
		e.FeatureID = feature.ID
	}

	if e.MetaData == nil {
		e.MetaData = map[string]interface{}{
			"limit_type": "days",
			"limit":      30,
		}
	}

	planFeature := e.client.db.EntitlementPlanFeature.Create().
		SetPlanID(e.PlanID).
		SetFeatureID(e.FeatureID).
		SetMetadata(e.MetaData).
		SaveX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(e.client.fga)

	return planFeature
}

// MustNew entity type builder is used to create, without authz checks, entity types in the database
func (e *EntityTypeBuilder) MustNew(ctx context.Context, t *testing.T) *ent.EntityType {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if e.Name == "" {
		e.Name = gofakeit.AppName()
	}

	entityType := e.client.db.EntityType.Create().
		SetName(e.Name).
		SaveX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(e.client.fga)

	return entityType
}

// MustDelete is used to cleanup, without authz checks, entities in the database
func (e *EntityTypeCleanup) MustDelete(ctx context.Context, t *testing.T) {
	mock_fga.ClearMocks(e.client.fga)

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	e.client.db.EntityType.DeleteOneID(e.ID).ExecX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(e.client.fga)
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

	// clear mocks before going to tests
	mock_fga.ClearMocks(e.client.fga)

	return entity
}

// MustDelete is used to cleanup, without authz checks, entities in the database
func (e *EntityCleanup) MustDelete(ctx context.Context, t *testing.T) {
	mock_fga.ClearMocks(e.client.fga)

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	e.client.db.Entity.DeleteOneID(e.ID).ExecX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(e.client.fga)
}

// MustNew contact builder is used to create, without authz checks, contacts in the database
func (e *ContactBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Contact {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if e.Name == "" {
		e.Name = gofakeit.AppName()
	}

	if e.Email == "" {
		e.Email = gofakeit.Email()
	}

	if e.Phone == "" {
		e.Phone = gofakeit.Phone()
	}

	if e.Address == "" {
		address := gofakeit.Address()
		e.Address = fmt.Sprintf("%s, %s, %s, %s", address.Street, address.City, address.State, address.Zip)
	}

	if e.Title == "" {
		e.Title = gofakeit.JobTitle()
	}

	if e.Company == "" {
		e.Company = gofakeit.Company()
	}

	entity := e.client.db.Contact.Create().
		SetFullName(e.Name).
		SetEmail(e.Email).
		SetPhoneNumber(e.Phone).
		SetAddress(e.Address).
		SetTitle(e.Title).
		SetCompany(e.Company).
		SaveX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(e.client.fga)

	return entity
}

// MustDelete is used to cleanup, without authz checks, contacts in the database
func (e *ContactCleanup) MustDelete(ctx context.Context, t *testing.T) {
	mock_fga.ClearMocks(e.client.fga)

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	e.client.db.Contact.DeleteOneID(e.ID).ExecX(ctx)

	// clear mocks before going to tests
	mock_fga.ClearMocks(e.client.fga)
}
