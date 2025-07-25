// Code generated by ent, DO NOT EDIT.

package subcontrol

import (
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/pkg/enums"
)

const (
	// Label holds the string label denoting the subcontrol type in the database.
	Label = "subcontrol"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldCreatedBy holds the string denoting the created_by field in the database.
	FieldCreatedBy = "created_by"
	// FieldUpdatedBy holds the string denoting the updated_by field in the database.
	FieldUpdatedBy = "updated_by"
	// FieldDeletedAt holds the string denoting the deleted_at field in the database.
	FieldDeletedAt = "deleted_at"
	// FieldDeletedBy holds the string denoting the deleted_by field in the database.
	FieldDeletedBy = "deleted_by"
	// FieldDisplayID holds the string denoting the display_id field in the database.
	FieldDisplayID = "display_id"
	// FieldTags holds the string denoting the tags field in the database.
	FieldTags = "tags"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldReferenceID holds the string denoting the reference_id field in the database.
	FieldReferenceID = "reference_id"
	// FieldAuditorReferenceID holds the string denoting the auditor_reference_id field in the database.
	FieldAuditorReferenceID = "auditor_reference_id"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldSource holds the string denoting the source field in the database.
	FieldSource = "source"
	// FieldReferenceFramework holds the string denoting the reference_framework field in the database.
	FieldReferenceFramework = "reference_framework"
	// FieldControlType holds the string denoting the control_type field in the database.
	FieldControlType = "control_type"
	// FieldCategory holds the string denoting the category field in the database.
	FieldCategory = "category"
	// FieldCategoryID holds the string denoting the category_id field in the database.
	FieldCategoryID = "category_id"
	// FieldSubcategory holds the string denoting the subcategory field in the database.
	FieldSubcategory = "subcategory"
	// FieldMappedCategories holds the string denoting the mapped_categories field in the database.
	FieldMappedCategories = "mapped_categories"
	// FieldAssessmentObjectives holds the string denoting the assessment_objectives field in the database.
	FieldAssessmentObjectives = "assessment_objectives"
	// FieldAssessmentMethods holds the string denoting the assessment_methods field in the database.
	FieldAssessmentMethods = "assessment_methods"
	// FieldControlQuestions holds the string denoting the control_questions field in the database.
	FieldControlQuestions = "control_questions"
	// FieldImplementationGuidance holds the string denoting the implementation_guidance field in the database.
	FieldImplementationGuidance = "implementation_guidance"
	// FieldExampleEvidence holds the string denoting the example_evidence field in the database.
	FieldExampleEvidence = "example_evidence"
	// FieldReferences holds the string denoting the references field in the database.
	FieldReferences = "references"
	// FieldControlOwnerID holds the string denoting the control_owner_id field in the database.
	FieldControlOwnerID = "control_owner_id"
	// FieldDelegateID holds the string denoting the delegate_id field in the database.
	FieldDelegateID = "delegate_id"
	// FieldOwnerID holds the string denoting the owner_id field in the database.
	FieldOwnerID = "owner_id"
	// FieldRefCode holds the string denoting the ref_code field in the database.
	FieldRefCode = "ref_code"
	// FieldControlID holds the string denoting the control_id field in the database.
	FieldControlID = "control_id"
	// EdgeEvidence holds the string denoting the evidence edge name in mutations.
	EdgeEvidence = "evidence"
	// EdgeControlObjectives holds the string denoting the control_objectives edge name in mutations.
	EdgeControlObjectives = "control_objectives"
	// EdgeTasks holds the string denoting the tasks edge name in mutations.
	EdgeTasks = "tasks"
	// EdgeNarratives holds the string denoting the narratives edge name in mutations.
	EdgeNarratives = "narratives"
	// EdgeRisks holds the string denoting the risks edge name in mutations.
	EdgeRisks = "risks"
	// EdgeActionPlans holds the string denoting the action_plans edge name in mutations.
	EdgeActionPlans = "action_plans"
	// EdgeProcedures holds the string denoting the procedures edge name in mutations.
	EdgeProcedures = "procedures"
	// EdgeInternalPolicies holds the string denoting the internal_policies edge name in mutations.
	EdgeInternalPolicies = "internal_policies"
	// EdgeControlOwner holds the string denoting the control_owner edge name in mutations.
	EdgeControlOwner = "control_owner"
	// EdgeDelegate holds the string denoting the delegate edge name in mutations.
	EdgeDelegate = "delegate"
	// EdgeOwner holds the string denoting the owner edge name in mutations.
	EdgeOwner = "owner"
	// EdgeControl holds the string denoting the control edge name in mutations.
	EdgeControl = "control"
	// EdgeControlImplementations holds the string denoting the control_implementations edge name in mutations.
	EdgeControlImplementations = "control_implementations"
	// EdgeScheduledJobs holds the string denoting the scheduled_jobs edge name in mutations.
	EdgeScheduledJobs = "scheduled_jobs"
	// EdgeMappedToSubcontrols holds the string denoting the mapped_to_subcontrols edge name in mutations.
	EdgeMappedToSubcontrols = "mapped_to_subcontrols"
	// EdgeMappedFromSubcontrols holds the string denoting the mapped_from_subcontrols edge name in mutations.
	EdgeMappedFromSubcontrols = "mapped_from_subcontrols"
	// Table holds the table name of the subcontrol in the database.
	Table = "subcontrols"
	// EvidenceTable is the table that holds the evidence relation/edge. The primary key declared below.
	EvidenceTable = "evidence_subcontrols"
	// EvidenceInverseTable is the table name for the Evidence entity.
	// It exists in this package in order to avoid circular dependency with the "evidence" package.
	EvidenceInverseTable = "evidences"
	// ControlObjectivesTable is the table that holds the control_objectives relation/edge. The primary key declared below.
	ControlObjectivesTable = "subcontrol_control_objectives"
	// ControlObjectivesInverseTable is the table name for the ControlObjective entity.
	// It exists in this package in order to avoid circular dependency with the "controlobjective" package.
	ControlObjectivesInverseTable = "control_objectives"
	// TasksTable is the table that holds the tasks relation/edge. The primary key declared below.
	TasksTable = "subcontrol_tasks"
	// TasksInverseTable is the table name for the Task entity.
	// It exists in this package in order to avoid circular dependency with the "task" package.
	TasksInverseTable = "tasks"
	// NarrativesTable is the table that holds the narratives relation/edge.
	NarrativesTable = "narratives"
	// NarrativesInverseTable is the table name for the Narrative entity.
	// It exists in this package in order to avoid circular dependency with the "narrative" package.
	NarrativesInverseTable = "narratives"
	// NarrativesColumn is the table column denoting the narratives relation/edge.
	NarrativesColumn = "subcontrol_narratives"
	// RisksTable is the table that holds the risks relation/edge. The primary key declared below.
	RisksTable = "subcontrol_risks"
	// RisksInverseTable is the table name for the Risk entity.
	// It exists in this package in order to avoid circular dependency with the "risk" package.
	RisksInverseTable = "risks"
	// ActionPlansTable is the table that holds the action_plans relation/edge.
	ActionPlansTable = "action_plans"
	// ActionPlansInverseTable is the table name for the ActionPlan entity.
	// It exists in this package in order to avoid circular dependency with the "actionplan" package.
	ActionPlansInverseTable = "action_plans"
	// ActionPlansColumn is the table column denoting the action_plans relation/edge.
	ActionPlansColumn = "subcontrol_action_plans"
	// ProceduresTable is the table that holds the procedures relation/edge. The primary key declared below.
	ProceduresTable = "subcontrol_procedures"
	// ProceduresInverseTable is the table name for the Procedure entity.
	// It exists in this package in order to avoid circular dependency with the "procedure" package.
	ProceduresInverseTable = "procedures"
	// InternalPoliciesTable is the table that holds the internal_policies relation/edge. The primary key declared below.
	InternalPoliciesTable = "internal_policy_subcontrols"
	// InternalPoliciesInverseTable is the table name for the InternalPolicy entity.
	// It exists in this package in order to avoid circular dependency with the "internalpolicy" package.
	InternalPoliciesInverseTable = "internal_policies"
	// ControlOwnerTable is the table that holds the control_owner relation/edge.
	ControlOwnerTable = "subcontrols"
	// ControlOwnerInverseTable is the table name for the Group entity.
	// It exists in this package in order to avoid circular dependency with the "group" package.
	ControlOwnerInverseTable = "groups"
	// ControlOwnerColumn is the table column denoting the control_owner relation/edge.
	ControlOwnerColumn = "control_owner_id"
	// DelegateTable is the table that holds the delegate relation/edge.
	DelegateTable = "subcontrols"
	// DelegateInverseTable is the table name for the Group entity.
	// It exists in this package in order to avoid circular dependency with the "group" package.
	DelegateInverseTable = "groups"
	// DelegateColumn is the table column denoting the delegate relation/edge.
	DelegateColumn = "delegate_id"
	// OwnerTable is the table that holds the owner relation/edge.
	OwnerTable = "subcontrols"
	// OwnerInverseTable is the table name for the Organization entity.
	// It exists in this package in order to avoid circular dependency with the "organization" package.
	OwnerInverseTable = "organizations"
	// OwnerColumn is the table column denoting the owner relation/edge.
	OwnerColumn = "owner_id"
	// ControlTable is the table that holds the control relation/edge.
	ControlTable = "subcontrols"
	// ControlInverseTable is the table name for the Control entity.
	// It exists in this package in order to avoid circular dependency with the "control" package.
	ControlInverseTable = "controls"
	// ControlColumn is the table column denoting the control relation/edge.
	ControlColumn = "control_id"
	// ControlImplementationsTable is the table that holds the control_implementations relation/edge. The primary key declared below.
	ControlImplementationsTable = "subcontrol_control_implementations"
	// ControlImplementationsInverseTable is the table name for the ControlImplementation entity.
	// It exists in this package in order to avoid circular dependency with the "controlimplementation" package.
	ControlImplementationsInverseTable = "control_implementations"
	// ScheduledJobsTable is the table that holds the scheduled_jobs relation/edge. The primary key declared below.
	ScheduledJobsTable = "scheduled_job_subcontrols"
	// ScheduledJobsInverseTable is the table name for the ScheduledJob entity.
	// It exists in this package in order to avoid circular dependency with the "scheduledjob" package.
	ScheduledJobsInverseTable = "scheduled_jobs"
	// MappedToSubcontrolsTable is the table that holds the mapped_to_subcontrols relation/edge. The primary key declared below.
	MappedToSubcontrolsTable = "mapped_control_to_subcontrols"
	// MappedToSubcontrolsInverseTable is the table name for the MappedControl entity.
	// It exists in this package in order to avoid circular dependency with the "mappedcontrol" package.
	MappedToSubcontrolsInverseTable = "mapped_controls"
	// MappedFromSubcontrolsTable is the table that holds the mapped_from_subcontrols relation/edge. The primary key declared below.
	MappedFromSubcontrolsTable = "mapped_control_from_subcontrols"
	// MappedFromSubcontrolsInverseTable is the table name for the MappedControl entity.
	// It exists in this package in order to avoid circular dependency with the "mappedcontrol" package.
	MappedFromSubcontrolsInverseTable = "mapped_controls"
)

