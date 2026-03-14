//go:build test

package emailruntime_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/theopenlane/utils/testutils"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/emailruntime"
	"github.com/theopenlane/core/internal/ent/entconfig"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/enttest"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/entdb"
	coreutils "github.com/theopenlane/core/internal/testutils"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/iam/auth"
	fgatest "github.com/theopenlane/iam/fgax/testutils"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/compose"
)

const (
	fgaModelFile = "../../fga/model/model.fga"
)

// TestServiceTestSuite runs the service integration test suite.
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

// ServiceTestSuite handles setup and teardown for emailruntime service integration tests.
type ServiceTestSuite struct {
	suite.Suite
	client *generated.Client
	tf     *testutils.TestFixture
	ofgaTF *fgatest.OpenFGATestFixture
}

// SetupSuite runs once before the suite.
func (suite *ServiceTestSuite) SetupSuite() {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	suite.client = suite.setupClient()
}

// TearDownSuite runs once after the suite.
func (suite *ServiceTestSuite) TearDownSuite() {
	t := suite.T()

	err := suite.client.Close()
	require.NoError(t, err)

	testutils.TeardownFixture(suite.tf)
}

// setupClient creates an ent client backed by a Postgres test container and FGA.
func (suite *ServiceTestSuite) setupClient() *generated.Client {
	t := suite.T()

	suite.ofgaTF = fgatest.NewFGATestcontainer(context.Background(), fgatest.WithModelFile(fgaModelFile))

	ctx := context.Background()

	fgaClient, err := suite.ofgaTF.NewFgaClient(ctx)
	require.NoError(t, err)

	tm, err := coreutils.CreateTokenManager(-15 * time.Minute) //nolint:mnd
	require.NoError(t, err)

	sm := coreutils.CreateSessionManager()
	rc := coreutils.NewRedisClient()

	sessionConfig := sessions.NewSessionConfig(
		sm,
		sessions.WithPersistence(rc),
	)
	sessionConfig.CookieConfig = sessions.DebugOnlyCookieConfig

	entCfg := &entconfig.Config{
		EntityTypes: []string{},
		Modules: entconfig.Modules{
			Enabled: true,
		},
	}

	entitlementsCfg := &entitlements.StripeClient{
		Config: &entitlements.Config{
			Enabled: false,
		},
	}

	opts := []generated.Option{
		generated.Authz(*fgaClient),
		generated.TokenManager(tm),
		generated.SessionConfig(&sessionConfig),
		generated.Emailer(&compose.Config{}),
		generated.EntConfig(entCfg),
		generated.EntitlementManager(entitlementsCfg),
	}

	suite.tf = entdb.NewTestFixture()

	db, err := sql.Open("postgres", suite.tf.URI)
	require.NoError(t, err)

	defer db.Close()

	return enttest.Open(t, dialect.Postgres, suite.tf.URI,
		enttest.WithMigrateOptions(entdb.EnablePostgresOption(db)),
		enttest.WithOptions(opts...),
	)
}

// systemAdminCtx returns a context carrying a system-admin caller.
func systemAdminCtx() context.Context {
	return auth.WithCaller(context.Background(), &auth.Caller{
		SubjectID:          ulids.New().String(),
		Capabilities:       auth.CapSystemAdmin,
		AuthenticationType: auth.JWTAuthentication,
	})
}

// allowCtx wraps ctx with a privacy.Allow decision so all privacy rules are bypassed.
func allowCtx(ctx context.Context) context.Context {
	return privacy.DecisionContext(ctx, privacy.Allow)
}

// createSystemEmailTemplate creates an EmailTemplate directly via the ent client.
func (suite *ServiceTestSuite) createSystemEmailTemplate(ctx context.Context, key, subject, body, text string) *generated.EmailTemplate {
	t := suite.T()

	et, err := suite.client.EmailTemplate.Create().
		SetKey(key).
		SetName(key).
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetSubjectTemplate(subject).
		SetBodyTemplate(body).
		SetTextTemplate(text).
		SetActive(true).
		SetVersion(1).
		SetSystemOwned(true).
		Save(allowCtx(ctx))
	require.NoError(t, err)

	return et
}

