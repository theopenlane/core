package taskrules

// SetupTaskNotificationGates declares, by schema snake name, which task-rule batches broadcast a
// SUGGESTED_TASKS notification on completion; the value is an optional CEL gate ("" = always)
var SetupTaskNotificationGates = map[string]string{
	"organization": notPersonalOrg,
	"onboarding":   "",
}
