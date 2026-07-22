package taskrules

import (
	"fmt"

	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/enums"
)

// NotificationTaskRules generate suggested tasks from notification events. Trigger is
// create-only: a notification's topic is set once and never changes
var NotificationTaskRules = []entx.TaskRuleSpec{
	{
		RuleID:     "review-domain-scan",
		Expression: fmt.Sprintf("value == %q", enums.NotificationTopicDomainScan.String()),
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
}
