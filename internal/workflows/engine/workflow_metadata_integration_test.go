//go:build test

package engine_test

import (
	"encoding/json"

	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/graphapi"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	slackdef "github.com/theopenlane/core/internal/integrations/definitions/slack"
)

type metadataProvider struct {
	Provider      string                 `json:"provider"`
	Available     bool                   `json:"available"`
	Installations []metadataInstallation `json:"installations"`
	Operations    []metadataOperation    `json:"operations"`
}

type metadataInstallation struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type metadataOperation struct {
	Name      string             `json:"name"`
	Templates []metadataTemplate `json:"templates"`
}

type metadataTemplate struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// TestWorkflowMetadataExtensions_AvailableWithIntegrationAndTemplates verifies
// that the workflow metadata extensions surface correctly reflects org-scoped
// availability: a provider is marked available only when the calling org has
// both a connected installation and at least one active workflow-action template
// whose topic_pattern matches an operation
func (s *WorkflowEngineTestSuite) TestWorkflowMetadataExtensions_AvailableWithIntegrationAndTemplates() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	// Create a connected Slack integration for this org
	_, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Slack for Metadata Test").
		SetDefinitionID(slackdef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		Save(seedCtx)
	s.Require().NoError(err)

	// Create a connected Email integration for this org
	_, err = s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Email for Metadata Test").
		SetDefinitionID(emaildef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		Save(seedCtx)
	s.Require().NoError(err)

	// Create a notification template matching the Slack message.send operation
	slackTemplateKey := "metadata-slack-" + ulid.Make().String()

	_, err = s.client.NotificationTemplate.Create().
		SetOwnerID(orgID).
		SetKey(slackTemplateKey).
		SetName("Slack Alert Template").
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTemplateContext(enums.TemplateContextWorkflowAction).
		SetTopicPattern(slackdef.MessageSendOp.Name()).
		SetDefaults(map[string]any{
			"channel": "C-METADATA",
			"text":    "alert from workflow",
		}).
		SetActive(true).
		Save(seedCtx)
	s.Require().NoError(err)

	defer func() {
		_, _ = s.client.NotificationTemplate.Delete().Where().Exec(seedCtx)
	}()

	extensions := graphapi.WorkflowMetadataExtensions(userCtx, s.integrationsRT, s.client)
	s.Require().NotNil(extensions)

	providers := unmarshalMetadataProviders(s, extensions)
	s.Require().NotEmpty(providers, "should have at least one provider")

	// Slack has both a connected installation AND a matching template → Available
	slack := findProvider(s, providers, slackdef.DefinitionID.ID())
	s.True(slack.Available, "Slack should be available (has installation + template)")
	s.Len(slack.Installations, 1)
	s.Equal("Slack for Metadata Test", slack.Installations[0].Name)

	msgSendOp, opFound := lo.Find(slack.Operations, func(op metadataOperation) bool {
		return op.Name == slackdef.MessageSendOp.Name()
	})
	s.Require().True(opFound, "MessageSendOperation should be in operations")
	s.Require().Len(msgSendOp.Templates, 1)
	s.Equal(slackTemplateKey, msgSendOp.Templates[0].Key)
	s.Equal("Slack Alert Template", msgSendOp.Templates[0].Name)

	// Email has an installation but NO workflow action template → not available
	email := findProvider(s, providers, emaildef.DefinitionID.ID())
	s.False(email.Available, "Email should NOT be available (has installation but no workflow action template)")
	s.Len(email.Installations, 1)
	s.Equal("Email for Metadata Test", email.Installations[0].Name)
}

// TestWorkflowMetadataExtensions_DisconnectedIntegrationNotAvailable verifies
// that a disabled integration does not make the provider available even when
// matching templates exist
func (s *WorkflowEngineTestSuite) TestWorkflowMetadataExtensions_DisconnectedIntegrationNotAvailable() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	// Create a disabled Slack integration
	_, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Disabled Slack").
		SetDefinitionID(slackdef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusDisabled).
		Save(seedCtx)
	s.Require().NoError(err)

	// Create a matching template
	_, err = s.client.NotificationTemplate.Create().
		SetOwnerID(orgID).
		SetKey("metadata-disconnected-" + ulid.Make().String()).
		SetName("Slack Template for Disabled").
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTemplateContext(enums.TemplateContextWorkflowAction).
		SetTopicPattern(slackdef.MessageSendOp.Name()).
		SetDefaults(map[string]any{"text": "unreachable"}).
		SetActive(true).
		Save(seedCtx)
	s.Require().NoError(err)

	defer func() {
		_, _ = s.client.NotificationTemplate.Delete().Where().Exec(seedCtx)
	}()

	extensions := graphapi.WorkflowMetadataExtensions(userCtx, s.integrationsRT, s.client)
	providers := unmarshalMetadataProviders(s, extensions)

	slack := findProvider(s, providers, slackdef.DefinitionID.ID())
	s.False(slack.Available, "disabled integration should not make provider available")
	s.Empty(slack.Installations, "disabled integration should not appear in installations")
}

