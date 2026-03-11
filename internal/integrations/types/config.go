package types

import (
	"encoding/json"
	"time"

	"github.com/theopenlane/core/common/enums"
)

// RetentionPolicy replaces common/openapi.IntegrationRetentionPolicy and defines
// storage settings for integration payloads
type RetentionPolicy struct {
	// StoreRawPayload indicates whether to store the raw payload
	StoreRawPayload bool `json:"storeRawPayload,omitempty"`
	// PayloadTTL defines how long raw payloads are retained
	PayloadTTL time.Duration `json:"payloadTtl,omitempty"`
}

// IntegrationConfig replaces common/openapi.IntegrationConfig and represents
// the per-org runtime instance config stored as a JSON field in the ent schema
type IntegrationConfig struct {
	// OperationTemplates holds saved operation templates keyed by operation name
	OperationTemplates map[string]OperationTemplate `json:"operationTemplates,omitempty"`
	// EnabledOperations lists which operations are enabled
	EnabledOperations []string `json:"enabledOperations,omitempty"`
	// ClientConfig holds provider-specific client configuration
	ClientConfig json.RawMessage `json:"clientConfig,omitempty"`
	// CollectionStrategy defines how data is collected from the provider
	CollectionStrategy string `json:"collectionStrategy,omitempty"`
	// Schedule defines the integration schedule
	Schedule string `json:"schedule,omitempty"`
	// PollInterval defines how often to poll the provider for new data
	PollInterval time.Duration `json:"pollInterval,omitempty"`
	// MappingOverrides holds user-configurable CEL expression overrides keyed by schema name
	MappingOverrides map[string]MappingOverride `json:"mappingOverrides,omitempty"`
	// RetentionPolicy defines the data retention policy
	RetentionPolicy *RetentionPolicy `json:"retentionPolicy,omitempty"`
	// SCIMProvisionMode controls how SCIM push events are persisted;
	// zero value ("") is treated as USERS by SCIM handlers —
	// only meaningful for SCIM-type integrations
	SCIMProvisionMode enums.SCIMProvisionMode `json:"scimProvisionMode,omitempty"`
}
