package awsauditmanager

import (
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID             = types.NewDefinitionRef("def_01K0AWSAUDITM0000000000001")
	AuditManagerClient       = types.NewClientRef[*auditmanager.Client]()
	HealthDefaultOperation   = types.NewOperationRef[HealthCheck]("health.default")
	AssessmentsListOperation = types.NewOperationRef[AssessmentsList]("assessments.list")
)

// Slug is the unique identifier for the AWS Audit Manager integration
const Slug = "aws_audit_manager"