// Columns holds all SQL columns for subcontrol fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldCreatedBy,
	FieldUpdatedBy,
	FieldDeletedAt,
	FieldDeletedBy,
	FieldDisplayID,
	FieldTags,
	FieldDescription,
	FieldReferenceID,
	FieldAuditorReferenceID,
	FieldStatus,
	FieldSource,
	FieldReferenceFramework,
	FieldControlType,
	FieldCategory,
	FieldCategoryID,
	FieldSubcategory,
	FieldMappedCategories,
	FieldAssessmentObjectives,
	FieldAssessmentMethods,
	FieldControlQuestions,
	FieldImplementationGuidance,
	FieldExampleEvidence,
	FieldReferences,
	FieldControlOwnerID,
	FieldDelegateID,
	FieldOwnerID,
	FieldRefCode,
	FieldControlID,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "subcontrols"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"program_subcontrols",
	"user_subcontrols",
}

var (
	// EvidencePrimaryKey and EvidenceColumn2 are the table columns denoting the
	// primary key for the evidence relation (M2M).
	EvidencePrimaryKey = []string{"evidence_id", "subcontrol_id"}
	// ControlObjectivesPrimaryKey and ControlObjectivesColumn2 are the table columns denoting the
	// primary key for the control_objectives relation (M2M).
	ControlObjectivesPrimaryKey = []string{"subcontrol_id", "control_objective_id"}
	// TasksPrimaryKey and TasksColumn2 are the table columns denoting the
	// primary key for the tasks relation (M2M).
	TasksPrimaryKey = []string{"subcontrol_id", "task_id"}
	// RisksPrimaryKey and RisksColumn2 are the table columns denoting the
	// primary key for the risks relation (M2M).
	RisksPrimaryKey = []string{"subcontrol_id", "risk_id"}
	// ProceduresPrimaryKey and ProceduresColumn2 are the table columns denoting the
	// primary key for the procedures relation (M2M).
	ProceduresPrimaryKey = []string{"subcontrol_id", "procedure_id"}
	// InternalPoliciesPrimaryKey and InternalPoliciesColumn2 are the table columns denoting the
	// primary key for the internal_policies relation (M2M).
	InternalPoliciesPrimaryKey = []string{"internal_policy_id", "subcontrol_id"}
	// ControlImplementationsPrimaryKey and ControlImplementationsColumn2 are the table columns denoting the
	// primary key for the control_implementations relation (M2M).
	ControlImplementationsPrimaryKey = []string{"subcontrol_id", "control_implementation_id"}
	// ScheduledJobsPrimaryKey and ScheduledJobsColumn2 are the table columns denoting the
	// primary key for the scheduled_jobs relation (M2M).
	ScheduledJobsPrimaryKey = []string{"scheduled_job_id", "subcontrol_id"}
	// MappedToSubcontrolsPrimaryKey and MappedToSubcontrolsColumn2 are the table columns denoting the
	// primary key for the mapped_to_subcontrols relation (M2M).
	MappedToSubcontrolsPrimaryKey = []string{"mapped_control_id", "subcontrol_id"}
	// MappedFromSubcontrolsPrimaryKey and MappedFromSubcontrolsColumn2 are the table columns denoting the
	// primary key for the mapped_from_subcontrols relation (M2M).
	MappedFromSubcontrolsPrimaryKey = []string{"mapped_control_id", "subcontrol_id"}
)

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
			return true
		}
	}
	return false
}

