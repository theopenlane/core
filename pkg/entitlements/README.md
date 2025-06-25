# Stripe integration

To test webhooks locally:

`brew install stripe/stripe-cli/stripe`

`stripe login` and then use your associated credentials

`stripe listen --forward-to localhost:17608/v1/stripe/webhook`

`stripe trigger payment_intent.succeeded`

## Unprocessed

If your webhook endpoint temporarily can’t process events, Stripe automatically resends the undelivered events to your endpoint for up to three days, increasing the time for your webhook endpoint to eventually receive and process all events.

stripe fixtures ./fixtures.json

stripe trigger payment_intent.succeeded --override payment_intent:amount=5000 --override payment_intent:currency=usd --add payment_intent:customer=cus_xxx

https://github.com/stripe/stripe-cli/tree/master/pkg/fixtures/triggers

## Subscription status

Basic information pulled up out of stripe's docs indicating the various scenarios in which a subscription can be in a given status

https://stripe.com/docs/billing/subscriptions/overview#subscription-status

### Active

- subscription moves into active status when trial ends and a payment method has been added
- if initial payment attempt fails, and moves into incomplete, but then the payment is successful, it moves into active
- if trial ends and no payment method has been added, subs moves into paused status; if payment method is added + processed, it moves back to active

### Incomplete

- a subscription moves into incomplete if the initial payment attempt fails

### IncompleteExpired

- if the first invoice is not paid within 23 hours, the subscription transitions to incomplete_expired

### PastDue

- when collection_method=charge_automatically, subs becomes past_due when payment is required but cannot be paid (due to failed payment or awaiting additional user actions)

### Paused

- A subscription can only enter a paused status when a trial ends without a payment method

### Trialing

- subscription status is in trailing if we create the initial subscription with a trial period

### Canceled or Unpaid

- after exhausting all payment retry attempts, the subscription will become canceled or unpaid
- subscription moves into cancelled if we set cancel_at_period_end: true and the period end passes