// createSystemNotificationTemplate creates a NotificationTemplate directly via the ent client.
func (suite *ServiceTestSuite) createSystemNotificationTemplate(ctx context.Context, key string, emailTemplateID string) *generated.NotificationTemplate {
	t := suite.T()

	builder := suite.client.NotificationTemplate.Create().
		SetKey(key).
		SetName(key).
		SetLocale("en-US").
		SetChannel(enums.ChannelEmail).
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTopicPattern(key).
		SetActive(true).
		SetVersion(1).
		SetSystemOwned(true)

	if emailTemplateID != "" {
		builder = builder.SetEmailTemplateID(emailTemplateID)
	}

	nt, err := builder.Save(allowCtx(ctx))
	require.NoError(t, err)

	return nt
}

// TestComposeFromNotificationTemplate_MissingTemplateReference verifies that compose
// returns ErrMissingTemplateReference when no template key or ID is provided.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_MissingTemplateReference() {
	t := suite.T()
	ctx := systemAdminCtx()
	_, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		To:   []string{"user@example.com"},
		From: "noreply@example.com",
	})

	require.ErrorIs(t, err, emailruntime.ErrMissingTemplateReference)
}

// TestComposeFromNotificationTemplate_MissingRecipient verifies that compose returns
// ErrMissingRecipientAddress when no recipient address is provided.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_MissingRecipient() {
	t := suite.T()
	ctx := systemAdminCtx()
	_, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: "some_key"},
		From:     "noreply@example.com",
	})

	require.ErrorIs(t, err, emailruntime.ErrMissingRecipientAddress)
}

// TestComposeFromNotificationTemplate_MissingSender verifies that compose returns
// ErrMissingSenderAddress when no sender address is provided.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_MissingSender() {
	t := suite.T()
	ctx := systemAdminCtx()
	_, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: "some_key"},
		To:       []string{"user@example.com"},
	})

	require.ErrorIs(t, err, emailruntime.ErrMissingSenderAddress)
}

// TestComposeFromNotificationTemplate_TemplateNotFound verifies that compose returns
// ErrNotificationTemplateNotFound when the requested template key does not exist.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_TemplateNotFound() {
	t := suite.T()
	ctx := systemAdminCtx()
	_, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: "nonexistent_key_xyzzy"},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
	})

	require.ErrorIs(t, err, emailruntime.ErrNotificationTemplateNotFound)
}

// TestComposeFromNotificationTemplate_ByKey verifies that compose resolves a notification
// template by key and produces a valid email message.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_ByKey() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_compose_by_key"

	et := suite.createSystemEmailTemplate(ctx, key,
		"Hello {{.FirstName}}",
		"<p>Hello {{.FirstName}}</p>",
		"Hello {{.FirstName}}",
	)
	suite.createSystemNotificationTemplate(ctx, key, et.ID)

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
		Data:     map[string]any{"FirstName": "Ada"},
	})

	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "Hello Ada", msg.Subject)
	require.Contains(t, msg.HTML, "Hello Ada")
	require.Contains(t, msg.Text, "Hello Ada")
	require.Equal(t, "noreply@example.com", msg.From)
	require.Equal(t, []string{"user@example.com"}, msg.To)
}

// TestComposeFromNotificationTemplate_ByID verifies that compose resolves a notification
// template by ID and produces a valid email message.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_ByID() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_compose_by_id"

	et := suite.createSystemEmailTemplate(ctx, key, "ID subject", "<p>Body</p>", "Body")
	nt := suite.createSystemNotificationTemplate(ctx, key, et.ID)

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{ID: nt.ID},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
	})

	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "ID subject", msg.Subject)
}

// TestComposeFromNotificationTemplate_SynthesizedEmailTemplate verifies that when a
// notification template has no linked email template, the engine synthesizes one from
// the notification template fields and composes successfully.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_SynthesizedEmailTemplate() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_synthesized_email"

	// no email template — pass empty ID
	suite.createSystemNotificationTemplate(ctx, key, "")

	// update the notification template to include subject and body directly
	nt, err := suite.client.NotificationTemplate.Query().
		Where().
		All(allowCtx(ctx))
	require.NoError(t, err)

	var ntID string

	for _, n := range nt {
		if n.Key == key {
			ntID = n.ID

			break
		}
	}

	require.NotEmpty(t, ntID)

	_, err = suite.client.NotificationTemplate.UpdateOneID(ntID).
		SetSubjectTemplate("Synthesized {{.Greeting}}").
		SetBodyTemplate("<p>Synthesized {{.Greeting}}</p>").
		Save(allowCtx(ctx))
	require.NoError(t, err)

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
		Data:     map[string]any{"Greeting": "world"},
	})

	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "Synthesized world", msg.Subject)
	require.Contains(t, msg.HTML, "Synthesized world")
}

