package graphapi_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/evidence"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/programmembership"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/ulids"
)

type OrganizationBuilder struct {
	client   *client
	Features []models.OrgModule

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

type JobRunnerBuilder struct {
	client *client
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
	RiskID     string
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
	Status          enums.ProgramStatus
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
	Name                    string
	BlockedGroupIDs         []string
	EditorGroupIDs          []string
	SkipApprovalRequirement bool
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
	RefCode                 string
	Aliases                 []string
	Title                   string
	ProgramID               string
	StandardID              string
	ControlOwnerID          string
	ControlEditorGroupID    string
	ControlImplementationID string
	// AllFields will set all direct fields on the control with random data
	AllFields   bool
	Category    string
	Subcategory string
}

type SubcontrolBuilder struct {
	client *client

	// Fields
	Name        string
	ControlID   string
	Category    string
	Subcategory string
}

type MappedControlBuilder struct {
	client *client

	// Fields
	FromControlIDs    []string
	ToControlIDs      []string
	FromSubcontrolIDs []string
	ToSubcontrolIDs   []string
	MappingType       enums.MappingType
	Relation          string
	Confidence        int
	Source            enums.MappingSource
	InternalID        string
	InternalNotes     string
}

type EvidenceBuilder struct {
	client *client

	// Fields
	Name        string
	ProgramID   string
	ControlID   string
	IncludeFile bool
}

type StandardBuilder struct {
	client *client

	// Fields
	Name      string
	Framework string
	IsPublic  bool
}

type SubprocessorBuilder struct {
	client *client

	// Fields
	Name          string
	Description   string
	LogoRemoteURL string
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
	Name   string
	ZoneID string
}

type FileBuilder struct {
	client *client

	// Fields
	Name string
}

type TemplateBuilder struct {
	client *client

	// Fields
	Name         string
	Description  string
	Kind         enums.TemplateKind
	TemplateType enums.DocumentType
	JSONConfig   map[string]any
	UISchema     map[string]any
}

type AssessmentBuilder struct {
	client *client

	// Fields
	Name                string
	AssessmentType      enums.AssessmentType
	TemplateID          string
	ResponseDueDuration int64
	Tags                []string
}

type AssessmentResponseBuilder struct {
	client *client

	// Fields
	AssessmentID   string
	Email          string
	OwnerID        string
	DueDate        *time.Time
	DocumentDataID string
}

type TagDefinitionBuilder struct {
	client *client

	// Fields
	Name  string
	Color string
}

type CustomTypeEnumBuilder struct {
	client *client

	// Fields
	Name        string
	Description string
	Color       string
	ObjectType  string
}

// Faker structs with random injected data
type Faker struct {
	Name string
}

func randomName(t *testing.T) string {
	var f Faker
	err := gofakeit.Struct(&f)
	requireNoError(err)

	var b strings.Builder
	for _, r := range f.Name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String() + "_" + ulids.New().String()
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
//		MustDelete(testUser1.UserCtx, t)
//
// Special handling for standards - update them to be private before deletion
// this is to allow the system admin to delete public standards
// and controls that are linked to them
// this is a workaround to avoid the cascade delete hook on standard
// that would otherwise prevent the deletion of public standards
// and controls that are linked to them
func (c *Cleanup[DeleteExec]) MustDelete(ctx context.Context, t *testing.T) {
	// add client to context for hooks that expect the client to be in the context
	ctx = setContext(ctx, suite.client.db)

	// Special handling for standards - update them to be private before deletion
	// Only do this for system admins
	if _, ok := any(c.client).(*ent.StandardClient); ok && auth.IsSystemAdminFromContext(ctx) {
		if c.ID != "" {
			err := suite.client.db.Standard.UpdateOneID(c.ID).SetIsPublic(false).Exec(ctx)
			requireNoError(err)
		}
		for _, id := range c.IDs {
			err := suite.client.db.Standard.UpdateOneID(id).SetIsPublic(false).Exec(ctx)
			requireNoError(err)
		}
	}

	for _, id := range c.IDs {
		err := c.client.DeleteOneID(id).Exec(ctx)
		requireNoError(err)
	}

	if c.ID != "" {
		err := c.client.DeleteOneID(c.ID).Exec(ctx)
		requireNoError(err)
	}
}

// setContext is a helper function to set the context for the client
// setting privacy to allow and adding the client to the context
func setContext(ctx context.Context, db *ent.Client) context.Context {
	ctx = ent.NewContext(rule.WithInternalContext(ctx), db)

	// add the GraphQL response context to prevent panics from interceptors that call graphql.AddError
	return graphql.WithResponseContext(ctx, gqlerrors.ErrorPresenter, graphql.DefaultRecover)
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
		desc := gofakeit.HipsterSentence()
		o.Description = &desc
	}

	m := o.client.db.Organization.Create().SetName(o.Name).SetDescription(*o.Description).SetDisplayName(o.DisplayName).SetPersonalOrg(o.PersonalOrg)

	if o.ParentOrgID != "" {
		m.SetParentID(o.ParentOrgID)
	}

	org, err := m.Save(ctx)
	requireNoError(err)

	if o.AllowedDomains != nil {
		orgSetting, err := org.Setting(ctx)
		requireNoError(err)

		err = orgSetting.Update().SetAllowedEmailDomains(o.AllowedDomains).Exec(ctx)
		requireNoError(err)
	}

	o.enableModules(ctx, t, org.ID)

	return org
}

// enableModules enables the selected organization modules for the given organization
func (o *OrganizationBuilder) enableModules(ctx context.Context, t *testing.T, orgID string) {
	features := o.Features

	if len(o.Features) == 0 {
		features = models.AllOrgModules
	}

	err := entitlements.CreateFeatureTuples(ctx, o.client.fga, orgID, features)
	assert.NilError(t, err)

	for _, feature := range features {
		n, err := o.client.db.OrgModule.Update().
			Where(
				orgmodule.OwnerID(orgID),
				orgmodule.Module(feature),
			).
			SetActive(true).
			Save(ctx)
		assert.NilError(t, err)

		// if no rows were updated, the module wasn't created - create it now
		if n == 0 {
			err = o.client.db.OrgModule.Create().
				SetOwnerID(orgID).
				SetModule(feature).
				SetActive(true).
				SetPrice(models.Price{Amount: 0, Interval: "month"}).
				Exec(ctx)
			assert.NilError(t, err)
		}
	}

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
		// ensure email has a valid domain for email verification tests
		u.Email = strings.ToLower(fmt.Sprintf("%s.%s.%s@%s", u.FirstName, u.LastName, ulids.New().String(), "theopenlane.io"))
	}

	if u.Password == "" {
		u.Password = gofakeit.Password(true, true, true, true, false, 20)
	}

	// create user setting
	userSetting, err := u.client.db.UserSetting.Create().Save(ctx)
	requireNoError(err)

	user, err := u.client.db.User.Create().
		SetFirstName(u.FirstName).
		SetLastName(u.LastName).
		SetEmail(u.Email).
		SetPassword(u.Password).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SetSetting(userSetting).
		Save(ctx)
	requireNoError(err)

	_, err = user.Edges.Setting.DefaultOrg(ctx)
	requireNoError(err)

	return user
}

