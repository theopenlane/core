//go:build test

package testharness

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/consts"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/v2/internal/ent/generated/evidence"
	"github.com/theopenlane/core/v2/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/programmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/sladefinition"
	"github.com/theopenlane/core/v2/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

type OrganizationBuilder struct {
	Client   *Client
	Features []models.OrgModule

	// Fields
	SystemOrg      bool
	Name           string
	DisplayName    string
	Description    *string
	ParentOrgID    string
	PersonalOrg    bool
	AllowedDomains []string
}

type GroupBuilder struct {
	Client *Client

	// Fields
	Name              string
	ControlEditorsIDs []string
	ProgramEditorsIDs []string
}

type UserBuilder struct {
	Client *Client

	// Fields
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type TFASettingBuilder struct {
	Client *Client

	totpAllowed *bool
}

type WebauthnBuilder struct {
	Client *Client
}

type OrgMemberBuilder struct {
	Client *Client

	// Fields
	UserID string
	Role   string
}

type GroupMemberBuilder struct {
	Client *Client

	// Fields
	UserID  string
	GroupID string
	Role    string
}

type InviteBuilder struct {
	Client *Client

	// Fields
	Recipient string
	Role      string
}

type PersonalAccessTokenBuilder struct {
	Client *Client

	// Fields
	Name            string
	Token           string
	Abilities       []string
	Description     string
	ExpiresAt       *time.Time
	OrganizationIDs []string
}

type APITokenBuilder struct {
	Client *Client

	// Fields
	Name        string
	Token       string
	Scopes      []string
	Description string
	ExpiresAt   *time.Time
}

type SubscriberBuilder struct {
	Client *Client

	// Fields
	Email string
}

type EntityBuilder struct {
	Client *Client

	// Fields
	Name        string
	DisplayName string
	TypeID      string
	Description string
	Tier        enums.VendorTier
}

type EntityTypeBuilder struct {
	Client *Client

	// Fields
	Name string
}

type IdentityHolderBuilder struct {
	Client *Client

	// Fields
	FullName   string
	Email      string
	Phone      string
	Title      string
	Department string
	Team       string
	Location   string
}

type DirectoryAccountBuilder struct {
	Client *Client

	// Fields
	ExternalID     string
	CanonicalEmail *string
	DisplayName    string
	GivenName      *string
	FamilyName     *string
	DirectoryName  *string
	Status         enums.DirectoryAccountStatus
	PrimarySource  bool
	JobTitle       *string
	Department     *string
	OwnerID        string
	PhoneNumber    *string
	EmailAliases   []string
}

type ContactBuilder struct {
	Client *Client

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
	Client *Client

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
	Client *Client

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
	Client *Client

	// Fields
	UserID    string
	ProgramID string
	Role      string
}

type ProcedureBuilder struct {
	Client *Client

	// Fields
	Name    string
	GroupID string
}

type InternalPolicyBuilder struct {
	Client *Client

	// Fields
	Name                    string
	BlockedGroupIDs         []string
	EditorGroupIDs          []string
	SkipApprovalRequirement bool
}

type RiskBuilder struct {
	Client *Client

	// Fields
	Name      string
	ProgramID string
}

type ControlObjectiveBuilder struct {
	Client *Client

	// Fields
	Name      string
	ProgramID string
}

type NarrativeBuilder struct {
	Client *Client

	// Fields
	Name      string
	ProgramID string
}

type ControlBuilder struct {
	Client *Client

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
	AllFields          bool
	Category           string
	Subcategory        string
	ReferenceFramework *string
}

type SubcontrolBuilder struct {
	Client *Client

	// Fields
	Name        string
	ControlID   string
	Category    string
	Subcategory string
}

type MappedControlBuilder struct {
	Client *Client

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
	Client *Client

	// Fields
	Name             string
	ProgramID        string
	ControlID        string
	InternalPolicyID string
	ProcedureID      string
	IncludeFile      bool
	Status           *enums.EvidenceStatus
}

type StandardBuilder struct {
	Client *Client

	// Fields
	Name      string
	Framework string
	IsPublic  bool
}

type SubprocessorBuilder struct {
	Client *Client

	// Fields
	Name          string
	Description   string
	LogoRemoteURL string
}

type NoteBuilder struct {
	Client *Client

	// Fields
	Text          string
	TaskID        string
	FileIDs       []string
	TrustCenterID string
}

type ControlImplementationBuilder struct {
	Client *Client

	// Fields
	Details            string
	ImplementationDate time.Time
	ControlIDs         []string
	SubcontrolIDs      []string
}

type MappableDomainBuilder struct {
	Client *Client

	// Fields
	Name   string
	ZoneID string
}

type FileBuilder struct {
	Client *Client

	// Fields
	Name    string
	MD5Hash string
}

type TemplateBuilder struct {
	Client *Client

	// Fields
	Name          string
	Description   string
	Kind          enums.TemplateKind
	TemplateType  enums.DocumentType
	JSONConfig    map[string]any
	UISchema      map[string]any
	TrustCenterID string
	FileIDs       []string
}

type AssessmentBuilder struct {
	Client *Client

	// Fields
	Name                string
	AssessmentType      enums.AssessmentType
	TemplateID          string
	ResponseDueDuration int64
	Tags                []string
}

type AssessmentResponseBuilder struct {
	Client *Client

	// Fields
	AssessmentID   string
	Email          string
	OwnerID        string
	DueDate        *time.Time
	DocumentDataID string
}

type TagDefinitionBuilder struct {
	Client *Client

	// Fields
	Name  string
	Color string
}

type CustomTypeEnumBuilder struct {
	Client *Client

	// Fields
	Name        string
	Description string
	Color       string
	ObjectType  string
	Field       string
}

type AssetBuilder struct {
	Client *Client

	// Fields
	Name string
}

type SLADefinitionBuilder struct {
	Client *Client

	// Fields
	SLADays       int
	SecurityLevel enums.SecurityLevel
}

type PlatformBuilder struct {
	Client *Client

	// Fields
	Name string
}

// Faker structs with random injected data
type Faker struct {
	Name string
}

func RandomName(t *testing.T) string {
	var f Faker
	err := gofakeit.Struct(&f)
	RequireNoError(t, err)

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
	Client DeleteClient[T]

	// Fields
	ID  string
	IDs []string
}

// MustDelete deletes the entities without authz checks for the type
// this should normally look like this:
// type: generated.OrganizationDeleteOne (replace Organization with the entity you want to delete)
// Client: suite.Client.DB.Organization (replace Organization with the entity you want to delete)
//
//	(&Cleanup[*generated.OrganizationDeleteOne]{
//		Client: suite.Client.DB.Organization,
//		ID: resp.CreateOrganization.Organization.ID}).
//		MustDelete(SharedTestUser1.UserCtx, t)
//
// Special handling for standards - update them to be private before deletion
// this is to allow the system admin to delete public standards
// and controls that are linked to them
// this is a workaround to avoid the cascade delete hook on standard
// that would otherwise prevent the deletion of public standards
// and controls that are linked to them
func (c *Cleanup[DeleteExec]) MustDelete(ctx context.Context, t *testing.T) {
	// add client to context for hooks that expect the client to be in the context
	ctx = SetContext(ctx, Suite.Client.DB)

	// Special handling for standards - update them to be private before deletion
	// Only do this for system admins
	stdAdminCaller, stdAdminOk := auth.CallerFromContext(ctx)
	if _, ok := any(c.Client).(*ent.StandardClient); ok && stdAdminOk && stdAdminCaller != nil && stdAdminCaller.Has(auth.CapSystemAdmin) {
		if c.ID != "" {
			err := Suite.Client.DB.Standard.UpdateOneID(c.ID).SetIsPublic(false).Exec(ctx)
			RequireNoError(t, err)
		}
		for _, id := range c.IDs {
			err := Suite.Client.DB.Standard.UpdateOneID(id).SetIsPublic(false).Exec(ctx)
			RequireNoError(t, err)
		}
	}

	for _, id := range c.IDs {
		err := c.Client.DeleteOneID(id).Exec(ctx)
		RequireNoError(t, err)
	}

	if c.ID != "" {
		err := c.Client.DeleteOneID(c.ID).Exec(ctx)
		RequireNoError(t, err)
	}
}

// SetContext is a helper function to set the context for the client
// setting privacy to allow and adding the client to the context
func SetContext(ctx context.Context, db *ent.Client) context.Context {
	ctx = ent.NewContext(rule.WithInternalContext(ctx), db)

	// add the GraphQL response context to prevent panics from interceptors that call graphql.AddError
	return graphql.WithResponseContext(ctx, gqlerrors.ErrorPresenter, graphql.DefaultRecover)
}

// MustNew organization builder is used to create, without authz checks, orgs in the database
func (o *OrganizationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Organization {
	// no auth, so allow policy
	ctx = SetContext(ctx, o.Client.DB)

	if o.SystemOrg {
		systemOrg, err := o.Client.DB.Organization.Create().SetID(consts.SystemAdminOrgID).SetName("System Admin Organization").SetDisplayName("System Admin Organization").SetPersonalOrg(true).SetDescription("Organization for system administrators").Save(ctx)
		RequireNoError(t, err)

		return systemOrg
	}

	if o.Name == "" {
		o.Name = RandomName(t)
	}

	if o.DisplayName == "" {
		o.DisplayName = gofakeit.LetterN(40)
	}

	if o.Description == nil {
		desc := gofakeit.HipsterSentence()
		o.Description = &desc
	}

	m := o.Client.DB.Organization.Create().SetName(o.Name).SetDescription(*o.Description).SetDisplayName(o.DisplayName).SetPersonalOrg(o.PersonalOrg)

	if o.ParentOrgID != "" {
		m.SetParentID(o.ParentOrgID)
	}

	org, err := m.Save(ctx)
	RequireNoError(t, err)

	orgSetting, err := org.Setting(ctx)
	RequireNoError(t, err)
	update := orgSetting.Update()

	if o.AllowedDomains != nil {
		update.SetAllowedEmailDomains(o.AllowedDomains)
	}

	// turn on so all tests have this by default
	err = update.SetAllowSupportAccess(true).Exec(ctx)
	RequireNoError(t, err)

	o.enableModules(ctx, t, org.ID)

	return org
}

// enableModules enables the selected organization modules for the given organization
func (o *OrganizationBuilder) enableModules(ctx context.Context, t *testing.T, orgID string) {
	features := o.Features

	if len(o.Features) == 0 {
		features = models.AllOrgModules
	}

	err := entitlements.CreateFeatureTuples(ctx, o.Client.FGA, orgID, features)
	assert.NilError(t, err)

	for _, feature := range features {
		n, err := o.Client.DB.OrgModule.Update().
			Where(
				orgmodule.OwnerID(orgID),
				orgmodule.Module(feature),
			).
			SetActive(true).
			Save(ctx)
		assert.NilError(t, err)

		// if no rows were updated, the module wasn't created - create it now
		if n == 0 {
			err = o.Client.DB.OrgModule.Create().
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
	ctx = SetContext(ctx, u.Client.DB)

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
	userSetting, err := u.Client.DB.UserSetting.Create().Save(ctx)
	RequireNoError(t, err)

	user, err := u.Client.DB.User.Create().
		SetFirstName(u.FirstName).
		SetLastName(u.LastName).
		SetEmail(u.Email).
		SetPassword(u.Password).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SetSetting(userSetting).
		Save(ctx)
	RequireNoError(t, err)

	_, err = user.Edges.Setting.DefaultOrg(ctx)
	RequireNoError(t, err)

	return user
}

// MustNew tfa settings builder is used to create, without authz checks, tfa settings in the database
func (tf *TFASettingBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TFASetting {
	if tf.totpAllowed == nil {
		tf.totpAllowed = lo.ToPtr(true)
	}

	setting, err := tf.Client.DB.TFASetting.Create().
		SetTotpAllowed(*tf.totpAllowed).
		Save(ctx)

	// if the setting is not created, it means the user already has a setting
	// and let's skip for seeding
	if errors.Is(err, generated.ConstraintError{}) {
		return nil
	}

	return setting
}

func getValidIPAddress(t *testing.T) string {
	maxAttempts := 10
	attempts := 0
	for {
		ip := gofakeit.IPv4Address()
		if err := models.ValidateIP(ip); err == nil {
			return ip
		}
		attempts++

		if attempts >= maxAttempts {
			t.Fail()
		}
	}
}

// MustNew webauthn settings builder is used to create passkeys without the browser setup process
func (w *WebauthnBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Webauthn {
	uuidBytes, err := uuid.NewUUID()
	RequireNoError(t, err)

	wn, err := w.Client.DB.Webauthn.Create().
		SetAaguid(models.ToAAGUID(uuidBytes[:])).
		SetAttestationType("type").
		SetBackupEligible(true).
		SetBackupState(true).
		SetSignCount(10).
		SetCredentialID([]byte(uuid.NewString())).
		SetTransports([]string{uuid.NewString()}).
		Save(ctx)
	RequireNoError(t, err)

	return wn
}

// MustNew org members builder is used to create, without authz checks, org members in the database
func (om *OrgMemberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.OrgMembership {
	ctx = SetContext(ctx, om.Client.DB)

	if om.UserID == "" {
		user := (&UserBuilder{Client: om.Client}).MustNew(ctx, t)
		om.UserID = user.ID
	}

	role := enums.ToRole(om.Role)
	if role == &enums.RoleInvalid {
		role = &enums.RoleMember
	}

	orgMember, err := om.Client.DB.OrgMembership.Create().
		SetUserID(om.UserID).
		SetRole(*role).
		Save(ctx)
	RequireNoError(t, err)

	return orgMember
}

// MustNew group builder is used to create, without authz checks, groups in the database
func (g *GroupBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Group {
	ctx = SetContext(ctx, g.Client.DB)

	if g.Name == "" {
		g.Name = RandomName(t)
	}

	mutation := g.Client.DB.Group.Create().SetName(g.Name)

	if len(g.ControlEditorsIDs) > 0 {
		mutation.AddControlEditorIDs(g.ControlEditorsIDs...)
	}

	if len(g.ProgramEditorsIDs) > 0 {
		mutation.AddProgramEditorIDs(g.ProgramEditorsIDs...)
	}

	group, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return group
}

// MustNew invite builder is used to create, without authz checks, invites in the database
func (i *InviteBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Invite {
	ctx = SetContext(ctx, i.Client.DB)

	// create user if not provided
	rec := i.Recipient

	if rec == "" {
		rec = strings.ToLower(fmt.Sprintf("%s@%s", ulids.New().String(), "theopenlane.io"))
	}

	inviteQuery := i.Client.DB.Invite.Create().
		SetRecipient(rec)

	if i.Role != "" {
		inviteQuery.SetRole(*enums.ToRole(i.Role))
	}

	invite, err := inviteQuery.Save(ctx)
	RequireNoError(t, err)

	return invite
}

// MustNew subscriber builder is used to create, without authz checks, subscribers in the database
func (i *SubscriberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Subscriber {
	reqCtx := SetContext(ctx, i.Client.DB)

	// create user if not provided
	rec := i.Email

	if rec == "" {
		rec = gofakeit.Email()
	}

	sub, err := i.Client.DB.Subscriber.Create().
		SetEmail(rec).
		SetActive(true).Save(reqCtx)
	RequireNoError(t, err)

	return sub
}

// MustNew personal access tokens builder is used to create, without authz checks, personal access tokens in the database
func (pat *PersonalAccessTokenBuilder) MustNew(ctx context.Context, t *testing.T) *ent.PersonalAccessToken {
	ctx = SetContext(ctx, pat.Client.DB)

	if pat.Name == "" {
		pat.Name = gofakeit.AppName()
	}

	if pat.Description == "" {
		pat.Description = gofakeit.HipsterSentence()
	}

	if pat.OrganizationIDs == nil {
		// default to adding the test users organization ID
		pat.OrganizationIDs = []string{SharedTestUser1.OrganizationID}
	}

	request := pat.Client.DB.PersonalAccessToken.Create().
		SetName(pat.Name).
		SetDescription(pat.Description).
		AddOrganizationIDs(pat.OrganizationIDs...)

	if pat.ExpiresAt != nil {
		request.SetExpiresAt(*pat.ExpiresAt)
	}

	token, err := request.Save(ctx)
	RequireNoError(t, err)

	return token
}

// MustNew api tokens builder is used to create, without authz checks, api tokens in the database
func (at *APITokenBuilder) MustNew(ctx context.Context, t *testing.T) *ent.APIToken {
	ctx = SetContext(ctx, at.Client.DB)

	if at.Name == "" {
		at.Name = gofakeit.AppName()
	}

	if at.Description == "" {
		at.Description = gofakeit.HipsterSentence()
	}

	request := at.Client.DB.APIToken.Create().
		SetName(at.Name).
		SetDescription(at.Description)

	if at.Scopes != nil {
		request.SetScopes(at.Scopes)
	}

	if at.ExpiresAt != nil {
		request.SetExpiresAt(*at.ExpiresAt)
	}

	token, err := request.Save(ctx)
	RequireNoError(t, err)

	return token
}

// MustNew user builder is used to create, without authz checks, group members in the database
func (gm *GroupMemberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.GroupMembership {
	ctx = SetContext(ctx, gm.Client.DB)

	if gm.GroupID == "" {
		group := (&GroupBuilder{Client: gm.Client}).MustNew(ctx, t)
		gm.GroupID = group.ID
	}

	if gm.UserID == "" {
		orgMember := (&OrgMemberBuilder{Client: gm.Client}).MustNew(ctx, t)
		gm.UserID = orgMember.UserID
	}

	mut := gm.Client.DB.GroupMembership.Create().
		SetUserID(gm.UserID).
		SetGroupID(gm.GroupID)

	if gm.Role != "" {
		mut.SetRole(*enums.ToRole(gm.Role))
	}

	groupMember, err := mut.Save(ctx)
	RequireNoError(t, err)

	gmToReturn, err := gm.Client.DB.GroupMembership.Query().
		WithUser().
		WithOrgMembership().
		Where(groupmembership.ID(groupMember.ID)).Only(ctx)
	RequireNoError(t, err)

	return gmToReturn
}

// MustNew entity type builder is used to create, without authz checks, entity types in the database
func (e *EntityTypeBuilder) MustNew(ctx context.Context, t *testing.T) *ent.EntityType {
	ctx = SetContext(ctx, e.Client.DB)

	if e.Name == "" {
		e.Name = RandomName(t)
	}

	entityType, err := e.Client.DB.EntityType.Create().
		SetName(e.Name).
		Save(ctx)
	RequireNoError(t, err)

	return entityType
}

// MustNew entity builder is used to create, without authz checks, entities in the database
func (e *EntityBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Entity {
	ctx = SetContext(ctx, e.Client.DB)

	if e.Name == "" {
		e.Name = gofakeit.LoremIpsumWord() + ulids.New().String()
	}

	if e.DisplayName == "" {
		e.DisplayName = e.Name
	}

	if e.Description == "" {
		e.Description = gofakeit.HipsterSentence()
	}

	if e.Tier == "" {
		e.Tier = enums.VendorTierStandard
	}

	if e.TypeID == "" {
		et := (&EntityTypeBuilder{Client: e.Client}).MustNew(ctx, t)
		e.TypeID = et.ID
	}

	entity := e.Client.DB.Entity.Create().
		SetName(e.Name).
		SetDisplayName(e.DisplayName).
		SetEntityTypeID(e.TypeID).
		SetDescription(e.Description).
		SetTier(e.Tier)

	savedEntity, err := entity.Save(ctx)
	RequireNoError(t, err)

	return savedEntity
}

// MustNew identity holder builder is used to create, without authz checks, identity holders in the database
func (i *IdentityHolderBuilder) MustNew(ctx context.Context, t *testing.T) *ent.IdentityHolder {
	ctx = SetContext(ctx, i.Client.DB)

	if i.FullName == "" {
		i.FullName = gofakeit.Name()
	}

	if i.Email == "" {
		i.Email = gofakeit.Email()
	}

	if i.Phone == "" {
		i.Phone = gofakeit.Phone()
	}

	if i.Title == "" {
		i.Title = gofakeit.JobTitle()
	}

	if i.Department == "" {
		i.Department = gofakeit.JobDescriptor()
	}

	if i.Team == "" {
		i.Team = gofakeit.AppName()
	}

	if i.Location == "" {
		i.Location = gofakeit.City()
	}

	entity, err := i.Client.DB.IdentityHolder.Create().
		SetFullName(i.FullName).
		SetEmail(i.Email).
		SetPhoneNumber(i.Phone).
		SetTitle(i.Title).
		SetDepartment(i.Department).
		SetTeam(i.Team).
		SetLocation(i.Location).
		Save(ctx)
	RequireNoError(t, err)

	return entity
}

// MustNew directory account builder is used to create, without authz checks, directory accounts in the database
func (d *DirectoryAccountBuilder) MustNew(ctx context.Context, t *testing.T) *ent.DirectoryAccount {
	ctx = SetContext(ctx, d.Client.DB)

	if d.ExternalID == "" {
		d.ExternalID = ulids.New().String()
	}

	if d.Status == "" {
		d.Status = enums.DirectoryAccountStatusActive
	}

	create := d.Client.DB.DirectoryAccount.Create().
		SetExternalID(d.ExternalID).
		SetDisplayName(d.DisplayName).
		SetStatus(d.Status).
		SetPrimarySource(d.PrimarySource).
		SetNillableCanonicalEmail(d.CanonicalEmail).
		SetNillableGivenName(d.GivenName).
		SetNillableFamilyName(d.FamilyName).
		SetNillableDirectoryName(d.DirectoryName).
		SetNillableJobTitle(d.JobTitle).
		SetNillableDepartment(d.Department).
		SetNillablePhoneNumber(d.PhoneNumber).
		SetEmailAliases(d.EmailAliases)

	if d.OwnerID != "" {
		create.SetOwnerID(d.OwnerID)
	}

	entity, err := create.Save(ctx)
	RequireNoError(t, err)

	return entity
}

// MustNew contact builder is used to create, without authz checks, contacts in the database
func (c *ContactBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Contact {
	ctx = SetContext(ctx, c.Client.DB)

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

	entity, err := c.Client.DB.Contact.Create().
		SetFullName(c.Name).
		SetEmail(c.Email).
		SetPhoneNumber(c.Phone).
		SetAddress(c.Address).
		SetTitle(c.Title).
		SetCompany(c.Company).
		Save(ctx)
	RequireNoError(t, err)

	return entity
}

// MustNew task builder is used to create, without authz checks, tasks in the database
func (c *TaskBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Task {
	ctx = SetContext(ctx, c.Client.DB)

	if c.Title == "" {
		c.Title = gofakeit.AppName()
	}

	if c.Details == "" {
		c.Details = gofakeit.HipsterSentence()
	}

	taskCreate := c.Client.DB.Task.Create().
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
	RequireNoError(t, err)

	return task
}

// MustNew program builder is used to create, without authz checks, programs in the database
func (p *ProgramBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Program {
	ctx = SetContext(ctx, p.Client.DB)

	if p.Name == "" {
		p.Name = gofakeit.AppName()
	}

	mutation := p.Client.DB.Program.Create().
		SetName(p.Name)

	if p.WithProcedure {
		procedure := (&ProcedureBuilder{Client: p.Client, Name: gofakeit.AppName()}).MustNew(ctx, t)
		mutation.AddProcedureIDs(procedure.ID)
	}

	if p.WithPolicy {
		policy := (&InternalPolicyBuilder{Client: p.Client, Name: gofakeit.AppName()}).MustNew(ctx, t)
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
	RequireNoError(t, err)

	return program
}

// MustNew user builder is used to create, without authz checks, program members in the database
func (pm *ProgramMemberBuilder) MustNew(ctx context.Context, t *testing.T) *ent.ProgramMembership {
	ctx = SetContext(ctx, pm.Client.DB)

	if pm.ProgramID == "" {
		program := (&ProgramBuilder{Client: pm.Client}).MustNew(ctx, t)
		pm.ProgramID = program.ID
	}

	if pm.UserID == "" {
		// first create an org member
		orgMember := (&OrgMemberBuilder{Client: pm.Client}).MustNew(ctx, t)
		pm.UserID = orgMember.UserID
	}

	mutation := pm.Client.DB.ProgramMembership.Create().
		SetUserID(pm.UserID).
		SetProgramID(pm.ProgramID)

	if pm.Role != "" {
		mutation.SetRole(*enums.ToRole(pm.Role))
	}

	programMember, err := mutation.Save(ctx)
	RequireNoError(t, err)

	programMember, err = pm.Client.DB.ProgramMembership.Query().
		WithUser().
		WithOrgMembership().
		Where(programmembership.ID(programMember.ID)).Only(ctx)
	RequireNoError(t, err)

	return programMember
}

// MustNew procedure builder is used to create, without authz checks, procedures in the database
func (p *ProcedureBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Procedure {
	ctx = SetContext(ctx, p.Client.DB)

	if p.Name == "" {
		p.Name = gofakeit.AppName()
	}

	mutation := p.Client.DB.Procedure.Create().
		SetName(p.Name)

	if p.GroupID != "" {
		mutation.AddEditorIDs(p.GroupID)
	}

	procedure, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return procedure
}

// MustNew policy builder is used to create, without authz checks, policies in the database
func (p *InternalPolicyBuilder) MustNew(ctx context.Context, t *testing.T) *ent.InternalPolicy {
	ctx = SetContext(ctx, p.Client.DB)

	if p.Name == "" {
		p.Name = gofakeit.AppName()
	}

	mut := p.Client.DB.InternalPolicy.Create().
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
	RequireNoError(t, err)

	return policy
}

// MustNew risk builder is used to create, without authz checks, risks in the database
func (r *RiskBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Risk {
	ctx = SetContext(ctx, r.Client.DB)

	if r.Name == "" {
		r.Name = gofakeit.AppName()
	}

	mutation := r.Client.DB.Risk.Create().
		SetName(r.Name)

	if r.ProgramID != "" {
		mutation.AddProgramIDs(r.ProgramID)
	}

	risk, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return risk
}

// MustNew control objective builder is used to create, without authz checks, control objectives in the database
func (c *ControlObjectiveBuilder) MustNew(ctx context.Context, t *testing.T) *ent.ControlObjective {
	ctx = SetContext(ctx, c.Client.DB)

	if c.Name == "" {
		c.Name = gofakeit.AppName()
	}

	mutation := c.Client.DB.ControlObjective.Create().
		SetName(c.Name)

	if c.ProgramID != "" {
		mutation.AddProgramIDs(c.ProgramID)
	}

	co, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return co
}

// MustNew narrative builder is used to create, without authz checks, narratives in the database
func (n *NarrativeBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Narrative {
	ctx = SetContext(ctx, n.Client.DB)

	if n.Name == "" {
		n.Name = gofakeit.AppName()
	}

	mutation := n.Client.DB.Narrative.Create().
		SetName(n.Name)

	if n.ProgramID != "" {
		mutation.AddProgramIDs(n.ProgramID)
	}

	narrative, err := mutation.
		Save(ctx)
	RequireNoError(t, err)

	return narrative
}

// MustNew control builder is used to create, without authz checks, controls in the database
func (c *ControlBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Control {
	ctx = SetContext(ctx, c.Client.DB)

	if c.RefCode == "" {
		c.RefCode = gofakeit.UUID()
	}

	if c.Title == "" {
		c.Title = gofakeit.HipsterSentence()
	}

	mutation := c.Client.DB.Control.Create().
		SetRefCode(c.RefCode).SetTitle(c.Title).SetNillableReferenceFramework(c.ReferenceFramework)

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
	RequireNoError(t, err)

	return control
}

// MustNew subcontrol builder is used to create, without authz checks, subcontrols in the database
func (s *SubcontrolBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Subcontrol {
	ctx = SetContext(ctx, s.Client.DB)

	if s.Name == "" {
		s.Name = gofakeit.UUID()
	}

	mutation := s.Client.DB.Subcontrol.Create().
		SetRefCode(s.Name)

	if s.ControlID == "" {
		control := (&ControlBuilder{Client: s.Client}).MustNew(ctx, t)
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

	RequireNoError(t, err)

	return sc
}

// MustNew control builder is used to create, without authz checks, controls in the database
func (e *EvidenceBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Evidence {
	ctx = SetContext(ctx, e.Client.DB)

	if e.Name == "" {
		e.Name = gofakeit.AppName()
	}

	mutation := e.Client.DB.Evidence.Create().
		SetCreationDate(models.DateTime(time.Now().Add(-time.Minute))).
		SetName(e.Name).
		SetNillableStatus(e.Status)

	if e.ProgramID != "" {
		mutation.AddProgramIDs(e.ProgramID)
	}

	if e.ControlID != "" {
		mutation.AddControlIDs(e.ControlID)
	}

	if e.InternalPolicyID != "" {
		mutation.AddInternalPolicyIDs(e.InternalPolicyID)
	}

	if e.ProcedureID != "" {
		mutation.AddProcedureIDs(e.ProcedureID)
	}

	if e.IncludeFile {
		file := (&FileBuilder{Client: e.Client, Name: e.Name}).MustNew(ctx, t)

		mutation.AddFileIDs(file.ID)
	}

	ev, err := mutation.
		Save(ctx)
	RequireNoError(t, err)

	if e.IncludeFile {
		ev, err := e.Client.DB.Evidence.Query().WithFiles().Where(evidence.ID(ev.ID)).Only(ctx)
		RequireNoError(t, err)

		return ev
	}

	return ev
}

// MustNew standard builder is used to create, without authz checks, standards in the database
func (s *StandardBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Standard {
	ctx = SetContext(ctx, s.Client.DB)

	if s.Name == "" {
		s.Name = gofakeit.AppName()
	}

	if s.Framework == "" {
		s.Framework = "MITB Framework"
	}

	mut := s.Client.DB.Standard.Create().
		SetName(s.Name).
		SetFramework(s.Framework).
		SetIsPublic(s.IsPublic)

	standard, err := mut.Save(ctx)
	RequireNoError(t, err)

	return standard
}

// MustNew subprocessor builder is used to create, without authz checks, subprocessors in the database
func (s *SubprocessorBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Subprocessor {
	ctx = SetContext(ctx, s.Client.DB)

	if s.Name == "" {
		for {
			s.Name = gofakeit.Company()
			_, err := s.Client.DB.Subprocessor.Query().Where(subprocessor.Name(s.Name)).Only(ctx)
			if err != nil {
				break
			}
		}
	}

	mutation := s.Client.DB.Subprocessor.Create().
		SetName(s.Name)

	if s.Description != "" {
		mutation.SetDescription(s.Description)
	}

	if s.LogoRemoteURL != "" {
		mutation.SetLogoRemoteURL(s.LogoRemoteURL)
	}

	subprocessor, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return subprocessor
}

// MustNew note builder is used to create, without authz checks, notes in the database
func (n *NoteBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Note {
	ctx = SetContext(ctx, n.Client.DB)

	if n.Text == "" {
		n.Text = gofakeit.HipsterSentence()
	}

	mutation := n.Client.DB.Note.Create().
		SetText(n.Text)

	if n.TaskID != "" {
		mutation.SetTaskID(n.TaskID)
	}

	if len(n.FileIDs) > 0 {
		mutation.AddFileIDs(n.FileIDs...)
	}

	if n.TrustCenterID != "" {
		mutation.SetTrustCenterID(n.TrustCenterID)
	}

	note, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return note
}

// MustNew controlImplementation builder is used to create, without authz checks, controlImplementations in the database
func (e *ControlImplementationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.ControlImplementation {
	ctx = SetContext(ctx, e.Client.DB)

	if e.Details == "" {
		e.Details = gofakeit.Paragraph()
	}

	if e.ImplementationDate.IsZero() {
		e.ImplementationDate = time.Now()
	}

	mutation := e.Client.DB.ControlImplementation.Create().
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
	RequireNoError(t, err)

	return controlImplementation
}

// MustNew controlImplementation builder is used to create, without authz checks, controlImplementations in the database
func (e *MappedControlBuilder) MustNew(ctx context.Context, t *testing.T) *ent.MappedControl {
	if ctx == SharedSystemAdminUser.UserCtx {
		if e.InternalID == "" {
			e.InternalID = ulids.New().String()
		}

		if e.InternalNotes == "" {
			e.InternalNotes = "Created by system admin user"
		}
	}

	ctx = SetContext(ctx, e.Client.DB)

	if len(e.FromControlIDs) == 0 && len(e.FromSubcontrolIDs) == 0 {
		fromControl := (&ControlBuilder{Client: e.Client}).MustNew(ctx, t)
		e.FromControlIDs = []string{fromControl.ID}
	}

	if len(e.ToControlIDs) == 0 && len(e.ToSubcontrolIDs) == 0 {
		toControl := (&ControlBuilder{Client: e.Client}).MustNew(ctx, t)
		e.ToControlIDs = []string{toControl.ID}
	}

	mutation := e.Client.DB.MappedControl.Create().
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
	RequireNoError(t, err)

	res, err := e.Client.DB.MappedControl.Query().
		WithFromControls().
		WithFromSubcontrols().
		WithToControls().
		WithToSubcontrols().
		Where(mappedcontrol.ID(mappedControl.ID)).Only(ctx)

	return res
}

// MustNew mappable domain builder is used to create, without authz checks, mappable domains in the database
func (e *MappableDomainBuilder) MustNew(ctx context.Context, t *testing.T) *ent.MappableDomain {
	ctx = SetContext(ctx, e.Client.DB)

	if e.Name == "" {
		e.Name = gofakeit.DomainName()
	}
	if e.ZoneID == "" {
		e.ZoneID = gofakeit.UUID()
	}

	mappableDomain, err := e.Client.DB.MappableDomain.Create().
		SetName(e.Name).
		SetZoneID(e.ZoneID).
		Save(ctx)
	RequireNoError(t, err)

	return mappableDomain
}

// CustomDomainBuilder is used to create custom domains
type CustomDomainBuilder struct {
	Client *Client

	// Fields
	CnameRecord      string
	MappableDomainID string
}

// DNSVerificationBuilder is used to create DNS verifications
type DNSVerificationBuilder struct {
	Client *Client

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
	ctx = SetContext(ctx, c.Client.DB)

	if c.CnameRecord == "" {
		c.CnameRecord = gofakeit.DomainName()
	}

	if c.MappableDomainID == "" {
		mappableDomain := (&MappableDomainBuilder{Client: c.Client}).MustNew(ctx, t)
		c.MappableDomainID = mappableDomain.ID
	}

	customDomain, err := c.Client.DB.CustomDomain.Create().
		SetCnameRecord(c.CnameRecord).
		SetMappableDomainID(c.MappableDomainID).
		Save(ctx)
	RequireNoError(t, err)

	return customDomain
}

// MustNew DNS verification builder is used to create, without authz checks, DNS verifications in the database
func (d *DNSVerificationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.DNSVerification {
	ctx = SetContext(ctx, d.Client.DB)

	if d.CloudflareHostnameID == "" {
		d.CloudflareHostnameID = gofakeit.UUID()
	}

	if d.DNSTxtRecord == "" {
		d.DNSTxtRecord = "_cf-verify." + gofakeit.DomainName()
	}

	if d.DNSTxtValue == "" {
		d.DNSTxtValue = gofakeit.UUID()
	}

	mutation := d.Client.DB.DNSVerification.Create().
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
	RequireNoError(t, err)

	return dnsVerification
}

const testScriptURL = "https://raw.githubusercontent.com/theopenlane/jobs-examples/refs/heads/main/basic/print.go"

// TrustCenterBuilder is used to create trust centers
type TrustCenterBuilder struct {
	Client *Client

	// Fields
	Slug           string
	CustomDomainID string
}

// TrustCenterSettingBuilder is used to create trust center settings
type TrustCenterSettingBuilder struct {
	Client *Client

	// Fields
	Title         string
	Overview      string
	PrimaryColor  string
	TrustCenterID string
	Tags          []string
}

// TrustCenterComplianceBuilder is used to create trust center compliance
type TrustCenterComplianceBuilder struct {
	Client *Client

	// Fields
	TrustCenterID string
	StandardID    string
	Tags          []string
}

// EmailTemplateBuilder is used to create email templates
type EmailTemplateBuilder struct {
	Client *Client

	// Fields
	Name            string
	Key             string
	TemplateContext *enums.TemplateContext
}

// MustNew trust center builder is used to create, without authz checks, trust centers in the database
func (tc *TrustCenterBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TrustCenter {
	// Add the database client to context so the authz client is available for feature checks
	// Do not use internal ctx or skip privacy checks so the owner_id can be applied correctly
	ctx = ent.NewContext(ctx, tc.Client.DB)
	ctx = graphql.WithResponseContext(ctx, gqlerrors.ErrorPresenter, graphql.DefaultRecover)

	if tc.Slug == "" {
		tc.Slug = RandomName(t)
	}

	mutation := tc.Client.DB.TrustCenter.Create().
		SetSlug(tc.Slug)

	if tc.CustomDomainID != "" {
		mutation.SetCustomDomainID(tc.CustomDomainID)
	}

	// set the org owner_id, this is done via hooks when using the api
	caller, _ := auth.CallerFromContext(ctx)
	if caller != nil && caller.OrganizationID != "" {
		mutation.SetOwnerID(caller.OrganizationID)
	}

	trustCenter, err := mutation.Save(ctx)
	RequireNoError(t, err)

	// the trust center create hook seeds a customizable message-updates email template; remove it when
	// the test ends so it does not leak into shared-org email template assertions
	t.Cleanup(func() {
		cleanupCtx := privacy.DecisionContext(SetContext(ctx, tc.Client.DB), privacy.Allow)
		_, _ = tc.Client.DB.EmailTemplate.Delete().Where(emailtemplate.TrustCenterID(trustCenter.ID)).Exec(cleanupCtx)
	})

	return trustCenter
}

// MustNew trust center setting builder is used to create, without authz checks, trust center settings in the database
func (tcs *TrustCenterSettingBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TrustCenterSetting {
	userCtx := ctx
	ctx = SetContext(ctx, tcs.Client.DB)

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
		trustCenter := (&TrustCenterBuilder{Client: tcs.Client}).MustNew(userCtx, t)
		tcs.TrustCenterID = trustCenter.ID
	}

	if len(tcs.Tags) == 0 {
		tcs.Tags = []string{"test", "trust-center"}
	}

	mutation := tcs.Client.DB.TrustCenterSetting.Create().
		SetTitle(tcs.Title).
		SetOverview(tcs.Overview).
		SetPrimaryColor(tcs.PrimaryColor).
		SetTrustCenterID(tcs.TrustCenterID)

	trustCenterSetting, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return trustCenterSetting
}

func (tccb *TrustCenterComplianceBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TrustCenterCompliance {
	userCtx := ctx
	ctx = SetContext(ctx, tccb.Client.DB)

	if tccb.TrustCenterID == "" {
		trustCenter := (&TrustCenterBuilder{Client: tccb.Client}).MustNew(userCtx, t)
		tccb.TrustCenterID = trustCenter.ID
	}

	if tccb.StandardID == "" {
		standard := (&StandardBuilder{Client: tccb.Client}).MustNew(ctx, t)
		tccb.StandardID = standard.ID
	}

	mutation := tccb.Client.DB.TrustCenterCompliance.Create().
		SetTrustCenterID(tccb.TrustCenterID).
		SetStandardID(tccb.StandardID)

	if len(tccb.Tags) > 0 {
		mutation.SetTags(tccb.Tags)
	}

	trustCenterCompliance, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return trustCenterCompliance
}

// TrustCenterEntityBuilder is used to create trustcenter entities
type TrustCenterEntityBuilder struct {
	Client *Client

	// Fields
	Name          string
	URL           *string
	TrustCenterID string
	LogoFileID    *string
}

func (te *TrustCenterEntityBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TrustCenterEntity {
	userCtx := ctx
	ctx = ent.NewContext(ctx, te.Client.DB)
	ctx = graphql.WithResponseContext(ctx, gqlerrors.ErrorPresenter, graphql.DefaultRecover)

	if te.Name == "" {
		te.Name = gofakeit.Company()
	}

	if te.TrustCenterID == "" {
		trustCenter := (&TrustCenterBuilder{Client: te.Client}).MustNew(userCtx, t)
		te.TrustCenterID = trustCenter.ID
	}

	mutation := te.Client.DB.TrustCenterEntity.Create().
		SetName(te.Name).
		SetTrustCenterID(te.TrustCenterID)

	if te.URL != nil {
		mutation.SetURL(*te.URL)
	}

	if te.LogoFileID != nil {
		mutation.SetLogoFileID(*te.LogoFileID)
	}

	trustCenterEntity, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return trustCenterEntity
}

// IntegrationBuilder is used to create integrations
type IntegrationBuilder struct {
	Client *Client

	// Fields
	Name        string
	Description string
	Kind        string
}

// MustNew integration builder is used to create, without authz checks, integrations in the database
func (ib *IntegrationBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Integration {
	ctx = SetContext(ctx, ib.Client.DB)

	if ib.Name == "" {
		ib.Name = "GitHub Integration Test"
	}

	if ib.Description == "" {
		ib.Description = "Test integration for GraphQL tests"
	}

	if ib.Kind == "" {
		ib.Kind = "github"
	}

	mutation := ib.Client.DB.Integration.Create().
		SetName(ib.Name).
		SetDescription(ib.Description).
		SetKind(ib.Kind)

	integration, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return integration
}

// SecretBuilder is used to create secrets (hush)
type SecretBuilder struct {
	Client *Client

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
	ctx = SetContext(ctx, sb.Client.DB)

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

	mutation := sb.Client.DB.Hush.Create().
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
	RequireNoError(t, err)

	return secret
}

// IntegrationCleanup is used to delete integrations
type IntegrationCleanup struct {
	Client *Client
	ID     string
}

// MustDelete deletes the integration
func (ic *IntegrationCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = SetContext(ctx, ic.Client.DB)

	err := ic.Client.DB.Integration.DeleteOneID(ic.ID).Exec(ctx)
	RequireNoError(t, err)
}

// SecretCleanup is used to delete secrets
type SecretCleanup struct {
	Client *Client
	ID     string
}

// MustDelete deletes the secret
func (sc *SecretCleanup) MustDelete(ctx context.Context, t *testing.T) {
	ctx = SetContext(ctx, sc.Client.DB)

	err := sc.Client.DB.Hush.DeleteOneID(sc.ID).Exec(ctx)
	RequireNoError(t, err)
}

// MustNew file builder is used to create, without authz checks, files in the database
func (fb *FileBuilder) MustNew(ctx context.Context, t *testing.T) *ent.File {
	ctx = SetContext(ctx, fb.Client.DB)

	if fb.Name == "" {
		fb.Name = gofakeit.Name()
	}

	url := gofakeit.URL()

	mutation := fb.Client.DB.File.Create().
		SetProvidedFileName(fb.Name).
		SetProvidedFileExtension("csv").
		SetDetectedContentType("application/csv").
		SetURI(url)

	if fb.MD5Hash != "" {
		mutation.SetMd5Hash(fb.MD5Hash)
	}

	file, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return file
}

// MustNew template builder is used to create, without authz checks, templates in the database
func (tb *TemplateBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Template {
	ctx = SetContext(ctx, tb.Client.DB)

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
	mutation := tb.Client.DB.Template.Create().
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

	if tb.TrustCenterID != "" {
		mutation.SetTrustCenterID(tb.TrustCenterID)
	}

	if len(tb.FileIDs) > 0 {
		mutation.AddFileIDs(tb.FileIDs...)
	}

	template, err := mutation.Save(ctx)
	RequireNoError(t, err)

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

	ctx = SetContext(ctx, ab.Client.DB)

	if ab.Name == "" {
		ab.Name = gofakeit.Company() + "-" + ulids.New().String()
	}

	if ab.TemplateID == "" {
		template := (&TemplateBuilder{Client: ab.Client}).MustNew(ctx, t)
		ab.TemplateID = template.ID
	}

	mutation := ab.Client.DB.Assessment.Create().
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
	RequireNoError(t, err)

	return assessment
}

// MustNew assessment response builder creates responses without authz checks
// This uses a questionnaire caller to bypass auth, simulating anonymous user creation
func (arb *AssessmentResponseBuilder) MustNew(ctx context.Context, t *testing.T) *ent.AssessmentResponse {
	ctx = SetContext(ctx, arb.Client.DB)

	var assessment *ent.Assessment

	if arb.AssessmentID == "" {
		assessment = (&AssessmentBuilder{Client: arb.Client}).MustNew(ctx, t)
		arb.AssessmentID = assessment.ID
	}

	if arb.Email == "" {
		arb.Email = gofakeit.Email()
	}

	if arb.OwnerID == "" {
		if assessment == nil {
			var err error
			assessment, err = arb.Client.DB.Assessment.Get(ctx, arb.AssessmentID)
			RequireNoError(t, err)
		}
		arb.OwnerID = assessment.OwnerID
	}

	// Use questionnaire caller to bypass auth checks (simulates anonymous JWT)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = auth.WithCaller(allowCtx, auth.NewQuestionnaireCaller(arb.OwnerID, "", "", ""))

	mutation := arb.Client.DB.AssessmentResponse.Create().
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
	RequireNoError(t, err)

	return response
}

// TrustCenterWatermarkConfigBuilder is used to create trust center watermark configs
type TrustCenterWatermarkConfigBuilder struct {
	Client *Client

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
	Client *Client

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
		tcdb.Category = gofakeit.Word() + "-" + ulids.New().String()
	}

	if tcdb.TrustCenterID == "" {
		trustCenter := (&TrustCenterBuilder{Client: tcdb.Client}).MustNew(userCtx, t)
		tcdb.TrustCenterID = trustCenter.ID
	}

	if len(tcdb.Tags) == 0 {
		tcdb.Tags = []string{"test", "document"}
	}

	(&CustomTypeEnumBuilder{
		Client:     tcdb.Client,
		Name:       tcdb.Category,
		ObjectType: "trust_center_doc",
	}).MustNew(userCtx, t)

	// Create a test PDF file for upload
	pdfFile, err := storage.NewUploadFile(filepath.Join(repoRoot(), "internal", "graphapi", "testdata", "uploads", "hello.pdf"))
	RequireNoError(t, err)

	fileUpload := graphql.Upload{
		File:        pdfFile.RawFile,
		Filename:    pdfFile.OriginalName,
		Size:        pdfFile.Size,
		ContentType: pdfFile.ContentType,
	}

	// Prepare the GraphQL input
	input := testclient.CreateTrustCenterDocInput{
		Title:                  tcdb.Title,
		TrustCenterDocKindName: &tcdb.Category,
		TrustCenterID:          &tcdb.TrustCenterID,
		Tags:                   tcdb.Tags,
	}

	if tcdb.Visibility != "" {
		input.Visibility = &tcdb.Visibility
	}

	// Expect the file upload in the object store
	ExpectUpload(t, tcdb.Client.MockProvider, []graphql.Upload{fileUpload})

	// Create the trust center document using the GraphQL API
	resp, err := tcdb.Client.API.CreateTrustCenterDoc(ctx, input, fileUpload)
	RequireNoError(t, err)

	// Convert the GraphQL response to an ent entity
	// We need to fetch it from the database to get the full ent.TrustCenterDoc
	dbCtx := SetContext(ctx, tcdb.Client.DB)
	trustCenterDoc, err := tcdb.Client.DB.TrustCenterDoc.Get(dbCtx, resp.CreateTrustCenterDoc.TrustCenterDoc.ID)
	RequireNoError(t, err)

	return trustCenterDoc
}

// MustNew trust center watermark config builder is used to create, without authz checks, trust center watermark configs in the database
func (tcwcb *TrustCenterWatermarkConfigBuilder) MustNew(ctx context.Context, t *testing.T, trustCenterID string) *ent.TrustCenterWatermarkConfig {
	ctx = SetContext(ctx, tcwcb.Client.DB)

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

	mutation := tcwcb.Client.DB.TrustCenterWatermarkConfig.Create().
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
	RequireNoError(t, err)

	return watermarkConfig
}

func (td *TagDefinitionBuilder) MustNew(ctx context.Context, t *testing.T) *ent.TagDefinition {
	ctx = SetContext(ctx, td.Client.DB)

	if td.Name == "" {
		// ensure unique name by appending ULID
		td.Name = gofakeit.HipsterWord() + ulids.New().String()
	}

	mutation := td.Client.DB.TagDefinition.Create().
		SetName(td.Name)

	if td.Color != "" {
		mutation.SetColor(td.Color)
	}

	tagDefinition, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return tagDefinition
}

func (td *CustomTypeEnumBuilder) MustNew(ctx context.Context, t *testing.T) *ent.CustomTypeEnum {
	ctx = SetContext(ctx, td.Client.DB)

	if td.Name == "" {
		td.Name = gofakeit.HipsterWord() + "-" + ulids.New().String()
	}

	mutation := td.Client.DB.CustomTypeEnum.Create().
		SetName(td.Name).
		SetObjectType(td.ObjectType)

	// default to "task" only if not explicitly set (including empty for global enums)
	if td.ObjectType == "" && td.Field == "" {
		mutation.SetObjectType("task")
	}

	if td.Field != "" {
		mutation.SetField(td.Field)
	}

	if td.Description != "" {
		mutation.SetDescription(td.Description)
	}

	if td.Color != "" {
		mutation.SetColor(td.Color)
	}

	customTypeEnum, err := mutation.Save(ctx)
	RequireNoError(t, err)

	return customTypeEnum
}

// MustNew asset builder is used to create, without authz checks, assets in the database
func (a *AssetBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Asset {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if a.Name == "" {
		a.Name = gofakeit.AppName()
	}

	asset := a.Client.DB.Asset.Create().
		SetName(a.Name).
		SaveX(ctx)

	return asset
}

// MustNew SLADefinition builder is used to create, without authz checks, SLA definitions in the database.
func (s *SLADefinitionBuilder) MustNew(ctx context.Context, t *testing.T) *ent.SLADefinition {
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	if s.SLADays == 0 {
		s.SLADays = 30
	}

	if s.SecurityLevel == "" {
		s.SecurityLevel = enums.SecurityLevelNone
	}

	sla, err := s.Client.DB.SLADefinition.Create().
		SetSLADays(s.SLADays).
		SetSecurityLevel(s.SecurityLevel).
		Save(ctx)
	if err == nil {
		return sla
	}

	existing, err := s.Client.DB.SLADefinition.Query().
		Where(
			sladefinition.SecurityLevelEQ(s.SecurityLevel),
		).
		Only(ctx)
	if err != nil {
		t.Fatalf("failed to find existing SLA definition: %v", err)
	}

	return existing
}

func (e *EmailTemplateBuilder) MustNew(ctx context.Context, t *testing.T) *ent.EmailTemplate {
	ctx = SetContext(ctx, e.Client.DB)

	if e.Name == "" {
		e.Name = gofakeit.HipsterWord() + " Template"
	}

	if e.Key == "" {
		e.Key = emaildef.BrandedMessageOp.Name()
	}

	if e.TemplateContext == nil {
		e.TemplateContext = &enums.TemplateContextCampaignRecipient
	}

	emailTemplate, err := e.Client.DB.EmailTemplate.Create().
		SetName(e.Name).
		SetKey(e.Key).
		SetTemplateContext(*e.TemplateContext).
		SetDefaults(map[string]any{
			"subject": "Test subject",
			"title":   "Test title",
			"intros":  []any{"Test body"},
		}).
		Save(ctx)
	RequireNoError(t, err)

	return emailTemplate
}

func (p *PlatformBuilder) MustNew(ctx context.Context, t *testing.T) *ent.Platform {
	ctx = SetContext(ctx, p.Client.DB)

	if p.Name == "" {
		p.Name = gofakeit.AppName() + ulids.New().String()
	}

	platform, err := p.Client.DB.Platform.Create().
		SetName(p.Name).
		Save(ctx)
	RequireNoError(t, err)

	return platform
}
