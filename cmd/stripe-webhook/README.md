# Stripe Webhook Migration Tool

Automated tool for managing Stripe webhook API version upgrades.

## Problem

When upgrading the Stripe Go SDK to a new API version, webhooks must be manually recreated through the Stripe dashboard with the correct API version. This process is:
- Error-prone and manual
- Requires careful coordination to avoid missing events
- Difficult to rollback if issues occur
- Lacks automation and documentation

## Solution

This CLI tool automates the entire webhook migration process following Stripe's recommended approach for zero-downtime migrations.

## Migration Workflow

### Stage 1: Ready
Your existing webhook uses an older API version than your SDK.

```bash
stripe-webhook status
```

### Stage 2: Create New Webhook (Disabled)
Creates a new webhook at the same URL with a query parameter (`?stripe_api_version=new`) in disabled state.

```bash
stripe-webhook migrate --step create
```

**Important**: Save the webhook secret from the output. Update your configuration:
```bash
# Add the new secret to your environment
STRIPE_WEBHOOK_SECRET_NEW=whsec_...
```

### Stage 3: Enable Dual Processing
Enable the new webhook to receive events alongside the old webhook.

```bash
stripe-webhook migrate --step enable
```

Both webhooks now send events to your endpoint. Your application receives duplicate events.

### Stage 4: Update Application Code
Update your webhook handler to:
1. Check the `stripe_api_version` query parameter
2. Process only events from the new webhook (version=new)
3. Return 400 for events from the old webhook (version=old or no parameter)

```go
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    version := r.URL.Query().Get("stripe_api_version")

    // Process only new version events
    if version != "new" {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    // Continue processing with new Stripe SDK...
}
```

Deploy this code to all environments.

### Stage 5: Disable Old Webhook
After confirming the new webhook processes events correctly, disable the old webhook.

```bash
stripe-webhook migrate --step disable
```

Only the new webhook is now active. Monitor for any issues.

### Stage 6: Cleanup Old Webhook
After a monitoring period (recommended: 24-48 hours), delete the old webhook.

```bash
stripe-webhook migrate --step cleanup
```

The old webhook is deleted but history is preserved in Stripe.

### Stage 7: Promote (Optional)
Remove the version query parameter from the new webhook URL for cleaner URLs.

```bash
stripe-webhook migrate --step promote
```

Updates the webhook URL from `https://api.openlane.com/v1/stripe/webhook?stripe_api_version=new` to `https://api.openlane.com/v1/stripe/webhook?stripe_api_version=old`

## Commands

### List Webhooks
```bash
stripe-webhook list
```

Shows all webhook endpoints in your Stripe account.

### Check Migration Status
```bash
stripe-webhook status --webhook-url https://api.openlane.com/v1/stripe/webhook
```

Displays current migration state and next action.

### Automated Migration
```bash
stripe-webhook migrate --auto
```

Automatically executes the next migration step (except code deployment).

### Manual Step Execution
```bash
# Create new webhook
stripe-webhook migrate --step create

# Enable new webhook
stripe-webhook migrate --step enable

# Disable old webhook (after code deployment)
stripe-webhook migrate --step disable

# Cleanup old webhook
stripe-webhook migrate --step cleanup

# Rollback migration
stripe-webhook migrate --step rollback

# Promote new webhook
stripe-webhook migrate --step promote
```

### Custom Events
```bash
stripe-webhook migrate --step create --events customer.subscription.updated --events invoice.paid
```

## Configuration

### Environment Variables
```bash
# Stripe API key (required)
export STRIPE_API_KEY=sk_test_...
# or
export STRIPE_SECRET_KEY=sk_test_...

# Webhook URL (required)
export STRIPE_WEBHOOK_URL=https://api.openlane.com/v1/stripe/webhook
```

### CLI Flags
```bash
stripe-webhook status \
  --stripe-key sk_test_... \
  --webhook-url https://api.openlane.com/v1/stripe/webhook
```

## Rollback Process

If issues occur during migration, rollback to the old webhook:

```bash
stripe-webhook migrate --step rollback
```

This will:
1. Disable the new webhook
2. Re-enable the old webhook
3. Restore normal operation

Then:
1. Revert your application code changes
2. Investigate the issue
3. Fix problems before retrying migration

## Migration States

| State | Description | Can Rollback |
|-------|-------------|--------------|
| `none` | No migration needed (versions match) | N/A |
| `ready` | Version mismatch detected | No |
| `new_created` | New webhook created (disabled) | Yes |
| `dual_processing` | Both webhooks active | Yes |
| `transitioned` | Only new webhook active | Yes |
| `complete` | Old webhook deleted | No |

## Safety Features

- **State validation**: Each step validates the current state before executing
- **Disabled creation**: New webhooks created in disabled state to prevent premature event processing
- **Rollback capability**: Can revert to old webhook at any stage before cleanup
- **Query parameters**: Uses URL query parameters to distinguish webhook versions without URL conflicts
- **Event parity**: Automatically copies event subscriptions from old to new webhook

## Complete Example

```bash
# 1. Check if migration is needed
stripe-webhook status
# Output: Migration Stage: ready

# 2. Create new webhook
stripe-webhook migrate --step create
# Save the webhook secret: whsec_abc123...

# 3. Update environment variables
export STRIPE_WEBHOOK_SECRET_NEW=whsec_abc123...

# 4. Enable dual processing
stripe-webhook migrate --step enable

# 5. Update application code to check query parameter
# Deploy code to all environments

# 6. Disable old webhook
stripe-webhook migrate --step disable

# 7. Monitor for 24-48 hours
stripe-webhook status

# 8. Cleanup old webhook
stripe-webhook migrate --step cleanup

# 9. (Optional) Promote new webhook
stripe-webhook migrate --step promote

# 10. Update environment variables (remove old secret)
unset STRIPE_WEBHOOK_SECRET_NEW
export STRIPE_WEBHOOK_SECRET=whsec_abc123...
```

## Troubleshooting

### Webhook Not Found
```
Error: webhook endpoint not found
```
Solution: Verify `STRIPE_WEBHOOK_URL` matches an existing webhook in Stripe dashboard.

### Multiple Webhooks Found
```
Error: multiple webhook endpoints found for URL
```
Solution: Manually remove duplicate webhooks from Stripe dashboard before running migration.

### Version Mismatch
```
Migration Stage: none
```
Solution: No migration needed. Your webhook already uses the current SDK API version.

### Invalid State
```
Error: invalid migration state for requested operation
```
Solution: Run `stripe-webhook status` to check current state and follow the recommended next action.

## Building

```bash
go build -o stripe-webhook ./cmd/stripe-webhook
```

## Testing

The migration logic is in `pkg/entitlements/webhook_migration.go` with comprehensive functions for each step.

## Resources

- [Stripe Webhook Versioning Guide](https://docs.stripe.com/webhooks/versioning)
- [Stripe API Versioning](https://docs.stripe.com/api/versioning)
- [Webhook Best Practices](https://docs.stripe.com/webhooks/best-practices)