// TestComposeFromNotificationTemplate_JsonconfigToolingOnly verifies jsonconfig does not
// block sends; missing keys render empty values at compose time.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_JsonconfigToolingOnly() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_jsonconfig_tooling_only"

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"UserName": map[string]any{"type": "string"},
		},
		"required": []any{"UserName"},
	}

	nt, err := suite.client.NotificationTemplate.Create().
		SetKey(key).
		SetName(key).
		SetLocale("en-US").
		SetChannel(enums.ChannelEmail).
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTopicPattern(key).
		SetSubjectTemplate("Hello {{.UserName}}").
		SetBodyTemplate("<p>Hello {{.UserName}}</p>").
		SetJsonconfig(schema).
		SetActive(true).
		SetVersion(1).
		SetSystemOwned(true).
		Save(allowCtx(ctx))
	require.NoError(t, err)
	require.NotNil(t, nt)

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
		Data:     map[string]any{},
	})

	require.NoError(t, err)
	require.Contains(t, msg.Subject, "Hello")
	require.Contains(t, msg.HTML, "<p>Hello")
}

// TestComposeFromNotificationTemplate_OwnerOnlyExcludesSystemTemplates verifies that
// when OwnerOnly is true, system-owned templates are not returned even when they match
// the requested key.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_OwnerOnlyExcludesSystemTemplates() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_owner_only_exclusion"

	suite.createSystemNotificationTemplate(ctx, key, "")

	_, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template:  emailruntime.TemplateRef{Key: key},
		OwnerID:   ulids.New().String(),
		OwnerOnly: true,
		To:        []string{"user@example.com"},
		From:      "noreply@example.com",
	})

	require.ErrorIs(t, err, emailruntime.ErrNotificationTemplateNotFound)
}

// TestComposeFromNotificationTemplate_OptionalFields verifies that optional fields
// (ReplyTo, Headers) are applied correctly to the composed email message.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_OptionalFields() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_optional_fields"

	nt, err := suite.client.NotificationTemplate.Create().
		SetKey(key).
		SetName(key).
		SetLocale("en-US").
		SetChannel(enums.ChannelEmail).
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTopicPattern(key).
		SetSubjectTemplate("Optional").
		SetActive(true).
		SetVersion(1).
		SetSystemOwned(true).
		Save(allowCtx(ctx))
	require.NoError(t, err)
	require.NotNil(t, nt)

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
		ReplyTo:  "replies@example.com",
		Headers:  map[string]string{"X-Custom": "test-value"},
	})

	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "replies@example.com", msg.ReplyTo)
	require.Equal(t, "test-value", msg.Headers["X-Custom"])
}

// TestComposeFromNotificationTemplate_HTMLBodyWithEmbeddedQuotes verifies that an HTML
// body template containing embedded double-quotes and special characters survives the
// ent jsonb storage roundtrip and renders correctly.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_HTMLBodyWithEmbeddedQuotes() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_html_embedded_quotes"

	// HTML with embedded double quotes, angle brackets, and ampersands — all of which
	// could be mangled if the body were serialized through JSON without correct escaping.
	rawHTML := `<a href="https://example.com?foo=1&bar=2" data-label="click &amp; go">Link &amp; Text</a>`

	et := suite.createSystemEmailTemplate(ctx, key, "Subject", rawHTML, "")
	suite.createSystemNotificationTemplate(ctx, key, et.ID)

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
	})

	require.NoError(t, err)
	require.Contains(t, msg.HTML, `href="https://example.com?foo=1&bar=2"`)
	require.Contains(t, msg.HTML, `data-label="click &amp; go"`)
	require.Contains(t, msg.HTML, `Link &amp; Text`)
}

// TestComposeFromNotificationTemplate_HTMLBodyWithGoTemplateInAttrs verifies that Go
// template expressions embedded inside HTML attributes render correctly after the ent
// storage roundtrip.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_HTMLBodyWithGoTemplateInAttrs() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_html_template_in_attrs"

	// html/template auto-escapes attribute values, so the rendered href will contain the
	// token value passed in Data without raw injection.
	rawHTML := `<a href="https://example.com/verify?token={{.Token}}">Verify</a>`

	et := suite.createSystemEmailTemplate(ctx, key, "Verify your email", rawHTML, "")
	suite.createSystemNotificationTemplate(ctx, key, et.ID)

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
		Data:     map[string]any{"Token": "abc-123"},
	})

	require.NoError(t, err)
	require.Contains(t, msg.HTML, "abc-123")
	require.Contains(t, msg.HTML, "verify?token=")
}