// Note that the variables below are initialized by the runtime
// package on the initialization of the application. Therefore,
// it should be imported in the main as follows:
//
//	import _ "github.com/theopenlane/core/internal/ent/generated/runtime"
var (
	Hooks        [11]ent.Hook
	Interceptors [4]ent.Interceptor
	Policy       ent.Policy
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DisplayIDValidator is a validator for the "display_id" field. It is called by the builders before save.
	DisplayIDValidator func(string) error
	// DefaultTags holds the default value on creation for the "tags" field.
	DefaultTags []string
	// OwnerIDValidator is a validator for the "owner_id" field. It is called by the builders before save.
	OwnerIDValidator func(string) error
	// RefCodeValidator is a validator for the "ref_code" field. It is called by the builders before save.
	RefCodeValidator func(string) error
	// ControlIDValidator is a validator for the "control_id" field. It is called by the builders before save.
	ControlIDValidator func(string) error
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
)

const DefaultStatus enums.ControlStatus = "NOT_IMPLEMENTED"

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s enums.ControlStatus) error {
	switch s.String() {
	case "PREPARING", "NEEDS_APPROVAL", "CHANGES_REQUESTED", "APPROVED", "ARCHIVED", "NOT_IMPLEMENTED":
		return nil
	default:
		return fmt.Errorf("subcontrol: invalid enum value for status field: %q", s)
	}
}

