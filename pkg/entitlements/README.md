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