// TestComposeFromNotificationTemplate_JsonconfigRoundtrip verifies that a JSON schema
// stored in the jsonb Jsonconfig column survives a write-read cycle and still validates
// template data correctly.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_JsonconfigRoundtrip() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_jsonconfig_roundtrip"

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"Count": map[string]any{"type": "integer"},
		},
		"required": []any{"Count"},
	}

	et, err := suite.client.EmailTemplate.Create().
		SetKey(key).
		SetName(key).
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetSubjectTemplate("Count is {{.Count}}").
		SetBodyTemplate("<p>Count is {{.Count}}</p>").
		SetJsonconfig(schema).
		SetActive(true).
		SetVersion(1).
		SetSystemOwned(true).
		Save(allowCtx(ctx))
	require.NoError(t, err)

	suite.createSystemNotificationTemplate(ctx, key, et.ID)

	// reload from DB to confirm schema roundtripped correctly
	reloaded, err := suite.client.EmailTemplate.Get(allowCtx(ctx), et.ID)
	require.NoError(t, err)

	reloadedBytes, err := json.Marshal(reloaded.Jsonconfig)
	require.NoError(t, err)

	originalBytes, err := json.Marshal(schema)
	require.NoError(t, err)

	require.JSONEq(t, string(originalBytes), string(reloadedBytes))

	// valid data should compose
	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
		Data:     map[string]any{"Count": 42},
	})
	require.NoError(t, err)
	require.Contains(t, msg.HTML, "42")

	// invalid data type should still render because jsonconfig is for tooling only.
	msg, err = emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
		Data:     map[string]any{"Count": "not-an-int"},
	})
	require.NoError(t, err)
	require.Contains(t, msg.HTML, "not-an-int")
}

// TestComposeFromNotificationTemplate_NestedTemplateData verifies that deeply nested
// map structures in template data render correctly through Go's html/template engine
// after passing through ent's jsonb column as the Jsonconfig.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_NestedTemplateData() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_nested_template_data"

	rawHTML := `<p>Hello {{.User.FirstName}}, org: {{.Org.Name}}</p>`

	et := suite.createSystemEmailTemplate(ctx, key, "Hello {{.User.FirstName}}", rawHTML, "")
	suite.createSystemNotificationTemplate(ctx, key, et.ID)

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
		Data: map[string]any{
			"User": map[string]any{"FirstName": "Ada"},
			"Org":  map[string]any{"Name": "Openlane"},
		},
	})

	require.NoError(t, err)
	require.Contains(t, msg.HTML, "Hello Ada")
	require.Contains(t, msg.HTML, "org: Openlane")
	require.Equal(t, "Hello Ada", msg.Subject)
}

// TestComposeFromNotificationTemplate_MetadataRoundtrip verifies that template Metadata
// stored as jsonb is retrieved and interpreted correctly by the rendering engine (e.g.
// render_mode key controls dispatch to the correct renderer).
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_MetadataRoundtrip() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_metadata_roundtrip"

	metadata := map[string]any{
		emailruntime.MetadataKeyRenderMode.String(): emailruntime.RenderModeRawHTML.String(),
	}

	et, err := suite.client.EmailTemplate.Create().
		SetKey(key).
		SetName(key).
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetSubjectTemplate("Meta subject").
		SetBodyTemplate("<p>Meta body</p>").
		SetMetadata(metadata).
		SetActive(true).
		SetVersion(1).
		SetSystemOwned(true).
		Save(allowCtx(ctx))
	require.NoError(t, err)

	suite.createSystemNotificationTemplate(ctx, key, et.ID)

	// reload from DB to confirm metadata roundtripped correctly
	reloaded, err := suite.client.EmailTemplate.Get(allowCtx(ctx), et.ID)
	require.NoError(t, err)
	require.Equal(t, emailruntime.RenderModeRawHTML.String(), reloaded.Metadata[emailruntime.MetadataKeyRenderMode.String()])

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
	})

	require.NoError(t, err)
	require.Contains(t, msg.HTML, "Meta body")
}