const DefaultSource enums.ControlSource = "USER_DEFINED"

// SourceValidator is a validator for the "source" field enum values. It is called by the builders before save.
func SourceValidator(s enums.ControlSource) error {
	switch s.String() {
	case "FRAMEWORK", "TEMPLATE", "USER_DEFINED", "IMPORTED":
		return nil
	default:
		return fmt.Errorf("subcontrol: invalid enum value for source field: %q", s)
	}
}

const DefaultControlType enums.ControlType = "PREVENTATIVE"

// ControlTypeValidator is a validator for the "control_type" field enum values. It is called by the builders before save.
func ControlTypeValidator(ct enums.ControlType) error {
	switch ct.String() {
	case "PREVENTATIVE", "DETECTIVE", "CORRECTIVE", "DETERRENT":
		return nil
	default:
		return fmt.Errorf("subcontrol: invalid enum value for control_type field: %q", ct)
	}
}

// OrderOption defines the ordering options for the Subcontrol queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByCreatedBy orders the results by the created_by field.
func ByCreatedBy(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedBy, opts...).ToFunc()
}

// ByUpdatedBy orders the results by the updated_by field.
func ByUpdatedBy(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedBy, opts...).ToFunc()
}

// ByDeletedAt orders the results by the deleted_at field.
func ByDeletedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDeletedAt, opts...).ToFunc()
}

