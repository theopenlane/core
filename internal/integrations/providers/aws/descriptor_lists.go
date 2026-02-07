package aws

import "github.com/theopenlane/core/common/integrations/types"

// awsOperations returns the operation descriptors supported by the AWS provider
func awsOperations() []types.OperationDescriptor {
	ops := []types.OperationDescriptor{
		awsHealthOperation(),
	}
	ops = append(ops, awsAuditManagerOperations()...)
	ops = append(ops, awsSecurityHubOperations()...)
	return ops
}

// awsClientDescriptors returns the client descriptors supported by the AWS provider
func awsClientDescriptors() []types.ClientDescriptor {
	var descriptors []types.ClientDescriptor
	descriptors = append(descriptors, awsAuditManagerClientDescriptors()...)
	descriptors = append(descriptors, awsSecurityHubClientDescriptors()...)
	return descriptors
}
