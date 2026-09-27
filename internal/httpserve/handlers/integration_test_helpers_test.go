package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// HelperTestHealthCheck is the config type for the helper test health check operation
type HelperTestHealthCheck struct{}

// githubTestCredential is the credential type stored by the GitHub disconnect test definition
type githubTestCredential struct {
	// Token is the stored token
	Token string `json:"token"`
}

var (
	githubAppDefinitionID   = githubapp.DefinitionID.ID()
	githubTestCredentialRef = types.NewCredentialRef[githubTestCredential]()
	_, _                    = providerkit.OperationSchema[HelperTestHealthCheck]()
)

// withDefinitionRuntime returns a restore function that resets IntegrationsConfig.
// All definitions and gala listeners are registered once in SetupSuite
func (suite *HandlerTestSuite) withDefinitionRuntime(_ *testing.T, _ []registry.Builder) func() {
	originalConfig := suite.h.IntegrationsConfig

	return func() {
		suite.h.IntegrationsConfig = originalConfig
	}
}

// withGitHubAppIntegrationRuntime sets the handler's GitHubApp config for the test
// and returns a restore function that resets it
func (suite *HandlerTestSuite) withGitHubAppIntegrationRuntime(t *testing.T, cfg githubapp.Config) func() {
	t.Helper()

	originalConfig := suite.h.IntegrationsConfig
	suite.h.IntegrationsConfig.GitHubApp = cfg

	return func() {
		suite.h.IntegrationsConfig = originalConfig
	}
}

// githubTestDefinitionBuilder returns a minimal test definition used for disconnect tests.
// The definition has no credentials schema or auth flow; it only needs to be present in
// the registry so the handler can resolve the provider by ID.
func githubTestDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "GitHub",
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         githubTestCredentialRef,
					Name:        "GitHub Test Credential",
					Description: "Credential slot used by the GitHub disconnect test definition.",
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  githubTestCredentialRef.ID(),
					Name:           "GitHub Test Connection",
					Description:    "Test connection used for handler disconnect flows.",
					CredentialRefs: []types.CredentialSlotID{githubTestCredentialRef.ID()},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: githubTestCredentialRef.ID(),
						Description:   "Remove the persisted GitHub test credential and disconnect this installation.",
					},
				},
			},
		}, nil
	})
}

const (
	webhookTestDefinitionID = "def_01K0TESTWBHK0000000000001"
	webhookTestPath         = "/v1/integrations/webhooks/:endpointID"
	webhookTestSecret       = "webhook-test-secret"
)

// WebhookTestHealthCheck is the config type for the webhook test health check operation
type WebhookTestHealthCheck struct{}

// webhookTestAlertEnvelope is the payload type for the test webhook events
type webhookTestAlertEnvelope struct{}

var (
	webhookTestCredentialRef                         = types.NewCredentialSlotID("webhook_test")
	webhookHealthSchema, webhookHealthCheckOperation = providerkit.OperationSchema[WebhookTestHealthCheck]()
	webhookAlertCreatedEvent                         = types.NewWebhookEventRef[webhookTestAlertEnvelope]("alert.created")
)

func webhookTestDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Webhook Test",
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:    webhookTestCredentialRef,
					Name:   "Webhook Test Credential",
					Schema: json.RawMessage(`{"type":"object","properties":{"token":{"type":"string"}}}`),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  webhookTestCredentialRef,
					Name:           "Webhook Test Connection",
					CredentialRefs: []types.CredentialSlotID{webhookTestCredentialRef},
				},
			},
			Webhooks: []types.WebhookRegistration{
				{
					Name: "inbound.events",
					Event: func(req types.WebhookInboundRequest) (types.WebhookReceivedEvent, error) {
						var envelope struct {
							Event      string `json:"event"`
							DeliveryID string `json:"delivery_id"`
						}
						if err := json.Unmarshal(req.Payload, &envelope); err != nil {
							return types.WebhookReceivedEvent{}, err
						}

						if envelope.Event == "" {
							return types.WebhookReceivedEvent{}, nil
						}

						return types.WebhookReceivedEvent{
							Name:       envelope.Event,
							DeliveryID: envelope.DeliveryID,
							Payload:    req.Payload,
						}, nil
					},
					Events: []types.WebhookEventRegistration{
						{
							Name:  webhookAlertCreatedEvent.Name(),
							Topic: types.NewDefinitionRef(definitionID).WebhookEventTopic(webhookAlertCreatedEvent.Name()),
							Handle: func(context.Context, types.WebhookHandleRequest) error {
								return nil
							},
						},
					},
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         webhookHealthCheckOperation.Name(),
					Description:  "Health check",
					Topic:        types.NewDefinitionRef(definitionID).OperationTopic(webhookHealthCheckOperation.Name()),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: webhookHealthSchema,
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"ok":true}`), nil
					},
				},
			},
		}, nil
	})
}

const (
	operationTestDefinitionID       = "def_01K0TESTOPS00000000000001"
	operationTestInlineDefinitionID = "def_01K0TESTOPS00000000000002"
	operationTestPath               = "/v1/integrations/:definitionID/operations"
)

// OperationTestHealthCheck is the config type for the test health check operation
type OperationTestHealthCheck struct{}

// OperationTestRepoSync is the config type for the test repository sync operation
type OperationTestRepoSync struct{}

// OperationTestValidated is the config type for the test validated operation
type OperationTestValidated struct {
	// Target is the required target field
	Target string `json:"target" jsonschema:"required"`
}

var (
	operationTestCredentialRef                      = types.NewCredentialSlotID("op_test")
	opTestHealthSchema, opTestHealthCheckOperation  = providerkit.OperationSchema[OperationTestHealthCheck]()
	opTestRepoSyncSchema, opTestRepoSyncOperation   = providerkit.OperationSchema[OperationTestRepoSync]()
	opTestValidatedSchema, opTestValidatedOperation = providerkit.OperationSchema[OperationTestValidated]()
)

func operationTestDefinitionBuilder(definitionID string, inlineNonHealth bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Operation Test",
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:    operationTestCredentialRef,
					Name:   "Op Test Credential",
					Schema: json.RawMessage(`{"type":"object","properties":{"token":{"type":"string"}}}`),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  operationTestCredentialRef,
					Name:           "Op Test Connection",
					CredentialRefs: []types.CredentialSlotID{operationTestCredentialRef},
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         opTestHealthCheckOperation.Name(),
					Description:  "Validate the test credential",
					Topic:        types.NewDefinitionRef(definitionID).OperationTopic(opTestHealthCheckOperation.Name()),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: opTestHealthSchema,
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"ok":true}`), nil
					},
				},
				{
					Name:         opTestRepoSyncOperation.Name(),
					Description:  "Sync repositories",
					Topic:        types.NewDefinitionRef(definitionID).OperationTopic(opTestRepoSyncOperation.Name()),
					Policy:       types.ExecutionPolicy{Inline: inlineNonHealth},
					ConfigSchema: opTestRepoSyncSchema,
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"synced":true}`), nil
					},
				},
				{
					Name:         opTestValidatedOperation.Name(),
					Description:  "Operation with config schema",
					Topic:        types.NewDefinitionRef(definitionID).OperationTopic(opTestValidatedOperation.Name()),
					ConfigSchema: opTestValidatedSchema,
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"validated":true}`), nil
					},
				},
			},
		}, nil
	})
}

