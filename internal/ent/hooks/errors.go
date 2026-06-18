package hooks

import (
	"errors"
	"strings"
)

var (
	// ErrInternalServerError is returned when an internal error occurs.
	ErrInternalServerError = errors.New("internal server error")
	// ErrInvalidInput is returned when the input is invalid.
	ErrInvalidInput = errors.New("invalid input")
	// ErrPersonalOrgsNoChildren is returned when personal org attempts to add a child org
	ErrPersonalOrgsNoChildren = errors.New("personal organizations are not allowed to have child organizations")
	// ErrPersonalOrgsNoMembers is returned when personal org attempts to add members
	ErrPersonalOrgsNoMembers = errors.New("personal organizations are not allowed to have members other than the owner")
	// ErrOrgOwnerCannotBeDeleted is returned when an org owner is attempted to be deleted
	ErrOrgOwnerCannotBeDeleted = errors.New("organization owner cannot be deleted, it must be transferred to a new owner first")
	// ErrOrgOwnerCannotBeUpdated is returned when an org owner role is attempted to be changed directly via mutations
	ErrOrgOwnerCannotBeUpdated = errors.New("organization owner role cannot be changed directly, it must be transferred")
	// ErrPersonalOrgsNoUser is returned when personal org has no user associated, so no permissions can be added
	ErrPersonalOrgsNoUser = errors.New("personal organizations missing user association")
	// ErrUserNotInOrg is returned when a user is not a member of an organization when trying to add them to a group
	ErrUserNotInOrg = errors.New("user not in organization")
	// ErrUnsupportedFGARole is returned when a role is assigned that is not supported in our fine grained authorization system
	ErrUnsupportedFGARole = errors.New("unsupported role")
	// ErrMissingRole is returned when an update request is made that contains no role
	ErrMissingRole = errors.New("missing role in update")
	// ErrUserAlreadyOrgMember is returned when an user attempts to be invited to an org they are already a member of
	ErrUserAlreadyOrgMember = errors.New("user already member of organization")
	// ErrUserAlreadySubscriber is returned when an user attempts to subscribe to an organization but is already a subscriber
	ErrUserAlreadySubscriber = errors.New("subscriber already exists")
	// ErrEmailRequired is returned when an email is required but not provided
	ErrEmailRequired = errors.New("email is required but not provided")
	// ErrMaxAttempts is returned when a user has reached the max attempts to resend an invitation to an org
	ErrMaxAttempts = errors.New("too many attempts to resend org invitation")
	// ErrMaxAttemptsAssessments is returned when a user has reached the max attempts to resend an assessment
	ErrMaxAttemptsAssessments = errors.New("too many attempts to resend assessment invitation")
	// ErrMaxSubscriptionAttempts is returned when a user has reached the max attempts to subscribe to an org
	ErrMaxSubscriptionAttempts = errors.New("too many attempts to resend org subscription email")
	// ErrAssessmentInCompleted is returned when attempting to resend an email for an assessment that is already completed
	ErrAssessmentInCompleted = errors.New("assessment is already completed")
	// ErrMissingRecipientEmail is returned when an email is required but not provided
	ErrMissingRecipientEmail = errors.New("recipient email is required but not provided")
	// ErrMissingRequiredName is returned when a name is required but not provided
	ErrMissingRequiredName = errors.New("name or display name is required but not provided")
	// ErrTooManyAvatarFiles is returned when a user attempts to upload more than one avatar file
	ErrTooManyAvatarFiles = errors.New("too many avatar files uploaded, only one is allowed")
	// ErrNoControls is returned when a subcontrol has no controls assigned
	ErrNoControls = errors.New("subcontrol must have at least one control assigned")
	// ErrUnableToCast is returned when a type assertion fails
	ErrUnableToCast = errors.New("unable to cast")
	// ErrNoSubscriptions is returned when an organization has no subscriptions
	ErrNoSubscriptions = errors.New("organization has no subscriptions")
	// ErrTooManySubscriptions is returned when an organization has too many subscriptions
	ErrTooManySubscriptions = errors.New("organization has too many subscriptions")
	// ErrTooManyPrices is returned when an organization has too many subscriptions
	ErrTooManyPrices = errors.New("organization has too many prices on a subscription")
	// ErrNoPrices is returned when a subscription has no price
	ErrNoPrices = errors.New("subscription has no price")
	// ErrManagedGroup is returned when a user attempts to modify a managed group
	ErrManagedGroup = errors.New("managed groups cannot be modified")
	// ErrMaxAttemptsOrganization is returned when the max attempts have been reached to create an organization via onboarding
	ErrMaxAttemptsOrganization = errors.New("too many attempts to create organization")
	// ErrEmailDomainNotAllowed is returned when an email domain is not allowed to be used for an organization
	ErrEmailDomainNotAllowed = errors.New("email domain not allowed in organization")
	// ErrUserNotFound is returned when a user is not found in the system
	ErrUserNotFound = errors.New("user not found")
	// ErrCronRequired is returned when a user does not provide a cron expression
	ErrCronRequired = errors.New("cron expression must be specified")
	// ErrZeroTimeNotAllowed is returned when you try to set a non usable time value
	ErrZeroTimeNotAllowed = errors.New("time cannot be empty. Provide a valid time/date")
	// ErrFutureTimeNotAllowed is returned when you try to set a time into the future.
	// future being any second/minute past the current time of validation
	ErrFutureTimeNotAllowed = errors.New("time cannot be in the future")
	// ErrPastTimeNotAllowed is returned when you try to set a time into the past.
	ErrPastTimeNotAllowed = errors.New("time cannot be in the past")
	// ErrFieldRequired is returned when a field is required but not provided
	ErrFieldRequired = errors.New("field is required but not provided")
	// ErrOwnerIDNotExists is returned when an owner_id cannot be found
	ErrOwnerIDNotExists = errors.New("owner_id is required")
	// ErrArchivedProgramUpdateNotAllowed is returned when an archived program is updated. It only
	// allows updates if the status is changed
	ErrArchivedProgramUpdateNotAllowed = errors.New("you cannot update an archived program")
	// ErrNotSingularUpload is returned when a user is importing content to create a schema
	// and they upload more than one file
	ErrNotSingularUpload = errors.New("multiple uploads not supported")
	// ErrSSONotEnforceable makes sure the connection has been tested before it can be enforced for an org
	ErrSSONotEnforceable = errors.New("you cannot enforce sso without testing the connection works correctly")
	// ErrExemptDomainAllowedConflict is returned when a domain appears in both the sso exempt domains and the allowed email domains lists
	ErrExemptDomainAllowedConflict = errors.New("a domain cannot be in both the sso exempt domains and allowed email domains lists")
	// ErrUnableToDetermineEventID is returned when we cannot determine the event ID for an event
	ErrUnableToDetermineEventID = errors.New("unable to determine event ID")
	// ErrGalaRuntimeUnavailable is returned when gala dual emit is enabled but runtime is unavailable
	ErrGalaRuntimeUnavailable = errors.New("gala runtime unavailable")
	// ErrGalaMutationEnqueueFailed is returned when gala mutation enqueue fails.
	ErrGalaMutationEnqueueFailed = errors.New("gala mutation enqueue failed")
	// ErrNotSingularTrustCenter is returned when an org is trying to create multiple trust centers
	ErrNotSingularTrustCenter = errors.New("you can only create/manage one trust center at a time")
	// ErrStatusApprovedNotAllowed is returned when a user attempts to set status to APPROVED without being in the approver or delegate group
	ErrStatusApprovedNotAllowed = errors.New("you must be in the approver group to mark as approved")
	// ErrInvalidChannel is returned when an invalid notification channel is provided
	ErrInvalidChannel = errors.New("invalid channel")
	// ErrTemplateIDRequired is returned when an assessment is created without a template
	ErrTemplateIDRequired = errors.New("template id required when creating an assessment")
	// ErrTemplateNotFound is returned when an assessment is created with a non existing template
	ErrTemplateNotFound = errors.New("template does not exist")
	// ErrTemplateNotQuestionnaire is returned when an assessment tries to use a wrong template type
	ErrTemplateNotQuestionnaire = errors.New("template must be a questionnaire")
	// ErrTrustCenterIDRequired is returned when the trustcenter id is not provided
	// when creating a customer for the trust center
	ErrTrustCenterIDRequired = errors.New("trustcenter entity must include a trustcenter id")
	// ErrUnableToCreateContact is returned when a contact could not be created
	ErrUnableToCreateContact = errors.New("unable to create a contact")
	// ErrUnableToCreateAssessmentResponse is returned when an assessment response could not be created
	ErrUnableToCreateAssessmentResponse = errors.New("unable to create assessment response")
	// ErrTooManyLogoFiles is returned when a user attempts to upload more than one logo file
	ErrTooManyLogoFiles = errors.New("too many logo files uploaded, only one is allowed")
	// ErrTooManyFaviconFiles is returned when a user attempts to upload more than one favicon file
	ErrTooManyFaviconFiles = errors.New("too many favicon files uploaded, only one is allowed")
	// ErrTooManyHeroImageFiles is returned when a user attempts to upload more than one hero image file
	ErrTooManyHeroImageFiles = errors.New("too many hero image files uploaded, only one is allowed")
	// ErrMissingTrustCenterID is returned when a trust center id is required but not provided
	ErrMissingTrustCenterID = errors.New("trust center id is required")
	// ErrMissingFileID is returned when a file id is required but not provided
	ErrMissingFileID = errors.New("missing file id")
	// ErrCannotSetFileOnCreate is returned when trying to set a file id on create mutations
	ErrCannotSetFileOnCreate = errors.New("cannot set file id on create")
	// ErrCacheRefreshFailed is returned when the cache refresh request fails
	ErrCacheRefreshFailed = errors.New("cache refresh request failed")
	// ErrNoOrganizationID is returned when no organization ID is found in context
	ErrNoOrganizationID = errors.New("no valid organization ID found")
	// ErrNDATemplateRequired is returned when nda requests mutation runs but there is no file for the
	// user to sign
	ErrNDATemplateRequired = errors.New("you need a nda template before a request can be made")
	// ErrMutationMissingID is returned when a mutation does not have an ID
	ErrMutationMissingID = errors.New("mutation missing ID")
	// ErrProposedChangesNotSupported is returned when proposed changes are not supported for a schema type
	ErrProposedChangesNotSupported = errors.New("proposed changes not supported for this schema type")
	// ErrFailedToGetUserFromContext is returned when the user cannot be resolved from the context
	ErrFailedToGetUserFromContext = errors.New("failed to get user from context")
	// ErrFailedToGetObjectOwnerID is returned when the owner id cannot be resolved for a workflow object
	ErrFailedToGetObjectOwnerID = errors.New("failed to get object owner id")
	// ErrFailedToQueryObjectRefs is returned when workflow object refs cannot be queried
	ErrFailedToQueryObjectRefs = errors.New("failed to query object refs")
	// ErrFailedToComputeProposalHash is returned when a workflow proposal hash cannot be computed
	ErrFailedToComputeProposalHash = errors.New("failed to compute proposal hash")
	// ErrFailedToQueryExistingProposal is returned when an existing proposal cannot be queried
	ErrFailedToQueryExistingProposal = errors.New("failed to query existing proposal")
	// ErrFailedToUpdateProposal is returned when a workflow proposal cannot be updated
	ErrFailedToUpdateProposal = errors.New("failed to update proposal")
	// ErrFailedToBeginTransaction is returned when a workflow transaction cannot be started
	ErrFailedToBeginTransaction = errors.New("failed to begin transaction")
	// ErrFailedToCreateWorkflowInstance is returned when a workflow instance cannot be created
	ErrFailedToCreateWorkflowInstance = errors.New("failed to create workflow instance")
	// ErrFailedToCreateWorkflowObjectRef is returned when a workflow object ref cannot be created
	ErrFailedToCreateWorkflowObjectRef = errors.New("failed to create workflow object ref")
	// ErrFailedToCreateWorkflowProposal is returned when a workflow proposal cannot be created
	ErrFailedToCreateWorkflowProposal = errors.New("failed to create workflow proposal inside of ent hooks")
	// ErrFailedToCreateWorkflowProposal is returned when a workflow proposal cannot be created
	ErrFailedToUpdateWorkflowProposal = errors.New("failed to update workflow proposal inside of ent hooks")
	// ErrFailedToLinkProposalToInstance is returned when a proposal cannot be linked to an instance
	ErrFailedToLinkProposalToInstance = errors.New("failed to link proposal to instance")
	// ErrFailedToCommitProposalTransaction is returned when proposal staging transaction cannot be committed
	ErrFailedToCommitProposalTransaction = errors.New("failed to commit proposal transaction")
	// ErrFailedToQueryAssignments is returned when workflow assignments cannot be queried
	ErrFailedToQueryAssignments = errors.New("failed to query assignments")
	// ErrFailedToInvalidateAssignment is returned when an assignment cannot be invalidated
	ErrFailedToInvalidateAssignment = errors.New("failed to invalidate assignment")
	// ErrWorkflowProposalMissingObjectRef is returned when a proposal is missing its object ref edge
	ErrWorkflowProposalMissingObjectRef = errors.New("workflow proposal missing object ref")
	// ErrFailedToDeriveObjectFromRef is returned when a workflow object cannot be derived from a ref
	ErrFailedToDeriveObjectFromRef = errors.New("failed to derive object from ref")
	// ErrFailedToLoadWorkflowObject is returned when a workflow object cannot be loaded
	ErrFailedToLoadWorkflowObject = errors.New("failed to load workflow object")
	// ErrFailedToFindMatchingDefinitions is returned when workflow definitions cannot be matched
	ErrFailedToFindMatchingDefinitions = errors.New("failed to find matching definitions")
	// ErrFailedToRecordAssignmentInvalidationEvent is returned when assignment invalidation events cannot be recorded
	ErrFailedToRecordAssignmentInvalidationEvent = errors.New("failed to record assignment invalidation event")
	// ErrFailedToResolveInvalidationNotificationOwner is returned when the owner for invalidation notifications cannot be resolved
	ErrFailedToResolveInvalidationNotificationOwner = errors.New("failed to resolve invalidation notification owner")
	// ErrFailedToSendInvalidationNotification is returned when invalidation notifications cannot be sent
	ErrFailedToSendInvalidationNotification = errors.New("failed to send invalidation notification")
	// ErrFailedToQueryWorkflowProposal is returned when a workflow proposal cannot be queried
	ErrFailedToQueryWorkflowProposal = errors.New("failed to query workflow proposal")
	// ErrFailedToQueryWorkflowInstances is returned when workflow instances cannot be queried
	ErrFailedToQueryWorkflowInstances = errors.New("failed to query workflow instances")
	// ErrFailedToLoadWorkflowProposalForTrigger is returned when a submitted proposal cannot be loaded for triggering
	ErrFailedToLoadWorkflowProposalForTrigger = errors.New("failed to load workflow proposal for trigger")
	// ErrFailedToResumeWorkflowInstance is returned when a workflow instance cannot be resumed
	ErrFailedToResumeWorkflowInstance = errors.New("failed to resume workflow instance")
	// ErrFailedToTriggerWorkflow is returned when a workflow cannot be triggered
	ErrFailedToTriggerWorkflow = errors.New("failed to trigger workflow")
	// ErrMissingIDForTrustCenterNDARequest is returned when a mutation for trust center nda request is missing the ID field, which is required to determine the trust center and send the appropriate email
	ErrMissingIDForTrustCenterNDARequest = errors.New("missing ID for trust center NDA request mutation")
	// ErrVendorScoringQuestionNotFound is returned when a question key cannot be resolved in the scoring config
	ErrVendorScoringQuestionNotFound = errors.New("vendor scoring question not found in config")
	// ErrStartDateLaterThanEndDate is returned when a program mutation has a start date that is later than the end date
	ErrStartDateLaterThanEndDate = errors.New("mutation's start date cannot be later than end date")
	// ErrFailedToGetOldEndDate is returned when a program mutation's old end date cannot be retrieved, which is necessary to validate that the start date is not later than the old end date
	ErrFailedToGetOldEndDate = errors.New("could not get old end date for mutation")
	// ErrFailedToGetStartDate is returned when a program mutation's start date cannot be retrieved, which is necessary to validate that the old start date is not later than the end date
	ErrFailedToGetOldStartDate = errors.New("could not get old start date for mutation")
	// ErrFailedToGetProgramByID is returned when a program cannot be retrieved by ID when updating
	ErrFailedToGetProgramByID = errors.New("could not get program by id when updating")
	// ErrFailedToGetIdsForProgramUpdate is returned when a program mutation's ids cannot be retrieved, which is necessary to validate bulk updates
	ErrFailedToGetIDsForProgramUpdate = errors.New("could not get ids for mutation program update")
	// ErrInvalidTemplateDefaults is returned when template defaults do not satisfy the template's jsonconfig schema
	ErrInvalidTemplateDefaults = errors.New("template defaults do not satisfy the template schema")
	// ErrMissingNDATemplateFile is returned when the NDA template has no attached file for attestation
	ErrMissingNDATemplateFile = errors.New("NDA template must have at least one attached file")
	// errMissingTemplate is returned when an nda request is made but there is no template for the user to sign
	errMissingTemplate = errors.New("missing template")
	// errDocInfoDoesNotMatchCaller is returned when the document data submitted for an NDA does not match the authenticated user's information, such as email
	errDocInfoDoesNotMatchCaller = errors.New("NDA submission does not match authenticated user")
	// errUserHasAlreadySignedNDA is returned when a user attempts to submit an NDA document but has already signed the NDA
	errUserHasAlreadySignedNDA = errors.New("user has already signed the NDA")
	// errValidationFailed is returned when validation of an NDA submission fails for any reason not covered by the other specific errors
	errValidationFailed = errors.New("validation failed")
	// errMustBeAnonymousUser is returned when an NDA submission is attempted by an authenticated user, as only anonymous users should be submitting ANDAs for signing
	errMustBeAnonymousUser = errors.New("must be an anonymous user")
	// errMissingResponse is returned when the document data mutation does not include the response field, which is required for NDA submissions
	errMissingResponse = errors.New("missing response")
	// errOnlyOneDocumentData is returned when a user attempts to submit more than one document data file for the same NDA, as only one submission should be allowed per NDA
	errOnlyOneDocumentData = errors.New("you can only upload one document data file for an nda")
	// ErrFailedToUploadAttestedPDF is returned when the attested NDA PDF upload fails
	ErrFailedToUploadAttestedPDF = errors.New("failed to upload attested PDF")
	// ErrNoUploadedFiles is returned when the upload pipeline returns zero files
	ErrNoUploadedFiles = errors.New("no files returned from upload")
	// ErrFailedToAssociateFile is returned when linking an uploaded file to its document data record fails
	ErrFailedToAssociateFile = errors.New("failed to associate file with document data")
	// ErrFailedToFetchNDATemplate is returned when the NDA template cannot be queried
	ErrFailedToFetchNDATemplate = errors.New("failed to fetch NDA template")
	// ErrFailedToFetchNDATemplateFiles is returned when the template's file edges cannot be loaded
	ErrFailedToFetchNDATemplateFiles = errors.New("failed to fetch NDA template files")
	// ErrFailedToMarshalDocumentData is returned when document data cannot be serialized to JSON
	ErrFailedToMarshalDocumentData = errors.New("failed to marshal document data")
	// ErrFailedToUnmarshalNDAMetadata is returned when document data JSON cannot be deserialized into the NDA struct
	ErrFailedToUnmarshalNDAMetadata = errors.New("failed to unmarshal NDA metadata")
	// ErrFailedToDownloadNDAPDF is returned when the original NDA PDF cannot be downloaded from storage
	ErrFailedToDownloadNDAPDF = errors.New("failed to download original NDA PDF")
	// ErrFailedToCreateAttestedPDF is returned when appending the attestation page to the PDF fails
	ErrFailedToCreateAttestedPDF = errors.New("failed to create attested PDF")
	// ErrFailedToFetchTrustCenter is returned when the trust center record cannot be queried
	ErrFailedToFetchTrustCenter = errors.New("failed to fetch trust center")
	// ErrFailedToCreateAttestationCert is returned when generating the attestation certificate PDF page fails
	ErrFailedToCreateAttestationCert = errors.New("failed to create attestation certificate")
	// ErrFailedToMergeAttestationPage is returned when merging the attestation page into the original PDF fails
	ErrFailedToMergeAttestationPage = errors.New("failed to merge attestation page")
	// ErrFailedToGenerateAttestationPDF is returned when the attestation PDF output fails
	ErrFailedToGenerateAttestationPDF = errors.New("failed to generate attestation PDF")
	// ErrInvalidScope is returned when a scope is not assignable to service subjects
	ErrInvalidScope = errors.New("scope is not assignable to service subjects")
)

// IsUniqueConstraintError reports if the error resulted from a DB uniqueness constraint violation.
// e.g. duplicate value in unique index.
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	for _, s := range []string{
		"Error 1062",                 // MySQL
		"violates unique constraint", // Postgres
		"UNIQUE constraint failed",   // SQLite
	} {
		if strings.Contains(err.Error(), s) {
			return true
		}
	}

	return false
}