// ByDeletedBy orders the results by the deleted_by field.
func ByDeletedBy(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDeletedBy, opts...).ToFunc()
}

// ByDisplayID orders the results by the display_id field.
func ByDisplayID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDisplayID, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
}

// ByReferenceID orders the results by the reference_id field.
func ByReferenceID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldReferenceID, opts...).ToFunc()
}

// ByAuditorReferenceID orders the results by the auditor_reference_id field.
func ByAuditorReferenceID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAuditorReferenceID, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// BySource orders the results by the source field.
func BySource(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSource, opts...).ToFunc()
}

// ByReferenceFramework orders the results by the reference_framework field.
func ByReferenceFramework(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldReferenceFramework, opts...).ToFunc()
}

// ByControlType orders the results by the control_type field.
func ByControlType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldControlType, opts...).ToFunc()
}

// ByCategory orders the results by the category field.
func ByCategory(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCategory, opts...).ToFunc()
}

// ByCategoryID orders the results by the category_id field.
func ByCategoryID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCategoryID, opts...).ToFunc()
}

// BySubcategory orders the results by the subcategory field.
func BySubcategory(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSubcategory, opts...).ToFunc()
}

// ByControlOwnerID orders the results by the control_owner_id field.
func ByControlOwnerID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldControlOwnerID, opts...).ToFunc()
}

// ByDelegateID orders the results by the delegate_id field.
func ByDelegateID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDelegateID, opts...).ToFunc()
}

// ByOwnerID orders the results by the owner_id field.
func ByOwnerID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOwnerID, opts...).ToFunc()
}

// ByRefCode orders the results by the ref_code field.
func ByRefCode(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRefCode, opts...).ToFunc()
}

// ByControlID orders the results by the control_id field.
func ByControlID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldControlID, opts...).ToFunc()
}

// ByEvidenceCount orders the results by evidence count.
func ByEvidenceCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newEvidenceStep(), opts...)
	}
}

// ByEvidence orders the results by evidence terms.
func ByEvidence(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newEvidenceStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByControlObjectivesCount orders the results by control_objectives count.
func ByControlObjectivesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newControlObjectivesStep(), opts...)
	}
}

// ByControlObjectives orders the results by control_objectives terms.
func ByControlObjectives(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newControlObjectivesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByTasksCount orders the results by tasks count.
func ByTasksCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newTasksStep(), opts...)
	}
}

// ByTasks orders the results by tasks terms.
func ByTasks(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newTasksStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByNarrativesCount orders the results by narratives count.
func ByNarrativesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newNarrativesStep(), opts...)
	}
}

// ByNarratives orders the results by narratives terms.
func ByNarratives(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newNarrativesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByRisksCount orders the results by risks count.
func ByRisksCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newRisksStep(), opts...)
	}
}

// ByRisks orders the results by risks terms.
func ByRisks(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRisksStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByActionPlansCount orders the results by action_plans count.
func ByActionPlansCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newActionPlansStep(), opts...)
	}
}

// ByActionPlans orders the results by action_plans terms.
func ByActionPlans(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newActionPlansStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByProceduresCount orders the results by procedures count.
func ByProceduresCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newProceduresStep(), opts...)
	}
}

// ByProcedures orders the results by procedures terms.
func ByProcedures(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newProceduresStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByInternalPoliciesCount orders the results by internal_policies count.
func ByInternalPoliciesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newInternalPoliciesStep(), opts...)
	}
}

// ByInternalPolicies orders the results by internal_policies terms.
func ByInternalPolicies(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newInternalPoliciesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByControlOwnerField orders the results by control_owner field.
func ByControlOwnerField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newControlOwnerStep(), sql.OrderByField(field, opts...))
	}
}

// ByDelegateField orders the results by delegate field.
func ByDelegateField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newDelegateStep(), sql.OrderByField(field, opts...))
	}
}