const (
	configTestProviderID           = "def_01K0TESTCFG00000000000001"
	configTestFailHealthProviderID = "def_01K0TESTCFG00000000000002"
	configTestVendorProviderID     = "def_01K0TESTCFG00000000000003"
	configTestVendorFamily         = "configtestvendor"
	configTestProbeProviderID      = "def_01K0TESTCFG00000000000004"
	configTestProbeFailProviderID  = "def_01K0TESTCFG00000000000005"
	configTestProbeOperation       = "ConfigTestProbe"
	configTestUnprobedOperation    = "ConfigTestUnprobed"
)

var errConfigTestProbeFailed = errors.New("probe prerequisites missing")

var configTestCredentialRef = types.NewCredentialSlotID("config_test")

func configTestProbeDefinitionBuilder(definitionID string, probeErr error) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def, err := configTestDefinitionBuilder(definitionID, false)()
		if err != nil {
			return types.Definition{}, err
		}

		def.Operations = append(def.Operations,
			types.OperationRegistration{
				Name:  configTestProbeOperation,
				Topic: types.NewDefinitionRef(definitionID).OperationTopic(configTestProbeOperation),
				Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
					return json.RawMessage(`{}`), nil
				},
				HealthCheck: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
					return nil, probeErr
				},
			},
			types.OperationRegistration{
				Name:  configTestUnprobedOperation,
				Topic: types.NewDefinitionRef(definitionID).OperationTopic(configTestUnprobedOperation),
				Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
					return json.RawMessage(`{}`), nil
				},
			},
		)

		return def, nil
	})
}

func configTestFamilyDefinitionBuilder(definitionID, family string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		def, err := configTestDefinitionBuilder(definitionID, false)()
		if err != nil {
			return types.Definition{}, err
		}

		def.Family = family

		return def, nil
	})
}

func configTestDefinitionBuilder(definitionID string, failHealth bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		healthHandler := func(context.Context, types.OperationRequest) (json.RawMessage, error) {
			if failHealth {
				return nil, errors.New("health failed")
			}

			return json.RawMessage(`{"ok":true}`), nil
		}

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Config Test",
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: json.RawMessage(`{"type":"object","properties":{"filterExpr":{"type":"string"}}}`),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         configTestCredentialRef,
					Name:        "Config Test Credential",
					Description: "Credential slot used by the config test definition.",
					Schema:      json.RawMessage(`{"type":"object","required":["projectId","serviceAccountEmail"],"properties":{"projectId":{"type":"string"},"serviceAccountEmail":{"type":"string"}}}`),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  configTestCredentialRef,
					Name:           "Config Test Connection",
					Description:    "Connect the config test definition using the configured credential payload.",
					CredentialRefs: []types.CredentialSlotID{configTestCredentialRef},
					HealthCheck:    &types.HealthCheckRegistration{Handle: healthHandler},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: configTestCredentialRef,
						Description:   "Remove the persisted config test credential and disconnect this installation.",
					},
				},
			},
		}, nil
	})
}

func userInputOnlyTestDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "User Input Test",
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: json.RawMessage(`{"type":"object","properties":{"filterExpr":{"type":"string"}}}`),
			},
		}, nil
	})
}
