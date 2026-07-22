package taskrules

import (
	"fmt"

	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/enums"
)

// Notification task rule IDs
const (
	RuleReviewDomainScan = "review-domain-scan"
)

// NotificationTaskRules generate suggested tasks from notification events. Trigger is
// create-only: a notification's topic is set once and never changes
var NotificationTaskRules = []entx.TaskRuleSpec{
	{
		RuleID:     RuleReviewDomainScan,
		Expression: fmt.Sprintf("value == %q", enums.NotificationTopicDomainScan.String()),
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
}
