// Package slacknotify provides Slack webhook notification helpers shared across packages.
// Centralising the config and send helpers here breaks the import cycle that would otherwise
// arise when integration definitions (e.g. githubapp) need to send Slack messages using
// infrastructure that lives in ent/hooks.
package slacknotify