// ByOwnerField orders the results by owner field.
func ByOwnerField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOwnerStep(), sql.OrderByField(field, opts...))
	}
}

// ByControlField orders the results by control field.
func ByControlField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newControlStep(), sql.OrderByField(field, opts...))
	}
}

// ByControlImplementationsCount orders the results by control_implementations count.
func ByControlImplementationsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newControlImplementationsStep(), opts...)
	}
}

// ByControlImplementations orders the results by control_implementations terms.
func ByControlImplementations(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newControlImplementationsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByScheduledJobsCount orders the results by scheduled_jobs count.
func ByScheduledJobsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newScheduledJobsStep(), opts...)
	}
}

// ByScheduledJobs orders the results by scheduled_jobs terms.
func ByScheduledJobs(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newScheduledJobsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByMappedToSubcontrolsCount orders the results by mapped_to_subcontrols count.
func ByMappedToSubcontrolsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newMappedToSubcontrolsStep(), opts...)
	}
}

// ByMappedToSubcontrols orders the results by mapped_to_subcontrols terms.
func ByMappedToSubcontrols(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newMappedToSubcontrolsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByMappedFromSubcontrolsCount orders the results by mapped_from_subcontrols count.
func ByMappedFromSubcontrolsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newMappedFromSubcontrolsStep(), opts...)
	}
}

// ByMappedFromSubcontrols orders the results by mapped_from_subcontrols terms.
func ByMappedFromSubcontrols(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newMappedFromSubcontrolsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newEvidenceStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(EvidenceInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, EvidenceTable, EvidencePrimaryKey...),
	)
}
func newControlObjectivesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ControlObjectivesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, ControlObjectivesTable, ControlObjectivesPrimaryKey...),
	)
}
func newTasksStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(TasksInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, TasksTable, TasksPrimaryKey...),
	)
}
func newNarrativesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(NarrativesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, NarrativesTable, NarrativesColumn),
	)
}
func newRisksStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(RisksInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, RisksTable, RisksPrimaryKey...),
	)
}
func newActionPlansStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ActionPlansInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, ActionPlansTable, ActionPlansColumn),
	)
}
func newProceduresStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ProceduresInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, ProceduresTable, ProceduresPrimaryKey...),
	)
}
func newInternalPoliciesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(InternalPoliciesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, InternalPoliciesTable, InternalPoliciesPrimaryKey...),
	)
}
func newControlOwnerStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ControlOwnerInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, ControlOwnerTable, ControlOwnerColumn),
	)
}
func newDelegateStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(DelegateInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, DelegateTable, DelegateColumn),
	)
}
func newOwnerStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OwnerInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, OwnerTable, OwnerColumn),
	)
}
func newControlStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ControlInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, ControlTable, ControlColumn),
	)
}
func newControlImplementationsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ControlImplementationsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, ControlImplementationsTable, ControlImplementationsPrimaryKey...),
	)
}
func newScheduledJobsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ScheduledJobsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, ScheduledJobsTable, ScheduledJobsPrimaryKey...),
	)
}
func newMappedToSubcontrolsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(MappedToSubcontrolsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, MappedToSubcontrolsTable, MappedToSubcontrolsPrimaryKey...),
	)
}
func newMappedFromSubcontrolsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(MappedFromSubcontrolsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, MappedFromSubcontrolsTable, MappedFromSubcontrolsPrimaryKey...),
	)
}

var (
	// enums.ControlStatus must implement graphql.Marshaler.
	_ graphql.Marshaler = (*enums.ControlStatus)(nil)
	// enums.ControlStatus must implement graphql.Unmarshaler.
	_ graphql.Unmarshaler = (*enums.ControlStatus)(nil)
)

var (
	// enums.ControlSource must implement graphql.Marshaler.
	_ graphql.Marshaler = (*enums.ControlSource)(nil)
	// enums.ControlSource must implement graphql.Unmarshaler.
	_ graphql.Unmarshaler = (*enums.ControlSource)(nil)
)

var (
	// enums.ControlType must implement graphql.Marshaler.
	_ graphql.Marshaler = (*enums.ControlType)(nil)
	// enums.ControlType must implement graphql.Unmarshaler.
	_ graphql.Unmarshaler = (*enums.ControlType)(nil)
)