// MustNew tfa settings builder is used to create, without authz checks, tfa settings in the database
func (tf *TFASettingBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TFASetting {
	if tf.totpAllowed == nil {
		tf.totpAllowed = lo.ToPtr(true)
	}

	setting, err := tf.client.db.TFASetting.Create().
		SetTotpAllowed(*tf.totpAllowed).
		Save(ctx)

	// if the setting is not created, it means the user already has a setting
	// and let's skip for seeding
	if errors.Is(err, generated.ConstraintError{}) {
		return nil
	}

	return setting
}

// MustNew JobRunner settings builder is used to create runners
func (w *JobRunnerBuilder) MustNew(ctx context.Context, t *testing.T) *ent.JobRunner {
	ctx = setContext(ctx, w.client.db)

	wn, err := w.client.db.JobRunner.Create().
		SetName(randomName(t)).
		SetIPAddress(gofakeit.IPv4Address()).
		Save(ctx)
	requireNoError(err)

	return wn
}

// MustNew webauthn settings builder is used to create passkeys without the browser setup process
func (w *WebauthnBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Webauthn {
	uuidBytes, err := uuid.NewUUID()
	requireNoError(err)

	wn, err := w.client.db.Webauthn.Create().
		SetAaguid(models.ToAAGUID(uuidBytes[:])).
		SetAttestationType("type").
		SetBackupEligible(true).
		SetBackupState(true).
		SetSignCount(10).
		SetCredentialID([]byte(uuid.NewString())).
		SetTransports([]string{uuid.NewString()}).
		Save(ctx)
	requireNoError(err)

	return wn
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

	orgMember, err := om.client.db.OrgMembership.Create().
		SetUserID(om.UserID).
		SetRole(*role).
		Save(ctx)
	requireNoError(err)

	return orgMember
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
	requireNoError(err)

	return group
}

// MustNew invite builder is used to create, without authz checks, invites in the database
func (i *InviteBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Invite {
	ctx = setContext(ctx, i.client.db)

	// create user if not provided
	rec := i.Recipient

	if rec == "" {
		rec = strings.ToLower(fmt.Sprintf("%s@%s", ulids.New().String(), "theopenlane.io"))
	}

	inviteQuery := i.client.db.Invite.Create().
		SetRecipient(rec)

	if i.Role != "" {
		inviteQuery.SetRole(*enums.ToRole(i.Role))
	}

	invite, err := inviteQuery.Save(ctx)
	requireNoError(err)

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

	sub, err := i.client.db.Subscriber.Create().
		SetEmail(rec).
		SetActive(true).Save(reqCtx)
	requireNoError(err)

	return sub
}

// MustNew personal access tokens builder is used to create, without authz checks, personal access tokens in the database
func (pat *PersonalAccessTokenBuilder) MustNew(ctx context.Context, t *testing.T) *ent.PersonalAccessToken {
	ctx = setContext(ctx, pat.client.db)

	if pat.Name == "" {
		pat.Name = gofakeit.AppName()
	}

	if pat.Description == "" {
		pat.Description = gofakeit.HipsterSentence()
	}

	if pat.OrganizationIDs == nil {
		// default to adding the test users organization ID
		pat.OrganizationIDs = []string{testUser1.OrganizationID}
	}

	request := pat.client.db.PersonalAccessToken.Create().
		SetName(pat.Name).
		SetDescription(pat.Description).
		AddOrganizationIDs(pat.OrganizationIDs...)

	if pat.ExpiresAt != nil {
		request.SetExpiresAt(*pat.ExpiresAt)
	}

	token, err := request.Save(ctx)
	requireNoError(err)

	return token
}

// MustNew api tokens builder is used to create, without authz checks, api tokens in the database
func (at *APITokenBuilder) MustNew(ctx context.Context, t *testing.T) *ent.APIToken {
	ctx = setContext(ctx, at.client.db)

	if at.Name == "" {
		at.Name = gofakeit.AppName()
	}

	if at.Description == "" {
		at.Description = gofakeit.HipsterSentence()
	}

	if at.Scopes == nil {
		at.Scopes = []string{"read", "write", "group_manager"}
	}

	request := at.client.db.APIToken.Create().
		SetName(at.Name).
		SetDescription(at.Description).
		SetScopes(at.Scopes)

	if at.ExpiresAt != nil {
		request.SetExpiresAt(*at.ExpiresAt)
	}

	token, err := request.Save(ctx)
	requireNoError(err)

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
		orgMember := (&OrgMemberBuilder{client: gm.client}).MustNew(ctx, t)
		gm.UserID = orgMember.UserID
	}

	mut := gm.client.db.GroupMembership.Create().
		SetUserID(gm.UserID).
		SetGroupID(gm.GroupID)

	if gm.Role != "" {
		mut.SetRole(*enums.ToRole(gm.Role))
	}

	groupMember, err := mut.Save(ctx)
	requireNoError(err)

	gmToReturn, err := gm.client.db.GroupMembership.Query().
		WithUser().
		WithOrgMembership().
		Where(groupmembership.ID(groupMember.ID)).Only(ctx)
	requireNoError(err)

	return gmToReturn
}

// MustNew entity type builder is used to create, without authz checks, entity types in the database
func (e *EntityTypeBuilder) MustNew(ctx context.Context, t *testing.T) *ent.EntityType {
	ctx = setContext(ctx, e.client.db)

	if e.Name == "" {
		e.Name = randomName(t)
	}

	entityType, err := e.client.db.EntityType.Create().
		SetName(e.Name).
		Save(ctx)
	requireNoError(err)

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
		e.Description = gofakeit.HipsterSentence()
	}

	if e.TypeID == "" {
		et := (&EntityTypeBuilder{client: e.client}).MustNew(ctx, t)
		e.TypeID = et.ID
	}

	entity, err := e.client.db.Entity.Create().
		SetName(e.Name).
		SetDisplayName(e.DisplayName).
		SetEntityTypeID(e.TypeID).
		SetDescription(e.Description).
		Save(ctx)
	requireNoError(err)

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

	entity, err := c.client.db.Contact.Create().
		SetFullName(c.Name).
		SetEmail(c.Email).
		SetPhoneNumber(c.Phone).
		SetAddress(c.Address).
		SetTitle(c.Title).
		SetCompany(c.Company).
		Save(ctx)
	requireNoError(err)

	return entity
}

// MustNew task builder is used to create, without authz checks, tasks in the database
func (c *TaskBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Task {
	ctx = setContext(ctx, c.client.db)

	if c.Title == "" {
		c.Title = gofakeit.AppName()
	}

	if c.Details == "" {
		c.Details = gofakeit.HipsterSentence()
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

	if c.RiskID != "" {
		taskCreate.AddRiskIDs(c.RiskID)
	}

	task, err := taskCreate.Save(ctx)
	requireNoError(err)

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

	if p.Status.String() != "" {
		mutation.SetStatus(p.Status)
	}

	program, err := mutation.
		Save(ctx)
	requireNoError(err)

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
		orgMember := (&OrgMemberBuilder{client: pm.client}).MustNew(ctx, t)
		pm.UserID = orgMember.UserID
	}

	mutation := pm.client.db.ProgramMembership.Create().
		SetUserID(pm.UserID).
		SetProgramID(pm.ProgramID)

	if pm.Role != "" {
		mutation.SetRole(*enums.ToRole(pm.Role))
	}

	programMember, err := mutation.Save(ctx)
	requireNoError(err)

	programMember, err = pm.client.db.ProgramMembership.Query().
		WithUser().
		WithOrgMembership().
		Where(programmembership.ID(programMember.ID)).Only(ctx)
	requireNoError(err)

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

	procedure, err := mutation.Save(ctx)
	requireNoError(err)

	return procedure
}

// MustNew policy builder is used to create, without authz checks, policies in the database
func (p *InternalPolicyBuilder) MustNew(ctx context.Context, t *testing.T) *ent.InternalPolicy {
	ctx = setContext(ctx, p.client.db)

	if p.Name == "" {
		p.Name = gofakeit.AppName()
	}

	mut := p.client.db.InternalPolicy.Create().
		SetName(p.Name)

	if len(p.BlockedGroupIDs) > 0 {
		mut.AddBlockedGroupIDs(p.BlockedGroupIDs...)
	}

	if len(p.EditorGroupIDs) > 0 {
		mut.AddEditorIDs(p.EditorGroupIDs...)
	}

	if p.SkipApprovalRequirement {
		mut.SetApprovalRequired(false)
	}

	policy, err := mut.Save(ctx)
	requireNoError(err)

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

	risk, err := mutation.Save(ctx)
	requireNoError(err)

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

	co, err := mutation.Save(ctx)
	requireNoError(err)

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

	narrative, err := mutation.
		Save(ctx)
	requireNoError(err)

	return narrative
}

// MustNew control builder is used to create, without authz checks, controls in the database
func (c *ControlBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Control {
	ctx = setContext(ctx, c.client.db)

	if c.RefCode == "" {
		c.RefCode = gofakeit.UUID()
	}

	if c.Title == "" {
		c.Title = gofakeit.HipsterSentence()
	}

	mutation := c.client.db.Control.Create().
		SetRefCode(c.RefCode).SetTitle(c.Title)

	if c.ProgramID != "" {
		mutation.AddProgramIDs(c.ProgramID)
	}

	if c.StandardID != "" {
		mutation.SetStandardID(c.StandardID)
		mutation.SetSource(enums.ControlSourceFramework)
	} else {
		mutation.SetSource(enums.ControlSourceUserDefined)
	}

	if c.ControlOwnerID != "" {
		mutation.SetControlOwnerID(c.ControlOwnerID)
	}

	if c.ControlEditorGroupID != "" {
		mutation.AddEditorIDs(c.ControlEditorGroupID)
	}

	if c.ControlImplementationID != "" {
		mutation.AddControlImplementationIDs(c.ControlImplementationID)
	}

	if c.AllFields {
		mutation.SetDescription(gofakeit.HipsterSentence()).
			// add a unique string to ensure we know the number of controls created per category is singular
			// this field doesn't actually need to be unique, but is an easy way to do the tests
			SetCategory(gofakeit.Adjective() + ulids.New().String()).
			SetCategoryID(ulids.New().String()).
			SetSubcategory(gofakeit.Adjective() + ulids.New().String()).
			SetControlType(enums.ControlTypeDetective).
			SetExampleEvidence([]models.ExampleEvidence{
				{
					DocumentationType: "Documentation",
					Description:       gofakeit.HipsterSentence(),
				},
			}).
			SetImplementationGuidance([]models.ImplementationGuidance{
				{
					ReferenceID: ulids.New().String(),
					Guidance: []string{
						gofakeit.HipsterSentence(),
						gofakeit.HipsterSentence(),
					},
				},
			}).
			SetAssessmentMethods([]models.AssessmentMethod{
				{
					ID:     ulids.New().String(),
					Type:   "test",
					Method: gofakeit.HipsterSentence(),
				},
			}).
			SetMappedCategories([]string{"Governance", "Risk Management"}).
			SetTags([]string{"tag1", "tag2"}).
			SetReferences([]models.Reference{
				{
					Name: gofakeit.HipsterSentence(),
					URL:  gofakeit.URL(),
				},
			}).
			SetAliases([]string{gofakeit.UUID(), gofakeit.UUID()})
	}

	if c.Category != "" {
		mutation.SetCategory(c.Category)
	}

	if c.Subcategory != "" {
		mutation.SetSubcategory(c.Subcategory)
	}

	if c.Aliases != nil {
		mutation.SetAliases(c.Aliases)
	}

	control, err := mutation.
		Save(ctx)
	requireNoError(err)

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

	if s.Category != "" {
		mutation.SetCategory(s.Category)
	}

	if s.Subcategory != "" {
		mutation.SetSubcategory(s.Subcategory)
	}

	sc, err := mutation.
		Save(ctx)

	requireNoError(err)

	return sc
}

// MustNew control builder is used to create, without authz checks, controls in the database
func (e *EvidenceBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Evidence {
	ctx = setContext(ctx, e.client.db)

	if e.Name == "" {
		e.Name = gofakeit.AppName()
	}

	mutation := e.client.db.Evidence.Create().
		SetCreationDate(time.Now().Add(-time.Minute)).
		SetName(e.Name)

	if e.ProgramID != "" {
		mutation.AddProgramIDs(e.ProgramID)
	}

	if e.ControlID != "" {
		mutation.AddControlIDs(e.ControlID)
	}

	if e.IncludeFile {
		file := (&FileBuilder{client: e.client, Name: e.Name}).MustNew(ctx, t)

		mutation.AddFileIDs(file.ID)
	}

	ev, err := mutation.
		Save(ctx)
	requireNoError(err)

	if e.IncludeFile {
		ev, err := e.client.db.Evidence.Query().WithFiles().Where(evidence.ID(ev.ID)).Only(ctx)
		requireNoError(err)

		return ev
	}

	return ev
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

	mut := s.client.db.Standard.Create().
		SetName(s.Name).
		SetFramework(s.Framework).
		SetIsPublic(s.IsPublic)

	standard, err := mut.Save(ctx)
	requireNoError(err)

	return standard
}

// MustNew subprocessor builder is used to create, without authz checks, subprocessors in the database
func (s *SubprocessorBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Subprocessor {
	ctx = setContext(ctx, s.client.db)

	if s.Name == "" {
		for {
			s.Name = gofakeit.Company()
			_, err := s.client.db.Subprocessor.Query().Where(subprocessor.Name(s.Name)).Only(ctx)
			if err != nil {
				break
			}
		}
	}

	mutation := s.client.db.Subprocessor.Create().
		SetName(s.Name)

	if s.Description != "" {
		mutation.SetDescription(s.Description)
	}

	if s.LogoRemoteURL != "" {
		mutation.SetLogoRemoteURL(s.LogoRemoteURL)
	}

	subprocessor, err := mutation.Save(ctx)
	requireNoError(err)

	return subprocessor
}

// MustNew note builder is used to create, without authz checks, notes in the database
func (n *NoteBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Note {
	ctx = setContext(ctx, n.client.db)

	if n.Text == "" {
		n.Text = gofakeit.HipsterSentence()
	}

	mutation := n.client.db.Note.Create().
		SetText(n.Text)

	if n.TaskID != "" {
		mutation.SetTaskID(n.TaskID)
	}

	if len(n.FileIDs) > 0 {
		mutation.AddFileIDs(n.FileIDs...)
	}

	note, err := mutation.Save(ctx)
	requireNoError(err)

	return note
}

// MustNew controlImplementation builder is used to create, without authz checks, controlImplementations in the database
func (e *ControlImplementationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.ControlImplementation {
	ctx = setContext(ctx, e.client.db)

	if e.Details == "" {
		e.Details = gofakeit.Paragraph()
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
	requireNoError(err)

	return controlImplementation
}

// MustNew controlImplementation builder is used to create, without authz checks, controlImplementations in the database
func (e *MappedControlBuilder) MustNew(ctx context.Context, t *testing.T) *ent.MappedControl {
	if ctx == systemAdminUser.UserCtx {
		if e.InternalID == "" {
			e.InternalID = ulids.New().String()
		}

		if e.InternalNotes == "" {
			e.InternalNotes = "Created by system admin user"
		}
	}

	ctx = setContext(ctx, e.client.db)

	if len(e.FromControlIDs) == 0 && len(e.FromSubcontrolIDs) == 0 {
		fromControl := (&ControlBuilder{client: e.client}).MustNew(ctx, t)
		e.FromControlIDs = []string{fromControl.ID}
	}

	if len(e.ToControlIDs) == 0 && len(e.ToSubcontrolIDs) == 0 {
		toControl := (&ControlBuilder{client: e.client}).MustNew(ctx, t)
		e.ToControlIDs = []string{toControl.ID}
	}

	mutation := e.client.db.MappedControl.Create().
		AddFromControlIDs(e.FromControlIDs...).
		AddToControlIDs(e.ToControlIDs...)

	if len(e.FromSubcontrolIDs) > 0 {
		mutation.AddFromSubcontrolIDs(e.FromSubcontrolIDs...)
	}

	if len(e.ToSubcontrolIDs) > 0 {
		mutation.AddToSubcontrolIDs(e.ToSubcontrolIDs...)
	}

	if e.MappingType != "" {
		mutation.SetMappingType(e.MappingType)
	}

	if e.Relation != "" {
		mutation.SetRelation(e.Relation)
	}

	if e.Confidence != 0 {
		mutation.SetConfidence(e.Confidence)
	}

	if e.Source != "" {
		mutation.SetSource(e.Source)
	}

	if e.InternalID != "" {
		mutation.SetSystemInternalID(e.InternalID)
	}

	if e.InternalNotes != "" {
		mutation.SetInternalNotes(e.InternalNotes)
	}

	mappedControl, err := mutation.Save(ctx)
	requireNoError(err)

	res, err := e.client.db.MappedControl.Query().
		WithFromControls().
		WithFromSubcontrols().
		WithToControls().
		WithToSubcontrols().
		Where(mappedcontrol.ID(mappedControl.ID)).Only(ctx)

	return res
}

// MustNew mappable domain builder is used to create, without authz checks, mappable domains in the database
func (e *MappableDomainBuilder) MustNew(ctx context.Context, t *testing.T) *ent.MappableDomain {
	ctx = setContext(ctx, e.client.db)

	if e.Name == "" {
		e.Name = gofakeit.DomainName()
	}
	if e.ZoneID == "" {
		e.ZoneID = gofakeit.UUID()
	}

	mappableDomain, err := e.client.db.MappableDomain.Create().
		SetName(e.Name).
		SetZoneID(e.ZoneID).
		Save(ctx)
	requireNoError(err)

	return mappableDomain
}

// CustomDomainBuilder is used to create custom domains
type CustomDomainBuilder struct {
	client *client

	// Fields
	CnameRecord      string
	MappableDomainID string
}

// DNSVerificationBuilder is used to create DNS verifications
type DNSVerificationBuilder struct {
	client *client

	// Fields
	CloudflareHostnameID        string
	DNSTxtRecord                string
	DNSTxtValue                 string
	DNSVerificationStatus       *enums.DNSVerificationStatus
	DNSVerificationStatusReason *string
	AcmeChallengePath           string
	ExpectedAcmeChallengeValue  string
	AcmeChallengeStatus         *enums.SSLVerificationStatus
	AcmeChallengeStatusReason   *string
	CustomDomainIDs             []string
}

// MustNew custom domain builder is used to create, without authz checks, custom domains in the database
func (c *CustomDomainBuilder) MustNew(ctx context.Context, t *testing.T) *ent.CustomDomain {
	ctx = setContext(ctx, c.client.db)

	if c.CnameRecord == "" {
		c.CnameRecord = gofakeit.DomainName()
	}

	if c.MappableDomainID == "" {
		mappableDomain := (&MappableDomainBuilder{client: c.client}).MustNew(ctx, t)
		c.MappableDomainID = mappableDomain.ID
	}

	customDomain, err := c.client.db.CustomDomain.Create().
		SetCnameRecord(c.CnameRecord).
		SetMappableDomainID(c.MappableDomainID).
		Save(ctx)
	requireNoError(err)

	return customDomain
}

type JobRunnerTokenBuilder struct {
	client *client

	// Fields
	JobRunnerID string
	ExpiresAt   *time.Time
	IsActive    *bool
}

func (j *JobRunnerTokenBuilder) MustNew(ctx context.Context, t *testing.T) *generated.JobRunnerToken {
	ctx = setContext(ctx, j.client.db)

	create := j.client.db.JobRunnerToken.Create()

	if j.JobRunnerID == "" {
		jobRunner := (&JobRunnerBuilder{client: j.client}).MustNew(ctx, t)
		create.AddJobRunnerIDs(jobRunner.ID)
	}

	if j.ExpiresAt != nil {
		create.SetExpiresAt(*j.ExpiresAt)
	}

	if j.IsActive != nil {
		create.SetIsActive(*j.IsActive)
	}

	token, err := create.Save(ctx)
	requireNoError(err)

	return token
}

type JobRunnerRegistrationTokenBuilder struct {
	client *client

	WithRunner bool
}

func (j *JobRunnerRegistrationTokenBuilder) MustNew(ctx context.Context, t *testing.T) *generated.JobRunnerRegistrationToken {
	ctx = setContext(ctx, j.client.db)

	create := j.client.db.JobRunnerRegistrationToken.Create()
	if j.WithRunner {
		jobRunner := (&JobRunnerBuilder{client: j.client}).MustNew(ctx, t)
		create.SetJobRunnerID(jobRunner.ID)
	}

	token, err := create.Save(ctx)
	requireNoError(err)

	return token
}

// MustNew DNS verification builder is used to create, without authz checks, DNS verifications in the database
func (d *DNSVerificationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.DNSVerification {
	ctx = setContext(ctx, d.client.db)

	if d.CloudflareHostnameID == "" {
		d.CloudflareHostnameID = gofakeit.UUID()
	}

	if d.DNSTxtRecord == "" {
		d.DNSTxtRecord = "_cf-verify." + gofakeit.DomainName()
	}

	if d.DNSTxtValue == "" {
		d.DNSTxtValue = gofakeit.UUID()
	}

	mutation := d.client.db.DNSVerification.Create().
		SetCloudflareHostnameID(d.CloudflareHostnameID).
		SetDNSTxtRecord(d.DNSTxtRecord).
		SetDNSTxtValue(d.DNSTxtValue).
		SetAcmeChallengePath(d.AcmeChallengePath).
		SetExpectedAcmeChallengeValue(d.ExpectedAcmeChallengeValue)

	if d.DNSVerificationStatus != nil {
		mutation.SetDNSVerificationStatus(*d.DNSVerificationStatus)
	}

	if d.DNSVerificationStatusReason != nil {
		mutation.SetDNSVerificationStatusReason(*d.DNSVerificationStatusReason)
	}

	if d.AcmeChallengeStatus != nil {
		mutation.SetAcmeChallengeStatus(*d.AcmeChallengeStatus)
	}

	if d.AcmeChallengeStatusReason != nil {
		mutation.SetAcmeChallengeStatusReason(*d.AcmeChallengeStatusReason)
	}

	if len(d.CustomDomainIDs) > 0 {
		mutation.AddCustomDomainIDs(d.CustomDomainIDs...)
	}

	dnsVerification, err := mutation.Save(ctx)
	requireNoError(err)

	return dnsVerification
}

type JobTemplateBuilder struct {
	client *client

	Title        string
	Description  *string
	Cron         string
	Platform     enums.JobPlatformType
	DownloadURL  string
	WindmillPath string
}

const testScriptURL = "https://raw.githubusercontent.com/theopenlane/jobs-examples/refs/heads/main/basic/print.go"

func (j *JobTemplateBuilder) MustNew(ctx context.Context, t *testing.T) *ent.JobTemplate {
	ctx = setContext(ctx, j.client.db)

	if j.Title == "" {
		j.Title = "Test Job Template"
	}

	if j.DownloadURL == "" {
		j.DownloadURL = testScriptURL
	}

	if j.Platform == "" {
		j.Platform = enums.JobPlatformTypeGo
	}

	if j.WindmillPath == "" {
		j.WindmillPath = "u/admin/gifted_script"
	}

	mut := j.client.db.JobTemplate.Create().
		SetTitle(j.Title).
		SetDownloadURL(j.DownloadURL).
		SetPlatform(j.Platform).
		SetWindmillPath(j.WindmillPath)

	if j.Description != nil {
		mut.SetDescription(*j.Description)
	}

	if j.Cron != "" {
		mut.SetCron(models.Cron(j.Cron))
	}

	if j.Description != nil {
		mut.SetDescription(*j.Description)
	}

	jt, err := mut.Save(ctx)
	requireNoError(err)

	return jt
}

type ScheduledJobBuilder struct {
	client *client

	// Fields
	JobID         string
	Configuration models.JobConfiguration
	Cron          *string
	JobRunnerID   string
	ControlIDs    []string
	Active        bool
}

func (b *ScheduledJobBuilder) MustNew(ctx context.Context, t *testing.T) *generated.ScheduledJob {
	ctx = setContext(ctx, b.client.db)

	job := b.client.db.ScheduledJob.Create().
		SetJobID(b.JobID)

	if b.JobRunnerID != "" {
		job.SetJobRunnerID(b.JobRunnerID)
	}

	if len(b.ControlIDs) > 0 {
		job.AddControlIDs(b.ControlIDs...)
	}

	result, err := job.Save(ctx)
	requireNoError(err)

	return result
}

// TrustCenterBuilder is used to create trust centers
type TrustCenterBuilder struct {
	client *client

	// Fields
	Slug           string
	CustomDomainID string
}

// TrustCenterSettingBuilder is used to create trust center settings
type TrustCenterSettingBuilder struct {
	client *client

	// Fields
	Title         string
	Overview      string
	PrimaryColor  string
	TrustCenterID string
	Tags          []string
}

// TrustCenterComplianceBuilder is used to create trust center compliance
type TrustCenterComplianceBuilder struct {
	client *client

	// Fields
	TrustCenterID string
	StandardID    string
	Tags          []string
}

// MustNew trust center builder is used to create, without authz checks, trust centers in the database
func (tc *TrustCenterBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TrustCenter {
	// do not use internal ctx or skip the checks so
	// the owner_id can be applied
	ctx = ent.NewContext(ctx, tc.client.db)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)
	ctx = graphql.WithResponseContext(ctx, gqlerrors.ErrorPresenter, graphql.DefaultRecover)

	if tc.Slug == "" {
		tc.Slug = randomName(t)
	}

	mutation := tc.client.db.TrustCenter.Create().
		SetSlug(tc.Slug)

	if tc.CustomDomainID != "" {
		mutation.SetCustomDomainID(tc.CustomDomainID)
	}

	trustCenter, err := mutation.Save(ctx)
	requireNoError(err)

	// Create the organization parent tuple for the trust center
	// This is normally done by the orgOwnedMixin, but since we're bypassing hooks, we need to do it manually
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		requireNoError(err)
	}

	parentReq := fgax.TupleRequest{
		SubjectID:   orgID,
		SubjectType: "organization",
		ObjectID:    trustCenter.ID,
		ObjectType:  "trust_center",
		Relation:    "parent",
	}

	tuple := fgax.GetTupleKey(parentReq)
	if _, err := tc.client.db.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{tuple}, nil); err != nil {
		requireNoError(err)
	}

	return trustCenter
}

// MustNew trust center setting builder is used to create, without authz checks, trust center settings in the database
func (tcs *TrustCenterSettingBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TrustCenterSetting {
	userCtx := ctx
	ctx = setContext(ctx, tcs.client.db)

	if tcs.Title == "" {
		tcs.Title = gofakeit.Company() + " Trust Center"
	}

	if tcs.Overview == "" {
		tcs.Overview = gofakeit.Sentence()
	}

	if tcs.PrimaryColor == "" {
		tcs.PrimaryColor = gofakeit.HexColor()
	}

	if tcs.TrustCenterID == "" {
		trustCenter := (&TrustCenterBuilder{client: tcs.client}).MustNew(userCtx, t)
		tcs.TrustCenterID = trustCenter.ID
	}

	if len(tcs.Tags) == 0 {
		tcs.Tags = []string{"test", "trust-center"}
	}

	mutation := tcs.client.db.TrustCenterSetting.Create().
		SetTitle(tcs.Title).
		SetOverview(tcs.Overview).
		SetPrimaryColor(tcs.PrimaryColor).
		SetTrustCenterID(tcs.TrustCenterID)

	trustCenterSetting, err := mutation.Save(ctx)
	requireNoError(err)

	return trustCenterSetting
}

func (tccb *TrustCenterComplianceBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TrustCenterCompliance {
	userCtx := ctx
	ctx = setContext(ctx, tccb.client.db)

	if tccb.TrustCenterID == "" {
		trustCenter := (&TrustCenterBuilder{client: tccb.client}).MustNew(userCtx, t)
		tccb.TrustCenterID = trustCenter.ID
	}

	if tccb.StandardID == "" {
		standard := (&StandardBuilder{client: tccb.client}).MustNew(ctx, t)
		tccb.StandardID = standard.ID
	}

	mutation := tccb.client.db.TrustCenterCompliance.Create().
		SetTrustCenterID(tccb.TrustCenterID).
		SetStandardID(tccb.StandardID)

	if len(tccb.Tags) > 0 {
		mutation.SetTags(tccb.Tags)
	}

	trustCenterCompliance, err := mutation.Save(ctx)
	requireNoError(err)

	return trustCenterCompliance
}

// IntegrationBuilder is used to create integrations
type IntegrationBuilder struct {
	client *client

	// Fields
	Name        string
	Description string
	Kind        string
}

// MustNew integration builder is used to create, without authz checks, integrations in the database
func (ib *IntegrationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Integration {
	ctx = setContext(ctx, ib.client.db)

	if ib.Name == "" {
		ib.Name = "GitHub Integration Test"
	}

	if ib.Description == "" {
		ib.Description = "Test integration for GraphQL tests"
	}

	if ib.Kind == "" {
		ib.Kind = "github"
	}

	mutation := ib.client.db.Integration.Create().
		SetName(ib.Name).
		SetDescription(ib.Description).
		SetKind(ib.Kind)

	integration, err := mutation.Save(ctx)
	requireNoError(err)

	return integration
}

// SecretBuilder is used to create secrets (hush)
type SecretBuilder struct {
	client *client

	// Fields
	Name           string
	Description    string
	Kind           string
	SecretName     string
	SecretValue    string
	IntegrationIDs []string
}

// WithIntegration adds an integration ID to the secret
func (sb *SecretBuilder) WithIntegration(integrationID string) *SecretBuilder {
	sb.IntegrationIDs = append(sb.IntegrationIDs, integrationID)
	return sb
}

// WithSecretName sets the secret name
func (sb *SecretBuilder) WithSecretName(name string) *SecretBuilder {
	sb.SecretName = name
	return sb
}

// WithSecretValue sets the secret value
func (sb *SecretBuilder) WithSecretValue(value string) *SecretBuilder {
	sb.SecretValue = value
	return sb
}

// MustNew secret builder is used to create, without authz checks, secrets in the database
func (sb *SecretBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Hush {
	ctx = setContext(ctx, sb.client.db)

	if sb.Name == "" {
		sb.Name = "Test Secret"
	}

	if sb.Description == "" {
		sb.Description = "Test secret for GraphQL tests"
	}

	if sb.Kind == "" {
		sb.Kind = "oauth_token"
	}

	if sb.SecretName == "" {
		sb.SecretName = "github_access_token"
	}

	if sb.SecretValue == "" {
		sb.SecretValue = "gho_test_token_123456"
	}

	mutation := sb.client.db.Hush.Create().
		SetName(sb.Name).
		SetDescription(sb.Description).
		SetKind(sb.Kind).
		SetSecretName(sb.SecretName).
		SetSecretValue(sb.SecretValue)

	// Add integration associations if provided
	if len(sb.IntegrationIDs) > 0 {
		mutation.AddIntegrationIDs(sb.IntegrationIDs...)
	}

	secret, err := mutation.Save(ctx)
	requireNoError(err)

	return secret
}

// IntegrationCleanup is used to delete integrations
type IntegrationCleanup struct {
	client *client
	ID     string
}

// MustDelete deletes the integration
func (ic *IntegrationCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = setContext(ctx, ic.client.db)

	err := ic.client.db.Integration.DeleteOneID(ic.ID).Exec(ctx)
	requireNoError(err)
}

// SecretCleanup is used to delete secrets
type SecretCleanup struct {
	client *client
	ID     string
}

// MustDelete deletes the secret
func (sc *SecretCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = setContext(ctx, sc.client.db)

	err := sc.client.db.Hush.DeleteOneID(sc.ID).Exec(ctx)
	requireNoError(err)
}

// MustNew file builder is used to create, without authz checks, files in the database
func (fb *FileBuilder) MustNew(ctx context.Context, t *testing.T) *ent.File {
	ctx = setContext(ctx, fb.client.db)

	if fb.Name == "" {
		fb.Name = gofakeit.Name()
	}

	url := gofakeit.URL()

	mutation := fb.client.db.File.Create().
		SetProvidedFileName(fb.Name).
		SetProvidedFileExtension("csv").
		SetDetectedContentType("application/csv").
		SetURI(url)

	file, err := mutation.Save(ctx)
	requireNoError(err)

	return file
}

// MustNew template builder is used to create, without authz checks, templates in the database
func (tb *TemplateBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Template {
	ctx = setContext(ctx, tb.client.db)

	if tb.Name == "" {
		tb.Name = gofakeit.Name()
	}

	if tb.Description == "" {
		tb.Description = gofakeit.HipsterSentence()
	}

	if tb.JSONConfig == nil {
		tb.JSONConfig = map[string]any{
			"key":   "value",
			"array": []string{"one", "two", "three"},
		}
	}
	mutation := tb.client.db.Template.Create().
		SetName(tb.Name).
		SetDescription(tb.Description).
		SetJsonconfig(tb.JSONConfig)

	if tb.Kind != "" {
		mutation.SetKind(tb.Kind)
	}

	if tb.TemplateType != "" {
		mutation.SetTemplateType(tb.TemplateType)
	}

	if tb.UISchema != nil {
		mutation.SetUischema(tb.UISchema)
	}

	template, err := mutation.Save(ctx)
	requireNoError(err)

	return template
}

// MustNew assessment builder is used to create, without authz checks, assessments in the database
func (ab *AssessmentBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Assessment {

	jsonConfig := map[string]any{
		"title":       "Test Assessment Template Missing",
		"description": "A test questionnaire template that will be deleted",
		"questions": []map[string]any{
			{
				"id":       "q1",
				"question": "What is your name?",
				"type":     "text",
			},
		},
	}

	ctx = setContext(ctx, ab.client.db)

	if ab.Name == "" {
		ab.Name = gofakeit.Company() + "-" + ulids.New().String()
	}

	if ab.TemplateID == "" {
		template := (&TemplateBuilder{client: ab.client}).MustNew(ctx, t)
		ab.TemplateID = template.ID
	}

	mutation := ab.client.db.Assessment.Create().
		SetName(ab.Name).
		SetTemplateID(ab.TemplateID)

	if ab.AssessmentType != "" {
		mutation.SetAssessmentType(ab.AssessmentType)
	}

	if ab.ResponseDueDuration > 0 {
		mutation.SetResponseDueDuration(ab.ResponseDueDuration)
	}

	if len(ab.Tags) > 0 {
		mutation.SetTags(ab.Tags)
	}

	mutation.SetJsonconfig(jsonConfig)

	assessment, err := mutation.Save(ctx)
	requireNoError(err)

	return assessment
}

// MustNew assessment response builder creates responses without authz checks
// This uses the QuestionnaireContextKey to bypass auth, simulating anonymous user creation
func (arb *AssessmentResponseBuilder) MustNew(ctx context.Context, t *testing.T) *ent.AssessmentResponse {
	ctx = setContext(ctx, arb.client.db)

	var assessment *ent.Assessment

	if arb.AssessmentID == "" {
		assessment = (&AssessmentBuilder{client: arb.client}).MustNew(ctx, t)
		arb.AssessmentID = assessment.ID
	}

	if arb.Email == "" {
		arb.Email = gofakeit.Email()
	}

	if arb.OwnerID == "" {
		if assessment == nil {
			var err error
			assessment, err = arb.client.db.Assessment.Get(ctx, arb.AssessmentID)
			requireNoError(err)
		}
		arb.OwnerID = assessment.OwnerID
	}

	// Use QuestionnaireContextKey to bypass auth checks (simulates anonymous JWT)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.QuestionnaireContextKey{})

	mutation := arb.client.db.AssessmentResponse.Create().
		SetAssessmentID(arb.AssessmentID).
		SetEmail(arb.Email).
		SetOwnerID(arb.OwnerID)

	if arb.DueDate != nil {
		mutation.SetDueDate(*arb.DueDate)
	}

	if arb.DocumentDataID != "" {
		mutation.SetDocumentDataID(arb.DocumentDataID)
	}

	response, err := mutation.Save(allowCtx)
	requireNoError(err)

	return response
}

// TrustCenterWatermarkConfigBuilder is used to create trust center watermark configs
type TrustCenterWatermarkConfigBuilder struct {
	client *client

	// Fields
	TrustCenterID string
	LogoID        *string
	Text          string
	FontSize      float64
	Opacity       float64
	Rotation      float64
	Color         string
	Font          enums.Font
}

// TrustCenterDocBuilder is used to create trust center documents
type TrustCenterDocBuilder struct {
	client *client

	// Fields
	Title         string
	Category      string
	TrustCenterID string
	FileID        string
	Visibility    enums.TrustCenterDocumentVisibility
	Tags          []string
}

// MustNew trust center doc builder is used to create trust center docs using the GraphQL API
func (tcdb *TrustCenterDocBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TrustCenterDoc {
	// save original context for trust center creation to preserve org scoping
	userCtx := ctx

	if tcdb.Title == "" {
		tcdb.Title = gofakeit.Sentence()
	}

	if tcdb.Category == "" {
		tcdb.Category = gofakeit.Word()
	}

	if tcdb.TrustCenterID == "" {
		trustCenter := (&TrustCenterBuilder{client: tcdb.client}).MustNew(userCtx, t)
		tcdb.TrustCenterID = trustCenter.ID
	}

	if len(tcdb.Tags) == 0 {
		tcdb.Tags = []string{"test", "document"}
	}

	// Create a test PDF file for upload
	pdfFile, err := storage.NewUploadFile("testdata/uploads/hello.pdf")
	requireNoError(err)

	fileUpload := graphql.Upload{
		File:        pdfFile.RawFile,
		Filename:    pdfFile.OriginalName,
		Size:        pdfFile.Size,
		ContentType: pdfFile.ContentType,
	}

	// Prepare the GraphQL input
	input := testclient.CreateTrustCenterDocInput{
		Title:         tcdb.Title,
		Category:      tcdb.Category,
		TrustCenterID: &tcdb.TrustCenterID,
		Tags:          tcdb.Tags,
	}

	if tcdb.Visibility != "" {
		input.Visibility = &tcdb.Visibility
	}

	// Expect the file upload in the object store
	expectUpload(t, tcdb.client.mockProvider, []graphql.Upload{fileUpload})

	// Create the trust center document using the GraphQL API
	resp, err := tcdb.client.api.CreateTrustCenterDoc(ctx, input, fileUpload)
	requireNoError(err)

	// Convert the GraphQL response to an ent entity
	// We need to fetch it from the database to get the full ent.TrustCenterDoc
	dbCtx := setContext(ctx, tcdb.client.db)
	trustCenterDoc, err := tcdb.client.db.TrustCenterDoc.Get(dbCtx, resp.CreateTrustCenterDoc.TrustCenterDoc.ID)
	requireNoError(err)

	return trustCenterDoc
}

// MustNew trust center watermark config builder is used to create, without authz checks, trust center watermark configs in the database
func (tcwcb *TrustCenterWatermarkConfigBuilder) MustNew(ctx context.Context, t *testing.T, trustCenterID string) *ent.TrustCenterWatermarkConfig {
	ctx = setContext(ctx, tcwcb.client.db)

	// Set the trust center ID from the parameter
	tcwcb.TrustCenterID = trustCenterID

	// Set default values if not provided
	if tcwcb.Text == "" && tcwcb.LogoID == nil {
		tcwcb.Text = "Test Watermark"
	}

	if tcwcb.FontSize == 0 {
		tcwcb.FontSize = 48.0
	}

	if tcwcb.Opacity == 0 {
		tcwcb.Opacity = 0.3
	}

	if tcwcb.Rotation == 0 {
		tcwcb.Rotation = 45.0
	}

	if tcwcb.Color == "" {
		tcwcb.Color = "#808080"
	}

	if tcwcb.Font == "" {
		tcwcb.Font = enums.FontHelvetica
	}

	mutation := tcwcb.client.db.TrustCenterWatermarkConfig.Create().
		SetTrustCenterID(tcwcb.TrustCenterID).
		SetFontSize(tcwcb.FontSize).
		SetOpacity(tcwcb.Opacity).
		SetRotation(tcwcb.Rotation).
		SetColor(tcwcb.Color).
		SetFont(tcwcb.Font)

	if tcwcb.Text != "" {
		mutation.SetText(tcwcb.Text)
	}

	if tcwcb.LogoID != nil {
		mutation.SetLogoID(*tcwcb.LogoID)
	}

	watermarkConfig, err := mutation.Save(ctx)
	requireNoError(err)

	return watermarkConfig
}

func (td *TagDefinitionBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TagDefinition {
	ctx = setContext(ctx, td.client.db)

	if td.Name == "" {
		td.Name = gofakeit.HipsterWord()
	}

	mutation := td.client.db.TagDefinition.Create().
		SetName(td.Name)

	if td.Color != "" {
		mutation.SetColor(td.Color)
	}

	tagDefinition, err := mutation.Save(ctx)
	requireNoError(err)

	return tagDefinition
}

func (td *CustomTypeEnumBuilder) MustNew(ctx context.Context, t *testing.T) *ent.CustomTypeEnum {
	ctx = setContext(ctx, td.client.db)

	if td.Name == "" {
		td.Name = gofakeit.HipsterWord()
	}

	if td.ObjectType == "" {
		td.ObjectType = "task"
	}

	mutation := td.client.db.CustomTypeEnum.Create().
		SetName(td.Name).
		SetObjectType(td.ObjectType)

	if td.Description != "" {
		mutation.SetDescription(td.Description)
	}

	if td.Color != "" {
		mutation.SetColor(td.Color)
	}

	customTypeEnum, err := mutation.Save(ctx)
	requireNoError(err)

	return customTypeEnum
}
