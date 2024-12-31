# Stripe integration

`brew install stripe/stripe-cli/stripe`

`stripe login` and then use your associated credentials

`stripe listen --forward-to localhost:17608/v1/stripe/webhook`

`stripe trigger payment_intent.succeeded`

## Unprocessed

If your webhook endpoint temporarily can’t process events, Stripe automatically resends the undelivered events to your endpoint for up to three days, increasing the time for your webhook endpoint to eventually receive and process all events.

This guide explains how to speed up that process by manually processing the undelivered events.

