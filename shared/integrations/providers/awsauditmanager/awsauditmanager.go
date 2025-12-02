package awsauditmanager

import (
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/awssts"
	"github.com/theopenlane/shared/integrations/types"
)

// TypeAWSAuditManager identifies the AWS Audit Manager provider
const TypeAWSAuditManager = types.ProviderType("aws_audit_manager")

// Builder returns the AWS Audit Manager provider builder
func Builder() providers.Builder {
	return awssts.Builder(TypeAWSAuditManager, awssts.WithOperations(awsAuditOperations()))
}
