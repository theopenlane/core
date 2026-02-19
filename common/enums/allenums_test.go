package enums_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
)

// gqlEnum is the interface satisfied by all enum types in this package
type gqlEnum interface {
	// Values returns all valid enum values as strings
	Values() []string
	// String returns the string representation of the enum value
	String() string
	// MarshalGQL writes the quoted enum value to the given writer
	MarshalGQL(io.Writer)
}

// enumTestCase holds the data needed to exercise all standard methods on a single enum type
type enumTestCase struct {
	// name is the human-readable enum type name used as the subtest identifier
	name string
	// value is any concrete enum value; drives Values, String, and MarshalGQL coverage
	value gqlEnum
	// unmarshal allocates a zero-value receiver and calls UnmarshalGQL on it
	unmarshal func(any) error
	// parse calls the To* parser function with a known-valid input string
	parse func()
	// errType is the expected error on a bad unmarshal input; nil defaults to ErrInvalidType
	errType error
}

// TestEnumCoverage iterates through all enum types in the package and exercises their core methods using a shared test structure
// this is pretty gross to read but it ensures that every enum type gets coverage on all of the same methods without needing to write a separate test function for each one
func TestEnumCoverage(t *testing.T) {
	cases := []enumTestCase{
		{name: "AssessmentResponseStatus", value: enums.AssessmentResponseStatusNotStarted,
			unmarshal: func(v any) error { var e enums.AssessmentResponseStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToAssessmentResponseStatus("NOT_STARTED") }},
		{name: "AssessmentType", value: enums.AssessmentTypeInternal,
			unmarshal: func(v any) error { var e enums.AssessmentType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToAssessmentType("INTERNAL") }},
		{name: "AssetType", value: enums.AssetTypeTechnology,
			unmarshal: func(v any) error { var e enums.AssetType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToAssetType("TECHNOLOGY") }},
		{name: "AuthProvider", value: enums.AuthProviderCredentials,
			unmarshal: func(v any) error { var e enums.AuthProvider; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToAuthProvider("CREDENTIALS") }},
		{name: "CampaignStatus", value: enums.CampaignStatusDraft,
			unmarshal: func(v any) error { var e enums.CampaignStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToCampaignStatus("DRAFT") }},
		{name: "CampaignType", value: enums.CampaignTypeQuestionnaire,
			unmarshal: func(v any) error { var e enums.CampaignType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToCampaignType("QUESTIONNAIRE") }},
		{name: "Channel", value: enums.ChannelInApp,
			unmarshal: func(v any) error { var e enums.Channel; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToChannel("IN_APP") }},
		{name: "ControlSource", value: enums.ControlSourceFramework,
			unmarshal: func(v any) error { var e enums.ControlSource; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToControlSource("FRAMEWORK") }},
		{name: "ControlStatus", value: enums.ControlStatusPreparing,
			unmarshal: func(v any) error { var e enums.ControlStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToControlStatus("PREPARING") }},
		{name: "ControlType", value: enums.ControlTypePreventative,
			unmarshal: func(v any) error { var e enums.ControlType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToControlType("PREVENTATIVE") }},
		{name: "DNSVerificationStatus", value: enums.DNSVerificationStatusActive,
			unmarshal: func(v any) error { var e enums.DNSVerificationStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToDNSVerificationStatus("ACTIVE") }},
		{name: "DirectoryAccountMFAState", value: enums.DirectoryAccountMFAStateUnknown,
			unmarshal: func(v any) error { var e enums.DirectoryAccountMFAState; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToDirectoryAccountMFAState("UNKNOWN") }},
		{name: "DirectoryAccountStatus", value: enums.DirectoryAccountStatusActive,
			unmarshal: func(v any) error { var e enums.DirectoryAccountStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToDirectoryAccountStatus("ACTIVE") }},
		{name: "DirectoryAccountType", value: enums.DirectoryAccountTypeUser,
			unmarshal: func(v any) error { var e enums.DirectoryAccountType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToDirectoryAccountType("USER") }},
		{name: "DirectoryGroupClassification", value: enums.DirectoryGroupClassificationSecurity,
			unmarshal: func(v any) error { var e enums.DirectoryGroupClassification; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToDirectoryGroupClassification("SECURITY") }},
		{name: "DirectoryGroupStatus", value: enums.DirectoryGroupStatusActive,
			unmarshal: func(v any) error { var e enums.DirectoryGroupStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToDirectoryGroupStatus("ACTIVE") }},
		{name: "DirectoryMembershipRole", value: enums.DirectoryMembershipRoleMember,
			unmarshal: func(v any) error { var e enums.DirectoryMembershipRole; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToDirectoryMembershipRole("MEMBER") }},
		{name: "DirectorySyncRunStatus", value: enums.DirectorySyncRunStatusPending,
			unmarshal: func(v any) error { var e enums.DirectorySyncRunStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToDirectorySyncRunStatus("PENDING") }},
		{name: "DocumentStatus", value: enums.DocumentPublished,
			unmarshal: func(v any) error { var e enums.DocumentStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToDocumentStatus("PUBLISHED") }},
		{name: "DocumentType", value: enums.RootTemplate,
			unmarshal: func(v any) error { var e enums.DocumentType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToDocumentType("ROOTTEMPLATE") }},
		{name: "EntityStatus", value: enums.EntityStatusDraft,
			unmarshal: func(v any) error { var e enums.EntityStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToEntityStatus("DRAFT") }},
		{name: "EvidenceStatus", value: enums.EvidenceStatusSubmitted,
			unmarshal: func(v any) error { var e enums.EvidenceStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToEvidenceStatus("SUBMITTED") }},
		{name: "ExportFormat", value: enums.ExportFormatCsv,
			unmarshal: func(v any) error { var e enums.ExportFormat; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToExportFormat("CSV") }},
		{name: "ExportStatus", value: enums.ExportStatusPending,
			unmarshal: func(v any) error { var e enums.ExportStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToExportStatus("PENDING") }},
		{name: "ExportType", value: enums.ExportTypeAsset,
			unmarshal: func(v any) error { var e enums.ExportType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToExportType("ASSET") }},
		{name: "Font", value: enums.FontCourier,
			unmarshal: func(v any) error { var e enums.Font; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToFont("COURIER") }},
		{name: "Frequency", value: enums.FrequencyYearly,
			unmarshal: func(v any) error { var e enums.Frequency; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToFrequency("YEARLY") }},
		{name: "IdentityHolderType", value: enums.IdentityHolderTypeEmployee,
			unmarshal: func(v any) error { var e enums.IdentityHolderType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToIdentityHolderType("EMPLOYEE") }},
		{name: "ImpersonationAction", value: enums.ImpersonationActionStart,
			unmarshal: func(v any) error { var e enums.ImpersonationAction; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToImpersonationAction("START") }},
		{name: "ImpersonationType", value: enums.ImpersonationTypeSupport,
			unmarshal: func(v any) error { var e enums.ImpersonationType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToImpersonationType("SUPPORT") }},
		{name: "IntegrationOperationKind", value: enums.IntegrationOperationKindSync,
			unmarshal: func(v any) error { var e enums.IntegrationOperationKind; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToIntegrationOperationKind("SYNC") }},
		{name: "IntegrationRunStatus", value: enums.IntegrationRunStatusPending,
			unmarshal: func(v any) error { var e enums.IntegrationRunStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToIntegrationRunStatus("PENDING") }},
		{name: "IntegrationRunType", value: enums.IntegrationRunTypeManual,
			unmarshal: func(v any) error { var e enums.IntegrationRunType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToIntegrationRunType("MANUAL") }},
		{name: "IntegrationWebhookStatus", value: enums.IntegrationWebhookStatusActive,
			unmarshal: func(v any) error { var e enums.IntegrationWebhookStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToIntegrationWebhookStatus("ACTIVE") }},
		{name: "InviteStatus", value: enums.InvitationSent,
			unmarshal: func(v any) error { var e enums.InviteStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToInviteStatus("INVITATION_SENT") }},
		{name: "JobCadenceFrequency", value: enums.JobCadenceFrequencyDaily,
			unmarshal: func(v any) error { var e enums.JobCadenceFrequency; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToJobCadenceFrequency("DAILY") }},
		{name: "JobEnvironment", value: enums.JobEnvironmentOpenlane,
			unmarshal: func(v any) error { var e enums.JobEnvironment; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToJobEnvironment("OPENLANE") }},
		{name: "JobExecutionStatus", value: enums.JobExecutionStatusCanceled,
			unmarshal: func(v any) error { var e enums.JobExecutionStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToJobExecutionStatus("CANCELED") }},
		{name: "JobPlatformType", value: enums.JobPlatformTypeGo,
			unmarshal: func(v any) error { var e enums.JobPlatformType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToJobPlatformType("GO") }},
		{name: "JobRunnerStatus", value: enums.JobRunnerStatusOnline,
			unmarshal: func(v any) error { var e enums.JobRunnerStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToJobRunnerStatus("ONLINE") }},
		{name: "JobWeekday", value: enums.JobWeekdaySunday,
			unmarshal: func(v any) error { var e enums.JobWeekday; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToJobWeekday("SUNDAY") }},
		{name: "JoinPolicy", value: enums.JoinPolicyOpen,
			unmarshal: func(v any) error { var e enums.JoinPolicy; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToGroupJoinPolicy("OPEN") }},
		{name: "MappingSource", value: enums.MappingSourceManual,
			unmarshal: func(v any) error { var e enums.MappingSource; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToMappingSource("MANUAL") }},
		{name: "MappingType", value: enums.MappingTypeEqual,
			unmarshal: func(v any) error { var e enums.MappingType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToMappingType("EQUAL") }},
		{name: "NotificationCadence", value: enums.NotificationCadenceImmediate,
			unmarshal: func(v any) error { var e enums.NotificationCadence; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToNotificationCadence("IMMEDIATE") }},
		{name: "NotificationChannelStatus", value: enums.NotificationChannelStatusEnabled,
			unmarshal: func(v any) error { var e enums.NotificationChannelStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToNotificationChannelStatus("ENABLED") }},
		{name: "NotificationTemplateFormat", value: enums.NotificationTemplateFormatText,
			unmarshal: func(v any) error { var e enums.NotificationTemplateFormat; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToNotificationTemplateFormat("TEXT") }},
		{name: "NotificationTopic", value: enums.NotificationTopicTaskAssignment,
			unmarshal: func(v any) error { var e enums.NotificationTopic; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToNotificationTopic("TASK_ASSIGNMENT") }},
		{name: "NotificationType", value: enums.NotificationTypeOrganization,
			unmarshal: func(v any) error { var e enums.NotificationType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToNotificationType("ORGANIZATION") }},
		{name: "ObjectiveStatus", value: enums.ObjectiveDraftStatus,
			unmarshal: func(v any) error { var e enums.ObjectiveStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToObjectiveStatus("DRAFT") }},
		{name: "Permission", value: enums.Editor,
			unmarshal: func(v any) error { var e enums.Permission; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToPermission("EDITOR") }},
		{name: "PlatformStatus", value: enums.PlatformStatusActive,
			unmarshal: func(v any) error { var e enums.PlatformStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToPlatformStatus("ACTIVE") }},
		{name: "Priority", value: enums.PriorityLow,
			unmarshal: func(v any) error { var e enums.Priority; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToPriority("LOW") }},
		{name: "ProgramStatus", value: enums.ProgramStatusNotStarted,
			unmarshal: func(v any) error { var e enums.ProgramStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToProgramStatus("NOT_STARTED") }},
		{name: "ProgramType", value: enums.ProgramTypeFramework,
			unmarshal: func(v any) error { var e enums.ProgramType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToProgramType("FRAMEWORK") }},
		{name: "Region", value: enums.Amer,
			unmarshal: func(v any) error { var e enums.Region; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToRegion("AMER") }},
		{name: "RiskImpact", value: enums.RiskImpactLow,
			unmarshal: func(v any) error { var e enums.RiskImpact; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToRiskImpact("LOW") }},
		{name: "RiskLikelihood", value: enums.RiskLikelihoodLow,
			unmarshal: func(v any) error { var e enums.RiskLikelihood; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToRiskLikelihood("UNLIKELY") }},
		{name: "RiskStatus", value: enums.RiskOpen,
			unmarshal: func(v any) error { var e enums.RiskStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToRiskStatus("OPEN") }},
		{name: "Role", value: enums.RoleAdmin,
			unmarshal: func(v any) error { var e enums.Role; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToRole("ADMIN") }},
		{name: "ScanStatus", value: enums.ScanStatusPending,
			unmarshal: func(v any) error { var e enums.ScanStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToScanStatus("PENDING") }},
		{name: "ScanType", value: enums.ScanTypeDomain,
			unmarshal: func(v any) error { var e enums.ScanType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToScanType("DOMAIN") }},
		{name: "ScheduledJobRunStatus", value: enums.ScheduledJobRunStatusPending,
			unmarshal: func(v any) error { var e enums.ScheduledJobRunStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToScheduledJobRunStatus("PENDING") }},
		{name: "SourceType", value: enums.SourceTypeManual,
			unmarshal: func(v any) error { var e enums.SourceType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToSourceType("MANUAL") }},
		{name: "SSOProvider", value: enums.SSOProviderOkta,
			unmarshal: func(v any) error { var e enums.SSOProvider; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToSSOProvider("OKTA") }},
		{name: "SSLVerificationStatus", value: enums.SSLVerificationStatusActive,
			unmarshal: func(v any) error { var e enums.SSLVerificationStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToSSLVerificationStatus("ACTIVE") }},
		{name: "StandardStatus", value: enums.StandardActive,
			unmarshal: func(v any) error { var e enums.StandardStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToStandardStatus("ACTIVE") }},
		{name: "TaskStatus", value: enums.TaskStatusOpen,
			unmarshal: func(v any) error { var e enums.TaskStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToTaskStatus("OPEN") }},
		{name: "TemplateKind", value: enums.TemplateKindQuestionnaire,
			unmarshal: func(v any) error { var e enums.TemplateKind; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToTemplateKind("QUESTIONNAIRE") }},
		{name: "Tier", value: enums.TierFree,
			unmarshal: func(v any) error { var e enums.Tier; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToTier("FREE") }},
		{name: "TrustCenterDocumentVisibility", value: enums.TrustCenterDocumentVisibilityPubliclyVisible,
			unmarshal: func(v any) error { var e enums.TrustCenterDocumentVisibility; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToTrustCenterDocumentVisibility("PUBLICLY_VISIBLE") }},
		{name: "TrustCenterEnvironment", value: enums.TrustCenterEnvironmentLive,
			unmarshal: func(v any) error { var e enums.TrustCenterEnvironment; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToTrustCenterEnvironment("LIVE") }},
		{name: "TrustCenterNDARequestAccessLevel", value: enums.TrustCenterNDARequestAccessLevelFull,
			unmarshal: func(v any) error { var e enums.TrustCenterNDARequestAccessLevel; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToTrustCenterNDARequestAccessLevel("FULL") }},
		{name: "TrustCenterNDARequestStatus", value: enums.TrustCenterNDARequestStatusRequested,
			unmarshal: func(v any) error { var e enums.TrustCenterNDARequestStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToTrustCenterNDARequestStatus("REQUESTED") }},
		{name: "TrustCenterPreviewStatus", value: enums.TrustCenterPreviewStatusProvisioning,
			unmarshal: func(v any) error { var e enums.TrustCenterPreviewStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToTrustCenterPreviewStatus("PROVISIONING") }},
		{name: "TrustCenterThemeMode", value: enums.TrustCenterThemeModeEasy,
			unmarshal: func(v any) error { var e enums.TrustCenterThemeMode; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToTrustCenterThemeMode("EASY") }},
		{name: "UserStatus", value: enums.UserStatusActive,
			unmarshal: func(v any) error { var e enums.UserStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToUserStatus("ACTIVE") }},
		{name: "Visibility", value: enums.VisibilityPublic,
			unmarshal: func(v any) error { var e enums.Visibility; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToGroupVisibility("PUBLIC") }},
		{name: "WatermarkStatus", value: enums.WatermarkStatusPending,
			unmarshal: func(v any) error { var e enums.WatermarkStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWatermarkStatus("PENDING") }},

		// Workflow enums
		{name: "WorkflowActionType", value: enums.WorkflowActionTypeApproval,
			unmarshal: func(v any) error { var e enums.WorkflowActionType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWorkflowActionType("REQUEST_APPROVAL") }},
		{name: "WorkflowApprovalSubmissionMode", value: enums.WorkflowApprovalSubmissionModeManualSubmit,
			unmarshal: func(v any) error { var e enums.WorkflowApprovalSubmissionMode; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWorkflowApprovalSubmissionMode("MANUAL_SUBMIT") }},
		{name: "WorkflowApprovalTiming", value: enums.WorkflowApprovalTimingPreCommit,
			unmarshal: func(v any) error { var e enums.WorkflowApprovalTiming; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWorkflowApprovalTiming("PRE_COMMIT") }},
		{name: "WorkflowAssignmentStatus", value: enums.WorkflowAssignmentStatusPending,
			unmarshal: func(v any) error { var e enums.WorkflowAssignmentStatus; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWorkflowAssignmentStatus("PENDING") }},
		{name: "WorkflowEventType", value: enums.WorkflowEventTypeAction,
			unmarshal: func(v any) error { var e enums.WorkflowEventType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWorkflowEventType("ACTION") }},
		{name: "WorkflowInstanceState", value: enums.WorkflowInstanceStateRunning,
			unmarshal: func(v any) error { var e enums.WorkflowInstanceState; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWorkflowInstanceState("RUNNING") }},
		{name: "WorkflowKind", value: enums.WorkflowKindApproval,
			unmarshal: func(v any) error { var e enums.WorkflowKind; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWorkflowKind("APPROVAL") }},
		{name: "WorkflowProposalState", value: enums.WorkflowProposalStateDraft,
			unmarshal: func(v any) error { var e enums.WorkflowProposalState; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWorkflowProposalState("DRAFT") }},
		{name: "WorkflowTargetType", value: enums.WorkflowTargetTypeUser,
			unmarshal: func(v any) error { var e enums.WorkflowTargetType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWorkflowTargetType("USER") }},

		// WorkflowObjectType has a custom UnmarshalGQL with a different error type
		{name: "WorkflowObjectType", value: enums.WorkflowObjectTypeActionPlan,
			unmarshal: func(v any) error { var e enums.WorkflowObjectType; return e.UnmarshalGQL(v) },
			parse:     func() { enums.ToWorkflowObjectType("ActionPlan") },
			errType:   enums.ErrWrongTypeWorkflowObjectType},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotEmpty(t, tc.value.Values())

			str := tc.value.String()
			assert.NotEmpty(t, str)

			var buf bytes.Buffer
			tc.value.MarshalGQL(&buf)
			assert.Equal(t, `"`+str+`"`, buf.String())

			require.NoError(t, tc.unmarshal(str))

			errType := tc.errType
			if errType == nil {
				errType = enums.ErrInvalidType
			}

			require.ErrorIs(t, tc.unmarshal(42), errType)

			tc.parse()
		})
	}
}