// TestComposeFromNotificationTemplate_BaseTemplateAssembly verifies that a customer body template
// containing a {{define "content"}} block is assembled with a system base template at render time.
// The base template provides the outer layout; the customer block injects into {{block "content" .}}.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_BaseTemplateAssembly() {
	t := suite.T()
	ctx := systemAdminCtx()

	const baseKey = "test_base_layout"
	const customerKey = "test_base_assembly_customer"

	baseBody := `<html><body>` +
		`{{block "content" .}}<p>default</p>{{end}}` +
		`</body></html>`

	suite.createSystemEmailTemplate(ctx, baseKey, "", baseBody, "")

	customerBody := `{{define "content"}}<p>Hello {{.Name}}</p>{{end}}`

	metadata := map[string]any{
		emailruntime.MetadataKeyRenderMode.String():   emailruntime.RenderModeRawHTML.String(),
		emailruntime.MetadataKeyBaseTemplate.String(): baseKey,
	}

	et, err := suite.client.EmailTemplate.Create().
		SetKey(customerKey).
		SetName(customerKey).
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetSubjectTemplate("Hello {{.Name}}").
		SetBodyTemplate(customerBody).
		SetMetadata(metadata).
		SetActive(true).
		SetVersion(1).
		SetSystemOwned(true).
		Save(allowCtx(ctx))
	require.NoError(t, err)

	suite.createSystemNotificationTemplate(ctx, customerKey, et.ID)

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: customerKey},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
		Data:     map[string]any{"Name": "Ada"},
	})

	require.NoError(t, err)
	require.Contains(t, msg.HTML, "<html>")
	require.Contains(t, msg.HTML, "<body>")
	require.Contains(t, msg.HTML, "<p>Hello Ada</p>")
	require.NotContains(t, msg.HTML, "<p>default</p>")
}

// TestComposeFromNotificationTemplate_DynamicAttachments verifies that attachments passed via
// ComposeRequest.Attachments are included in the composed EmailMessage.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_DynamicAttachments() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_dynamic_attachments"

	et := suite.createSystemEmailTemplate(ctx, key, "Subject", "<p>Body</p>", "")
	suite.createSystemNotificationTemplate(ctx, key, et.ID)

	attachment := newman.NewAttachment("report.pdf", []byte("pdf-content"))

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template:    emailruntime.TemplateRef{Key: key},
		To:          []string{"user@example.com"},
		From:        "noreply@example.com",
		Attachments: []*newman.Attachment{attachment},
	})

	require.NoError(t, err)
	require.Len(t, msg.Attachments, 1)
	require.Equal(t, "report.pdf", msg.Attachments[0].Filename)
}

// TestComposeFromNotificationTemplate_StaticAttachments verifies files attached to an
// EmailTemplate are loaded and included as static attachments.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_StaticAttachments() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_static_attachments"

	et := suite.createSystemEmailTemplate(ctx, key, "Subject", "<p>Body</p>", "")
	suite.createSystemNotificationTemplate(ctx, key, et.ID)

	fileRecord, err := suite.client.File.Create().
		SetProvidedFileName("invoice").
		SetProvidedFileExtension("pdf").
		SetDetectedContentType("application/pdf").
		SetDetectedMimeType("application/pdf").
		SetFileContents([]byte("pdf-content")).
		Save(allowCtx(ctx))
	require.NoError(t, err)

	_, err = suite.client.EmailTemplate.UpdateOneID(et.ID).
		AddFileIDs(fileRecord.ID).
		Save(allowCtx(ctx))
	require.NoError(t, err)

	msg, err := emailruntime.ComposeFromNotificationTemplate(ctx, suite.client, emailruntime.ComposeRequest{
		Template: emailruntime.TemplateRef{Key: key},
		To:       []string{"user@example.com"},
		From:     "noreply@example.com",
	})

	require.NoError(t, err)
	require.Len(t, msg.Attachments, 1)
	require.Equal(t, "invoice.pdf", msg.Attachments[0].Filename)
	require.Equal(t, "application/pdf", msg.Attachments[0].ContentType)
	require.Equal(t, []byte("pdf-content"), msg.Attachments[0].Content)
}

// TestComposeFromNotificationTemplate_RevisionBumpOnUpdate verifies that the RevisionMixin hook
// auto-increments the patch version on every update mutation.
func (suite *ServiceTestSuite) TestComposeFromNotificationTemplate_RevisionBumpOnUpdate() {
	t := suite.T()
	ctx := systemAdminCtx()

	const key = "test_revision_bump"

	et := suite.createSystemEmailTemplate(ctx, key, "Original subject", "<p>Original</p>", "")

	require.Equal(t, "v0.0.1", et.Revision)

	updated, err := suite.client.EmailTemplate.UpdateOneID(et.ID).
		SetSubjectTemplate("Updated subject").
		Save(allowCtx(ctx))
	require.NoError(t, err)

	require.Equal(t, "v0.0.2", updated.Revision)
}