// TestWorkflowMetadataExtensions_ReconnectedIntegrationBecomesAvailable verifies
// that when a new connected installation replaces a disabled one, the provider
// becomes available again without any change to the workflow definition
func (s *WorkflowEngineTestSuite) TestWorkflowMetadataExtensions_ReconnectedIntegrationBecomesAvailable() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	// Start with a disabled integration
	_, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Old Slack (disabled)").
		SetDefinitionID(slackdef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusDisabled).
		Save(seedCtx)
	s.Require().NoError(err)

	// Create a matching template
	_, err = s.client.NotificationTemplate.Create().
		SetOwnerID(orgID).
		SetKey("metadata-reconnect-" + ulid.Make().String()).
		SetName("Slack Reconnect Template").
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTemplateContext(enums.TemplateContextWorkflowAction).
		SetTopicPattern(slackdef.MessageSendOp.Name()).
		SetDefaults(map[string]any{"text": "reconnected"}).
		SetActive(true).
		Save(seedCtx)
	s.Require().NoError(err)

	defer func() {
		_, _ = s.client.NotificationTemplate.Delete().Where().Exec(seedCtx)
	}()

	// Before reconnect: not available
	extensions := graphapi.WorkflowMetadataExtensions(userCtx, s.integrationsRT, s.client)
	providers := unmarshalMetadataProviders(s, extensions)
	slack := findProvider(s, providers, slackdef.DefinitionID.ID())
	s.False(slack.Available, "should not be available before reconnection")

	// Reconnect: create a new connected installation for the same definition
	_, err = s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("New Slack (connected)").
		SetDefinitionID(slackdef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		Save(seedCtx)
	s.Require().NoError(err)

	// After reconnect: available, new installation appears
	extensions = graphapi.WorkflowMetadataExtensions(userCtx, s.integrationsRT, s.client)
	providers = unmarshalMetadataProviders(s, extensions)
	slack = findProvider(s, providers, slackdef.DefinitionID.ID())
	s.True(slack.Available, "should be available after reconnection")
	s.Len(slack.Installations, 1, "only the connected installation should appear")
	s.Equal("New Slack (connected)", slack.Installations[0].Name)
}

// unmarshalMetadataProviders extracts the providers list from the extensions map
func unmarshalMetadataProviders(s *WorkflowEngineTestSuite, extensions map[string]any) []metadataProvider {
	s.T().Helper()

	raw, err := json.Marshal(extensions)
	s.Require().NoError(err)

	var doc struct {
		Integrations struct {
			Providers []metadataProvider `json:"providers"`
		} `json:"integrations"`
	}
	s.Require().NoError(json.Unmarshal(raw, &doc))

	return doc.Integrations.Providers
}

// findProvider locates a provider by definition ID, failing the test if not found
func findProvider(s *WorkflowEngineTestSuite, providers []metadataProvider, definitionID string) metadataProvider {
	s.T().Helper()

	p, found := lo.Find(providers, func(p metadataProvider) bool {
		return p.Provider == definitionID
	})
	s.Require().True(found, "provider %s not found in metadata", definitionID)

	return p
}