// TestToTimeWeekday covers the JobWeekday to time.Weekday conversion
func TestToTimeWeekday(t *testing.T) {
	tests := []struct {
		input    enums.JobWeekday
		expected time.Weekday
	}{
		{enums.JobWeekdaySunday, time.Sunday},
		{enums.JobWeekdayMonday, time.Monday},
		{enums.JobWeekdayTuesday, time.Tuesday},
		{enums.JobWeekdayWednesday, time.Wednesday},
		{enums.JobWeekdayThursday, time.Thursday},
		{enums.JobWeekdayFriday, time.Friday},
		{enums.JobWeekdaySaturday, time.Saturday},
		{enums.JobWeekday("INVALID"), time.Weekday(10)},
	}

	for _, tc := range tests {
		t.Run(string(tc.input), func(t *testing.T) {
			assert.Equal(t, tc.expected, enums.ToTimeWeekday(tc.input))
		})
	}
}

// TestWorkflowObjectTypeParse covers all branches of the generated ToWorkflowObjectType switch
func TestWorkflowObjectTypeParse(t *testing.T) {
	tests := []struct {
		input    string
		expected *enums.WorkflowObjectType
	}{
		{"ActionPlan", &enums.WorkflowObjectTypeActionPlan},
		{"Campaign", &enums.WorkflowObjectTypeCampaign},
		{"CampaignTarget", &enums.WorkflowObjectTypeCampaignTarget},
		{"Control", &enums.WorkflowObjectTypeControl},
		{"Evidence", &enums.WorkflowObjectTypeEvidence},
		{"IdentityHolder", &enums.WorkflowObjectTypeIdentityHolder},
		{"InternalPolicy", &enums.WorkflowObjectTypeInternalPolicy},
		{"Platform", &enums.WorkflowObjectTypePlatform},
		{"Procedure", &enums.WorkflowObjectTypeProcedure},
		{"Subcontrol", &enums.WorkflowObjectTypeSubcontrol},
		{"nonexistent", nil},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := enums.ToWorkflowObjectType(tc.input)
			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tc.expected, *result)
			}
		})
	}
}

// TestFontToFontStr covers the Font.ToFontStr switch statement
func TestFontToFontStr(t *testing.T) {
	tests := []struct {
		font     enums.Font
		expected string
	}{
		{enums.FontCourier, "Courier"},
		{enums.FontCourierBold, "Courier-Bold"},
		{enums.FontCourierBoldOblique, "Courier-BoldOblique"},
		{enums.FontCourierOblique, "Courier-Oblique"},
		{enums.FontHelvetica, "Helvetica"},
		{enums.FontHelveticaBold, "Helvetica-Bold"},
		{enums.FontHelveticaBoldOblique, "Helvetica-BoldOblique"},
		{enums.FontHelveticaOblique, "Helvetica-Oblique"},
		{enums.FontSymbol, "Symbol"},
		{enums.FontTimesBold, "Times-Bold"},
		{enums.FontTimesBoldItalic, "Times-BoldItalic"},
		{enums.FontTimesItalic, "Times-Italic"},
		{enums.FontTimesRoman, "Times-Roman"},
		{enums.FontInvalid, ""},
	}

	for _, tc := range tests {
		t.Run(string(tc.font), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.font.ToFontStr())
		})
	}
}
