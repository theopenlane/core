package taskrules

import (
	"fmt"

	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/enums"
)

// NotificationTaskRuleTopics are the notification topics a task rule below can fire on
var NotificationTaskRuleTopics = []string{enums.NotificationTopicDomainScan.String()}

// NotificationTaskRules generate suggested tasks from notification events. Trigger is
// create-only: a notification's topic is set once and never changes
var NotificationTaskRules = []entx.TaskRuleSpec{
	{
		RuleID:     "review-domain-scan",
		Expression: fmt.Sprintf("value == %q", enums.NotificationTopicDomainScan.String()),
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
}
