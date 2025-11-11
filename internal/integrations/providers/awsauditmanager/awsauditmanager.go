package awsauditmanager

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeAWSAuditManager identifies the AWS Audit Manager provider
const TypeAWSAuditManager = types.ProviderType("aws_audit_manager")

// Builder returns the AWS Audit Manager provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeAWSAuditManager)
}
